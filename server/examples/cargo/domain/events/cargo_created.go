package events

import (
	"fmt"
	"time"

	"cqrs"
)

// CargoCreatedEventData contains the data for cargo creation
type CargoCreatedEventData struct {
	CargoID     string    `json:"cargo_id"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	MaxWeight   float64   `json:"max_weight"`
	MaxVolume   float64   `json:"max_volume"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

// CargoCreatedEvent represents the event when a new cargo is created
type CargoCreatedEvent struct {
	*cqrs.BaseDomainEventMessage
	Data CargoCreatedEventData `json:"data"`
}

// NewCargoCreatedEvent creates a new cargo created event
func NewCargoCreatedEvent(cargoID, origin, destination string, maxWeight, maxVolume float64, createdBy string) *CargoCreatedEvent {
	eventData := CargoCreatedEventData{
		CargoID:     cargoID,
		Origin:      origin,
		Destination: destination,
		MaxWeight:   maxWeight,
		MaxVolume:   maxVolume,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
	}

	return &CargoCreatedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"CargoCreated",
			cargoID,
			"Cargo",
			1,
			eventData,
			createdBy,
			cqrs.UserIssuer,
		),
		Data: eventData,
	}
}

// GetCargoID returns the cargo ID
func (e *CargoCreatedEvent) GetCargoID() string {
	return e.Data.CargoID
}

// GetOrigin returns the origin location
func (e *CargoCreatedEvent) GetOrigin() string {
	return e.Data.Origin
}

// GetDestination returns the destination location
func (e *CargoCreatedEvent) GetDestination() string {
	return e.Data.Destination
}

// GetMaxWeight returns the maximum weight capacity
func (e *CargoCreatedEvent) GetMaxWeight() float64 {
	return e.Data.MaxWeight
}

// GetMaxVolume returns the maximum volume capacity
func (e *CargoCreatedEvent) GetMaxVolume() float64 {
	return e.Data.MaxVolume
}

// GetCreatedBy returns who created the cargo
func (e *CargoCreatedEvent) GetCreatedBy() string {
	return e.Data.CreatedBy
}

// GetCreatedAt returns when the cargo was created
func (e *CargoCreatedEvent) GetCreatedAt() time.Time {
	return e.Data.CreatedAt
}

// ValidateEvent validates the cargo created event
func (e *CargoCreatedEvent) ValidateEvent() error {
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
	if e.Data.MaxWeight <= 0 {
		return fmt.Errorf("max weight must be positive")
	}
	if e.Data.MaxVolume <= 0 {
		return fmt.Errorf("max volume must be positive")
	}
	if e.Data.CreatedBy == "" {
		return fmt.Errorf("created by cannot be empty")
	}

	return nil
}
