package gameaccount

import (
	"errors"
	"time"

	"defense-allies-server/pkg/cqrs"
)

const AggregateType = "GameAccount"

type GameAccount struct {
	*cqrs.BaseAggregate
	
	username     string
	displayName  string
	status       GameAccountStatus
	metadata     map[string]interface{}
	lastLoginAt  *time.Time
}

type GameAccountStatus string

const (
	GameAccountStatusActive    GameAccountStatus = "active"
	GameAccountStatusSuspended GameAccountStatus = "suspended"
	GameAccountStatusDeleted   GameAccountStatus = "deleted"
)

func NewGameAccount(id string, username string, displayName string, metadata map[string]interface{}) (*GameAccount, error) {
	if id == "" {
		return nil, errors.New("game account ID cannot be empty")
	}
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	gameAccount := &GameAccount{
		BaseAggregate: cqrs.NewBaseAggregate(id, AggregateType),
		username:      username,
		displayName:   displayName,
		status:        GameAccountStatusActive,
		metadata:      metadata,
	}

	event := NewGameAccountCreatedEvent(id, username, displayName, metadata)
	if err := gameAccount.BaseAggregate.ApplyEvent(event); err != nil {
		return nil, err
	}

	gameAccount.apply(event)
	return gameAccount, nil
}

func LoadGameAccount(id string, options ...cqrs.BaseAggregateOption) *GameAccount {
	return &GameAccount{
		BaseAggregate: cqrs.NewBaseAggregate(id, AggregateType, options...),
		metadata:      make(map[string]interface{}),
		status:        GameAccountStatusActive,
	}
}

func (g *GameAccount) UpdateDisplayName(displayName string) error {
	if g.status != GameAccountStatusActive {
		return errors.New("cannot update display name of inactive account")
	}

	if displayName == g.displayName {
		return errors.New("display name is already set to this value")
	}

	event := NewDisplayNameUpdatedEvent(g.ID(), displayName)
	if err := g.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	g.apply(event)
	return nil
}

func (g *GameAccount) UpdateMetadata(metadata map[string]interface{}) error {
	if g.status != GameAccountStatusActive {
		return errors.New("cannot update metadata of inactive account")
	}

	event := NewMetadataUpdatedEvent(g.ID(), metadata)
	if err := g.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	g.apply(event)
	return nil
}

func (g *GameAccount) RecordLogin() error {
	if g.status != GameAccountStatusActive {
		return errors.New("cannot record login for inactive account")
	}

	now := time.Now()
	event := NewLoginRecordedEvent(g.ID(), now)
	if err := g.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	g.apply(event)
	return nil
}

func (g *GameAccount) Suspend() error {
	if g.status == GameAccountStatusDeleted {
		return errors.New("cannot suspend deleted account")
	}

	if g.status == GameAccountStatusSuspended {
		return errors.New("account already suspended")
	}

	event := NewGameAccountStatusChangedEvent(g.ID(), GameAccountStatusSuspended)
	if err := g.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	g.apply(event)
	return nil
}

func (g *GameAccount) Activate() error {
	if g.status == GameAccountStatusDeleted {
		return errors.New("cannot activate deleted account")
	}

	if g.status == GameAccountStatusActive {
		return errors.New("account already active")
	}

	event := NewGameAccountStatusChangedEvent(g.ID(), GameAccountStatusActive)
	if err := g.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	g.apply(event)
	return nil
}

func (g *GameAccount) Delete() error {
	if g.status == GameAccountStatusDeleted {
		return errors.New("account already deleted")
	}

	event := NewGameAccountStatusChangedEvent(g.ID(), GameAccountStatusDeleted)
	if err := g.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	g.apply(event)
	return nil
}

func (g *GameAccount) apply(event cqrs.EventMessage) {
	switch e := event.(type) {
	case *GameAccountCreatedEvent:
		g.username = e.Username
		g.displayName = e.DisplayName
		g.metadata = e.GameMetadata
		g.status = GameAccountStatusActive
	case *DisplayNameUpdatedEvent:
		g.displayName = e.DisplayName
	case *MetadataUpdatedEvent:
		g.metadata = e.GameMetadata
	case *LoginRecordedEvent:
		g.lastLoginAt = &e.LoginTime
	case *GameAccountStatusChangedEvent:
		g.status = e.Status
	}
}

func (g *GameAccount) Username() string {
	return g.username
}

func (g *GameAccount) DisplayName() string {
	return g.displayName
}

func (g *GameAccount) Status() GameAccountStatus {
	return g.status
}

func (g *GameAccount) Metadata() map[string]interface{} {
	metadata := make(map[string]interface{})
	for k, v := range g.metadata {
		metadata[k] = v
	}
	return metadata
}

func (g *GameAccount) LastLoginAt() *time.Time {
	return g.lastLoginAt
}

func (g *GameAccount) IsActive() bool {
	return g.status == GameAccountStatusActive
}