package domain

import (
	"fmt"
	"time"
)

// MemberStatus represents the status of a guild member
type MemberStatus int

const (
	// StatusPending represents a member with pending invitation
	StatusPending MemberStatus = iota
	// StatusActive represents an active guild member
	StatusActive
	// StatusInactive represents an inactive guild member
	StatusInactive
	// StatusLeft represents a member who left the guild
	StatusLeft
	// StatusKicked represents a member who was kicked from the guild
	StatusKicked
)

// String returns the string representation of the member status
func (s MemberStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusActive:
		return "Active"
	case StatusInactive:
		return "Inactive"
	case StatusLeft:
		return "Left"
	case StatusKicked:
		return "Kicked"
	default:
		return "Unknown"
	}
}

// GuildMember represents a member of a guild
type GuildMember struct {
	UserID       string       `json:"user_id"`
	Username     string       `json:"username"`
	Role         GuildRole    `json:"role"`
	Status       MemberStatus `json:"status"`
	JoinedAt     time.Time    `json:"joined_at"`
	LastActiveAt time.Time    `json:"last_active_at"`
	InvitedBy    string       `json:"invited_by,omitempty"`
	KickedBy     string       `json:"kicked_by,omitempty"`
	KickedReason string       `json:"kicked_reason,omitempty"`
	Contribution int64        `json:"contribution"` // Total contribution points
}

// NewGuildMember creates a new guild member
func NewGuildMember(userID, username, invitedBy string) *GuildMember {
	now := time.Now()
	return &GuildMember{
		UserID:       userID,
		Username:     username,
		Role:         RoleMember,
		Status:       StatusPending,
		JoinedAt:     now,
		LastActiveAt: now,
		InvitedBy:    invitedBy,
		Contribution: 0,
	}
}

// Activate activates the member (accepts invitation)
func (m *GuildMember) Activate() error {
	if m.Status != StatusPending {
		return fmt.Errorf("member must be pending to activate, current status: %s", m.Status.String())
	}
	m.Status = StatusActive
	m.LastActiveAt = time.Now()
	return nil
}

// Deactivate deactivates the member
func (m *GuildMember) Deactivate() error {
	if m.Status != StatusActive {
		return fmt.Errorf("member must be active to deactivate, current status: %s", m.Status.String())
	}
	m.Status = StatusInactive
	return nil
}

// Leave marks the member as left
func (m *GuildMember) Leave() error {
	if m.Status != StatusActive && m.Status != StatusInactive {
		return fmt.Errorf("member must be active or inactive to leave, current status: %s", m.Status.String())
	}
	m.Status = StatusLeft
	return nil
}

// Kick marks the member as kicked
func (m *GuildMember) Kick(kickedBy, reason string) error {
	if m.Status != StatusActive && m.Status != StatusInactive {
		return fmt.Errorf("member must be active or inactive to be kicked, current status: %s", m.Status.String())
	}
	m.Status = StatusKicked
	m.KickedBy = kickedBy
	m.KickedReason = reason
	return nil
}

// Promote promotes the member to a higher role
func (m *GuildMember) Promote(newRole GuildRole) error {
	if m.Status != StatusActive {
		return fmt.Errorf("only active members can be promoted")
	}
	if newRole <= m.Role {
		return fmt.Errorf("new role must be higher than current role")
	}
	m.Role = newRole
	return nil
}

// Demote demotes the member to a lower role
func (m *GuildMember) Demote(newRole GuildRole) error {
	if m.Status != StatusActive {
		return fmt.Errorf("only active members can be demoted")
	}
	if newRole >= m.Role {
		return fmt.Errorf("new role must be lower than current role")
	}
	m.Role = newRole
	return nil
}

// AddContribution adds contribution points to the member
func (m *GuildMember) AddContribution(points int64) {
	if points > 0 {
		m.Contribution += points
		m.LastActiveAt = time.Now()
	}
}

// UpdateLastActive updates the last active timestamp
func (m *GuildMember) UpdateLastActive() {
	m.LastActiveAt = time.Now()
}

// IsActive returns true if the member is active
func (m *GuildMember) IsActive() bool {
	return m.Status == StatusActive
}

// HasPermission checks if the member has a specific permission
func (m *GuildMember) HasPermission(permission Permission) bool {
	if !m.IsActive() {
		return false
	}
	return m.Role.HasPermission(permission)
}

// CanPromote checks if this member can promote another member to a target role
func (m *GuildMember) CanPromote(targetRole GuildRole) bool {
	if !m.IsActive() {
		return false
	}
	return m.Role.CanPromoteTo(targetRole)
}

// CanDemote checks if this member can demote another member from a source role
func (m *GuildMember) CanDemote(sourceRole GuildRole) bool {
	if !m.IsActive() {
		return false
	}
	return m.Role.CanDemoteFrom(sourceRole)
}

// CanKick checks if this member can kick another member with a specific role
func (m *GuildMember) CanKick(targetRole GuildRole) bool {
	if !m.IsActive() {
		return false
	}
	return m.Role.CanKick(targetRole)
}

// GetDaysInGuild returns the number of days the member has been in the guild
func (m *GuildMember) GetDaysInGuild() int {
	return int(time.Since(m.JoinedAt).Hours() / 24)
}

// GetDaysSinceLastActive returns the number of days since the member was last active
func (m *GuildMember) GetDaysSinceLastActive() int {
	return int(time.Since(m.LastActiveAt).Hours() / 24)
}

// Clone creates a deep copy of the guild member
func (m *GuildMember) Clone() *GuildMember {
	return &GuildMember{
		UserID:       m.UserID,
		Username:     m.Username,
		Role:         m.Role,
		Status:       m.Status,
		JoinedAt:     m.JoinedAt,
		LastActiveAt: m.LastActiveAt,
		InvitedBy:    m.InvitedBy,
		KickedBy:     m.KickedBy,
		KickedReason: m.KickedReason,
		Contribution: m.Contribution,
	}
}

// Validate validates the guild member data
func (m *GuildMember) Validate() error {
	if m.UserID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	if m.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if m.JoinedAt.IsZero() {
		return fmt.Errorf("joined at cannot be zero")
	}
	if m.LastActiveAt.IsZero() {
		return fmt.Errorf("last active at cannot be zero")
	}
	if m.Contribution < 0 {
		return fmt.Errorf("contribution cannot be negative")
	}
	return nil
}
