package cqrs

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BaseSnapshotData provides a base implementation of SnapshotData interface
// (This was already defined in snapshot.go, but we'll extend it here)

// InMemorySnapshotStore provides an in-memory implementation of SnapshotStore
type InMemorySnapshotStore struct {
	snapshots map[string]SnapshotData
	mutex     sync.RWMutex
}

// NewInMemorySnapshotStore creates a new in-memory snapshot store
func NewInMemorySnapshotStore() *InMemorySnapshotStore {
	return &InMemorySnapshotStore{
		snapshots: make(map[string]SnapshotData),
	}
}

// SnapshotStore interface implementation

func (s *InMemorySnapshotStore) Save(ctx context.Context, snapshot SnapshotData) error {
	if snapshot == nil {
		return NewCQRSError(ErrCodeSnapshotValidationFailed.String(), "snapshot cannot be nil", nil)
	}

	if err := snapshot.Validate(); err != nil {
		return NewCQRSError(ErrCodeSnapshotValidationFailed.String(), "snapshot validation failed", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key := s.getSnapshotKey(snapshot.ID(), snapshot.Type())
	s.snapshots[key] = snapshot

	return nil
}

func (s *InMemorySnapshotStore) Load(ctx context.Context, aggregateID string) (SnapshotData, error) {
	if aggregateID == "" {
		return nil, NewCQRSError(ErrCodeInvalidAggregate.String(), "aggregate ID cannot be empty", nil)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Try to find snapshot with any aggregate type (since we only have aggregateID)
	for _, snapshot := range s.snapshots {
		if snapshot.ID() == aggregateID {
			return snapshot, nil
		}
	}

	return nil, NewCQRSError(ErrCodeSnapshotValidationFailed.String(), fmt.Sprintf("snapshot not found for aggregate: %s", aggregateID), ErrSnapshotNotFound)
}

func (s *InMemorySnapshotStore) Delete(ctx context.Context, aggregateID string) error {
	if aggregateID == "" {
		return NewCQRSError(ErrCodeInvalidAggregate.String(), "aggregate ID cannot be empty", nil)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Find and delete snapshot with matching aggregate ID
	for key, snapshot := range s.snapshots {
		if snapshot.ID() == aggregateID {
			delete(s.snapshots, key)
			return nil
		}
	}

	return NewCQRSError(ErrCodeSnapshotValidationFailed.String(), fmt.Sprintf("snapshot not found for aggregate: %s", aggregateID), ErrSnapshotNotFound)
}

func (s *InMemorySnapshotStore) Exists(ctx context.Context, aggregateID string) bool {
	if aggregateID == "" {
		return false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, snapshot := range s.snapshots {
		if snapshot.ID() == aggregateID {
			return true
		}
	}

	return false
}

// Helper methods

func (s *InMemorySnapshotStore) getSnapshotKey(aggregateID, aggregateType string) string {
	return fmt.Sprintf("%s:%s", aggregateType, aggregateID)
}

// GetSnapshotCount returns the number of stored snapshots
func (s *InMemorySnapshotStore) GetSnapshotCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.snapshots)
}

// GetAllSnapshots returns all stored snapshots (for testing/debugging)
func (s *InMemorySnapshotStore) GetAllSnapshots() map[string]SnapshotData {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	snapshots := make(map[string]SnapshotData)
	for k, v := range s.snapshots {
		snapshots[k] = v
	}
	return snapshots
}

// Clear removes all snapshots
func (s *InMemorySnapshotStore) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.snapshots = make(map[string]SnapshotData)
}

// LoadByType loads snapshot by aggregate ID and type
func (s *InMemorySnapshotStore) LoadByType(ctx context.Context, aggregateID, aggregateType string) (SnapshotData, error) {
	if aggregateID == "" {
		return nil, NewCQRSError(ErrCodeInvalidAggregate.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return nil, NewCQRSError(ErrCodeInvalidAggregate.String(), "aggregate type cannot be empty", nil)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	key := s.getSnapshotKey(aggregateID, aggregateType)
	if snapshot, exists := s.snapshots[key]; exists {
		return snapshot, nil
	}

	return nil, NewCQRSError(ErrCodeSnapshotValidationFailed.String(), fmt.Sprintf("snapshot not found for aggregate: %s:%s", aggregateType, aggregateID), ErrSnapshotNotFound)
}

// GetSnapshotsByType returns all snapshots of a specific aggregate type
func (s *InMemorySnapshotStore) GetSnapshotsByType(aggregateType string) []SnapshotData {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var snapshots []SnapshotData
	for _, snapshot := range s.snapshots {
		if snapshot.Type() == aggregateType {
			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots
}

// GetOldSnapshots returns snapshots older than the specified duration
func (s *InMemorySnapshotStore) GetOldSnapshots(olderThan time.Duration) []SnapshotData {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	cutoff := time.Now().Add(-olderThan)
	var oldSnapshots []SnapshotData

	for _, snapshot := range s.snapshots {
		if snapshot.Timestamp().Before(cutoff) {
			oldSnapshots = append(oldSnapshots, snapshot)
		}
	}

	return oldSnapshots
}
