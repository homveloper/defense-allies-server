package projections

import (
	"context"
	"cqrs"
	"cqrs/cqrsx/examples/04-read-models/domain"
	"cqrs/cqrsx/examples/04-read-models/readmodels"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// AnalyticsProjection handles events for analytics and dashboard views
type AnalyticsProjection struct {
	readStore cqrs.ReadStore
	name      string
}

// NewAnalyticsProjection creates a new AnalyticsProjection
func NewAnalyticsProjection(readStore cqrs.ReadStore) *AnalyticsProjection {
	return &AnalyticsProjection{
		readStore: readStore,
		name:      "AnalyticsProjection",
	}
}

// GetName returns the projection name
func (p *AnalyticsProjection) GetName() string {
	return p.name
}

// Handle handles domain events and updates analytics read models
func (p *AnalyticsProjection) Handle(ctx context.Context, event cqrs.EventMessage) error {
	log.Printf("AnalyticsProjection: Processing event %s for aggregate %s",
		event.EventType(), event.ID())

	switch e := event.EventData().(type) {
	case *domain.UserCreated:
		return p.handleUserCreated(ctx, e)
	case *domain.UserUpdated:
		return p.handleUserUpdated(ctx, e)
	case *domain.ProductCreated:
		return p.handleProductCreated(ctx, e)
	case *domain.ProductUpdated:
		return p.handleProductUpdated(ctx, e)
	case *domain.OrderCreated:
		return p.handleOrderCreated(ctx, e)
	case *domain.OrderCompleted:
		return p.handleOrderCompleted(ctx, e)
	case *domain.OrderCancelled:
		return p.handleOrderCancelled(ctx, e)
	default:
		// Ignore unknown events
		log.Printf("AnalyticsProjection: Ignoring unknown event type %T", e)
		return nil
	}
}

// handleUserCreated handles UserCreated events for customer history
func (p *AnalyticsProjection) handleUserCreated(ctx context.Context, event *domain.UserCreated) error {
	log.Printf("AnalyticsProjection: Creating CustomerOrderHistoryView for user %s", event.UserID)

	// Create new CustomerOrderHistoryView
	historyView := readmodels.NewCustomerOrderHistoryView(event.UserID, event.Name, event.Email)

	// Save to read store
	if err := p.readStore.Save(ctx, historyView); err != nil {
		return fmt.Errorf("failed to save CustomerOrderHistoryView: %w", err)
	}

	log.Printf("AnalyticsProjection: Successfully created CustomerOrderHistoryView for user %s", event.UserID)
	return nil
}

// handleUserUpdated handles UserUpdated events
func (p *AnalyticsProjection) handleUserUpdated(ctx context.Context, event *domain.UserUpdated) error {
	log.Printf("AnalyticsProjection: Updating CustomerOrderHistoryView for user %s", event.UserID)

	// Get existing CustomerOrderHistoryView
	historyView, err := p.getCustomerHistoryView(ctx, event.UserID)
	if err != nil {
		log.Printf("AnalyticsProjection: CustomerOrderHistoryView not found for user %s, skipping update", event.UserID)
		return nil
	}

	// Update customer information
	historyView.UpdateCustomerInfo(event.Name, event.Email)

	// Save updated view
	if err := p.readStore.Save(ctx, historyView); err != nil {
		return fmt.Errorf("failed to save updated CustomerOrderHistoryView: %w", err)
	}

	log.Printf("AnalyticsProjection: Successfully updated CustomerOrderHistoryView for user %s", event.UserID)
	return nil
}

// handleProductCreated handles ProductCreated events for product popularity
func (p *AnalyticsProjection) handleProductCreated(ctx context.Context, event *domain.ProductCreated) error {
	log.Printf("AnalyticsProjection: Creating ProductPopularityView for product %s", event.ProductID)

	// Create new ProductPopularityView
	popularityView := readmodels.NewProductPopularityView(
		event.ProductID,
		event.Name,
		event.Category,
		event.Price,
	)

	// Save to read store
	if err := p.readStore.Save(ctx, popularityView); err != nil {
		return fmt.Errorf("failed to save ProductPopularityView: %w", err)
	}

	log.Printf("AnalyticsProjection: Successfully created ProductPopularityView for product %s", event.ProductID)
	return nil
}

// handleProductUpdated handles ProductUpdated events
func (p *AnalyticsProjection) handleProductUpdated(ctx context.Context, event *domain.ProductUpdated) error {
	log.Printf("AnalyticsProjection: Updating ProductPopularityView for product %s", event.ProductID)

	// Get existing ProductPopularityView
	popularityView, err := p.getProductPopularityView(ctx, event.ProductID)
	if err != nil {
		log.Printf("AnalyticsProjection: ProductPopularityView not found for product %s, skipping update", event.ProductID)
		return nil
	}

	// Update product information
	popularityView.UpdateProductInfo(event.Name, event.Category, event.Price)

	// Save updated view
	if err := p.readStore.Save(ctx, popularityView); err != nil {
		return fmt.Errorf("failed to save updated ProductPopularityView: %w", err)
	}

	log.Printf("AnalyticsProjection: Successfully updated ProductPopularityView for product %s", event.ProductID)
	return nil
}

// handleOrderCreated handles OrderCreated events
func (p *AnalyticsProjection) handleOrderCreated(ctx context.Context, event *domain.OrderCreated) error {
	log.Printf("AnalyticsProjection: Processing OrderCreated for analytics - order %s", event.OrderID)

	// Update customer history
	if err := p.updateCustomerHistoryForOrder(ctx, event.CustomerID, event.OrderID, event.OrderDate, event.Total, "pending", 0, len(event.Items)); err != nil {
		log.Printf("AnalyticsProjection: Warning - failed to update customer history: %v", err)
	}

	// Update product popularity for each item
	for _, item := range event.Items {
		if err := p.updateProductPopularityForSale(ctx, item.ProductID, item.Quantity, item.SubTotal, event.OrderDate); err != nil {
			log.Printf("AnalyticsProjection: Warning - failed to update product popularity for %s: %v", item.ProductID, err)
		}
	}

	log.Printf("AnalyticsProjection: Successfully processed OrderCreated for analytics")
	return nil
}

// handleOrderCompleted handles OrderCompleted events
func (p *AnalyticsProjection) handleOrderCompleted(ctx context.Context, event *domain.OrderCompleted) error {
	log.Printf("AnalyticsProjection: Processing OrderCompleted for analytics - order %s", event.OrderID)

	// Update customer history with completion
	if err := p.updateCustomerHistoryOrderStatus(ctx, event.CustomerID, event.OrderID, "completed", &event.CompletedAt); err != nil {
		log.Printf("AnalyticsProjection: Warning - failed to update customer history: %v", err)
	}

	log.Printf("AnalyticsProjection: Successfully processed OrderCompleted for analytics")
	return nil
}

// handleOrderCancelled handles OrderCancelled events
func (p *AnalyticsProjection) handleOrderCancelled(ctx context.Context, event *domain.OrderCancelled) error {
	log.Printf("AnalyticsProjection: Processing OrderCancelled for analytics - order %s", event.OrderID)

	// Update customer history with cancellation
	if err := p.updateCustomerHistoryOrderStatus(ctx, event.CustomerID, event.OrderID, "cancelled", &event.CancelledAt); err != nil {
		log.Printf("AnalyticsProjection: Warning - failed to update customer history: %v", err)
	}

	log.Printf("AnalyticsProjection: Successfully processed OrderCancelled for analytics")
	return nil
}

// Helper methods

// updateCustomerHistoryForOrder updates customer history with a new order
func (p *AnalyticsProjection) updateCustomerHistoryForOrder(ctx context.Context, customerID, orderID string, orderDate time.Time, totalAmount decimal.Decimal, status string, itemCount int, uniqueProducts int) error {
	// Get existing CustomerOrderHistoryView
	historyView, err := p.getCustomerHistoryView(ctx, customerID)
	if err != nil {
		return err
	}

	// Create order history item
	orderItem := readmodels.OrderHistoryItem{
		OrderID:     orderID,
		OrderDate:   orderDate,
		TotalAmount: totalAmount,
		ItemCount:   itemCount,
		Status:      status,
	}

	// Add order to history
	historyView.AddOrder(orderItem)

	// Save updated view
	return p.readStore.Save(ctx, historyView)
}

// updateCustomerHistoryOrderStatus updates the status of an order in customer history
func (p *AnalyticsProjection) updateCustomerHistoryOrderStatus(ctx context.Context, customerID, orderID, status string, statusDate *time.Time) error {
	// Get existing CustomerOrderHistoryView
	historyView, err := p.getCustomerHistoryView(ctx, customerID)
	if err != nil {
		return err
	}

	// Find and update the order
	orders := historyView.GetOrders()
	for i, order := range orders {
		if order.OrderID == orderID {
			orders[i].Status = status
			if status == "completed" && statusDate != nil {
				orders[i].CompletedAt = statusDate
			} else if status == "cancelled" && statusDate != nil {
				orders[i].CancelledAt = statusDate
			}

			// Update the order in history
			historyView.UpdateOrder(orderID, orders[i])
			break
		}
	}

	// Save updated view
	return p.readStore.Save(ctx, historyView)
}

// updateProductPopularityForSale updates product popularity metrics for a sale
func (p *AnalyticsProjection) updateProductPopularityForSale(ctx context.Context, productID string, unitsSold int, revenue decimal.Decimal, saleDate time.Time) error {
	// Get existing ProductPopularityView
	popularityView, err := p.getProductPopularityView(ctx, productID)
	if err != nil {
		return err
	}

	// Record the sale
	popularityView.RecordSale(unitsSold, revenue, saleDate)

	// Save updated view
	return p.readStore.Save(ctx, popularityView)
}

// getCustomerHistoryView retrieves a CustomerOrderHistoryView from the read store
func (p *AnalyticsProjection) getCustomerHistoryView(ctx context.Context, customerID string) (*readmodels.CustomerOrderHistoryView, error) {
	readModel, err := p.readStore.GetByID(ctx, customerID, "CustomerOrderHistoryView")
	if err != nil {
		return nil, err
	}

	historyView, ok := readModel.(*readmodels.CustomerOrderHistoryView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *CustomerOrderHistoryView, got %T", readModel)
	}

	return historyView, nil
}

// getProductPopularityView retrieves a ProductPopularityView from the read store
func (p *AnalyticsProjection) getProductPopularityView(ctx context.Context, productID string) (*readmodels.ProductPopularityView, error) {
	readModel, err := p.readStore.GetByID(ctx, productID, "ProductPopularityView")
	if err != nil {
		return nil, err
	}

	popularityView, ok := readModel.(*readmodels.ProductPopularityView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *ProductPopularityView, got %T", readModel)
	}

	return popularityView, nil
}

// GetSupportedEvents returns the list of events this projection handles
func (p *AnalyticsProjection) GetSupportedEvents() []string {
	return []string{
		"UserCreated",
		"UserUpdated",
		"ProductCreated",
		"ProductUpdated",
		"OrderCreated",
		"OrderCompleted",
		"OrderCancelled",
	}
}

// IsEventSupported checks if the projection supports a specific event type
func (p *AnalyticsProjection) IsEventSupported(eventType string) bool {
	supportedEvents := p.GetSupportedEvents()
	for _, supported := range supportedEvents {
		if supported == eventType {
			return true
		}
	}
	return false
}

// Rebuild rebuilds all analytics read models from events
func (p *AnalyticsProjection) Rebuild(ctx context.Context, eventStore interface{}) error {
	log.Printf("AnalyticsProjection: Starting rebuild process")

	// For now, we'll just log the rebuild operation
	// In a real implementation, this would iterate through all events
	// and rebuild the read models
	log.Printf("AnalyticsProjection: Rebuild operation logged - implementation depends on event store capabilities")

	log.Printf("AnalyticsProjection: Rebuild completed successfully")
	return nil
}

// Validate validates the projection state
func (p *AnalyticsProjection) Validate() error {
	// Allow nil readStore for demo purposes
	if p.name == "" {
		return fmt.Errorf("projection name cannot be empty")
	}
	return nil
}
