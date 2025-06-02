package cqrs

import (
	"context"
	"time"
)

// SnapshotData interface (serialization handled separately)
type SnapshotData interface {
	AggregateID() string
	AggregateType() string
	Version() int
	Data() interface{} // Snapshot data (serialization handled separately)
	Timestamp() time.Time

	// Validation
	Validate() error
	GetChecksum() string // Integrity verification
}

// SnapshotStore interface for snapshot storage
type SnapshotStore interface {
	Save(ctx context.Context, snapshot SnapshotData) error
	Load(ctx context.Context, aggregateID string) (SnapshotData, error)
	Delete(ctx context.Context, aggregateID string) error
	Exists(ctx context.Context, aggregateID string) bool
}

// BaseSnapshotData provides a base implementation of SnapshotData interface
type BaseSnapshotData struct {
	aggregateID   string
	aggregateType string
	version       int
	data          interface{}
	timestamp     time.Time
	checksum      string
}

// NewBaseSnapshotData creates a new BaseSnapshotData
func NewBaseSnapshotData(aggregateID, aggregateType string, version int, data interface{}) *BaseSnapshotData {
	snapshot := &BaseSnapshotData{
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		version:       version,
		data:          data,
		timestamp:     time.Now(),
	}
	snapshot.calculateChecksum()
	return snapshot
}

// SnapshotData interface implementation

func (s *BaseSnapshotData) AggregateID() string {
	return s.aggregateID
}

func (s *BaseSnapshotData) AggregateType() string {
	return s.aggregateType
}

func (s *BaseSnapshotData) Version() int {
	return s.version
}

func (s *BaseSnapshotData) Data() interface{} {
	return s.data
}

func (s *BaseSnapshotData) Timestamp() time.Time {
	return s.timestamp
}

func (s *BaseSnapshotData) Validate() error {
	if s.aggregateID == "" {
		return ErrInvalidAggregateID
	}
	if s.aggregateType == "" {
		return ErrInvalidAggregateType
	}
	if s.version < 0 {
		return ErrInvalidVersion
	}
	if s.data == nil {
		return ErrInvalidSnapshotData
	}
	return nil
}

func (s *BaseSnapshotData) GetChecksum() string {
	if s.checksum == "" {
		s.calculateChecksum()
	}
	return s.checksum
}

// Helper methods

func (s *BaseSnapshotData) calculateChecksum() {
	// Simple checksum calculation - in production, use proper hashing
	s.checksum = calculateDataChecksum(s.aggregateID, s.aggregateType, s.version, s.data)
}

// SetTimestamp sets the timestamp (used when loading from storage)
func (s *BaseSnapshotData) SetTimestamp(timestamp time.Time) {
	s.timestamp = timestamp
}

// SetChecksum sets the checksum (used when loading from storage)
func (s *BaseSnapshotData) SetChecksum(checksum string) {
	s.checksum = checksum
}

// VerifyChecksum verifies the integrity of the snapshot data
func (s *BaseSnapshotData) VerifyChecksum() bool {
	expectedChecksum := calculateDataChecksum(s.aggregateID, s.aggregateType, s.version, s.data)
	return s.checksum == expectedChecksum
}

// GetSnapshotInfo returns basic snapshot information as a map
func (s *BaseSnapshotData) GetSnapshotInfo() map[string]interface{} {
	return map[string]interface{}{
		"aggregate_id":   s.aggregateID,
		"aggregate_type": s.aggregateType,
		"version":        s.version,
		"timestamp":      s.timestamp,
		"checksum":       s.checksum,
	}
}
