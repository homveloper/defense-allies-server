package domain

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// User represents a user aggregate in the system
type User struct {
	*cqrs.BaseAggregate
	name        string
	email       string
	totalOrders int
	totalSpent  decimal.Decimal
	isVIP       bool
	createdAt   time.Time
	updatedAt   time.Time
}

// NewUser creates a new User aggregate
func NewUser(id, name, email string) *User {
	user := &User{
		BaseAggregate: cqrs.NewBaseAggregate(id, "User"),
		name:          name,
		email:         email,
		totalOrders:   0,
		totalSpent:    decimal.Zero,
		isVIP:         false,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}

	// Apply creation event
	event := NewUserCreated(id, name, email)
	user.ApplyEvent(event)

	return user
}

// LoadUserFromHistory loads a User aggregate from event history
func LoadUserFromHistory(id string, events []cqrs.EventMessage) (*User, error) {
	user := &User{
		BaseAggregate: cqrs.NewBaseAggregate(id, "User"),
		totalSpent:    decimal.Zero,
	}

	for _, event := range events {
		if err := user.applyDomainEvent(event); err != nil {
			return nil, fmt.Errorf("failed to apply event: %w", err)
		}
	}

	return user, nil
}

// Business Methods

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(name, email string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if u.name == name && u.email == email {
		return nil // No changes
	}

	event := NewUserUpdated(u.AggregateID(), name, email)
	u.ApplyEvent(event)

	return nil
}

// RecordOrderCompletion records that the user completed an order
func (u *User) RecordOrderCompletion(orderAmount decimal.Decimal) {
	u.totalOrders++
	u.totalSpent = u.totalSpent.Add(orderAmount)
	u.updatedAt = time.Now()

	// Check if user becomes VIP (spent more than $1000)
	if u.totalSpent.GreaterThan(decimal.NewFromInt(1000)) && !u.isVIP {
		u.isVIP = true
	}
}

// Getters

func (u *User) GetName() string {
	return u.name
}

func (u *User) GetEmail() string {
	return u.email
}

func (u *User) GetTotalOrders() int {
	return u.totalOrders
}

func (u *User) GetTotalSpent() decimal.Decimal {
	return u.totalSpent
}

func (u *User) IsVIP() bool {
	return u.isVIP
}

func (u *User) GetCreatedAt() time.Time {
	return u.createdAt
}

func (u *User) GetUpdatedAt() time.Time {
	return u.updatedAt
}

// Event Application

// applyDomainEvent applies domain events to the aggregate
func (u *User) applyDomainEvent(event cqrs.EventMessage) error {
	switch e := event.EventData().(type) {
	case *UserCreated:
		return u.applyUserCreated(e)
	case *UserUpdated:
		return u.applyUserUpdated(e)
	default:
		// Ignore unknown events
		return nil
	}
}

// applyUserCreated applies UserCreated event
func (u *User) applyUserCreated(event *UserCreated) error {
	u.name = event.Name
	u.email = event.Email
	u.totalOrders = 0
	u.totalSpent = decimal.Zero
	u.isVIP = false
	u.createdAt = event.Timestamp()
	u.updatedAt = event.Timestamp()
	return nil
}

// applyUserUpdated applies UserUpdated event
func (u *User) applyUserUpdated(event *UserUpdated) error {
	u.name = event.Name
	u.email = event.Email
	u.updatedAt = event.Timestamp()
	return nil
}

// Validation

// Validate validates the user aggregate state
func (u *User) Validate() error {
	if u.AggregateID() == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if u.name == "" {
		return fmt.Errorf("user name cannot be empty")
	}
	if u.email == "" {
		return fmt.Errorf("user email cannot be empty")
	}
	if u.totalOrders < 0 {
		return fmt.Errorf("total orders cannot be negative")
	}
	if u.totalSpent.IsNegative() {
		return fmt.Errorf("total spent cannot be negative")
	}
	return nil
}

// Repository Interface

// UserRepository defines the interface for user persistence
type UserRepository interface {
	cqrs.EventSourcedRepository

	// FindByEmail finds a user by email address
	FindByEmail(ctx context.Context, email string) (*User, error)

	// GetUserStats gets user statistics
	GetUserStats(ctx context.Context, userID string) (*UserStats, error)
}

// UserStats represents user statistics
type UserStats struct {
	UserID      string          `json:"user_id"`
	TotalOrders int             `json:"total_orders"`
	TotalSpent  decimal.Decimal `json:"total_spent"`
	IsVIP       bool            `json:"is_vip"`
	LastOrderAt *time.Time      `json:"last_order_at,omitempty"`
}

// Commands

// CreateUserCommand represents a command to create a user
type CreateUserCommand struct {
	*cqrs.BaseCommand
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewCreateUserCommand creates a new CreateUserCommand
func NewCreateUserCommand(userID, name, email string) *CreateUserCommand {
	return &CreateUserCommand{
		BaseCommand: cqrs.NewBaseCommand("CreateUser", userID, "User", nil),
		Name:        name,
		Email:       email,
	}
}

// UpdateUserCommand represents a command to update a user
type UpdateUserCommand struct {
	*cqrs.BaseCommand
	Name  string `json:"name"`
	Email string `json:"email"`
}

// NewUpdateUserCommand creates a new UpdateUserCommand
func NewUpdateUserCommand(userID, name, email string) *UpdateUserCommand {
	return &UpdateUserCommand{
		BaseCommand: cqrs.NewBaseCommand("UpdateUser", userID, "User", nil),
		Name:        name,
		Email:       email,
	}
}

// Command Handlers

// UserCommandHandler handles user-related commands
type UserCommandHandler struct {
	repository UserRepository
}

// NewUserCommandHandler creates a new UserCommandHandler
func NewUserCommandHandler(repository UserRepository) *UserCommandHandler {
	return &UserCommandHandler{
		repository: repository,
	}
}

// Handle handles user commands
func (h *UserCommandHandler) Handle(ctx context.Context, command cqrs.Command) (interface{}, error) {
	switch cmd := command.(type) {
	case *CreateUserCommand:
		return h.handleCreateUser(ctx, cmd)
	case *UpdateUserCommand:
		return h.handleUpdateUser(ctx, cmd)
	default:
		return nil, fmt.Errorf("unknown command type: %T", command)
	}
}

// handleCreateUser handles CreateUserCommand
func (h *UserCommandHandler) handleCreateUser(ctx context.Context, cmd *CreateUserCommand) (*User, error) {
	// Check if repository is available
	if h.repository != nil {
		// Check if user already exists
		existing, err := h.repository.FindByEmail(ctx, cmd.Email)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("user with email %s already exists", cmd.Email)
		}
	}

	// Create new user
	user := NewUser(cmd.AggregateID(), cmd.Name, cmd.Email)

	// Validate
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Save if repository is available (버전 관리 자동화)
	if h.repository != nil {
		if err := h.repository.Save(ctx, user, 0); err != nil {
			return nil, fmt.Errorf("failed to save user: %w", err)
		}
	}

	return user, nil
}

// handleUpdateUser handles UpdateUserCommand
func (h *UserCommandHandler) handleUpdateUser(ctx context.Context, cmd *UpdateUserCommand) (*User, error) {
	// Check if repository is available
	if h.repository == nil {
		return nil, fmt.Errorf("repository not available")
	}

	// Load user
	user, err := h.repository.GetByID(ctx, cmd.AggregateID())
	if err != nil {
		return nil, fmt.Errorf("failed to load user: %w", err)
	}

	userAggregate, ok := user.(*User)
	if !ok {
		return nil, fmt.Errorf("invalid aggregate type")
	}

	// Update profile
	if err := userAggregate.UpdateProfile(cmd.Name, cmd.Email); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Save (버전 관리 자동화)
	if err := h.repository.Save(ctx, userAggregate, 0); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return userAggregate, nil
}
