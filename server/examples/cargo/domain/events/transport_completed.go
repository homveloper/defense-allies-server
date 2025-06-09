package events

import (
	"fmt"
	"time"

	"cqrs"
)

// TransportCompletedEventData contains the data for transport completion
type TransportCompletedEventData struct {
	CargoID              string        `json:"cargo_id"`
	Origin               string        `json:"origin"`
	Destination          string        `json:"destination"`
	CompletedBy          string        `json:"completed_by"`
	CompletedAt          time.Time     `json:"completed_at"`
	StartedAt            time.Time     `json:"started_at"`
	ActualDuration       time.Duration `json:"actual_duration"`
	EstimatedDuration    time.Duration `json:"estimated_duration"`
	VehicleID            string        `json:"vehicle_id"`
	DriverID             string        `json:"driver_id,omitempty"`
	FinalLocation        string        `json:"final_location"`
	DeliveredShipments   int           `json:"delivered_shipments"`
	UndeliveredShipments int           `json:"undelivered_shipments"`
	TotalDistance        float64       `json:"total_distance"`    // in kilometers
	FuelConsumed         float64       `json:"fuel_consumed"`     // in liters
	CompletionStatus     string        `json:"completion_status"` // success, partial, failed
	Notes                string        `json:"notes,omitempty"`
}

// TransportCompletedEvent represents the event when cargo transport is completed
type TransportCompletedEvent struct {
	*cqrs.BaseDomainEventMessage
	Data TransportCompletedEventData `json:"data"`
}

// NewTransportCompletedEvent creates a new transport completed event
func NewTransportCompletedEvent(
	cargoID, origin, destination, completedBy string,
	startedAt time.Time,
	estimatedDuration time.Duration,
	vehicleID, driverID, finalLocation string,
	deliveredShipments, undeliveredShipments int,
	totalDistance, fuelConsumed float64,
	completionStatus, notes string,
) *TransportCompletedEvent {
	completedAt := time.Now()
	actualDuration := completedAt.Sub(startedAt)

	eventData := TransportCompletedEventData{
		CargoID:              cargoID,
		Origin:               origin,
		Destination:          destination,
		CompletedBy:          completedBy,
		CompletedAt:          completedAt,
		StartedAt:            startedAt,
		ActualDuration:       actualDuration,
		EstimatedDuration:    estimatedDuration,
		VehicleID:            vehicleID,
		DriverID:             driverID,
		FinalLocation:        finalLocation,
		DeliveredShipments:   deliveredShipments,
		UndeliveredShipments: undeliveredShipments,
		TotalDistance:        totalDistance,
		FuelConsumed:         fuelConsumed,
		CompletionStatus:     completionStatus,
		Notes:                notes,
	}

	return &TransportCompletedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"TransportCompleted",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			completedBy,
			cqrs.UserIssuer,
		),
		Data: eventData,
	}
}

// NewTransportCompletedEventBySystem creates a new transport completed event by system
func NewTransportCompletedEventBySystem(
	cargoID, origin, destination string,
	startedAt time.Time,
	estimatedDuration time.Duration,
	vehicleID, finalLocation string,
	deliveredShipments, undeliveredShipments int,
	totalDistance, fuelConsumed float64,
	completionStatus, notes string,
) *TransportCompletedEvent {
	completedAt := time.Now()
	actualDuration := completedAt.Sub(startedAt)

	eventData := TransportCompletedEventData{
		CargoID:              cargoID,
		Origin:               origin,
		Destination:          destination,
		CompletedBy:          "system",
		CompletedAt:          completedAt,
		StartedAt:            startedAt,
		ActualDuration:       actualDuration,
		EstimatedDuration:    estimatedDuration,
		VehicleID:            vehicleID,
		FinalLocation:        finalLocation,
		DeliveredShipments:   deliveredShipments,
		UndeliveredShipments: undeliveredShipments,
		TotalDistance:        totalDistance,
		FuelConsumed:         fuelConsumed,
		CompletionStatus:     completionStatus,
		Notes:                notes,
	}

	return &TransportCompletedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"TransportCompleted",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			"transport-tracker",
			cqrs.SystemIssuer,
		),
		Data: eventData,
	}
}

// GetCargoID returns the cargo ID
func (e *TransportCompletedEvent) GetCargoID() string {
	return e.Data.CargoID
}

// GetOrigin returns the origin location
func (e *TransportCompletedEvent) GetOrigin() string {
	return e.Data.Origin
}

// GetDestination returns the destination location
func (e *TransportCompletedEvent) GetDestination() string {
	return e.Data.Destination
}

// GetCompletedBy returns who completed the transport
func (e *TransportCompletedEvent) GetCompletedBy() string {
	return e.Data.CompletedBy
}

// GetCompletedAt returns when the transport was completed
func (e *TransportCompletedEvent) GetCompletedAt() time.Time {
	return e.Data.CompletedAt
}

// GetStartedAt returns when the transport started
func (e *TransportCompletedEvent) GetStartedAt() time.Time {
	return e.Data.StartedAt
}

// GetActualDuration returns the actual transport duration
func (e *TransportCompletedEvent) GetActualDuration() time.Duration {
	return e.Data.ActualDuration
}

// GetEstimatedDuration returns the estimated transport duration
func (e *TransportCompletedEvent) GetEstimatedDuration() time.Duration {
	return e.Data.EstimatedDuration
}

// GetVehicleID returns the vehicle ID
func (e *TransportCompletedEvent) GetVehicleID() string {
	return e.Data.VehicleID
}

// GetDriverID returns the driver ID
func (e *TransportCompletedEvent) GetDriverID() string {
	return e.Data.DriverID
}

// GetFinalLocation returns the final location
func (e *TransportCompletedEvent) GetFinalLocation() string {
	return e.Data.FinalLocation
}

// GetDeliveredShipments returns the number of delivered shipments
func (e *TransportCompletedEvent) GetDeliveredShipments() int {
	return e.Data.DeliveredShipments
}

// GetUndeliveredShipments returns the number of undelivered shipments
func (e *TransportCompletedEvent) GetUndeliveredShipments() int {
	return e.Data.UndeliveredShipments
}

// GetTotalDistance returns the total distance traveled
func (e *TransportCompletedEvent) GetTotalDistance() float64 {
	return e.Data.TotalDistance
}

// GetFuelConsumed returns the fuel consumed
func (e *TransportCompletedEvent) GetFuelConsumed() float64 {
	return e.Data.FuelConsumed
}

// GetCompletionStatus returns the completion status
func (e *TransportCompletedEvent) GetCompletionStatus() string {
	return e.Data.CompletionStatus
}

// GetNotes returns any additional notes
func (e *TransportCompletedEvent) GetNotes() string {
	return e.Data.Notes
}

// IsSuccessful checks if the transport was completed successfully
func (e *TransportCompletedEvent) IsSuccessful() bool {
	return e.Data.CompletionStatus == "success"
}

// IsPartialSuccess checks if the transport was partially successful
func (e *TransportCompletedEvent) IsPartialSuccess() bool {
	return e.Data.CompletionStatus == "partial"
}

// IsFailed checks if the transport failed
func (e *TransportCompletedEvent) IsFailed() bool {
	return e.Data.CompletionStatus == "failed"
}

// IsOnTime checks if the transport was completed on time
func (e *TransportCompletedEvent) IsOnTime() bool {
	return e.Data.ActualDuration <= e.Data.EstimatedDuration
}

// GetDelayDuration returns the delay duration (positive if late, negative if early)
func (e *TransportCompletedEvent) GetDelayDuration() time.Duration {
	return e.Data.ActualDuration - e.Data.EstimatedDuration
}

// IsSystemCompleted checks if the transport was completed by system
func (e *TransportCompletedEvent) IsSystemCompleted() bool {
	return e.IssuerType() == cqrs.SystemIssuer
}

// ValidateEvent validates the transport completed event
func (e *TransportCompletedEvent) ValidateEvent() error {
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
	if e.Data.CompletedBy == "" {
		return fmt.Errorf("completed by cannot be empty")
	}
	if e.Data.VehicleID == "" {
		return fmt.Errorf("vehicle ID cannot be empty")
	}
	if e.Data.FinalLocation == "" {
		return fmt.Errorf("final location cannot be empty")
	}
	if e.Data.CompletedAt.Before(e.Data.StartedAt) {
		return fmt.Errorf("completed time cannot be before start time")
	}
	if e.Data.DeliveredShipments < 0 {
		return fmt.Errorf("delivered shipments cannot be negative")
	}
	if e.Data.UndeliveredShipments < 0 {
		return fmt.Errorf("undelivered shipments cannot be negative")
	}
	if e.Data.TotalDistance < 0 {
		return fmt.Errorf("total distance cannot be negative")
	}
	if e.Data.FuelConsumed < 0 {
		return fmt.Errorf("fuel consumed cannot be negative")
	}
	if e.Data.CompletionStatus == "" {
		return fmt.Errorf("completion status cannot be empty")
	}

	return nil
}
