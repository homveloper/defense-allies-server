package domain

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs"

	"github.com/pkg/errors"
)

// UserStatus represents the status of a user
type UserStatus int

const (
	UserStatusActive UserStatus = iota
	UserStatusInactive
	UserStatusDeactivated
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusActive:
		return "active"
	case UserStatusInactive:
		return "inactive"
	case UserStatusDeactivated:
		return "deactivated"
	default:
		return "unknown"
	}
}

// User represents a user aggregate
type User struct {
	*cqrs.BaseAggregate
	email              string
	name               string
	status             UserStatus
	lastLoginAt        *time.Time
	deactivatedAt      *time.Time
	deactivationReason string

	// Role management
	roleManager *RoleManager

	// Profile information
	profile *UserProfile
}

// NewUser creates a new User aggregate
func NewUser(userID, email, name string) (*User, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	user := &User{
		BaseAggregate: cqrs.NewBaseAggregate(userID, "User"),
		email:         email,
		name:          name,
		status:        UserStatusActive,
		roleManager:   NewRoleManager(),
		profile:       NewUserProfile("", name), // Initialize with name as display name
	}

	// Assign default user role
	defaultRole := NewRole(RoleTypeUser, "system")
	user.roleManager.AddRole(defaultRole)

	// Apply the creation event
	event := NewUserCreatedEvent(userID, email, name)
	user.Apply(event, true)

	return user, nil
}

// LoadUserFromHistory creates a User aggregate from event history
func LoadUserFromHistory(userID string, events []cqrs.EventMessage) (*User, error) {
	user := &User{
		BaseAggregate: cqrs.NewBaseAggregate(userID, "User"),
		roleManager:   NewRoleManager(),
		profile:       NewUserProfile("", ""), // Will be populated from events
	}

	for _, event := range events {
		if err := user.applyEvent(event); err != nil {
			return nil, errors.Wrapf(err, "failed to apply event %s", event.EventType())
		}
		user.SetOriginalVersion(event.Version())
	}

	user.ClearChanges()
	return user, nil
}

// Business methods

// ChangeEmail changes the user's email address
func (u *User) ChangeEmail(newEmail string) error {
	if u.IsDeleted() {
		return errors.New("cannot change email of deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot change email of deactivated user")
	}

	if newEmail == u.email {
		return errors.New("new email is the same as current email")
	}

	if !isValidEmail(newEmail) {
		return errors.Errorf("invalid email format: %s", newEmail)
	}

	event := NewEmailChangedEvent(u.ID(), u.email, newEmail, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// Deactivate deactivates the user
func (u *User) Deactivate(reason string) error {
	if u.IsDeleted() {
		return errors.New("cannot deactivate deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("user is already deactivated")
	}

	if reason == "" {
		return errors.New("deactivation reason cannot be empty")
	}

	event := NewUserDeactivatedEvent(u.ID(), reason, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// Activate activates the user
func (u *User) Activate() error {
	if u.IsDeleted() {
		return errors.New("cannot activate deleted user")
	}

	if u.status == UserStatusActive {
		return errors.New("user is already active")
	}

	event := NewUserActivatedEvent(u.ID(), u.Version()+1)
	u.Apply(event, true)

	return nil
}

// RecordLogin records a user login
func (u *User) RecordLogin() error {
	if u.IsDeleted() {
		return errors.New("cannot record login for deleted user")
	}

	if u.status != UserStatusActive {
		return errors.New("cannot record login for inactive user")
	}

	now := time.Now()
	u.lastLoginAt = &now
	// Update the aggregate's updated time
	u.IncrementVersion()

	return nil
}

// Getters

// Email returns the user's email
func (u *User) Email() string {
	return u.email
}

// Name returns the user's name
func (u *User) Name() string {
	return u.name
}

// Status returns the user's status
func (u *User) Status() UserStatus {
	return u.status
}

// LastLoginAt returns the user's last login time
func (u *User) LastLoginAt() *time.Time {
	return u.lastLoginAt
}

// DeactivatedAt returns when the user was deactivated
func (u *User) DeactivatedAt() *time.Time {
	return u.deactivatedAt
}

// DeactivationReason returns the reason for deactivation
func (u *User) DeactivationReason() string {
	return u.deactivationReason
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.status == UserStatusActive && !u.IsDeleted()
}

// Role management methods

// AssignRole assigns a role to the user
func (u *User) AssignRole(roleType RoleType, assignedBy string) error {
	if u.IsDeleted() {
		return errors.New("cannot assign role to deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot assign role to deactivated user")
	}

	role := NewRole(roleType, assignedBy)
	u.roleManager.AddRole(role)

	event := NewRoleAssignedEvent(u.ID(), roleType, assignedBy, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// AssignRoleWithExpiry assigns a role with expiration to the user
func (u *User) AssignRoleWithExpiry(roleType RoleType, assignedBy string, expiresAt time.Time) error {
	if u.IsDeleted() {
		return errors.New("cannot assign role to deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot assign role to deactivated user")
	}

	if expiresAt.Before(time.Now()) {
		return errors.New("expiration time cannot be in the past")
	}

	role := NewRoleWithExpiry(roleType, assignedBy, expiresAt)
	u.roleManager.AddRole(role)

	event := NewRoleAssignedWithExpiryEvent(u.ID(), roleType, assignedBy, expiresAt, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// RevokeRole revokes a role from the user
func (u *User) RevokeRole(roleType RoleType, revokedBy string) error {
	if u.IsDeleted() {
		return errors.New("cannot revoke role from deleted user")
	}

	if !u.roleManager.HasRole(roleType) {
		return errors.Errorf("user does not have role: %s", roleType.String())
	}

	// Cannot revoke the last user role
	if roleType == RoleTypeUser && len(u.roleManager.GetActiveRoles()) == 1 {
		return errors.New("cannot revoke the last user role")
	}

	u.roleManager.RemoveRole(roleType)

	event := NewRoleRevokedEvent(u.ID(), roleType, revokedBy, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleType RoleType) bool {
	return u.roleManager.HasRole(roleType)
}

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(permission string) bool {
	return u.roleManager.HasPermission(permission)
}

// GetRoles returns all active roles
func (u *User) GetRoles() []*Role {
	return u.roleManager.GetActiveRoles()
}

// GetPermissions returns all permissions from active roles
func (u *User) GetPermissions() []string {
	return u.roleManager.GetAllPermissions()
}

// Profile management methods

// UpdateProfile updates the user's profile information
func (u *User) UpdateProfile(firstName, lastName, bio string) error {
	if u.IsDeleted() {
		return errors.New("cannot update profile of deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot update profile of deactivated user")
	}

	u.profile.UpdateBasicInfo(firstName, lastName, bio)

	// Create changes map for event
	changes := map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"bio":        bio,
	}

	event := NewProfileUpdatedEvent(u.ID(), changes, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// UpdateDisplayName updates the user's display name
func (u *User) UpdateDisplayName(displayName string) error {
	if u.IsDeleted() {
		return errors.New("cannot update display name of deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot update display name of deactivated user")
	}

	if err := u.profile.UpdateDisplayName(displayName); err != nil {
		return err
	}

	changes := map[string]interface{}{
		"display_name": displayName,
	}

	event := NewProfileUpdatedEvent(u.ID(), changes, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// UpdateContactInfo updates the user's contact information
func (u *User) UpdateContactInfo(phoneNumber, address, city, country, postalCode string) error {
	if u.IsDeleted() {
		return errors.New("cannot update contact info of deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot update contact info of deactivated user")
	}

	u.profile.UpdateContactInfo(phoneNumber, address, city, country, postalCode)

	changes := map[string]interface{}{
		"phone_number": phoneNumber,
		"address":      address,
		"city":         city,
		"country":      country,
		"postal_code":  postalCode,
	}

	event := NewProfileUpdatedEvent(u.ID(), changes, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// SetAvatar sets the user's avatar
func (u *User) SetAvatar(avatarURL string) error {
	if u.IsDeleted() {
		return errors.New("cannot set avatar of deleted user")
	}

	if u.status == UserStatusDeactivated {
		return errors.New("cannot set avatar of deactivated user")
	}

	u.profile.SetAvatar(avatarURL)

	changes := map[string]interface{}{
		"avatar": avatarURL,
	}

	event := NewProfileUpdatedEvent(u.ID(), changes, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// SetPreference sets a user preference
func (u *User) SetPreference(key string, value interface{}) error {
	if u.IsDeleted() {
		return errors.New("cannot set preference of deleted user")
	}

	u.profile.SetPreference(key, value)

	changes := map[string]interface{}{
		"preferences": map[string]interface{}{
			key: value,
		},
	}

	event := NewProfileUpdatedEvent(u.ID(), changes, u.Version()+1)
	u.Apply(event, true)

	return nil
}

// GetProfile returns the user's profile
func (u *User) GetProfile() *UserProfile {
	return u.profile
}

// Apply applies an event to the aggregate
func (u *User) Apply(event cqrs.EventMessage, isNew bool) {
	u.BaseAggregate.Apply(event, isNew)
	if err := u.applyEvent(event); err != nil {
		// In a real implementation, you might want to handle this differently
		panic(fmt.Sprintf("failed to apply event: %v", err))
	}
}

// applyEvent applies the event to the aggregate state
func (u *User) applyEvent(event cqrs.EventMessage) error {
	switch e := event.(type) {
	case *UserCreatedEvent:
		u.email = e.Email
		u.name = e.Name
		u.status = UserStatusActive

	case *EmailChangedEvent:
		u.email = e.NewEmail

	case *UserDeactivatedEvent:
		u.status = UserStatusDeactivated
		u.deactivatedAt = &e.DeactivatedAt
		u.deactivationReason = e.Reason

	case *UserActivatedEvent:
		u.status = UserStatusActive
		u.deactivatedAt = nil
		u.deactivationReason = ""

	case *RoleAssignedEvent:
		role := NewRole(e.RoleType, e.AssignedBy)
		role.AssignedAt = e.AssignedAt
		u.roleManager.AddRole(role)

	case *RoleAssignedWithExpiryEvent:
		role := NewRoleWithExpiry(e.RoleType, e.AssignedBy, e.ExpiresAt)
		role.AssignedAt = e.AssignedAt
		u.roleManager.AddRole(role)

	case *RoleRevokedEvent:
		u.roleManager.RemoveRole(e.RoleType)

	case *ProfileUpdatedEvent:
		// Apply profile changes from event
		if u.profile == nil {
			u.profile = NewUserProfile("", "")
		}

		for key, value := range e.Changes {
			switch key {
			case "first_name":
				if v, ok := value.(string); ok {
					u.profile.FirstName = v
				}
			case "last_name":
				if v, ok := value.(string); ok {
					u.profile.LastName = v
				}
			case "bio":
				if v, ok := value.(string); ok {
					u.profile.Bio = v
				}
			case "display_name":
				if v, ok := value.(string); ok {
					u.profile.DisplayName = v
				}
			case "avatar":
				if v, ok := value.(string); ok {
					u.profile.Avatar = v
				}
			case "phone_number":
				if v, ok := value.(string); ok {
					u.profile.PhoneNumber = v
				}
			case "address":
				if v, ok := value.(string); ok {
					u.profile.Address = v
				}
			case "city":
				if v, ok := value.(string); ok {
					u.profile.City = v
				}
			case "country":
				if v, ok := value.(string); ok {
					u.profile.Country = v
				}
			case "postal_code":
				if v, ok := value.(string); ok {
					u.profile.PostalCode = v
				}
			case "preferences":
				if prefs, ok := value.(map[string]interface{}); ok {
					if u.profile.Preferences == nil {
						u.profile.Preferences = make(map[string]interface{})
					}
					for k, v := range prefs {
						u.profile.Preferences[k] = v
					}
				}
			}
		}
		u.profile.UpdatedAt = e.UpdatedAt

	default:
		return errors.Errorf("unknown event type: %T", event)
	}

	return nil
}

// Validate validates the user aggregate
func (u *User) Validate() error {
	if err := u.BaseAggregate.Validate(); err != nil {
		return errors.Wrap(err, "base aggregate validation failed")
	}

	if u.email == "" {
		return errors.New("user email cannot be empty")
	}

	if u.name == "" {
		return errors.New("user name cannot be empty")
	}

	if !isValidEmail(u.email) {
		return errors.Errorf("invalid email format: %s", u.email)
	}

	return nil
}
