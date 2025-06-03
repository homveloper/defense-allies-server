package cqrs

import "time"

// AggregateRoot is the core interface that all domain aggregates must implement.
// This interface provides the foundation for CQRS and Event Sourcing patterns,
// ensuring consistency, versioning, and proper event handling across the system.
// The interface is designed to be compatible with the go.cqrs framework while
// providing additional functionality for Redis-based storage and hybrid approaches.
//
// Key responsibilities:
//   - Identity and version management for optimistic concurrency control
//   - Event application and change tracking for event sourcing
//   - Metadata management for auditing and debugging
//   - Business rule validation and state management
//   - Soft deletion support for data retention policies
type AggregateRoot interface {
	// Basic identification and version management

	// AggregateID returns the unique identifier for this aggregate instance.
	// This ID should be immutable and globally unique within the aggregate type.
	//
	// Returns:
	//   - string: The unique aggregate identifier (typically UUID)
	AggregateID() string

	// OriginalVersion returns the version number when the aggregate was loaded from storage.
	// This is used for optimistic concurrency control to detect concurrent modifications.
	//
	// Returns:
	//   - int: The version at load time (0 for new aggregates)
	OriginalVersion() int

	// CurrentVersion returns the current version number of the aggregate.
	// This version is incremented each time the aggregate state changes.
	//
	// Returns:
	//   - int: The current version number
	CurrentVersion() int

	// IncrementVersion increases the current version by 1.
	// This should be called whenever the aggregate state is modified.
	IncrementVersion()

	// Event application and tracking

	// Apply processes an event and updates the aggregate state accordingly.
	// This method is the core of event sourcing pattern implementation.
	//
	// Parameters:
	//   - event: The event to apply to the aggregate state
	//   - isNew: true if this is a new event being generated, false if replaying from history
	Apply(event EventMessage, isNew bool)

	// TrackChange records an event as an uncommitted change.
	// These changes will be persisted when the aggregate is saved.
	//
	// Parameters:
	//   - event: The event to track as an uncommitted change
	TrackChange(event EventMessage)

	// GetChanges returns all uncommitted events that have been applied to this aggregate.
	// These events represent the changes that need to be persisted.
	//
	// Returns:
	//   - []EventMessage: Slice of uncommitted events in chronological order
	GetChanges() []EventMessage

	// ClearChanges removes all uncommitted changes from the aggregate.
	// This should be called after successfully persisting the changes.
	ClearChanges()

	// Additional methods needed for Redis implementation

	// SetOriginalVersion sets the version that was loaded from storage.
	// This is used when loading aggregates from persistent storage.
	//
	// Parameters:
	//   - version: The version number to set as the original version
	SetOriginalVersion(version int)

	// Additional metadata

	// AggregateType returns the type name of this aggregate.
	// This is used for polymorphic storage and event routing.
	//
	// Returns:
	//   - string: The aggregate type name (e.g., "User", "Order", "Product")
	AggregateType() string

	// CreatedAt returns the timestamp when this aggregate was first created.
	//
	// Returns:
	//   - time.Time: The creation timestamp
	CreatedAt() time.Time

	// UpdatedAt returns the timestamp when this aggregate was last modified.
	//
	// Returns:
	//   - time.Time: The last modification timestamp
	UpdatedAt() time.Time

	// Validation

	// Validate checks if the current aggregate state satisfies all business rules.
	// This method should be called before persisting the aggregate.
	//
	// Returns:
	//   - error: nil if valid, error describing validation failure otherwise
	Validate() error

	// State management

	// IsDeleted returns whether this aggregate has been soft deleted.
	// Soft deleted aggregates are retained for audit purposes but marked as inactive.
	//
	// Returns:
	//   - bool: true if the aggregate is soft deleted, false otherwise
	IsDeleted() bool

	// MarkAsDeleted marks this aggregate as soft deleted.
	// This preserves the aggregate data while indicating it should not be used.
	MarkAsDeleted()
}

// EventSourcedAggregate extends AggregateRoot with event sourcing capabilities.
// This interface is for aggregates that maintain their state by replaying events
// from an event store. It provides advanced features like snapshots and optimized
// event replay for performance in high-volume scenarios.
//
// Use this interface when:
//   - You need complete audit trail of all changes
//   - Complex business logic requires event replay
//   - Temporal queries (state at specific points in time) are needed
//   - High write volume requires snapshot optimization
type EventSourcedAggregate interface {
	AggregateRoot

	// Event history management

	// LoadFromHistory reconstructs the aggregate state by replaying events.
	// This method applies events in chronological order to rebuild the current state.
	//
	// Parameters:
	//   - events: Slice of events to replay in chronological order
	//
	// Returns:
	//   - error: nil on success, error if event replay fails
	LoadFromHistory(events []EventMessage) error

	// ApplyEvent applies a single event to the aggregate state with error handling.
	// This method is specifically designed for event replay scenarios where
	// detailed error information is needed. It should update the aggregate state
	// without tracking the event as a new change.
	//
	// Parameters:
	//   - event: The event to apply to the aggregate state
	//
	// Returns:
	//   - error: nil on success, error if event application fails
	//
	// Note: This method should call Apply(event, false) internally for consistency
	ApplyEvent(event EventMessage) error

	// Snapshot support

	// CreateSnapshot generates a snapshot of the current aggregate state.
	// Snapshots are used to optimize event replay by providing a baseline state.
	//
	// Returns:
	//   - SnapshotData: The snapshot data containing the current state
	//   - error: nil on success, error if snapshot creation fails
	CreateSnapshot() (SnapshotData, error)

	// LoadFromSnapshot restores the aggregate state from a snapshot.
	// This provides a performance optimization for event replay.
	//
	// Parameters:
	//   - snapshot: The snapshot data to restore from
	//
	// Returns:
	//   - error: nil on success, error if snapshot loading fails
	LoadFromSnapshot(snapshot SnapshotData) error

	// ShouldCreateSnapshot determines if a new snapshot should be created.
	// This is typically based on the number of events since the last snapshot.
	//
	// Returns:
	//   - bool: true if a snapshot should be created, false otherwise
	ShouldCreateSnapshot() bool

	// Event replay optimization

	// GetLastSnapshotVersion returns the version of the last created snapshot.
	// This is used to determine which events need to be replayed.
	//
	// Returns:
	//   - int: The version number of the last snapshot (0 if no snapshots exist)
	GetLastSnapshotVersion() int

	// CanReplayFrom checks if the aggregate can be reconstructed from a specific version.
	// This is used to validate that sufficient event history exists.
	//
	// Parameters:
	//   - version: The version number to check replay capability from
	//
	// Returns:
	//   - bool: true if replay is possible from the specified version, false otherwise
	CanReplayFrom(version int) bool
}

// StateBasedAggregate extends AggregateRoot for traditional CRUD operations.
// This interface is for aggregates that store their complete state directly
// rather than reconstructing it from events. It's simpler and more performant
// for scenarios where event history is not required.
//
// Use this interface when:
//   - Simple CRUD operations are sufficient
//   - Event history is not needed
//   - Performance is critical and complexity should be minimized
//   - Working with legacy systems or simple data models
type StateBasedAggregate interface {
	AggregateRoot

	// Direct state load/save

	// LoadState loads the complete aggregate state from storage.
	// This method should populate all aggregate fields from the persistent store.
	//
	// Returns:
	//   - error: nil on success, error if state loading fails
	LoadState() error

	// SaveState persists the complete aggregate state to storage.
	// This method should save all aggregate fields to the persistent store.
	//
	// Returns:
	//   - error: nil on success, error if state saving fails
	SaveState() error

	// State comparison (Optimistic Concurrency Control)

	// HasChanged returns whether the aggregate state has been modified.
	// This is used to optimize storage operations and detect concurrent changes.
	//
	// Returns:
	//   - bool: true if the aggregate has been modified, false otherwise
	HasChanged() bool

	// GetStateHash returns a hash of the current aggregate state.
	// This hash is used for change detection and optimistic concurrency control.
	//
	// Returns:
	//   - string: A hash string representing the current state
	GetStateHash() string
}

// HybridAggregate combines event sourcing and state storage capabilities.
// This interface allows aggregates to use both event sourcing for audit trails
// and state storage for performance optimization. It provides the flexibility
// to choose the appropriate storage strategy based on specific requirements.
//
// Use this interface when:
//   - You need both audit trails and high performance
//   - Different operations require different storage strategies
//   - Migrating from state-based to event-sourced systems
//   - Complex scenarios requiring both approaches
type HybridAggregate interface {
	EventSourcedAggregate
	StateBasedAggregate

	// Hybrid specific functionality

	// SyncStateFromEvents synchronizes the state storage with the event history.
	// This ensures consistency between the two storage mechanisms.
	//
	// Returns:
	//   - error: nil on success, error if synchronization fails
	SyncStateFromEvents() error

	// ValidateStateConsistency checks if the state storage matches the event history.
	// This is used to detect and resolve inconsistencies between storage mechanisms.
	//
	// Returns:
	//   - error: nil if consistent, error describing the inconsistency
	ValidateStateConsistency() error

	// GetStorageStrategy returns the current storage strategy being used.
	// This allows dynamic switching between event sourcing and state storage.
	//
	// Returns:
	//   - StorageStrategy: The current storage strategy (EventSourcing, StateBased, or Hybrid)
	GetStorageStrategy() StorageStrategy
}
