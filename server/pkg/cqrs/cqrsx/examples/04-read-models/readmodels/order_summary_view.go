package readmodels

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"time"

	"github.com/shopspring/decimal"
)

// OrderItemView represents an item in an order summary
type OrderItemView struct {
	ProductID   string          `json:"product_id"`
	ProductName string          `json:"product_name"`
	Price       decimal.Decimal `json:"price"`
	Quantity    int             `json:"quantity"`
	SubTotal    decimal.Decimal `json:"sub_total"`
}

// OrderSummaryView represents an order read model with calculated totals
type OrderSummaryView struct {
	*cqrs.BaseReadModel
	CustomerID     string            `json:"customer_id"`
	CustomerName   string            `json:"customer_name"`
	CustomerEmail  string            `json:"customer_email"`
	Items          []OrderItemView   `json:"items"`
	SubTotal       decimal.Decimal   `json:"sub_total"`
	TaxAmount      decimal.Decimal   `json:"tax_amount"`
	DiscountAmount decimal.Decimal   `json:"discount_amount"`
	TotalAmount    decimal.Decimal   `json:"total_amount"`
	Status         string            `json:"status"`
	OrderDate      time.Time         `json:"order_date"`
	CompletedAt    *time.Time        `json:"completed_at,omitempty"`
	CancelledAt    *time.Time        `json:"cancelled_at,omitempty"`
	CancelReason   string            `json:"cancel_reason,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

// NewOrderSummaryView creates a new OrderSummaryView
func NewOrderSummaryView(orderID, customerID, customerName, customerEmail string, items []OrderItemView, subTotal, taxAmount, totalAmount decimal.Decimal, orderDate time.Time) *OrderSummaryView {
	now := time.Now()
	return &OrderSummaryView{
		BaseReadModel:  cqrs.NewBaseReadModel(orderID, "OrderSummaryView", nil),
		CustomerID:     customerID,
		CustomerName:   customerName,
		CustomerEmail:  customerEmail,
		Items:          items,
		SubTotal:       subTotal,
		TaxAmount:      taxAmount,
		DiscountAmount: decimal.Zero,
		TotalAmount:    totalAmount,
		Status:         "pending",
		OrderDate:      orderDate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Business Methods

// UpdateStatus updates the order status
func (osv *OrderSummaryView) UpdateStatus(status string) {
	osv.Status = status
	osv.UpdatedAt = time.Now()
	osv.IncrementVersion()
}

// CompleteOrder marks the order as completed
func (osv *OrderSummaryView) CompleteOrder(completedAt time.Time) {
	osv.Status = "completed"
	osv.CompletedAt = &completedAt
	osv.UpdatedAt = time.Now()
	osv.IncrementVersion()
}

// CancelOrder marks the order as cancelled
func (osv *OrderSummaryView) CancelOrder(cancelledAt time.Time, reason string) {
	osv.Status = "cancelled"
	osv.CancelledAt = &cancelledAt
	osv.CancelReason = reason
	osv.UpdatedAt = time.Now()
	osv.IncrementVersion()
}

// AddItem adds an item to the order
func (osv *OrderSummaryView) AddItem(item OrderItemView) {
	osv.Items = append(osv.Items, item)
	osv.recalculateTotals()
	osv.UpdatedAt = time.Now()
	osv.IncrementVersion()
}

// RemoveItem removes an item from the order
func (osv *OrderSummaryView) RemoveItem(productID string) {
	for i, item := range osv.Items {
		if item.ProductID == productID {
			osv.Items = append(osv.Items[:i], osv.Items[i+1:]...)
			break
		}
	}
	osv.recalculateTotals()
	osv.UpdatedAt = time.Now()
	osv.IncrementVersion()
}

// ApplyDiscount applies a discount to the order
func (osv *OrderSummaryView) ApplyDiscount(discountAmount decimal.Decimal) {
	osv.DiscountAmount = discountAmount
	osv.TotalAmount = osv.SubTotal.Add(osv.TaxAmount).Sub(osv.DiscountAmount)
	osv.UpdatedAt = time.Now()
	osv.IncrementVersion()
}

// recalculateTotals recalculates order totals
func (osv *OrderSummaryView) recalculateTotals() {
	osv.SubTotal = decimal.Zero
	for _, item := range osv.Items {
		osv.SubTotal = osv.SubTotal.Add(item.SubTotal)
	}
	
	// Calculate tax (10% tax rate)
	taxRate := decimal.NewFromFloat(0.1)
	osv.TaxAmount = osv.SubTotal.Mul(taxRate)
	
	// Calculate total
	osv.TotalAmount = osv.SubTotal.Add(osv.TaxAmount).Sub(osv.DiscountAmount)
}

// Getters

func (osv *OrderSummaryView) GetCustomerID() string {
	return osv.CustomerID
}

func (osv *OrderSummaryView) GetCustomerName() string {
	return osv.CustomerName
}

func (osv *OrderSummaryView) GetCustomerEmail() string {
	return osv.CustomerEmail
}

func (osv *OrderSummaryView) GetItems() []OrderItemView {
	return osv.Items
}

func (osv *OrderSummaryView) GetSubTotal() decimal.Decimal {
	return osv.SubTotal
}

func (osv *OrderSummaryView) GetTaxAmount() decimal.Decimal {
	return osv.TaxAmount
}

func (osv *OrderSummaryView) GetDiscountAmount() decimal.Decimal {
	return osv.DiscountAmount
}

func (osv *OrderSummaryView) GetTotalAmount() decimal.Decimal {
	return osv.TotalAmount
}

func (osv *OrderSummaryView) GetStatus() string {
	return osv.Status
}

func (osv *OrderSummaryView) GetOrderDate() time.Time {
	return osv.OrderDate
}

func (osv *OrderSummaryView) GetCompletedAt() *time.Time {
	return osv.CompletedAt
}

func (osv *OrderSummaryView) GetCancelledAt() *time.Time {
	return osv.CancelledAt
}

func (osv *OrderSummaryView) GetCancelReason() string {
	return osv.CancelReason
}

func (osv *OrderSummaryView) GetCreatedAt() time.Time {
	return osv.CreatedAt
}

func (osv *OrderSummaryView) GetUpdatedAt() time.Time {
	return osv.UpdatedAt
}

// Helper Methods

// IncrementVersion increments the version and updates timestamp
func (osv *OrderSummaryView) IncrementVersion() {
	osv.BaseReadModel.IncrementVersion()
	osv.UpdatedAt = time.Now()
}

// GetItemCount returns the total number of items
func (osv *OrderSummaryView) GetItemCount() int {
	totalQuantity := 0
	for _, item := range osv.Items {
		totalQuantity += item.Quantity
	}
	return totalQuantity
}

// GetUniqueProductCount returns the number of unique products
func (osv *OrderSummaryView) GetUniqueProductCount() int {
	return len(osv.Items)
}

// IsCompleted checks if the order is completed
func (osv *OrderSummaryView) IsCompleted() bool {
	return osv.Status == "completed"
}

// IsCancelled checks if the order is cancelled
func (osv *OrderSummaryView) IsCancelled() bool {
	return osv.Status == "cancelled"
}

// IsPending checks if the order is pending
func (osv *OrderSummaryView) IsPending() bool {
	return osv.Status == "pending"
}

// Validation

// Validate validates the OrderSummaryView state
func (osv *OrderSummaryView) Validate() error {
	if err := osv.BaseReadModel.Validate(); err != nil {
		return err
	}
	if osv.CustomerID == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "customer ID cannot be empty", nil)
	}
	if osv.CustomerName == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "customer name cannot be empty", nil)
	}
	if osv.TotalAmount.IsNegative() {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total amount cannot be negative", nil)
	}
	if osv.Status == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "order status cannot be empty", nil)
	}
	return nil
}

// Repository Interface

// OrderSummaryViewRepository defines the interface for OrderSummaryView persistence
type OrderSummaryViewRepository interface {
	// Save saves an OrderSummaryView
	Save(ctx context.Context, orderView *OrderSummaryView) error
	
	// GetByID retrieves an OrderSummaryView by ID
	GetByID(ctx context.Context, orderID string) (*OrderSummaryView, error)
	
	// GetByCustomerID retrieves OrderSummaryViews by customer ID
	GetByCustomerID(ctx context.Context, customerID string) ([]*OrderSummaryView, error)
	
	// GetByStatus retrieves OrderSummaryViews by status
	GetByStatus(ctx context.Context, status string) ([]*OrderSummaryView, error)
	
	// GetByDateRange retrieves OrderSummaryViews within date range
	GetByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*OrderSummaryView, error)
	
	// GetRecentOrders retrieves recent orders (last N orders)
	GetRecentOrders(ctx context.Context, limit int) ([]*OrderSummaryView, error)
	
	// Delete removes an OrderSummaryView
	Delete(ctx context.Context, orderID string) error
}
