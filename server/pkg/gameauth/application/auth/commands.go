package auth

import (
	"time"

	"defense-allies-server/pkg/gameauth/domain/common"
)

type LoginRequest struct {
	ProviderType common.ProviderType `json:"provider_type"`
	Credentials  interface{}         `json:"credentials"`
	ClientInfo   common.ClientInfo   `json:"client_info"`
}

type LoginResponse struct {
	GameAccountID   string                                          `json:"game_account_id"`
	SessionToken    string                                          `json:"session_token"`
	RefreshToken    string                                          `json:"refresh_token"`
	ExpiresAt       time.Time                                       `json:"expires_at"`
	IsNewAccount    bool                                            `json:"is_new_account"`
	ProviderType    common.ProviderType                             `json:"provider_type"`
	LinkedProviders map[common.ProviderType]common.AuthProviderInfo `json:"linked_providers"`
}

type RefreshSessionRequest struct {
	RefreshToken string            `json:"refresh_token"`
	ClientInfo   common.ClientInfo `json:"client_info"`
}

type RefreshSessionResponse struct {
	SessionToken string    `json:"session_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type LogoutRequest struct {
	SessionToken string `json:"session_token"`
}

type ValidateSessionRequest struct {
	SessionToken string `json:"session_token"`
}

type ValidateSessionResponse struct {
	GameAccountID   string                                          `json:"game_account_id"`
	AccountLinkID   string                                          `json:"account_link_id"`
	ProviderType    common.ProviderType                             `json:"provider_type"`
	ExpiresAt       time.Time                                       `json:"expires_at"`
	LinkedProviders map[common.ProviderType]common.AuthProviderInfo `json:"linked_providers"`
}
