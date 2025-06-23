package accountlink

import (
	"cqrs"
	"defense-allies-server/pkg/gameauth/domain/common"
)

const (
	EventTypeAccountLinkCreated       = "AccountLinkCreated"
	EventTypeProviderLinked           = "ProviderLinked"
	EventTypeProviderUnlinked         = "ProviderUnlinked"
	EventTypeMetadataUpdated          = "MetadataUpdated"
	EventTypeAccountLinkStatusChanged = "AccountLinkStatusChanged"
)

type AccountLinkCreatedEvent struct {
	*cqrs.BaseEventMessage
	GameAccountID string                  `json:"game_account_id"`
	ProviderInfo  common.AuthProviderInfo `json:"provider_info"`
}

func NewAccountLinkCreatedEvent(accountLinkID string, gameAccountID string, providerInfo common.AuthProviderInfo) *AccountLinkCreatedEvent {
	return &AccountLinkCreatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeAccountLinkCreated),
		GameAccountID:    gameAccountID,
		ProviderInfo:     providerInfo,
	}
}

type ProviderLinkedEvent struct {
	*cqrs.BaseEventMessage
	ProviderInfo common.AuthProviderInfo `json:"provider_info"`
}

func NewProviderLinkedEvent(accountLinkID string, providerInfo common.AuthProviderInfo) *ProviderLinkedEvent {
	return &ProviderLinkedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeProviderLinked),
		ProviderInfo:     providerInfo,
	}
}

type ProviderUnlinkedEvent struct {
	*cqrs.BaseEventMessage
	ProviderType common.ProviderType `json:"provider_type"`
}

func NewProviderUnlinkedEvent(accountLinkID string, providerType common.ProviderType) *ProviderUnlinkedEvent {
	return &ProviderUnlinkedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeProviderUnlinked),
		ProviderType:     providerType,
	}
}

type MetadataUpdatedEvent struct {
	*cqrs.BaseEventMessage
	AccountMetadata map[string]interface{} `json:"account_metadata"`
}

func NewMetadataUpdatedEvent(accountLinkID string, metadata map[string]interface{}) *MetadataUpdatedEvent {
	return &MetadataUpdatedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeMetadataUpdated),
		AccountMetadata:  metadata,
	}
}

type AccountLinkStatusChangedEvent struct {
	*cqrs.BaseEventMessage
	Status AccountLinkStatus `json:"status"`
}

func NewAccountLinkStatusChangedEvent(accountLinkID string, status AccountLinkStatus) *AccountLinkStatusChangedEvent {
	return &AccountLinkStatusChangedEvent{
		BaseEventMessage: cqrs.NewBaseEventMessage(EventTypeAccountLinkStatusChanged),
		Status:           status,
	}
}
