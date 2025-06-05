package cqrsx

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
)

// MongoEventSourcedRepository implements EventSourcedRepository using MongoDB
// Uses standard Event Sourcing pattern with pre-designed schema
type MongoEventSourcedRepository struct {
	eventStore    *MongoEventStore
	snapshotStore *MongoSnapshotStore
	aggregateType string
}

// MongoStateBasedRepository implements StateBasedRepository using MongoDB
// Uses standard state-based storage pattern with pre-designed schema
type MongoStateBasedRepository struct {
	stateStore    *MongoStateStore
	aggregateType string
}

// MongoHybridRepository implements HybridRepository using MongoDB
// Combines Event Sourcing and state-based storage for optimal performance
type MongoHybridRepository struct {
	eventStore    *MongoEventStore
	stateStore    *MongoStateStore
	snapshotStore *MongoSnapshotStore
	aggregateType string
}

// NewMongoEventSourcedRepository creates a new MongoDB event sourced repository
// Uses standard Event Sourcing schema that developers don't need to design
func NewMongoEventSourcedRepository(client *MongoClientManager, aggregateType string) *MongoEventSourcedRepository {
	eventStore := NewMongoEventStore(client, "events")
	snapshotStore := NewMongoSnapshotStore(client, "snapshots")

	return &MongoEventSourcedRepository{
		eventStore:    eventStore,
		snapshotStore: snapshotStore,
		aggregateType: aggregateType,
	}
}

// NewMongoStateBasedRepository creates a new MongoDB state-based repository
// Uses standard state storage schema that developers don't need to design
func NewMongoStateBasedRepository(client *MongoClientManager, aggregateType string) *MongoStateBasedRepository {
	stateStore := NewMongoStateStore(client, "aggregates")

	return &MongoStateBasedRepository{
		stateStore:    stateStore,
		aggregateType: aggregateType,
	}
}

// NewMongoHybridRepository creates a new MongoDB hybrid repository
// Combines Event Sourcing and state storage for best of both worlds
func NewMongoHybridRepository(client *MongoClientManager, aggregateType string) *MongoHybridRepository {
	eventStore := NewMongoEventStore(client, "events")
	stateStore := NewMongoStateStore(client, "aggregates")
	snapshotStore := NewMongoSnapshotStore(client, "snapshots")

	return &MongoHybridRepository{
		eventStore:    eventStore,
		stateStore:    stateStore,
		snapshotStore: snapshotStore,
		aggregateType: aggregateType,
	}
}

// MongoEventSourcedRepository implementation

// Save saves an aggregate using Event Sourcing pattern
func (r *MongoEventSourcedRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	if aggregate == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate cannot be nil", nil)
	}

	// Get uncommitted changes (events)
	changes := aggregate.GetChanges()
	if len(changes) == 0 {
		return nil // No changes to save
	}

	// Save events using standard Event Sourcing pattern
	err := r.eventStore.SaveEvents(ctx, aggregate.AggregateID(), changes, expectedVersion)
	if err != nil {
		return err
	}

	// Clear changes after successful save
	aggregate.ClearChanges()

	return nil
}

// GetByID loads an aggregate by ID using Event Sourcing pattern
func (r *MongoEventSourcedRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	if id == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}

	var aggregate cqrs.AggregateRoot
	var fromVersion int = 0

	// Try to load from snapshot first (performance optimization)
	if r.snapshotStore != nil {
		snapshot, err := r.snapshotStore.LoadSnapshot(ctx, id, r.aggregateType)
		if err == nil {
			aggregate = snapshot
			fromVersion = aggregate.CurrentVersion() + 1
		} else if !cqrs.IsNotFoundError(err) {
			return nil, err
		}
	}

	// Load events from the snapshot version (or from beginning if no snapshot)
	events, err := r.eventStore.LoadEvents(ctx, id, r.aggregateType, fromVersion, 0)
	if err != nil {
		return nil, err
	}

	// If no snapshot and no events, aggregate doesn't exist
	if aggregate == nil && len(events) == 0 {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeAggregateNotFound.String(),
			fmt.Sprintf("aggregate not found: %s", id), nil)
	}

	// Create aggregate instance if no snapshot was loaded
	if aggregate == nil {
		aggregate, err = cqrs.CreateAggregateInstance(r.aggregateType, id)
		if err != nil {
			return nil, err
		}
	}

	// Apply events to rebuild/update state
	for _, event := range events {
		aggregate.Apply(event, false)
	}

	// Clear changes (these are historical events, not new changes)
	aggregate.ClearChanges()

	return aggregate, nil
}

// GetVersion gets the current version of an aggregate
func (r *MongoEventSourcedRepository) GetVersion(ctx context.Context, id string) (int, error) {
	if id == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}

	return r.eventStore.GetLastEventVersion(ctx, id, r.aggregateType)
}

// Exists checks if an aggregate exists
func (r *MongoEventSourcedRepository) Exists(ctx context.Context, id string) bool {
	version, err := r.GetVersion(ctx, id)
	return err == nil && version > 0
}

// SaveSnapshot saves a snapshot of the aggregate (performance optimization)
func (r *MongoEventSourcedRepository) SaveSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	if r.snapshotStore == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}

	return r.snapshotStore.SaveSnapshot(ctx, aggregate)
}

// MongoStateBasedRepository implementation

// Save saves an aggregate using state-based storage
func (r *MongoStateBasedRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	if aggregate == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate cannot be nil", nil)
	}

	return r.stateStore.SaveAggregate(ctx, aggregate, expectedVersion)
}

// GetByID loads an aggregate by ID using state-based storage
func (r *MongoStateBasedRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	if id == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}

	return r.stateStore.LoadAggregate(ctx, id, r.aggregateType)
}

// GetVersion gets the current version of an aggregate
func (r *MongoStateBasedRepository) GetVersion(ctx context.Context, id string) (int, error) {
	if id == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}

	return r.stateStore.GetAggregateVersion(ctx, id, r.aggregateType)
}

// Exists checks if an aggregate exists
func (r *MongoStateBasedRepository) Exists(ctx context.Context, id string) bool {
	version, err := r.GetVersion(ctx, id)
	return err == nil && version >= 0
}

// MongoHybridRepository implementation

// Save saves an aggregate using hybrid storage (both events and state)
func (r *MongoHybridRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	if aggregate == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate cannot be nil", nil)
	}

	// Save events first (for audit trail and event sourcing)
	changes := aggregate.GetChanges()
	if len(changes) > 0 {
		err := r.eventStore.SaveEvents(ctx, aggregate.AggregateID(), changes, expectedVersion)
		if err != nil {
			return err
		}
	}

	// Save current state (for fast retrieval)
	err := r.stateStore.SaveAggregate(ctx, aggregate, expectedVersion)
	if err != nil {
		return err
	}

	// Clear changes after successful save
	aggregate.ClearChanges()

	return nil
}

// GetByID loads an aggregate by ID using hybrid storage (prefer state, fallback to events)
func (r *MongoHybridRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	if id == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}

	// Try to load from state store first (faster)
	aggregate, err := r.stateStore.LoadAggregate(ctx, id, r.aggregateType)
	if err == nil {
		return aggregate, nil
	}

	// If not found in state store, try to rebuild from events
	if cqrs.IsNotFoundError(err) {
		return r.rebuildFromEvents(ctx, id)
	}

	return nil, err
}

// rebuildFromEvents rebuilds aggregate from events (fallback for hybrid repository)
func (r *MongoHybridRepository) rebuildFromEvents(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	var aggregate cqrs.AggregateRoot
	var fromVersion int = 0

	// Try to load from snapshot first
	if r.snapshotStore != nil {
		snapshot, err := r.snapshotStore.LoadSnapshot(ctx, id, r.aggregateType)
		if err == nil {
			aggregate = snapshot
			fromVersion = aggregate.CurrentVersion() + 1
		} else if !cqrs.IsNotFoundError(err) {
			return nil, err
		}
	}

	// Load events
	events, err := r.eventStore.LoadEvents(ctx, id, r.aggregateType, fromVersion, 0)
	if err != nil {
		return nil, err
	}

	// If no snapshot and no events, aggregate doesn't exist
	if aggregate == nil && len(events) == 0 {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeAggregateNotFound.String(),
			fmt.Sprintf("aggregate not found: %s", id), nil)
	}

	// Create aggregate instance if no snapshot was loaded
	if aggregate == nil {
		aggregate, err = cqrs.CreateAggregateInstance(r.aggregateType, id)
		if err != nil {
			return nil, err
		}
	}

	// Apply events to rebuild state
	for _, event := range events {
		aggregate.Apply(event, false)
	}

	// Clear changes
	aggregate.ClearChanges()

	return aggregate, nil
}

// GetVersion gets the current version of an aggregate
func (r *MongoHybridRepository) GetVersion(ctx context.Context, id string) (int, error) {
	if id == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "aggregate ID cannot be empty", nil)
	}

	// Try state store first
	version, err := r.stateStore.GetAggregateVersion(ctx, id, r.aggregateType)
	if err == nil {
		return version, nil
	}

	// Fallback to event store
	if cqrs.IsNotFoundError(err) {
		return r.eventStore.GetLastEventVersion(ctx, id, r.aggregateType)
	}

	return -1, err
}

// Exists checks if an aggregate exists
func (r *MongoHybridRepository) Exists(ctx context.Context, id string) bool {
	// Check state store first
	if r.stateStore.AggregateExists(ctx, id, r.aggregateType) {
		return true
	}

	// Check event store
	version, err := r.eventStore.GetLastEventVersion(ctx, id, r.aggregateType)
	return err == nil && version > 0
}

// SaveSnapshot saves a snapshot of the aggregate (performance optimization)
func (r *MongoHybridRepository) SaveSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	if r.snapshotStore == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}

	return r.snapshotStore.SaveSnapshot(ctx, aggregate)
}
