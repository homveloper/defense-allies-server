package domain

import (
	"encoding/json"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// V1 이벤트들 - 초기 버전 (기본 필드만 포함)

// UserCreatedV1 represents the initial version of user creation event
type UserCreatedV1 struct {
	*cqrs.BaseEventMessage
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// NewUserCreatedV1 creates a new V1 user created event
func NewUserCreatedV1(userID, name, email string) *UserCreatedV1 {
	return &UserCreatedV1{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserCreated",
			userID,
			"User",
			1, // V1
			map[string]interface{}{
				"version": "1.0",
				"schema":  "user_created_v1",
			},
		),
		UserID: userID,
		Name:   name,
		Email:  email,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserCreatedV1) EventData() interface{} {
	return map[string]interface{}{
		"user_id": e.UserID,
		"name":    e.Name,
		"email":   e.Email,
	}
}

// UserUpdatedV1 represents the initial version of user update event
type UserUpdatedV1 struct {
	*cqrs.BaseEventMessage
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

// NewUserUpdatedV1 creates a new V1 user updated event
func NewUserUpdatedV1(userID, name, email string) *UserUpdatedV1 {
	return &UserUpdatedV1{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserUpdated",
			userID,
			"User",
			1, // V1
			map[string]interface{}{
				"version": "1.0",
				"schema":  "user_updated_v1",
			},
		),
		UserID: userID,
		Name:   name,
		Email:  email,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserUpdatedV1) EventData() interface{} {
	return map[string]interface{}{
		"user_id": e.UserID,
		"name":    e.Name,
		"email":   e.Email,
	}
}

// UserDeletedV1 represents the initial version of user deletion event
type UserDeletedV1 struct {
	*cqrs.BaseEventMessage
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
	Reason    string    `json:"reason"`
}

// NewUserDeletedV1 creates a new V1 user deleted event
func NewUserDeletedV1(userID, reason string) *UserDeletedV1 {
	return &UserDeletedV1{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserDeleted",
			userID,
			"User",
			1, // V1
			map[string]interface{}{
				"version": "1.0",
				"schema":  "user_deleted_v1",
			},
		),
		UserID:    userID,
		DeletedAt: time.Now(),
		Reason:    reason,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserDeletedV1) EventData() interface{} {
	return map[string]interface{}{
		"user_id":    e.UserID,
		"deleted_at": e.DeletedAt,
		"reason":     e.Reason,
	}
}

// V1EventFactory creates V1 events from raw data
type V1EventFactory struct{}

// CreateEvent creates a V1 event from event type and data
func (f *V1EventFactory) CreateEvent(eventType string, aggregateID string, eventData []byte, metadata map[string]interface{}) (cqrs.EventMessage, error) {
	switch eventType {
	case "UserCreated":
		return f.createUserCreatedV1(aggregateID, eventData, metadata)
	case "UserUpdated":
		return f.createUserUpdatedV1(aggregateID, eventData, metadata)
	case "UserDeleted":
		return f.createUserDeletedV1(aggregateID, eventData, metadata)
	default:
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "unknown event type: "+eventType, nil)
	}
}

// createUserCreatedV1 creates UserCreatedV1 from raw data
func (f *V1EventFactory) createUserCreatedV1(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserCreatedV1, error) {
	var data struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserCreatedV1", err)
	}

	event := &UserCreatedV1{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserCreated",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID: data.UserID,
		Name:   data.Name,
		Email:  data.Email,
	}

	return event, nil
}

// createUserUpdatedV1 creates UserUpdatedV1 from raw data
func (f *V1EventFactory) createUserUpdatedV1(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserUpdatedV1, error) {
	var data struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserUpdatedV1", err)
	}

	event := &UserUpdatedV1{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserUpdated",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID: data.UserID,
		Name:   data.Name,
		Email:  data.Email,
	}

	return event, nil
}

// createUserDeletedV1 creates UserDeletedV1 from raw data
func (f *V1EventFactory) createUserDeletedV1(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserDeletedV1, error) {
	var data struct {
		UserID    string    `json:"user_id"`
		DeletedAt time.Time `json:"deleted_at"`
		Reason    string    `json:"reason"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserDeletedV1", err)
	}

	event := &UserDeletedV1{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserDeleted",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:    data.UserID,
		DeletedAt: data.DeletedAt,
		Reason:    data.Reason,
	}

	return event, nil
}

// GetVersion returns the version of V1 events
func (f *V1EventFactory) GetVersion() int {
	return 1
}

// GetSupportedEvents returns the list of supported event types
func (f *V1EventFactory) GetSupportedEvents() []string {
	return []string{"UserCreated", "UserUpdated", "UserDeleted"}
}
