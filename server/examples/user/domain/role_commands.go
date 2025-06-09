package domain

import (
	"fmt"
	"time"

	"cqrs"
)

// AssignRoleCommand represents a command to assign a role to a user
type AssignRoleCommand struct {
	*cqrs.BaseCommand
	RoleType   RoleType `json:"role_type"`
	AssignedBy string   `json:"assigned_by"`
}

// NewAssignRoleCommand creates a new AssignRoleCommand
func NewAssignRoleCommand(userID string, roleType RoleType, assignedBy string) *AssignRoleCommand {
	return &AssignRoleCommand{
		BaseCommand: cqrs.NewBaseCommand(
			AssignRoleCommandType,
			userID,
			"User",
			map[string]interface{}{
				"role_type":   roleType.String(),
				"assigned_by": assignedBy,
			},
		),
		RoleType:   roleType,
		AssignedBy: assignedBy,
	}
}

// Validate validates the AssignRoleCommand
func (c *AssignRoleCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.AssignedBy == "" {
		return fmt.Errorf("assigned_by cannot be empty")
	}

	return nil
}

// AssignRoleWithExpiryCommand represents a command to assign a role with expiry to a user
type AssignRoleWithExpiryCommand struct {
	*cqrs.BaseCommand
	RoleType   RoleType  `json:"role_type"`
	AssignedBy string    `json:"assigned_by"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// NewAssignRoleWithExpiryCommand creates a new AssignRoleWithExpiryCommand
func NewAssignRoleWithExpiryCommand(userID string, roleType RoleType, assignedBy string, expiresAt time.Time) *AssignRoleWithExpiryCommand {
	return &AssignRoleWithExpiryCommand{
		BaseCommand: cqrs.NewBaseCommand(
			AssignRoleWithExpiryCommandType,
			userID,
			"User",
			map[string]interface{}{
				"role_type":   roleType.String(),
				"assigned_by": assignedBy,
				"expires_at":  expiresAt,
			},
		),
		RoleType:   roleType,
		AssignedBy: assignedBy,
		ExpiresAt:  expiresAt,
	}
}

// Validate validates the AssignRoleWithExpiryCommand
func (c *AssignRoleWithExpiryCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.AssignedBy == "" {
		return fmt.Errorf("assigned_by cannot be empty")
	}

	if c.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expires_at cannot be in the past")
	}

	return nil
}

// RevokeRoleCommand represents a command to revoke a role from a user
type RevokeRoleCommand struct {
	*cqrs.BaseCommand
	RoleType  RoleType `json:"role_type"`
	RevokedBy string   `json:"revoked_by"`
}

// NewRevokeRoleCommand creates a new RevokeRoleCommand
func NewRevokeRoleCommand(userID string, roleType RoleType, revokedBy string) *RevokeRoleCommand {
	return &RevokeRoleCommand{
		BaseCommand: cqrs.NewBaseCommand(
			RevokeRoleCommandType,
			userID,
			"User",
			map[string]interface{}{
				"role_type":  roleType.String(),
				"revoked_by": revokedBy,
			},
		),
		RoleType:  roleType,
		RevokedBy: revokedBy,
	}
}

// Validate validates the RevokeRoleCommand
func (c *RevokeRoleCommand) Validate() error {
	if err := c.BaseCommand.Validate(); err != nil {
		return err
	}

	if c.RevokedBy == "" {
		return fmt.Errorf("revoked_by cannot be empty")
	}

	return nil
}
