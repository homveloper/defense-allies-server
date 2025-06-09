package readmodels

import (
	"context"
	"cqrs"
	"time"

	"github.com/shopspring/decimal"
)

// OrderHistoryItem represents a single order in customer history
type OrderHistoryItem struct {
	OrderID      string          `json:"order_id"`
	OrderDate    time.Time       `json:"order_date"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	ItemCount    int             `json:"item_count"`
	Status       string          `json:"status"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	CancelledAt  *time.Time      `json:"cancelled_at,omitempty"`
	CancelReason string          `json:"cancel_reason,omitempty"`
}

// CustomerOrderHistoryView represents a denormalized view of customer order history
type CustomerOrderHistoryView struct {
	*cqrs.BaseReadModel
	CustomerID      string             `json:"customer_id"`
	CustomerName    string             `json:"customer_name"`
	CustomerEmail   string             `json:"customer_email"`
	Orders          []OrderHistoryItem `json:"orders"`
	TotalOrders     int                `json:"total_orders"`
	CompletedOrders int                `json:"completed_orders"`
	CancelledOrders int                `json:"cancelled_orders"`
	TotalSpent      decimal.Decimal    `json:"total_spent"`
	AverageOrder    decimal.Decimal    `json:"average_order"`
	LastOrderDate   *time.Time         `json:"last_order_date,omitempty"`
	FirstOrderDate  *time.Time         `json:"first_order_date,omitempty"`
	Tags            []string           `json:"tags"` // VIP, Frequent, etc.
	CustomerTier    string             `json:"customer_tier"`
	IsActive        bool               `json:"is_active"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// NewCustomerOrderHistoryView creates a new CustomerOrderHistoryView
func NewCustomerOrderHistoryView(customerID, customerName, customerEmail string) *CustomerOrderHistoryView {
	now := time.Now()
	return &CustomerOrderHistoryView{
		BaseReadModel: cqrs.NewBaseReadModel(customerID, "CustomerOrderHistoryView", nil),
		CustomerID:    customerID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		Orders:        make([]OrderHistoryItem, 0),
		TotalOrders:   0,
		TotalSpent:    decimal.Zero,
		AverageOrder:  decimal.Zero,
		Tags:          make([]string, 0),
		CustomerTier:  "Bronze",
		IsActive:      false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Business Methods

// AddOrder adds an order to the customer history
func (chv *CustomerOrderHistoryView) AddOrder(orderItem OrderHistoryItem) {
	chv.Orders = append(chv.Orders, orderItem)
	chv.recalculateStats()
	chv.updateTags()
	chv.UpdatedAt = time.Now()
	chv.IncrementVersion()
}

// UpdateOrder updates an existing order in the history
func (chv *CustomerOrderHistoryView) UpdateOrder(orderID string, updatedItem OrderHistoryItem) {
	for i, order := range chv.Orders {
		if order.OrderID == orderID {
			chv.Orders[i] = updatedItem
			break
		}
	}
	chv.recalculateStats()
	chv.updateTags()
	chv.UpdatedAt = time.Now()
	chv.IncrementVersion()
}

// RemoveOrder removes an order from the history
func (chv *CustomerOrderHistoryView) RemoveOrder(orderID string) {
	for i, order := range chv.Orders {
		if order.OrderID == orderID {
			chv.Orders = append(chv.Orders[:i], chv.Orders[i+1:]...)
			break
		}
	}
	chv.recalculateStats()
	chv.updateTags()
	chv.UpdatedAt = time.Now()
	chv.IncrementVersion()
}

// UpdateCustomerInfo updates customer information
func (chv *CustomerOrderHistoryView) UpdateCustomerInfo(name, email string) {
	chv.CustomerName = name
	chv.CustomerEmail = email
	chv.UpdatedAt = time.Now()
	chv.IncrementVersion()
}

// recalculateStats recalculates all statistics
func (chv *CustomerOrderHistoryView) recalculateStats() {
	chv.TotalOrders = len(chv.Orders)
	chv.CompletedOrders = 0
	chv.CancelledOrders = 0
	chv.TotalSpent = decimal.Zero

	var firstOrderDate, lastOrderDate *time.Time

	for _, order := range chv.Orders {
		switch order.Status {
		case "completed":
			chv.CompletedOrders++
			chv.TotalSpent = chv.TotalSpent.Add(order.TotalAmount)
		case "cancelled":
			chv.CancelledOrders++
		}

		// Track first and last order dates
		if firstOrderDate == nil || order.OrderDate.Before(*firstOrderDate) {
			firstOrderDate = &order.OrderDate
		}
		if lastOrderDate == nil || order.OrderDate.After(*lastOrderDate) {
			lastOrderDate = &order.OrderDate
		}
	}

	chv.FirstOrderDate = firstOrderDate
	chv.LastOrderDate = lastOrderDate

	// Calculate average order value
	if chv.CompletedOrders > 0 {
		chv.AverageOrder = chv.TotalSpent.Div(decimal.NewFromInt(int64(chv.CompletedOrders)))
	} else {
		chv.AverageOrder = decimal.Zero
	}

	// Update customer tier
	chv.updateCustomerTier()

	// Update active status (ordered within 30 days)
	chv.updateActiveStatus()
}

// updateCustomerTier updates customer tier based on total spent
func (chv *CustomerOrderHistoryView) updateCustomerTier() {
	if chv.TotalSpent.GreaterThanOrEqual(decimal.NewFromInt(5000)) {
		chv.CustomerTier = "Platinum"
	} else if chv.TotalSpent.GreaterThanOrEqual(decimal.NewFromInt(1000)) {
		chv.CustomerTier = "Gold"
	} else if chv.TotalSpent.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		chv.CustomerTier = "Silver"
	} else {
		chv.CustomerTier = "Bronze"
	}
}

// updateActiveStatus updates active status based on recent orders
func (chv *CustomerOrderHistoryView) updateActiveStatus() {
	if chv.LastOrderDate == nil {
		chv.IsActive = false
		return
	}

	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	chv.IsActive = chv.LastOrderDate.After(thirtyDaysAgo)
}

// updateTags updates customer tags based on behavior
func (chv *CustomerOrderHistoryView) updateTags() {
	chv.Tags = make([]string, 0)

	// VIP tag
	if chv.TotalSpent.GreaterThan(decimal.NewFromInt(1000)) {
		chv.Tags = append(chv.Tags, "VIP")
	}

	// Frequent customer tag (more than 10 orders)
	if chv.TotalOrders > 10 {
		chv.Tags = append(chv.Tags, "Frequent")
	}

	// High value tag (average order > $100)
	if chv.AverageOrder.GreaterThan(decimal.NewFromInt(100)) {
		chv.Tags = append(chv.Tags, "HighValue")
	}

	// Active tag
	if chv.IsActive {
		chv.Tags = append(chv.Tags, "Active")
	}

	// New customer tag (first order within 30 days)
	if chv.FirstOrderDate != nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		if chv.FirstOrderDate.After(thirtyDaysAgo) {
			chv.Tags = append(chv.Tags, "New")
		}
	}
}

// Getters

func (chv *CustomerOrderHistoryView) GetCustomerID() string {
	return chv.CustomerID
}

func (chv *CustomerOrderHistoryView) GetCustomerName() string {
	return chv.CustomerName
}

func (chv *CustomerOrderHistoryView) GetCustomerEmail() string {
	return chv.CustomerEmail
}

func (chv *CustomerOrderHistoryView) GetOrders() []OrderHistoryItem {
	return chv.Orders
}

func (chv *CustomerOrderHistoryView) GetTotalOrders() int {
	return chv.TotalOrders
}

func (chv *CustomerOrderHistoryView) GetCompletedOrders() int {
	return chv.CompletedOrders
}

func (chv *CustomerOrderHistoryView) GetCancelledOrders() int {
	return chv.CancelledOrders
}

func (chv *CustomerOrderHistoryView) GetTotalSpent() decimal.Decimal {
	return chv.TotalSpent
}

func (chv *CustomerOrderHistoryView) GetAverageOrder() decimal.Decimal {
	return chv.AverageOrder
}

func (chv *CustomerOrderHistoryView) GetLastOrderDate() *time.Time {
	return chv.LastOrderDate
}

func (chv *CustomerOrderHistoryView) GetFirstOrderDate() *time.Time {
	return chv.FirstOrderDate
}

func (chv *CustomerOrderHistoryView) GetTags() []string {
	return chv.Tags
}

func (chv *CustomerOrderHistoryView) GetCustomerTier() string {
	return chv.CustomerTier
}

func (chv *CustomerOrderHistoryView) GetIsActive() bool {
	return chv.IsActive
}

func (chv *CustomerOrderHistoryView) GetCreatedAt() time.Time {
	return chv.CreatedAt
}

func (chv *CustomerOrderHistoryView) GetUpdatedAt() time.Time {
	return chv.UpdatedAt
}

// Helper Methods

// IncrementVersion increments the version and updates timestamp
func (chv *CustomerOrderHistoryView) IncrementVersion() {
	chv.BaseReadModel.IncrementVersion()
	chv.UpdatedAt = time.Now()
}

// GetRecentOrders returns orders from the last N days
func (chv *CustomerOrderHistoryView) GetRecentOrders(days int) []OrderHistoryItem {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	recentOrders := make([]OrderHistoryItem, 0)

	for _, order := range chv.Orders {
		if order.OrderDate.After(cutoffDate) {
			recentOrders = append(recentOrders, order)
		}
	}

	return recentOrders
}

// GetOrdersByStatus returns orders with specific status
func (chv *CustomerOrderHistoryView) GetOrdersByStatus(status string) []OrderHistoryItem {
	filteredOrders := make([]OrderHistoryItem, 0)

	for _, order := range chv.Orders {
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}

	return filteredOrders
}

// HasTag checks if customer has a specific tag
func (chv *CustomerOrderHistoryView) HasTag(tag string) bool {
	for _, t := range chv.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// Validation

// Validate validates the CustomerOrderHistoryView state
func (chv *CustomerOrderHistoryView) Validate() error {
	if err := chv.BaseReadModel.Validate(); err != nil {
		return err
	}
	if chv.CustomerID == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "customer ID cannot be empty", nil)
	}
	if chv.CustomerName == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "customer name cannot be empty", nil)
	}
	if chv.TotalOrders < 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total orders cannot be negative", nil)
	}
	if chv.TotalSpent.IsNegative() {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total spent cannot be negative", nil)
	}
	return nil
}

// Repository Interface

// CustomerOrderHistoryViewRepository defines the interface for CustomerOrderHistoryView persistence
type CustomerOrderHistoryViewRepository interface {
	// Save saves a CustomerOrderHistoryView
	Save(ctx context.Context, historyView *CustomerOrderHistoryView) error

	// GetByID retrieves a CustomerOrderHistoryView by customer ID
	GetByID(ctx context.Context, customerID string) (*CustomerOrderHistoryView, error)

	// GetByTier retrieves CustomerOrderHistoryViews by tier
	GetByTier(ctx context.Context, tier string) ([]*CustomerOrderHistoryView, error)

	// GetByTag retrieves CustomerOrderHistoryViews by tag
	GetByTag(ctx context.Context, tag string) ([]*CustomerOrderHistoryView, error)

	// GetActiveCustomers retrieves active customers
	GetActiveCustomers(ctx context.Context) ([]*CustomerOrderHistoryView, error)

	// GetTopCustomers retrieves top customers by total spent
	GetTopCustomers(ctx context.Context, limit int) ([]*CustomerOrderHistoryView, error)

	// Delete removes a CustomerOrderHistoryView
	Delete(ctx context.Context, customerID string) error
}
