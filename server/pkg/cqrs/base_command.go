package cqrs

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BaseCommand provides a base implementation of Command interface
type BaseCommand struct {
	commandID     string
	commandType   string
	aggregateID   string
	aggregateType string
	timestamp     time.Time
	userID        string
	correlationID string
	data          interface{}
}

// NewBaseCommand creates a new BaseCommand
func NewBaseCommand(commandType, aggregateID, aggregateType string, data interface{}) *BaseCommand {
	return &BaseCommand{
		commandID:     uuid.New().String(),
		commandType:   commandType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		timestamp:     time.Now(),
		data:          data,
	}
}

// Command interface implementation

func (c *BaseCommand) CommandID() string {
	return c.commandID
}

func (c *BaseCommand) CommandType() string {
	return c.commandType
}

func (c *BaseCommand) ID() string {
	return c.aggregateID
}

func (c *BaseCommand) Type() string {
	return c.aggregateType
}

func (c *BaseCommand) Timestamp() time.Time {
	return c.timestamp
}

func (c *BaseCommand) UserID() string {
	return c.userID
}

func (c *BaseCommand) CorrelationID() string {
	return c.correlationID
}

func (c *BaseCommand) GetData() interface{} {
	return c.data
}

func (c *BaseCommand) Validate() error {
	if c.commandID == "" {
		return fmt.Errorf("command ID cannot be empty")
	}
	if c.commandType == "" {
		return fmt.Errorf("command type cannot be empty")
	}
	if c.aggregateID == "" {
		return fmt.Errorf("aggregate ID cannot be empty")
	}
	if c.aggregateType == "" {
		return fmt.Errorf("aggregate type cannot be empty")
	}
	return nil
}

// Helper methods

// SetCommandID sets the command ID (used when loading from storage)
func (c *BaseCommand) SetCommandID(commandID string) {
	c.commandID = commandID
}

// SetTimestamp sets the timestamp (used when loading from storage)
func (c *BaseCommand) SetTimestamp(timestamp time.Time) {
	c.timestamp = timestamp
}

// SetUserID sets the user ID
func (c *BaseCommand) SetUserID(userID string) {
	c.userID = userID
}

// SetCorrelationID sets the correlation ID
func (c *BaseCommand) SetCorrelationID(correlationID string) {
	c.correlationID = correlationID
}

// SetData sets the command data
func (c *BaseCommand) SetData(data interface{}) {
	c.data = data
}

// GetCommandInfo returns basic command information as a map
func (c *BaseCommand) GetCommandInfo() map[string]interface{} {
	return map[string]interface{}{
		"command_id":     c.commandID,
		"command_type":   c.commandType,
		"aggregate_id":   c.aggregateID,
		"aggregate_type": c.aggregateType,
		"timestamp":      c.timestamp,
		"user_id":        c.userID,
		"correlation_id": c.correlationID,
	}
}
