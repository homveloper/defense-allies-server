package events

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// ShipmentData represents shipment data in events (to avoid circular imports)
type ShipmentData struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Type        int        `json:"type"`
	Status      int        `json:"status"`
	Weight      float64    `json:"weight"`
	Length      float64    `json:"length"`
	Width       float64    `json:"width"`
	Height      float64    `json:"height"`
	Value       float64    `json:"value"`
	Origin      string     `json:"origin"`
	Destination string     `json:"destination"`
	CreatedAt   time.Time  `json:"created_at"`
	LoadedAt    *time.Time `json:"loaded_at,omitempty"`

	// Special handling requirements
	RequiresRefrigeration bool    `json:"requires_refrigeration"`
	MaxTemperature        float64 `json:"max_temperature,omitempty"`
	MinTemperature        float64 `json:"min_temperature,omitempty"`
	HandlingInstructions  string  `json:"handling_instructions,omitempty"`
}

// Validate validates the shipment data
func (s *ShipmentData) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("shipment ID cannot be empty")
	}
	if s.Description == "" {
		return fmt.Errorf("shipment description cannot be empty")
	}
	if s.Weight <= 0 {
		return fmt.Errorf("shipment weight must be positive")
	}
	if s.Length <= 0 || s.Width <= 0 || s.Height <= 0 {
		return fmt.Errorf("shipment dimensions must be positive")
	}
	if s.Origin == "" {
		return fmt.Errorf("origin cannot be empty")
	}
	if s.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}
	return nil
}

// ShipmentLoadedEventData contains the data for shipment loading
type ShipmentLoadedEventData struct {
	CargoID     string        `json:"cargo_id"`
	Shipment    ShipmentData  `json:"shipment"`
	LoadedBy    string        `json:"loaded_by"`
	LoadedAt    time.Time     `json:"loaded_at"`
	LoadingTime time.Duration `json:"loading_time"` // Time taken to load
	Position    int           `json:"position"`     // Position in cargo
}

// ShipmentLoadedEvent represents the event when a shipment is loaded into cargo
type ShipmentLoadedEvent struct {
	*cqrs.BaseDomainEventMessage
	Data ShipmentLoadedEventData `json:"data"`
}

// NewShipmentLoadedEvent creates a new shipment loaded event
func NewShipmentLoadedEvent(cargoID string, shipment ShipmentData, loadedBy string, loadingTime time.Duration, position int) *ShipmentLoadedEvent {
	eventData := ShipmentLoadedEventData{
		CargoID:     cargoID,
		Shipment:    shipment,
		LoadedBy:    loadedBy,
		LoadedAt:    time.Now(),
		LoadingTime: loadingTime,
		Position:    position,
	}

	return &ShipmentLoadedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"ShipmentLoaded",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			loadedBy,
			cqrs.UserIssuer,
		),
		Data: eventData,
	}
}

// NewShipmentLoadedEventBySystem creates a new shipment loaded event by system
func NewShipmentLoadedEventBySystem(cargoID string, shipment ShipmentData, loadingTime time.Duration, position int) *ShipmentLoadedEvent {
	eventData := ShipmentLoadedEventData{
		CargoID:     cargoID,
		Shipment:    shipment,
		LoadedBy:    "system",
		LoadedAt:    time.Now(),
		LoadingTime: loadingTime,
		Position:    position,
	}

	return &ShipmentLoadedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessageWithIssuer(
			"ShipmentLoaded",
			cargoID,
			"Cargo",
			0, // Version will be set by aggregate
			eventData,
			"auto-loader",
			cqrs.SystemIssuer,
		),
		Data: eventData,
	}
}

// GetCargoID returns the cargo ID
func (e *ShipmentLoadedEvent) GetCargoID() string {
	return e.Data.CargoID
}

// GetShipment returns the loaded shipment
func (e *ShipmentLoadedEvent) GetShipment() ShipmentData {
	return e.Data.Shipment
}

// GetShipmentID returns the shipment ID
func (e *ShipmentLoadedEvent) GetShipmentID() string {
	return e.Data.Shipment.ID
}

// GetLoadedBy returns who loaded the shipment
func (e *ShipmentLoadedEvent) GetLoadedBy() string {
	return e.Data.LoadedBy
}

// GetLoadedAt returns when the shipment was loaded
func (e *ShipmentLoadedEvent) GetLoadedAt() time.Time {
	return e.Data.LoadedAt
}

// GetLoadingTime returns the time taken to load the shipment
func (e *ShipmentLoadedEvent) GetLoadingTime() time.Duration {
	return e.Data.LoadingTime
}

// GetPosition returns the position of the shipment in cargo
func (e *ShipmentLoadedEvent) GetPosition() int {
	return e.Data.Position
}

// IsSystemLoaded checks if the shipment was loaded by system
func (e *ShipmentLoadedEvent) IsSystemLoaded() bool {
	return e.IssuerType() == cqrs.SystemIssuer
}

// ValidateEvent validates the shipment loaded event
func (e *ShipmentLoadedEvent) ValidateEvent() error {
	if err := e.BaseDomainEventMessage.ValidateEvent(); err != nil {
		return err
	}

	if e.Data.CargoID == "" {
		return fmt.Errorf("cargo ID cannot be empty")
	}
	if e.Data.Shipment.ID == "" {
		return fmt.Errorf("shipment ID cannot be empty")
	}
	if err := e.Data.Shipment.Validate(); err != nil {
		return fmt.Errorf("invalid shipment data: %w", err)
	}
	if e.Data.LoadedBy == "" {
		return fmt.Errorf("loaded by cannot be empty")
	}
	if e.Data.Position < 0 {
		return fmt.Errorf("position cannot be negative")
	}

	return nil
}
