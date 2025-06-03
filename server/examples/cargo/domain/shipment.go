package domain

import (
	"fmt"
	"time"
)

// ShipmentType represents the type of shipment
type ShipmentType int

const (
	GeneralCargo ShipmentType = iota
	FragileCargo
	HazardousCargo
	PerishableCargo
	ValuableCargo
)

func (st ShipmentType) String() string {
	switch st {
	case GeneralCargo:
		return "general"
	case FragileCargo:
		return "fragile"
	case HazardousCargo:
		return "hazardous"
	case PerishableCargo:
		return "perishable"
	case ValuableCargo:
		return "valuable"
	default:
		return "unknown"
	}
}

// ShipmentStatus represents the current status of a shipment
type ShipmentStatus int

const (
	ShipmentPending ShipmentStatus = iota
	ShipmentLoaded
	ShipmentInTransit
	ShipmentUnloaded
	ShipmentDelivered
)

func (ss ShipmentStatus) String() string {
	switch ss {
	case ShipmentPending:
		return "pending"
	case ShipmentLoaded:
		return "loaded"
	case ShipmentInTransit:
		return "in_transit"
	case ShipmentUnloaded:
		return "unloaded"
	case ShipmentDelivered:
		return "delivered"
	default:
		return "unknown"
	}
}

// Dimensions represents the physical dimensions of a shipment
type Dimensions struct {
	Length float64 `json:"length"` // in meters
	Width  float64 `json:"width"`  // in meters
	Height float64 `json:"height"` // in meters
}

// Volume calculates the volume in cubic meters
func (d Dimensions) Volume() float64 {
	return d.Length * d.Width * d.Height
}

// Validate checks if dimensions are valid
func (d Dimensions) Validate() error {
	if d.Length <= 0 {
		return fmt.Errorf("length must be positive")
	}
	if d.Width <= 0 {
		return fmt.Errorf("width must be positive")
	}
	if d.Height <= 0 {
		return fmt.Errorf("height must be positive")
	}
	return nil
}

// Shipment represents an individual item being transported
type Shipment struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Type        ShipmentType   `json:"type"`
	Status      ShipmentStatus `json:"status"`
	Weight      float64        `json:"weight"`      // in kilograms
	Dimensions  Dimensions     `json:"dimensions"`
	Value       float64        `json:"value"`       // monetary value
	Origin      string         `json:"origin"`
	Destination string         `json:"destination"`
	
	// Timing information
	CreatedAt    time.Time  `json:"created_at"`
	LoadedAt     *time.Time `json:"loaded_at,omitempty"`
	UnloadedAt   *time.Time `json:"unloaded_at,omitempty"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	
	// Special handling requirements
	RequiresRefrigeration bool    `json:"requires_refrigeration"`
	MaxTemperature       float64 `json:"max_temperature,omitempty"`
	MinTemperature       float64 `json:"min_temperature,omitempty"`
	HandlingInstructions string  `json:"handling_instructions,omitempty"`
}

// NewShipment creates a new shipment
func NewShipment(id, description string, shipmentType ShipmentType, weight float64, dimensions Dimensions, value float64, origin, destination string) (*Shipment, error) {
	if id == "" {
		return nil, fmt.Errorf("shipment ID cannot be empty")
	}
	if description == "" {
		return nil, fmt.Errorf("shipment description cannot be empty")
	}
	if weight <= 0 {
		return nil, fmt.Errorf("shipment weight must be positive")
	}
	if err := dimensions.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dimensions: %w", err)
	}
	if origin == "" {
		return nil, fmt.Errorf("origin cannot be empty")
	}
	if destination == "" {
		return nil, fmt.Errorf("destination cannot be empty")
	}

	return &Shipment{
		ID:          id,
		Description: description,
		Type:        shipmentType,
		Status:      ShipmentPending,
		Weight:      weight,
		Dimensions:  dimensions,
		Value:       value,
		Origin:      origin,
		Destination: destination,
		CreatedAt:   time.Now(),
	}, nil
}

// Load marks the shipment as loaded
func (s *Shipment) Load() error {
	if s.Status != ShipmentPending {
		return fmt.Errorf("shipment %s is not in pending status, current status: %s", s.ID, s.Status.String())
	}
	
	now := time.Now()
	s.Status = ShipmentLoaded
	s.LoadedAt = &now
	return nil
}

// StartTransit marks the shipment as in transit
func (s *Shipment) StartTransit() error {
	if s.Status != ShipmentLoaded {
		return fmt.Errorf("shipment %s is not loaded, current status: %s", s.ID, s.Status.String())
	}
	
	s.Status = ShipmentInTransit
	return nil
}

// Unload marks the shipment as unloaded
func (s *Shipment) Unload() error {
	if s.Status != ShipmentInTransit {
		return fmt.Errorf("shipment %s is not in transit, current status: %s", s.ID, s.Status.String())
	}
	
	now := time.Now()
	s.Status = ShipmentUnloaded
	s.UnloadedAt = &now
	return nil
}

// Deliver marks the shipment as delivered
func (s *Shipment) Deliver() error {
	if s.Status != ShipmentUnloaded {
		return fmt.Errorf("shipment %s is not unloaded, current status: %s", s.ID, s.Status.String())
	}
	
	now := time.Now()
	s.Status = ShipmentDelivered
	s.DeliveredAt = &now
	return nil
}

// GetTransportDuration calculates the time spent in transport
func (s *Shipment) GetTransportDuration() time.Duration {
	if s.LoadedAt == nil || s.UnloadedAt == nil {
		return 0
	}
	return s.UnloadedAt.Sub(*s.LoadedAt)
}

// IsHazardous checks if the shipment requires special handling
func (s *Shipment) IsHazardous() bool {
	return s.Type == HazardousCargo
}

// IsFragile checks if the shipment is fragile
func (s *Shipment) IsFragile() bool {
	return s.Type == FragileCargo
}

// IsPerishable checks if the shipment is perishable
func (s *Shipment) IsPerishable() bool {
	return s.Type == PerishableCargo || s.RequiresRefrigeration
}

// GetVolume returns the volume of the shipment
func (s *Shipment) GetVolume() float64 {
	return s.Dimensions.Volume()
}

// Validate checks if the shipment data is valid
func (s *Shipment) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("shipment ID cannot be empty")
	}
	if s.Description == "" {
		return fmt.Errorf("shipment description cannot be empty")
	}
	if s.Weight <= 0 {
		return fmt.Errorf("shipment weight must be positive")
	}
	if err := s.Dimensions.Validate(); err != nil {
		return fmt.Errorf("invalid dimensions: %w", err)
	}
	if s.Origin == "" {
		return fmt.Errorf("origin cannot be empty")
	}
	if s.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}
	
	// Temperature validation for refrigerated shipments
	if s.RequiresRefrigeration {
		if s.MaxTemperature <= s.MinTemperature {
			return fmt.Errorf("max temperature must be greater than min temperature")
		}
	}
	
	return nil
}

// Clone creates a deep copy of the shipment
func (s *Shipment) Clone() *Shipment {
	clone := *s
	
	// Copy pointer fields
	if s.LoadedAt != nil {
		loadedAt := *s.LoadedAt
		clone.LoadedAt = &loadedAt
	}
	if s.UnloadedAt != nil {
		unloadedAt := *s.UnloadedAt
		clone.UnloadedAt = &unloadedAt
	}
	if s.DeliveredAt != nil {
		deliveredAt := *s.DeliveredAt
		clone.DeliveredAt = &deliveredAt
	}
	
	return &clone
}
