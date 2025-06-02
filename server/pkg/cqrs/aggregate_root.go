package cqrs

import "time"

// AggregateRoot is the interface that all aggregates should implement
// This interface is compatible with go.cqrs framework
type AggregateRoot interface {
	// Basic identification and version management
	AggregateID() string
	OriginalVersion() int // Version at load time
	CurrentVersion() int  // Current version
	IncrementVersion()    // Increment version

	// Event application and tracking
	Apply(event EventMessage, isNew bool) // Apply event
	TrackChange(event EventMessage)       // Track changes
	GetChanges() []EventMessage           // Get uncommitted changes
	ClearChanges()                        // Clear changes

	// Additional methods needed for Redis implementation
	SetOriginalVersion(version int) // Set original version (for loading)

	// Additional metadata
	AggregateType() string // Aggregate type identification
	CreatedAt() time.Time  // Creation time
	UpdatedAt() time.Time  // Last update time

	// Validation
	Validate() error // Business rule validation

	// State management
	IsDeleted() bool // Check if deleted
	MarkAsDeleted()  // Mark as soft deleted
}

// EventSourcedAggregate supports event sourcing (optional)
type EventSourcedAggregate interface {
	AggregateRoot

	// Event history management
	LoadFromHistory(events []EventMessage) error
	ApplyEvent(event EventMessage) error

	// Snapshot support
	CreateSnapshot() (SnapshotData, error)
	LoadFromSnapshot(snapshot SnapshotData) error
	ShouldCreateSnapshot() bool // Check snapshot creation condition

	// Event replay optimization
	GetLastSnapshotVersion() int
	CanReplayFrom(version int) bool
}

// StateBasedAggregate for traditional CRUD operations
type StateBasedAggregate interface {
	AggregateRoot

	// Direct state load/save
	LoadState() error
	SaveState() error

	// State comparison (Optimistic Concurrency Control)
	HasChanged() bool
	GetStateHash() string // For change detection
}

// HybridAggregate combines event sourcing and state storage
type HybridAggregate interface {
	EventSourcedAggregate
	StateBasedAggregate

	// Hybrid specific functionality
	SyncStateFromEvents() error      // Sync state from events
	ValidateStateConsistency() error // Validate state consistency
	GetStorageStrategy() StorageStrategy
}
