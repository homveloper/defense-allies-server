package cqrs

import (
	"context"
	"time"
)

// Repository interface defines the core contract for aggregate persistence in CQRS systems.
// This interface is compatible with the go.cqrs framework and provides the fundamental
// operations needed for aggregate lifecycle management. It abstracts the underlying
// storage mechanism, allowing different implementations (in-memory, Redis, SQL, etc.).
//
// Key responsibilities:
//   - Aggregate persistence with optimistic concurrency control
//   - Aggregate retrieval by unique identifier
//   - Version management for conflict detection
//   - Existence checking for validation
//
// Implementation guidelines:
//   - Ensure thread safety for concurrent operations
//   - Implement proper error handling with meaningful messages
//   - Support optimistic concurrency control through version checking
//   - Provide consistent behavior across different storage backends
//   - Handle aggregate serialization/deserialization transparently
type Repository interface {
	// Save persists an aggregate to the repository with optimistic concurrency control.
	// This method stores the aggregate state and increments its version to prevent
	// concurrent modification conflicts.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregate: The aggregate to save (must be non-nil and valid)
	//   - expectedVersion: Expected current version for concurrency control
	//
	// Returns:
	//   - error: nil on success, error on validation failure, version conflict, or storage error
	//
	// Error conditions:
	//   - aggregate is nil: Returns validation error
	//   - aggregate validation fails: Returns validation error with details
	//   - version conflict: Returns concurrency error (expectedVersion != currentVersion)
	//   - storage failure: Returns repository error with underlying cause
	//
	// Concurrency: This method implements optimistic locking through version checking
	Save(ctx context.Context, aggregate AggregateRoot, expectedVersion int) error

	// GetByID retrieves an aggregate by its unique identifier.
	// This method loads the complete aggregate state from storage and reconstructs
	// the aggregate object with proper type casting.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - id: Unique identifier of the aggregate to retrieve (must be non-empty)
	//
	// Returns:
	//   - AggregateRoot: The retrieved aggregate with current state
	//   - error: nil on success, error if not found or retrieval fails
	//
	// Error conditions:
	//   - id is empty: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - deserialization fails: Returns serialization error
	//   - storage failure: Returns repository error with underlying cause
	GetByID(ctx context.Context, id string) (AggregateRoot, error)

	// GetVersion retrieves the current version of an aggregate without loading its full state.
	// This method is optimized for version checking and concurrency control scenarios.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - id: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - int: Current version number of the aggregate
	//   - error: nil on success, error if not found or retrieval fails
	//
	// Error conditions:
	//   - id is empty: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Performance: This method should be optimized for minimal data transfer
	GetVersion(ctx context.Context, id string) (int, error)

	// Exists checks if an aggregate with the given ID exists in the repository.
	// This method provides a lightweight way to verify aggregate existence without
	// loading the full aggregate state.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - id: Unique identifier of the aggregate to check (must be non-empty)
	//
	// Returns:
	//   - bool: true if aggregate exists, false otherwise
	//
	// Note: This method should not return errors for "not found" cases,
	// only for actual system failures. Missing aggregates return false.
	//
	// Performance: This method should be optimized for minimal resource usage
	Exists(ctx context.Context, id string) bool
}

// QueryCriteria represents the search and filtering conditions for repository queries.
// This structure provides a flexible way to specify complex query requirements
// including filtering, sorting, and pagination. It's designed to be storage-agnostic
// and can be translated to different query languages (SQL, NoSQL, etc.).
//
// Usage patterns:
//   - Use Filters for field-based filtering with various operators
//   - Combine SortBy and SortOrder for result ordering
//   - Use Limit and Offset for pagination
//   - Chain multiple criteria for complex queries
type QueryCriteria struct {
	Filters   map[string]interface{} `json:"filters"`    // Field-value pairs for filtering (supports operators)
	SortBy    string                 `json:"sort_by"`    // Field name to sort by (empty for no sorting)
	SortOrder SortOrder              `json:"sort_order"` // Sort direction (Ascending or Descending)
	Limit     int                    `json:"limit"`      // Maximum number of results (0 for no limit)
	Offset    int                    `json:"offset"`     // Number of results to skip (for pagination)
}

// StorageMetrics represents performance and usage metrics for aggregate storage.
// This structure provides insights into storage efficiency, access patterns,
// and resource utilization. It's useful for monitoring, optimization, and
// capacity planning in production environments.
//
// Metrics categories:
//   - Volume metrics: EventCount, SnapshotCount, StateSize
//   - Access metrics: LastAccessed
//   - Performance metrics: (can be extended with timing data)
type StorageMetrics struct {
	EventCount    int64     `json:"event_count"`    // Total number of events stored for the aggregate
	SnapshotCount int64     `json:"snapshot_count"` // Number of snapshots created for the aggregate
	StateSize     int64     `json:"state_size"`     // Size of stored state in bytes
	LastAccessed  time.Time `json:"last_accessed"`  // Timestamp of last access (read or write)
}

// EventSourcedRepository extends Repository with event sourcing capabilities.
// This interface provides advanced functionality for event-driven aggregate persistence,
// including event storage, snapshot management, and performance optimizations.
// It's designed for scenarios requiring complete audit trails and temporal queries.
//
// Key features:
//   - Event stream storage and retrieval
//   - Snapshot management for performance optimization
//   - Event compaction for storage efficiency
//   - Streaming support for real-time processing
//
// Use cases:
//   - Financial systems requiring complete audit trails
//   - Complex business processes with temporal requirements
//   - Systems needing event replay capabilities
//   - High-volume scenarios requiring snapshot optimization
type EventSourcedRepository interface {
	Repository

	// Event store related methods

	// SaveEvents persists a batch of events for an aggregate with version control.
	// This method ensures atomicity of event storage and maintains event ordering.
	// It's the core method for event sourcing persistence.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//   - events: Slice of events to save in chronological order (must be non-empty)
	//   - expectedVersion: Expected current version for optimistic concurrency control
	//
	// Returns:
	//   - error: nil on success, error on validation failure, version conflict, or storage error
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - events slice is empty: Returns validation error
	//   - version conflict: Returns concurrency error
	//   - event validation fails: Returns validation error with details
	//   - storage failure: Returns repository error with underlying cause
	//
	// Atomicity: All events are saved atomically or none are saved
	SaveEvents(ctx context.Context, aggregateID string, events []EventMessage, expectedVersion int) error

	// GetEventHistory retrieves events for an aggregate starting from a specific version.
	// This method is used for aggregate reconstruction and event replay scenarios.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//   - fromVersion: Starting version number (0 for all events, >0 for partial history)
	//
	// Returns:
	//   - []EventMessage: Events in chronological order starting from fromVersion
	//   - error: nil on success, error if aggregate not found or retrieval fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - fromVersion is negative: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Performance: Consider using snapshots for aggregates with many events
	GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]EventMessage, error)

	// GetEventStream returns a channel for streaming events of an aggregate.
	// This method enables real-time event processing and reactive patterns.
	// The channel is closed when all events have been sent or an error occurs.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - <-chan EventMessage: Read-only channel for receiving events in chronological order
	//   - error: nil on success, error if aggregate not found or streaming setup fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - streaming setup fails: Returns repository error with underlying cause
	//
	// Usage: Always check for channel closure and handle context cancellation
	GetEventStream(ctx context.Context, aggregateID string) (<-chan EventMessage, error)

	// Snapshot related methods

	// SaveSnapshot persists a snapshot of aggregate state for performance optimization.
	// Snapshots reduce the number of events needed for aggregate reconstruction.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - snapshot: Snapshot data containing aggregate state and metadata
	//
	// Returns:
	//   - error: nil on success, error on validation failure or storage error
	//
	// Error conditions:
	//   - snapshot is invalid: Returns validation error
	//   - snapshot data is corrupted: Returns serialization error
	//   - storage failure: Returns repository error with underlying cause
	SaveSnapshot(ctx context.Context, snapshot SnapshotData) error

	// GetSnapshot retrieves the latest snapshot for an aggregate.
	// This method is used to optimize aggregate reconstruction by providing a baseline state.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - SnapshotData: Latest snapshot data for the aggregate
	//   - error: nil on success, error if not found or retrieval fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - no snapshot exists: Returns not found error
	//   - snapshot is corrupted: Returns serialization error
	//   - storage failure: Returns repository error with underlying cause
	GetSnapshot(ctx context.Context, aggregateID string) (SnapshotData, error)

	// DeleteSnapshot removes the snapshot for an aggregate.
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
	//   - storage failure: Returns repository error with underlying cause
	//
	// Note: Deleting a non-existent snapshot is not considered an error
	DeleteSnapshot(ctx context.Context, aggregateID string) error

	// Optimization methods

	// GetLastEventVersion retrieves the version number of the latest event for an aggregate.
	// This method is optimized for version checking without loading event data.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - int: Version number of the latest event (0 if no events exist)
	//   - error: nil on success, error if retrieval fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Performance: This method should be optimized for minimal data transfer
	GetLastEventVersion(ctx context.Context, aggregateID string) (int, error)

	// CompactEvents removes old events before a specific version to save storage space.
	// This method is used for storage optimization while preserving recent event history.
	// Events are typically compacted after creating snapshots.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//   - beforeVersion: Version number before which events should be removed
	//
	// Returns:
	//   - error: nil on success, error if compaction fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - beforeVersion is invalid: Returns validation error
	//   - no snapshot exists for beforeVersion: Returns validation error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Warning: Ensure snapshots exist before compacting events to avoid data loss
	CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error
}

// StateBasedRepository extends Repository with traditional CRUD operations.
// This interface provides familiar database-style operations for aggregates,
// making it suitable for scenarios where event history is not required.
// It focuses on current state storage and efficient querying capabilities.
//
// Key features:
//   - Traditional CRUD operations (Create, Read, Update, Delete)
//   - Advanced querying with filtering, sorting, and pagination
//   - Batch operations for high-throughput scenarios
//   - Optimized for current state access patterns
//
// Use cases:
//   - Simple business applications without audit requirements
//   - Legacy system integration
//   - Performance-critical scenarios where event replay overhead is unacceptable
//   - Reporting and analytics workloads
type StateBasedRepository interface {
	Repository

	// CRUD operations

	// Create persists a new aggregate to the repository.
	// This method is specifically for creating new aggregates and will fail if the aggregate already exists.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregate: The new aggregate to create (must be non-nil and valid)
	//
	// Returns:
	//   - error: nil on success, error on validation failure, duplicate key, or storage error
	//
	// Error conditions:
	//   - aggregate is nil: Returns validation error
	//   - aggregate already exists: Returns duplicate key error
	//   - aggregate validation fails: Returns validation error with details
	//   - storage failure: Returns repository error with underlying cause
	Create(ctx context.Context, aggregate AggregateRoot) error

	// Update modifies an existing aggregate in the repository.
	// This method uses optimistic concurrency control through version checking.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregate: The aggregate with updated state (must be non-nil and valid)
	//
	// Returns:
	//   - error: nil on success, error on validation failure, version conflict, or storage error
	//
	// Error conditions:
	//   - aggregate is nil: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - version conflict: Returns concurrency error
	//   - aggregate validation fails: Returns validation error with details
	//   - storage failure: Returns repository error with underlying cause
	Update(ctx context.Context, aggregate AggregateRoot) error

	// Delete removes an aggregate from the repository by its identifier.
	// This method performs a hard delete, permanently removing the aggregate.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - id: Unique identifier of the aggregate to delete (must be non-empty)
	//
	// Returns:
	//   - error: nil on success, error if deletion fails
	//
	// Error conditions:
	//   - id is empty: Returns validation error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Note: Deleting a non-existent aggregate is not considered an error
	Delete(ctx context.Context, id string) error

	// Query functionality

	// FindBy retrieves aggregates matching the specified criteria.
	// This method supports complex filtering, sorting, and pagination.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - criteria: Query criteria including filters, sorting, and pagination
	//
	// Returns:
	//   - []AggregateRoot: Aggregates matching the criteria in the specified order
	//   - error: nil on success, error if query execution fails
	//
	// Error conditions:
	//   - invalid criteria: Returns validation error
	//   - query execution fails: Returns repository error with underlying cause
	//
	// Performance: Use appropriate indexes for filter fields to ensure good performance
	FindBy(ctx context.Context, criteria QueryCriteria) ([]AggregateRoot, error)

	// Count returns the number of aggregates matching the specified criteria.
	// This method is optimized for counting without loading aggregate data.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - criteria: Query criteria for filtering (sorting and pagination are ignored)
	//
	// Returns:
	//   - int64: Number of aggregates matching the criteria
	//   - error: nil on success, error if count operation fails
	//
	// Error conditions:
	//   - invalid criteria: Returns validation error
	//   - count operation fails: Returns repository error with underlying cause
	//
	// Performance: This method should be optimized for minimal resource usage
	Count(ctx context.Context, criteria QueryCriteria) (int64, error)

	// Batch operations

	// SaveBatch persists multiple aggregates in a single operation.
	// This method provides better performance for bulk operations through batching.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregates: Slice of aggregates to save (can be mix of new and existing)
	//
	// Returns:
	//   - error: nil on success, error if any aggregate fails validation or storage
	//
	// Error conditions:
	//   - any aggregate is nil: Returns validation error
	//   - any aggregate validation fails: Returns validation error with details
	//   - version conflict on any aggregate: Returns concurrency error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Atomicity: Implementation may choose to make this operation atomic or not
	SaveBatch(ctx context.Context, aggregates []AggregateRoot) error

	// DeleteBatch removes multiple aggregates in a single operation.
	// This method provides better performance for bulk deletions through batching.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - ids: Slice of aggregate identifiers to delete (must be non-empty strings)
	//
	// Returns:
	//   - error: nil on success, error if deletion fails
	//
	// Error conditions:
	//   - any id is empty: Returns validation error
	//   - storage failure: Returns repository error with underlying cause
	//
	// Note: Non-existent aggregates are ignored and not considered errors
	DeleteBatch(ctx context.Context, ids []string) error
}

// HybridRepository combines event sourcing and state storage capabilities.
// This interface provides the flexibility to use both storage strategies within
// the same repository, allowing optimization for different access patterns
// and requirements. It's ideal for complex systems with varying needs.
//
// Key features:
//   - Full event sourcing capabilities for audit and replay
//   - State-based storage for performance-critical operations
//   - Synchronization between event and state storage
//   - Consistency validation across storage mechanisms
//   - Comprehensive metrics for monitoring and optimization
//
// Use cases:
//   - Systems transitioning from state-based to event-sourced architecture
//   - Applications requiring both audit trails and high-performance queries
//   - Complex domains with mixed storage requirements
//   - Systems needing flexible storage strategy selection
type HybridRepository interface {
	EventSourcedRepository
	StateBasedRepository

	// Hybrid specific methods

	// SyncStateFromEvents synchronizes the state storage with the event history.
	// This method rebuilds the current state by replaying events from the event store.
	// It's used to ensure consistency between the two storage mechanisms.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate to synchronize (must be non-empty)
	//
	// Returns:
	//   - error: nil on success, error if synchronization fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - aggregate not found in event store: Returns not found error
	//   - event replay fails: Returns event sourcing error
	//   - state update fails: Returns repository error with underlying cause
	//
	// Performance: This operation can be expensive for aggregates with many events
	SyncStateFromEvents(ctx context.Context, aggregateID string) error

	// ValidateConsistency checks if the state storage matches the event history.
	// This method compares the current state with the state derived from event replay
	// to detect inconsistencies between storage mechanisms.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate to validate (must be non-empty)
	//
	// Returns:
	//   - error: nil if consistent, error describing the inconsistency or validation failure
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - state and events are inconsistent: Returns consistency error with details
	//   - validation process fails: Returns repository error with underlying cause
	//
	// Usage: Run this periodically to detect and resolve storage inconsistencies
	ValidateConsistency(ctx context.Context, aggregateID string) error

	// GetStorageMetrics retrieves comprehensive metrics for an aggregate's storage usage.
	// This method provides insights into storage efficiency, access patterns, and
	// performance characteristics across both storage mechanisms.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeout, and request-scoped values
	//   - aggregateID: Unique identifier of the aggregate (must be non-empty)
	//
	// Returns:
	//   - *StorageMetrics: Comprehensive metrics including event count, snapshot count, state size, etc.
	//   - error: nil on success, error if metrics collection fails
	//
	// Error conditions:
	//   - aggregateID is empty: Returns validation error
	//   - aggregate not found: Returns not found error
	//   - metrics collection fails: Returns repository error with underlying cause
	//
	// Usage: Use for monitoring, capacity planning, and storage optimization decisions
	GetStorageMetrics(ctx context.Context, aggregateID string) (*StorageMetrics, error)
}
