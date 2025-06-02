package cqrs

import (
	"context"
	"time"
)

// Command interface (serialization handled separately)
type Command interface {
	// Basic identification information
	CommandID() string     // Unique command ID
	CommandType() string   // Command type
	AggregateID() string   // Target aggregate ID
	AggregateType() string // Target aggregate type

	// Metadata
	Timestamp() time.Time  // Command creation time
	UserID() string        // Command executing user
	CorrelationID() string // Correlation tracking ID

	// Validation
	Validate() error // Command validation

	// Data access (serialization handled separately)
	GetData() interface{} // Command data
}

// CommandResult represents the result of command execution
type CommandResult struct {
	Success       bool              // Success status
	Error         error             // Error information
	Events        []EventMessage    // Generated events
	AggregateID   string           // Processed aggregate ID
	Version       int              // Aggregate version after processing
	Data          interface{}      // Response data (if needed)
	ExecutionTime time.Duration    // Execution time
}

// CommandHandler interface for handling commands
type CommandHandler interface {
	Handle(ctx context.Context, command Command) (*CommandResult, error)
	CanHandle(commandType string) bool
	GetHandlerName() string
}

// CommandDispatcher interface for dispatching commands
type CommandDispatcher interface {
	Dispatch(ctx context.Context, command Command) (*CommandResult, error)
	RegisterHandler(commandType string, handler CommandHandler) error
	UnregisterHandler(commandType string) error
}
