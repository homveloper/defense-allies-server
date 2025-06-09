package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cqrs"
	"defense-allies-server/examples/cargo/application/commands"
	"defense-allies-server/examples/cargo/application/handlers"
	"defense-allies-server/examples/cargo/domain"

	"github.com/google/uuid"
)

func main() {
	fmt.Println("🚛 Cargo Transport System - Event Sourcing CQRS Example")
	fmt.Println("========================================================")

	// Initialize CQRS infrastructure
	ctx := context.Background()

	// Create in-memory repository for this example
	repository := NewInMemoryCargoRepository()

	// Create command dispatcher
	commandDispatcher := cqrs.NewInMemoryCommandDispatcher()

	// Create and register command handler
	cargoHandler := handlers.NewCargoCommandHandler(repository)
	if err := commandDispatcher.RegisterHandler("CreateCargo", cargoHandler); err != nil {
		log.Fatalf("Failed to register CreateCargo handler: %v", err)
	}
	if err := commandDispatcher.RegisterHandler("LoadShipment", cargoHandler); err != nil {
		log.Fatalf("Failed to register LoadShipment handler: %v", err)
	}

	// Create event bus for projections
	eventBus := cqrs.NewInMemoryEventBus()
	if err := eventBus.Start(ctx); err != nil {
		log.Fatalf("Failed to start event bus: %v", err)
	}
	defer eventBus.Stop(ctx)

	fmt.Println("\n✅ CQRS Infrastructure initialized successfully")

	// Run the cargo transport example
	if err := runCargoExample(ctx, commandDispatcher, repository); err != nil {
		log.Fatalf("Example failed: %v", err)
	}

	fmt.Println("\n🎉 Cargo Transport Example completed successfully!")
}

func runCargoExample(ctx context.Context, dispatcher cqrs.CommandDispatcher, repository cqrs.EventSourcedRepository) error {
	fmt.Println("\n📦 Starting Cargo Transport Example...")

	// Step 1: Create a new cargo
	fmt.Println("\n1️⃣ Creating new cargo...")
	cargoID := uuid.New().String()
	userID := "user123"

	createCargoCmd := commands.NewCreateCargoCommandWithID(
		cargoID,
		"Seoul, South Korea",
		"Busan, South Korea",
		10000.0, // 10 tons max weight
		50.0,    // 50 cubic meters max volume
		userID,
	)

	result, err := dispatcher.Dispatch(ctx, createCargoCmd)
	if err != nil {
		return fmt.Errorf("failed to create cargo: %w", err)
	}

	fmt.Printf("   ✅ Cargo created successfully!")
	fmt.Printf("   📋 Cargo ID: %s\n", cargoID)
	fmt.Printf("   📍 Route: %s → %s\n", createCargoCmd.GetOrigin(), createCargoCmd.GetDestination())
	fmt.Printf("   ⚖️  Max Weight: %.1f kg\n", createCargoCmd.GetMaxWeight())
	fmt.Printf("   📦 Max Volume: %.1f m³\n", createCargoCmd.GetMaxVolume())
	fmt.Printf("   🔢 Version: %d\n", result.Version)

	// Step 2: Load first shipment
	fmt.Println("\n2️⃣ Loading first shipment...")
	shipment1ID := uuid.New().String()

	loadShipment1Cmd := commands.NewLoadShipmentCommand(
		cargoID,
		shipment1ID,
		"Electronics - Smartphones",
		int(domain.ValuableCargo),
		500.0,   // 500 kg
		1.2,     // 1.2m length
		0.8,     // 0.8m width
		0.6,     // 0.6m height
		50000.0, // $50,000 value
		"Seoul, South Korea",
		"Busan, South Korea",
		15*time.Minute, // 15 minutes loading time
		userID,
	)

	result, err = dispatcher.Dispatch(ctx, loadShipment1Cmd)
	if err != nil {
		return fmt.Errorf("failed to load first shipment: %w", err)
	}

	fmt.Printf("   ✅ First shipment loaded successfully!")
	fmt.Printf("   📦 Shipment ID: %s\n", shipment1ID)
	fmt.Printf("   📝 Description: %s\n", loadShipment1Cmd.GetDescription())
	fmt.Printf("   ⚖️  Weight: %.1f kg\n", loadShipment1Cmd.GetWeight())
	fmt.Printf("   📐 Volume: %.2f m³\n", loadShipment1Cmd.GetVolume())
	fmt.Printf("   ⏱️  Loading Time: %v\n", loadShipment1Cmd.GetLoadingTime())
	fmt.Printf("   🔢 Version: %d\n", result.Version)

	// Step 3: Load second shipment
	fmt.Println("\n3️⃣ Loading second shipment...")
	shipment2ID := uuid.New().String()

	loadShipment2Cmd := commands.NewLoadShipmentCommand(
		cargoID,
		shipment2ID,
		"Automotive Parts - Engine Components",
		int(domain.GeneralCargo),
		1500.0,  // 1500 kg
		2.0,     // 2.0m length
		1.5,     // 1.5m width
		1.0,     // 1.0m height
		25000.0, // $25,000 value
		"Seoul, South Korea",
		"Busan, South Korea",
		30*time.Minute, // 30 minutes loading time
		userID,
	)

	result, err = dispatcher.Dispatch(ctx, loadShipment2Cmd)
	if err != nil {
		return fmt.Errorf("failed to load second shipment: %w", err)
	}

	fmt.Printf("   ✅ Second shipment loaded successfully!")
	fmt.Printf("   📦 Shipment ID: %s\n", shipment2ID)
	fmt.Printf("   📝 Description: %s\n", loadShipment2Cmd.GetDescription())
	fmt.Printf("   ⚖️  Weight: %.1f kg\n", loadShipment2Cmd.GetWeight())
	fmt.Printf("   📐 Volume: %.2f m³\n", loadShipment2Cmd.GetVolume())
	fmt.Printf("   ⏱️  Loading Time: %v\n", loadShipment2Cmd.GetLoadingTime())
	fmt.Printf("   🔢 Version: %d\n", result.Version)

	// Step 4: Check current cargo state
	fmt.Println("\n4️⃣ Checking current cargo state...")

	aggregateRoot, err := repository.GetByID(ctx, cargoID)
	if err != nil {
		return fmt.Errorf("failed to get cargo: %w", err)
	}

	cargo, ok := aggregateRoot.(*domain.CargoAggregate)
	if !ok {
		return fmt.Errorf("invalid aggregate type")
	}

	fmt.Printf("   📊 Cargo Status: %s\n", cargo.GetStatus().String())
	fmt.Printf("   📦 Shipments Count: %d\n", cargo.GetShipmentCount())
	fmt.Printf("   ⚖️  Current Weight: %.1f kg (%.1f%% of capacity)\n",
		cargo.GetCurrentWeight(),
		(cargo.GetCurrentWeight()/cargo.GetMaxWeight())*100)
	fmt.Printf("   📐 Current Volume: %.2f m³ (%.1f%% of capacity)\n",
		cargo.GetCurrentVolume(),
		(cargo.GetCurrentVolume()/cargo.GetMaxVolume())*100)
	fmt.Printf("   🆓 Available Weight: %.1f kg\n", cargo.GetAvailableWeight())
	fmt.Printf("   🆓 Available Volume: %.2f m³\n", cargo.GetAvailableVolume())

	// Step 5: Display event history
	fmt.Println("\n5️⃣ Event History:")
	events, err := repository.GetEventHistory(ctx, cargoID, 0) // Get all events from version 0
	if err != nil {
		return fmt.Errorf("failed to get event history: %w", err)
	}

	for i, event := range events {
		if domainEvent, ok := event.(cqrs.DomainEventMessage); ok {
			fmt.Printf("   %d. %s (v%d) - Issued by: %s (%s) at %s\n",
				i+1,
				event.EventType(),
				event.Version(),
				domainEvent.IssuerID(),
				domainEvent.IssuerType().String(),
				event.Timestamp().Format("15:04:05"),
			)
		} else {
			fmt.Printf("   %d. %s (v%d) at %s\n",
				i+1,
				event.EventType(),
				event.Version(),
				event.Timestamp().Format("15:04:05"),
			)
		}
	}

	// Step 6: Display shipment details
	fmt.Println("\n6️⃣ Loaded Shipments:")
	shipments := cargo.GetShipments()
	for i, shipment := range shipments {
		fmt.Printf("   📦 Shipment %s:\n", i)
		fmt.Printf("      📝 Description: %s\n", shipment.Description)
		fmt.Printf("      🏷️  Type: %s\n", shipment.Type.String())
		fmt.Printf("      📊 Status: %s\n", shipment.Status.String())
		fmt.Printf("      ⚖️  Weight: %.1f kg\n", shipment.Weight)
		fmt.Printf("      📐 Volume: %.2f m³\n", shipment.GetVolume())
		fmt.Printf("      💰 Value: $%.2f\n", shipment.Value)
		if shipment.LoadedAt != nil {
			fmt.Printf("      ⏰ Loaded At: %s\n", shipment.LoadedAt.Format("15:04:05"))
		}
		fmt.Println()
	}

	return nil
}

var _ cqrs.EventSourcedRepository = (*InMemoryCargoRepository)(nil)

// InMemoryCargoRepository is a simple in-memory repository for the cargo example
type InMemoryCargoRepository struct {
	cargos map[string]*domain.CargoAggregate
	events map[string][]cqrs.EventMessage // aggregateID -> events
}

// NewInMemoryCargoRepository creates a new InMemoryCargoRepository
func NewInMemoryCargoRepository() *InMemoryCargoRepository {
	return &InMemoryCargoRepository{
		cargos: make(map[string]*domain.CargoAggregate),
		events: make(map[string][]cqrs.EventMessage),
	}
}

// Save saves an aggregate
func (r *InMemoryCargoRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	cargo, ok := aggregate.(*domain.CargoAggregate)
	if !ok {
		return fmt.Errorf("invalid aggregate type: expected *domain.CargoAggregate, got %T", aggregate)
	}

	// Check version for optimistic concurrency control
	if existing, exists := r.cargos[cargo.ID()]; exists {
		// For existing aggregates, check if the expected version matches the stored version
		if existing.OriginalVersion() != expectedVersion {
			return fmt.Errorf("version conflict: expected %d, got %d", expectedVersion, existing.OriginalVersion())
		}
	} else {
		// For new aggregates, expected version should be 0
		if expectedVersion != 0 {
			return fmt.Errorf("new aggregate version conflict: expected 0, got %d", expectedVersion)
		}
	}

	// Store events for history
	changes := cargo.GetChanges()
	if len(changes) > 0 {
		if r.events[cargo.ID()] == nil {
			r.events[cargo.ID()] = make([]cqrs.EventMessage, 0)
		}
		r.events[cargo.ID()] = append(r.events[cargo.ID()], changes...)
	}

	// Clone the cargo to avoid external modifications
	clonedCargo := *cargo
	r.cargos[cargo.ID()] = &clonedCargo

	// Clear changes after saving
	cargo.ClearChanges()

	return nil
}

// GetByID gets an aggregate by ID
func (r *InMemoryCargoRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	cargo, exists := r.cargos[id]
	if !exists {
		return nil, fmt.Errorf("cargo with ID %s not found", id)
	}

	// Clone the cargo to avoid external modifications
	clonedCargo := *cargo
	return &clonedCargo, nil
}

// GetVersion gets the version of an aggregate
func (r *InMemoryCargoRepository) GetVersion(ctx context.Context, id string) (int, error) {
	cargo, exists := r.cargos[id]
	if !exists {
		return 0, fmt.Errorf("cargo with ID %s not found", id)
	}
	return cargo.Version(), nil
}

// Exists checks if an aggregate exists
func (r *InMemoryCargoRepository) Exists(ctx context.Context, id string) bool {
	_, exists := r.cargos[id]
	return exists
}

// EventSourcedRepository interface implementation

// SaveEvents saves events for an aggregate
func (r *InMemoryCargoRepository) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	// Check version for optimistic concurrency control
	if existing, exists := r.cargos[aggregateID]; exists {
		if existing.Version() != expectedVersion {
			return fmt.Errorf("version conflict: expected %d, got %d", expectedVersion, existing.Version())
		}
	}

	// Store events for history
	if len(events) > 0 {
		if r.events[aggregateID] == nil {
			r.events[aggregateID] = make([]cqrs.EventMessage, 0)
		}
		r.events[aggregateID] = append(r.events[aggregateID], events...)
	}

	return nil
}

// GetEventHistory gets the event history for an aggregate starting from a specific version
func (r *InMemoryCargoRepository) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]cqrs.EventMessage, error) {
	events, exists := r.events[aggregateID]
	if !exists {
		return []cqrs.EventMessage{}, nil
	}

	// Filter events from the specified version
	var result []cqrs.EventMessage
	for _, event := range events {
		if event.Version() >= fromVersion {
			result = append(result, event)
		}
	}

	return result, nil
}

// GetEventStream returns a channel for streaming events (simplified implementation)
func (r *InMemoryCargoRepository) GetEventStream(ctx context.Context, aggregateID string) (<-chan cqrs.EventMessage, error) {
	events, exists := r.events[aggregateID]
	if !exists {
		events = []cqrs.EventMessage{}
	}

	ch := make(chan cqrs.EventMessage, len(events))
	go func() {
		defer close(ch)
		for _, event := range events {
			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// SaveSnapshot saves a snapshot (simplified implementation)
func (r *InMemoryCargoRepository) SaveSnapshot(ctx context.Context, snapshot cqrs.SnapshotData) error {
	// For this example, we'll just ignore snapshots
	return nil
}

// GetSnapshot gets the latest snapshot (simplified implementation)
func (r *InMemoryCargoRepository) GetSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	// For this example, we'll return nil (no snapshot)
	return nil, fmt.Errorf("no snapshot found for aggregate %s", aggregateID)
}

// DeleteSnapshot removes the snapshot for an aggregate (simplified implementation)
func (r *InMemoryCargoRepository) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	// For this example, we'll just ignore snapshot deletion
	return nil
}

// LoadFromSnapshot loads an aggregate from snapshot (simplified implementation)
func (r *InMemoryCargoRepository) LoadFromSnapshot(ctx context.Context, aggregateID string) (cqrs.AggregateRoot, error) {
	// For this example, we'll just use regular GetByID
	return r.GetByID(ctx, aggregateID)
}

// GetLastEventVersion gets the last event version for an aggregate
func (r *InMemoryCargoRepository) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	events, exists := r.events[aggregateID]
	if !exists || len(events) == 0 {
		return 0, nil
	}

	// Return the version of the last event
	lastEvent := events[len(events)-1]
	return lastEvent.Version(), nil
}

// CompactEvents removes old events before a specific version (simplified implementation)
func (r *InMemoryCargoRepository) CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error {
	events, exists := r.events[aggregateID]
	if !exists {
		return nil
	}

	// Keep only events from the specified version onwards
	var compactedEvents []cqrs.EventMessage
	for _, event := range events {
		if event.Version() >= beforeVersion {
			compactedEvents = append(compactedEvents, event)
		}
	}

	r.events[aggregateID] = compactedEvents
	return nil
}
