package cqrsx

import (
	"context"
	"cqrs"
	"fmt"
)

// RedisEventSourcedRepository implements EventSourcedRepository using Redis
type RedisEventSourcedRepository struct {
	eventStore    *RedisEventStore
	snapshotStore cqrs.SnapshotStore
	aggregateType string
}

// NewRedisEventSourcedRepository creates a new Redis event sourced repository
func NewRedisEventSourcedRepository(eventStore *RedisEventStore, snapshotStore cqrs.SnapshotStore, aggregateType string) *RedisEventSourcedRepository {
	return &RedisEventSourcedRepository{
		eventStore:    eventStore,
		snapshotStore: snapshotStore,
		aggregateType: aggregateType,
	}
}

// RedisEventSourcedRepository implementation

func (r *RedisEventSourcedRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	if aggregate.Type() != r.aggregateType {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("aggregate type mismatch: expected %s, got %s", r.aggregateType, aggregate.Type()), nil)
	}

	// Get uncommitted events
	events := aggregate.Changes()
	if len(events) == 0 {
		return nil // No changes to save
	}

	// Save events
	err := r.eventStore.SaveEvents(ctx, aggregate.ID(), events, expectedVersion)
	if err != nil {
		return err
	}

	// Clear changes after successful save
	aggregate.ClearChanges()

	return nil
}

func (r *RedisEventSourcedRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	// Try to load from snapshot first
	var aggregate cqrs.AggregateRoot
	var fromVersion int = 0

	if r.snapshotStore != nil {
		snapshot, err := r.snapshotStore.Load(ctx, id)
		if err == nil && snapshot != nil {
			// Create aggregate from snapshot
			aggregate = cqrs.NewBaseAggregate(id, r.aggregateType, cqrs.WithOriginalVersion(snapshot.Version()))
			// Note: In real implementation, you'd need to restore aggregate state from snapshot
			fromVersion = snapshot.Version() + 1
		}
	}

	// If no snapshot, create new aggregate
	if aggregate == nil {
		aggregate = cqrs.NewBaseAggregate(id, r.aggregateType)
	}

	// Load events from event store
	events, err := r.eventStore.GetEventHistory(ctx, id, r.aggregateType, fromVersion)
	if err != nil {
		return nil, err
	}

	// Apply events to aggregate
	for _, event := range events {
		aggregate.ReplayEvent(event) // false = existing event, don't track as change
	}

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

func (r *RedisEventSourcedRepository) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	return r.eventStore.SaveEvents(ctx, aggregateID, events, expectedVersion)
}

func (r *RedisEventSourcedRepository) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]cqrs.EventMessage, error) {
	return r.eventStore.GetEventHistory(ctx, aggregateID, r.aggregateType, fromVersion)
}

func (r *RedisEventSourcedRepository) GetEventStream(ctx context.Context, aggregateID string) (<-chan cqrs.EventMessage, error) {
	// Note: This would require Redis Streams implementation
	return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "event streaming not implemented yet", nil)
}

func (r *RedisEventSourcedRepository) SaveSnapshot(ctx context.Context, snapshot cqrs.SnapshotData) error {
	if r.snapshotStore == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}
	return r.snapshotStore.Save(ctx, snapshot)
}

func (r *RedisEventSourcedRepository) GetSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	if r.snapshotStore == nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}
	return r.snapshotStore.Load(ctx, aggregateID)
}

func (r *RedisEventSourcedRepository) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	if r.snapshotStore == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "snapshot store not configured", nil)
	}
	return r.snapshotStore.Delete(ctx, aggregateID)
}

func (r *RedisEventSourcedRepository) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	return r.eventStore.GetLastEventVersion(ctx, aggregateID, r.aggregateType)
}

func (r *RedisEventSourcedRepository) CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error {
	return r.eventStore.CompactEvents(ctx, aggregateID, r.aggregateType, beforeVersion)
}
