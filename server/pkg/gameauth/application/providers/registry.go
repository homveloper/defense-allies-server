package providers

import (
	"defense-allies-server/pkg/gameauth/domain/common"
)

type ProviderRegistry struct {
	providers map[common.ProviderType]AuthProvider
}

func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[common.ProviderType]AuthProvider),
	}
}

func (r *ProviderRegistry) Register(provider AuthProvider) {
	r.providers[provider.ProviderType()] = provider
}

func (r *ProviderRegistry) GetProvider(providerType common.ProviderType) (AuthProvider, bool) {
	provider, exists := r.providers[providerType]
	return provider, exists
}

func (r *ProviderRegistry) GetAllProviders() map[common.ProviderType]AuthProvider {
	providers := make(map[common.ProviderType]AuthProvider)
	for k, v := range r.providers {
		providers[k] = v
	}
	return providers
}