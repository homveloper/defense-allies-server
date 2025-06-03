package commands

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"

	"github.com/google/uuid"
)

// CreateCargoCommandData contains the data for creating a cargo
type CreateCargoCommandData struct {
	CargoID     string  `json:"cargo_id"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	MaxWeight   float64 `json:"max_weight"`
	MaxVolume   float64 `json:"max_volume"`
}

// CreateCargoCommand represents a command to create a new cargo
type CreateCargoCommand struct {
	*cqrs.BaseCommand
	Data CreateCargoCommandData `json:"data"`
}

// NewCreateCargoCommand creates a new create cargo command
func NewCreateCargoCommand(origin, destination string, maxWeight, maxVolume float64, userID string) *CreateCargoCommand {
	cargoID := uuid.New().String()

	commandData := CreateCargoCommandData{
		CargoID:     cargoID,
		Origin:      origin,
		Destination: destination,
		MaxWeight:   maxWeight,
		MaxVolume:   maxVolume,
	}

	cmd := &CreateCargoCommand{
		BaseCommand: cqrs.NewBaseCommand(
			"CreateCargo",
			cargoID,
			"Cargo",
			commandData,
		),
		Data: commandData,
	}
	cmd.SetUserID(userID)
	return cmd
}

// NewCreateCargoCommandWithID creates a new create cargo command with specific ID
func NewCreateCargoCommandWithID(cargoID, origin, destination string, maxWeight, maxVolume float64, userID string) *CreateCargoCommand {
	commandData := CreateCargoCommandData{
		CargoID:     cargoID,
		Origin:      origin,
		Destination: destination,
		MaxWeight:   maxWeight,
		MaxVolume:   maxVolume,
	}

	cmd := &CreateCargoCommand{
		BaseCommand: cqrs.NewBaseCommand(
			"CreateCargo",
			cargoID,
			"Cargo",
			commandData,
		),
		Data: commandData,
	}
	cmd.SetUserID(userID)
	return cmd
}

// GetCargoID returns the cargo ID
func (c *CreateCargoCommand) GetCargoID() string {
	return c.Data.CargoID
}

// GetOrigin returns the origin location
func (c *CreateCargoCommand) GetOrigin() string {
	return c.Data.Origin
}

// GetDestination returns the destination location
func (c *CreateCargoCommand) GetDestination() string {
	return c.Data.Destination
}

// GetMaxWeight returns the maximum weight capacity
func (c *CreateCargoCommand) GetMaxWeight() float64 {
	return c.Data.MaxWeight
}

// GetMaxVolume returns the maximum volume capacity
func (c *CreateCargoCommand) GetMaxVolume() float64 {
	return c.Data.MaxVolume
}

// Validate validates the create cargo command
func (c *CreateCargoCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.Data.CargoID == "" {
		return fmt.Errorf("cargo ID cannot be empty")
	}
	if c.Data.Origin == "" {
		return fmt.Errorf("origin cannot be empty")
	}
	if c.Data.Destination == "" {
		return fmt.Errorf("destination cannot be empty")
	}
	if c.Data.Origin == c.Data.Destination {
		return fmt.Errorf("origin and destination cannot be the same")
	}
	if c.Data.MaxWeight <= 0 {
		return fmt.Errorf("max weight must be positive")
	}
	if c.Data.MaxVolume <= 0 {
		return fmt.Errorf("max volume must be positive")
	}

	return nil
}

// LoadShipmentCommandData contains the data for loading a shipment
type LoadShipmentCommandData struct {
	CargoID      string        `json:"cargo_id"`
	ShipmentID   string        `json:"shipment_id"`
	Description  string        `json:"description"`
	ShipmentType int           `json:"shipment_type"`
	Weight       float64       `json:"weight"`
	Length       float64       `json:"length"`
	Width        float64       `json:"width"`
	Height       float64       `json:"height"`
	Value        float64       `json:"value"`
	Origin       string        `json:"origin"`
	Destination  string        `json:"destination"`
	LoadingTime  time.Duration `json:"loading_time"`

	// Special handling
	RequiresRefrigeration bool    `json:"requires_refrigeration"`
	MaxTemperature        float64 `json:"max_temperature,omitempty"`
	MinTemperature        float64 `json:"min_temperature,omitempty"`
	HandlingInstructions  string  `json:"handling_instructions,omitempty"`
}

// LoadShipmentCommand represents a command to load a shipment into cargo
type LoadShipmentCommand struct {
	*cqrs.BaseCommand
	Data LoadShipmentCommandData `json:"data"`
}

// NewLoadShipmentCommand creates a new load shipment command
func NewLoadShipmentCommand(
	cargoID, shipmentID, description string,
	shipmentType int,
	weight, length, width, height, value float64,
	origin, destination string,
	loadingTime time.Duration,
	userID string,
) *LoadShipmentCommand {
	commandData := LoadShipmentCommandData{
		CargoID:      cargoID,
		ShipmentID:   shipmentID,
		Description:  description,
		ShipmentType: shipmentType,
		Weight:       weight,
		Length:       length,
		Width:        width,
		Height:       height,
		Value:        value,
		Origin:       origin,
		Destination:  destination,
		LoadingTime:  loadingTime,
	}

	cmd := &LoadShipmentCommand{
		BaseCommand: cqrs.NewBaseCommand(
			"LoadShipment",
			cargoID,
			"Cargo",
			commandData,
		),
		Data: commandData,
	}
	cmd.SetUserID(userID)
	return cmd
}

// GetCargoID returns the cargo ID
func (c *LoadShipmentCommand) GetCargoID() string {
	return c.Data.CargoID
}

// GetShipmentID returns the shipment ID
func (c *LoadShipmentCommand) GetShipmentID() string {
	return c.Data.ShipmentID
}

// GetDescription returns the shipment description
func (c *LoadShipmentCommand) GetDescription() string {
	return c.Data.Description
}

// GetWeight returns the shipment weight
func (c *LoadShipmentCommand) GetWeight() float64 {
	return c.Data.Weight
}

// GetVolume returns the shipment volume
func (c *LoadShipmentCommand) GetVolume() float64 {
	return c.Data.Length * c.Data.Width * c.Data.Height
}

// GetLoadingTime returns the loading time
func (c *LoadShipmentCommand) GetLoadingTime() time.Duration {
	return c.Data.LoadingTime
}

// Validate validates the load shipment command
func (c *LoadShipmentCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.Data.CargoID == "" {
		return fmt.Errorf("cargo ID cannot be empty")
	}
	if c.Data.ShipmentID == "" {
		return fmt.Errorf("shipment ID cannot be empty")
	}
	if c.Data.Description == "" {
		return fmt.Errorf("shipment description cannot be empty")
	}
	if c.Data.Weight <= 0 {
		return fmt.Errorf("shipment weight must be positive")
	}
	if c.Data.Length <= 0 || c.Data.Width <= 0 || c.Data.Height <= 0 {
		return fmt.Errorf("shipment dimensions must be positive")
	}
	if c.Data.Origin == "" {
		return fmt.Errorf("shipment origin cannot be empty")
	}
	if c.Data.Destination == "" {
		return fmt.Errorf("shipment destination cannot be empty")
	}
	if c.Data.LoadingTime < 0 {
		return fmt.Errorf("loading time cannot be negative")
	}

	// Temperature validation for refrigerated shipments
	if c.Data.RequiresRefrigeration {
		if c.Data.MaxTemperature <= c.Data.MinTemperature {
			return fmt.Errorf("max temperature must be greater than min temperature")
		}
	}

	return nil
}
