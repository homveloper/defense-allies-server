package auth

import (
	"context"
	"fmt"
)

func (s *Service) RefreshSession(ctx context.Context, req *RefreshSessionRequest) (*RefreshSessionResponse, error) {
	session, err := s.authSessionRepo.FindByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to find session by refresh token: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active")
	}

	newSessionToken := s.idGenerator.NewID()
	newRefreshToken := s.idGenerator.NewID()
	newExpiresAt := session.ExpiresAt().Add(s.sessionTTL)

	if err := session.Refresh(newSessionToken, newRefreshToken, newExpiresAt); err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	if err := s.authSessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save refreshed session: %w", err)
	}

	return &RefreshSessionResponse{
		SessionToken: newSessionToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    newExpiresAt,
	}, nil
}

func (s *Service) ValidateSession(ctx context.Context, req *ValidateSessionRequest) (*ValidateSessionResponse, error) {
	session, err := s.authSessionRepo.FindBySessionToken(ctx, req.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("failed to find session by token: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("invalid session token")
	}

	if !session.IsActive() {
		return nil, fmt.Errorf("session is not active")
	}

	if err := session.UpdateActivity(); err != nil {
		return nil, fmt.Errorf("failed to update session activity: %w", err)
	}

	if err := s.authSessionRepo.Save(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to save session activity: %w", err)
	}

	accountLink, err := s.accountLinkRepo.Load(ctx, session.AccountLinkID())
	if err != nil {
		return nil, fmt.Errorf("failed to load account link: %w", err)
	}

	return &ValidateSessionResponse{
		GameAccountID:   session.GameAccountID(),
		AccountLinkID:   session.AccountLinkID(),
		ProviderType:    session.ProviderType(),
		ExpiresAt:       session.ExpiresAt(),
		LinkedProviders: accountLink.AuthProviders(),
	}, nil
}

func (s *Service) Logout(ctx context.Context, req *LogoutRequest) error {
	session, err := s.authSessionRepo.FindBySessionToken(ctx, req.SessionToken)
	if err != nil {
		return fmt.Errorf("failed to find session by token: %w", err)
	}
	if session == nil {
		return fmt.Errorf("invalid session token")
	}

	if err := session.Revoke(); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	if err := s.authSessionRepo.Save(ctx, session); err != nil {
		return fmt.Errorf("failed to save revoked session: %w", err)
	}

	return nil
}
