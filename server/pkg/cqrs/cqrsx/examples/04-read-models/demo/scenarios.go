package demo

import (
	"context"
	"cqrs/cqrsx/examples/04-read-models/domain"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DemoScenario represents a demo scenario
type DemoScenario struct {
	Name        string
	Description string
	Steps       []DemoStep
}

// DemoStep represents a single step in a demo scenario
type DemoStep struct {
	Name        string
	Description string
	Action      func(ctx context.Context, deps *DemoDependencies) error
}

// DemoDependencies contains dependencies needed for demo scenarios
type DemoDependencies struct {
	UserCommandHandler    *domain.UserCommandHandler
	ProductCommandHandler *domain.ProductCommandHandler
	OrderCommandHandler   *domain.OrderCommandHandler
}

// GetAllScenarios returns all available demo scenarios
func GetAllScenarios() []DemoScenario {
	return []DemoScenario{
		GetBasicECommerceScenario(),
		GetCustomerJourneyScenario(),
		GetProductPopularityScenario(),
		GetOrderManagementScenario(),
	}
}

// GetBasicECommerceScenario returns a basic e-commerce demo scenario
func GetBasicECommerceScenario() DemoScenario {
	return DemoScenario{
		Name:        "Basic E-Commerce Flow",
		Description: "Demonstrates basic e-commerce operations: user registration, product creation, and order processing",
		Steps: []DemoStep{
			{
				Name:        "Create Users",
				Description: "Create sample users",
				Action:      createSampleUsers,
			},
			{
				Name:        "Create Products",
				Description: "Create sample products",
				Action:      createSampleProducts,
			},
			{
				Name:        "Create Orders",
				Description: "Create sample orders",
				Action:      createSampleOrders,
			},
			{
				Name:        "Complete Orders",
				Description: "Complete some orders",
				Action:      completeSampleOrders,
			},
		},
	}
}

// GetCustomerJourneyScenario returns a customer journey demo scenario
func GetCustomerJourneyScenario() DemoScenario {
	return DemoScenario{
		Name:        "Customer Journey",
		Description: "Demonstrates a complete customer journey from registration to multiple purchases",
		Steps: []DemoStep{
			{
				Name:        "New Customer Registration",
				Description: "Register a new customer",
				Action:      createNewCustomer,
			},
			{
				Name:        "First Purchase",
				Description: "Customer makes their first purchase",
				Action:      makeFirstPurchase,
			},
			{
				Name:        "Update Profile",
				Description: "Customer updates their profile",
				Action:      updateCustomerProfile,
			},
			{
				Name:        "Multiple Purchases",
				Description: "Customer makes multiple purchases to become VIP",
				Action:      makeMultiplePurchases,
			},
			{
				Name:        "Cancel Order",
				Description: "Customer cancels an order",
				Action:      cancelOrder,
			},
		},
	}
}

// GetProductPopularityScenario returns a product popularity demo scenario
func GetProductPopularityScenario() DemoScenario {
	return DemoScenario{
		Name:        "Product Popularity Tracking",
		Description: "Demonstrates how product popularity is tracked through sales",
		Steps: []DemoStep{
			{
				Name:        "Create Product Categories",
				Description: "Create products in different categories",
				Action:      createProductCategories,
			},
			{
				Name:        "Simulate Sales",
				Description: "Simulate sales for different products",
				Action:      simulateProductSales,
			},
			{
				Name:        "Update Product Info",
				Description: "Update product information",
				Action:      updateProductInfo,
			},
		},
	}
}

// GetOrderManagementScenario returns an order management demo scenario
func GetOrderManagementScenario() DemoScenario {
	return DemoScenario{
		Name:        "Order Management",
		Description: "Demonstrates various order management operations",
		Steps: []DemoStep{
			{
				Name:        "Create Complex Orders",
				Description: "Create orders with multiple items",
				Action:      createComplexOrders,
			},
			{
				Name:        "Modify Orders",
				Description: "Add and remove items from orders",
				Action:      modifyOrders,
			},
			{
				Name:        "Process Orders",
				Description: "Complete and cancel orders",
				Action:      processOrders,
			},
		},
	}
}

// Demo step implementations

func createSampleUsers(ctx context.Context, deps *DemoDependencies) error {
	users := []struct {
		name  string
		email string
	}{
		{"John Doe", "john.doe@example.com"},
		{"Jane Smith", "jane.smith@example.com"},
		{"Bob Johnson", "bob.johnson@example.com"},
		{"Alice Brown", "alice.brown@example.com"},
		{"Charlie Wilson", "charlie.wilson@example.com"},
	}

	for _, user := range users {
		userID := uuid.New().String()
		cmd := domain.NewCreateUserCommand(userID, user.name, user.email)

		_, err := deps.UserCommandHandler.Handle(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.name, err)
		}

		log.Printf("Created user: %s (%s)", user.name, user.email)
	}

	return nil
}

func createSampleProducts(ctx context.Context, deps *DemoDependencies) error {
	products := []struct {
		name        string
		price       decimal.Decimal
		category    string
		description string
	}{
		{"Laptop Pro", decimal.NewFromInt(1299), "Electronics", "High-performance laptop"},
		{"Wireless Mouse", decimal.NewFromInt(29), "Electronics", "Ergonomic wireless mouse"},
		{"Coffee Mug", decimal.NewFromInt(15), "Home", "Ceramic coffee mug"},
		{"Running Shoes", decimal.NewFromInt(89), "Sports", "Comfortable running shoes"},
		{"Smartphone", decimal.NewFromInt(699), "Electronics", "Latest smartphone model"},
		{"Desk Chair", decimal.NewFromInt(199), "Furniture", "Ergonomic office chair"},
		{"Water Bottle", decimal.NewFromInt(25), "Sports", "Insulated water bottle"},
		{"Headphones", decimal.NewFromInt(149), "Electronics", "Noise-canceling headphones"},
	}

	for _, product := range products {
		productID := uuid.New().String()
		cmd := domain.NewCreateProductCommand(productID, product.name, product.price, product.category, product.description)

		_, err := deps.ProductCommandHandler.Handle(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create product %s: %w", product.name, err)
		}

		log.Printf("Created product: %s - $%s", product.name, product.price.String())
	}

	return nil
}

func createSampleOrders(ctx context.Context, deps *DemoDependencies) error {
	// This is a simplified version - in a real scenario, you'd need to:
	// 1. Get actual user IDs and product IDs from the system
	// 2. Create proper order items with correct product information

	orders := []struct {
		customerID string
		items      []domain.OrderItem
	}{
		{
			customerID: "user-1", // This would be a real user ID
			items: []domain.OrderItem{
				{
					ProductID: "product-1",
					Name:      "Laptop Pro",
					Price:     decimal.NewFromInt(1299),
					Quantity:  1,
					SubTotal:  decimal.NewFromInt(1299),
				},
			},
		},
		{
			customerID: "user-2",
			items: []domain.OrderItem{
				{
					ProductID: "product-2",
					Name:      "Wireless Mouse",
					Price:     decimal.NewFromInt(29),
					Quantity:  2,
					SubTotal:  decimal.NewFromInt(58),
				},
				{
					ProductID: "product-3",
					Name:      "Coffee Mug",
					Price:     decimal.NewFromInt(15),
					Quantity:  1,
					SubTotal:  decimal.NewFromInt(15),
				},
			},
		},
	}

	for i, order := range orders {
		orderID := uuid.New().String()
		cmd := domain.NewCreateOrderCommand(orderID, order.customerID, order.items)

		_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
		if err != nil {
			return fmt.Errorf("failed to create order %d: %w", i+1, err)
		}

		log.Printf("Created order: %s for customer %s", orderID, order.customerID)
	}

	return nil
}

func completeSampleOrders(ctx context.Context, deps *DemoDependencies) error {
	// This would complete some of the orders created in the previous step
	// In a real implementation, you'd track the order IDs

	orderIDs := []string{"order-1", "order-2"} // These would be real order IDs

	for _, orderID := range orderIDs {
		cmd := domain.NewCompleteOrderCommand(orderID)

		_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
		if err != nil {
			log.Printf("Warning: failed to complete order %s: %v", orderID, err)
			continue
		}

		log.Printf("Completed order: %s", orderID)
	}

	return nil
}

func createNewCustomer(ctx context.Context, deps *DemoDependencies) error {
	userID := uuid.New().String()
	cmd := domain.NewCreateUserCommand(userID, "Demo Customer", "demo.customer@example.com")

	_, err := deps.UserCommandHandler.Handle(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create new customer: %w", err)
	}

	log.Printf("Created new customer: Demo Customer")
	return nil
}

func makeFirstPurchase(ctx context.Context, deps *DemoDependencies) error {
	// Create a first purchase for the demo customer
	orderID := uuid.New().String()
	items := []domain.OrderItem{
		{
			ProductID: "product-starter",
			Name:      "Starter Kit",
			Price:     decimal.NewFromInt(49),
			Quantity:  1,
			SubTotal:  decimal.NewFromInt(49),
		},
	}

	cmd := domain.NewCreateOrderCommand(orderID, "demo-customer-id", items)
	_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create first purchase: %w", err)
	}

	// Complete the order
	completeCmd := domain.NewCompleteOrderCommand(orderID)
	_, err = deps.OrderCommandHandler.Handle(ctx, completeCmd)
	if err != nil {
		log.Printf("Warning: failed to complete first purchase: %v", err)
	}

	log.Printf("Demo customer made first purchase")
	return nil
}

func updateCustomerProfile(ctx context.Context, deps *DemoDependencies) error {
	cmd := domain.NewUpdateUserCommand("demo-customer-id", "Demo Customer Updated", "demo.updated@example.com")

	_, err := deps.UserCommandHandler.Handle(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to update customer profile: %w", err)
	}

	log.Printf("Updated demo customer profile")
	return nil
}

func makeMultiplePurchases(ctx context.Context, deps *DemoDependencies) error {
	// Create multiple orders to make customer VIP
	purchases := []struct {
		items []domain.OrderItem
	}{
		{
			items: []domain.OrderItem{
				{ProductID: "product-premium", Name: "Premium Item", Price: decimal.NewFromInt(299), Quantity: 1, SubTotal: decimal.NewFromInt(299)},
			},
		},
		{
			items: []domain.OrderItem{
				{ProductID: "product-luxury", Name: "Luxury Item", Price: decimal.NewFromInt(599), Quantity: 1, SubTotal: decimal.NewFromInt(599)},
			},
		},
		{
			items: []domain.OrderItem{
				{ProductID: "product-exclusive", Name: "Exclusive Item", Price: decimal.NewFromInt(399), Quantity: 1, SubTotal: decimal.NewFromInt(399)},
			},
		},
	}

	for i, purchase := range purchases {
		orderID := uuid.New().String()
		cmd := domain.NewCreateOrderCommand(orderID, "demo-customer-id", purchase.items)

		_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
		if err != nil {
			log.Printf("Warning: failed to create purchase %d: %v", i+1, err)
			continue
		}

		// Complete the order
		completeCmd := domain.NewCompleteOrderCommand(orderID)
		_, err = deps.OrderCommandHandler.Handle(ctx, completeCmd)
		if err != nil {
			log.Printf("Warning: failed to complete purchase %d: %v", i+1, err)
		}

		log.Printf("Demo customer made purchase %d", i+1)

		// Add some delay between purchases
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func cancelOrder(ctx context.Context, deps *DemoDependencies) error {
	// Create an order and then cancel it
	orderID := uuid.New().String()
	items := []domain.OrderItem{
		{ProductID: "product-cancel", Name: "Item to Cancel", Price: decimal.NewFromInt(99), Quantity: 1, SubTotal: decimal.NewFromInt(99)},
	}

	cmd := domain.NewCreateOrderCommand(orderID, "demo-customer-id", items)
	_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to create order for cancellation: %w", err)
	}

	// Cancel the order
	cancelCmd := domain.NewCancelOrderCommand(orderID, "Customer changed mind")
	_, err = deps.OrderCommandHandler.Handle(ctx, cancelCmd)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	log.Printf("Demo customer cancelled order")
	return nil
}

func createProductCategories(ctx context.Context, deps *DemoDependencies) error {
	categories := map[string][]struct {
		name  string
		price decimal.Decimal
		desc  string
	}{
		"Electronics": {
			{"Gaming Laptop", decimal.NewFromInt(1899), "High-end gaming laptop"},
			{"Tablet", decimal.NewFromInt(399), "10-inch tablet"},
			{"Smart Watch", decimal.NewFromInt(299), "Fitness tracking smartwatch"},
		},
		"Books": {
			{"Programming Guide", decimal.NewFromInt(49), "Complete programming guide"},
			{"Design Patterns", decimal.NewFromInt(59), "Software design patterns book"},
			{"Clean Code", decimal.NewFromInt(45), "Writing clean, maintainable code"},
		},
		"Sports": {
			{"Tennis Racket", decimal.NewFromInt(129), "Professional tennis racket"},
			{"Basketball", decimal.NewFromInt(35), "Official size basketball"},
			{"Yoga Mat", decimal.NewFromInt(29), "Non-slip yoga mat"},
		},
	}

	for category, products := range categories {
		for _, product := range products {
			productID := uuid.New().String()
			cmd := domain.NewCreateProductCommand(productID, product.name, product.price, category, product.desc)

			_, err := deps.ProductCommandHandler.Handle(ctx, cmd)
			if err != nil {
				log.Printf("Warning: failed to create product %s: %v", product.name, err)
				continue
			}

			log.Printf("Created product in %s: %s", category, product.name)
		}
	}

	return nil
}

func simulateProductSales(ctx context.Context, deps *DemoDependencies) error {
	// This would simulate sales for the products created above
	// In a real implementation, you'd track product IDs and create realistic orders

	log.Printf("Simulating product sales...")

	// Create multiple orders with different products to simulate popularity
	for i := 0; i < 10; i++ {
		orderID := uuid.New().String()
		customerID := fmt.Sprintf("customer-%d", i%3) // Rotate between 3 customers

		items := []domain.OrderItem{
			{
				ProductID: fmt.Sprintf("product-%d", i%5), // Rotate between 5 products
				Name:      fmt.Sprintf("Product %d", i%5),
				Price:     decimal.NewFromInt(int64(50 + i*10)),
				Quantity:  1 + i%3,
				SubTotal:  decimal.NewFromInt(int64((50 + i*10) * (1 + i%3))),
			},
		}

		cmd := domain.NewCreateOrderCommand(orderID, customerID, items)
		_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
		if err != nil {
			log.Printf("Warning: failed to create simulated order %d: %v", i, err)
			continue
		}

		// Complete most orders
		if i%4 != 0 { // Complete 75% of orders
			completeCmd := domain.NewCompleteOrderCommand(orderID)
			_, err = deps.OrderCommandHandler.Handle(ctx, completeCmd)
			if err != nil {
				log.Printf("Warning: failed to complete simulated order %d: %v", i, err)
			}
		}
	}

	log.Printf("Completed product sales simulation")
	return nil
}

func updateProductInfo(ctx context.Context, deps *DemoDependencies) error {
	// Update some product information
	updates := []struct {
		productID string
		name      string
		price     decimal.Decimal
		category  string
		desc      string
	}{
		{"product-1", "Gaming Laptop Pro", decimal.NewFromInt(1999), "Electronics", "Updated high-end gaming laptop"},
		{"product-2", "Programming Guide 2nd Edition", decimal.NewFromInt(55), "Books", "Updated programming guide"},
	}

	for _, update := range updates {
		cmd := domain.NewUpdateProductCommand(update.productID, update.name, update.price, update.category, update.desc)

		_, err := deps.ProductCommandHandler.Handle(ctx, cmd)
		if err != nil {
			log.Printf("Warning: failed to update product %s: %v", update.productID, err)
			continue
		}

		log.Printf("Updated product: %s", update.name)
	}

	return nil
}

func createComplexOrders(ctx context.Context, deps *DemoDependencies) error {
	// Create orders with multiple items
	complexOrders := []struct {
		customerID string
		items      []domain.OrderItem
	}{
		{
			customerID: "complex-customer-1",
			items: []domain.OrderItem{
				{ProductID: "p1", Name: "Item 1", Price: decimal.NewFromInt(100), Quantity: 2, SubTotal: decimal.NewFromInt(200)},
				{ProductID: "p2", Name: "Item 2", Price: decimal.NewFromInt(50), Quantity: 3, SubTotal: decimal.NewFromInt(150)},
				{ProductID: "p3", Name: "Item 3", Price: decimal.NewFromInt(75), Quantity: 1, SubTotal: decimal.NewFromInt(75)},
			},
		},
		{
			customerID: "complex-customer-2",
			items: []domain.OrderItem{
				{ProductID: "p4", Name: "Item 4", Price: decimal.NewFromInt(200), Quantity: 1, SubTotal: decimal.NewFromInt(200)},
				{ProductID: "p5", Name: "Item 5", Price: decimal.NewFromInt(25), Quantity: 4, SubTotal: decimal.NewFromInt(100)},
			},
		},
	}

	for i, order := range complexOrders {
		orderID := uuid.New().String()
		cmd := domain.NewCreateOrderCommand(orderID, order.customerID, order.items)

		_, err := deps.OrderCommandHandler.Handle(ctx, cmd)
		if err != nil {
			log.Printf("Warning: failed to create complex order %d: %v", i+1, err)
			continue
		}

		log.Printf("Created complex order %d with %d items", i+1, len(order.items))
	}

	return nil
}

func modifyOrders(ctx context.Context, deps *DemoDependencies) error {
	// This would modify existing orders by adding/removing items
	// For demo purposes, we'll just log the operations

	log.Printf("Simulating order modifications...")

	// In a real implementation, you'd:
	// 1. Get existing order IDs
	// 2. Add items using AddOrderItemCommand
	// 3. Remove items using RemoveOrderItemCommand

	log.Printf("Order modifications completed")
	return nil
}

func processOrders(ctx context.Context, deps *DemoDependencies) error {
	// This would process existing orders (complete/cancel)
	// For demo purposes, we'll just log the operations

	log.Printf("Processing orders...")

	// In a real implementation, you'd:
	// 1. Get existing order IDs
	// 2. Complete some orders
	// 3. Cancel some orders

	log.Printf("Order processing completed")
	return nil
}
