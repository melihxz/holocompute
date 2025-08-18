package dsm

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/melihxz/holocompute/internal/log"
	"github.com/stretchr/testify/assert"
)

func TestLeaseManager_AcquireReadLease(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a read lease
	lease, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-1", 1)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, lease)
	assert.Equal(t, ReadLease, lease.Type)
	assert.Equal(t, "client-1", lease.Owner)

	// Acquire another read lease on the same page (should succeed)
	lease2, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-2", 1)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, lease2)
}

func TestLeaseManager_AcquireWriteLease(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a write lease
	lease, err := lm.AcquireLease(context.Background(), "array-1", 0, WriteLease, "client-1", 1)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, lease)
	assert.Equal(t, WriteLease, lease.Type)
	assert.Equal(t, "client-1", lease.Owner)
}

func TestLeaseManager_WriteLeaseBlocksRead(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a write lease
	_, err := lm.AcquireLease(context.Background(), "array-1", 0, WriteLease, "client-1", 1)
	assert.NoError(t, err)

	// Try to acquire a read lease on the same page (should fail)
	_, err = lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-2", 1)
	assert.Error(t, err)
}

func TestLeaseManager_ReadLeaseBlocksWrite(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a read lease
	_, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-1", 1)
	assert.NoError(t, err)

	// Try to acquire a write lease on the same page (should fail)
	_, err = lm.AcquireLease(context.Background(), "array-1", 0, WriteLease, "client-2", 1)
	assert.Error(t, err)
}

func TestLeaseManager_ReleaseLease(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a lease
	lease, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-1", 1)
	assert.NoError(t, err)

	// Release the lease
	err = lm.ReleaseLease(context.Background(), lease.ID)
	assert.NoError(t, err)

	// Try to validate the released lease (should fail)
	_, err = lm.ValidateLease(context.Background(), lease.ID)
	assert.Error(t, err)
}

func TestLeaseManager_ValidateLease(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a lease
	lease, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-1", 1)
	assert.NoError(t, err)

	// Validate the lease (should succeed)
	validLease, err := lm.ValidateLease(context.Background(), lease.ID)
	assert.NoError(t, err)
	assert.Equal(t, lease, validLease)
}

func TestLeaseManager_ExpiredLease(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Millisecond, logger) // Very short TTL

	// Acquire a lease
	lease, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-1", 1)
	assert.NoError(t, err)

	// Wait for lease to expire
	time.Sleep(time.Millisecond * 30)

	// Validate the expired lease (should fail)
	_, err = lm.ValidateLease(context.Background(), lease.ID)
	assert.Error(t, err)
}

func TestLeaseManager_RevokeLease(t *testing.T) {
	logger := log.New(slog.LevelDebug)
	lm := NewLeaseManager(time.Minute, logger)

	// Acquire a lease
	_, err := lm.AcquireLease(context.Background(), "array-1", 0, ReadLease, "client-1", 1)
	assert.NoError(t, err)

	// Revoke the lease
	err = lm.RevokeLease(context.Background(), "array-1", 0)
	assert.NoError(t, err)

	// Try to acquire a write lease (should succeed now)
	_, err = lm.AcquireLease(context.Background(), "array-1", 0, WriteLease, "client-2", 1)
	assert.NoError(t, err)
}
