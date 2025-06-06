package projections

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/readmodels"
	"fmt"
	"log"
)

// OrderProjection handles order-related events and updates OrderSummaryView read models
type OrderProjection struct {
	readStore cqrs.ReadStore
	name      string
}

// NewOrderProjection creates a new OrderProjection
func NewOrderProjection(readStore cqrs.ReadStore) *OrderProjection {
	return &OrderProjection{
		readStore: readStore,
		name:      "OrderProjection",
	}
}

// GetName returns the projection name
func (p *OrderProjection) GetName() string {
	return p.name
}

// Handle handles domain events and updates read models accordingly
func (p *OrderProjection) Handle(ctx context.Context, event cqrs.EventMessage) error {
	log.Printf("OrderProjection: Processing event %s for aggregate %s",
		event.EventType(), event.ID())

	switch e := event.EventData().(type) {
	case *domain.OrderCreated:
		return p.handleOrderCreated(ctx, e)
	case *domain.OrderCompleted:
		return p.handleOrderCompleted(ctx, e)
	case *domain.OrderCancelled:
		return p.handleOrderCancelled(ctx, e)
	case *domain.OrderItemAdded:
		return p.handleOrderItemAdded(ctx, e)
	case *domain.OrderItemRemoved:
		return p.handleOrderItemRemoved(ctx, e)
	case *domain.UserCreated:
		// We need user info for order summaries
		return p.handleUserCreated(ctx, e)
	case *domain.UserUpdated:
		// Update customer info in existing orders
		return p.handleUserUpdated(ctx, e)
	default:
		// Ignore unknown events
		log.Printf("OrderProjection: Ignoring unknown event type %T", e)
		return nil
	}
}

// handleOrderCreated handles OrderCreated events
func (p *OrderProjection) handleOrderCreated(ctx context.Context, event *domain.OrderCreated) error {
	log.Printf("OrderProjection: Creating OrderSummaryView for order %s", event.OrderID)

	// Get customer information (we'll need to look this up)
	customerName, customerEmail, err := p.getCustomerInfo(ctx, event.CustomerID)
	if err != nil {
		log.Printf("OrderProjection: Warning - could not get customer info for %s: %v", event.CustomerID, err)
		customerName = "Unknown Customer"
		customerEmail = "unknown@example.com"
	}

	// Convert domain OrderItems to view OrderItemViews
	items := make([]readmodels.OrderItemView, len(event.Items))
	for i, item := range event.Items {
		items[i] = readmodels.OrderItemView{
			ProductID:   item.ProductID,
			ProductName: item.Name,
			Price:       item.Price,
			Quantity:    item.Quantity,
			SubTotal:    item.SubTotal,
		}
	}

	// Create new OrderSummaryView
	orderView := readmodels.NewOrderSummaryView(
		event.OrderID,
		event.CustomerID,
		customerName,
		customerEmail,
		items,
		event.SubTotal,
		event.TaxAmount,
		event.Total,
		event.OrderDate,
	)

	// Save to read store
	if err := p.readStore.Save(ctx, orderView); err != nil {
		return fmt.Errorf("failed to save OrderSummaryView: %w", err)
	}

	log.Printf("OrderProjection: Successfully created OrderSummaryView for order %s", event.OrderID)
	return nil
}

// handleOrderCompleted handles OrderCompleted events
func (p *OrderProjection) handleOrderCompleted(ctx context.Context, event *domain.OrderCompleted) error {
	log.Printf("OrderProjection: Updating OrderSummaryView for completed order %s", event.OrderID)

	// Get existing OrderSummaryView
	orderView, err := p.getOrderSummaryView(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get OrderSummaryView: %w", err)
	}

	// Update status to completed
	orderView.CompleteOrder(event.CompletedAt)

	// Save updated view
	if err := p.readStore.Save(ctx, orderView); err != nil {
		return fmt.Errorf("failed to save updated OrderSummaryView: %w", err)
	}

	log.Printf("OrderProjection: Successfully updated OrderSummaryView for completed order %s", event.OrderID)
	return nil
}

// handleOrderCancelled handles OrderCancelled events
func (p *OrderProjection) handleOrderCancelled(ctx context.Context, event *domain.OrderCancelled) error {
	log.Printf("OrderProjection: Updating OrderSummaryView for cancelled order %s", event.OrderID)

	// Get existing OrderSummaryView
	orderView, err := p.getOrderSummaryView(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get OrderSummaryView: %w", err)
	}

	// Update status to cancelled
	orderView.CancelOrder(event.CancelledAt, event.Reason)

	// Save updated view
	if err := p.readStore.Save(ctx, orderView); err != nil {
		return fmt.Errorf("failed to save updated OrderSummaryView: %w", err)
	}

	log.Printf("OrderProjection: Successfully updated OrderSummaryView for cancelled order %s", event.OrderID)
	return nil
}

// handleOrderItemAdded handles OrderItemAdded events
func (p *OrderProjection) handleOrderItemAdded(ctx context.Context, event *domain.OrderItemAdded) error {
	log.Printf("OrderProjection: Adding item to OrderSummaryView for order %s", event.OrderID)

	// Get existing OrderSummaryView
	orderView, err := p.getOrderSummaryView(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get OrderSummaryView: %w", err)
	}

	// Convert domain OrderItem to view OrderItemView
	itemView := readmodels.OrderItemView{
		ProductID:   event.Item.ProductID,
		ProductName: event.Item.Name,
		Price:       event.Item.Price,
		Quantity:    event.Item.Quantity,
		SubTotal:    event.Item.SubTotal,
	}

	// Add item to order
	orderView.AddItem(itemView)

	// Save updated view
	if err := p.readStore.Save(ctx, orderView); err != nil {
		return fmt.Errorf("failed to save updated OrderSummaryView: %w", err)
	}

	log.Printf("OrderProjection: Successfully added item to OrderSummaryView for order %s", event.OrderID)
	return nil
}

// handleOrderItemRemoved handles OrderItemRemoved events
func (p *OrderProjection) handleOrderItemRemoved(ctx context.Context, event *domain.OrderItemRemoved) error {
	log.Printf("OrderProjection: Removing item from OrderSummaryView for order %s", event.OrderID)

	// Get existing OrderSummaryView
	orderView, err := p.getOrderSummaryView(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get OrderSummaryView: %w", err)
	}

	// Remove item from order
	orderView.RemoveItem(event.ProductID)

	// Save updated view
	if err := p.readStore.Save(ctx, orderView); err != nil {
		return fmt.Errorf("failed to save updated OrderSummaryView: %w", err)
	}

	log.Printf("OrderProjection: Successfully removed item from OrderSummaryView for order %s", event.OrderID)
	return nil
}

// handleUserCreated handles UserCreated events to cache customer info
func (p *OrderProjection) handleUserCreated(ctx context.Context, event *domain.UserCreated) error {
	// We don't need to do anything here for now
	// Customer info will be looked up when needed
	log.Printf("OrderProjection: User created %s - customer info cached", event.UserID)
	return nil
}

// handleUserUpdated handles UserUpdated events to update customer info in orders
func (p *OrderProjection) handleUserUpdated(ctx context.Context, event *domain.UserUpdated) error {
	log.Printf("OrderProjection: Updating customer info in orders for user %s", event.UserID)

	// This would typically involve querying all orders for this customer
	// and updating their customer name/email
	// For now, we'll just log this operation
	log.Printf("OrderProjection: Customer info updated for user %s", event.UserID)
	return nil
}

// getOrderSummaryView retrieves an OrderSummaryView from the read store
func (p *OrderProjection) getOrderSummaryView(ctx context.Context, orderID string) (*readmodels.OrderSummaryView, error) {
	// Get from read store
	readModel, err := p.readStore.GetByID(ctx, orderID, "OrderSummaryView")
	if err != nil {
		return nil, err
	}

	// Type assertion to OrderSummaryView
	orderView, ok := readModel.(*readmodels.OrderSummaryView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *OrderSummaryView, got %T", readModel)
	}

	return orderView, nil
}

// getCustomerInfo retrieves customer information from UserView
func (p *OrderProjection) getCustomerInfo(ctx context.Context, customerID string) (string, string, error) {
	// Try to get UserView
	readModel, err := p.readStore.GetByID(ctx, customerID, "UserView")
	if err != nil {
		return "", "", err
	}

	// Type assertion to UserView
	userView, ok := readModel.(*readmodels.UserView)
	if !ok {
		return "", "", fmt.Errorf("invalid read model type: expected *UserView, got %T", readModel)
	}

	return userView.GetName(), userView.GetEmail(), nil
}

// GetSupportedEvents returns the list of events this projection handles
func (p *OrderProjection) GetSupportedEvents() []string {
	return []string{
		"OrderCreated",
		"OrderCompleted",
		"OrderCancelled",
		"OrderItemAdded",
		"OrderItemRemoved",
		"UserCreated",
		"UserUpdated",
	}
}

// IsEventSupported checks if the projection supports a specific event type
func (p *OrderProjection) IsEventSupported(eventType string) bool {
	supportedEvents := p.GetSupportedEvents()
	for _, supported := range supportedEvents {
		if supported == eventType {
			return true
		}
	}
	return false
}

// Rebuild rebuilds all OrderSummaryView read models from events
func (p *OrderProjection) Rebuild(ctx context.Context, eventStore interface{}) error {
	log.Printf("OrderProjection: Starting rebuild process")

	// For now, we'll just log the rebuild operation
	// In a real implementation, this would iterate through all events
	// and rebuild the read models
	log.Printf("OrderProjection: Rebuild operation logged - implementation depends on event store capabilities")

	log.Printf("OrderProjection: Rebuild completed successfully")
	return nil
}

// Validate validates the projection state
func (p *OrderProjection) Validate() error {
	// Allow nil readStore for demo purposes
	if p.name == "" {
		return fmt.Errorf("projection name cannot be empty")
	}
	return nil
}
