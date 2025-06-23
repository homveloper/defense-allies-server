package accountlink

import (
	"errors"
	"time"

	"cqrs"

	"defense-allies-server/pkg/gameauth/domain/common"
)

const AggregateType = "AccountLink"

type AccountLink struct {
	*cqrs.BaseAggregate

	gameAccountID string
	authProviders map[common.ProviderType]common.AuthProviderInfo
	metadata      map[string]interface{}
	status        AccountLinkStatus
}

type AccountLinkStatus string

const (
	AccountLinkStatusActive    AccountLinkStatus = "active"
	AccountLinkStatusSuspended AccountLinkStatus = "suspended"
	AccountLinkStatusDeleted   AccountLinkStatus = "deleted"
)

func NewAccountLink(id string, gameAccountID string, providerType common.ProviderType, externalID string, metadata map[string]interface{}) (*AccountLink, error) {
	if id == "" {
		return nil, errors.New("account link ID cannot be empty")
	}
	if gameAccountID == "" {
		return nil, errors.New("game account ID cannot be empty")
	}
	if externalID == "" {
		return nil, errors.New("external ID cannot be empty")
	}

	accountLink := &AccountLink{
		BaseAggregate: cqrs.NewBaseAggregate(id, AggregateType),
		gameAccountID: gameAccountID,
		authProviders: make(map[common.ProviderType]common.AuthProviderInfo),
		metadata:      metadata,
		status:        AccountLinkStatusActive,
	}

	now := time.Now()
	providerInfo := common.AuthProviderInfo{
		ProviderType: providerType,
		ExternalID:   externalID,
		LinkedAt:     now,
		Metadata:     metadata,
	}

	event := NewAccountLinkCreatedEvent(id, gameAccountID, providerInfo)
	if err := accountLink.BaseAggregate.ApplyEvent(event); err != nil {
		return nil, err
	}

	accountLink.apply(event)
	return accountLink, nil
}

func LoadAccountLink(id string, options ...cqrs.BaseAggregateOption) *AccountLink {
	return &AccountLink{
		BaseAggregate: cqrs.NewBaseAggregate(id, AggregateType, options...),
		authProviders: make(map[common.ProviderType]common.AuthProviderInfo),
		metadata:      make(map[string]interface{}),
		status:        AccountLinkStatusActive,
	}
}

func (a *AccountLink) LinkProvider(providerType common.ProviderType, externalID string, metadata map[string]interface{}) error {
	if externalID == "" {
		return errors.New("external ID cannot be empty")
	}

	if _, exists := a.authProviders[providerType]; exists {
		return errors.New("provider already linked")
	}

	if a.status != AccountLinkStatusActive {
		return errors.New("cannot link provider to inactive account")
	}

	now := time.Now()
	providerInfo := common.AuthProviderInfo{
		ProviderType: providerType,
		ExternalID:   externalID,
		LinkedAt:     now,
		Metadata:     metadata,
	}

	event := NewProviderLinkedEvent(a.ID(), providerInfo)
	if err := a.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	a.apply(event)
	return nil
}

func (a *AccountLink) UnlinkProvider(providerType common.ProviderType) error {
	if _, exists := a.authProviders[providerType]; !exists {
		return errors.New("provider not linked")
	}

	if len(a.authProviders) <= 1 {
		return errors.New("cannot unlink last provider")
	}

	event := NewProviderUnlinkedEvent(a.ID(), providerType)
	if err := a.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	a.apply(event)
	return nil
}

func (a *AccountLink) UpdateMetadata(metadata map[string]interface{}) error {
	if a.status != AccountLinkStatusActive {
		return errors.New("cannot update metadata of inactive account")
	}

	event := NewMetadataUpdatedEvent(a.ID(), metadata)
	if err := a.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	a.apply(event)
	return nil
}

func (a *AccountLink) Suspend() error {
	if a.status == AccountLinkStatusDeleted {
		return errors.New("cannot suspend deleted account")
	}

	if a.status == AccountLinkStatusSuspended {
		return errors.New("account already suspended")
	}

	event := NewAccountLinkStatusChangedEvent(a.ID(), AccountLinkStatusSuspended)
	if err := a.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	a.apply(event)
	return nil
}

func (a *AccountLink) Activate() error {
	if a.status == AccountLinkStatusDeleted {
		return errors.New("cannot activate deleted account")
	}

	if a.status == AccountLinkStatusActive {
		return errors.New("account already active")
	}

	event := NewAccountLinkStatusChangedEvent(a.ID(), AccountLinkStatusActive)
	if err := a.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	a.apply(event)
	return nil
}

func (a *AccountLink) apply(event cqrs.EventMessage) {
	switch e := event.(type) {
	case *AccountLinkCreatedEvent:
		a.gameAccountID = e.GameAccountID
		a.authProviders[e.ProviderInfo.ProviderType] = e.ProviderInfo
		a.status = AccountLinkStatusActive
	case *ProviderLinkedEvent:
		a.authProviders[e.ProviderInfo.ProviderType] = e.ProviderInfo
	case *ProviderUnlinkedEvent:
		delete(a.authProviders, e.ProviderType)
	case *MetadataUpdatedEvent:
		a.metadata = e.AccountMetadata
	case *AccountLinkStatusChangedEvent:
		a.status = e.Status
	}
}

func (a *AccountLink) GameAccountID() string {
	return a.gameAccountID
}

func (a *AccountLink) AuthProviders() map[common.ProviderType]common.AuthProviderInfo {
	providers := make(map[common.ProviderType]common.AuthProviderInfo)
	for k, v := range a.authProviders {
		providers[k] = v
	}
	return providers
}

func (a *AccountLink) HasProvider(providerType common.ProviderType) bool {
	_, exists := a.authProviders[providerType]
	return exists
}

func (a *AccountLink) GetProvider(providerType common.ProviderType) (common.AuthProviderInfo, bool) {
	provider, exists := a.authProviders[providerType]
	return provider, exists
}

func (a *AccountLink) Metadata() map[string]interface{} {
	metadata := make(map[string]interface{})
	for k, v := range a.metadata {
		metadata[k] = v
	}
	return metadata
}

func (a *AccountLink) Status() AccountLinkStatus {
	return a.status
}

func (a *AccountLink) IsActive() bool {
	return a.status == AccountLinkStatusActive
}
