package handlers

import (
	"context"
	"fmt"

	"defense-allies-server/examples/guild/application/commands"
	"defense-allies-server/examples/guild/domain"
	"defense-allies-server/pkg/cqrs"
)

// GuildCommandHandler handles guild-related commands
type GuildCommandHandler struct {
	*cqrs.BaseCommandHandler
	repository cqrs.EventSourcedRepository
}

// NewGuildCommandHandler creates a new GuildCommandHandler
func NewGuildCommandHandler(repository cqrs.EventSourcedRepository) *GuildCommandHandler {
	supportedCommands := []string{
		commands.CreateGuildCommandType,
		commands.UpdateGuildInfoCommandType,
		commands.UpdateGuildSettingsCommandType,
		commands.InviteMemberCommandType,
		commands.AcceptInvitationCommandType,
		commands.KickMemberCommandType,
		commands.PromoteMemberCommandType,
	}

	return &GuildCommandHandler{
		BaseCommandHandler: cqrs.NewBaseCommandHandler("GuildCommandHandler", supportedCommands),
		repository:         repository,
	}
}

// Handle handles the incoming command
func (h *GuildCommandHandler) Handle(ctx context.Context, command cqrs.Command) (*cqrs.CommandResult, error) {
	// Validate command
	if err := command.Validate(); err != nil {
		return nil, fmt.Errorf("command validation failed: %w", err)
	}

	switch cmd := command.(type) {
	case *commands.CreateGuildCommand:
		return h.handleCreateGuild(ctx, cmd)
	case *commands.UpdateGuildInfoCommand:
		return h.handleUpdateGuildInfo(ctx, cmd)
	case *commands.UpdateGuildSettingsCommand:
		return h.handleUpdateGuildSettings(ctx, cmd)
	case *commands.InviteMemberCommand:
		return h.handleInviteMember(ctx, cmd)
	case *commands.AcceptInvitationCommand:
		return h.handleAcceptInvitation(ctx, cmd)
	case *commands.KickMemberCommand:
		return h.handleKickMember(ctx, cmd)
	case *commands.PromoteMemberCommand:
		return h.handlePromoteMember(ctx, cmd)
	default:
		return nil, fmt.Errorf("unsupported command type: %s", command.CommandType())
	}
}

// handleCreateGuild handles the CreateGuildCommand
func (h *GuildCommandHandler) handleCreateGuild(ctx context.Context, cmd *commands.CreateGuildCommand) (*cqrs.CommandResult, error) {
	// Check if guild already exists
	exists := h.repository.Exists(ctx, cmd.ID())
	if exists {
		return nil, fmt.Errorf("guild with ID %s already exists", cmd.ID())
	}

	// Create new guild aggregate
	guild := domain.NewGuildAggregate(
		cmd.ID(),
		cmd.Name,
		cmd.Description,
		cmd.FounderID,
		cmd.FounderUsername,
	)

	// Validate the guild
	if err := guild.Validate(); err != nil {
		return nil, fmt.Errorf("guild validation failed: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, 0); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"guild_id": cmd.ID(),
			"name":     cmd.Name,
			"message":  "Guild created successfully",
		},
	}, nil
}

// handleUpdateGuildInfo handles the UpdateGuildInfoCommand
func (h *GuildCommandHandler) handleUpdateGuildInfo(ctx context.Context, cmd *commands.UpdateGuildInfoCommand) (*cqrs.CommandResult, error) {
	// Load guild aggregate
	guild, err := h.loadGuild(ctx, cmd.ID())
	if err != nil {
		return nil, err
	}

	// Update guild info
	if err := guild.UpdateInfo(cmd.Name, cmd.Description, cmd.Notice, cmd.Tag, cmd.UpdatedBy); err != nil {
		return nil, fmt.Errorf("failed to update guild info: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"message": "Guild info updated successfully",
		},
	}, nil
}

// handleUpdateGuildSettings handles the UpdateGuildSettingsCommand
func (h *GuildCommandHandler) handleUpdateGuildSettings(ctx context.Context, cmd *commands.UpdateGuildSettingsCommand) (*cqrs.CommandResult, error) {
	// Load guild aggregate
	guild, err := h.loadGuild(ctx, cmd.ID())
	if err != nil {
		return nil, err
	}

	// Update guild settings
	if err := guild.UpdateSettings(cmd.MaxMembers, cmd.MinLevel, cmd.IsPublic, cmd.RequireApproval, cmd.UpdatedBy); err != nil {
		return nil, fmt.Errorf("failed to update guild settings: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"message": "Guild settings updated successfully",
		},
	}, nil
}

// handleInviteMember handles the InviteMemberCommand
func (h *GuildCommandHandler) handleInviteMember(ctx context.Context, cmd *commands.InviteMemberCommand) (*cqrs.CommandResult, error) {
	// Load guild aggregate
	guild, err := h.loadGuild(ctx, cmd.ID())
	if err != nil {
		return nil, err
	}

	// Invite member
	if err := guild.InviteMember(cmd.UserID(), cmd.Username, cmd.InvitedBy); err != nil {
		return nil, fmt.Errorf("failed to invite member: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"user_id":    cmd.UserID(),
			"username":   cmd.Username,
			"invited_by": cmd.InvitedBy,
			"message":    "Member invited successfully",
		},
	}, nil
}

// handleAcceptInvitation handles the AcceptInvitationCommand
func (h *GuildCommandHandler) handleAcceptInvitation(ctx context.Context, cmd *commands.AcceptInvitationCommand) (*cqrs.CommandResult, error) {
	// Load guild aggregate
	guild, err := h.loadGuild(ctx, cmd.ID())
	if err != nil {
		return nil, err
	}

	// Accept invitation
	if err := guild.AcceptInvitation(cmd.UserID()); err != nil {
		return nil, fmt.Errorf("failed to accept invitation: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"user_id": cmd.UserID(),
			"message": "Invitation accepted successfully",
		},
	}, nil
}

// handleKickMember handles the KickMemberCommand
func (h *GuildCommandHandler) handleKickMember(ctx context.Context, cmd *commands.KickMemberCommand) (*cqrs.CommandResult, error) {
	// Load guild aggregate
	guild, err := h.loadGuild(ctx, cmd.ID())
	if err != nil {
		return nil, err
	}

	// Kick member
	if err := guild.KickMember(cmd.UserID(), cmd.KickedBy, cmd.Reason); err != nil {
		return nil, fmt.Errorf("failed to kick member: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"user_id":   cmd.UserID(),
			"kicked_by": cmd.KickedBy,
			"reason":    cmd.Reason,
			"message":   "Member kicked successfully",
		},
	}, nil
}

// handlePromoteMember handles the PromoteMemberCommand
func (h *GuildCommandHandler) handlePromoteMember(ctx context.Context, cmd *commands.PromoteMemberCommand) (*cqrs.CommandResult, error) {
	// Load guild aggregate
	guild, err := h.loadGuild(ctx, cmd.ID())
	if err != nil {
		return nil, err
	}

	// Parse new role
	newRole, err := domain.ParseGuildRole(cmd.NewRole)
	if err != nil {
		return nil, fmt.Errorf("invalid role: %w", err)
	}

	// Promote member
	if err := guild.PromoteMember(cmd.UserID(), cmd.PromotedBy, newRole); err != nil {
		return nil, fmt.Errorf("failed to promote member: %w", err)
	}

	// Save the guild
	if err := h.repository.Save(ctx, guild, guild.OriginalVersion()); err != nil {
		return nil, fmt.Errorf("failed to save guild: %w", err)
	}

	return &cqrs.CommandResult{
		AggregateID: cmd.ID(),
		Success:     true,
		Data: map[string]interface{}{
			"user_id":     cmd.UserID(),
			"new_role":    cmd.NewRole,
			"promoted_by": cmd.PromotedBy,
			"message":     "Member promoted successfully",
		},
	}, nil
}

// loadGuild loads a guild aggregate from the repository
func (h *GuildCommandHandler) loadGuild(ctx context.Context, guildID string) (*domain.GuildAggregate, error) {
	// Check if guild exists
	exists := h.repository.Exists(ctx, guildID)
	if !exists {
		return nil, fmt.Errorf("guild with ID %s not found", guildID)
	}

	// Load events
	events, err := h.repository.GetEventHistory(ctx, guildID, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to load guild events: %w", err)
	}

	// Reconstruct guild from events
	guild, err := domain.LoadGuildAggregate(guildID, events)
	if err != nil {
		return nil, fmt.Errorf("failed to load guild aggregate: %w", err)
	}

	return guild, nil
}
