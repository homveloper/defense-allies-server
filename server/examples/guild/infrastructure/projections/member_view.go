package projections

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/examples/guild/domain"
	"defense-allies-server/pkg/cqrs"
)

// MemberView represents a read model for guild member data
type MemberView struct {
	*cqrs.BaseReadModel
	GuildID      string    `json:"guild_id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Invitation information
	InvitedBy string `json:"invited_by,omitempty"`

	// Kick information
	KickedBy     string `json:"kicked_by,omitempty"`
	KickedReason string `json:"kicked_reason,omitempty"`

	// Statistics
	Contribution int64 `json:"contribution"`
	DaysInGuild  int   `json:"days_in_guild"`

	// Permissions (derived from role)
	Permissions []string `json:"permissions"`
}

// NewMemberView creates a new MemberView
func NewMemberView(guildID, userID string) *MemberView {
	now := time.Now()
	memberView := &MemberView{
		BaseReadModel: cqrs.NewBaseReadModel(fmt.Sprintf("%s:%s", guildID, userID), "MemberView", map[string]interface{}{}),
		GuildID:       guildID,
		UserID:        userID,
		JoinedAt:      now,
		LastActiveAt:  now,
		UpdatedAt:     now,
		Contribution:  0,
		Permissions:   make([]string, 0),
	}
	return memberView
}

// GetData returns the MemberView data as a map for serialization
func (mv *MemberView) GetData() interface{} {
	return map[string]interface{}{
		"guild_id":       mv.GuildID,
		"user_id":        mv.UserID,
		"username":       mv.Username,
		"role":           mv.Role,
		"status":         mv.Status,
		"joined_at":      mv.JoinedAt,
		"last_active_at": mv.LastActiveAt,
		"updated_at":     mv.UpdatedAt,
		"invited_by":     mv.InvitedBy,
		"kicked_by":      mv.KickedBy,
		"kicked_reason":  mv.KickedReason,
		"contribution":   mv.Contribution,
		"days_in_guild":  mv.DaysInGuild,
		"permissions":    mv.Permissions,
	}
}

// UpdateDaysInGuild calculates and updates the days in guild
func (mv *MemberView) UpdateDaysInGuild() {
	mv.DaysInGuild = int(time.Since(mv.JoinedAt).Hours() / 24)
}

// UpdatePermissions updates permissions based on role
func (mv *MemberView) UpdatePermissions() {
	role, err := domain.ParseGuildRole(mv.Role)
	if err != nil {
		mv.Permissions = []string{}
		return
	}

	permissions := role.GetPermissions()
	mv.Permissions = make([]string, len(permissions))
	for i, perm := range permissions {
		mv.Permissions[i] = perm.String()
	}
}

// IsActive returns true if the member is active
func (mv *MemberView) IsActive() bool {
	return mv.Status == "Active"
}

// HasPermission checks if the member has a specific permission
func (mv *MemberView) HasPermission(permission string) bool {
	for _, perm := range mv.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// MemberViewProjection handles guild member events and updates the MemberView read model
type MemberViewProjection struct {
	*cqrs.BaseProjection
	readStore cqrs.ReadStore
}

// NewMemberViewProjection creates a new MemberViewProjection
func NewMemberViewProjection(readStore cqrs.ReadStore) *MemberViewProjection {
	supportedEvents := []string{
		domain.GuildCreatedEventType,
		domain.MemberInvitedEventType,
		domain.MemberJoinedEventType,
		domain.MemberKickedEventType,
		domain.MemberPromotedEventType,
	}

	return &MemberViewProjection{
		BaseProjection: cqrs.NewBaseProjection("MemberViewProjection", "1.0.0", supportedEvents),
		readStore:      readStore,
	}
}

// Project processes the event and updates the read model
func (p *MemberViewProjection) Project(ctx context.Context, event cqrs.EventMessage) error {
	// Call base implementation first
	if err := p.BaseProjection.Project(ctx, event); err != nil {
		return err
	}

	switch e := event.(type) {
	case *domain.GuildCreatedEvent:
		return p.handleGuildCreated(ctx, e)
	case *domain.MemberInvitedEvent:
		return p.handleMemberInvited(ctx, e)
	case *domain.MemberJoinedEvent:
		return p.handleMemberJoined(ctx, e)
	case *domain.MemberKickedEvent:
		return p.handleMemberKicked(ctx, e)
	case *domain.MemberPromotedEvent:
		return p.handleMemberPromoted(ctx, e)
	default:
		return fmt.Errorf("unsupported event type: %T", event)
	}
}

// Event handlers

// handleGuildCreated handles GuildCreatedEvent (creates founder member)
func (p *MemberViewProjection) handleGuildCreated(ctx context.Context, event *domain.GuildCreatedEvent) error {
	guildID := event.ID()
	founderID := event.FounderID
	founderUsername := event.FounderUsername

	memberView := NewMemberView(guildID, founderID)
	memberView.Username = founderUsername
	memberView.Role = "Leader"
	memberView.Status = "Active"
	memberView.JoinedAt = event.Timestamp()
	memberView.LastActiveAt = event.Timestamp()
	memberView.UpdatedAt = event.Timestamp()
	memberView.SetVersion(event.Version())

	memberView.UpdatePermissions()
	memberView.UpdateDaysInGuild()

	return p.readStore.Save(ctx, memberView)
}

// handleMemberInvited handles MemberInvitedEvent
func (p *MemberViewProjection) handleMemberInvited(ctx context.Context, event *domain.MemberInvitedEvent) error {
	guildID := event.ID()
	userID := event.UserID
	username := event.Username
	invitedBy := event.InvitedBy

	memberView := NewMemberView(guildID, userID)
	memberView.Username = username
	memberView.Role = "Member"
	memberView.Status = "Pending"
	memberView.InvitedBy = invitedBy
	memberView.JoinedAt = event.Timestamp()
	memberView.LastActiveAt = event.Timestamp()
	memberView.UpdatedAt = event.Timestamp()
	memberView.SetVersion(event.Version())

	memberView.UpdatePermissions()
	memberView.UpdateDaysInGuild()

	return p.readStore.Save(ctx, memberView)
}

// handleMemberJoined handles MemberJoinedEvent
func (p *MemberViewProjection) handleMemberJoined(ctx context.Context, event *domain.MemberJoinedEvent) error {
	guildID := event.ID()
	userID := event.UserID
	memberID := fmt.Sprintf("%s:%s", guildID, userID)

	// Load existing member view
	readModel, err := p.readStore.GetByID(ctx, memberID, "MemberView")
	if err != nil {
		return fmt.Errorf("failed to load member view: %w", err)
	}

	memberView, ok := readModel.(*MemberView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *MemberView, got %T", readModel)
	}

	// Update status to active
	memberView.Status = "Active"
	memberView.LastActiveAt = event.Timestamp()
	memberView.UpdatedAt = event.Timestamp()
	memberView.SetVersion(event.Version())

	memberView.UpdatePermissions()
	memberView.UpdateDaysInGuild()

	return p.readStore.Save(ctx, memberView)
}

// handleMemberKicked handles MemberKickedEvent
func (p *MemberViewProjection) handleMemberKicked(ctx context.Context, event *domain.MemberKickedEvent) error {
	guildID := event.ID()
	userID := event.UserID
	kickedBy := event.KickedBy
	reason := event.Reason
	memberID := fmt.Sprintf("%s:%s", guildID, userID)

	// Load existing member view
	readModel, err := p.readStore.GetByID(ctx, memberID, "MemberView")
	if err != nil {
		return fmt.Errorf("failed to load member view: %w", err)
	}

	memberView, ok := readModel.(*MemberView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *MemberView, got %T", readModel)
	}

	// Update status to kicked
	memberView.Status = "Kicked"
	memberView.KickedBy = kickedBy
	memberView.KickedReason = reason
	memberView.UpdatedAt = event.Timestamp()
	memberView.SetVersion(event.Version())

	memberView.UpdatePermissions()
	memberView.UpdateDaysInGuild()

	return p.readStore.Save(ctx, memberView)
}

// handleMemberPromoted handles MemberPromotedEvent
func (p *MemberViewProjection) handleMemberPromoted(ctx context.Context, event *domain.MemberPromotedEvent) error {
	guildID := event.ID()
	userID := event.UserID
	newRole := event.NewRole.String()
	memberID := fmt.Sprintf("%s:%s", guildID, userID)

	// Load existing member view
	readModel, err := p.readStore.GetByID(ctx, memberID, "MemberView")
	if err != nil {
		return fmt.Errorf("failed to load member view: %w", err)
	}

	memberView, ok := readModel.(*MemberView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *MemberView, got %T", readModel)
	}

	// Update role
	memberView.Role = newRole
	memberView.LastActiveAt = event.Timestamp()
	memberView.UpdatedAt = event.Timestamp()
	memberView.SetVersion(event.Version())

	memberView.UpdatePermissions()
	memberView.UpdateDaysInGuild()

	return p.readStore.Save(ctx, memberView)
}
