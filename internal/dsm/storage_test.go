package dsm

import (
	"context"
	"log/slog"
	"testing"

	"github.com/melihxz/holocompute/internal/hyperbus"
	"github.com/melihxz/holocompute/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestPageStorage(t *testing.T) {
	// Create a page storage
	storage := newPageStorage(PageSize)

	// Test writing and reading an int64
	err := storage.setInt64(0, 42)
	assert.NoError(t, err)

	value, err := storage.getInt64(0)
	assert.NoError(t, err)
	assert.Equal(t, int64(42), value)

	// Test writing and reading a float32
	err = storage.setFloat32(8, 3.0)
	assert.NoError(t, err)

	fvalue, err := storage.getFloat32(8)
	assert.NoError(t, err)
	assert.Equal(t, float32(3.0), fvalue)

	// Test bounds checking
	err = storage.setInt64(PageSize-7, 42)
	assert.Error(t, err)

	_, err = storage.getInt64(PageSize - 7)
	assert.Error(t, err)
}

func TestPage(t *testing.T) {
	// Create a page
	page := NewPage(0, 1)

	// Test writing and reading an int64 at element index 0
	err := page.SetInt64(0, 100)
	assert.NoError(t, err)

	value, err := page.GetInt64(0)
	assert.NoError(t, err)
	assert.Equal(t, int64(100), value)

	// Test writing and reading a float32 at element index 1
	err = page.SetFloat32(1, 2.0)
	assert.NoError(t, err)

	fvalue, err := page.GetFloat32(1)
	assert.NoError(t, err)
	assert.Equal(t, float32(2.0), fvalue)
}

func TestMemoryManager(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	bus := &hyperbus.Bus{} // Mock bus

	// Create memory manager
	mm := NewMemoryManager(bus, logger)

	// Create array
	array, err := mm.CreateArray(context.Background(), 1000)
	assert.NoError(t, err)
	assert.NotNil(t, array)

	// Test getting a local page
	page, err := mm.getLocalPage(context.Background(), array.ID, 0, 1)
	assert.NoError(t, err)
	assert.NotNil(t, page)

	// Test storing and retrieving the same page
	err = mm.storePage(context.Background(), array.ID, 0, page)
	assert.NoError(t, err)

	page2, err := mm.getLocalPage(context.Background(), array.ID, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, page, page2)
}
