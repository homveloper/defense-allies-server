package cqrs

import (
	"context"
	"time"
)

// Repository interface (go.cqrs compatible)
type Repository interface {
	Save(ctx context.Context, aggregate AggregateRoot, expectedVersion int) error
	GetByID(ctx context.Context, id string) (AggregateRoot, error)
	GetVersion(ctx context.Context, id string) (int, error)
	Exists(ctx context.Context, id string) bool
}

// QueryCriteria represents query conditions for repositories
type QueryCriteria struct {
	Filters   map[string]interface{}
	SortBy    string
	SortOrder SortOrder
	Limit     int
	Offset    int
}

// StorageMetrics represents storage metrics
type StorageMetrics struct {
	EventCount    int64
	SnapshotCount int64
	StateSize     int64
	LastAccessed  time.Time
}

// EventSourcedRepository for event sourcing (optional)
type EventSourcedRepository interface {
	Repository

	// Event store related
	SaveEvents(ctx context.Context, aggregateID string, events []EventMessage, expectedVersion int) error
	GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]EventMessage, error)
	GetEventStream(ctx context.Context, aggregateID string) (<-chan EventMessage, error)

	// Snapshot related
	SaveSnapshot(ctx context.Context, snapshot SnapshotData) error
	GetSnapshot(ctx context.Context, aggregateID string) (SnapshotData, error)
	DeleteSnapshot(ctx context.Context, aggregateID string) error

	// Optimization
	GetLastEventVersion(ctx context.Context, aggregateID string) (int, error)
	CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error
}

// StateBasedRepository for traditional CRUD operations
type StateBasedRepository interface {
	Repository

	// CRUD operations
	Create(ctx context.Context, aggregate AggregateRoot) error
	Update(ctx context.Context, aggregate AggregateRoot) error
	Delete(ctx context.Context, id string) error

	// Query functionality
	FindBy(ctx context.Context, criteria QueryCriteria) ([]AggregateRoot, error)
	Count(ctx context.Context, criteria QueryCriteria) (int64, error)

	// Batch operations
	SaveBatch(ctx context.Context, aggregates []AggregateRoot) error
	DeleteBatch(ctx context.Context, ids []string) error
}

// HybridRepository combines event sourcing and state storage
type HybridRepository interface {
	EventSourcedRepository
	StateBasedRepository

	// Hybrid specific
	SyncStateFromEvents(ctx context.Context, aggregateID string) error
	ValidateConsistency(ctx context.Context, aggregateID string) error
	GetStorageMetrics(ctx context.Context, aggregateID string) (*StorageMetrics, error)
}
