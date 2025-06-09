package demo

import (
	"context"
	"cqrs"
	"cqrs/cqrsx/examples/04-read-models/readmodels"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// QueryExamples demonstrates various read model queries
type QueryExamples struct {
	readStore cqrs.ReadStore
}

// NewQueryExamples creates a new QueryExamples instance
func NewQueryExamples(readStore cqrs.ReadStore) *QueryExamples {
	return &QueryExamples{
		readStore: readStore,
	}
}

// RunAllExamples runs all query examples
func (qe *QueryExamples) RunAllExamples(ctx context.Context) error {
	log.Printf("=== Running Read Model Query Examples ===")

	examples := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"User View Queries", qe.demonstrateUserViewQueries},
		{"Order Summary Queries", qe.demonstrateOrderSummaryQueries},
		{"Customer History Queries", qe.demonstrateCustomerHistoryQueries},
		{"Product Popularity Queries", qe.demonstrateProductPopularityQueries},
		{"Dashboard Queries", qe.demonstrateDashboardQueries},
		{"Complex Analytics Queries", qe.demonstrateComplexQueries},
	}

	for _, example := range examples {
		log.Printf("\n--- %s ---", example.name)
		if err := example.fn(ctx); err != nil {
			log.Printf("Error in %s: %v", example.name, err)
		}
	}

	log.Printf("\n=== Query Examples Completed ===")
	return nil
}

// demonstrateUserViewQueries shows UserView query examples
func (qe *QueryExamples) demonstrateUserViewQueries(ctx context.Context) error {
	log.Printf("Demonstrating UserView queries...")

	// Example 1: Get user by ID
	userID := "user-123"
	userView, err := qe.getUserView(ctx, userID)
	if err != nil {
		log.Printf("User not found: %v", err)
		// Create a sample user for demonstration
		userView = qe.createSampleUserView(userID)
		if err := qe.readStore.Save(ctx, userView); err != nil {
			return fmt.Errorf("failed to save sample user: %w", err)
		}
	}

	log.Printf("User: %s (%s)", userView.GetName(), userView.GetEmail())
	log.Printf("Total Orders: %d, Total Spent: $%s", userView.GetTotalOrders(), userView.GetTotalSpent().String())
	log.Printf("VIP Status: %t, Customer Tier: %s", userView.GetIsVIP(), userView.GetCustomerTier())

	// Example 2: Demonstrate user statistics
	log.Printf("Average Order Value: $%s", userView.GetAverageOrderValue().String())
	log.Printf("Is Active Customer: %t", userView.IsActiveCustomer())

	return nil
}

// demonstrateOrderSummaryQueries shows OrderSummaryView query examples
func (qe *QueryExamples) demonstrateOrderSummaryQueries(ctx context.Context) error {
	log.Printf("Demonstrating OrderSummaryView queries...")

	// Create a sample order for demonstration
	orderID := "order-456"
	orderView := qe.createSampleOrderSummaryView(orderID)
	if qe.readStore != nil {
		if err := qe.readStore.Save(ctx, orderView); err != nil {
			return fmt.Errorf("failed to save sample order: %w", err)
		}
	}

	// Example 1: Get order by ID
	retrievedOrder, err := qe.getOrderSummaryView(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}

	log.Printf("Order ID: %s", retrievedOrder.GetID())
	log.Printf("Customer: %s (%s)", retrievedOrder.GetCustomerName(), retrievedOrder.GetCustomerEmail())
	log.Printf("Status: %s", retrievedOrder.GetStatus())
	log.Printf("Total Amount: $%s", retrievedOrder.GetTotalAmount().String())
	log.Printf("Item Count: %d", retrievedOrder.GetItemCount())
	log.Printf("Unique Products: %d", retrievedOrder.GetUniqueProductCount())

	// Example 2: Demonstrate order analysis
	log.Printf("Is Completed: %t", retrievedOrder.IsCompleted())
	log.Printf("Is Pending: %t", retrievedOrder.IsPending())

	return nil
}

// demonstrateCustomerHistoryQueries shows CustomerOrderHistoryView query examples
func (qe *QueryExamples) demonstrateCustomerHistoryQueries(ctx context.Context) error {
	log.Printf("Demonstrating CustomerOrderHistoryView queries...")

	// Create a sample customer history for demonstration
	customerID := "customer-789"
	historyView := qe.createSampleCustomerHistoryView(customerID)
	if qe.readStore != nil {
		if err := qe.readStore.Save(ctx, historyView); err != nil {
			return fmt.Errorf("failed to save sample customer history: %w", err)
		}
	}

	// Example 1: Get customer history
	retrievedHistory, err := qe.getCustomerHistoryView(ctx, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer history: %w", err)
	}

	log.Printf("Customer: %s (%s)", retrievedHistory.GetCustomerName(), retrievedHistory.GetCustomerEmail())
	log.Printf("Total Orders: %d (Completed: %d, Cancelled: %d)",
		retrievedHistory.GetTotalOrders(),
		retrievedHistory.GetCompletedOrders(),
		retrievedHistory.GetCancelledOrders())
	log.Printf("Total Spent: $%s", retrievedHistory.GetTotalSpent().String())
	log.Printf("Average Order: $%s", retrievedHistory.GetAverageOrder().String())
	log.Printf("Customer Tier: %s", retrievedHistory.GetCustomerTier())
	log.Printf("Is Active: %t", retrievedHistory.GetIsActive())
	log.Printf("Tags: %v", retrievedHistory.GetTags())

	// Example 2: Get recent orders
	recentOrders := retrievedHistory.GetRecentOrders(30) // Last 30 days
	log.Printf("Recent Orders (last 30 days): %d", len(recentOrders))

	// Example 3: Get orders by status
	completedOrders := retrievedHistory.GetOrdersByStatus("completed")
	log.Printf("Completed Orders: %d", len(completedOrders))

	return nil
}

// demonstrateProductPopularityQueries shows ProductPopularityView query examples
func (qe *QueryExamples) demonstrateProductPopularityQueries(ctx context.Context) error {
	log.Printf("Demonstrating ProductPopularityView queries...")

	// Create a sample product popularity view for demonstration
	productID := "product-101"
	popularityView := qe.createSampleProductPopularityView(productID)
	if qe.readStore != nil {
		if err := qe.readStore.Save(ctx, popularityView); err != nil {
			return fmt.Errorf("failed to save sample product popularity: %w", err)
		}
	}

	// Example 1: Get product popularity
	retrievedPopularity, err := qe.getProductPopularityView(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get product popularity: %w", err)
	}

	log.Printf("Product: %s (Category: %s)", retrievedPopularity.GetProductName(), retrievedPopularity.GetCategory())
	log.Printf("Current Price: $%s", retrievedPopularity.GetCurrentPrice().String())
	log.Printf("Total Units Sold: %d", retrievedPopularity.GetTotalUnitsSold())
	log.Printf("Total Revenue: $%s", retrievedPopularity.GetTotalRevenue().String())
	log.Printf("Total Orders: %d", retrievedPopularity.GetTotalOrders())
	log.Printf("Average Order Value: $%s", retrievedPopularity.GetAverageOrderValue().String())
	log.Printf("Average Units Per Order: %s", retrievedPopularity.GetAverageUnitsPerOrder().String())
	log.Printf("Popularity Score: %s", retrievedPopularity.GetPopularityScore().String())
	log.Printf("Popularity Rank: %d", retrievedPopularity.GetPopularityRank())
	log.Printf("Tags: %v", retrievedPopularity.GetTags())

	// Example 2: Check product status
	log.Printf("Is Recently Sold (7 days): %t", retrievedPopularity.IsRecentlySold(7))
	log.Printf("Has Bestseller Tag: %t", retrievedPopularity.HasTag("Bestseller"))

	return nil
}

// demonstrateDashboardQueries shows DashboardView query examples
func (qe *QueryExamples) demonstrateDashboardQueries(ctx context.Context) error {
	log.Printf("Demonstrating DashboardView queries...")

	// Create a sample dashboard view for demonstration
	dashboardView := qe.createSampleDashboardView()
	if qe.readStore != nil {
		if err := qe.readStore.Save(ctx, dashboardView); err != nil {
			return fmt.Errorf("failed to save sample dashboard: %w", err)
		}
	}

	// Example 1: Get dashboard data
	retrievedDashboard, err := qe.getDashboardView(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dashboard: %w", err)
	}

	log.Printf("=== Dashboard Overview ===")
	log.Printf("Total Users: %d", retrievedDashboard.GetTotalUsers())
	log.Printf("Total Orders: %d", retrievedDashboard.GetTotalOrders())
	log.Printf("Total Products: %d", retrievedDashboard.GetTotalProducts())
	log.Printf("Total Revenue: $%s", retrievedDashboard.GetTotalRevenue().String())

	// Example 2: Sales metrics
	salesMetrics := retrievedDashboard.GetSalesMetrics()
	log.Printf("=== Sales Metrics ===")
	log.Printf("Today Revenue: $%s", salesMetrics.TodayRevenue.String())
	log.Printf("Week Revenue: $%s", salesMetrics.WeekRevenue.String())
	log.Printf("Month Revenue: $%s", salesMetrics.MonthRevenue.String())
	log.Printf("Growth Rate: %s%%", salesMetrics.GrowthRate.String())

	// Example 3: Customer stats
	customerStats := retrievedDashboard.GetCustomerStats()
	log.Printf("=== Customer Statistics ===")
	log.Printf("Total Customers: %d", customerStats.TotalCustomers)
	log.Printf("New Customers: %d", customerStats.NewCustomers)
	log.Printf("Active Customers: %d", customerStats.ActiveCustomers)
	log.Printf("VIP Customers: %d", customerStats.VIPCustomers)

	// Example 4: Popular products
	popularProducts := retrievedDashboard.GetTopProductsByRevenue(3)
	log.Printf("=== Top Products ===")
	for i, product := range popularProducts {
		log.Printf("%d. %s - $%s (%d units)", i+1, product.ProductName, product.Revenue.String(), product.UnitsSold)
	}

	// Example 5: Dashboard freshness
	log.Printf("=== Dashboard Status ===")
	log.Printf("Generated At: %s", retrievedDashboard.GetGeneratedAt().Format(time.RFC3339))
	log.Printf("Data Freshness: %s", retrievedDashboard.GetDataFreshness().String())
	log.Printf("Is Expired: %t", retrievedDashboard.IsExpired())
	log.Printf("Time Until Expiry: %s", retrievedDashboard.GetTimeUntilExpiry().String())
	log.Printf("Has Alerts: %t", retrievedDashboard.HasAlerts())

	return nil
}

// demonstrateComplexQueries shows complex analytical queries
func (qe *QueryExamples) demonstrateComplexQueries(ctx context.Context) error {
	log.Printf("Demonstrating complex analytical queries...")

	// This would demonstrate more complex queries that combine multiple read models
	// For now, we'll simulate some analytical insights

	log.Printf("=== Customer Segmentation Analysis ===")
	log.Printf("VIP Customers: 15 (12% of total)")
	log.Printf("Active Customers: 89 (71% of total)")
	log.Printf("New Customers (this month): 23")
	log.Printf("At-risk Customers: 8 (no orders in 60 days)")

	log.Printf("=== Product Performance Analysis ===")
	log.Printf("Best Performing Category: Electronics (45% of revenue)")
	log.Printf("Trending Products: 12 products with >50% growth")
	log.Printf("Underperforming Products: 5 products with <10 sales")

	log.Printf("=== Revenue Analysis ===")
	log.Printf("Month-over-Month Growth: +15.3%")
	log.Printf("Average Order Value: $127.50")
	log.Printf("Customer Lifetime Value: $456.78")

	return nil
}

// Helper methods to create sample data

func (qe *QueryExamples) createSampleUserView(userID string) *readmodels.UserView {
	userView := readmodels.NewUserView(userID, "John Doe", "john.doe@example.com")

	// Simulate some order history
	userView.RecordOrder(decimal.NewFromInt(150), time.Now().AddDate(0, 0, -10))
	userView.RecordOrder(decimal.NewFromInt(89), time.Now().AddDate(0, 0, -5))
	userView.RecordOrder(decimal.NewFromInt(299), time.Now().AddDate(0, 0, -2))

	return userView
}

func (qe *QueryExamples) createSampleOrderSummaryView(orderID string) *readmodels.OrderSummaryView {
	items := []readmodels.OrderItemView{
		{
			ProductID:   "prod-1",
			ProductName: "Laptop",
			Price:       decimal.NewFromInt(999),
			Quantity:    1,
			SubTotal:    decimal.NewFromInt(999),
		},
		{
			ProductID:   "prod-2",
			ProductName: "Mouse",
			Price:       decimal.NewFromInt(29),
			Quantity:    2,
			SubTotal:    decimal.NewFromInt(58),
		},
	}

	orderView := readmodels.NewOrderSummaryView(
		orderID,
		"customer-123",
		"Jane Smith",
		"jane.smith@example.com",
		items,
		decimal.NewFromInt(1057),
		decimal.NewFromInt(105),
		decimal.NewFromInt(1162),
		time.Now().AddDate(0, 0, -1),
	)

	orderView.CompleteOrder(time.Now())
	return orderView
}

func (qe *QueryExamples) createSampleCustomerHistoryView(customerID string) *readmodels.CustomerOrderHistoryView {
	historyView := readmodels.NewCustomerOrderHistoryView(customerID, "Alice Johnson", "alice.johnson@example.com")

	// Add some order history
	orders := []readmodels.OrderHistoryItem{
		{
			OrderID:     "order-1",
			OrderDate:   time.Now().AddDate(0, 0, -30),
			TotalAmount: decimal.NewFromInt(150),
			ItemCount:   2,
			Status:      "completed",
			CompletedAt: &[]time.Time{time.Now().AddDate(0, 0, -29)}[0],
		},
		{
			OrderID:     "order-2",
			OrderDate:   time.Now().AddDate(0, 0, -15),
			TotalAmount: decimal.NewFromInt(299),
			ItemCount:   1,
			Status:      "completed",
			CompletedAt: &[]time.Time{time.Now().AddDate(0, 0, -14)}[0],
		},
		{
			OrderID:     "order-3",
			OrderDate:   time.Now().AddDate(0, 0, -5),
			TotalAmount: decimal.NewFromInt(89),
			ItemCount:   3,
			Status:      "pending",
		},
	}

	for _, order := range orders {
		historyView.AddOrder(order)
	}

	return historyView
}

func (qe *QueryExamples) createSampleProductPopularityView(productID string) *readmodels.ProductPopularityView {
	popularityView := readmodels.NewProductPopularityView(
		productID,
		"Gaming Laptop Pro",
		"Electronics",
		decimal.NewFromInt(1299),
	)

	// Simulate some sales
	popularityView.RecordSale(5, decimal.NewFromInt(6495), time.Now().AddDate(0, 0, -20))
	popularityView.RecordSale(3, decimal.NewFromInt(3897), time.Now().AddDate(0, 0, -10))
	popularityView.RecordSale(2, decimal.NewFromInt(2598), time.Now().AddDate(0, 0, -3))

	popularityView.UpdatePopularityRank(5)

	return popularityView
}

func (qe *QueryExamples) createSampleDashboardView() *readmodels.DashboardView {
	dashboardView := readmodels.NewDashboardView()

	// Sample metrics
	salesMetrics := readmodels.SalesMetrics{
		TodayRevenue: decimal.NewFromInt(1250),
		WeekRevenue:  decimal.NewFromInt(8900),
		MonthRevenue: decimal.NewFromInt(35600),
		YearRevenue:  decimal.NewFromInt(425000),
		TodayOrders:  12,
		WeekOrders:   89,
		MonthOrders:  356,
		YearOrders:   4250,
		GrowthRate:   decimal.NewFromFloat(15.3),
	}

	customerStats := readmodels.CustomerStats{
		TotalCustomers:    125,
		NewCustomers:      23,
		ActiveCustomers:   89,
		VIPCustomers:      15,
		AverageOrderValue: decimal.NewFromFloat(127.50),
		CustomerRetention: decimal.NewFromFloat(78.5),
	}

	popularProducts := []readmodels.ProductStats{
		{ProductID: "p1", ProductName: "Gaming Laptop", Category: "Electronics", UnitsSold: 45, Revenue: decimal.NewFromInt(58500), Rank: 1},
		{ProductID: "p2", ProductName: "Wireless Headphones", Category: "Electronics", UnitsSold: 89, Revenue: decimal.NewFromInt(13350), Rank: 2},
		{ProductID: "p3", ProductName: "Office Chair", Category: "Furniture", UnitsSold: 23, Revenue: decimal.NewFromInt(4577), Rank: 3},
	}

	recentOrders := []readmodels.OrderSummary{
		{OrderID: "o1", CustomerName: "John Doe", TotalAmount: decimal.NewFromInt(299), Status: "completed", OrderDate: time.Now().AddDate(0, 0, -1), ItemCount: 2},
		{OrderID: "o2", CustomerName: "Jane Smith", TotalAmount: decimal.NewFromInt(150), Status: "pending", OrderDate: time.Now(), ItemCount: 1},
	}

	topCategories := []string{"Electronics", "Furniture", "Sports", "Books"}
	alerts := []string{"Low inventory: Gaming Mouse", "High return rate: Bluetooth Speaker"}

	dashboardView.RefreshAll(
		125,                        // totalUsers
		356,                        // totalOrders
		89,                         // totalProducts
		decimal.NewFromInt(425000), // totalRevenue
		salesMetrics,
		customerStats,
		popularProducts,
		recentOrders,
		topCategories,
		alerts,
	)

	return dashboardView
}

// Helper methods to retrieve read models

func (qe *QueryExamples) getUserView(ctx context.Context, userID string) (*readmodels.UserView, error) {
	if qe.readStore == nil {
		// Return a mock UserView for demo purposes
		return readmodels.NewUserView(userID, "Demo User", "demo@example.com"), nil
	}

	readModel, err := qe.readStore.GetByID(ctx, userID, "UserView")
	if err != nil {
		return nil, err
	}

	userView, ok := readModel.(*readmodels.UserView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *UserView, got %T", readModel)
	}

	return userView, nil
}

func (qe *QueryExamples) getOrderSummaryView(ctx context.Context, orderID string) (*readmodels.OrderSummaryView, error) {
	if qe.readStore == nil {
		// Return a mock OrderSummaryView for demo purposes
		items := []readmodels.OrderItemView{}
		return readmodels.NewOrderSummaryView(orderID, "user-1", "Demo User", "demo@example.com", items, decimal.NewFromInt(100), decimal.NewFromInt(10), decimal.NewFromInt(110), time.Now()), nil
	}

	readModel, err := qe.readStore.GetByID(ctx, orderID, "OrderSummaryView")
	if err != nil {
		return nil, err
	}

	orderView, ok := readModel.(*readmodels.OrderSummaryView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *OrderSummaryView, got %T", readModel)
	}

	return orderView, nil
}

func (qe *QueryExamples) getCustomerHistoryView(ctx context.Context, customerID string) (*readmodels.CustomerOrderHistoryView, error) {
	if qe.readStore == nil {
		// Return a mock CustomerOrderHistoryView for demo purposes
		return readmodels.NewCustomerOrderHistoryView(customerID, "Demo User", "demo@example.com"), nil
	}

	readModel, err := qe.readStore.GetByID(ctx, customerID, "CustomerOrderHistoryView")
	if err != nil {
		return nil, err
	}

	historyView, ok := readModel.(*readmodels.CustomerOrderHistoryView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *CustomerOrderHistoryView, got %T", readModel)
	}

	return historyView, nil
}

func (qe *QueryExamples) getProductPopularityView(ctx context.Context, productID string) (*readmodels.ProductPopularityView, error) {
	if qe.readStore == nil {
		// Return a mock ProductPopularityView for demo purposes
		return readmodels.NewProductPopularityView(productID, "Demo Product", "Electronics", decimal.NewFromInt(99)), nil
	}

	readModel, err := qe.readStore.GetByID(ctx, productID, "ProductPopularityView")
	if err != nil {
		return nil, err
	}

	popularityView, ok := readModel.(*readmodels.ProductPopularityView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *ProductPopularityView, got %T", readModel)
	}

	return popularityView, nil
}

func (qe *QueryExamples) getDashboardView(ctx context.Context) (*readmodels.DashboardView, error) {
	if qe.readStore == nil {
		// Return a mock DashboardView for demo purposes
		return readmodels.NewDashboardView(), nil
	}

	readModel, err := qe.readStore.GetByID(ctx, "dashboard", "DashboardView")
	if err != nil {
		return nil, err
	}

	dashboardView, ok := readModel.(*readmodels.DashboardView)
	if !ok {
		return nil, fmt.Errorf("invalid read model type: expected *DashboardView, got %T", readModel)
	}

	return dashboardView, nil
}
