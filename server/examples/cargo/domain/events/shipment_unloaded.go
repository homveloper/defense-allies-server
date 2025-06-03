package events

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// ShipmentUnloadedEventData contains the data for shipment unloading
type ShipmentUnloadedEventData struct {
	CargoID       string        `json:"cargo_id"`
	ShipmentID    string        `json:"shipment_id"`
	UnloadedBy    string        `json:"unloaded_by"`
	UnloadedAt    time.Time     `json:"unloaded_at"`
	UnloadingTime time.Duration `json:"unloading_time"` // Time taken to unload
	Reason        string        `json:"reason"`         // Reason for unloading
	Location      string        `json:"location"`       // Where it was unloaded
}

// ShipmentUnloadedEvent represents the event when a shipment is unloaded from cargo
type ShipmentUnloadedEvent struct {
	*cqrs.BaseDomainEventMessage
	Data ShipmentUnloadedEventData `json:"data"`
}

// NewShipmentUnloadedEvent creates a new shipment unloaded event
func NewShipmentUnloadedEvent(cargoID, shipmentID, unloadedBy string, unloadingTime time.Duration, reason, location string) *ShipmentUnloadedEvent {
	eventData := ShipmentUnloadedEventData{
		CargoID:       cargoID,
		ShipmentID:    shipmentID,
		UnloadedBy:    unloadedBy,
		UnloadedAt:    time.Now(),
		UnloadingTime: unloadingTime,
		Reason:        reason,
		Location:      location,
	}

	return &ShipmentUnloadedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"ShipmentUnloaded",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			unloadedBy,
			cqrs.UserIssuer,
		),
		Data: eventData,
	}
}

// NewShipmentUnloadedEventBySystem creates a new shipment unloaded event by system
func NewShipmentUnloadedEventBySystem(cargoID, shipmentID string, unloadingTime time.Duration, reason, location string) *ShipmentUnloadedEvent {
	eventData := ShipmentUnloadedEventData{
		CargoID:       cargoID,
		ShipmentID:    shipmentID,
		UnloadedBy:    "system",
		UnloadedAt:    time.Now(),
		UnloadingTime: unloadingTime,
		Reason:        reason,
		Location:      location,
	}

	return &ShipmentUnloadedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"ShipmentUnloaded",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			"auto-unloader",
			cqrs.SystemIssuer,
		),
		Data: eventData,
	}
}

// GetCargoID returns the cargo ID
func (e *ShipmentUnloadedEvent) GetCargoID() string {
	return e.Data.CargoID
}

// GetShipmentID returns the shipment ID
func (e *ShipmentUnloadedEvent) GetShipmentID() string {
	return e.Data.ShipmentID
}

// GetUnloadedBy returns who unloaded the shipment
func (e *ShipmentUnloadedEvent) GetUnloadedBy() string {
	return e.Data.UnloadedBy
}

// GetUnloadedAt returns when the shipment was unloaded
func (e *ShipmentUnloadedEvent) GetUnloadedAt() time.Time {
	return e.Data.UnloadedAt
}

// GetUnloadingTime returns the time taken to unload the shipment
func (e *ShipmentUnloadedEvent) GetUnloadingTime() time.Duration {
	return e.Data.UnloadingTime
}

// GetReason returns the reason for unloading
func (e *ShipmentUnloadedEvent) GetReason() string {
	return e.Data.Reason
}

// GetLocation returns where the shipment was unloaded
func (e *ShipmentUnloadedEvent) GetLocation() string {
	return e.Data.Location
}

// IsSystemUnloaded checks if the shipment was unloaded by system
func (e *ShipmentUnloadedEvent) IsSystemUnloaded() bool {
	return e.IssuerType() == cqrs.SystemIssuer
}

// IsDestinationUnload checks if this is a destination unload
func (e *ShipmentUnloadedEvent) IsDestinationUnload() bool {
	return e.Data.Reason == "destination_reached" || e.Data.Reason == "delivery"
}

// IsEmergencyUnload checks if this is an emergency unload
func (e *ShipmentUnloadedEvent) IsEmergencyUnload() bool {
	return e.Data.Reason == "emergency" || e.Data.Reason == "damage" || e.Data.Reason == "hazard"
}

// ValidateEvent validates the shipment unloaded event
func (e *ShipmentUnloadedEvent) ValidateEvent() error {
	if err := e.BaseDomainEventMessage.ValidateEvent(); err != nil {
		return err
	}

	if e.Data.CargoID == "" {
		return fmt.Errorf("cargo ID cannot be empty")
	}
	if e.Data.ShipmentID == "" {
		return fmt.Errorf("shipment ID cannot be empty")
	}
	if e.Data.UnloadedBy == "" {
		return fmt.Errorf("unloaded by cannot be empty")
	}
	if e.Data.Location == "" {
		return fmt.Errorf("location cannot be empty")
	}

	return nil
}
