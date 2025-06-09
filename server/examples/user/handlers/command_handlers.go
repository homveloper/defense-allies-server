package handlers

import (
	"context"
	"cqrs"
	"fmt"
	"time"

	"defense-allies-server/examples/user/domain"
)

// UserCommandHandler handles user-related commands
type UserCommandHandler struct {
	*cqrs.BaseCommandHandler
	repository cqrs.Repository
	eventBus   cqrs.EventBus
}

// NewUserCommandHandler creates a new UserCommandHandler
func NewUserCommandHandler(repository cqrs.Repository, eventBus cqrs.EventBus) *UserCommandHandler {
	handler := &UserCommandHandler{
		BaseCommandHandler: cqrs.NewBaseCommandHandler("UserCommandHandler", []string{}),
		repository:         repository,
		eventBus:           eventBus,
	}

	// Register supported command types
	handler.AddCommandType(domain.CreateUserCommandType)
	handler.AddCommandType(domain.ChangeEmailCommandType)
	handler.AddCommandType(domain.DeactivateUserCommandType)
	handler.AddCommandType(domain.ActivateUserCommandType)

	return handler
}

// Handle handles the command
func (h *UserCommandHandler) Handle(ctx context.Context, command cqrs.Command) (*cqrs.CommandResult, error) {
	startTime := time.Now()

	// Validate command
	if err := command.Validate(); err != nil {
		return &cqrs.CommandResult{
			Success:       false,
			Error:         fmt.Errorf("command validation failed: %w", err),
			AggregateID:   command.ID(),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	var result *cqrs.CommandResult
	var err error

	switch cmd := command.(type) {
	case *domain.CreateUserCommand:
		result, err = h.handleCreateUser(ctx, cmd)
	case *domain.ChangeEmailCommand:
		result, err = h.handleChangeEmail(ctx, cmd)
	case *domain.DeactivateUserCommand:
		result, err = h.handleDeactivateUser(ctx, cmd)
	case *domain.ActivateUserCommand:
		result, err = h.handleActivateUser(ctx, cmd)
	default:
		return &cqrs.CommandResult{
			Success:       false,
			Error:         fmt.Errorf("unsupported command type: %T", command),
			AggregateID:   command.ID(),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	if err != nil {
		return &cqrs.CommandResult{
			Success:       false,
			Error:         err,
			AggregateID:   command.ID(),
			ExecutionTime: time.Since(startTime),
		}, nil
	}

	result.ExecutionTime = time.Since(startTime)
	return result, nil
}

// handleCreateUser handles CreateUserCommand
func (h *UserCommandHandler) handleCreateUser(ctx context.Context, cmd *domain.CreateUserCommand) (*cqrs.CommandResult, error) {
	// Check if user already exists
	if h.repository.Exists(ctx, cmd.ID()) {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("user with ID %s already exists", cmd.ID()),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Create new user aggregate
	user, err := domain.NewUser(cmd.ID(), cmd.Email, cmd.Name)
	if err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to create user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Save the aggregate (for new user, expected version should be 0)
	if err := h.repository.Save(ctx, user, 0); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to save user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Publish events
	events := user.GetChanges()
	if err := h.publishEvents(ctx, events); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to publish events: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	return &cqrs.CommandResult{
		Success:     true,
		Events:      events,
		AggregateID: cmd.ID(),
		Version:     user.Version(),
		Data: map[string]interface{}{
			"user_id": user.ID(),
			"email":   user.Email(),
			"name":    user.Name(),
			"status":  user.Status().String(),
		},
	}, nil
}

// handleChangeEmail handles ChangeEmailCommand
func (h *UserCommandHandler) handleChangeEmail(ctx context.Context, cmd *domain.ChangeEmailCommand) (*cqrs.CommandResult, error) {
	// Load user aggregate
	aggregate, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to load user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	user, ok := aggregate.(*domain.User)
	if !ok {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("invalid aggregate type: expected *domain.User, got %T", aggregate),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Debug: Print version info
	fmt.Printf("DEBUG: Before ChangeEmail - Original: %d, Current: %d\n", user.OriginalVersion(), user.Version())

	// Execute business logic
	if err := user.ChangeEmail(cmd.NewEmail); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to change email: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Debug: Print version info after change
	fmt.Printf("DEBUG: After ChangeEmail - Original: %d, Current: %d\n", user.OriginalVersion(), user.Version())

	// Save the aggregate
	if err := h.repository.Save(ctx, user, user.OriginalVersion()); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to save user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Publish events
	events := user.GetChanges()
	if err := h.publishEvents(ctx, events); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to publish events: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	return &cqrs.CommandResult{
		Success:     true,
		Events:      events,
		AggregateID: cmd.ID(),
		Version:     user.Version(),
		Data: map[string]interface{}{
			"user_id": user.ID(),
			"email":   user.Email(),
		},
	}, nil
}

// handleDeactivateUser handles DeactivateUserCommand
func (h *UserCommandHandler) handleDeactivateUser(ctx context.Context, cmd *domain.DeactivateUserCommand) (*cqrs.CommandResult, error) {
	// Load user aggregate
	aggregate, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to load user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	user, ok := aggregate.(*domain.User)
	if !ok {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("invalid aggregate type: expected *domain.User, got %T", aggregate),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Execute business logic
	if err := user.Deactivate(cmd.Reason); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to deactivate user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Save the aggregate
	if err := h.repository.Save(ctx, user, user.OriginalVersion()); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to save user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Publish events
	events := user.GetChanges()
	if err := h.publishEvents(ctx, events); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to publish events: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	return &cqrs.CommandResult{
		Success:     true,
		Events:      events,
		AggregateID: cmd.ID(),
		Version:     user.Version(),
		Data: map[string]interface{}{
			"user_id": user.ID(),
			"status":  user.Status().String(),
			"reason":  user.DeactivationReason(),
		},
	}, nil
}

// handleActivateUser handles ActivateUserCommand
func (h *UserCommandHandler) handleActivateUser(ctx context.Context, cmd *domain.ActivateUserCommand) (*cqrs.CommandResult, error) {
	// Load user aggregate
	aggregate, err := h.repository.GetByID(ctx, cmd.ID())
	if err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to load user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	user, ok := aggregate.(*domain.User)
	if !ok {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("invalid aggregate type: expected *domain.User, got %T", aggregate),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Execute business logic
	if err := user.Activate(); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to activate user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Save the aggregate
	if err := h.repository.Save(ctx, user, user.OriginalVersion()); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to save user: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	// Publish events
	events := user.GetChanges()
	if err := h.publishEvents(ctx, events); err != nil {
		return &cqrs.CommandResult{
			Success:     false,
			Error:       fmt.Errorf("failed to publish events: %w", err),
			AggregateID: cmd.ID(),
		}, nil
	}

	return &cqrs.CommandResult{
		Success:     true,
		Events:      events,
		AggregateID: cmd.ID(),
		Version:     user.Version(),
		Data: map[string]interface{}{
			"user_id": user.ID(),
			"status":  user.Status().String(),
		},
	}, nil
}

// publishEvents publishes events to the event bus
func (h *UserCommandHandler) publishEvents(ctx context.Context, events []cqrs.EventMessage) error {
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			return fmt.Errorf("failed to publish event %s: %w", event.EventType(), err)
		}
	}
	return nil
}
