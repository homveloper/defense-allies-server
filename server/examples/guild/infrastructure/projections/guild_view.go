package projections

import (
	"context"
	"fmt"
	"strings"
	"time"

	"defense-allies-server/examples/guild/domain"
	"defense-allies-server/pkg/cqrs"
)

// GuildView represents a read model for guild data
type GuildView struct {
	*cqrs.BaseReadModel
	GuildID     string    `json:"guild_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Notice      string    `json:"notice"`
	Tag         string    `json:"tag"`
	Status      string    `json:"status"`
	FoundedAt   time.Time `json:"founded_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Guild settings
	MaxMembers      int  `json:"max_members"`
	MinLevel        int  `json:"min_level"`
	IsPublic        bool `json:"is_public"`
	RequireApproval bool `json:"require_approval"`

	// Guild statistics
	MemberCount       int   `json:"member_count"`
	ActiveMemberCount int   `json:"active_member_count"`
	Treasury          int64 `json:"treasury"`
	Level             int   `json:"level"`
	Experience        int64 `json:"experience"`
	TotalContribution int64 `json:"total_contribution"`

	// Founder information
	FounderID       string `json:"founder_id"`
	FounderUsername string `json:"founder_username"`

	// Searchable text for full-text search
	SearchableText string `json:"searchable_text"`
}

// NewGuildView creates a new GuildView
func NewGuildView(guildID string) *GuildView {
	guildView := &GuildView{
		BaseReadModel: cqrs.NewBaseReadModel(guildID, "GuildView", map[string]interface{}{}),
		GuildID:       guildID,
		FoundedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		MaxMembers:    50,
		MinLevel:      1,
		IsPublic:      true,
		Level:         1,
	}
	return guildView
}

// GetData returns the GuildView data as a map for serialization
func (gv *GuildView) GetData() interface{} {
	return map[string]interface{}{
		"guild_id":            gv.GuildID,
		"name":                gv.Name,
		"description":         gv.Description,
		"notice":              gv.Notice,
		"tag":                 gv.Tag,
		"status":              gv.Status,
		"founded_at":          gv.FoundedAt,
		"updated_at":          gv.UpdatedAt,
		"max_members":         gv.MaxMembers,
		"min_level":           gv.MinLevel,
		"is_public":           gv.IsPublic,
		"require_approval":    gv.RequireApproval,
		"member_count":        gv.MemberCount,
		"active_member_count": gv.ActiveMemberCount,
		"treasury":            gv.Treasury,
		"level":               gv.Level,
		"experience":          gv.Experience,
		"total_contribution":  gv.TotalContribution,
		"founder_id":          gv.FounderID,
		"founder_username":    gv.FounderUsername,
		"searchable_text":     gv.SearchableText,
	}
}

// UpdateSearchableText updates the searchable text field
func (gv *GuildView) UpdateSearchableText() {
	var parts []string

	if gv.Name != "" {
		parts = append(parts, gv.Name)
	}
	if gv.Tag != "" {
		parts = append(parts, gv.Tag)
	}
	if gv.Description != "" {
		parts = append(parts, gv.Description)
	}
	if gv.Notice != "" {
		parts = append(parts, gv.Notice)
	}
	if gv.FounderUsername != "" {
		parts = append(parts, gv.FounderUsername)
	}

	gv.SearchableText = strings.Join(parts, " ")
}

// GetDisplayName returns a display-friendly name
func (gv *GuildView) GetDisplayName() string {
	if gv.Tag != "" {
		return fmt.Sprintf("[%s] %s", gv.Tag, gv.Name)
	}
	return gv.Name
}

// IsActive returns true if the guild is active
func (gv *GuildView) IsActive() bool {
	return gv.Status == "Active"
}

// GetMemberCapacityPercentage returns the member capacity as a percentage
func (gv *GuildView) GetMemberCapacityPercentage() float64 {
	if gv.MaxMembers == 0 {
		return 0
	}
	return float64(gv.MemberCount) / float64(gv.MaxMembers) * 100
}

// GuildViewProjection handles guild events and updates the GuildView read model
type GuildViewProjection struct {
	*cqrs.BaseProjection
	readStore cqrs.ReadStore
}

// NewGuildViewProjection creates a new GuildViewProjection
func NewGuildViewProjection(readStore cqrs.ReadStore) *GuildViewProjection {
	supportedEvents := []string{
		domain.GuildCreatedEventType,
		domain.GuildInfoUpdatedEventType,
		domain.GuildSettingsUpdatedEventType,
		domain.MemberInvitedEventType,
		domain.MemberJoinedEventType,
		domain.MemberKickedEventType,
		domain.MemberPromotedEventType,
	}

	return &GuildViewProjection{
		BaseProjection: cqrs.NewBaseProjection("GuildViewProjection", "1.0.0", supportedEvents),
		readStore:      readStore,
	}
}

// Project processes the event and updates the read model
func (p *GuildViewProjection) Project(ctx context.Context, event cqrs.EventMessage) error {
	// Call base implementation first
	if err := p.BaseProjection.Project(ctx, event); err != nil {
		return err
	}

	switch e := event.(type) {
	case *domain.GuildCreatedEvent:
		return p.handleGuildCreated(ctx, e)
	case *domain.GuildInfoUpdatedEvent:
		return p.handleGuildInfoUpdated(ctx, e)
	case *domain.GuildSettingsUpdatedEvent:
		return p.handleGuildSettingsUpdated(ctx, e)
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

// handleGuildCreated handles GuildCreatedEvent
func (p *GuildViewProjection) handleGuildCreated(ctx context.Context, event *domain.GuildCreatedEvent) error {
	guildView := NewGuildView(event.AggregateID())
	guildView.Name = event.Name
	guildView.Description = event.Description
	guildView.Status = "Active"
	guildView.FoundedAt = event.Timestamp()
	guildView.UpdatedAt = event.Timestamp()
	guildView.FounderID = event.FounderID
	guildView.FounderUsername = event.FounderUsername
	guildView.MemberCount = 1 // Founder is the first member
	guildView.ActiveMemberCount = 1
	guildView.SetVersion(event.Version())

	guildView.UpdateSearchableText()

	return p.readStore.Save(ctx, guildView)
}

// handleGuildInfoUpdated handles GuildInfoUpdatedEvent
func (p *GuildViewProjection) handleGuildInfoUpdated(ctx context.Context, event *domain.GuildInfoUpdatedEvent) error {
	// Load existing guild view
	readModel, err := p.readStore.GetByID(ctx, event.AggregateID(), "GuildView")
	if err != nil {
		return fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*GuildView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	// Update guild info
	guildView.Name = event.Name
	guildView.Description = event.Description
	guildView.Notice = event.Notice
	guildView.Tag = event.Tag
	guildView.UpdatedAt = event.Timestamp()
	guildView.SetVersion(event.Version())

	guildView.UpdateSearchableText()

	return p.readStore.Save(ctx, guildView)
}

// handleGuildSettingsUpdated handles GuildSettingsUpdatedEvent
func (p *GuildViewProjection) handleGuildSettingsUpdated(ctx context.Context, event *domain.GuildSettingsUpdatedEvent) error {
	// Load existing guild view
	readModel, err := p.readStore.GetByID(ctx, event.AggregateID(), "GuildView")
	if err != nil {
		return fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*GuildView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	// Update guild settings
	guildView.MaxMembers = event.MaxMembers
	guildView.MinLevel = event.MinLevel
	guildView.IsPublic = event.IsPublic
	guildView.RequireApproval = event.RequireApproval
	guildView.UpdatedAt = event.Timestamp()
	guildView.SetVersion(event.Version())

	return p.readStore.Save(ctx, guildView)
}

// handleMemberInvited handles MemberInvitedEvent
func (p *GuildViewProjection) handleMemberInvited(ctx context.Context, event *domain.MemberInvitedEvent) error {
	// Load existing guild view
	readModel, err := p.readStore.GetByID(ctx, event.AggregateID(), "GuildView")
	if err != nil {
		return fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*GuildView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	// Update member count (invited member is pending)
	guildView.MemberCount++
	guildView.UpdatedAt = event.Timestamp()
	guildView.SetVersion(event.Version())

	return p.readStore.Save(ctx, guildView)
}

// handleMemberJoined handles MemberJoinedEvent
func (p *GuildViewProjection) handleMemberJoined(ctx context.Context, event *domain.MemberJoinedEvent) error {
	// Load existing guild view
	readModel, err := p.readStore.GetByID(ctx, event.AggregateID(), "GuildView")
	if err != nil {
		return fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*GuildView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	// Update active member count (member accepted invitation)
	guildView.ActiveMemberCount++
	guildView.UpdatedAt = event.Timestamp()
	guildView.SetVersion(event.Version())

	return p.readStore.Save(ctx, guildView)
}

// handleMemberKicked handles MemberKickedEvent
func (p *GuildViewProjection) handleMemberKicked(ctx context.Context, event *domain.MemberKickedEvent) error {
	// Load existing guild view
	readModel, err := p.readStore.GetByID(ctx, event.AggregateID(), "GuildView")
	if err != nil {
		return fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*GuildView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	// Update member counts (member was kicked)
	guildView.MemberCount--
	guildView.ActiveMemberCount--
	guildView.UpdatedAt = event.Timestamp()
	guildView.SetVersion(event.Version())

	return p.readStore.Save(ctx, guildView)
}

// handleMemberPromoted handles MemberPromotedEvent
func (p *GuildViewProjection) handleMemberPromoted(ctx context.Context, event *domain.MemberPromotedEvent) error {
	// Load existing guild view
	readModel, err := p.readStore.GetByID(ctx, event.AggregateID(), "GuildView")
	if err != nil {
		return fmt.Errorf("failed to load guild view: %w", err)
	}

	guildView, ok := readModel.(*GuildView)
	if !ok {
		return fmt.Errorf("invalid read model type: expected *GuildView, got %T", readModel)
	}

	// Update timestamp (promotion doesn't change counts but is an activity)
	guildView.UpdatedAt = event.Timestamp()
	guildView.SetVersion(event.Version())

	return p.readStore.Save(ctx, guildView)
}
