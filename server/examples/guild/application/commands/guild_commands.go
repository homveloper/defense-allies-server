package commands

import (
	"fmt"

	"defense-allies-server/pkg/cqrs"
)

// Command type constants
const (
	// Guild management commands
	CreateGuildCommandType         = "CreateGuild"
	UpdateGuildInfoCommandType     = "UpdateGuildInfo"
	UpdateGuildSettingsCommandType = "UpdateGuildSettings"
	DisbandGuildCommandType        = "DisbandGuild"

	// Member management commands
	InviteMemberCommandType     = "InviteMember"
	AcceptInvitationCommandType = "AcceptInvitation"
	RejectInvitationCommandType = "RejectInvitation"
	LeaveMemberCommandType      = "LeaveMember"
	KickMemberCommandType       = "KickMember"
	PromoteMemberCommandType    = "PromoteMember"
	DemoteMemberCommandType     = "DemoteMember"
)

// Guild Management Commands

// CreateGuildCommand represents a command to create a new guild
type CreateGuildCommand struct {
	*cqrs.BaseCommand
	Name            string `json:"name"`
	Description     string `json:"description"`
	FounderID       string `json:"founder_id"`
	FounderUsername string `json:"founder_username"`
}

// NewCreateGuildCommand creates a new CreateGuildCommand
func NewCreateGuildCommand(guildID, name, description, founderID, founderUsername string) *CreateGuildCommand {
	return &CreateGuildCommand{
		BaseCommand: cqrs.NewBaseCommand(
			CreateGuildCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"name":             name,
				"description":      description,
				"founder_id":       founderID,
				"founder_username": founderUsername,
			},
		),
		Name:            name,
		Description:     description,
		FounderID:       founderID,
		FounderUsername: founderUsername,
	}
}

// Validate validates the create guild command
func (c *CreateGuildCommand) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("guild name cannot be empty")
	}
	if len(c.Name) < 3 || len(c.Name) > 50 {
		return fmt.Errorf("guild name must be between 3 and 50 characters")
	}
	if c.FounderID == "" {
		return fmt.Errorf("founder ID cannot be empty")
	}
	if c.FounderUsername == "" {
		return fmt.Errorf("founder username cannot be empty")
	}
	return nil
}

// UpdateGuildInfoCommand represents a command to update guild information
type UpdateGuildInfoCommand struct {
	*cqrs.BaseCommand
	Name        string `json:"name"`
	Description string `json:"description"`
	Notice      string `json:"notice"`
	Tag         string `json:"tag"`
	UpdatedBy   string `json:"updated_by"`
}

// NewUpdateGuildInfoCommand creates a new UpdateGuildInfoCommand
func NewUpdateGuildInfoCommand(guildID, name, description, notice, tag, updatedBy string) *UpdateGuildInfoCommand {
	return &UpdateGuildInfoCommand{
		BaseCommand: cqrs.NewBaseCommand(
			UpdateGuildInfoCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"name":        name,
				"description": description,
				"notice":      notice,
				"tag":         tag,
				"updated_by":  updatedBy,
			},
		),
		Name:        name,
		Description: description,
		Notice:      notice,
		Tag:         tag,
		UpdatedBy:   updatedBy,
	}
}

// Validate validates the update guild info command
func (c *UpdateGuildInfoCommand) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("guild name cannot be empty")
	}
	if len(c.Name) < 3 || len(c.Name) > 50 {
		return fmt.Errorf("guild name must be between 3 and 50 characters")
	}
	if c.UpdatedBy == "" {
		return fmt.Errorf("updated by cannot be empty")
	}
	if len(c.Tag) > 10 {
		return fmt.Errorf("guild tag cannot be longer than 10 characters")
	}
	return nil
}

// UpdateGuildSettingsCommand represents a command to update guild settings
type UpdateGuildSettingsCommand struct {
	*cqrs.BaseCommand
	MaxMembers      int    `json:"max_members"`
	MinLevel        int    `json:"min_level"`
	IsPublic        bool   `json:"is_public"`
	RequireApproval bool   `json:"require_approval"`
	UpdatedBy       string `json:"updated_by"`
}

// NewUpdateGuildSettingsCommand creates a new UpdateGuildSettingsCommand
func NewUpdateGuildSettingsCommand(guildID string, maxMembers, minLevel int, isPublic, requireApproval bool, updatedBy string) *UpdateGuildSettingsCommand {
	return &UpdateGuildSettingsCommand{
		BaseCommand: cqrs.NewBaseCommand(
			UpdateGuildSettingsCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"max_members":      maxMembers,
				"min_level":        minLevel,
				"is_public":        isPublic,
				"require_approval": requireApproval,
				"updated_by":       updatedBy,
			},
		),
		MaxMembers:      maxMembers,
		MinLevel:        minLevel,
		IsPublic:        isPublic,
		RequireApproval: requireApproval,
		UpdatedBy:       updatedBy,
	}
}

// Validate validates the update guild settings command
func (c *UpdateGuildSettingsCommand) Validate() error {
	if c.MaxMembers < 1 || c.MaxMembers > 200 {
		return fmt.Errorf("max members must be between 1 and 200")
	}
	if c.MinLevel < 1 || c.MinLevel > 100 {
		return fmt.Errorf("min level must be between 1 and 100")
	}
	if c.UpdatedBy == "" {
		return fmt.Errorf("updated by cannot be empty")
	}
	return nil
}

// Member Management Commands

// InviteMemberCommand represents a command to invite a member to the guild
type InviteMemberCommand struct {
	*cqrs.BaseCommand
	Username  string `json:"username"`
	InvitedBy string `json:"invited_by"`
}

// NewInviteMemberCommand creates a new InviteMemberCommand
func NewInviteMemberCommand(guildID, userID, username, invitedBy string) *InviteMemberCommand {
	cmd := &InviteMemberCommand{
		BaseCommand: cqrs.NewBaseCommand(
			InviteMemberCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"user_id":    userID,
				"username":   username,
				"invited_by": invitedBy,
			},
		),
		Username:  username,
		InvitedBy: invitedBy,
	}

	cmd.SetUserID(userID)
	return cmd
}

// Validate validates the invite member command
func (c *InviteMemberCommand) Validate() error {
	if c.UserID() == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if c.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if c.InvitedBy == "" {
		return fmt.Errorf("invited by cannot be empty")
	}
	if c.UserID() == c.InvitedBy {
		return fmt.Errorf("cannot invite yourself")
	}
	return nil
}

// AcceptInvitationCommand represents a command to accept a guild invitation
type AcceptInvitationCommand struct {
	*cqrs.BaseCommand
}

// NewAcceptInvitationCommand creates a new AcceptInvitationCommand
func NewAcceptInvitationCommand(guildID, userID string) *AcceptInvitationCommand {
	cmd := &AcceptInvitationCommand{
		BaseCommand: cqrs.NewBaseCommand(
			AcceptInvitationCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"user_id": userID,
			},
		),
	}

	cmd.SetUserID(userID)
	return cmd
}

// Validate validates the accept invitation command
func (c *AcceptInvitationCommand) Validate() error {
	if c.UserID() == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	return nil
}

// KickMemberCommand represents a command to kick a member from the guild
type KickMemberCommand struct {
	*cqrs.BaseCommand
	KickedBy string `json:"kicked_by"`
	Reason   string `json:"reason"`
}

// NewKickMemberCommand creates a new KickMemberCommand
func NewKickMemberCommand(guildID, userID, kickedBy, reason string) *KickMemberCommand {
	cmd := &KickMemberCommand{
		BaseCommand: cqrs.NewBaseCommand(
			KickMemberCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"user_id":   userID,
				"kicked_by": kickedBy,
				"reason":    reason,
			},
		),
		KickedBy: kickedBy,
		Reason:   reason,
	}
	cmd.SetUserID(userID)
	return cmd
}

// Validate validates the kick member command
func (c *KickMemberCommand) Validate() error {
	if c.UserID() == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if c.KickedBy == "" {
		return fmt.Errorf("kicked by cannot be empty")
	}
	if c.UserID() == c.KickedBy {
		return fmt.Errorf("cannot kick yourself")
	}
	return nil
}

// PromoteMemberCommand represents a command to promote a member
type PromoteMemberCommand struct {
	*cqrs.BaseCommand
	NewRole    string `json:"new_role"`
	PromotedBy string `json:"promoted_by"`
}

// NewPromoteMemberCommand creates a new PromoteMemberCommand
func NewPromoteMemberCommand(guildID, userID, newRole, promotedBy string) *PromoteMemberCommand {
	cmd := &PromoteMemberCommand{
		BaseCommand: cqrs.NewBaseCommand(
			PromoteMemberCommandType,
			guildID,
			"Guild",
			map[string]interface{}{
				"user_id":     userID,
				"new_role":    newRole,
				"promoted_by": promotedBy,
			},
		),
		NewRole:    newRole,
		PromotedBy: promotedBy,
	}

	cmd.SetUserID(userID)
	return cmd
}

// Validate validates the promote member command
func (c *PromoteMemberCommand) Validate() error {
	if c.UserID() == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if c.NewRole == "" {
		return fmt.Errorf("new role cannot be empty")
	}
	if c.PromotedBy == "" {
		return fmt.Errorf("promoted by cannot be empty")
	}
	if c.UserID() == c.PromotedBy {
		return fmt.Errorf("cannot promote yourself")
	}
	return nil
}
