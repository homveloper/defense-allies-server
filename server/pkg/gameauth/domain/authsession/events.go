package authsession

import (
	"time"

	"cqrs"
	"defense-allies-server/pkg/gameauth/domain/common"
)

const (
	EventTypeAuthSessionCreated   = "AuthSessionCreated"
	EventTypeAuthSessionRefreshed = "AuthSessionRefreshed"
	EventTypeActivityUpdated      = "ActivityUpdated"
	EventTypeAuthSessionRevoked   = "AuthSessionRevoked"
	EventTypeAuthSessionExpired   = "AuthSessionExpired"
)

type AuthSessionCreatedEvent struct {
	*cqrs.BaseEventMessage
	GameAccountID  string              `json:"game_account_id"`
	AccountLinkID  string              `json:"account_link_id"`
	ProviderType   common.ProviderType `json:"provider_type"`
	SessionToken   string              `json:"session_token"`
	RefreshToken   string              `json:"refresh_token"`
	ExpiresAt      time.Time           `json:"expires_at"`
	LastActivityAt time.Time           `json:"last_activity_at"`
	ClientInfo     common.ClientInfo   `json:"client_info"`
}

func NewAuthSessionCreatedEvent(sessionID string, gameAccountID string, accountLinkID string, providerType common.ProviderType, sessionToken string, refreshToken string, expiresAt time.Time, clientInfo common.ClientInfo) *AuthSessionCreatedEvent {
	return &AuthSessionCreatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeAuthSessionCreated),
		GameAccountID:    gameAccountID,
		AccountLinkID:    accountLinkID,
		ProviderType:     providerType,
		SessionToken:     sessionToken,
		RefreshToken:     refreshToken,
		ExpiresAt:        expiresAt,
		LastActivityAt:   time.Now(),
		ClientInfo:       clientInfo,
	}
}

type AuthSessionRefreshedEvent struct {
	*cqrs.BaseEventMessage
	NewSessionToken string    `json:"new_session_token"`
	NewRefreshToken string    `json:"new_refresh_token"`
	NewExpiresAt    time.Time `json:"new_expires_at"`
}

func NewAuthSessionRefreshedEvent(sessionID string, newSessionToken string, newRefreshToken string, newExpiresAt time.Time) *AuthSessionRefreshedEvent {
	return &AuthSessionRefreshedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeAuthSessionRefreshed),
		NewSessionToken:  newSessionToken,
		NewRefreshToken:  newRefreshToken,
		NewExpiresAt:     newExpiresAt,
	}
}

type ActivityUpdatedEvent struct {
	*cqrs.BaseEventMessage
	ActivityTime time.Time `json:"activity_time"`
}

func NewActivityUpdatedEvent(sessionID string, activityTime time.Time) *ActivityUpdatedEvent {
	return &ActivityUpdatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeActivityUpdated),
		ActivityTime:     activityTime,
	}
}

type AuthSessionRevokedEvent struct {
	*cqrs.BaseEventMessage
}

func NewAuthSessionRevokedEvent(sessionID string) *AuthSessionRevokedEvent {
	return &AuthSessionRevokedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeAuthSessionRevoked),
	}
}

type AuthSessionExpiredEvent struct {
	*cqrs.BaseEventMessage
}

func NewAuthSessionExpiredEvent(sessionID string) *AuthSessionExpiredEvent {
	return &AuthSessionExpiredEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeAuthSessionExpired),
	}
}
