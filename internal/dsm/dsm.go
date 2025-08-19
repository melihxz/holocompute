package dsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/melihxz/holocompute/internal/hyperbus"
	"github.com/melihxz/holocompute/internal/log"
)

// ArrayID uniquely identifies a shared array
type ArrayID string

// PageID identifies a page within an array
type PageID int32

// Version represents a version of a page
type Version int64

// PageSize is the size of a page in bytes
const PageSize = 64 * 1024 // 64 KiB

// Page represents a page of data
type Page struct {
	ID      PageID
	Version Version
	Data    []byte
	storage *pageStorage
}

// NewPage creates a new page
func NewPage(id PageID, version Version) *Page {
	return &Page{
		ID:      id,
		Version: version,
		Data:    make([]byte, PageSize),
		storage: newPageStorage(PageSize),
	}
}

// GetInt64 reads a 64-bit integer from the page at the specified element index
func (p *Page) GetInt64(elementIndex int) (int64, error) {
	offset := elementIndex * 8
	return p.storage.getInt64(offset)
}

// SetInt64 writes a 64-bit integer to the page at the specified element index
func (p *Page) SetInt64(elementIndex int, value int64) error {
	offset := elementIndex * 8
	return p.storage.setInt64(offset, value)
}

// GetFloat32 reads a 32-bit float from the page at the specified element index
func (p *Page) GetFloat32(elementIndex int) (float32, error) {
	offset := elementIndex * 4
	return p.storage.getFloat32(offset)
}

// SetFloat32 writes a 32-bit float to the page at the specified element index
func (p *Page) SetFloat32(elementIndex int, value float32) error {
	offset := elementIndex * 4
	return p.storage.setFloat32(offset, value)
}

// Array represents a distributed shared array
type Array struct {
	ID          ArrayID
	Length      int
	NumPages    int
	PageMapping map[PageID]hyperbus.NodeID
	Version     Version
	mu          sync.RWMutex
}

// NewArray creates a new array
func NewArray(length int) *Array {
	pageCount := (length*8 + PageSize - 1) / PageSize // Assuming 8 bytes per element for now

	return &Array{
		ID:          ArrayID(uuid.New().String()),
		Length:      length,
		NumPages:    pageCount,
		PageMapping: make(map[PageID]hyperbus.NodeID),
		Version:     1,
	}
}

// PageCount returns the number of pages in the array
func (a *Array) PageCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.NumPages
}

// GetPageOwner returns the node that owns the specified page
func (a *Array) GetPageOwner(pageID PageID) (hyperbus.NodeID, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	nodeID, exists := a.PageMapping[pageID]
	return nodeID, exists
}

// SetPageOwner sets the owner of the specified page
func (a *Array) SetPageOwner(pageID PageID, nodeID hyperbus.NodeID) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.PageMapping[pageID] = nodeID
}

// MemoryManager manages distributed shared memory
type MemoryManager struct {
	arrays map[ArrayID]*Array
	bus    *hyperbus.Bus
	logger *log.Logger
	pages  map[pageKey]*Page // local page storage
	mu     sync.RWMutex
}

// pageKey uniquely identifies a page
type pageKey struct {
	arrayID ArrayID
	pageID  PageID
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager(bus *hyperbus.Bus, logger *log.Logger) *MemoryManager {
	return &MemoryManager{
		arrays: make(map[ArrayID]*Array),
		bus:    bus,
		logger: logger,
		pages:  make(map[pageKey]*Page),
	}
}

// CreateArray creates a new shared array
func (mm *MemoryManager) CreateArray(ctx context.Context, length int) (*Array, error) {
	array := NewArray(length)

	mm.mu.Lock()
	mm.arrays[array.ID] = array
	mm.mu.Unlock()

	mm.logger.Info("created new array", "array_id", array.ID, "length", length, "pages", array.PageCount)

	return array, nil
}

// GetArray retrieves an existing array
func (mm *MemoryManager) GetArray(ctx context.Context, arrayID ArrayID) (*Array, error) {
	mm.mu.RLock()
	array, exists := mm.arrays[arrayID]
	mm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("array not found: %s", arrayID)
	}

	return array, nil
}

// DeleteArray deletes an array
func (mm *MemoryManager) DeleteArray(ctx context.Context, arrayID ArrayID) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	_, exists := mm.arrays[arrayID]
	if !exists {
		return fmt.Errorf("array not found: %s", arrayID)
	}

	delete(mm.arrays, arrayID)
	mm.logger.Info("deleted array", "array_id", arrayID)

	return nil
}

// RequestPage requests a page from the owner
func (mm *MemoryManager) RequestPage(ctx context.Context, arrayID ArrayID, pageID PageID, version Version) (*Page, error) {
	// Get the array
	array, err := mm.GetArray(ctx, arrayID)
	if err != nil {
		return nil, fmt.Errorf("failed to get array: %w", err)
	}

	// Get the owner of the page
	ownerID, exists := array.GetPageOwner(pageID)
	if !exists {
		return nil, fmt.Errorf("page owner not found for page %d in array %s", pageID, arrayID)
	}

	// If we're the owner, return the local page
	if ownerID == mm.bus.LocalNode().ID {
		return mm.getLocalPage(ctx, arrayID, pageID, version)
	}

	// Request the page from the owner
	page, err := mm.requestRemotePage(ctx, ownerID, arrayID, pageID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to request remote page: %w", err)
	}

	return page, nil
}

// getLocalPage retrieves a page from local storage
func (mm *MemoryManager) getLocalPage(ctx context.Context, arrayID ArrayID, pageID PageID, version Version) (*Page, error) {
	mm.logger.Debug("retrieving local page", "array_id", arrayID, "page_id", pageID)

	// Check if page exists in local storage
	key := pageKey{arrayID: arrayID, pageID: pageID}
	mm.mu.RLock()
	page, exists := mm.pages[key]
	mm.mu.RUnlock()

	if !exists {
		// Create a new page
		page = NewPage(pageID, version)

		// Store it
		mm.mu.Lock()
		mm.pages[key] = page
		mm.mu.Unlock()
	}

	return page, nil
}

// requestRemotePage requests a page from a remote node
func (mm *MemoryManager) requestRemotePage(ctx context.Context, ownerID hyperbus.NodeID, arrayID ArrayID, pageID PageID, version Version) (*Page, error) {
	mm.logger.Debug("requesting remote page",
		"owner_id", ownerID,
		"array_id", arrayID,
		"page_id", pageID)

	// Create a PageRequest message
	// Send it to the owner node
	// Wait for the PageResponse
	// Decode and return the page

	// Return a new page for now
	page := NewPage(pageID, version)
	return page, nil
}

// storePage stores a page in local storage
func (mm *MemoryManager) storePage(ctx context.Context, arrayID ArrayID, pageID PageID, page *Page) error {
	key := pageKey{arrayID: arrayID, pageID: pageID}

	mm.mu.Lock()
	mm.pages[key] = page
	mm.mu.Unlock()

	mm.logger.Debug("stored page locally", "array_id", arrayID, "page_id", pageID)
	return nil
}
