package authsession

import (
	"context"
	"time"
)

type Repository interface {
	Save(ctx context.Context, session *AuthSession) error
	Load(ctx context.Context, id string) (*AuthSession, error)
	FindBySessionToken(ctx context.Context, sessionToken string) (*AuthSession, error)
	FindByRefreshToken(ctx context.Context, refreshToken string) (*AuthSession, error)
	FindActiveByGameAccountID(ctx context.Context, gameAccountID string) ([]*AuthSession, error)
	DeleteExpired(ctx context.Context, before time.Time) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}