package dsm

import (
	"encoding/binary"
	"fmt"
)

// pageStorage handles the actual storage of page data
type pageStorage struct {
	data []byte
}

// newPageStorage creates a new page storage with the specified size
func newPageStorage(size int) *pageStorage {
	return &pageStorage{
		data: make([]byte, size),
	}
}

// getInt64 reads a 64-bit integer from the page
func (ps *pageStorage) getInt64(offset int) (int64, error) {
	if offset < 0 || offset+8 > len(ps.data) {
		return 0, fmt.Errorf("offset out of bounds: %d", offset)
	}
	
	return int64(binary.LittleEndian.Uint64(ps.data[offset : offset+8])), nil
}

// setInt64 writes a 64-bit integer to the page
func (ps *pageStorage) setInt64(offset int, value int64) error {
	if offset < 0 || offset+8 > len(ps.data) {
		return fmt.Errorf("offset out of bounds: %d", offset)
	}
	
	binary.LittleEndian.PutUint64(ps.data[offset:offset+8], uint64(value))
	return nil
}

// getFloat32 reads a 32-bit float from the page
func (ps *pageStorage) getFloat32(offset int) (float32, error) {
	if offset < 0 || offset+4 > len(ps.data) {
		return 0, fmt.Errorf("offset out of bounds: %d", offset)
	}
	
	return float32(binary.LittleEndian.Uint32(ps.data[offset : offset+4])), nil
}

// setFloat32 writes a 32-bit float to the page
func (ps *pageStorage) setFloat32(offset int, value float32) error {
	if offset < 0 || offset+4 > len(ps.data) {
		return fmt.Errorf("offset out of bounds: %d", offset)
	}
	
	binary.LittleEndian.PutUint32(ps.data[offset:offset+4], uint32(value))
	return nil
}