package accountlink

import (
	"context"

	"defense-allies-server/pkg/gameauth/domain/common"
)

type Repository interface {
	Save(ctx context.Context, accountLink *AccountLink) error
	Load(ctx context.Context, id string) (*AccountLink, error)
	FindByGameAccountID(ctx context.Context, gameAccountID string) (*AccountLink, error)
	FindByProvider(ctx context.Context, providerType common.ProviderType, externalID string) (*AccountLink, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}