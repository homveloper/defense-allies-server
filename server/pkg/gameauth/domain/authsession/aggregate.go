package authsession

import (
	"errors"
	"time"

	"cqrs"
	"defense-allies-server/pkg/gameauth/domain/common"
)

const AggregateType = "AuthSession"

type AuthSession struct {
	*cqrs.BaseAggregate

	gameAccountID  string
	accountLinkID  string
	providerType   common.ProviderType
	sessionToken   string
	refreshToken   string
	status         common.SessionStatus
	expiresAt      time.Time
	lastActivityAt time.Time
	clientInfo     common.ClientInfo
}

func NewAuthSession(id string, gameAccountID string, accountLinkID string, providerType common.ProviderType, sessionToken string, refreshToken string, expiresAt time.Time, clientInfo common.ClientInfo) (*AuthSession, error) {
	if id == "" {
		return nil, errors.New("session ID cannot be empty")
	}
	if gameAccountID == "" {
		return nil, errors.New("game account ID cannot be empty")
	}
	if accountLinkID == "" {
		return nil, errors.New("account link ID cannot be empty")
	}
	if sessionToken == "" {
		return nil, errors.New("session token cannot be empty")
	}

	authSession := &AuthSession{
		BaseAggregate:  cqrs.NewBaseAggregate(id, AggregateType),
		gameAccountID:  gameAccountID,
		accountLinkID:  accountLinkID,
		providerType:   providerType,
		sessionToken:   sessionToken,
		refreshToken:   refreshToken,
		status:         common.SessionStatusActive,
		expiresAt:      expiresAt,
		lastActivityAt: time.Now(),
		clientInfo:     clientInfo,
	}

	event := NewAuthSessionCreatedEvent(id, gameAccountID, accountLinkID, providerType, sessionToken, refreshToken, expiresAt, clientInfo)
	if err := authSession.BaseAggregate.ApplyEvent(event); err != nil {
		return nil, err
	}

	authSession.apply(event)
	return authSession, nil
}

func LoadAuthSession(id string, options ...cqrs.BaseAggregateOption) *AuthSession {
	return &AuthSession{
		BaseAggregate: cqrs.NewBaseAggregate(id, AggregateType, options...),
		status:        common.SessionStatusActive,
	}
}

func (s *AuthSession) Refresh(newSessionToken string, newRefreshToken string, newExpiresAt time.Time) error {
	if s.status != common.SessionStatusActive {
		return errors.New("cannot refresh inactive session")
	}

	if s.IsExpired() {
		return errors.New("cannot refresh expired session")
	}

	event := NewAuthSessionRefreshedEvent(s.ID(), newSessionToken, newRefreshToken, newExpiresAt)
	if err := s.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	s.apply(event)
	return nil
}

func (s *AuthSession) UpdateActivity() error {
	if s.status != common.SessionStatusActive {
		return errors.New("cannot update activity for inactive session")
	}

	if s.IsExpired() {
		return errors.New("cannot update activity for expired session")
	}

	event := NewActivityUpdatedEvent(s.ID(), time.Now())
	if err := s.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	s.apply(event)
	return nil
}

func (s *AuthSession) Revoke() error {
	if s.status == common.SessionStatusRevoked {
		return errors.New("session already revoked")
	}

	event := NewAuthSessionRevokedEvent(s.ID())
	if err := s.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	s.apply(event)
	return nil
}

func (s *AuthSession) Expire() error {
	if s.status == common.SessionStatusExpired {
		return errors.New("session already expired")
	}

	if s.status == common.SessionStatusRevoked {
		return errors.New("cannot expire revoked session")
	}

	event := NewAuthSessionExpiredEvent(s.ID())
	if err := s.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	s.apply(event)
	return nil
}

func (s *AuthSession) apply(event cqrs.EventMessage) {
	switch e := event.(type) {
	case *AuthSessionCreatedEvent:
		s.gameAccountID = e.GameAccountID
		s.accountLinkID = e.AccountLinkID
		s.providerType = e.ProviderType
		s.sessionToken = e.SessionToken
		s.refreshToken = e.RefreshToken
		s.status = common.SessionStatusActive
		s.expiresAt = e.ExpiresAt
		s.lastActivityAt = e.LastActivityAt
		s.clientInfo = e.ClientInfo
	case *AuthSessionRefreshedEvent:
		s.sessionToken = e.NewSessionToken
		s.refreshToken = e.NewRefreshToken
		s.expiresAt = e.NewExpiresAt
		s.lastActivityAt = time.Now()
	case *ActivityUpdatedEvent:
		s.lastActivityAt = e.ActivityTime
	case *AuthSessionRevokedEvent:
		s.status = common.SessionStatusRevoked
	case *AuthSessionExpiredEvent:
		s.status = common.SessionStatusExpired
	}
}

func (s *AuthSession) GameAccountID() string {
	return s.gameAccountID
}

func (s *AuthSession) AccountLinkID() string {
	return s.accountLinkID
}

func (s *AuthSession) ProviderType() common.ProviderType {
	return s.providerType
}

func (s *AuthSession) SessionToken() string {
	return s.sessionToken
}

func (s *AuthSession) RefreshToken() string {
	return s.refreshToken
}

func (s *AuthSession) Status() common.SessionStatus {
	return s.status
}

func (s *AuthSession) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s *AuthSession) LastActivityAt() time.Time {
	return s.lastActivityAt
}

func (s *AuthSession) ClientInfo() common.ClientInfo {
	return s.clientInfo
}

func (s *AuthSession) IsActive() bool {
	return s.status == common.SessionStatusActive && !s.IsExpired()
}

func (s *AuthSession) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

func (s *AuthSession) IsRevoked() bool {
	return s.status == common.SessionStatusRevoked
}
