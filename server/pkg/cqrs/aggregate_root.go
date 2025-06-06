package cqrs

// AggregateRoot is the core interface that all domain aggregates must implement.
// This interface provides the foundation for CQRS and Event Sourcing patterns,
// ensuring consistency, versioning, and proper event handling across the system.
// Designed specifically for Defense Allies with clean, simple, and intuitive API.
//
// Key responsibilities:
//   - Identity and version management for optimistic concurrency control
//   - Event application and change tracking for event sourcing
//   - Metadata management for auditing and debugging
//   - Business rule validation and state management
//   - Soft deletion support for data retention policies
type AggregateRoot interface {
	// Identity and version management

	// ID returns the unique identifier for this aggregate instance.
	// This ID should be immutable and globally unique within the aggregate type.
	ID() string

	// Type returns the type name of this aggregate.
	// This is used for polymorphic storage and event routing.
	Type() string

	// Version returns the current version number of the aggregate.
	// This version is incremented each time the aggregate state changes.
	Version() int

	// OriginalVersion returns the version number when the aggregate was loaded from storage.
	// This is used for optimistic concurrency control to detect concurrent modifications.
	OriginalVersion() int

	// Event application and tracking

	// ApplyEvent applies a newly generated event and tracks it as an uncommitted change.
	// This method is used when executing business logic that generates new events.
	ApplyEvent(event EventMessage) error

	// ReplayEvent applies an existing event during state reconstruction.
	// This method is used when replaying events from storage to rebuild aggregate state.
	// Events applied through this method are NOT tracked as uncommitted changes.
	ReplayEvent(event EventMessage) error

	// Changes returns all uncommitted events that have been applied to this aggregate.
	// These events represent the changes that need to be persisted.
	Changes() []EventMessage

	// ClearChanges removes all uncommitted changes from the aggregate.
	// This should be called after successfully persisting the changes.
	ClearChanges()

	// Validation

	// Validate checks if the current aggregate state satisfies all business rules.
	// This method should be called before persisting the aggregate.
	Validate() error
}

// EventSourcedAggregate extends AggregateRoot with event sourcing capabilities.
// This interface is for aggregates that maintain their state by replaying events
// from an event store. It provides advanced features like snapshots and optimized
// event replay for performance in high-volume scenarios.
type EventSourcedAggregate interface {
	AggregateRoot

	// LoadFromHistory reconstructs the aggregate state by replaying events.
	// This method applies events in chronological order to rebuild the current state.
	LoadFromHistory(events []EventMessage) error

	// Snapshot support

	// CreateSnapshot generates a snapshot of the current aggregate state.
	// Snapshots are used to optimize event replay by providing a baseline state.
	CreateSnapshot() (SnapshotData, error)

	// LoadFromSnapshot restores the aggregate state from a snapshot.
	// This provides a performance optimization for event replay.
	LoadFromSnapshot(snapshot SnapshotData) error

	// ShouldCreateSnapshot determines if a new snapshot should be created.
	// This is typically based on the number of events since the last snapshot.
	ShouldCreateSnapshot() bool

	// GetLastSnapshotVersion returns the version of the last created snapshot.
	// This is used to determine which events need to be replayed.
	GetLastSnapshotVersion() int

	// CanReplayFrom checks if the aggregate can be reconstructed from a specific version.
	// This is used to validate that sufficient event history exists.
	CanReplayFrom(version int) bool
}

// StateBasedAggregate extends AggregateRoot for traditional CRUD operations.
// This interface is for aggregates that store their complete state directly
// rather than reconstructing it from events. It's simpler and more performant
// for scenarios where event history is not required.
type StateBasedAggregate interface {
	AggregateRoot

	// LoadState loads the complete aggregate state from storage.
	// This method should populate all aggregate fields from the persistent store.
	LoadState() error

	// SaveState persists the complete aggregate state to storage.
	// This method should save all aggregate fields to the persistent store.
	SaveState() error

	// HasChanged returns whether the aggregate state has been modified.
	// This is used to optimize storage operations and detect concurrent changes.
	HasChanged() bool

	// GetStateHash returns a hash of the current aggregate state.
	// This hash is used for change detection and optimistic concurrency control.
	GetStateHash() string
}

// HybridAggregate combines event sourcing and state storage capabilities.
// This interface allows aggregates to use both event sourcing for audit trails
// and state storage for performance optimization.
type HybridAggregate interface {
	EventSourcedAggregate
	StateBasedAggregate

	// SyncStateFromEvents synchronizes the state storage with the event history.
	// This ensures consistency between the two storage mechanisms.
	SyncStateFromEvents() error

	// ValidateStateConsistency checks if the state storage matches the event history.
	// This is used to detect and resolve inconsistencies between storage mechanisms.
	ValidateStateConsistency() error

	// GetStorageStrategy returns the current storage strategy being used.
	// This allows dynamic switching between event sourcing and state storage.
	GetStorageStrategy() StorageStrategy
}
