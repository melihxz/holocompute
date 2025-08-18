package dsm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/melihxz/holocompute/internal/log"
)

// LeaseID uniquely identifies a lease
type LeaseID string

// LeaseType represents the type of lease
type LeaseType int

const (
	// ReadLease allows reading a page
	ReadLease LeaseType = iota
	// WriteLease allows reading and writing a page
	WriteLease
)

// Lease represents a lease on a page
type Lease struct {
	ID        LeaseID
	ArrayID   ArrayID
	PageID    PageID
	Type      LeaseType
	Owner     string // Node or client ID
	ExpiresAt time.Time
	Version   Version
}

// LeaseManager manages page leases
type LeaseManager struct {
	leases map[leaseKey]*Lease
	ttl    time.Duration
	logger *log.Logger
	mu     sync.RWMutex
}

// leaseKey uniquely identifies a leased page
type leaseKey struct {
	arrayID ArrayID
	pageID  PageID
}

// NewLeaseManager creates a new lease manager
func NewLeaseManager(ttl time.Duration, logger *log.Logger) *LeaseManager {
	return &LeaseManager{
		leases: make(map[leaseKey]*Lease),
		ttl:    ttl,
		logger: logger,
	}
}

// AcquireLease attempts to acquire a lease on a page
func (lm *LeaseManager) AcquireLease(ctx context.Context, arrayID ArrayID, pageID PageID, leaseType LeaseType, owner string, version Version) (*Lease, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	key := leaseKey{arrayID: arrayID, pageID: pageID}

	// Check if there's an existing lease
	if existingLease, exists := lm.leases[key]; exists {
		// If it's a write lease, reject all new requests
		if existingLease.Type == WriteLease {
			return nil, fmt.Errorf("write lease already exists for page %d in array %s", pageID, arrayID)
		}

		// If it's a read lease and we're requesting a write lease, reject
		if existingLease.Type == ReadLease && leaseType == WriteLease {
			return nil, fmt.Errorf("read lease exists, cannot acquire write lease for page %d in array %s", pageID, arrayID)
		}

		// If it's a read lease and we're requesting a read lease, allow (multi-reader)
		if existingLease.Type == ReadLease && leaseType == ReadLease {
			// Extend the existing lease
			existingLease.ExpiresAt = time.Now().Add(lm.ttl)
			return existingLease, nil
		}
	}

	// Create new lease
	lease := &Lease{
		ID:        LeaseID(uuid.New().String()),
		ArrayID:   arrayID,
		PageID:    pageID,
		Type:      leaseType,
		Owner:     owner,
		ExpiresAt: time.Now().Add(lm.ttl),
		Version:   version,
	}

	lm.leases[key] = lease
	lm.logger.Debug("acquired lease",
		"lease_id", lease.ID,
		"array_id", arrayID,
		"page_id", pageID,
		"type", leaseType,
		"owner", owner)

	return lease, nil
}

// ReleaseLease releases a lease
func (lm *LeaseManager) ReleaseLease(ctx context.Context, leaseID LeaseID) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Find the lease by ID
	for key, lease := range lm.leases {
		if lease.ID == leaseID {
			delete(lm.leases, key)
			lm.logger.Debug("released lease",
				"lease_id", leaseID,
				"array_id", lease.ArrayID,
				"page_id", lease.PageID)
			return nil
		}
	}

	return fmt.Errorf("lease not found: %s", leaseID)
}

// ValidateLease checks if a lease is still valid
func (lm *LeaseManager) ValidateLease(ctx context.Context, leaseID LeaseID) (*Lease, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Find the lease by ID
	for _, lease := range lm.leases {
		if lease.ID == leaseID {
			// Check if expired
			if time.Now().After(lease.ExpiresAt) {
				return nil, fmt.Errorf("lease expired: %s", leaseID)
			}
			return lease, nil
		}
	}

	return nil, fmt.Errorf("lease not found: %s", leaseID)
}

// HasWriteLease checks if there's a write lease on a page
func (lm *LeaseManager) HasWriteLease(ctx context.Context, arrayID ArrayID, pageID PageID) bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	key := leaseKey{arrayID: arrayID, pageID: pageID}
	lease, exists := lm.leases[key]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(lease.ExpiresAt) {
		return false
	}

	return lease.Type == WriteLease
}

// RevokeLease revokes a lease (e.g., when a writer commits)
func (lm *LeaseManager) RevokeLease(ctx context.Context, arrayID ArrayID, pageID PageID) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	key := leaseKey{arrayID: arrayID, pageID: pageID}
	lease, exists := lm.leases[key]
	if !exists {
		return nil // No lease to revoke
	}

	delete(lm.leases, key)
	lm.logger.Debug("revoked lease",
		"lease_id", lease.ID,
		"array_id", arrayID,
		"page_id", pageID)

	return nil
}

// CleanupExpiredLeases removes expired leases
func (lm *LeaseManager) CleanupExpiredLeases(ctx context.Context) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	now := time.Now()
	var expired []leaseKey

	for key, lease := range lm.leases {
		if now.After(lease.ExpiresAt) {
			expired = append(expired, key)
		}
	}

	for _, key := range expired {
		delete(lm.leases, key)
		lm.logger.Debug("cleaned up expired lease",
			"array_id", key.arrayID,
			"page_id", key.pageID)
	}
}
