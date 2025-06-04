package domain

import (
	"fmt"
	"strings"
)

// GuildRole represents the role of a guild member
type GuildRole int

const (
	// RoleGuest represents a guest or applicant
	RoleGuest GuildRole = iota
	// RoleMember represents a regular guild member
	RoleMember
	// RoleOfficer represents an officer with additional permissions
	RoleOfficer
	// RoleViceLeader represents a vice leader with management permissions
	RoleViceLeader
	// RoleLeader represents the guild leader with full permissions
	RoleLeader
)

// String returns the string representation of the guild role
func (r GuildRole) String() string {
	switch r {
	case RoleGuest:
		return "Guest"
	case RoleMember:
		return "Member"
	case RoleOfficer:
		return "Officer"
	case RoleViceLeader:
		return "ViceLeader"
	case RoleLeader:
		return "Leader"
	default:
		return "Unknown"
	}
}

// ParseGuildRole parses a string into a GuildRole
func ParseGuildRole(s string) (GuildRole, error) {
	switch strings.ToLower(s) {
	case "guest":
		return RoleGuest, nil
	case "member":
		return RoleMember, nil
	case "officer":
		return RoleOfficer, nil
	case "viceleader":
		return RoleViceLeader, nil
	case "leader":
		return RoleLeader, nil
	default:
		return RoleGuest, fmt.Errorf("invalid guild role: %s", s)
	}
}

// Permission represents a specific permission within the guild
type Permission int

const (
	// PermissionViewGuild allows viewing guild information
	PermissionViewGuild Permission = iota
	// PermissionInviteMembers allows inviting new members
	PermissionInviteMembers
	// PermissionKickMembers allows removing members
	PermissionKickMembers
	// PermissionPromoteMembers allows promoting members
	PermissionPromoteMembers
	// PermissionDemoteMembers allows demoting members
	PermissionDemoteMembers
	// PermissionManageGuild allows managing guild settings
	PermissionManageGuild
	// PermissionStartMining allows starting mining operations
	PermissionStartMining
	// PermissionManageMining allows managing mining operations
	PermissionManageMining
	// PermissionStartTransport allows starting transport operations
	PermissionStartTransport
	// PermissionManageTransport allows managing transport operations
	PermissionManageTransport
	// PermissionAttackTransport allows attacking other guild transports
	PermissionAttackTransport
	// PermissionChat allows chatting in guild channels
	PermissionChat
	// PermissionModerateChat allows moderating guild chat
	PermissionModerateChat
	// PermissionViewTreasury allows viewing guild treasury
	PermissionViewTreasury
	// PermissionManageTreasury allows managing guild treasury
	PermissionManageTreasury
)

// String returns the string representation of the permission
func (p Permission) String() string {
	switch p {
	case PermissionViewGuild:
		return "ViewGuild"
	case PermissionInviteMembers:
		return "InviteMembers"
	case PermissionKickMembers:
		return "KickMembers"
	case PermissionPromoteMembers:
		return "PromoteMembers"
	case PermissionDemoteMembers:
		return "DemoteMembers"
	case PermissionManageGuild:
		return "ManageGuild"
	case PermissionStartMining:
		return "StartMining"
	case PermissionManageMining:
		return "ManageMining"
	case PermissionStartTransport:
		return "StartTransport"
	case PermissionManageTransport:
		return "ManageTransport"
	case PermissionAttackTransport:
		return "AttackTransport"
	case PermissionChat:
		return "Chat"
	case PermissionModerateChat:
		return "ModerateChat"
	case PermissionViewTreasury:
		return "ViewTreasury"
	case PermissionManageTreasury:
		return "ManageTreasury"
	default:
		return "Unknown"
	}
}

// RolePermissions defines the permissions for each role
var RolePermissions = map[GuildRole][]Permission{
	RoleGuest: {
		PermissionViewGuild,
	},
	RoleMember: {
		PermissionViewGuild,
		PermissionStartMining,
		PermissionStartTransport,
		PermissionChat,
		PermissionViewTreasury,
	},
	RoleOfficer: {
		PermissionViewGuild,
		PermissionInviteMembers,
		PermissionStartMining,
		PermissionManageMining,
		PermissionStartTransport,
		PermissionManageTransport,
		PermissionAttackTransport,
		PermissionChat,
		PermissionModerateChat,
		PermissionViewTreasury,
	},
	RoleViceLeader: {
		PermissionViewGuild,
		PermissionInviteMembers,
		PermissionKickMembers,
		PermissionPromoteMembers,
		PermissionDemoteMembers,
		PermissionStartMining,
		PermissionManageMining,
		PermissionStartTransport,
		PermissionManageTransport,
		PermissionAttackTransport,
		PermissionChat,
		PermissionModerateChat,
		PermissionViewTreasury,
		PermissionManageTreasury,
	},
	RoleLeader: {
		PermissionViewGuild,
		PermissionInviteMembers,
		PermissionKickMembers,
		PermissionPromoteMembers,
		PermissionDemoteMembers,
		PermissionManageGuild,
		PermissionStartMining,
		PermissionManageMining,
		PermissionStartTransport,
		PermissionManageTransport,
		PermissionAttackTransport,
		PermissionChat,
		PermissionModerateChat,
		PermissionViewTreasury,
		PermissionManageTreasury,
	},
}

// HasPermission checks if a role has a specific permission
func (r GuildRole) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetPermissions returns all permissions for a role
func (r GuildRole) GetPermissions() []Permission {
	permissions, exists := RolePermissions[r]
	if !exists {
		return []Permission{}
	}
	return permissions
}

// CanPromoteTo checks if a role can promote another member to a target role
func (r GuildRole) CanPromoteTo(targetRole GuildRole) bool {
	// Only leaders and vice leaders can promote
	if r != RoleLeader && r != RoleViceLeader {
		return false
	}

	// Leaders can promote to any role except leader
	if r == RoleLeader {
		return targetRole != RoleLeader
	}

	// Vice leaders can promote up to officer
	if r == RoleViceLeader {
		return targetRole == RoleMember || targetRole == RoleOfficer
	}

	return false
}

// CanDemoteFrom checks if a role can demote another member from a source role
func (r GuildRole) CanDemoteFrom(sourceRole GuildRole) bool {
	// Only leaders and vice leaders can demote
	if r != RoleLeader && r != RoleViceLeader {
		return false
	}

	// Leaders can demote anyone except other leaders
	if r == RoleLeader {
		return sourceRole != RoleLeader
	}

	// Vice leaders can demote up to officer
	if r == RoleViceLeader {
		return sourceRole == RoleOfficer || sourceRole == RoleMember
	}

	return false
}

// CanKick checks if a role can kick another member with a specific role
func (r GuildRole) CanKick(targetRole GuildRole) bool {
	// Must have kick permission
	if !r.HasPermission(PermissionKickMembers) {
		return false
	}

	// Cannot kick someone with equal or higher role
	return r > targetRole
}
