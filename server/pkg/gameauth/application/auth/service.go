package auth

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/pkg/gameauth/application/providers"
	"defense-allies-server/pkg/gameauth/domain/accountlink"
	"defense-allies-server/pkg/gameauth/domain/authsession"
	"defense-allies-server/pkg/gameauth/domain/common"
	"defense-allies-server/pkg/gameauth/infrastructure/uuid"
)

type Service struct {
	providerRegistry *providers.ProviderRegistry
	accountLinkRepo  accountlink.Repository
	authSessionRepo  authsession.Repository
	idGenerator      uuid.Generator
	sessionTTL       time.Duration
}

func NewService(
	providerRegistry *providers.ProviderRegistry,
	accountLinkRepo accountlink.Repository,
	authSessionRepo authsession.Repository,
	idGenerator uuid.Generator,
	sessionTTL time.Duration,
) *Service {
	return &Service{
		providerRegistry: providerRegistry,
		accountLinkRepo:  accountLinkRepo,
		authSessionRepo:  authSessionRepo,
		idGenerator:      idGenerator,
		sessionTTL:       sessionTTL,
	}
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	provider, exists := s.providerRegistry.GetProvider(req.ProviderType)
	if !exists {
		return nil, fmt.Errorf("unsupported provider type: %s", req.ProviderType)
	}

	authResult, err := provider.Authenticate(ctx, req.Credentials)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	var accountLink *accountlink.AccountLink

	if authResult.IsNewAccount {
		accountLink, err = s.createAccountLink(ctx, authResult.GameAccountID, authResult)
		if err != nil {
			return nil, fmt.Errorf("failed to create account link: %w", err)
		}
	} else {
		accountLink, err = s.accountLinkRepo.FindByGameAccountID(ctx, authResult.GameAccountID)
		if err != nil {
			return nil, fmt.Errorf("failed to find account link: %w", err)
		}
		if accountLink == nil {
			return nil, fmt.Errorf("account link not found for game account: %s", authResult.GameAccountID)
		}
	}

	session, sessionToken, refreshToken, err := s.createAuthSession(ctx, accountLink, req.ClientInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", err)
	}

	return &LoginResponse{
		GameAccountID:   authResult.GameAccountID,
		SessionToken:    sessionToken,
		RefreshToken:    refreshToken,
		ExpiresAt:       session.ExpiresAt(),
		IsNewAccount:    authResult.IsNewAccount,
		ProviderType:    authResult.ProviderType,
		LinkedProviders: accountLink.AuthProviders(),
	}, nil
}

func (s *Service) createAccountLink(ctx context.Context, gameAccountID string, authResult *common.AuthResult) (*accountlink.AccountLink, error) {
	accountLinkID := s.idGenerator.NewID()

	accountLink, err := accountlink.NewAccountLink(
		accountLinkID,
		gameAccountID,
		authResult.ProviderType,
		authResult.ExternalID,
		authResult.Metadata,
	)
	if err != nil {
		return nil, err
	}

	if err := s.accountLinkRepo.Save(ctx, accountLink); err != nil {
		return nil, err
	}

	return accountLink, nil
}

func (s *Service) createAuthSession(ctx context.Context, accountLink *accountlink.AccountLink, clientInfo common.ClientInfo) (*authsession.AuthSession, string, string, error) {
	sessionID := s.idGenerator.NewID()
	sessionToken := s.idGenerator.NewID()
	refreshToken := s.idGenerator.NewID()
	expiresAt := time.Now().Add(s.sessionTTL)

	providerType := common.ProviderTypeGuest
	if providers := accountLink.AuthProviders(); len(providers) > 0 {
		for pt := range providers {
			providerType = pt
			break
		}
	}

	session, err := authsession.NewAuthSession(
		sessionID,
		accountLink.GameAccountID(),
		accountLink.ID(),
		providerType,
		sessionToken,
		refreshToken,
		expiresAt,
		clientInfo,
	)
	if err != nil {
		return nil, "", "", err
	}

	if err := s.authSessionRepo.Save(ctx, session); err != nil {
		return nil, "", "", err
	}

	return session, sessionToken, refreshToken, nil
}
