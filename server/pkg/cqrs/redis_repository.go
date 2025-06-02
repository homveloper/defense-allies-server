package cqrs

import (
	"context"
	"fmt"
)

// RedisEventSourcedRepository implements EventSourcedRepository using Redis
type RedisEventSourcedRepository struct {
	eventStore    *RedisEventStore
	snapshotStore SnapshotStore
	aggregateType string
}

// RedisStateBasedRepository implements StateBasedRepository using Redis
type RedisStateBasedRepository struct {
	stateStore    *RedisStateStore
	aggregateType string
}

// RedisHybridRepository implements HybridRepository using Redis
type RedisHybridRepository struct {
	eventStore    *RedisEventStore
	stateStore    *RedisStateStore
	snapshotStore SnapshotStore
	aggregateType string
}

// NewRedisEventSourcedRepository creates a new Redis event sourced repository
func NewRedisEventSourcedRepository(eventStore *RedisEventStore, snapshotStore SnapshotStore, aggregateType string) *RedisEventSourcedRepository {
	return &RedisEventSourcedRepository{
		eventStore:    eventStore,
		snapshotStore: snapshotStore,
		aggregateType: aggregateType,
	}
}

// NewRedisStateBasedRepository creates a new Redis state based repository
func NewRedisStateBasedRepository(stateStore *RedisStateStore, aggregateType string) *RedisStateBasedRepository {
	return &RedisStateBasedRepository{
		stateStore:    stateStore,
		aggregateType: aggregateType,
	}
}

// NewRedisHybridRepository creates a new Redis hybrid repository
func NewRedisHybridRepository(eventStore *RedisEventStore, stateStore *RedisStateStore, snapshotStore SnapshotStore, aggregateType string) *RedisHybridRepository {
	return &RedisHybridRepository{
		eventStore:    eventStore,
		stateStore:    stateStore,
		snapshotStore: snapshotStore,
		aggregateType: aggregateType,
	}
}

// RedisEventSourcedRepository implementation

func (r *RedisEventSourcedRepository) Save(ctx context.Context, aggregate AggregateRoot, expectedVersion int) error {
	if aggregate.AggregateType() != r.aggregateType {
		return NewCQRSError(ErrCodeRepositoryError.String(),
			fmt.Sprintf("aggregate type mismatch: expected %s, got %s", r.aggregateType, aggregate.AggregateType()), nil)
	}

	// Get uncommitted events
	events := aggregate.GetChanges()
	if len(events) == 0 {
		return nil // No changes to save
	}

	// Save events
	err := r.eventStore.SaveEvents(ctx, aggregate.AggregateID(), events, expectedVersion)
	if err != nil {
		return err
	}

	// Clear changes after successful save
	aggregate.ClearChanges()

	return nil
}

func (r *RedisEventSourcedRepository) GetByID(ctx context.Context, id string) (AggregateRoot, error) {
	// Try to load from snapshot first
	var aggregate AggregateRoot
	var fromVersion int = 0

	if r.snapshotStore != nil {
		snapshot, err := r.snapshotStore.Load(ctx, id)
		if err == nil && snapshot != nil {
			// Create aggregate from snapshot
			aggregate = NewBaseAggregate(id, r.aggregateType)
			// Note: In real implementation, you'd need to restore aggregate state from snapshot
			fromVersion = snapshot.Version() + 1
		}
	}

	// If no snapshot, create new aggregate
	if aggregate == nil {
		aggregate = NewBaseAggregate(id, r.aggregateType)
	}

	// Load events from event store
	events, err := r.eventStore.GetEventHistory(ctx, id, r.aggregateType, fromVersion)
	if err != nil {
		return nil, err
	}

	// Apply events to aggregate
	for _, event := range events {
		aggregate.Apply(event, false) // false = existing event, don't track as change
	}

	// Set original version
	aggregate.SetOriginalVersion(aggregate.CurrentVersion())

	return aggregate, nil
}

func (r *RedisEventSourcedRepository) GetVersion(ctx context.Context, id string) (int, error) {
	return r.eventStore.GetLastEventVersion(ctx, id, r.aggregateType)
}

func (r *RedisEventSourcedRepository) Exists(ctx context.Context, id string) bool {
	version, err := r.GetVersion(ctx, id)
	return err == nil && version > 0
}

// EventSourcedRepository specific methods

func (r *RedisEventSourcedRepository) SaveEvents(ctx context.Context, aggregateID string, events []EventMessage, expectedVersion int) error {
	return r.eventStore.SaveEvents(ctx, aggregateID, events, expectedVersion)
}

func (r *RedisEventSourcedRepository) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]EventMessage, error) {
	return r.eventStore.GetEventHistory(ctx, aggregateID, r.aggregateType, fromVersion)
}

func (r *RedisEventSourcedRepository) GetEventStream(ctx context.Context, aggregateID string) (<-chan EventMessage, error) {
	// Note: This would require Redis Streams implementation
	return nil, NewCQRSError(ErrCodeRepositoryError.String(), "event streaming not implemented yet", nil)
}

func (r *RedisEventSourcedRepository) SaveSnapshot(ctx context.Context, snapshot SnapshotData) error {
	if r.snapshotStore == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}
	return r.snapshotStore.Save(ctx, snapshot)
}

func (r *RedisEventSourcedRepository) GetSnapshot(ctx context.Context, aggregateID string) (SnapshotData, error) {
	if r.snapshotStore == nil {
		return nil, NewCQRSError(ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}
	return r.snapshotStore.Load(ctx, aggregateID)
}

func (r *RedisEventSourcedRepository) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	if r.snapshotStore == nil {
		return NewCQRSError(ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}
	return r.snapshotStore.Delete(ctx, aggregateID)
}

func (r *RedisEventSourcedRepository) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	return r.eventStore.GetLastEventVersion(ctx, aggregateID, r.aggregateType)
}

func (r *RedisEventSourcedRepository) CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error {
	return r.eventStore.CompactEvents(ctx, aggregateID, r.aggregateType, beforeVersion)
}

// RedisStateBasedRepository implementation

func (r *RedisStateBasedRepository) Save(ctx context.Context, aggregate AggregateRoot, expectedVersion int) error {
	return r.stateStore.Save(ctx, aggregate, expectedVersion)
}

func (r *RedisStateBasedRepository) GetByID(ctx context.Context, id string) (AggregateRoot, error) {
	return r.stateStore.GetByID(ctx, r.aggregateType, id)
}

func (r *RedisStateBasedRepository) GetVersion(ctx context.Context, id string) (int, error) {
	return r.stateStore.GetVersion(ctx, r.aggregateType, id)
}

func (r *RedisStateBasedRepository) Exists(ctx context.Context, id string) bool {
	return r.stateStore.Exists(ctx, r.aggregateType, id)
}

// StateBasedRepository specific methods

func (r *RedisStateBasedRepository) Create(ctx context.Context, aggregate AggregateRoot) error {
	return r.Save(ctx, aggregate, 0) // Expected version 0 for new aggregates
}

func (r *RedisStateBasedRepository) Update(ctx context.Context, aggregate AggregateRoot) error {
	return r.Save(ctx, aggregate, aggregate.OriginalVersion())
}

func (r *RedisStateBasedRepository) Delete(ctx context.Context, id string) error {
	return r.stateStore.Delete(ctx, r.aggregateType, id)
}

func (r *RedisStateBasedRepository) FindBy(ctx context.Context, criteria QueryCriteria) ([]AggregateRoot, error) {
	// Note: This would require implementing query functionality in Redis
	return nil, NewCQRSError(ErrCodeRepositoryError.String(), "query functionality not implemented yet", nil)
}

func (r *RedisStateBasedRepository) Count(ctx context.Context, criteria QueryCriteria) (int64, error) {
	// Note: This would require implementing query functionality in Redis
	return 0, NewCQRSError(ErrCodeRepositoryError.String(), "query functionality not implemented yet", nil)
}

func (r *RedisStateBasedRepository) SaveBatch(ctx context.Context, aggregates []AggregateRoot) error {
	// Note: This would require implementing batch operations
	return NewCQRSError(ErrCodeRepositoryError.String(), "batch operations not implemented yet", nil)
}

func (r *RedisStateBasedRepository) DeleteBatch(ctx context.Context, ids []string) error {
	// Note: This would require implementing batch operations
	return NewCQRSError(ErrCodeRepositoryError.String(), "batch operations not implemented yet", nil)
}
