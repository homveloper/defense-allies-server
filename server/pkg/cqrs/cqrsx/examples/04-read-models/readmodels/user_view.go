package readmodels

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"time"

	"github.com/shopspring/decimal"
)

// UserView represents a user read model optimized for queries
type UserView struct {
	*cqrs.BaseReadModel
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	TotalOrders   int             `json:"total_orders"`
	TotalSpent    decimal.Decimal `json:"total_spent"`
	LastOrderDate *time.Time      `json:"last_order_date,omitempty"`
	IsVIP         bool            `json:"is_vip"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// NewUserView creates a new UserView
func NewUserView(userID, name, email string) *UserView {
	now := time.Now()
	return &UserView{
		BaseReadModel: cqrs.NewBaseReadModel(userID, "UserView", nil),
		Name:          name,
		Email:         email,
		TotalOrders:   0,
		TotalSpent:    decimal.Zero,
		IsVIP:         false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Business Methods

// UpdateProfile updates user profile information
func (uv *UserView) UpdateProfile(name, email string) {
	uv.Name = name
	uv.Email = email
	uv.UpdatedAt = time.Now()
	uv.IncrementVersion()
}

// RecordOrder records a completed order
func (uv *UserView) RecordOrder(orderAmount decimal.Decimal, orderDate time.Time) {
	uv.TotalOrders++
	uv.TotalSpent = uv.TotalSpent.Add(orderAmount)
	uv.LastOrderDate = &orderDate
	uv.UpdatedAt = time.Now()

	// Check if user becomes VIP (spent more than $1000)
	if uv.TotalSpent.GreaterThan(decimal.NewFromInt(1000)) {
		uv.IsVIP = true
	}

	uv.IncrementVersion()
}

// Getters

func (uv *UserView) GetName() string {
	return uv.Name
}

func (uv *UserView) GetEmail() string {
	return uv.Email
}

func (uv *UserView) GetTotalOrders() int {
	return uv.TotalOrders
}

func (uv *UserView) GetTotalSpent() decimal.Decimal {
	return uv.TotalSpent
}

func (uv *UserView) GetLastOrderDate() *time.Time {
	return uv.LastOrderDate
}

func (uv *UserView) GetIsVIP() bool {
	return uv.IsVIP
}

func (uv *UserView) GetCreatedAt() time.Time {
	return uv.CreatedAt
}

func (uv *UserView) GetUpdatedAt() time.Time {
	return uv.UpdatedAt
}

// Helper Methods

// IncrementVersion increments the version and updates timestamp
func (uv *UserView) IncrementVersion() {
	uv.BaseReadModel.IncrementVersion()
	uv.UpdatedAt = time.Now()
}

// ToMap converts UserView to map for easy serialization
func (uv *UserView) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"id":           uv.GetID(),
		"type":         uv.GetType(),
		"version":      uv.GetVersion(),
		"name":         uv.Name,
		"email":        uv.Email,
		"total_orders": uv.TotalOrders,
		"total_spent":  uv.TotalSpent.String(),
		"is_vip":       uv.IsVIP,
		"created_at":   uv.CreatedAt,
		"updated_at":   uv.UpdatedAt,
		"last_updated": uv.GetLastUpdated(),
	}

	if uv.LastOrderDate != nil {
		result["last_order_date"] = *uv.LastOrderDate
	}

	return result
}

// FromMap creates UserView from map data
func (uv *UserView) FromMap(data map[string]interface{}) error {
	if name, ok := data["name"].(string); ok {
		uv.Name = name
	}
	if email, ok := data["email"].(string); ok {
		uv.Email = email
	}
	if totalOrders, ok := data["total_orders"].(int); ok {
		uv.TotalOrders = totalOrders
	}
	if totalSpentStr, ok := data["total_spent"].(string); ok {
		if totalSpent, err := decimal.NewFromString(totalSpentStr); err == nil {
			uv.TotalSpent = totalSpent
		}
	}
	if isVIP, ok := data["is_vip"].(bool); ok {
		uv.IsVIP = isVIP
	}
	if createdAt, ok := data["created_at"].(time.Time); ok {
		uv.CreatedAt = createdAt
	}
	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		uv.UpdatedAt = updatedAt
	}
	if lastOrderDate, ok := data["last_order_date"].(time.Time); ok {
		uv.LastOrderDate = &lastOrderDate
	}

	return nil
}

// Validation

// Validate validates the UserView state
func (uv *UserView) Validate() error {
	if err := uv.BaseReadModel.Validate(); err != nil {
		return err
	}
	if uv.Name == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "user name cannot be empty", nil)
	}
	if uv.Email == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "user email cannot be empty", nil)
	}
	if uv.TotalOrders < 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total orders cannot be negative", nil)
	}
	if uv.TotalSpent.IsNegative() {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total spent cannot be negative", nil)
	}
	return nil
}

// Query Helpers

// IsActiveCustomer checks if user has made orders recently (within 30 days)
func (uv *UserView) IsActiveCustomer() bool {
	if uv.LastOrderDate == nil {
		return false
	}
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	return uv.LastOrderDate.After(thirtyDaysAgo)
}

// GetCustomerTier returns customer tier based on total spent
func (uv *UserView) GetCustomerTier() string {
	if uv.TotalSpent.GreaterThanOrEqual(decimal.NewFromInt(5000)) {
		return "Platinum"
	} else if uv.TotalSpent.GreaterThanOrEqual(decimal.NewFromInt(1000)) {
		return "Gold"
	} else if uv.TotalSpent.GreaterThanOrEqual(decimal.NewFromInt(100)) {
		return "Silver"
	}
	return "Bronze"
}

// GetAverageOrderValue calculates average order value
func (uv *UserView) GetAverageOrderValue() decimal.Decimal {
	if uv.TotalOrders == 0 {
		return decimal.Zero
	}
	return uv.TotalSpent.Div(decimal.NewFromInt(int64(uv.TotalOrders)))
}

// Repository Interface

// UserViewRepository defines the interface for UserView persistence
type UserViewRepository interface {
	// Save saves a UserView
	Save(ctx context.Context, userView *UserView) error

	// GetByID retrieves a UserView by ID
	GetByID(ctx context.Context, userID string) (*UserView, error)

	// GetByEmail retrieves a UserView by email
	GetByEmail(ctx context.Context, email string) (*UserView, error)

	// GetVIPUsers retrieves all VIP users
	GetVIPUsers(ctx context.Context) ([]*UserView, error)

	// GetActiveCustomers retrieves active customers (ordered within 30 days)
	GetActiveCustomers(ctx context.Context) ([]*UserView, error)

	// GetByTier retrieves users by customer tier
	GetByTier(ctx context.Context, tier string) ([]*UserView, error)

	// Delete removes a UserView
	Delete(ctx context.Context, userID string) error
}
