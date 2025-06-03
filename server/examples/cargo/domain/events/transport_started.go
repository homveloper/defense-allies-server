package events

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// TransportStartedEventData contains the data for transport start
type TransportStartedEventData struct {
	CargoID          string    `json:"cargo_id"`
	Origin           string    `json:"origin"`
	Destination      string    `json:"destination"`
	StartedBy        string    `json:"started_by"`
	StartedAt        time.Time `json:"started_at"`
	EstimatedArrival time.Time `json:"estimated_arrival"`
	TransportMode    string    `json:"transport_mode"` // truck, ship, plane, train
	VehicleID        string    `json:"vehicle_id"`
	DriverID         string    `json:"driver_id,omitempty"`
	Route            string    `json:"route,omitempty"`
	TotalShipments   int       `json:"total_shipments"`
	TotalWeight      float64   `json:"total_weight"`
	TotalVolume      float64   `json:"total_volume"`
}

// TransportStartedEvent represents the event when cargo transport begins
type TransportStartedEvent struct {
	*cqrs.BaseDomainEventMessage
	Data TransportStartedEventData `json:"data"`
}

// NewTransportStartedEvent creates a new transport started event
func NewTransportStartedEvent(
	cargoID, origin, destination, startedBy string,
	estimatedArrival time.Time,
	transportMode, vehicleID, driverID, route string,
	totalShipments int,
	totalWeight, totalVolume float64,
) *TransportStartedEvent {
	eventData := TransportStartedEventData{
		CargoID:          cargoID,
		Origin:           origin,
		Destination:      destination,
		StartedBy:        startedBy,
		StartedAt:        time.Now(),
		EstimatedArrival: estimatedArrival,
		TransportMode:    transportMode,
		VehicleID:        vehicleID,
		DriverID:         driverID,
		Route:            route,
		TotalShipments:   totalShipments,
		TotalWeight:      totalWeight,
		TotalVolume:      totalVolume,
	}

	return &TransportStartedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"TransportStarted",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			startedBy,
			cqrs.UserIssuer,
		),
		Data: eventData,
	}
}

// NewTransportStartedEventBySystem creates a new transport started event by system
func NewTransportStartedEventBySystem(
	cargoID, origin, destination string,
	estimatedArrival time.Time,
	transportMode, vehicleID, route string,
	totalShipments int,
	totalWeight, totalVolume float64,
) *TransportStartedEvent {
	eventData := TransportStartedEventData{
		CargoID:          cargoID,
		Origin:           origin,
		Destination:      destination,
		StartedBy:        "system",
		StartedAt:        time.Now(),
		EstimatedArrival: estimatedArrival,
		TransportMode:    transportMode,
		VehicleID:        vehicleID,
		Route:            route,
		TotalShipments:   totalShipments,
		TotalWeight:      totalWeight,
		TotalVolume:      totalVolume,
	}

	return &TransportStartedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"TransportStarted",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			"transport-scheduler",
			cqrs.SystemIssuer,
		),
		Data: eventData,
	}
}

// GetCargoID returns the cargo ID
func (e *TransportStartedEvent) GetCargoID() string {
	return e.Data.CargoID
}

// GetOrigin returns the origin location
func (e *TransportStartedEvent) GetOrigin() string {
	return e.Data.Origin
}

// GetDestination returns the destination location
func (e *TransportStartedEvent) GetDestination() string {
	return e.Data.Destination
}

// GetStartedBy returns who started the transport
func (e *TransportStartedEvent) GetStartedBy() string {
	return e.Data.StartedBy
}

// GetStartedAt returns when the transport started
func (e *TransportStartedEvent) GetStartedAt() time.Time {
	return e.Data.StartedAt
}

// GetEstimatedArrival returns the estimated arrival time
func (e *TransportStartedEvent) GetEstimatedArrival() time.Time {
	return e.Data.EstimatedArrival
}

// GetTransportMode returns the mode of transport
func (e *TransportStartedEvent) GetTransportMode() string {
	return e.Data.TransportMode
}

// GetVehicleID returns the vehicle ID
func (e *TransportStartedEvent) GetVehicleID() string {
	return e.Data.VehicleID
}

// GetDriverID returns the driver ID
func (e *TransportStartedEvent) GetDriverID() string {
	return e.Data.DriverID
}

// GetRoute returns the transport route
func (e *TransportStartedEvent) GetRoute() string {
	return e.Data.Route
}

// GetTotalShipments returns the total number of shipments
func (e *TransportStartedEvent) GetTotalShipments() int {
	return e.Data.TotalShipments
}

// GetTotalWeight returns the total weight
func (e *TransportStartedEvent) GetTotalWeight() float64 {
	return e.Data.TotalWeight
}

// GetTotalVolume returns the total volume
func (e *TransportStartedEvent) GetTotalVolume() float64 {
	return e.Data.TotalVolume
}

// GetEstimatedDuration returns the estimated transport duration
func (e *TransportStartedEvent) GetEstimatedDuration() time.Duration {
	return e.Data.EstimatedArrival.Sub(e.Data.StartedAt)
}

// IsSystemStarted checks if the transport was started by system
func (e *TransportStartedEvent) IsSystemStarted() bool {
	return e.IssuerType() == cqrs.SystemIssuer
}

// ValidateEvent validates the transport started event
func (e *TransportStartedEvent) ValidateEvent() error {
	if err := e.BaseDomainEventMessage.ValidateEvent(); err != nil {
		return err
	}

	if e.Data.CargoID == "" {
		return fmt.Errorf("cargo ID cannot be empty")
	}
	if e.Data.Origin == "" {
		return fmt.Errorf("origin cannot be empty")
	}
	if e.Data.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}
	if e.Data.StartedBy == "" {
		return fmt.Errorf("started by cannot be empty")
	}
	if e.Data.TransportMode == "" {
		return fmt.Errorf("transport mode cannot be empty")
	}
	if e.Data.VehicleID == "" {
		return fmt.Errorf("vehicle ID cannot be empty")
	}
	if e.Data.EstimatedArrival.Before(e.Data.StartedAt) {
		return fmt.Errorf("estimated arrival cannot be before start time")
	}
	if e.Data.TotalShipments < 0 {
		return fmt.Errorf("total shipments cannot be negative")
	}
	if e.Data.TotalWeight < 0 {
		return fmt.Errorf("total weight cannot be negative")
	}
	if e.Data.TotalVolume < 0 {
		return fmt.Errorf("total volume cannot be negative")
	}

	return nil
}
