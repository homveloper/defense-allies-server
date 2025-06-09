package domain

import (
	"encoding/json"
	"time"

	"cqrs"
)

// V2 이벤트들 - 확장 버전 (추가 필드 포함)

// UserProfile represents user profile information (V2 addition)
type UserProfile struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
}

// UserPreferences represents user preferences (V2 addition)
type UserPreferences struct {
	Language           string                 `json:"language"`
	Timezone           string                 `json:"timezone"`
	EmailNotifications bool                   `json:"email_notifications"`
	SMSNotifications   bool                   `json:"sms_notifications"`
	Theme              string                 `json:"theme"`
	CustomSettings     map[string]interface{} `json:"custom_settings"`
}

// DefaultUserProfile returns default profile for migration
func DefaultUserProfile() UserProfile {
	return UserProfile{
		FirstName:   "",
		LastName:    "",
		DateOfBirth: time.Time{},
		Gender:      "not_specified",
		Avatar:      "",
		Bio:         "",
	}
}

// DefaultUserPreferences returns default preferences for migration
func DefaultUserPreferences() UserPreferences {
	return UserPreferences{
		Language:           "en",
		Timezone:           "UTC",
		EmailNotifications: true,
		SMSNotifications:   false,
		Theme:              "light",
		CustomSettings:     make(map[string]interface{}),
	}
}

// UserCreatedV2 represents the V2 version of user creation event
type UserCreatedV2 struct {
	*cqrs.BaseEventMessage
	UserID      string          `json:"user_id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Profile     UserProfile     `json:"profile"`     // V2 addition
	Preferences UserPreferences `json:"preferences"` // V2 addition
	CreatedAt   time.Time       `json:"created_at"`  // V2 addition
}

// NewUserCreatedV2 creates a new V2 user created event
func NewUserCreatedV2(userID, name, email string, profile UserProfile, preferences UserPreferences) *UserCreatedV2 {
	return &UserCreatedV2{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserCreated",
			userID,
			"User",
			1,
			map[string]interface{}{
				"version": "2.0",
				"schema":  "user_created_v2",
			},
		),
		UserID:      userID,
		Name:        name,
		Email:       email,
		Profile:     profile,
		Preferences: preferences,
		CreatedAt:   time.Now(),
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserCreatedV2) EventData() interface{} {
	return map[string]interface{}{
		"user_id":     e.UserID,
		"name":        e.Name,
		"email":       e.Email,
		"profile":     e.Profile,
		"preferences": e.Preferences,
		"created_at":  e.CreatedAt,
	}
}

// UserUpdatedV2 represents the V2 version of user update event
type UserUpdatedV2 struct {
	*cqrs.BaseEventMessage
	UserID      string          `json:"user_id"`
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Profile     UserProfile     `json:"profile"`     // V2 addition
	Preferences UserPreferences `json:"preferences"` // V2 addition
	UpdatedAt   time.Time       `json:"updated_at"`  // V2 addition
	UpdatedBy   string          `json:"updated_by"`  // V2 addition
}

// NewUserUpdatedV2 creates a new V2 user updated event
func NewUserUpdatedV2(userID, name, email, updatedBy string, profile UserProfile, preferences UserPreferences) *UserUpdatedV2 {
	return &UserUpdatedV2{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserUpdated",
			userID,
			"User",
			1,
			map[string]interface{}{
				"version": "2.0",
				"schema":  "user_updated_v2",
			},
		),
		UserID:      userID,
		Name:        name,
		Email:       email,
		Profile:     profile,
		Preferences: preferences,
		UpdatedAt:   time.Now(),
		UpdatedBy:   updatedBy,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserUpdatedV2) EventData() interface{} {
	return map[string]interface{}{
		"user_id":     e.UserID,
		"name":        e.Name,
		"email":       e.Email,
		"profile":     e.Profile,
		"preferences": e.Preferences,
		"updated_at":  e.UpdatedAt,
		"updated_by":  e.UpdatedBy,
	}
}

// UserDeletedV2 represents the V2 version of user deletion event
type UserDeletedV2 struct {
	*cqrs.BaseEventMessage
	UserID        string                 `json:"user_id"`
	DeletedAt     time.Time              `json:"deleted_at"`
	DeletedBy     string                 `json:"deleted_by"` // V2 addition
	Reason        string                 `json:"reason"`
	EventMetadata map[string]interface{} `json:"event_metadata"` // V2 addition (renamed to avoid conflict)
}

// NewUserDeletedV2 creates a new V2 user deleted event
func NewUserDeletedV2(userID, reason, deletedBy string, metadata map[string]interface{}) *UserDeletedV2 {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &UserDeletedV2{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserDeleted",
			userID,
			"User",
			1,
			map[string]interface{}{
				"version": "2.0",
				"schema":  "user_deleted_v2",
			},
		),
		UserID:        userID,
		DeletedAt:     time.Now(),
		DeletedBy:     deletedBy,
		Reason:        reason,
		EventMetadata: metadata,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserDeletedV2) EventData() interface{} {
	return map[string]interface{}{
		"user_id":        e.UserID,
		"deleted_at":     e.DeletedAt,
		"deleted_by":     e.DeletedBy,
		"reason":         e.Reason,
		"event_metadata": e.EventMetadata,
	}
}

// V2EventFactory creates V2 events from raw data
type V2EventFactory struct{}

// CreateEvent creates a V2 event from event type and data
func (f *V2EventFactory) CreateEvent(eventType string, aggregateID string, eventData []byte, metadata map[string]interface{}) (cqrs.EventMessage, error) {
	switch eventType {
	case "UserCreated":
		return f.createUserCreatedV2(aggregateID, eventData, metadata)
	case "UserUpdated":
		return f.createUserUpdatedV2(aggregateID, eventData, metadata)
	case "UserDeleted":
		return f.createUserDeletedV2(aggregateID, eventData, metadata)
	default:
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "unknown event type: "+eventType, nil)
	}
}

// createUserCreatedV2 creates UserCreatedV2 from raw data
func (f *V2EventFactory) createUserCreatedV2(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserCreatedV2, error) {
	var data struct {
		UserID      string          `json:"user_id"`
		Name        string          `json:"name"`
		Email       string          `json:"email"`
		Profile     UserProfile     `json:"profile"`
		Preferences UserPreferences `json:"preferences"`
		CreatedAt   time.Time       `json:"created_at"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserCreatedV2", err)
	}

	event := &UserCreatedV2{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserCreated",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:      data.UserID,
		Name:        data.Name,
		Email:       data.Email,
		Profile:     data.Profile,
		Preferences: data.Preferences,
		CreatedAt:   data.CreatedAt,
	}

	return event, nil
}

// createUserUpdatedV2 creates UserUpdatedV2 from raw data
func (f *V2EventFactory) createUserUpdatedV2(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserUpdatedV2, error) {
	var data struct {
		UserID      string          `json:"user_id"`
		Name        string          `json:"name"`
		Email       string          `json:"email"`
		Profile     UserProfile     `json:"profile"`
		Preferences UserPreferences `json:"preferences"`
		UpdatedAt   time.Time       `json:"updated_at"`
		UpdatedBy   string          `json:"updated_by"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserUpdatedV2", err)
	}

	event := &UserUpdatedV2{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserUpdated",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:      data.UserID,
		Name:        data.Name,
		Email:       data.Email,
		Profile:     data.Profile,
		Preferences: data.Preferences,
		UpdatedAt:   data.UpdatedAt,
		UpdatedBy:   data.UpdatedBy,
	}

	return event, nil
}

// createUserDeletedV2 creates UserDeletedV2 from raw data
func (f *V2EventFactory) createUserDeletedV2(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserDeletedV2, error) {
	var data struct {
		UserID        string                 `json:"user_id"`
		DeletedAt     time.Time              `json:"deleted_at"`
		DeletedBy     string                 `json:"deleted_by"`
		Reason        string                 `json:"reason"`
		EventMetadata map[string]interface{} `json:"event_metadata"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserDeletedV2", err)
	}

	event := &UserDeletedV2{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserDeleted",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:        data.UserID,
		DeletedAt:     data.DeletedAt,
		DeletedBy:     data.DeletedBy,
		Reason:        data.Reason,
		EventMetadata: data.EventMetadata,
	}

	return event, nil
}

// GetVersion returns the version of V2 events
func (f *V2EventFactory) GetVersion() int {
	return 2
}

// GetSupportedEvents returns the list of supported event types
func (f *V2EventFactory) GetSupportedEvents() []string {
	return []string{"UserCreated", "UserUpdated", "UserDeleted"}
}
