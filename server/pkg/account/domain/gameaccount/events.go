package gameaccount

import (
	"time"

	"defense-allies-server/pkg/cqrs"
)

const (
	EventTypeGameAccountCreated      = "GameAccountCreated"
	EventTypeDisplayNameUpdated      = "DisplayNameUpdated"
	EventTypeMetadataUpdated         = "MetadataUpdated"
	EventTypeLoginRecorded           = "LoginRecorded"
	EventTypeGameAccountStatusChanged = "GameAccountStatusChanged"
)

type GameAccountCreatedEvent struct {
	*cqrs.BaseEventMessage
	Username    string                 `json:"username"`
	DisplayName string                 `json:"display_name"`
	GameMetadata map[string]interface{} `json:"game_metadata"`
}

func NewGameAccountCreatedEvent(gameAccountID string, username string, displayName string, metadata map[string]interface{}) *GameAccountCreatedEvent {
	return &GameAccountCreatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeGameAccountCreated),
		Username:         username,
		DisplayName:      displayName,
		GameMetadata:     metadata,
	}
}

type DisplayNameUpdatedEvent struct {
	*cqrs.BaseEventMessage
	DisplayName string `json:"display_name"`
}

func NewDisplayNameUpdatedEvent(gameAccountID string, displayName string) *DisplayNameUpdatedEvent {
	return &DisplayNameUpdatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeDisplayNameUpdated),
		DisplayName:      displayName,
	}
}

type MetadataUpdatedEvent struct {
	*cqrs.BaseEventMessage
	GameMetadata map[string]interface{} `json:"game_metadata"`
}

func NewMetadataUpdatedEvent(gameAccountID string, metadata map[string]interface{}) *MetadataUpdatedEvent {
	return &MetadataUpdatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeMetadataUpdated),
		GameMetadata:     metadata,
	}
}

type LoginRecordedEvent struct {
	*cqrs.BaseEventMessage
	LoginTime time.Time `json:"login_time"`
}

func NewLoginRecordedEvent(gameAccountID string, loginTime time.Time) *LoginRecordedEvent {
	return &LoginRecordedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeLoginRecorded),
		LoginTime:        loginTime,
	}
}

type GameAccountStatusChangedEvent struct {
	*cqrs.BaseEventMessage
	Status GameAccountStatus `json:"status"`
}

func NewGameAccountStatusChangedEvent(gameAccountID string, status GameAccountStatus) *GameAccountStatusChangedEvent {
	return &GameAccountStatusChangedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeGameAccountStatusChanged),
		Status:           status,
	}
}