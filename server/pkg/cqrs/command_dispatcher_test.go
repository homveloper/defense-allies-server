package cqrs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test command implementation
type TestCommand struct {
	*BaseCommand
	TestData string
}

func NewTestCommand(aggregateID string, testData string) *TestCommand {
	return &TestCommand{
		BaseCommand: NewBaseCommand("TestCommand", aggregateID, "TestAggregate", testData),
		TestData:    testData,
	}
}

// Test command handler implementation
type TestCommandHandler struct {
	*BaseCommandHandler
	HandleFunc func(ctx context.Context, command Command) (*CommandResult, error)
}

func NewTestCommandHandler() *TestCommandHandler {
	return &TestCommandHandler{
		BaseCommandHandler: NewBaseCommandHandler("TestHandler", []string{"TestCommand"}),
	}
}

func (h *TestCommandHandler) Handle(ctx context.Context, command Command) (*CommandResult, error) {
	if h.HandleFunc != nil {
		return h.HandleFunc(ctx, command)
	}

	// Default implementation
	return &CommandResult{
		Success:     true,
		AggregateID: command.ID(),
		Version:     1,
		Events:      []EventMessage{},
	}, nil
}

func TestNewInMemoryCommandDispatcher(t *testing.T) {
	// Act
	dispatcher := NewInMemoryCommandDispatcher()

	// Assert
	assert.NotNil(t, dispatcher)
	assert.Equal(t, 0, dispatcher.GetHandlerCount())
	assert.Empty(t, dispatcher.GetRegisteredHandlers())
}

func TestCommandDispatcher_RegisterHandler(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler := NewTestCommandHandler()

	// Act
	err := dispatcher.RegisterHandler("TestCommand", handler)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, dispatcher.GetHandlerCount())
	assert.True(t, dispatcher.HasHandler("TestCommand"))
	assert.Contains(t, dispatcher.GetRegisteredHandlers(), "TestCommand")
}

func TestCommandDispatcher_RegisterHandler_EmptyCommandType(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler := NewTestCommandHandler()

	// Act
	err := dispatcher.RegisterHandler("", handler)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "command type cannot be empty")
}

func TestCommandDispatcher_RegisterHandler_NilHandler(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()

	// Act
	err := dispatcher.RegisterHandler("TestCommand", nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler cannot be nil")
}

func TestCommandDispatcher_RegisterHandler_Duplicate(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler1 := NewTestCommandHandler()
	handler2 := NewTestCommandHandler()

	// Act
	err1 := dispatcher.RegisterHandler("TestCommand", handler1)
	err2 := dispatcher.RegisterHandler("TestCommand", handler2)

	// Assert
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "handler already registered")
}

func TestCommandDispatcher_UnregisterHandler(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler := NewTestCommandHandler()
	dispatcher.RegisterHandler("TestCommand", handler)

	// Act
	err := dispatcher.UnregisterHandler("TestCommand")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, dispatcher.GetHandlerCount())
	assert.False(t, dispatcher.HasHandler("TestCommand"))
}

func TestCommandDispatcher_UnregisterHandler_NotFound(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()

	// Act
	err := dispatcher.UnregisterHandler("NonExistentCommand")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no handler registered")
}

func TestCommandDispatcher_Dispatch_Success(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler := NewTestCommandHandler()
	command := NewTestCommand("test-id", "test data")

	dispatcher.RegisterHandler("TestCommand", handler)

	// Act
	result, err := dispatcher.Dispatch(context.Background(), command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "test-id", result.AggregateID)
	assert.Equal(t, 1, result.Version)
}

func TestCommandDispatcher_Dispatch_NilCommand(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()

	// Act
	result, err := dispatcher.Dispatch(context.Background(), nil)

	// Assert
	assert.NoError(t, err) // Dispatcher returns result with error, not error itself
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "command cannot be nil")
}

func TestCommandDispatcher_Dispatch_InvalidCommand(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	command := NewTestCommand("", "test data") // Invalid - empty aggregate ID

	// Act
	result, err := dispatcher.Dispatch(context.Background(), command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "command validation failed")
}

func TestCommandDispatcher_Dispatch_NoHandler(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	command := NewTestCommand("test-id", "test data")

	// Act
	result, err := dispatcher.Dispatch(context.Background(), command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "no handler found")
}

func TestCommandDispatcher_Dispatch_HandlerError(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler := NewTestCommandHandler()
	command := NewTestCommand("test-id", "test data")

	// Set up handler to return error
	handler.HandleFunc = func(ctx context.Context, command Command) (*CommandResult, error) {
		return &CommandResult{
			Success: false,
			Error:   NewCQRSError(ErrCodeCommandValidation.String(), "handler error", nil),
		}, nil
	}

	dispatcher.RegisterHandler("TestCommand", handler)

	// Act
	result, err := dispatcher.Dispatch(context.Background(), command)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "handler error")
}

func TestCommandDispatcher_Clear(t *testing.T) {
	// Arrange
	dispatcher := NewInMemoryCommandDispatcher()
	handler := NewTestCommandHandler()
	dispatcher.RegisterHandler("TestCommand", handler)

	// Verify handler is registered
	assert.Equal(t, 1, dispatcher.GetHandlerCount())

	// Act
	dispatcher.Clear()

	// Assert
	assert.Equal(t, 0, dispatcher.GetHandlerCount())
	assert.Empty(t, dispatcher.GetRegisteredHandlers())
}

func TestBaseCommandHandler_CanHandle(t *testing.T) {
	// Arrange
	handler := NewBaseCommandHandler("TestHandler", []string{"Command1", "Command2"})

	// Act & Assert
	assert.True(t, handler.CanHandle("Command1"))
	assert.True(t, handler.CanHandle("Command2"))
	assert.False(t, handler.CanHandle("Command3"))
}

func TestBaseCommandHandler_GetHandlerName(t *testing.T) {
	// Arrange
	handlerName := "TestHandler"
	handler := NewBaseCommandHandler(handlerName, []string{"Command1"})

	// Act & Assert
	assert.Equal(t, handlerName, handler.GetHandlerName())
}

func TestBaseCommandHandler_AddRemoveCommandType(t *testing.T) {
	// Arrange
	handler := NewBaseCommandHandler("TestHandler", []string{"Command1"})

	// Initially can handle Command1 but not Command2
	assert.True(t, handler.CanHandle("Command1"))
	assert.False(t, handler.CanHandle("Command2"))

	// Act - Add Command2
	handler.AddCommandType("Command2")

	// Assert - Now can handle both
	assert.True(t, handler.CanHandle("Command1"))
	assert.True(t, handler.CanHandle("Command2"))

	// Act - Remove Command1
	handler.RemoveCommandType("Command1")

	// Assert - Now can only handle Command2
	assert.False(t, handler.CanHandle("Command1"))
	assert.True(t, handler.CanHandle("Command2"))
}

func TestBaseCommandHandler_GetSupportedCommandTypes(t *testing.T) {
	// Arrange
	commandTypes := []string{"Command1", "Command2", "Command3"}
	handler := NewBaseCommandHandler("TestHandler", commandTypes)

	// Act
	supportedTypes := handler.GetSupportedCommandTypes()

	// Assert
	assert.Len(t, supportedTypes, 3)
	for _, commandType := range commandTypes {
		assert.Contains(t, supportedTypes, commandType)
	}
}
