package repositories

import (
	"context"
	"fmt"

	"defense-allies-server/examples/guild/domain"
	"defense-allies-server/pkg/cqrs"
)

// InMemoryGuildRepository is a simple in-memory repository for the guild example
type InMemoryGuildRepository struct {
	guilds      map[string]*domain.GuildAggregate
	events      map[string][]cqrs.EventMessage // aggregateID -> events
	projections []cqrs.Projection
}

// NewInMemoryGuildRepository creates a new InMemoryGuildRepository
func NewInMemoryGuildRepository(projections []cqrs.Projection) *InMemoryGuildRepository {
	return &InMemoryGuildRepository{
		guilds:      make(map[string]*domain.GuildAggregate),
		events:      make(map[string][]cqrs.EventMessage),
		projections: projections,
	}
}

// Save saves an aggregate to the repository
func (r *InMemoryGuildRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	// Store the aggregate
	guild, ok := aggregate.(*domain.GuildAggregate)
	if !ok {
		return fmt.Errorf("invalid aggregate type: expected *GuildAggregate, got %T", aggregate)
	}

	// Create a copy to avoid reference issues
	guildCopy := *guild
	r.guilds[aggregate.ID()] = &guildCopy

	// Get uncommitted events
	events := aggregate.GetChanges()
	fmt.Printf("   🔧 Saving aggregate %s with %d events\n", aggregate.ID(), len(events))

	if len(events) > 0 {
		// Store events for history
		if r.events[aggregate.ID()] == nil {
			r.events[aggregate.ID()] = make([]cqrs.EventMessage, 0)
		}
		r.events[aggregate.ID()] = append(r.events[aggregate.ID()], events...)
		fmt.Printf("   🔧 Total events for %s: %d\n", aggregate.ID(), len(r.events[aggregate.ID()]))

		// Process events through projections
		for _, event := range events {
			for _, projection := range r.projections {
				if projection.CanHandle(event.EventType()) {
					if err := projection.Project(ctx, event); err != nil {
						return fmt.Errorf("failed to process event %s through projection %s: %w",
							event.EventType(), projection.GetProjectionName(), err)
					}
				}
			}
		}
	}

	// Clear changes after successful save
	aggregate.ClearChanges()

	return nil
}

// Load loads an aggregate by ID
func (r *InMemoryGuildRepository) Load(ctx context.Context, aggregateID string) (cqrs.AggregateRoot, error) {
	fmt.Printf("   🔧 Loading aggregate %s\n", aggregateID)
	events, exists := r.events[aggregateID]
	if !exists {
		fmt.Printf("   🔧 No events found for aggregate %s\n", aggregateID)
		return nil, fmt.Errorf("aggregate %s not found", aggregateID)
	}

	fmt.Printf("   🔧 Found %d events for aggregate %s\n", len(events), aggregateID)
	guild, err := domain.LoadGuildAggregate(aggregateID, events)
	if err != nil {
		return nil, fmt.Errorf("failed to load guild aggregate: %w", err)
	}

	return guild, nil
}

// GetByID gets an aggregate by ID (alias for Load)
func (r *InMemoryGuildRepository) GetByID(ctx context.Context, aggregateID string) (cqrs.AggregateRoot, error) {
	return r.Load(ctx, aggregateID)
}

// GetEventHistory returns the event history for an aggregate
func (r *InMemoryGuildRepository) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]cqrs.EventMessage, error) {
	events, exists := r.events[aggregateID]
	if !exists {
		return nil, fmt.Errorf("aggregate %s not found", aggregateID)
	}

	// Filter events from the specified version
	var filteredEvents []cqrs.EventMessage
	for _, event := range events {
		if event.Version() >= fromVersion {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// Exists checks if an aggregate exists
func (r *InMemoryGuildRepository) Exists(ctx context.Context, aggregateID string) bool {
	_, exists := r.events[aggregateID]
	return exists
}

// GetVersion returns the current version of an aggregate
func (r *InMemoryGuildRepository) GetVersion(ctx context.Context, aggregateID string) (int, error) {
	events, exists := r.events[aggregateID]
	if !exists {
		return 0, fmt.Errorf("aggregate %s not found", aggregateID)
	}

	if len(events) == 0 {
		return 0, nil
	}

	// Return the version of the last event
	return events[len(events)-1].Version(), nil
}

// EventSourcedRepository interface implementation

// SaveEvents saves events for an aggregate
func (r *InMemoryGuildRepository) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	// Check version for optimistic concurrency control
	if existing, exists := r.guilds[aggregateID]; exists {
		if existing.Version() != expectedVersion {
			return fmt.Errorf("version conflict: expected %d, got %d", expectedVersion, existing.Version())
		}
	}

	// Store events for history
	if len(events) > 0 {
		if r.events[aggregateID] == nil {
			r.events[aggregateID] = make([]cqrs.EventMessage, 0)
		}
		r.events[aggregateID] = append(r.events[aggregateID], events...)
	}

	return nil
}

// GetEventStream gets an event stream (not implemented for this example)
func (r *InMemoryGuildRepository) GetEventStream(ctx context.Context, aggregateID string) (<-chan cqrs.EventMessage, error) {
	return nil, fmt.Errorf("event streaming not implemented in this example")
}

// GetLastEventVersion returns the last event version for an aggregate
func (r *InMemoryGuildRepository) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	events, exists := r.events[aggregateID]
	if !exists || len(events) == 0 {
		return 0, nil
	}

	// Return the version of the last event
	lastEvent := events[len(events)-1]
	return lastEvent.Version(), nil
}

// LoadEvents loads events for an aggregate
func (r *InMemoryGuildRepository) LoadEvents(ctx context.Context, aggregateID string, fromVersion, toVersion int) ([]cqrs.EventMessage, error) {
	events, exists := r.events[aggregateID]
	if !exists {
		return []cqrs.EventMessage{}, nil
	}

	// Apply version filtering if needed
	if fromVersion > 0 || toVersion >= 0 {
		filteredEvents := make([]cqrs.EventMessage, 0)
		for _, event := range events {
			version := event.Version()
			if fromVersion > 0 && version < fromVersion {
				continue
			}
			if toVersion >= 0 && version > toVersion {
				break
			}
			filteredEvents = append(filteredEvents, event)
		}
		return filteredEvents, nil
	}

	return events, nil
}

// GetEventCount returns the number of events for an aggregate
func (r *InMemoryGuildRepository) GetEventCount(ctx context.Context, aggregateID string) (int, error) {
	events, exists := r.events[aggregateID]
	if !exists {
		return 0, nil
	}
	return len(events), nil
}

// SaveSnapshot saves a snapshot (not implemented for this example)
func (r *InMemoryGuildRepository) SaveSnapshot(ctx context.Context, snapshot cqrs.SnapshotData) error {
	return fmt.Errorf("snapshots not implemented in this example")
}

// LoadSnapshot loads a snapshot (not implemented for this example)
func (r *InMemoryGuildRepository) LoadSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	return nil, fmt.Errorf("snapshots not implemented in this example")
}

// GetSnapshot gets a snapshot (alias for LoadSnapshot)
func (r *InMemoryGuildRepository) GetSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	return r.LoadSnapshot(ctx, aggregateID)
}

// DeleteSnapshot deletes a snapshot (not implemented for this example)
func (r *InMemoryGuildRepository) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	return fmt.Errorf("snapshots not implemented in this example")
}

// CompactEvents compacts events (not implemented for this example)
func (r *InMemoryGuildRepository) CompactEvents(ctx context.Context, aggregateID string, toVersion int) error {
	return fmt.Errorf("event compaction not implemented in this example")
}
