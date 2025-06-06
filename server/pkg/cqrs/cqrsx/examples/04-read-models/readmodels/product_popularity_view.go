package readmodels

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"time"

	"github.com/shopspring/decimal"
)

// ProductSalesData represents sales data for a specific time period
type ProductSalesData struct {
	Period      string          `json:"period"`      // "daily", "weekly", "monthly"
	Date        time.Time       `json:"date"`
	UnitsSold   int             `json:"units_sold"`
	Revenue     decimal.Decimal `json:"revenue"`
	OrderCount  int             `json:"order_count"`
}

// ProductPopularityView represents a product's popularity and sales metrics
type ProductPopularityView struct {
	*cqrs.BaseReadModel
	ProductID          string             `json:"product_id"`
	ProductName        string             `json:"product_name"`
	Category           string             `json:"category"`
	CurrentPrice       decimal.Decimal    `json:"current_price"`
	TotalUnitsSold     int                `json:"total_units_sold"`
	TotalRevenue       decimal.Decimal    `json:"total_revenue"`
	TotalOrders        int                `json:"total_orders"`
	AverageOrderValue  decimal.Decimal    `json:"average_order_value"`
	AverageUnitsPerOrder decimal.Decimal  `json:"average_units_per_order"`
	PopularityScore    decimal.Decimal    `json:"popularity_score"`
	PopularityRank     int                `json:"popularity_rank"`
	LastSoldDate       *time.Time         `json:"last_sold_date,omitempty"`
	FirstSoldDate      *time.Time         `json:"first_sold_date,omitempty"`
	SalesData          []ProductSalesData `json:"sales_data"`
	Tags               []string           `json:"tags"` // "Bestseller", "Trending", "New", etc.
	IsActive           bool               `json:"is_active"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// NewProductPopularityView creates a new ProductPopularityView
func NewProductPopularityView(productID, productName, category string, currentPrice decimal.Decimal) *ProductPopularityView {
	now := time.Now()
	return &ProductPopularityView{
		BaseReadModel:        cqrs.NewBaseReadModel(productID, "ProductPopularityView", nil),
		ProductID:            productID,
		ProductName:          productName,
		Category:             category,
		CurrentPrice:         currentPrice,
		TotalUnitsSold:       0,
		TotalRevenue:         decimal.Zero,
		TotalOrders:          0,
		AverageOrderValue:    decimal.Zero,
		AverageUnitsPerOrder: decimal.Zero,
		PopularityScore:      decimal.Zero,
		PopularityRank:       0,
		SalesData:            make([]ProductSalesData, 0),
		Tags:                 make([]string, 0),
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// Business Methods

// RecordSale records a sale of this product
func (ppv *ProductPopularityView) RecordSale(unitsSold int, revenue decimal.Decimal, saleDate time.Time) {
	ppv.TotalUnitsSold += unitsSold
	ppv.TotalRevenue = ppv.TotalRevenue.Add(revenue)
	ppv.TotalOrders++
	
	// Update first/last sold dates
	if ppv.FirstSoldDate == nil || saleDate.Before(*ppv.FirstSoldDate) {
		ppv.FirstSoldDate = &saleDate
	}
	if ppv.LastSoldDate == nil || saleDate.After(*ppv.LastSoldDate) {
		ppv.LastSoldDate = &saleDate
	}
	
	ppv.recalculateMetrics()
	ppv.updateTags()
	ppv.UpdatedAt = time.Now()
	ppv.IncrementVersion()
}

// UpdateProductInfo updates product information
func (ppv *ProductPopularityView) UpdateProductInfo(name, category string, price decimal.Decimal) {
	ppv.ProductName = name
	ppv.Category = category
	ppv.CurrentPrice = price
	ppv.UpdatedAt = time.Now()
	ppv.IncrementVersion()
}

// UpdatePopularityRank updates the popularity rank
func (ppv *ProductPopularityView) UpdatePopularityRank(rank int) {
	ppv.PopularityRank = rank
	ppv.updateTags()
	ppv.UpdatedAt = time.Now()
	ppv.IncrementVersion()
}

// AddSalesData adds sales data for a specific period
func (ppv *ProductPopularityView) AddSalesData(salesData ProductSalesData) {
	// Check if data for this period already exists
	for i, data := range ppv.SalesData {
		if data.Period == salesData.Period && data.Date.Equal(salesData.Date) {
			ppv.SalesData[i] = salesData
			ppv.UpdatedAt = time.Now()
			ppv.IncrementVersion()
			return
		}
	}
	
	// Add new sales data
	ppv.SalesData = append(ppv.SalesData, salesData)
	ppv.UpdatedAt = time.Now()
	ppv.IncrementVersion()
}

// Activate activates the product
func (ppv *ProductPopularityView) Activate() {
	if !ppv.IsActive {
		ppv.IsActive = true
		ppv.updateTags()
		ppv.UpdatedAt = time.Now()
		ppv.IncrementVersion()
	}
}

// Deactivate deactivates the product
func (ppv *ProductPopularityView) Deactivate() {
	if ppv.IsActive {
		ppv.IsActive = false
		ppv.updateTags()
		ppv.UpdatedAt = time.Now()
		ppv.IncrementVersion()
	}
}

// recalculateMetrics recalculates all metrics
func (ppv *ProductPopularityView) recalculateMetrics() {
	// Calculate average order value
	if ppv.TotalOrders > 0 {
		ppv.AverageOrderValue = ppv.TotalRevenue.Div(decimal.NewFromInt(int64(ppv.TotalOrders)))
		ppv.AverageUnitsPerOrder = decimal.NewFromInt(int64(ppv.TotalUnitsSold)).Div(decimal.NewFromInt(int64(ppv.TotalOrders)))
	} else {
		ppv.AverageOrderValue = decimal.Zero
		ppv.AverageUnitsPerOrder = decimal.Zero
	}
	
	// Calculate popularity score (weighted combination of units sold, revenue, and recency)
	ppv.calculatePopularityScore()
}

// calculatePopularityScore calculates a popularity score based on various factors
func (ppv *ProductPopularityView) calculatePopularityScore() {
	// Base score from units sold (normalized)
	unitsScore := decimal.NewFromInt(int64(ppv.TotalUnitsSold))
	
	// Revenue score (normalized)
	revenueScore := ppv.TotalRevenue.Div(decimal.NewFromInt(100)) // Divide by 100 to normalize
	
	// Recency score (higher for recently sold items)
	recencyScore := decimal.Zero
	if ppv.LastSoldDate != nil {
		daysSinceLastSale := time.Since(*ppv.LastSoldDate).Hours() / 24
		if daysSinceLastSale <= 30 {
			recencyScore = decimal.NewFromFloat(30 - daysSinceLastSale)
		}
	}
	
	// Weighted combination
	ppv.PopularityScore = unitsScore.Mul(decimal.NewFromFloat(0.4)).
		Add(revenueScore.Mul(decimal.NewFromFloat(0.4))).
		Add(recencyScore.Mul(decimal.NewFromFloat(0.2)))
}

// updateTags updates product tags based on metrics
func (ppv *ProductPopularityView) updateTags() {
	ppv.Tags = make([]string, 0)
	
	// Bestseller tag (top 10 rank)
	if ppv.PopularityRank > 0 && ppv.PopularityRank <= 10 {
		ppv.Tags = append(ppv.Tags, "Bestseller")
	}
	
	// Trending tag (sold recently and high popularity score)
	if ppv.LastSoldDate != nil {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7)
		if ppv.LastSoldDate.After(sevenDaysAgo) && ppv.PopularityScore.GreaterThan(decimal.NewFromInt(100)) {
			ppv.Tags = append(ppv.Tags, "Trending")
		}
	}
	
	// High value tag (high average order value)
	if ppv.AverageOrderValue.GreaterThan(decimal.NewFromInt(100)) {
		ppv.Tags = append(ppv.Tags, "HighValue")
	}
	
	// New tag (first sold within 30 days)
	if ppv.FirstSoldDate != nil {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		if ppv.FirstSoldDate.After(thirtyDaysAgo) {
			ppv.Tags = append(ppv.Tags, "New")
		}
	}
	
	// Active tag
	if ppv.IsActive {
		ppv.Tags = append(ppv.Tags, "Active")
	}
	
	// Popular tag (high total units sold)
	if ppv.TotalUnitsSold > 100 {
		ppv.Tags = append(ppv.Tags, "Popular")
	}
}

// Getters

func (ppv *ProductPopularityView) GetProductID() string {
	return ppv.ProductID
}

func (ppv *ProductPopularityView) GetProductName() string {
	return ppv.ProductName
}

func (ppv *ProductPopularityView) GetCategory() string {
	return ppv.Category
}

func (ppv *ProductPopularityView) GetCurrentPrice() decimal.Decimal {
	return ppv.CurrentPrice
}

func (ppv *ProductPopularityView) GetTotalUnitsSold() int {
	return ppv.TotalUnitsSold
}

func (ppv *ProductPopularityView) GetTotalRevenue() decimal.Decimal {
	return ppv.TotalRevenue
}

func (ppv *ProductPopularityView) GetTotalOrders() int {
	return ppv.TotalOrders
}

func (ppv *ProductPopularityView) GetAverageOrderValue() decimal.Decimal {
	return ppv.AverageOrderValue
}

func (ppv *ProductPopularityView) GetAverageUnitsPerOrder() decimal.Decimal {
	return ppv.AverageUnitsPerOrder
}

func (ppv *ProductPopularityView) GetPopularityScore() decimal.Decimal {
	return ppv.PopularityScore
}

func (ppv *ProductPopularityView) GetPopularityRank() int {
	return ppv.PopularityRank
}

func (ppv *ProductPopularityView) GetLastSoldDate() *time.Time {
	return ppv.LastSoldDate
}

func (ppv *ProductPopularityView) GetFirstSoldDate() *time.Time {
	return ppv.FirstSoldDate
}

func (ppv *ProductPopularityView) GetSalesData() []ProductSalesData {
	return ppv.SalesData
}

func (ppv *ProductPopularityView) GetTags() []string {
	return ppv.Tags
}

func (ppv *ProductPopularityView) GetIsActive() bool {
	return ppv.IsActive
}

func (ppv *ProductPopularityView) GetCreatedAt() time.Time {
	return ppv.CreatedAt
}

func (ppv *ProductPopularityView) GetUpdatedAt() time.Time {
	return ppv.UpdatedAt
}

// Helper Methods

// IncrementVersion increments the version and updates timestamp
func (ppv *ProductPopularityView) IncrementVersion() {
	ppv.BaseReadModel.IncrementVersion()
	ppv.UpdatedAt = time.Now()
}

// GetSalesDataByPeriod returns sales data for a specific period
func (ppv *ProductPopularityView) GetSalesDataByPeriod(period string) []ProductSalesData {
	filteredData := make([]ProductSalesData, 0)
	
	for _, data := range ppv.SalesData {
		if data.Period == period {
			filteredData = append(filteredData, data)
		}
	}
	
	return filteredData
}

// HasTag checks if product has a specific tag
func (ppv *ProductPopularityView) HasTag(tag string) bool {
	for _, t := range ppv.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// IsRecentlySold checks if product was sold within the last N days
func (ppv *ProductPopularityView) IsRecentlySold(days int) bool {
	if ppv.LastSoldDate == nil {
		return false
	}
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return ppv.LastSoldDate.After(cutoffDate)
}

// Validation

// Validate validates the ProductPopularityView state
func (ppv *ProductPopularityView) Validate() error {
	if err := ppv.BaseReadModel.Validate(); err != nil {
		return err
	}
	if ppv.ProductID == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "product ID cannot be empty", nil)
	}
	if ppv.ProductName == "" {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "product name cannot be empty", nil)
	}
	if ppv.TotalUnitsSold < 0 {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total units sold cannot be negative", nil)
	}
	if ppv.TotalRevenue.IsNegative() {
		return cqrs.NewCQRSError("VALIDATION_ERROR", "total revenue cannot be negative", nil)
	}
	return nil
}

// Repository Interface

// ProductPopularityViewRepository defines the interface for ProductPopularityView persistence
type ProductPopularityViewRepository interface {
	// Save saves a ProductPopularityView
	Save(ctx context.Context, popularityView *ProductPopularityView) error
	
	// GetByID retrieves a ProductPopularityView by product ID
	GetByID(ctx context.Context, productID string) (*ProductPopularityView, error)
	
	// GetByCategory retrieves ProductPopularityViews by category
	GetByCategory(ctx context.Context, category string) ([]*ProductPopularityView, error)
	
	// GetByTag retrieves ProductPopularityViews by tag
	GetByTag(ctx context.Context, tag string) ([]*ProductPopularityView, error)
	
	// GetTopProducts retrieves top products by popularity score
	GetTopProducts(ctx context.Context, limit int) ([]*ProductPopularityView, error)
	
	// GetTrendingProducts retrieves trending products
	GetTrendingProducts(ctx context.Context) ([]*ProductPopularityView, error)
	
	// GetBestsellers retrieves bestseller products
	GetBestsellers(ctx context.Context) ([]*ProductPopularityView, error)
	
	// UpdateRankings updates popularity rankings for all products
	UpdateRankings(ctx context.Context) error
	
	// Delete removes a ProductPopularityView
	Delete(ctx context.Context, productID string) error
}
