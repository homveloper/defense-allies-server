package domain

import (
	"fmt"
	"regexp"
	"strings"

	"cqrs"
)

// Command type constants
const (
	CreateUserCommandType           = "CreateUser"
	ChangeEmailCommandType          = "ChangeEmail"
	DeactivateUserCommandType       = "DeactivateUser"
	ActivateUserCommandType         = "ActivateUser"
	AssignRoleCommandType           = "AssignRole"
	AssignRoleWithExpiryCommandType = "AssignRoleWithExpiry"
	RevokeRoleCommandType           = "RevokeRole"
	UpdateProfileCommandType        = "UpdateProfile"
	UpdateDisplayNameCommandType    = "UpdateDisplayName"
	UpdateContactInfoCommandType    = "UpdateContactInfo"
	SetAvatarCommandType            = "SetAvatar"
	SetPreferenceCommandType        = "SetPreference"
)

// CreateUserCommand represents a command to create a new user
type CreateUserCommand struct {
	*cqrs.BaseCommand
	UserId string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// NewCreateUserCommand creates a new CreateUserCommand
func NewCreateUserCommand(userID, email, name string) *CreateUserCommand {
	return &CreateUserCommand{
		BaseCommand: cqrs.NewBaseCommand(
			CreateUserCommandType,
			userID,
			"User",
			map[string]interface{}{
				"email": email,
				"name":  name,
			},
		),
		Email: email,
		Name:  name,
	}
}

// Validate validates the CreateUserCommand
func (c *CreateUserCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.Email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if !isValidEmail(c.Email) {
		return fmt.Errorf("invalid email format: %s", c.Email)
	}

	if c.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(c.Name) < 2 || len(c.Name) > 50 {
		return fmt.Errorf("name must be between 2 and 50 characters")
	}

	return nil
}

// ChangeEmailCommand represents a command to change user's email
type ChangeEmailCommand struct {
	*cqrs.BaseCommand
	NewEmail string `json:"new_email"`
}

// NewChangeEmailCommand creates a new ChangeEmailCommand
func NewChangeEmailCommand(userID, newEmail string) *ChangeEmailCommand {
	return &ChangeEmailCommand{
		BaseCommand: cqrs.NewBaseCommand(
			ChangeEmailCommandType,
			userID,
			"User",
			map[string]interface{}{
				"new_email": newEmail,
			},
		),
		NewEmail: newEmail,
	}
}

// Validate validates the ChangeEmailCommand
func (c *ChangeEmailCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.NewEmail == "" {
		return fmt.Errorf("new email cannot be empty")
	}

	if !isValidEmail(c.NewEmail) {
		return fmt.Errorf("invalid email format: %s", c.NewEmail)
	}

	return nil
}

// DeactivateUserCommand represents a command to deactivate a user
type DeactivateUserCommand struct {
	*cqrs.BaseCommand
	Reason string `json:"reason"`
}

// NewDeactivateUserCommand creates a new DeactivateUserCommand
func NewDeactivateUserCommand(userID, reason string) *DeactivateUserCommand {
	return &DeactivateUserCommand{
		BaseCommand: cqrs.NewBaseCommand(
			DeactivateUserCommandType,
			userID,
			"User",
			map[string]interface{}{
				"reason": reason,
			},
		),
		Reason: reason,
	}
}

// Validate validates the DeactivateUserCommand
func (c *DeactivateUserCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.Reason == "" {
		return fmt.Errorf("deactivation reason cannot be empty")
	}

	if len(c.Reason) > 200 {
		return fmt.Errorf("deactivation reason cannot exceed 200 characters")
	}

	return nil
}

// ActivateUserCommand represents a command to activate a user
type ActivateUserCommand struct {
	*cqrs.BaseCommand
}

// NewActivateUserCommand creates a new ActivateUserCommand
func NewActivateUserCommand(userID string) *ActivateUserCommand {
	return &ActivateUserCommand{
		BaseCommand: cqrs.NewBaseCommand(
			ActivateUserCommandType,
			userID,
			"User",
			map[string]interface{}{},
		),
	}
}

// Validate validates the ActivateUserCommand
func (c *ActivateUserCommand) Validate() error {
	return c.BaseCommand.Validate()
}

// Helper functions

// isValidEmail validates email format
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) == 0 || len(email) > 254 {
		return false
	}

	// Simple email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Command factory function for deserialization
func CreateCommandFromType(commandType string, aggregateID string, commandData map[string]interface{}) (cqrs.Command, error) {
	switch commandType {
	case CreateUserCommandType:
		email, ok := commandData["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid email in command data")
		}
		name, ok := commandData["name"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid name in command data")
		}
		return NewCreateUserCommand(aggregateID, email, name), nil

	case ChangeEmailCommandType:
		newEmail, ok := commandData["new_email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid new_email in command data")
		}
		return NewChangeEmailCommand(aggregateID, newEmail), nil

	case DeactivateUserCommandType:
		reason, ok := commandData["reason"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid reason in command data")
		}
		return NewDeactivateUserCommand(aggregateID, reason), nil

	case ActivateUserCommandType:
		return NewActivateUserCommand(aggregateID), nil

	default:
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeCommandValidation.String(), "unknown command type: "+commandType, nil)
	}
}
