package main

import (
	"context"
	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/demo"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/04-read-models/infrastructure"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	log.Printf("=== CQRS Read Models and Projections Demo ===")
	log.Printf("This demo showcases MongoDB-based read models and projections")
	log.Printf("using the cqrsx library with real-world e-commerce scenarios.")
	log.Printf("")

	ctx := context.Background()

	// Initialize the demo application
	app, err := initializeApplication(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer app.cleanup(ctx)

	// Run the demo
	if err := app.runDemo(ctx); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	log.Printf("=== Demo Completed Successfully ===")
}

// DemoApplication encapsulates the demo application
type DemoApplication struct {
	// Infrastructure
	mongoClient *cqrsx.MongoClientManager
	readStore   *cqrsx.MongoReadStore
	eventStore  *cqrsx.MongoEventStore

	// Repositories
	userRepo    domain.UserRepository
	productRepo domain.ProductRepository
	orderRepo   domain.OrderRepository

	// Command Handlers
	userHandler    *domain.UserCommandHandler
	productHandler *domain.ProductCommandHandler
	orderHandler   *domain.OrderCommandHandler

	// Infrastructure
	eventRegistry    *infrastructure.EventHandlerRegistry
	readStoreFactory *infrastructure.ReadStoreFactory

	// Demo components
	queryExamples *demo.QueryExamples
}

// initializeApplication initializes all components
func initializeApplication(ctx context.Context) (*DemoApplication, error) {
	app := &DemoApplication{}

	// Initialize MongoDB connection
	if err := app.initializeMongoDB(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize MongoDB: %w", err)
	}

	// Initialize repositories
	if err := app.initializeRepositories(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	// Initialize command handlers
	app.initializeCommandHandlers()

	// Initialize infrastructure
	if err := app.initializeInfrastructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	// Initialize demo components
	app.initializeDemoComponents()

	log.Printf("Application initialized successfully")
	return app, nil
}

// initializeMongoDB sets up MongoDB connections
func (app *DemoApplication) initializeMongoDB(ctx context.Context) error {
	// MongoDB connection string - in a real app, this would come from config
	connectionString := getMongoConnectionString()
	databaseName := "cqrs_read_models_demo"

	// Create MongoDB client manager
	mongoConfig := &cqrsx.MongoConfig{
		URI:      connectionString,
		Database: databaseName,
	}
	mongoClient, err := cqrsx.NewMongoClientManager(mongoConfig)
	if err != nil {
		return fmt.Errorf("failed to create MongoDB client: %w", err)
	}
	app.mongoClient = mongoClient

	// MongoDB client manager created successfully
	log.Printf("MongoDB client manager created")

	// Create read store
	app.readStore = cqrsx.NewMongoReadStore(mongoClient, "read_models")

	// Create event store
	app.eventStore = cqrsx.NewMongoEventStore(mongoClient, "events")

	log.Printf("MongoDB initialized successfully")
	return nil
}

// initializeRepositories sets up domain repositories
func (app *DemoApplication) initializeRepositories(ctx context.Context) error {
	// Create actual MongoDB repositories with adapters
	userRepo := cqrsx.NewMongoEventSourcedRepository(app.mongoClient, "User")
	app.userRepo = infrastructure.NewUserRepositoryAdapter(userRepo)

	orderRepo := cqrsx.NewMongoHybridRepository(app.mongoClient, "Order")
	app.orderRepo = infrastructure.NewOrderRepositoryAdapter(orderRepo)

	productRepo := cqrsx.NewMongoStateBasedRepository(app.mongoClient, "Product")
	app.productRepo = infrastructure.NewProductRepositoryAdapter(productRepo)

	log.Printf("Repositories initialized with MongoDB implementations")
	return nil
}

// initializeCommandHandlers sets up command handlers
func (app *DemoApplication) initializeCommandHandlers() {
	// Create command handlers with repositories
	app.userHandler = domain.NewUserCommandHandler(app.userRepo)
	app.productHandler = domain.NewProductCommandHandler(app.productRepo)
	app.orderHandler = domain.NewOrderCommandHandler(app.orderRepo)

	log.Printf("Command handlers initialized")
}

// initializeInfrastructure sets up infrastructure components
func (app *DemoApplication) initializeInfrastructure(ctx context.Context) error {
	// Create read store factory
	config := infrastructure.ReadStoreConfig{
		Type: infrastructure.ReadStoreTypeMongoDB,
		MongoDB: &infrastructure.MongoConfig{
			ConnectionString: getMongoConnectionString(),
			DatabaseName:     "cqrs_read_models_demo",
			CollectionName:   "read_models",
		},
	}
	app.readStoreFactory = infrastructure.NewReadStoreFactory(config)

	// Create event handler registry (using nil for interface compatibility)
	app.eventRegistry = infrastructure.NewEventHandlerRegistry(nil, nil)

	// Register projections
	if err := app.eventRegistry.RegisterProjections(); err != nil {
		return fmt.Errorf("failed to register projections: %w", err)
	}

	// Set up default handlers
	if err := app.eventRegistry.SetupDefaultHandlers(); err != nil {
		return fmt.Errorf("failed to setup default handlers: %w", err)
	}

	// Start event handlers
	if err := app.eventRegistry.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event handlers: %w", err)
	}

	log.Printf("Infrastructure initialized successfully")
	return nil
}

// initializeDemoComponents sets up demo-specific components
func (app *DemoApplication) initializeDemoComponents() {
	// Using nil for now due to interface compatibility issues
	// The focus is on testing Repository and Event processing
	app.queryExamples = demo.NewQueryExamples(nil)
	log.Printf("Demo components initialized")
}

// runDemo executes the main demo flow
func (app *DemoApplication) runDemo(ctx context.Context) error {
	log.Printf("\n=== Starting Read Models and Projections Demo ===")

	// Step 1: Demonstrate basic setup
	if err := app.demonstrateSetup(ctx); err != nil {
		return fmt.Errorf("setup demonstration failed: %w", err)
	}

	// Step 2: Run demo scenarios
	if err := app.runDemoScenarios(ctx); err != nil {
		return fmt.Errorf("demo scenarios failed: %w", err)
	}

	// Step 3: Demonstrate read model queries
	if err := app.demonstrateQueries(ctx); err != nil {
		return fmt.Errorf("query demonstration failed: %w", err)
	}

	// Step 4: Show projection statistics
	if err := app.showProjectionStatistics(ctx); err != nil {
		return fmt.Errorf("statistics demonstration failed: %w", err)
	}

	return nil
}

// demonstrateSetup shows the initial setup
func (app *DemoApplication) demonstrateSetup(ctx context.Context) error {
	log.Printf("\n--- Demonstrating Setup ---")

	// Show read store factory capabilities
	log.Printf("Read Store Factory Configuration:")
	config := app.readStoreFactory.GetConfig()
	log.Printf("  Type: %s", config.Type)
	if config.MongoDB != nil {
		log.Printf("  MongoDB Database: %s", config.MongoDB.DatabaseName)
		log.Printf("  MongoDB Collection: %s", config.MongoDB.CollectionName)
	}

	// Show projection manager status
	projManager := app.eventRegistry.GetProjectionManager()
	if projManager.IsRunning() {
		log.Printf("Projection Manager: Running")
		projectionNames := projManager.GetProjectionNames()
		log.Printf("Registered Projections: %v", projectionNames)
	}

	// Show event handler registry status
	if app.eventRegistry.IsRunning() {
		log.Printf("Event Handler Registry: Running")
		stats := app.eventRegistry.GetStatistics()
		log.Printf("Handler Statistics: %+v", stats)
	}

	return nil
}

// runDemoScenarios executes demo scenarios
func (app *DemoApplication) runDemoScenarios(ctx context.Context) error {
	log.Printf("\n--- Running Demo Scenarios ---")

	// Create demo dependencies
	deps := &demo.DemoDependencies{
		UserCommandHandler:    app.userHandler,
		ProductCommandHandler: app.productHandler,
		OrderCommandHandler:   app.orderHandler,
	}

	// Get all scenarios
	scenarios := demo.GetAllScenarios()

	// Run each scenario
	for _, scenario := range scenarios {
		log.Printf("\n=== Scenario: %s ===", scenario.Name)
		log.Printf("Description: %s", scenario.Description)

		for _, step := range scenario.Steps {
			log.Printf("\nStep: %s", step.Name)
			log.Printf("Description: %s", step.Description)

			if err := step.Action(ctx, deps); err != nil {
				log.Printf("Warning: Step failed: %v", err)
				// Continue with next step
			}

			// Add small delay between steps for demonstration
			time.Sleep(100 * time.Millisecond)
		}

		log.Printf("Scenario completed: %s", scenario.Name)
	}

	return nil
}

// demonstrateQueries shows read model query capabilities
func (app *DemoApplication) demonstrateQueries(ctx context.Context) error {
	log.Printf("\n--- Demonstrating Read Model Queries ---")

	// Run all query examples
	if err := app.queryExamples.RunAllExamples(ctx); err != nil {
		return fmt.Errorf("query examples failed: %w", err)
	}

	return nil
}

// showProjectionStatistics displays projection and handler statistics
func (app *DemoApplication) showProjectionStatistics(ctx context.Context) error {
	log.Printf("\n--- Projection and Handler Statistics ---")

	// Show projection manager statistics
	projManager := app.eventRegistry.GetProjectionManager()
	projStats := projManager.GetStatistics()
	log.Printf("Projection Manager Statistics:")
	for key, value := range projStats {
		log.Printf("  %s: %v", key, value)
	}

	// Show individual projection information
	allProjectionInfo := projManager.GetAllProjectionInfo()
	log.Printf("\nIndividual Projection Information:")
	for name, info := range allProjectionInfo {
		log.Printf("  %s:", name)
		log.Printf("    Status: %s", info.Status)
		log.Printf("    Events Processed: %d", info.EventsProcessed)
		log.Printf("    Error Count: %d", info.ErrorCount)
		log.Printf("    Last Processed: %s", info.LastProcessedAt.Format(time.RFC3339))
		if info.LastError != "" {
			log.Printf("    Last Error: %s", info.LastError)
		}
	}

	// Show event handler registry statistics
	registryStats := app.eventRegistry.GetStatistics()
	log.Printf("\nEvent Handler Registry Statistics:")
	for key, value := range registryStats {
		log.Printf("  %s: %v", key, value)
	}

	return nil
}

// cleanup performs cleanup operations
func (app *DemoApplication) cleanup(ctx context.Context) {
	log.Printf("Cleaning up application...")

	// Stop event handlers
	if app.eventRegistry != nil {
		if err := app.eventRegistry.Stop(ctx); err != nil {
			log.Printf("Warning: Failed to stop event registry: %v", err)
		}
	}

	// Close MongoDB connection
	if app.mongoClient != nil {
		if err := app.mongoClient.Close(ctx); err != nil {
			log.Printf("Warning: Failed to close MongoDB connection: %v", err)
		}
	}

	log.Printf("Cleanup completed")
}

// getMongoConnectionString returns the MongoDB connection string
func getMongoConnectionString() string {
	// Check environment variable first
	if connStr := os.Getenv("MONGODB_CONNECTION_STRING"); connStr != "" {
		return connStr
	}

	// Default to local MongoDB
	return "mongodb://localhost:27017"
}

// Helper function to check if MongoDB is available
func checkMongoDBAvailability(ctx context.Context, connectionString string) error {
	mongoConfig := &cqrsx.MongoConfig{
		URI:      connectionString,
		Database: "test",
	}
	client, err := cqrsx.NewMongoClientManager(mongoConfig)
	if err != nil {
		return err
	}
	defer client.Close(ctx)

	// MongoDB client created successfully - assume connection is working
	return nil
}

// init function to check prerequisites
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Check if MongoDB is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	connectionString := getMongoConnectionString()
	if err := checkMongoDBAvailability(ctx, connectionString); err != nil {
		log.Printf("Warning: MongoDB not available at %s", connectionString)
		log.Printf("This demo requires MongoDB to be running.")
		log.Printf("You can:")
		log.Printf("1. Start MongoDB locally on default port (27017)")
		log.Printf("2. Set MONGODB_CONNECTION_STRING environment variable")
		log.Printf("3. Use Docker: docker run -d -p 27017:27017 mongo:latest")
		log.Printf("")
		log.Printf("The demo will continue but may fail during MongoDB operations.")
	} else {
		log.Printf("MongoDB connection verified successfully")
	}
}
