package common

import "time"

type ProviderType string

const (
	ProviderTypeGuest  ProviderType = "guest"
	ProviderTypeApple  ProviderType = "apple"
	ProviderTypeGoogle ProviderType = "google"
)

type AuthProviderInfo struct {
	ProviderType ProviderType `json:"provider_type"`
	ExternalID   string       `json:"external_id"`
	LinkedAt     time.Time    `json:"linked_at"`
	Metadata     interface{}  `json:"metadata,omitempty"`
}

type DeviceInfo struct {
	DeviceID   string `json:"device_id"`
	DeviceType string `json:"device_type"`
	Platform   string `json:"platform"`
	Version    string `json:"version,omitempty"`
}

type ClientInfo struct {
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	Device    DeviceInfo `json:"device"`
}

type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusExpired SessionStatus = "expired"
	SessionStatusRevoked SessionStatus = "revoked"
)

type AuthResult struct {
	GameAccountID string                      `json:"game_account_id"`
	ProviderType  ProviderType                `json:"provider_type"`
	ExternalID    string                      `json:"external_id"`
	IsNewAccount  bool                        `json:"is_new_account"`
	Metadata      map[string]interface{}      `json:"metadata,omitempty"`
	Providers     map[ProviderType]AuthProviderInfo `json:"providers"`
}