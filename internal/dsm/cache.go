package dsm

import (
	"container/list"
	"sync"

	"github.com/melihxz/holocompute/internal/log"
)

// PageCache implements a 2Q cache for pages
type PageCache struct {
	capacity int
	cache    map[cacheKey]*list.Element
	// Two queues for 2Q algorithm
	freqList *list.List // Frequently accessed pages
	onceList *list.List // Pages accessed once
	logger   *log.Logger
	mu       sync.RWMutex
}

// cacheKey uniquely identifies a cached page
type cacheKey struct {
	arrayID ArrayID
	pageID  PageID
}

// cacheEntry holds a cached page
type cacheEntry struct {
	key      cacheKey
	page     *Page
	fromFreq bool // Whether this entry is from the frequent list
}

// NewPageCache creates a new page cache with the specified capacity
func NewPageCache(capacity int, logger *log.Logger) *PageCache {
	return &PageCache{
		capacity: capacity,
		cache:    make(map[cacheKey]*list.Element),
		freqList: list.New(),
		onceList: list.New(),
		logger:   logger,
	}
}

// Get retrieves a page from the cache
func (pc *PageCache) Get(arrayID ArrayID, pageID PageID) (*Page, bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	key := cacheKey{arrayID: arrayID, pageID: pageID}
	element, exists := pc.cache[key]
	if !exists {
		return nil, false
	}

	entry := element.Value.(*cacheEntry)

	// Move to frequent list if it's in the once list
	if !entry.fromFreq {
		// Remove from once list
		pc.onceList.Remove(element)

		// Add to frequent list
		entry.fromFreq = true
		element = pc.freqList.PushFront(entry)
		pc.cache[key] = element
	} else {
		// Move to front of frequent list
		pc.freqList.MoveToFront(element)
	}

	return entry.page, true
}

// Put adds a page to the cache
func (pc *PageCache) Put(arrayID ArrayID, pageID PageID, page *Page) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	key := cacheKey{arrayID: arrayID, pageID: pageID}

	// If already in cache, update it
	if element, exists := pc.cache[key]; exists {
		entry := element.Value.(*cacheEntry)
		entry.page = page

		// Move to frequent list if not already there
		if !entry.fromFreq {
			// Remove from once list
			pc.onceList.Remove(element)

			// Add to frequent list
			entry.fromFreq = true
			element = pc.freqList.PushFront(entry)
			pc.cache[key] = element
		} else {
			// Move to front of frequent list
			pc.freqList.MoveToFront(element)
		}
		return
	}

	// Add new entry to once list
	entry := &cacheEntry{
		key:      key,
		page:     page,
		fromFreq: false,
	}
	element := pc.onceList.PushFront(entry)
	pc.cache[key] = element

	// Evict if necessary
	if len(pc.cache) > pc.capacity {
		pc.evict()
	}
}

// evict removes the least recently used page from the cache
func (pc *PageCache) evict() {
	// First try to evict from once list
	if pc.onceList.Len() > 0 {
		element := pc.onceList.Back()
		if element != nil {
			entry := pc.onceList.Remove(element).(*cacheEntry)
			delete(pc.cache, entry.key)
			return
		}
	}

	// If once list is empty, evict from freq list
	if pc.freqList.Len() > 0 {
		element := pc.freqList.Back()
		if element != nil {
			entry := pc.freqList.Remove(element).(*cacheEntry)
			delete(pc.cache, entry.key)
			return
		}
	}
}

// Remove removes a page from the cache
func (pc *PageCache) Remove(arrayID ArrayID, pageID PageID) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	key := cacheKey{arrayID: arrayID, pageID: pageID}
	element, exists := pc.cache[key]
	if !exists {
		return
	}

	// Remove from whichever list it's in
	entry := element.Value.(*cacheEntry)
	if entry.fromFreq {
		pc.freqList.Remove(element)
	} else {
		pc.onceList.Remove(element)
	}

	delete(pc.cache, key)
}

// Size returns the current size of the cache
func (pc *PageCache) Size() int {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return len(pc.cache)
}

// Capacity returns the maximum capacity of the cache
func (pc *PageCache) Capacity() int {
	return pc.capacity
}
