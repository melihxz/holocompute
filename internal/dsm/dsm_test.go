package dsm

import (
	"context"
	"log/slog"
	"testing"

	"github.com/melihxz/holocompute/internal/hyperbus"
	"github.com/melihxz/holocompute/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestArray_PageCount(t *testing.T) {
	// Create an array with 100 elements
	// Assuming 8 bytes per element, that's 800 bytes
	// With 64 KiB pages, that's 1 page
	array := NewArray(100)

	// Verify page count
	assert.Equal(t, 1, array.PageCount())

	// Create an array with 10000000 elements (10M)
	// Assuming 8 bytes per element, that's 80MB
	// With 64 KiB pages, that's 1280 pages
	array2 := NewArray(10000000)

	// Verify page count
	// 10000000 * 8 = 80000000 bytes
	// 80000000 / (64 * 1024) = 1220.703125, rounded up to 1221
	assert.Equal(t, 1221, array2.PageCount())
}

func TestArray_PageOwner(t *testing.T) {
	array := NewArray(1000)

	// Set page owner
	nodeID := hyperbus.NodeID("node-1")
	array.SetPageOwner(0, nodeID)

	// Get page owner
	owner, exists := array.GetPageOwner(0)

	// Verify
	assert.True(t, exists)
	assert.Equal(t, nodeID, owner)

	// Try to get owner for non-existent page mapping
	_, exists = array.GetPageOwner(1)
	assert.False(t, exists)
}

func TestMemoryManager_CreateArray(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	bus := &hyperbus.Bus{} // Mock bus

	// Create memory manager
	mm := NewMemoryManager(bus, logger)

	// Create array
	array, err := mm.CreateArray(context.TODO(), 1000)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, array)
	assert.Equal(t, 1000, array.Length)

	// Verify array was stored
	storedArray, err := mm.GetArray(context.TODO(), array.ID)
	assert.NoError(t, err)
	assert.Equal(t, array, storedArray)
}

func TestMemoryManager_DeleteArray(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	bus := &hyperbus.Bus{} // Mock bus

	// Create memory manager
	mm := NewMemoryManager(bus, logger)

	// Create array
	array, err := mm.CreateArray(context.TODO(), 1000)
	assert.NoError(t, err)

	// Delete array
	err = mm.DeleteArray(context.TODO(), array.ID)
	assert.NoError(t, err)

	// Try to get deleted array
	_, err = mm.GetArray(context.TODO(), array.ID)
	assert.Error(t, err)

	// Try to delete non-existent array
	err = mm.DeleteArray(context.TODO(), "non-existent")
	assert.Error(t, err)
}

func TestPageCache_PutGet(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create cache with capacity 2
	cache := NewPageCache(2, logger)

	// Create a page
	page := &Page{
		ID:      0,
		Version: 1,
		Data:    make([]byte, PageSize),
	}

	// Put page in cache
	arrayID := ArrayID("array-1")
	cache.Put(arrayID, 0, page)

	// Get page from cache
	cachedPage, exists := cache.Get(arrayID, 0)

	// Verify
	assert.True(t, exists)
	assert.Equal(t, page, cachedPage)

	// Try to get non-existent page
	_, exists = cache.Get("array-2", 0)
	assert.False(t, exists)
}

func TestPageCache_Eviction(t *testing.T) {
	logger := log.New(slog.LevelDebug)

	// Create cache with capacity 2
	cache := NewPageCache(2, logger)

	// Create pages
	page1 := &Page{ID: 0, Version: 1, Data: make([]byte, PageSize)}
	page2 := &Page{ID: 1, Version: 1, Data: make([]byte, PageSize)}
	page3 := &Page{ID: 2, Version: 1, Data: make([]byte, PageSize)}

	// Put pages in cache
	arrayID := ArrayID("array-1")
	cache.Put(arrayID, 0, page1)
	cache.Put(arrayID, 1, page2)
	cache.Put(arrayID, 2, page3)

	// Verify cache size
	assert.Equal(t, 2, cache.Size())

	// The first page should have been evicted
	_, exists := cache.Get(arrayID, 0)
	assert.False(t, exists)
}
