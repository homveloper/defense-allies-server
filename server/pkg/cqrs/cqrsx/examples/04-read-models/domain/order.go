package domain

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Order represents an order aggregate in the system
type Order struct {
	*cqrs.BaseAggregate
	customerID     string
	items          []OrderItem
	subTotal       decimal.Decimal
	taxAmount      decimal.Decimal
	discountAmount decimal.Decimal
	totalAmount    decimal.Decimal
	status         OrderStatus
	orderDate      time.Time
	completedAt    *time.Time
	cancelledAt    *time.Time
	cancelReason   string
}

// NewOrder creates a new Order aggregate
func NewOrder(id, customerID string, items []OrderItem) *Order {
	subTotal := calculateSubTotal(items)
	taxAmount := calculateTax(subTotal)
	totalAmount := subTotal.Add(taxAmount)

	order := &Order{
		BaseAggregate:  cqrs.NewBaseAggregate(id, "Order"),
		customerID:     customerID,
		items:          items,
		subTotal:       subTotal,
		taxAmount:      taxAmount,
		discountAmount: decimal.Zero,
		totalAmount:    totalAmount,
		status:         OrderStatusPending,
		orderDate:      time.Now(),
	}

	// Apply creation event
	event := NewOrderCreated(id, customerID, items, subTotal, taxAmount, totalAmount)
	order.ApplyEvent(event)

	return order
}

// LoadOrderFromHistory loads an Order aggregate from event history
func LoadOrderFromHistory(id string, events []cqrs.EventMessage) (*Order, error) {
	order := &Order{
		BaseAggregate:  cqrs.NewBaseAggregate(id, "Order"),
		subTotal:       decimal.Zero,
		taxAmount:      decimal.Zero,
		discountAmount: decimal.Zero,
		totalAmount:    decimal.Zero,
		status:         OrderStatusPending,
	}

	for _, event := range events {
		if err := order.applyDomainEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event: %w", err)
		}
	}

	return order, nil
}

// Business Methods

// AddItem adds an item to the order
func (o *Order) AddItem(item OrderItem) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("cannot add items to %s order", o.status)
	}

	event := NewOrderItemAdded(o.ID(), item)
	o.ApplyEvent(event)

	return nil
}

// RemoveItem removes an item from the order
func (o *Order) RemoveItem(productID string) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("cannot remove items from %s order", o.status)
	}

	// Check if item exists
	found := false
	for _, item := range o.items {
		if item.ProductID == productID {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("item with product ID %s not found in order", productID)
	}

	event := NewOrderItemRemoved(o.ID(), productID)
	o.ApplyEvent(event)

	return nil
}

// Complete completes the order
func (o *Order) Complete() error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("cannot complete %s order", o.status)
	}

	if len(o.items) == 0 {
		return fmt.Errorf("cannot complete order with no items")
	}

	event := NewOrderCompleted(o.ID(), o.customerID, o.totalAmount)
	o.ApplyEvent(event)

	return nil
}

// Cancel cancels the order
func (o *Order) Cancel(reason string) error {
	if o.status != OrderStatusPending {
		return fmt.Errorf("cannot cancel %s order", o.status)
	}

	event := NewOrderCancelled(o.ID(), o.customerID, reason)
	o.ApplyEvent(event)

	return nil
}

// recalculateTotals recalculates order totals
func (o *Order) recalculateTotals() {
	o.subTotal = calculateSubTotal(o.items)
	o.taxAmount = calculateTax(o.subTotal)
	o.totalAmount = o.subTotal.Add(o.taxAmount).Sub(o.discountAmount)
}

// Getters

func (o *Order) GetCustomerID() string {
	return o.customerID
}

func (o *Order) GetItems() []OrderItem {
	return o.items
}

func (o *Order) GetSubTotal() decimal.Decimal {
	return o.subTotal
}

func (o *Order) GetTaxAmount() decimal.Decimal {
	return o.taxAmount
}

func (o *Order) GetDiscountAmount() decimal.Decimal {
	return o.discountAmount
}

func (o *Order) GetTotalAmount() decimal.Decimal {
	return o.totalAmount
}

func (o *Order) GetStatus() OrderStatus {
	return o.status
}

func (o *Order) GetOrderDate() time.Time {
	return o.orderDate
}

func (o *Order) GetCompletedAt() *time.Time {
	return o.completedAt
}

func (o *Order) GetCancelledAt() *time.Time {
	return o.cancelledAt
}

func (o *Order) GetCancelReason() string {
	return o.cancelReason
}

// Event Application

// applyDomainEvent applies domain events to the aggregate
func (o *Order) applyDomainEvent(event cqrs.EventMessage) error {
	switch e := event.EventData().(type) {
	case *OrderCreated:
		return o.applyOrderCreated(e)
	case *OrderCompleted:
		return o.applyOrderCompleted(e)
	case *OrderCancelled:
		return o.applyOrderCancelled(e)
	case *OrderItemAdded:
		return o.applyOrderItemAdded(e)
	case *OrderItemRemoved:
		return o.applyOrderItemRemoved(e)
	default:
		// Ignore unknown events
		return nil
	}
}

// applyOrderCreated applies OrderCreated event
func (o *Order) applyOrderCreated(event *OrderCreated) error {
	o.customerID = event.CustomerID
	o.items = event.Items
	o.subTotal = event.SubTotal
	o.taxAmount = event.TaxAmount
	o.totalAmount = event.Total
	o.status = OrderStatusPending
	o.orderDate = event.OrderDate
	return nil
}

// applyOrderCompleted applies OrderCompleted event
func (o *Order) applyOrderCompleted(event *OrderCompleted) error {
	o.status = OrderStatusCompleted
	o.completedAt = &event.CompletedAt
	return nil
}

// applyOrderCancelled applies OrderCancelled event
func (o *Order) applyOrderCancelled(event *OrderCancelled) error {
	o.status = OrderStatusCancelled
	o.cancelledAt = &event.CancelledAt
	o.cancelReason = event.Reason
	return nil
}

// applyOrderItemAdded applies OrderItemAdded event
func (o *Order) applyOrderItemAdded(event *OrderItemAdded) error {
	o.items = append(o.items, event.Item)
	o.recalculateTotals()
	return nil
}

// applyOrderItemRemoved applies OrderItemRemoved event
func (o *Order) applyOrderItemRemoved(event *OrderItemRemoved) error {
	for i, item := range o.items {
		if item.ProductID == event.ProductID {
			o.items = append(o.items[:i], o.items[i+1:]...)
			break
		}
	}
	o.recalculateTotals()
	return nil
}

// Validation

// Validate validates the order aggregate state
func (o *Order) Validate() error {
	if o.ID() == "" {
		return fmt.Errorf("order ID cannot be empty")
	}
	if o.customerID == "" {
		return fmt.Errorf("customer ID cannot be empty")
	}
	if o.totalAmount.IsNegative() {
		return fmt.Errorf("total amount cannot be negative")
	}
	return nil
}

// Helper functions

// calculateSubTotal calculates the subtotal for order items
func calculateSubTotal(items []OrderItem) decimal.Decimal {
	subTotal := decimal.Zero
	for _, item := range items {
		subTotal = subTotal.Add(item.SubTotal)
	}
	return subTotal
}

// calculateTax calculates tax amount (10% tax rate)
func calculateTax(subTotal decimal.Decimal) decimal.Decimal {
	taxRate := decimal.NewFromFloat(0.1) // 10% tax
	return subTotal.Mul(taxRate)
}

// Repository Interface

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	cqrs.EventSourcedRepository

	// FindByCustomerID finds orders by customer ID
	FindByCustomerID(ctx context.Context, customerID string) ([]*Order, error)

	// FindByStatus finds orders by status
	FindByStatus(ctx context.Context, status OrderStatus) ([]*Order, error)

	// GetOrderStats gets order statistics
	GetOrderStats(ctx context.Context, orderID string) (*OrderStats, error)
}

// OrderStats represents order statistics
type OrderStats struct {
	OrderID     string          `json:"order_id"`
	CustomerID  string          `json:"customer_id"`
	ItemCount   int             `json:"item_count"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Status      OrderStatus     `json:"status"`
	OrderDate   time.Time       `json:"order_date"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// Commands

// CreateOrderCommand represents a command to create an order
type CreateOrderCommand struct {
	*cqrs.BaseCommand
	CustomerID string      `json:"customer_id"`
	Items      []OrderItem `json:"items"`
}

// NewCreateOrderCommand creates a new CreateOrderCommand
func NewCreateOrderCommand(orderID, customerID string, items []OrderItem) *CreateOrderCommand {
	return &CreateOrderCommand{
		BaseCommand: cqrs.NewBaseCommand("CreateOrder", orderID, "Order", nil),
		CustomerID:  customerID,
		Items:       items,
	}
}

// CompleteOrderCommand represents a command to complete an order
type CompleteOrderCommand struct {
	*cqrs.BaseCommand
}

// NewCompleteOrderCommand creates a new CompleteOrderCommand
func NewCompleteOrderCommand(orderID string) *CompleteOrderCommand {
	return &CompleteOrderCommand{
		BaseCommand: cqrs.NewBaseCommand("CompleteOrder", orderID, "Order", nil),
	}
}

// CancelOrderCommand represents a command to cancel an order
type CancelOrderCommand struct {
	*cqrs.BaseCommand
	Reason string `json:"reason"`
}

// NewCancelOrderCommand creates a new CancelOrderCommand
func NewCancelOrderCommand(orderID, reason string) *CancelOrderCommand {
	return &CancelOrderCommand{
		BaseCommand: cqrs.NewBaseCommand("CancelOrder", orderID, "Order", nil),
		Reason:      reason,
	}
}

// AddOrderItemCommand represents a command to add an item to an order
type AddOrderItemCommand struct {
	*cqrs.BaseCommand
	Item OrderItem `json:"item"`
}

// NewAddOrderItemCommand creates a new AddOrderItemCommand
func NewAddOrderItemCommand(orderID string, item OrderItem) *AddOrderItemCommand {
	return &AddOrderItemCommand{
		BaseCommand: cqrs.NewBaseCommand("AddOrderItem", orderID, "Order", nil),
		Item:        item,
	}
}

// Command Handlers

// OrderCommandHandler handles order-related commands
type OrderCommandHandler struct {
	repository OrderRepository
}

// NewOrderCommandHandler creates a new OrderCommandHandler
func NewOrderCommandHandler(repository OrderRepository) *OrderCommandHandler {
	return &OrderCommandHandler{
		repository: repository,
	}
}

// Handle handles order commands
func (h *OrderCommandHandler) Handle(ctx context.Context, command cqrs.Command) (interface{}, error) {
	switch cmd := command.(type) {
	case *CreateOrderCommand:
		return h.handleCreateOrder(ctx, cmd)
	case *CompleteOrderCommand:
		return h.handleCompleteOrder(ctx, cmd)
	case *CancelOrderCommand:
		return h.handleCancelOrder(ctx, cmd)
	case *AddOrderItemCommand:
		return h.handleAddOrderItem(ctx, cmd)
	default:
		return nil, fmt.Errorf("unknown command type: %T", command)
	}
}

// handleCreateOrder handles CreateOrderCommand
func (h *OrderCommandHandler) handleCreateOrder(ctx context.Context, cmd *CreateOrderCommand) (*Order, error) {
	// Validate items
	if len(cmd.Items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	// Create new order
	order := NewOrder(cmd.ID(), cmd.CustomerID, cmd.Items)

	// Validate
	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("order validation failed: %w", err)
	}

	// Save if repository is available (버전 관리 자동화)
	if h.repository != nil {
		if err := h.repository.Save(ctx, order, 0); err != nil {
			return nil, fmt.Errorf("failed to save order: %w", err)
		}
	}

	return order, nil
}

// handleCompleteOrder handles CompleteOrderCommand
func (h *OrderCommandHandler) handleCompleteOrder(ctx context.Context, cmd *CompleteOrderCommand) (*Order, error) {
	// Check if repository is available
	if h.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	// Load order
	order, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load order: %w", err)
	}

	orderAggregate, ok := order.(*Order)
	if !ok {
		return nil, fmt.Errorf("invalid aggregate type")
	}

	// Complete order
	if err := orderAggregate.Complete(); err != nil {
		return nil, fmt.Errorf("failed to complete order: %w", err)
	}

	// Save (버전 관리 자동화)
	if err := h.repository.Save(ctx, orderAggregate, 0); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return orderAggregate, nil
}

// handleCancelOrder handles CancelOrderCommand
func (h *OrderCommandHandler) handleCancelOrder(ctx context.Context, cmd *CancelOrderCommand) (*Order, error) {
	// Check if repository is available
	if h.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	// Load order
	order, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load order: %w", err)
	}

	orderAggregate, ok := order.(*Order)
	if !ok {
		return nil, fmt.Errorf("invalid aggregate type")
	}

	// Cancel order
	if err := orderAggregate.Cancel(cmd.Reason); err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Save (버전 관리 자동화)
	if err := h.repository.Save(ctx, orderAggregate, 0); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return orderAggregate, nil
}

// handleAddOrderItem handles AddOrderItemCommand
func (h *OrderCommandHandler) handleAddOrderItem(ctx context.Context, cmd *AddOrderItemCommand) (*Order, error) {
	// Check if repository is available
	if h.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	// Load order
	order, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to load order: %w", err)
	}

	orderAggregate, ok := order.(*Order)
	if !ok {
		return nil, fmt.Errorf("invalid aggregate type")
	}

	// Add item
	if err := orderAggregate.AddItem(cmd.Item); err != nil {
		return nil, fmt.Errorf("failed to add item: %w", err)
	}

	// Save (버전 관리 자동화)
	if err := h.repository.Save(ctx, orderAggregate, 0); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	return orderAggregate, nil
}
