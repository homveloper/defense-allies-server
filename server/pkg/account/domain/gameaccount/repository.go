package gameaccount

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, gameAccount *GameAccount) error
	Load(ctx context.Context, id string) (*GameAccount, error)
	FindByUsername(ctx context.Context, username string) (*GameAccount, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}