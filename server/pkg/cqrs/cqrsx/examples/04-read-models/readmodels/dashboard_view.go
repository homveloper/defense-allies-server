package readmodels

import (
	"context"
	"cqrs"
	"time"

	"github.com/shopspring/decimal"
)

// ProductStats represents product statistics for dashboard
type ProductStats struct {
	ProductID   string          `json:"product_id"`
	ProductName string          `json:"product_name"`
	Category    string          `json:"category"`
	UnitsSold   int             `json:"units_sold"`
	Revenue     decimal.Decimal `json:"revenue"`
	Rank        int             `json:"rank"`
}

// OrderSummary represents order summary for dashboard
type OrderSummary struct {
	OrderID      string          `json:"order_id"`
	CustomerName string          `json:"customer_name"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Status       string          `json:"status"`
	OrderDate    time.Time       `json:"order_date"`
	ItemCount    int             `json:"item_count"`
}

// CustomerStats represents customer statistics for dashboard
type CustomerStats struct {
	TotalCustomers    int             `json:"total_customers"`
	NewCustomers      int             `json:"new_customers"`    // This month
	ActiveCustomers   int             `json:"active_customers"` // Last 30 days
	VIPCustomers      int             `json:"vip_customers"`
	AverageOrderValue decimal.Decimal `json:"average_order_value"`
	CustomerRetention decimal.Decimal `json:"customer_retention"` // Percentage
}

// SalesMetrics represents sales metrics for dashboard
type SalesMetrics struct {
	TodayRevenue decimal.Decimal `json:"today_revenue"`
	WeekRevenue  decimal.Decimal `json:"week_revenue"`
	MonthRevenue decimal.Decimal `json:"month_revenue"`
	YearRevenue  decimal.Decimal `json:"year_revenue"`
	TodayOrders  int             `json:"today_orders"`
	WeekOrders   int             `json:"week_orders"`
	MonthOrders  int             `json:"month_orders"`
	YearOrders   int             `json:"year_orders"`
	GrowthRate   decimal.Decimal `json:"growth_rate"` // Month over month
}

// DashboardView represents a dashboard read model with TTL caching
type DashboardView struct {
	*cqrs.BaseReadModel
	TotalUsers      int             `json:"total_users"`
	TotalOrders     int             `json:"total_orders"`
	TotalProducts   int             `json:"total_products"`
	TotalRevenue    decimal.Decimal `json:"total_revenue"`
	SalesMetrics    SalesMetrics    `json:"sales_metrics"`
	CustomerStats   CustomerStats   `json:"customer_stats"`
	PopularProducts []ProductStats  `json:"popular_products"`
	RecentOrders    []OrderSummary  `json:"recent_orders"`
	TopCategories   []string        `json:"top_categories"`
	Alerts          []string        `json:"alerts"`
	GeneratedAt     time.Time       `json:"generated_at"`
	ExpiresAt       time.Time       `json:"expires_at"`
	RefreshInterval time.Duration   `json:"refresh_interval"`
}

// NewDashboardView creates a new DashboardView with TTL
func NewDashboardView() *DashboardView {
	now := time.Now()
	refreshInterval := 5 * time.Minute // 5 minutes TTL

	return &DashboardView{
		BaseReadModel:   cqrs.NewBaseReadModel("dashboard", "DashboardView", nil),
		TotalUsers:      0,
		TotalOrders:     0,
		TotalProducts:   0,
		TotalRevenue:    decimal.Zero,
		PopularProducts: make([]ProductStats, 0),
		RecentOrders:    make([]OrderSummary, 0),
		TopCategories:   make([]string, 0),
		Alerts:          make([]string, 0),
		GeneratedAt:     now,
		ExpiresAt:       now.Add(refreshInterval),
		RefreshInterval: refreshInterval,
	}
}

// GetTTL returns the time-to-live duration for this view
func (dv *DashboardView) GetTTL() time.Duration {
	return dv.RefreshInterval
}

// IsExpired checks if the dashboard data has expired
func (dv *DashboardView) IsExpired() bool {
	return time.Now().After(dv.ExpiresAt)
}

// Business Methods

// UpdateTotals updates the total counts and revenue
func (dv *DashboardView) UpdateTotals(totalUsers, totalOrders, totalProducts int, totalRevenue decimal.Decimal) {
	dv.TotalUsers = totalUsers
	dv.TotalOrders = totalOrders
	dv.TotalProducts = totalProducts
	dv.TotalRevenue = totalRevenue
	dv.refreshTimestamps()
}

// UpdateSalesMetrics updates sales metrics
func (dv *DashboardView) UpdateSalesMetrics(metrics SalesMetrics) {
	dv.SalesMetrics = metrics
	dv.refreshTimestamps()
}

// UpdateCustomerStats updates customer statistics
func (dv *DashboardView) UpdateCustomerStats(stats CustomerStats) {
	dv.CustomerStats = stats
	dv.refreshTimestamps()
}

// UpdatePopularProducts updates the popular products list
func (dv *DashboardView) UpdatePopularProducts(products []ProductStats) {
	dv.PopularProducts = products
	dv.refreshTimestamps()
}

// UpdateRecentOrders updates the recent orders list
func (dv *DashboardView) UpdateRecentOrders(orders []OrderSummary) {
	dv.RecentOrders = orders
	dv.refreshTimestamps()
}

// UpdateTopCategories updates the top categories list
func (dv *DashboardView) UpdateTopCategories(categories []string) {
	dv.TopCategories = categories
	dv.refreshTimestamps()
}

// AddAlert adds an alert to the dashboard
func (dv *DashboardView) AddAlert(alert string) {
	dv.Alerts = append(dv.Alerts, alert)
	dv.refreshTimestamps()
}

// ClearAlerts clears all alerts
func (dv *DashboardView) ClearAlerts() {
	dv.Alerts = make([]string, 0)
	dv.refreshTimestamps()
}

// RefreshAll refreshes all dashboard data
func (dv *DashboardView) RefreshAll(
	totalUsers, totalOrders, totalProducts int,
	totalRevenue decimal.Decimal,
	salesMetrics SalesMetrics,
	customerStats CustomerStats,
	popularProducts []ProductStats,
	recentOrders []OrderSummary,
	topCategories []string,
	alerts []string,
) {
	dv.TotalUsers = totalUsers
	dv.TotalOrders = totalOrders
	dv.TotalProducts = totalProducts
	dv.TotalRevenue = totalRevenue
	dv.SalesMetrics = salesMetrics
	dv.CustomerStats = customerStats
	dv.PopularProducts = popularProducts
	dv.RecentOrders = recentOrders
	dv.TopCategories = topCategories
	dv.Alerts = alerts
	dv.refreshTimestamps()
}

// refreshTimestamps updates generation and expiration timestamps
func (dv *DashboardView) refreshTimestamps() {
	now := time.Now()
	dv.GeneratedAt = now
	dv.ExpiresAt = now.Add(dv.RefreshInterval)
	dv.IncrementVersion()
}

// Getters

func (dv *DashboardView) GetTotalUsers() int {
	return dv.TotalUsers
}

func (dv *DashboardView) GetTotalOrders() int {
	return dv.TotalOrders
}

func (dv *DashboardView) GetTotalProducts() int {
	return dv.TotalProducts
}

func (dv *DashboardView) GetTotalRevenue() decimal.Decimal {
	return dv.TotalRevenue
}

func (dv *DashboardView) GetSalesMetrics() SalesMetrics {
	return dv.SalesMetrics
}

func (dv *DashboardView) GetCustomerStats() CustomerStats {
	return dv.CustomerStats
}

func (dv *DashboardView) GetPopularProducts() []ProductStats {
	return dv.PopularProducts
}

func (dv *DashboardView) GetRecentOrders() []OrderSummary {
	return dv.RecentOrders
}

func (dv *DashboardView) GetTopCategories() []string {
	return dv.TopCategories
}

func (dv *DashboardView) GetAlerts() []string {
	return dv.Alerts
}

func (dv *DashboardView) GetGeneratedAt() time.Time {
	return dv.GeneratedAt
}

func (dv *DashboardView) GetExpiresAt() time.Time {
	return dv.ExpiresAt
}

func (dv *DashboardView) GetRefreshInterval() time.Duration {
	return dv.RefreshInterval
}

// Helper Methods

// IncrementVersion increments the version and updates timestamp
func (dv *DashboardView) IncrementVersion() {
	dv.BaseReadModel.IncrementVersion()
}

// GetTopProductsByRevenue returns top N products by revenue
func (dv *DashboardView) GetTopProductsByRevenue(limit int) []ProductStats {
	if limit <= 0 || limit > len(dv.PopularProducts) {
		return dv.PopularProducts
	}
	return dv.PopularProducts[:limit]
}

// GetRecentOrdersByStatus returns recent orders filtered by status
func (dv *DashboardView) GetRecentOrdersByStatus(status string) []OrderSummary {
	filteredOrders := make([]OrderSummary, 0)

	for _, order := range dv.RecentOrders {
		if order.Status == status {
			filteredOrders = append(filteredOrders, order)
		}
	}

	return filteredOrders
}

// HasAlerts checks if there are any alerts
func (dv *DashboardView) HasAlerts() bool {
	return len(dv.Alerts) > 0
}

// GetAlertCount returns the number of alerts
func (dv *DashboardView) GetAlertCount() int {
	return len(dv.Alerts)
}

// GetDataFreshness returns how fresh the data is (time since generation)
func (dv *DashboardView) GetDataFreshness() time.Duration {
	return time.Since(dv.GeneratedAt)
}

// GetTimeUntilExpiry returns time until data expires
func (dv *DashboardView) GetTimeUntilExpiry() time.Duration {
	if dv.IsExpired() {
		return 0
	}
	return time.Until(dv.ExpiresAt)
}

// Validation

// Validate validates the DashboardView state
func (dv *DashboardView) Validate() error {
	if err := dv.BaseReadModel.Validate(); err != nil {
		return err
	}
	if dv.TotalUsers < 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total users cannot be negative", nil)
	}
	if dv.TotalOrders < 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total orders cannot be negative", nil)
	}
	if dv.TotalProducts < 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total products cannot be negative", nil)
	}
	if dv.TotalRevenue.IsNegative() {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total revenue cannot be negative", nil)
	}
	if dv.RefreshInterval <= 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "refresh interval must be positive", nil)
	}
	return nil
}

// Repository Interface

// DashboardViewRepository defines the interface for DashboardView persistence
type DashboardViewRepository interface {
	// Save saves a DashboardView with TTL
	Save(ctx context.Context, dashboardView *DashboardView) error

	// Get retrieves the current DashboardView
	Get(ctx context.Context) (*DashboardView, error)

	// Refresh forces a refresh of dashboard data
	Refresh(ctx context.Context) (*DashboardView, error)

	// IsExpired checks if the cached dashboard data has expired
	IsExpired(ctx context.Context) (bool, error)

	// Delete removes the DashboardView (forces refresh on next access)
	Delete(ctx context.Context) error
}
