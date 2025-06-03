package domain

import (
	"fmt"
	"time"

	"defense-allies-server/examples/cargo/domain/events"
	"defense-allies-server/pkg/cqrs"
)

// CargoStatus represents the current status of cargo
type CargoStatus int

const (
	CargoCreated CargoStatus = iota
	CargoLoading
	CargoInTransit
	CargoUnloading
	CargoCompleted
)

func (cs CargoStatus) String() string {
	switch cs {
	case CargoCreated:
		return "created"
	case CargoLoading:
		return "loading"
	case CargoInTransit:
		return "in_transit"
	case CargoUnloading:
		return "unloading"
	case CargoCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// CargoAggregate represents a cargo container that holds multiple shipments
type CargoAggregate struct {
	*cqrs.BaseAggregate

	// Cargo properties
	origin      string
	destination string
	maxWeight   float64
	maxVolume   float64
	status      CargoStatus

	// Current state
	shipments     map[string]*Shipment // shipmentID -> shipment
	currentWeight float64
	currentVolume float64

	// Transport information
	transportStartedAt *time.Time
	transportMode      string
	vehicleID          string
	driverID           string
	route              string
	estimatedArrival   *time.Time
}

// NewCargoAggregate creates a new cargo aggregate
func NewCargoAggregate(id, origin, destination string, maxWeight, maxVolume float64) *CargoAggregate {
	cargo := &CargoAggregate{
		BaseAggregate: cqrs.NewBaseAggregate(id, "Cargo"),
		origin:        origin,
		destination:   destination,
		maxWeight:     maxWeight,
		maxVolume:     maxVolume,
		status:        CargoCreated,
		shipments:     make(map[string]*Shipment),
		currentWeight: 0,
		currentVolume: 0,
	}

	return cargo
}

// CreateCargo creates a new cargo and applies the creation event
func (c *CargoAggregate) CreateCargo(createdBy string) error {
	if c.status != CargoCreated {
		return fmt.Errorf("cargo %s is already created", c.AggregateID())
	}

	event := events.NewCargoCreatedEvent(
		c.AggregateID(),
		c.origin,
		c.destination,
		c.maxWeight,
		c.maxVolume,
		createdBy,
	)

	c.Apply(event, true)
	return nil
}

// LoadShipment loads a shipment into the cargo
func (c *CargoAggregate) LoadShipment(shipment *Shipment, loadedBy string, loadingTime time.Duration) error {
	if c.status != CargoCreated && c.status != CargoLoading {
		return fmt.Errorf("cargo %s is not available for loading, current status: %s", c.AggregateID(), c.status.String())
	}

	// Check if shipment already exists
	if _, exists := c.shipments[shipment.ID]; exists {
		return fmt.Errorf("shipment %s is already loaded in cargo %s", shipment.ID, c.AggregateID())
	}

	// Check weight capacity
	if c.currentWeight+shipment.Weight > c.maxWeight {
		return fmt.Errorf("adding shipment %s would exceed weight capacity (current: %.2f, max: %.2f, adding: %.2f)",
			shipment.ID, c.currentWeight, c.maxWeight, shipment.Weight)
	}

	// Check volume capacity
	shipmentVolume := shipment.GetVolume()
	if c.currentVolume+shipmentVolume > c.maxVolume {
		return fmt.Errorf("adding shipment %s would exceed volume capacity (current: %.2f, max: %.2f, adding: %.2f)",
			shipment.ID, c.currentVolume, c.maxVolume, shipmentVolume)
	}

	// Load the shipment
	if err := shipment.Load(); err != nil {
		return fmt.Errorf("failed to load shipment %s: %w", shipment.ID, err)
	}

	position := len(c.shipments) + 1

	// Convert shipment to event data format
	shipmentData := events.ShipmentData{
		ID:                    shipment.ID,
		Description:           shipment.Description,
		Type:                  int(shipment.Type),
		Status:                int(shipment.Status),
		Weight:                shipment.Weight,
		Length:                shipment.Dimensions.Length,
		Width:                 shipment.Dimensions.Width,
		Height:                shipment.Dimensions.Height,
		Value:                 shipment.Value,
		Origin:                shipment.Origin,
		Destination:           shipment.Destination,
		CreatedAt:             shipment.CreatedAt,
		LoadedAt:              shipment.LoadedAt,
		RequiresRefrigeration: shipment.RequiresRefrigeration,
		MaxTemperature:        shipment.MaxTemperature,
		MinTemperature:        shipment.MinTemperature,
		HandlingInstructions:  shipment.HandlingInstructions,
	}

	event := events.NewShipmentLoadedEvent(
		c.AggregateID(),
		shipmentData,
		loadedBy,
		loadingTime,
		position,
	)

	c.Apply(event, true)
	return nil
}

// UnloadShipment unloads a shipment from the cargo
func (c *CargoAggregate) UnloadShipment(shipmentID, unloadedBy string, unloadingTime time.Duration, reason, location string) error {
	shipment, exists := c.shipments[shipmentID]
	if !exists {
		return fmt.Errorf("shipment %s not found in cargo %s", shipmentID, c.AggregateID())
	}

	if shipment.Status != ShipmentLoaded && shipment.Status != ShipmentInTransit {
		return fmt.Errorf("shipment %s cannot be unloaded, current status: %s", shipmentID, shipment.Status.String())
	}

	// Unload the shipment
	if err := shipment.Unload(); err != nil {
		return fmt.Errorf("failed to unload shipment %s: %w", shipmentID, err)
	}

	event := events.NewShipmentUnloadedEvent(
		c.AggregateID(),
		shipmentID,
		unloadedBy,
		unloadingTime,
		reason,
		location,
	)

	c.Apply(event, true)
	return nil
}

// StartTransport starts the transport of the cargo
func (c *CargoAggregate) StartTransport(startedBy string, estimatedArrival time.Time, transportMode, vehicleID, driverID, route string) error {
	if c.status != CargoLoading && c.status != CargoCreated {
		return fmt.Errorf("cargo %s cannot start transport, current status: %s", c.AggregateID(), c.status.String())
	}

	if len(c.shipments) == 0 {
		return fmt.Errorf("cargo %s has no shipments to transport", c.AggregateID())
	}

	event := events.NewTransportStartedEvent(
		c.AggregateID(),
		c.origin,
		c.destination,
		startedBy,
		estimatedArrival,
		transportMode,
		vehicleID,
		driverID,
		route,
		len(c.shipments),
		c.currentWeight,
		c.currentVolume,
	)

	c.Apply(event, true)
	return nil
}

// CompleteTransport completes the transport of the cargo
func (c *CargoAggregate) CompleteTransport(
	completedBy string,
	finalLocation string,
	deliveredShipments, undeliveredShipments int,
	totalDistance, fuelConsumed float64,
	completionStatus, notes string,
) error {
	if c.status != CargoInTransit {
		return fmt.Errorf("cargo %s is not in transit, current status: %s", c.AggregateID(), c.status.String())
	}

	if c.transportStartedAt == nil {
		return fmt.Errorf("cargo %s transport start time is not set", c.AggregateID())
	}

	estimatedDuration := time.Duration(0)
	if c.estimatedArrival != nil {
		estimatedDuration = c.estimatedArrival.Sub(*c.transportStartedAt)
	}

	event := events.NewTransportCompletedEvent(
		c.AggregateID(),
		c.origin,
		c.destination,
		completedBy,
		*c.transportStartedAt,
		estimatedDuration,
		c.vehicleID,
		c.driverID,
		finalLocation,
		deliveredShipments,
		undeliveredShipments,
		totalDistance,
		fuelConsumed,
		completionStatus,
		notes,
	)

	c.Apply(event, true)
	return nil
}

// Apply applies events to the aggregate state
func (c *CargoAggregate) Apply(event cqrs.EventMessage, isNew bool) {
	// Call base implementation for infrastructure concerns
	c.BaseAggregate.Apply(event, isNew)

	// Apply domain-specific logic based on event type
	switch e := event.(type) {
	case *events.CargoCreatedEvent:
		c.applyCargoCreated(e)
	case *events.ShipmentLoadedEvent:
		c.applyShipmentLoaded(e)
	case *events.ShipmentUnloadedEvent:
		c.applyShipmentUnloaded(e)
	case *events.TransportStartedEvent:
		c.applyTransportStarted(e)
	case *events.TransportCompletedEvent:
		c.applyTransportCompleted(e)
	}
}

// applyCargoCreated applies the cargo created event
func (c *CargoAggregate) applyCargoCreated(event *events.CargoCreatedEvent) {
	c.origin = event.GetOrigin()
	c.destination = event.GetDestination()
	c.maxWeight = event.GetMaxWeight()
	c.maxVolume = event.GetMaxVolume()
	c.status = CargoCreated
}

// applyShipmentLoaded applies the shipment loaded event
func (c *CargoAggregate) applyShipmentLoaded(event *events.ShipmentLoadedEvent) {
	shipmentData := event.GetShipment()

	// Convert event data back to domain shipment
	shipment := &Shipment{
		ID:          shipmentData.ID,
		Description: shipmentData.Description,
		Type:        ShipmentType(shipmentData.Type),
		Status:      ShipmentStatus(shipmentData.Status),
		Weight:      shipmentData.Weight,
		Dimensions: Dimensions{
			Length: shipmentData.Length,
			Width:  shipmentData.Width,
			Height: shipmentData.Height,
		},
		Value:                 shipmentData.Value,
		Origin:                shipmentData.Origin,
		Destination:           shipmentData.Destination,
		CreatedAt:             shipmentData.CreatedAt,
		LoadedAt:              shipmentData.LoadedAt,
		RequiresRefrigeration: shipmentData.RequiresRefrigeration,
		MaxTemperature:        shipmentData.MaxTemperature,
		MinTemperature:        shipmentData.MinTemperature,
		HandlingInstructions:  shipmentData.HandlingInstructions,
	}

	c.shipments[shipment.ID] = shipment
	c.currentWeight += shipment.Weight
	c.currentVolume += shipment.GetVolume()
	c.status = CargoLoading

	// Update shipment status in our copy
	if s, exists := c.shipments[shipment.ID]; exists {
		s.Status = ShipmentLoaded
		loadedAt := event.GetLoadedAt()
		s.LoadedAt = &loadedAt
	}
}

// applyShipmentUnloaded applies the shipment unloaded event
func (c *CargoAggregate) applyShipmentUnloaded(event *events.ShipmentUnloadedEvent) {
	shipmentID := event.GetShipmentID()
	if shipment, exists := c.shipments[shipmentID]; exists {
		c.currentWeight -= shipment.Weight
		c.currentVolume -= shipment.GetVolume()

		// Update shipment status
		shipment.Status = ShipmentUnloaded
		unloadedAt := event.GetUnloadedAt()
		shipment.UnloadedAt = &unloadedAt

		// Remove from cargo if delivered
		if event.IsDestinationUnload() {
			delete(c.shipments, shipmentID)
		}
	}

	// Update cargo status
	if len(c.shipments) == 0 {
		c.status = CargoCompleted
	} else {
		c.status = CargoUnloading
	}
}

// applyTransportStarted applies the transport started event
func (c *CargoAggregate) applyTransportStarted(event *events.TransportStartedEvent) {
	c.status = CargoInTransit
	startedAt := event.GetStartedAt()
	c.transportStartedAt = &startedAt
	c.transportMode = event.GetTransportMode()
	c.vehicleID = event.GetVehicleID()
	c.driverID = event.GetDriverID()
	c.route = event.GetRoute()
	estimatedArrival := event.GetEstimatedArrival()
	c.estimatedArrival = &estimatedArrival

	// Update all shipments to in-transit status
	for _, shipment := range c.shipments {
		shipment.Status = ShipmentInTransit
	}
}

// applyTransportCompleted applies the transport completed event
func (c *CargoAggregate) applyTransportCompleted(event *events.TransportCompletedEvent) {
	c.status = CargoCompleted
}

// Getters for aggregate state

func (c *CargoAggregate) GetOrigin() string {
	return c.origin
}

func (c *CargoAggregate) GetDestination() string {
	return c.destination
}

func (c *CargoAggregate) GetMaxWeight() float64 {
	return c.maxWeight
}

func (c *CargoAggregate) GetMaxVolume() float64 {
	return c.maxVolume
}

func (c *CargoAggregate) GetStatus() CargoStatus {
	return c.status
}

func (c *CargoAggregate) GetShipments() map[string]*Shipment {
	// Return a copy to prevent external modification
	result := make(map[string]*Shipment)
	for k, v := range c.shipments {
		result[k] = v.Clone()
	}
	return result
}

func (c *CargoAggregate) GetCurrentWeight() float64 {
	return c.currentWeight
}

func (c *CargoAggregate) GetCurrentVolume() float64 {
	return c.currentVolume
}

func (c *CargoAggregate) GetShipmentCount() int {
	return len(c.shipments)
}

func (c *CargoAggregate) GetAvailableWeight() float64 {
	return c.maxWeight - c.currentWeight
}

func (c *CargoAggregate) GetAvailableVolume() float64 {
	return c.maxVolume - c.currentVolume
}

func (c *CargoAggregate) IsInTransit() bool {
	return c.status == CargoInTransit
}

func (c *CargoAggregate) IsCompleted() bool {
	return c.status == CargoCompleted
}

func (c *CargoAggregate) GetTransportStartedAt() *time.Time {
	return c.transportStartedAt
}

func (c *CargoAggregate) GetEstimatedArrival() *time.Time {
	return c.estimatedArrival
}

func (c *CargoAggregate) GetVehicleID() string {
	return c.vehicleID
}

func (c *CargoAggregate) GetDriverID() string {
	return c.driverID
}

func (c *CargoAggregate) GetRoute() string {
	return c.route
}

// Validate validates the aggregate state
func (c *CargoAggregate) Validate() error {
	if err := c.BaseAggregate.Validate(); err != nil {
		return err
	}

	if c.origin == "" {
		return fmt.Errorf("cargo origin cannot be empty")
	}
	if c.destination == "" {
		return fmt.Errorf("cargo destination cannot be empty")
	}
	if c.maxWeight <= 0 {
		return fmt.Errorf("cargo max weight must be positive")
	}
	if c.maxVolume <= 0 {
		return fmt.Errorf("cargo max volume must be positive")
	}
	if c.currentWeight > c.maxWeight {
		return fmt.Errorf("cargo current weight exceeds maximum")
	}
	if c.currentVolume > c.maxVolume {
		return fmt.Errorf("cargo current volume exceeds maximum")
	}

	return nil
}
