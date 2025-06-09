package domain

import (
	"cqrs"
	"time"

	"github.com/shopspring/decimal"
)

// User Events

// UserCreated represents a user creation event
type UserCreated struct {
	*cqrs.BaseEventMessage
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(userID, name, email string) *UserCreated {
	event := &UserCreated{
		UserID: userID,
		Name:   name,
		Email:  email,
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("UserCreated", userID, "User", 1, event)
	return event
}

// UserUpdated represents a user update event
type UserUpdated struct {
	*cqrs.BaseEventMessage
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// NewUserUpdated creates a new UserUpdated event
func NewUserUpdated(userID, name, email string) *UserUpdated {
	event := &UserUpdated{
		UserID: userID,
		Name:   name,
		Email:  email,
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("UserUpdated", userID, "User", 1, event)
	return event
}

// Product Events

// ProductCreated represents a product creation event
type ProductCreated struct {
	*cqrs.BaseEventMessage
	ProductID   string          `json:"product_id"`
	Name        string          `json:"name"`
	Price       decimal.Decimal `json:"price"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
}

// NewProductCreated creates a new ProductCreated event
func NewProductCreated(productID, name string, price decimal.Decimal, category, description string) *ProductCreated {
	event := &ProductCreated{
		ProductID:   productID,
		Name:        name,
		Price:       price,
		Category:    category,
		Description: description,
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("ProductCreated", productID, "Product", 1, event)
	return event
}

// ProductUpdated represents a product update event
type ProductUpdated struct {
	*cqrs.BaseEventMessage
	ProductID   string          `json:"product_id"`
	Name        string          `json:"name"`
	Price       decimal.Decimal `json:"price"`
	Category    string          `json:"category"`
	Description string          `json:"description"`
}

// NewProductUpdated creates a new ProductUpdated event
func NewProductUpdated(productID, name string, price decimal.Decimal, category, description string) *ProductUpdated {
	event := &ProductUpdated{
		ProductID:   productID,
		Name:        name,
		Price:       price,
		Category:    category,
		Description: description,
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("ProductUpdated", productID, "Product", 1, event)
	return event
}

// Order Events

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID string          `json:"product_id"`
	Name      string          `json:"name"`
	Price     decimal.Decimal `json:"price"`
	Quantity  int             `json:"quantity"`
	SubTotal  decimal.Decimal `json:"sub_total"`
}

// OrderCreated represents an order creation event
type OrderCreated struct {
	*cqrs.BaseEventMessage
	OrderID    string          `json:"order_id"`
	CustomerID string          `json:"customer_id"`
	Items      []OrderItem     `json:"items"`
	SubTotal   decimal.Decimal `json:"sub_total"`
	TaxAmount  decimal.Decimal `json:"tax_amount"`
	Total      decimal.Decimal `json:"total"`
	OrderDate  time.Time       `json:"order_date"`
}

// NewOrderCreated creates a new OrderCreated event
func NewOrderCreated(orderID, customerID string, items []OrderItem, subTotal, taxAmount, total decimal.Decimal) *OrderCreated {
	event := &OrderCreated{
		OrderID:    orderID,
		CustomerID: customerID,
		Items:      items,
		SubTotal:   subTotal,
		TaxAmount:  taxAmount,
		Total:      total,
		OrderDate:  time.Now(),
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("OrderCreated", orderID, "Order", 1, event)
	return event
}

// OrderCompleted represents an order completion event
type OrderCompleted struct {
	*cqrs.BaseEventMessage
	OrderID     string          `json:"order_id"`
	CustomerID  string          `json:"customer_id"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	CompletedAt time.Time       `json:"completed_at"`
}

// NewOrderCompleted creates a new OrderCompleted event
func NewOrderCompleted(orderID, customerID string, totalAmount decimal.Decimal) *OrderCompleted {
	event := &OrderCompleted{
		OrderID:     orderID,
		CustomerID:  customerID,
		TotalAmount: totalAmount,
		CompletedAt: time.Now(),
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("OrderCompleted", orderID, "Order", 1, event)
	return event
}

// OrderCancelled represents an order cancellation event
type OrderCancelled struct {
	*cqrs.BaseEventMessage
	OrderID     string    `json:"order_id"`
	CustomerID  string    `json:"customer_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

// NewOrderCancelled creates a new OrderCancelled event
func NewOrderCancelled(orderID, customerID, reason string) *OrderCancelled {
	event := &OrderCancelled{
		OrderID:     orderID,
		CustomerID:  customerID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("OrderCancelled", orderID, "Order", 1, event)
	return event
}

// OrderItemAdded represents adding an item to an order
type OrderItemAdded struct {
	*cqrs.BaseEventMessage
	OrderID string    `json:"order_id"`
	Item    OrderItem `json:"item"`
}

// NewOrderItemAdded creates a new OrderItemAdded event
func NewOrderItemAdded(orderID string, item OrderItem) *OrderItemAdded {
	event := &OrderItemAdded{
		OrderID: orderID,
		Item:    item,
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("OrderItemAdded", orderID, "Order", 1, event)
	return event
}

// OrderItemRemoved represents removing an item from an order
type OrderItemRemoved struct {
	*cqrs.BaseEventMessage
	OrderID   string `json:"order_id"`
	ProductID string `json:"product_id"`
}

// NewOrderItemRemoved creates a new OrderItemRemoved event
func NewOrderItemRemoved(orderID, productID string) *OrderItemRemoved {
	event := &OrderItemRemoved{
		OrderID:   orderID,
		ProductID: productID,
	}
	event.BaseEventMessage = cqrs.NewBaseEventMessage("OrderItemRemoved", orderID, "Order", 1, event)
	return event
}
