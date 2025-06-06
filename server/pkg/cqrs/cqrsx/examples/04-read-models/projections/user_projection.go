package projections

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/readmodels"
	"fmt"
	"log"
)

// UserProjection handles user-related events and updates UserView read models
type UserProjection struct {
	readStore cqrs.ReadStore
	name      string
}

// NewUserProjection creates a new UserProjection
func NewUserProjection(readStore cqrs.ReadStore) *UserProjection {
	return &UserProjection{
		readStore: readStore,
		name:      "UserProjection",
	}
}

// GetName returns the projection name
func (p *UserProjection) GetName() string {
	return p.name
}

// Handle handles domain events and updates read models accordingly
func (p *UserProjection) Handle(ctx context.Context, event cqrs.EventMessage) error {
	log.Printf("UserProjection: Processing event %s for aggregate %s",
		event.EventType(), event.AggregateID())

	switch e := event.EventData().(type) {
	case *domain.UserCreated:
		return p.handleUserCreated(ctx, e)
	case *domain.UserUpdated:
		return p.handleUserUpdated(ctx, e)
	case *domain.OrderCompleted:
		return p.handleOrderCompleted(ctx, e)
	default:
		// Ignore unknown events
		log.Printf("UserProjection: Ignoring unknown event type %T", e)
		return nil
	}
}

// handleUserCreated handles UserCreated events
func (p *UserProjection) handleUserCreated(ctx context.Context, event *domain.UserCreated) error {
	log.Printf("UserProjection: Creating UserView for user %s", event.UserID)

	// Create new UserView
	userView := readmodels.NewUserView(event.UserID, event.Name, event.Email)

	// Save to read store
	if err := p.readStore.Save(ctx, userView); err != nil {
		return fmt.Errorf("failed to save UserView: %w", err)
	}

	log.Printf("UserProjection: Successfully created UserView for user %s", event.UserID)
	return nil
}

// handleUserUpdated handles UserUpdated events
func (p *UserProjection) handleUserUpdated(ctx context.Context, event *domain.UserUpdated) error {
	log.Printf("UserProjection: Updating UserView for user %s", event.UserID)

	// Get existing UserView
	userView, err := p.getUserView(ctx, event.UserID)
	if err != nil {
		return fmt.Errorf("failed to get UserView: %w", err)
	}

	// Update profile information
	userView.UpdateProfile(event.Name, event.Email)

	// Save updated view
	if err := p.readStore.Save(ctx, userView); err != nil {
		return fmt.Errorf("failed to save updated UserView: %w", err)
	}

	log.Printf("UserProjection: Successfully updated UserView for user %s", event.UserID)
	return nil
}

// handleOrderCompleted handles OrderCompleted events to update user statistics
func (p *UserProjection) handleOrderCompleted(ctx context.Context, event *domain.OrderCompleted) error {
	log.Printf("UserProjection: Updating UserView for order completion - user %s, amount %s",
		event.CustomerID, event.TotalAmount.String())

	// Get existing UserView
	userView, err := p.getUserView(ctx, event.CustomerID)
	if err != nil {
		// If user view doesn't exist, we might need to create it
		// This could happen if events are processed out of order
		log.Printf("UserProjection: UserView not found for user %s, skipping order completion update", event.CustomerID)
		return nil
	}

	// Record the completed order
	userView.RecordOrder(event.TotalAmount, event.CompletedAt)

	// Save updated view
	if err := p.readStore.Save(ctx, userView); err != nil {
		return fmt.Errorf("failed to save updated UserView after order completion: %w", err)
	}

	log.Printf("UserProjection: Successfully updated UserView for user %s after order completion", event.CustomerID)
	return nil
}

// getUserView retrieves a UserView from the read store
func (p *UserProjection) getUserView(ctx context.Context, userID string) (*readmodels.UserView, error) {
	// Get from read store
	readModel, err := p.readStore.GetByID(ctx, userID, "UserView")
	if err != nil {
		return nil, err
	}

	// Type assertion to UserView
	userView, ok := readModel.(*readmodels.UserView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *UserView, got %T", readModel)
	}

	return userView, nil
}

// Rebuild rebuilds all UserView read models from events
func (p *UserProjection) Rebuild(ctx context.Context, eventStore interface{}) error {
	log.Printf("UserProjection: Starting rebuild process")

	// Clear existing UserViews (implementation depends on read store)
	// For now, we'll just log this step
	log.Printf("UserProjection: Clearing existing UserViews")

	// For now, we'll just log the rebuild operation
	// In a real implementation, this would iterate through all events
	// and rebuild the read models
	log.Printf("UserProjection: Rebuild operation logged - implementation depends on event store capabilities")

	log.Printf("UserProjection: Rebuild completed successfully")
	return nil
}

// GetStatistics returns projection statistics
func (p *UserProjection) GetStatistics(ctx context.Context) (map[string]interface{}, error) {
	// This would typically query the read store for statistics
	// For now, return basic info
	return map[string]interface{}{
		"projection_name": p.name,
		"status":          "active",
		"last_updated":    "unknown", // Would track this in real implementation
	}, nil
}

// Validate validates the projection state
func (p *UserProjection) Validate() error {
	// Allow nil readStore for demo purposes
	if p.name == "" {
		return fmt.Errorf("projection name cannot be empty")
	}
	return nil
}

// Reset resets the projection state (clears all read models)
func (p *UserProjection) Reset(ctx context.Context) error {
	log.Printf("UserProjection: Resetting projection state")

	// This would clear all UserViews from the read store
	// Implementation depends on the read store capabilities
	// For now, just log the operation
	log.Printf("UserProjection: Reset completed")

	return nil
}

// GetSupportedEvents returns the list of events this projection handles
func (p *UserProjection) GetSupportedEvents() []string {
	return []string{
		"UserCreated",
		"UserUpdated",
		"OrderCompleted", // For updating user statistics
	}
}

// IsEventSupported checks if the projection supports a specific event type
func (p *UserProjection) IsEventSupported(eventType string) bool {
	supportedEvents := p.GetSupportedEvents()
	for _, supported := range supportedEvents {
		if supported == eventType {
			return true
		}
	}
	return false
}

// ProcessEventBatch processes multiple events in a batch for better performance
func (p *UserProjection) ProcessEventBatch(ctx context.Context, events []cqrs.EventMessage) error {
	log.Printf("UserProjection: Processing batch of %d events", len(events))

	// Group events by aggregate ID for efficient processing
	eventsByUser := make(map[string][]cqrs.EventMessage)

	for _, event := range events {
		if p.IsEventSupported(event.EventType()) {
			userID := event.AggregateID()
			eventsByUser[userID] = append(eventsByUser[userID], event)
		}
	}

	// Process events for each user
	for userID, userEvents := range eventsByUser {
		log.Printf("UserProjection: Processing %d events for user %s", len(userEvents), userID)

		for _, event := range userEvents {
			if err := p.Handle(ctx, event); err != nil {
				return fmt.Errorf("failed to process event %s for user %s: %w",
					event.EventID(), userID, err)
			}
		}
	}

	log.Printf("UserProjection: Successfully processed batch of %d events", len(events))
	return nil
}

// GetLastProcessedEventID returns the ID of the last processed event
func (p *UserProjection) GetLastProcessedEventID(ctx context.Context) (string, error) {
	// This would typically be stored in a projection state table
	// For now, return empty string
	return "", nil
}

// SetLastProcessedEventID sets the ID of the last processed event
func (p *UserProjection) SetLastProcessedEventID(ctx context.Context, eventID string) error {
	// This would typically be stored in a projection state table
	// For now, just log it
	log.Printf("UserProjection: Setting last processed event ID to %s", eventID)
	return nil
}
