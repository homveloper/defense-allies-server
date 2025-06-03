package cqrs

import (
	"context"
	"time"
)

// SnapshotData interface defines the contract for aggregate state snapshots in event sourcing.
// Snapshots are point-in-time captures of aggregate state that optimize event replay
// performance by providing a baseline state. They contain the complete aggregate state
// at a specific version, allowing event replay to start from that point instead of
// replaying all events from the beginning.
//
// Key responsibilities:
//   - Capture complete aggregate state at a specific version
//   - Provide metadata for snapshot management and validation
//   - Support integrity verification through checksums
//   - Enable efficient serialization and storage
//
// Implementation guidelines:
//   - Snapshots should be immutable once created
//   - Include all necessary state to reconstruct the aggregate
//   - Implement proper validation to ensure data integrity
//   - Use checksums to detect corruption
//   - Handle serialization separately for flexibility
type SnapshotData interface {
	// AggregateID returns the unique identifier of the aggregate this snapshot represents.
	// This links the snapshot to its corresponding aggregate instance.
	//
	// Returns:
	//   - string: Unique aggregate identifier
	AggregateID() string

	// AggregateType returns the type of aggregate this snapshot represents.
	// This is used for polymorphic storage and type-safe reconstruction.
	//
	// Returns:
	//   - string: Aggregate type name (e.g., "User", "Order", "Product")
	AggregateType() string

	// Version returns the aggregate version at which this snapshot was taken.
	// This determines which events need to be replayed after loading the snapshot.
	//
	// Returns:
	//   - int: Aggregate version number at snapshot time
	Version() int

	// Data returns the actual aggregate state data captured in this snapshot.
	// This contains all the information needed to reconstruct the aggregate state.
	// Serialization is handled separately to allow different storage formats.
	//
	// Returns:
	//   - interface{}: Aggregate state data (should be serializable)
	Data() interface{}

	// Timestamp returns when this snapshot was created.
	// This is used for snapshot management, cleanup, and debugging.
	//
	// Returns:
	//   - time.Time: Snapshot creation timestamp
	Timestamp() time.Time

	// Validation methods

	// Validate checks if the snapshot data is valid and complete.
	// This method should verify all required fields and business rules.
	//
	// Returns:
	//   - error: nil if valid, descriptive error if validation fails
	//
	// Validation checks:
	//   - AggregateID is not empty
	//   - AggregateType is not empty
	//   - Version is non-negative
	//   - Data is not nil and contains valid state
	Validate() error

	// GetChecksum returns a checksum for integrity verification.
	// This is used to detect data corruption during storage or transmission.
	//
	// Returns:
	//   - string: Checksum string for integrity verification
	//
	// Usage: Compare checksums before and after storage operations
	GetChecksum() string
}

// SnapshotStore interface defines the contract for snapshot persistence.
// This interface abstracts the underlying storage mechanism for snapshots,
// allowing different implementations (Redis, SQL, file system, etc.).
// It provides the basic CRUD operations needed for snapshot management.
//
// Key responsibilities:
//   - Persist snapshots with proper error handling
//   - Retrieve snapshots by aggregate identifier
//   - Support snapshot cleanup and management
//   - Ensure data integrity during storage operations
//
// Implementation guidelines:
//   - Ensure thread safety for concurrent operations
//   - Implement proper error handling with meaningful messages
//   - Support atomic operations where possible
//   - Handle serialization/deserialization transparently
//   - Provide consistent behavior across different storage backends
type SnapshotStore interface {
	// Save persists a snapshot to the storage backend.
	// This method stores the snapshot data and ensures it can be retrieved later.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - snapshot: The snapshot data to save (must be non-nil and valid)
	//
	// Returns:
	//   - error: nil on success, error on validation failure or storage error
	//
	// Error conditions:
	//   - snapshot is nil: Returns validation error
	//   - snapshot validation fails: Returns validation error with details
	//   - serialization fails: Returns serialization error
	//   - storage failure: Returns storage error with underlying cause
	//
	// Behavior: Overwrites existing snapshots for the same aggregate
	Save(ctx context.Context, snapshot SnapshotData) error

	// Load retrieves a snapshot by aggregate identifier.
	// This method loads the most recent snapshot for the specified aggregate.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - SnapshotData: The loaded snapshot data
	//   - error: nil on success, error if not found or loading fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - snapshot not found: Returns not found error
	//   - deserialization fails: Returns serialization error
	//   - storage failure: Returns storage error with underlying cause
	Load(ctx context.Context, aggregateID string) (SnapshotData, error)

	// Delete removes a snapshot from the storage backend.
	// This method is used for cleanup or when snapshots become invalid.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - error: nil on success, error if deletion fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - storage failure: Returns storage error with underlying cause
	//
	// Note: Deleting a non-existent snapshot is not considered an error
	Delete(ctx context.Context, aggregateID string) error

	// Exists checks if a snapshot exists for the given aggregate.
	// This method provides a lightweight way to verify snapshot existence.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - bool: true if snapshot exists, false otherwise
	//
	// Note: This method should not return errors for "not found" cases,
	// only for actual system failures. Missing snapshots return false.
	Exists(ctx context.Context, aggregateID string) bool
}

// BaseSnapshotData provides a concrete implementation of the SnapshotData interface.
// This implementation handles the common functionality needed for snapshot management,
// including automatic checksum calculation, validation, and metadata management.
// It serves as a foundation for domain-specific snapshot implementations.
//
// Key features:
//   - Automatic checksum calculation for integrity verification
//   - Built-in validation for required fields
//   - Immutable design for thread safety
//   - Timestamp management for snapshot lifecycle
//
// Usage:
//   - Use directly for simple snapshot scenarios
//   - Embed in domain-specific snapshot types for extended functionality
//   - Suitable for most event sourcing use cases
type BaseSnapshotData struct {
	aggregateID   string      // Unique identifier of the aggregate
	aggregateType string      // Type of the aggregate (e.g., "User", "Order")
	version       int         // Aggregate version at snapshot time
	data          interface{} // Actual aggregate state data
	timestamp     time.Time   // When the snapshot was created
	checksum      string      // Integrity verification checksum
}

// NewBaseSnapshotData creates and initializes a new BaseSnapshotData instance.
// This constructor automatically sets the timestamp and calculates the checksum
// for integrity verification. The resulting snapshot is immutable and ready for storage.
//
// Parameters:
//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
//   - aggregateType: Type of the aggregate (must be non-empty)
//   - version: Aggregate version at snapshot time (must be non-negative)
//   - data: Aggregate state data to capture (must be non-nil and serializable)
//
// Returns:
//   - *BaseSnapshotData: A new snapshot instance ready for use
//
// Usage:
//
//	snapshot := NewBaseSnapshotData("user-123", "User", 5, userState)
//	err := snapshotStore.Save(ctx, snapshot)
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

// AggregateID returns the unique identifier of the aggregate this snapshot represents.
func (s *BaseSnapshotData) AggregateID() string {
	return s.aggregateID
}

// AggregateType returns the type of aggregate this snapshot represents.
func (s *BaseSnapshotData) AggregateType() string {
	return s.aggregateType
}

// Version returns the aggregate version at which this snapshot was taken.
func (s *BaseSnapshotData) Version() int {
	return s.version
}

// Data returns the actual aggregate state data captured in this snapshot.
func (s *BaseSnapshotData) Data() interface{} {
	return s.data
}

// Timestamp returns when this snapshot was created.
func (s *BaseSnapshotData) Timestamp() time.Time {
	return s.timestamp
}

// Validate checks if the snapshot data is valid and complete.
// This method performs comprehensive validation of all required fields
// and ensures the snapshot is in a consistent state.
//
// Returns:
//   - error: nil if valid, specific error describing the validation failure
//
// Validation rules:
//   - AggregateID must not be empty
//   - AggregateType must not be empty
//   - Version must be non-negative
//   - Data must not be nil
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

// GetChecksum returns the checksum for integrity verification.
// If the checksum hasn't been calculated yet, it calculates it automatically.
// This ensures the checksum is always available when needed.
//
// Returns:
//   - string: Checksum string for integrity verification
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
