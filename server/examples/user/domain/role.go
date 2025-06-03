package domain

import (
	"fmt"
	"strings"
	"time"
)

// RoleType represents different types of roles in the system
type RoleType int

const (
	// Basic user roles
	RoleTypeUser RoleType = iota
	RoleTypeModerator
	RoleTypeAdmin
	RoleTypeSuperAdmin

	// Game-specific roles
	RoleTypePlayer
	RoleTypeGameMaster
	RoleTypeBetaTester

	// Special roles
	RoleTypeGuest
	RoleTypeSupport
	RoleTypeDeveloper
)

// String returns the string representation of RoleType
func (rt RoleType) String() string {
	switch rt {
	case RoleTypeUser:
		return "user"
	case RoleTypeModerator:
		return "moderator"
	case RoleTypeAdmin:
		return "admin"
	case RoleTypeSuperAdmin:
		return "super_admin"
	case RoleTypePlayer:
		return "player"
	case RoleTypeGameMaster:
		return "game_master"
	case RoleTypeBetaTester:
		return "beta_tester"
	case RoleTypeGuest:
		return "guest"
	case RoleTypeSupport:
		return "support"
	case RoleTypeDeveloper:
		return "developer"
	default:
		return "unknown"
	}
}

// ParseRoleType parses a string to RoleType
func ParseRoleType(s string) (RoleType, error) {
	switch strings.ToLower(s) {
	case "user":
		return RoleTypeUser, nil
	case "moderator":
		return RoleTypeModerator, nil
	case "admin":
		return RoleTypeAdmin, nil
	case "super_admin":
		return RoleTypeSuperAdmin, nil
	case "player":
		return RoleTypePlayer, nil
	case "game_master":
		return RoleTypeGameMaster, nil
	case "beta_tester":
		return RoleTypeBetaTester, nil
	case "guest":
		return RoleTypeGuest, nil
	case "support":
		return RoleTypeSupport, nil
	case "developer":
		return RoleTypeDeveloper, nil
	default:
		return RoleTypeUser, fmt.Errorf("unknown role type: %s", s)
	}
}

// Role represents a user role with permissions and metadata
type Role struct {
	Type        RoleType  `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	AssignedAt  time.Time `json:"assigned_at"`
	AssignedBy  string    `json:"assigned_by,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool      `json:"is_active"`
}

// NewRole creates a new Role
func NewRole(roleType RoleType, assignedBy string) *Role {
	role := &Role{
		Type:        roleType,
		Name:        roleType.String(),
		Description: getDefaultRoleDescription(roleType),
		Permissions: getDefaultPermissions(roleType),
		AssignedAt:  time.Now(),
		AssignedBy:  assignedBy,
		IsActive:    true,
	}
	return role
}

// NewRoleWithExpiry creates a new Role with expiration
func NewRoleWithExpiry(roleType RoleType, assignedBy string, expiresAt time.Time) *Role {
	role := NewRole(roleType, assignedBy)
	role.ExpiresAt = &expiresAt
	return role
}

// IsExpired checks if the role has expired
func (r *Role) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// IsValid checks if the role is valid and active
func (r *Role) IsValid() bool {
	return r.IsActive && !r.IsExpired()
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permission string) bool {
	if !r.IsValid() {
		return false
	}
	
	for _, perm := range r.Permissions {
		if perm == permission || perm == "*" {
			return true
		}
	}
	return false
}

// Deactivate deactivates the role
func (r *Role) Deactivate() {
	r.IsActive = false
}

// Activate activates the role
func (r *Role) Activate() {
	r.IsActive = true
}

// getDefaultRoleDescription returns default description for role types
func getDefaultRoleDescription(roleType RoleType) string {
	switch roleType {
	case RoleTypeUser:
		return "Standard user with basic permissions"
	case RoleTypeModerator:
		return "Moderator with content management permissions"
	case RoleTypeAdmin:
		return "Administrator with system management permissions"
	case RoleTypeSuperAdmin:
		return "Super administrator with full system access"
	case RoleTypePlayer:
		return "Game player with gameplay permissions"
	case RoleTypeGameMaster:
		return "Game master with game management permissions"
	case RoleTypeBetaTester:
		return "Beta tester with testing permissions"
	case RoleTypeGuest:
		return "Guest user with limited permissions"
	case RoleTypeSupport:
		return "Support staff with customer service permissions"
	case RoleTypeDeveloper:
		return "Developer with development and debugging permissions"
	default:
		return "Unknown role"
	}
}

// getDefaultPermissions returns default permissions for role types
func getDefaultPermissions(roleType RoleType) []string {
	switch roleType {
	case RoleTypeUser:
		return []string{"user.read", "user.update_profile", "game.play"}
	case RoleTypeModerator:
		return []string{"user.read", "user.update_profile", "user.moderate", "content.moderate", "game.play"}
	case RoleTypeAdmin:
		return []string{"user.*", "content.*", "game.*", "system.read"}
	case RoleTypeSuperAdmin:
		return []string{"*"}
	case RoleTypePlayer:
		return []string{"user.read", "user.update_profile", "game.play", "game.stats"}
	case RoleTypeGameMaster:
		return []string{"user.read", "user.update_profile", "game.*", "event.manage"}
	case RoleTypeBetaTester:
		return []string{"user.read", "user.update_profile", "game.play", "game.test", "bug.report"}
	case RoleTypeGuest:
		return []string{"user.read"}
	case RoleTypeSupport:
		return []string{"user.read", "user.support", "ticket.manage"}
	case RoleTypeDeveloper:
		return []string{"user.*", "game.*", "system.*", "debug.*"}
	default:
		return []string{"user.read"}
	}
}

// RoleManager manages user roles
type RoleManager struct {
	roles map[RoleType]*Role
}

// NewRoleManager creates a new RoleManager
func NewRoleManager() *RoleManager {
	return &RoleManager{
		roles: make(map[RoleType]*Role),
	}
}

// AddRole adds a role to the manager
func (rm *RoleManager) AddRole(role *Role) {
	rm.roles[role.Type] = role
}

// RemoveRole removes a role from the manager
func (rm *RoleManager) RemoveRole(roleType RoleType) {
	delete(rm.roles, roleType)
}

// GetRole gets a role by type
func (rm *RoleManager) GetRole(roleType RoleType) (*Role, bool) {
	role, exists := rm.roles[roleType]
	return role, exists
}

// GetActiveRoles returns all active roles
func (rm *RoleManager) GetActiveRoles() []*Role {
	var activeRoles []*Role
	for _, role := range rm.roles {
		if role.IsValid() {
			activeRoles = append(activeRoles, role)
		}
	}
	return activeRoles
}

// HasRole checks if a specific role exists and is active
func (rm *RoleManager) HasRole(roleType RoleType) bool {
	role, exists := rm.roles[roleType]
	return exists && role.IsValid()
}

// HasPermission checks if any active role has the specified permission
func (rm *RoleManager) HasPermission(permission string) bool {
	for _, role := range rm.roles {
		if role.HasPermission(permission) {
			return true
		}
	}
	return false
}

// GetAllPermissions returns all permissions from all active roles
func (rm *RoleManager) GetAllPermissions() []string {
	permissionSet := make(map[string]bool)
	
	for _, role := range rm.roles {
		if role.IsValid() {
			for _, perm := range role.Permissions {
				permissionSet[perm] = true
			}
		}
	}
	
	var permissions []string
	for perm := range permissionSet {
		permissions = append(permissions, perm)
	}
	
	return permissions
}

// CleanupExpiredRoles removes expired roles
func (rm *RoleManager) CleanupExpiredRoles() {
	for roleType, role := range rm.roles {
		if role.IsExpired() {
			delete(rm.roles, roleType)
		}
	}
}
