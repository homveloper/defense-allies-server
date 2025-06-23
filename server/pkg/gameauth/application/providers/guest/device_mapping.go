package guest

import (
	"crypto/sha256"
	"fmt"
	
	"github.com/google/uuid"
)

// DeviceIDToGameAccountID는 DeviceID를 deterministic하게 UUID 기반 GameAccountID로 변환합니다
// 같은 DeviceID는 항상 같은 GameAccountID를 반환합니다
func DeviceIDToGameAccountID(deviceID string) (string, error) {
	if deviceID == "" {
		return "", fmt.Errorf("device ID cannot be empty")
	}

	// DeviceID를 해시하여 고정된 seed 생성
	// "guest:" 프리픽스를 추가하여 다른 provider와 구분
	hash := sha256.Sum256([]byte(fmt.Sprintf("guest:%s", deviceID)))
	
	// 해시의 처음 16바이트를 사용하여 deterministic UUID 생성
	gameAccountUUID, err := uuid.FromBytes(hash[:16])
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID from device ID: %w", err)
	}
	
	return gameAccountUUID.String(), nil
}

// GameAccountIDToDeviceMapping은 GameAccountID가 특정 DeviceID에서 생성되었는지 검증합니다
func GameAccountIDToDeviceMapping(gameAccountID string, deviceID string) bool {
	expectedGameAccountID, err := DeviceIDToGameAccountID(deviceID)
	if err != nil {
		return false
	}
	
	return gameAccountID == expectedGameAccountID
}

// IsValidGuestGameAccountID는 GameAccountID가 유효한 게스트 계정 ID인지 확인합니다
func IsValidGuestGameAccountID(gameAccountID string) bool {
	// UUID 형식인지 검증
	_, err := uuid.Parse(gameAccountID)
	return err == nil
}