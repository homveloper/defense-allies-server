package cqrs

import (
	"context"
	"fmt"
	"sync"
)

// InMemoryCommandDispatcher provides an in-memory implementation of CommandDispatcher
type InMemoryCommandDispatcher struct {
	handlers map[string]CommandHandler
	mutex    sync.RWMutex
}

// NewInMemoryCommandDispatcher creates a new in-memory command dispatcher
func NewInMemoryCommandDispatcher() *InMemoryCommandDispatcher {
	return &InMemoryCommandDispatcher{
		handlers: make(map[string]CommandHandler),
	}
}

// CommandDispatcher interface implementation

func (d *InMemoryCommandDispatcher) Dispatch(ctx context.Context, command Command) (*CommandResult, error) {
	if command == nil {
		return &CommandResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeCommandValidation.String(), "command cannot be nil", nil),
		}, nil
	}

	// Validate command
	if err := command.Validate(); err != nil {
		return &CommandResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeCommandValidation.String(), "command validation failed", err),
		}, nil
	}

	// Find handler
	d.mutex.RLock()
	handler, exists := d.handlers[command.CommandType()]
	d.mutex.RUnlock()

	if !exists {
		return &CommandResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeCommandValidation.String(), fmt.Sprintf("no handler found for command type: %s", command.CommandType()), ErrCommandHandlerNotFound),
		}, nil
	}

	// Execute command
	return handler.Handle(ctx, command)
}

func (d *InMemoryCommandDispatcher) RegisterHandler(commandType string, handler CommandHandler) error {
	if commandType == "" {
		return NewCQRSError(ErrCodeCommandValidation.String(), "command type cannot be empty", nil)
	}
	if handler == nil {
		return NewCQRSError(ErrCodeCommandValidation.String(), "handler cannot be nil", nil)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, exists := d.handlers[commandType]; exists {
		return NewCQRSError(ErrCodeCommandValidation.String(), fmt.Sprintf("handler already registered for command type: %s", commandType), nil)
	}

	d.handlers[commandType] = handler
	return nil
}

func (d *InMemoryCommandDispatcher) UnregisterHandler(commandType string) error {
	if commandType == "" {
		return NewCQRSError(ErrCodeCommandValidation.String(), "command type cannot be empty", nil)
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, exists := d.handlers[commandType]; !exists {
		return NewCQRSError(ErrCodeCommandValidation.String(), fmt.Sprintf("no handler registered for command type: %s", commandType), ErrCommandHandlerNotFound)
	}

	delete(d.handlers, commandType)
	return nil
}

// Helper methods

// GetRegisteredHandlers returns all registered command types
func (d *InMemoryCommandDispatcher) GetRegisteredHandlers() []string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	types := make([]string, 0, len(d.handlers))
	for commandType := range d.handlers {
		types = append(types, commandType)
	}
	return types
}

// HasHandler checks if a handler is registered for the given command type
func (d *InMemoryCommandDispatcher) HasHandler(commandType string) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	_, exists := d.handlers[commandType]
	return exists
}

// GetHandlerCount returns the number of registered handlers
func (d *InMemoryCommandDispatcher) GetHandlerCount() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return len(d.handlers)
}

// Clear removes all registered handlers
func (d *InMemoryCommandDispatcher) Clear() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.handlers = make(map[string]CommandHandler)
}

// BaseCommandHandler provides a base implementation of CommandHandler
type BaseCommandHandler struct {
	name         string
	commandTypes map[string]bool
}

// NewBaseCommandHandler creates a new base command handler
func NewBaseCommandHandler(name string, commandTypes []string) *BaseCommandHandler {
	typeMap := make(map[string]bool)
	for _, commandType := range commandTypes {
		typeMap[commandType] = true
	}

	return &BaseCommandHandler{
		name:         name,
		commandTypes: typeMap,
	}
}

// CommandHandler interface implementation

func (h *BaseCommandHandler) GetHandlerName() string {
	return h.name
}

func (h *BaseCommandHandler) CanHandle(commandType string) bool {
	return h.commandTypes[commandType]
}

// Handle method should be implemented by concrete handlers
func (h *BaseCommandHandler) Handle(ctx context.Context, command Command) (*CommandResult, error) {
	// Base implementation - should be overridden
	return &CommandResult{
		Success: false,
		Error:   NewCQRSError(ErrCodeCommandValidation.String(), "handle method not implemented", nil),
	}, nil
}

// Helper methods

// AddCommandType adds a command type that this handler can process
func (h *BaseCommandHandler) AddCommandType(commandType string) {
	h.commandTypes[commandType] = true
}

// RemoveCommandType removes a command type from this handler
func (h *BaseCommandHandler) RemoveCommandType(commandType string) {
	delete(h.commandTypes, commandType)
}

// GetSupportedCommandTypes returns all supported command types
func (h *BaseCommandHandler) GetSupportedCommandTypes() []string {
	types := make([]string, 0, len(h.commandTypes))
	for commandType := range h.commandTypes {
		types = append(types, commandType)
	}
	return types
}
