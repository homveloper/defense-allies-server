package domain

import (
	"time"

	"cqrs"
)

// Event type constants
const (
	UserCreatedEventType            = "UserCreated"
	EmailChangedEventType           = "EmailChanged"
	UserDeactivatedEventType        = "UserDeactivated"
	UserActivatedEventType          = "UserActivated"
	RoleAssignedEventType           = "RoleAssigned"
	RoleAssignedWithExpiryEventType = "RoleAssignedWithExpiry"
	RoleRevokedEventType            = "RoleRevoked"
	ProfileUpdatedEventType         = "ProfileUpdated"
)

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUserCreatedEvent creates a new UserCreatedEvent
func NewUserCreatedEvent(userID, email, name string) *UserCreatedEvent {
	event := &UserCreatedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			UserCreatedEventType,
			userID,
			"User",
			1,
			map[string]interface{}{
				"user_id":    userID,
				"email":      email,
				"name":       name,
				"created_at": time.Now(),
			},
		),
		UserID:    userID,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now(),
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.Priority)
	return event
}

// EmailChangedEvent represents an email change event
type EmailChangedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID   string `json:"user_id"`
	OldEmail string `json:"old_email"`
	NewEmail string `json:"new_email"`
}

// NewEmailChangedEvent creates a new EmailChangedEvent
func NewEmailChangedEvent(userID, oldEmail, newEmail string, version int) *EmailChangedEvent {
	event := &EmailChangedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			EmailChangedEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":   userID,
				"old_email": oldEmail,
				"new_email": newEmail,
			},
		),
		UserID:   userID,
		OldEmail: oldEmail,
		NewEmail: newEmail,
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.Priority)
	return event
}

// UserDeactivatedEvent represents a user deactivation event
type UserDeactivatedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID        string    `json:"user_id"`
	DeactivatedAt time.Time `json:"deactivated_at"`
	Reason        string    `json:"reason"`
}

// NewUserDeactivatedEvent creates a new UserDeactivatedEvent
func NewUserDeactivatedEvent(userID, reason string, version int) *UserDeactivatedEvent {
	event := &UserDeactivatedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			UserDeactivatedEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":        userID,
				"deactivated_at": time.Now(),
				"reason":         reason,
			},
		),
		UserID:        userID,
		DeactivatedAt: time.Now(),
		Reason:        reason,
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.PriorityHigh)
	return event
}

// UserActivatedEvent represents a user activation event
type UserActivatedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID      string    `json:"user_id"`
	ActivatedAt time.Time `json:"activated_at"`
}

// NewUserActivatedEvent creates a new UserActivatedEvent
func NewUserActivatedEvent(userID string, version int) *UserActivatedEvent {
	event := &UserActivatedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			UserActivatedEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":      userID,
				"activated_at": time.Now(),
			},
		),
		UserID:      userID,
		ActivatedAt: time.Now(),
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.Priority)
	return event
}

// RoleAssignedEvent represents a role assignment event
type RoleAssignedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID     string    `json:"user_id"`
	RoleType   RoleType  `json:"role_type"`
	AssignedBy string    `json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
}

// NewRoleAssignedEvent creates a new RoleAssignedEvent
func NewRoleAssignedEvent(userID string, roleType RoleType, assignedBy string, version int) *RoleAssignedEvent {
	event := &RoleAssignedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			RoleAssignedEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":     userID,
				"role_type":   roleType.String(),
				"assigned_by": assignedBy,
				"assigned_at": time.Now(),
			},
		),
		UserID:     userID,
		RoleType:   roleType,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.Priority)
	return event
}

// RoleAssignedWithExpiryEvent represents a role assignment with expiry event
type RoleAssignedWithExpiryEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID     string    `json:"user_id"`
	RoleType   RoleType  `json:"role_type"`
	AssignedBy string    `json:"assigned_by"`
	AssignedAt time.Time `json:"assigned_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// NewRoleAssignedWithExpiryEvent creates a new RoleAssignedWithExpiryEvent
func NewRoleAssignedWithExpiryEvent(userID string, roleType RoleType, assignedBy string, expiresAt time.Time, version int) *RoleAssignedWithExpiryEvent {
	event := &RoleAssignedWithExpiryEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			RoleAssignedWithExpiryEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":     userID,
				"role_type":   roleType.String(),
				"assigned_by": assignedBy,
				"assigned_at": time.Now(),
				"expires_at":  expiresAt,
			},
		),
		UserID:     userID,
		RoleType:   roleType,
		AssignedBy: assignedBy,
		AssignedAt: time.Now(),
		ExpiresAt:  expiresAt,
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.Priority)
	return event
}

// RoleRevokedEvent represents a role revocation event
type RoleRevokedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID    string    `json:"user_id"`
	RoleType  RoleType  `json:"role_type"`
	RevokedBy string    `json:"revoked_by"`
	RevokedAt time.Time `json:"revoked_at"`
}

// NewRoleRevokedEvent creates a new RoleRevokedEvent
func NewRoleRevokedEvent(userID string, roleType RoleType, revokedBy string, version int) *RoleRevokedEvent {
	event := &RoleRevokedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			RoleRevokedEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":    userID,
				"role_type":  roleType.String(),
				"revoked_by": revokedBy,
				"revoked_at": time.Now(),
			},
		),
		UserID:    userID,
		RoleType:  roleType,
		RevokedBy: revokedBy,
		RevokedAt: time.Now(),
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.PriorityHigh)
	return event
}

// ProfileUpdatedEvent represents a profile update event
type ProfileUpdatedEvent struct {
	*cqrs.BaseDomainEventMessage
	UserID    string                 `json:"user_id"`
	Changes   map[string]interface{} `json:"changes"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NewProfileUpdatedEvent creates a new ProfileUpdatedEvent
func NewProfileUpdatedEvent(userID string, changes map[string]interface{}, version int) *ProfileUpdatedEvent {
	event := &ProfileUpdatedEvent{
		BaseDomainEventMessage: cqrs.NewBaseDomainEventMessage(
			ProfileUpdatedEventType,
			userID,
			"User",
			version,
			map[string]interface{}{
				"user_id":    userID,
				"changes":    changes,
				"updated_at": time.Now(),
			},
		),
		UserID:    userID,
		Changes:   changes,
		UpdatedAt: time.Now(),
	}

	event.SetCategory(cqrs.DomainEvent)
	event.SetPriority(cqrs.Priority)
	return event
}

// Event factory function for deserialization
func CreateEventFromType(eventType string, eventData map[string]interface{}) (cqrs.EventMessage, error) {
	switch eventType {
	case UserCreatedEventType:
		return &UserCreatedEvent{
			UserID:    eventData["user_id"].(string),
			Email:     eventData["email"].(string),
			Name:      eventData["name"].(string),
			CreatedAt: eventData["created_at"].(time.Time),
		}, nil

	case EmailChangedEventType:
		return &EmailChangedEvent{
			UserID:   eventData["user_id"].(string),
			OldEmail: eventData["old_email"].(string),
			NewEmail: eventData["new_email"].(string),
		}, nil

	case UserDeactivatedEventType:
		return &UserDeactivatedEvent{
			UserID:        eventData["user_id"].(string),
			DeactivatedAt: eventData["deactivated_at"].(time.Time),
			Reason:        eventData["reason"].(string),
		}, nil

	case UserActivatedEventType:
		return &UserActivatedEvent{
			UserID:      eventData["user_id"].(string),
			ActivatedAt: eventData["activated_at"].(time.Time),
		}, nil

	default:
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventValidation.String(), "unknown event type: "+eventType, nil)
	}
}
