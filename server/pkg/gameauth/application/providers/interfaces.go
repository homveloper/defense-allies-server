package providers

import (
	"context"

	"defense-allies-server/pkg/gameauth/domain/common"
)

type AuthProvider interface {
	ProviderType() common.ProviderType
	Authenticate(ctx context.Context, credentials interface{}) (*common.AuthResult, error)
	GenerateGameID(ctx context.Context, externalID string) (string, error)
	ValidateCredentials(credentials interface{}) error
}

type GuestCredentials struct {
	DeviceID   string                 `json:"device_id"`
	DeviceInfo common.DeviceInfo      `json:"device_info"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Registry interface {
	Register(provider AuthProvider)
	GetProvider(providerType common.ProviderType) (AuthProvider, bool)
	GetAllProviders() map[common.ProviderType]AuthProvider
}