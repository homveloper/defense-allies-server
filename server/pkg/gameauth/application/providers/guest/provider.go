package guest

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"defense-allies-server/pkg/gameauth/application/providers"
	"defense-allies-server/pkg/gameauth/domain/accountlink"
	"defense-allies-server/pkg/gameauth/domain/common"

	"github.com/google/uuid"
)

type Provider struct {
	accountLinkRepo accountlink.Repository
}

func NewProvider(accountLinkRepo accountlink.Repository) *Provider {
	return &Provider{
		accountLinkRepo: accountLinkRepo,
	}
}

func (p *Provider) ProviderType() common.ProviderType {
	return common.ProviderTypeGuest
}

func (p *Provider) Authenticate(ctx context.Context, credentials interface{}) (*common.AuthResult, error) {
	guestCreds, ok := credentials.(*providers.GuestCredentials)
	if !ok {
		return nil, fmt.Errorf("invalid credentials type for guest provider")
	}

	if err := p.ValidateCredentials(guestCreds); err != nil {
		return nil, fmt.Errorf("credential validation failed: %w", err)
	}

	existingLink, err := p.accountLinkRepo.FindByProvider(ctx, common.ProviderTypeGuest, guestCreds.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing account link: %w", err)
	}

	if existingLink != nil {
		// 기존 계정이 있는 경우, DeviceID와 GameAccountID 매핑 검증
		if !GameAccountIDToDeviceMapping(existingLink.GameAccountID(), guestCreds.DeviceID) {
			return nil, fmt.Errorf("security violation: GameAccountID does not match DeviceID mapping")
		}

		return &common.AuthResult{
			GameAccountID: existingLink.GameAccountID(),
			ProviderType:  common.ProviderTypeGuest,
			ExternalID:    guestCreds.DeviceID,
			IsNewAccount:  false,
			Metadata:      guestCreds.Metadata,
			Providers:     existingLink.AuthProviders(),
		}, nil
	}

	gameAccountID, err := p.GenerateGameID(ctx, guestCreds.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate game ID: %w", err)
	}

	// 새 계정 생성 전 충돌 검사
	if err := p.validateGameAccountIDUniqueness(ctx, gameAccountID, guestCreds.DeviceID); err != nil {
		// 충돌 발생 시 대안 GameAccountID 생성
		alternativeGameAccountID, recoveryErr := p.HandleCollision(ctx, guestCreds.DeviceID, gameAccountID)
		if recoveryErr != nil {
			return nil, fmt.Errorf("game account ID collision detected and recovery failed: %w", recoveryErr)
		}
		gameAccountID = alternativeGameAccountID
	}

	metadata := make(map[string]interface{})
	if guestCreds.Metadata != nil {
		metadata = guestCreds.Metadata
	}

	metadata["device_info"] = guestCreds.DeviceInfo

	providers := map[common.ProviderType]common.AuthProviderInfo{
		common.ProviderTypeGuest: {
			ProviderType: common.ProviderTypeGuest,
			ExternalID:   guestCreds.DeviceID,
			Metadata:     metadata,
		},
	}

	return &common.AuthResult{
		GameAccountID: gameAccountID,
		ProviderType:  common.ProviderTypeGuest,
		ExternalID:    guestCreds.DeviceID,
		IsNewAccount:  true,
		Metadata:      metadata,
		Providers:     providers,
	}, nil
}

func (p *Provider) GenerateGameID(ctx context.Context, externalID string) (string, error) {
	if externalID == "" {
		return "", fmt.Errorf("external ID cannot be empty")
	}

	// DeviceID를 해시하여 고정된 seed 생성
	hash := sha256.Sum256([]byte(fmt.Sprintf("guest:%s", externalID)))

	// 해시를 사용하여 deterministic UUID 생성
	// 같은 DeviceID는 항상 같은 GameAccountID를 생성
	gameAccountUUID, err := uuid.FromBytes(hash[:16])
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID from device ID: %w", err)
	}

	return gameAccountUUID.String(), nil
}

func (p *Provider) ValidateCredentials(credentials interface{}) error {
	guestCreds, ok := credentials.(*providers.GuestCredentials)
	if !ok {
		return fmt.Errorf("invalid credentials type")
	}

	if guestCreds.DeviceID == "" {
		return fmt.Errorf("device ID cannot be empty")
	}

	if len(guestCreds.DeviceID) < 3 {
		return fmt.Errorf("device ID too short")
	}

	if strings.Contains(guestCreds.DeviceID, " ") {
		return fmt.Errorf("device ID cannot contain spaces")
	}

	if guestCreds.DeviceInfo.DeviceID == "" {
		return fmt.Errorf("device info device ID cannot be empty")
	}

	if guestCreds.DeviceInfo.DeviceID != guestCreds.DeviceID {
		return fmt.Errorf("device ID mismatch between credentials and device info")
	}

	return nil
}
