package handlers

import (
	"context"
	"fmt"

	"defense-allies-server/examples/cargo/application/commands"
	"defense-allies-server/examples/cargo/domain"
	"defense-allies-server/pkg/cqrs"
)

// CargoCommandHandler handles cargo-related commands
type CargoCommandHandler struct {
	*cqrs.BaseCommandHandler
	repository cqrs.EventSourcedRepository
}

// NewCargoCommandHandler creates a new cargo command handler
func NewCargoCommandHandler(repository cqrs.EventSourcedRepository) *CargoCommandHandler {
	supportedCommands := []string{
		"CreateCargo",
		"LoadShipment",
		"UnloadShipment",
		"StartTransport",
		"CompleteTransport",
	}

	handler := &CargoCommandHandler{
		BaseCommandHandler: cqrs.NewBaseCommandHandler("CargoCommandHandler", supportedCommands),
		repository:         repository,
	}

	return handler
}

// Handle handles the incoming command
func (h *CargoCommandHandler) Handle(ctx context.Context, command cqrs.Command) (*cqrs.CommandResult, error) {
	if command == nil {
		return nil, fmt.Errorf("command cannot be nil")
	}

	// Validate command
	if err := command.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	// Handle based on command type
	switch cmd := command.(type) {
	case *commands.CreateCargoCommand:
		return h.handleCreateCargo(ctx, cmd)
	case *commands.LoadShipmentCommand:
		return h.handleLoadShipment(ctx, cmd)
	default:
		return nil, fmt.Errorf("unsupported command type: %s", command.CommandType())
	}
}

// handleCreateCargo handles the create cargo command
func (h *CargoCommandHandler) handleCreateCargo(ctx context.Context, cmd *commands.CreateCargoCommand) (*cqrs.CommandResult, error) {
	// Check if cargo already exists
	existingCargo, err := h.repository.GetByID(ctx, cmd.GetCargoID())
	if err == nil && existingCargo != nil {
		return nil, fmt.Errorf("cargo %s already exists", cmd.GetCargoID())
	}

	// Create new cargo aggregate
	cargo := domain.NewCargoAggregate(
		cmd.GetCargoID(),
		cmd.GetOrigin(),
		cmd.GetDestination(),
		cmd.GetMaxWeight(),
		cmd.GetMaxVolume(),
	)

	// Execute the create cargo business logic
	if err := cargo.CreateCargo(cmd.UserID()); err != nil {
		return nil, fmt.Errorf("failed to create cargo: %w", err)
	}

	// Validate the aggregate state
	if err := cargo.Validate(); err != nil {
		return nil, fmt.Errorf("cargo validation failed: %w", err)
	}

	// Save the aggregate
	if err := h.repository.Save(ctx, cargo, cargo.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save cargo: %w", err)
	}

	// Return successful result
	return &cqrs.CommandResult{
		Success:     true,
		AggregateID: cargo.AggregateID(),
		Version:     cargo.CurrentVersion(),
		Events:      cargo.GetChanges(),
		Data: map[string]interface{}{
			"cargo_id":    cargo.AggregateID(),
			"origin":      cargo.GetOrigin(),
			"destination": cargo.GetDestination(),
			"max_weight":  cargo.GetMaxWeight(),
			"max_volume":  cargo.GetMaxVolume(),
			"status":      cargo.GetStatus().String(),
		},
	}, nil
}

// handleLoadShipment handles the load shipment command
func (h *CargoCommandHandler) handleLoadShipment(ctx context.Context, cmd *commands.LoadShipmentCommand) (*cqrs.CommandResult, error) {
	// Load the cargo aggregate
	aggregateRoot, err := h.repository.GetByID(ctx, cmd.GetCargoID())
	if err != nil {
		return nil, fmt.Errorf("cargo %s not found: %w", cmd.GetCargoID(), err)
	}

	cargo, ok := aggregateRoot.(*domain.CargoAggregate)
	if !ok {
		return nil, fmt.Errorf("invalid aggregate type")
	}

	// Create shipment from command data
	dimensions := domain.Dimensions{
		Length: cmd.Data.Length,
		Width:  cmd.Data.Width,
		Height: cmd.Data.Height,
	}

	shipment, err := domain.NewShipment(
		cmd.GetShipmentID(),
		cmd.GetDescription(),
		domain.ShipmentType(cmd.Data.ShipmentType),
		cmd.GetWeight(),
		dimensions,
		cmd.Data.Value,
		cmd.Data.Origin,
		cmd.Data.Destination,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shipment: %w", err)
	}

	// Set special handling requirements
	shipment.RequiresRefrigeration = cmd.Data.RequiresRefrigeration
	shipment.MaxTemperature = cmd.Data.MaxTemperature
	shipment.MinTemperature = cmd.Data.MinTemperature
	shipment.HandlingInstructions = cmd.Data.HandlingInstructions

	// Execute the load shipment business logic
	if err := cargo.LoadShipment(shipment, cmd.UserID(), cmd.GetLoadingTime()); err != nil {
		return nil, fmt.Errorf("failed to load shipment: %w", err)
	}

	// Validate the aggregate state
	if err := cargo.Validate(); err != nil {
		return nil, fmt.Errorf("cargo validation failed after loading shipment: %w", err)
	}

	// Save the aggregate
	if err := h.repository.Save(ctx, cargo, cargo.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save cargo after loading shipment: %w", err)
	}

	// Return successful result
	return &cqrs.CommandResult{
		Success:     true,
		AggregateID: cargo.AggregateID(),
		Version:     cargo.CurrentVersion(),
		Events:      cargo.GetChanges(),
		Data: map[string]interface{}{
			"cargo_id":         cargo.AggregateID(),
			"shipment_id":      shipment.ID,
			"shipment_count":   cargo.GetShipmentCount(),
			"current_weight":   cargo.GetCurrentWeight(),
			"current_volume":   cargo.GetCurrentVolume(),
			"available_weight": cargo.GetAvailableWeight(),
			"available_volume": cargo.GetAvailableVolume(),
			"status":           cargo.GetStatus().String(),
		},
	}, nil
}

// GetRepository returns the repository used by this handler
func (h *CargoCommandHandler) GetRepository() cqrs.EventSourcedRepository {
	return h.repository
}

// SetRepository sets the repository for this handler
func (h *CargoCommandHandler) SetRepository(repository cqrs.EventSourcedRepository) {
	h.repository = repository
}
