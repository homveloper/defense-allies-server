package cqrs

import (
	"context"
	"time"
)

// Command interface defines the contract for all commands in the CQRS system.
// Commands represent user intentions to modify system state and are the "write" side
// of CQRS. They encapsulate all information needed to perform a specific operation
// on an aggregate, including identification, metadata, and validation.
//
// Key principles:
//   - Commands are immutable once created
//   - Commands should be self-validating
//   - Commands contain intent, not implementation details
//   - Serialization is handled separately for flexibility
//
// Implementation guidelines:
//   - Use value objects for command data
//   - Include all necessary information for processing
//   - Implement meaningful validation rules
//   - Provide clear error messages for validation failures
type Command interface {
	// Basic identification information

	// CommandID returns the unique identifier for this command instance.
	// This ID is used for deduplication, logging, and correlation tracking.
	//
	// Returns:
	//   - string: Unique command identifier (typically UUID)
	CommandID() string

	// CommandType returns the type name of this command.
	// This is used for routing to appropriate handlers and serialization.
	//
	// Returns:
	//   - string: Command type name (e.g., "CreateUser", "UpdateOrder")
	CommandType() string

	// string returns the identifier of the target aggregate.
	// This specifies which aggregate instance the command should operate on.
	//
	// Returns:
	//   - string: Target aggregate identifier
	ID() string

	// AggregateType returns the type of the target aggregate.
	// This is used for routing and validation purposes.
	//
	// Returns:
	//   - string: Target aggregate type (e.g., "User", "Order", "Product")
	Type() string

	// Metadata

	// Timestamp returns when this command was created.
	// This is used for auditing, debugging, and temporal analysis.
	//
	// Returns:
	//   - time.Time: Command creation timestamp
	Timestamp() time.Time

	// UserID returns the identifier of the user executing this command.
	// This is used for authorization, auditing, and business logic.
	//
	// Returns:
	//   - string: User identifier (empty string if system-generated)
	UserID() string

	// CorrelationID returns the correlation identifier for request tracking.
	// This links related commands, events, and queries across service boundaries.
	//
	// Returns:
	//   - string: Correlation identifier for distributed tracing
	CorrelationID() string

	// Validation

	// Validate checks if the command satisfies all business rules and constraints.
	// This method should perform comprehensive validation including:
	//   - Required field validation
	//   - Format and range validation
	//   - Business rule validation
	//   - Cross-field validation
	//
	// Returns:
	//   - error: nil if valid, descriptive error if validation fails
	Validate() error

	// Data access (serialization handled separately)

	// GetData returns the command-specific data payload.
	// This contains the actual parameters needed to execute the command.
	// Serialization is handled separately to allow different formats.
	//
	// Returns:
	//   - interface{}: Command data (should be serializable)
	GetData() interface{}
}

// CommandResult represents the outcome of command execution in the CQRS system.
// This structure provides comprehensive information about the command processing,
// including success status, generated events, performance metrics, and error details.
// It serves as the primary feedback mechanism for command execution.
//
// Usage patterns:
//   - Check Success field first to determine outcome
//   - Use Events for event sourcing and projection updates
//   - Monitor ExecutionTime for performance analysis
//   - Use Version for optimistic concurrency control
type CommandResult struct {
	Success       bool           `json:"success"`        // Indicates if command executed successfully
	Error         error          `json:"error"`          // Error details if execution failed
	Events        []EventMessage `json:"events"`         // Events generated during command execution
	string        string         `json:"aggregate_id"`   // ID of the aggregate that was processed
	Version       int            `json:"version"`        // Aggregate version after command execution
	Data          interface{}    `json:"data"`           // Optional response data (e.g., created entity ID)
	ExecutionTime time.Duration  `json:"execution_time"` // Time taken to execute the command
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
