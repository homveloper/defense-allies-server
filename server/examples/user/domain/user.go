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
	}

	// Apply the creation event
	event := NewUserCreatedEvent(userID, email, name)
	user.Apply(event, true)

	return user, nil
}

// LoadUserFromHistory creates a User aggregate from event history
func LoadUserFromHistory(userID string, events []cqrs.EventMessage) (*User, error) {
	user := &User{
		BaseAggregate: cqrs.NewBaseAggregate(userID, "User"),
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

	event := NewEmailChangedEvent(u.AggregateID(), u.email, newEmail, u.CurrentVersion()+1)
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

	event := NewUserDeactivatedEvent(u.AggregateID(), reason, u.CurrentVersion()+1)
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

	event := NewUserActivatedEvent(u.AggregateID(), u.CurrentVersion()+1)
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
