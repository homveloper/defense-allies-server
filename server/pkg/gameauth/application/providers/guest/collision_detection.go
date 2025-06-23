package guest

import (
	"context"
	"fmt"

	"defense-allies-server/pkg/gameauth/domain/common"
)

// validateGameAccountIDUniqueness는 GameAccountID 충돌을 검사합니다
func (p *Provider) validateGameAccountIDUniqueness(ctx context.Context, gameAccountID string, deviceID string) error {
	// 1. 같은 GameAccountID를 가진 다른 AccountLink가 있는지 확인
	existingLinkByGameID, err := p.accountLinkRepo.FindByGameAccountID(ctx, gameAccountID)
	if err != nil {
		return fmt.Errorf("failed to check existing game account: %w", err)
	}

	if existingLinkByGameID != nil {
		// 기존 계정이 있다면, 같은 DeviceID인지 확인
		if provider, exists := existingLinkByGameID.GetProvider(common.ProviderTypeGuest); exists {
			if provider.ExternalID != deviceID {
				// 다른 DeviceID가 같은 GameAccountID를 생성했다면 충돌!
				return fmt.Errorf("hash collision detected: GameAccountID %s already exists for different DeviceID", gameAccountID)
			}
		}
	}

	// 2. DeviceID로 생성된 GameAccountID가 예상값과 일치하는지 확인
	expectedGameAccountID, err := DeviceIDToGameAccountID(deviceID)
	if err != nil {
		return fmt.Errorf("failed to generate expected game account ID: %w", err)
	}

	if gameAccountID != expectedGameAccountID {
		return fmt.Errorf("game account ID mismatch: expected %s, got %s", expectedGameAccountID, gameAccountID)
	}

	return nil
}

// HandleCollision은 충돌이 발생했을 때의 복구 로직입니다
func (p *Provider) HandleCollision(ctx context.Context, deviceID string, collisionGameAccountID string) (string, error) {
	// 충돌 발생 시 대안 전략:
	// 1. DeviceID에 타임스탬프 추가하여 새로운 GameAccountID 생성
	// 2. 또는 순차적 suffix 추가
	
	// 예시: DeviceID + collision counter
	for i := 1; i <= 10; i++ {
		alternativeDeviceID := fmt.Sprintf("%s_collision_%d", deviceID, i)
		alternativeGameAccountID, err := DeviceIDToGameAccountID(alternativeDeviceID)
		if err != nil {
			continue
		}

		// 새로운 GameAccountID가 충돌하지 않는지 확인
		if err := p.validateGameAccountIDUniqueness(ctx, alternativeGameAccountID, alternativeDeviceID); err == nil {
			return alternativeGameAccountID, nil
		}
	}

	return "", fmt.Errorf("failed to resolve collision after 10 attempts for DeviceID: %s", deviceID)
}