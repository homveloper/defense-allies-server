// versioning/event_upgrader.go - 이벤트 버전 업그레이드
package cqrsx

import (
	"fmt"
	"time"
)

// EventUpgrader는 이벤트 버전 업그레이드를 담당합니다
type EventUpgrader struct {
	upgraders map[string]map[int]UpgradeFunc
}

// UpgradeFunc는 이벤트 업그레이드 함수 타입입니다
type UpgradeFunc func(eventData map[string]interface{}) (map[string]interface{}, error)

// VersionedEvent는 버전 정보를 포함한 이벤트입니다
type VersionedEvent struct {
	EventType string                 `json:"eventType"`
	Version   int                    `json:"version"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// NewEventUpgrader는 새로운 이벤트 업그레이더를 생성합니다
func NewEventUpgrader() *EventUpgrader {
	upgrader := &EventUpgrader{
		upgraders: make(map[string]map[int]UpgradeFunc),
	}

	// 기본 업그레이더들 등록
	upgrader.registerDefaultUpgraders()

	return upgrader
}

// RegisterUpgrader는 특정 이벤트 타입의 업그레이더를 등록합니다
func (eu *EventUpgrader) RegisterUpgrader(eventType string, fromVersion int, upgradeFunc UpgradeFunc) {
	if eu.upgraders[eventType] == nil {
		eu.upgraders[eventType] = make(map[int]UpgradeFunc)
	}
	eu.upgraders[eventType][fromVersion] = upgradeFunc
}

// UpgradeEvent는 이벤트를 최신 버전으로 업그레이드합니다
func (eu *EventUpgrader) UpgradeEvent(eventType string, version int, eventData map[string]interface{}) (map[string]interface{}, int, error) {
	currentData := eventData
	currentVersion := version

	// 최신 버전까지 순차적으로 업그레이드
	for {
		upgradeFunc, exists := eu.upgraders[eventType][currentVersion]
		if !exists {
			// 더 이상 업그레이드할 버전이 없음
			break
		}

		upgradedData, err := upgradeFunc(currentData)
		if err != nil {
			return nil, currentVersion, fmt.Errorf("failed to upgrade event from version %d: %w", currentVersion, err)
		}

		currentData = upgradedData
		currentVersion++
	}

	return currentData, currentVersion, nil
}

// 기본 업그레이더들 등록
func (eu *EventUpgrader) registerDefaultUpgraders() {
	// GuildCreatedEvent v1 -> v2 업그레이드
	eu.RegisterUpgrader("GuildCreated", 1, func(data map[string]interface{}) (map[string]interface{}, error) {
		// v2에서 추가된 필드들
		data["region"] = "global"      // 기본 지역
		data["guildType"] = "standard" // 기본 길드 타입

		// 날짜 형식 변경 (문자열 -> 타임스탬프)
		if createdAtStr, ok := data["createdAt"].(string); ok {
			if timestamp, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				data["createdAt"] = timestamp.Unix()
			}
		}

		return data, nil
	})

	// MemberJoinedEvent v1 -> v2 업그레이드
	eu.RegisterUpgrader("MemberJoined", 1, func(data map[string]interface{}) (map[string]interface{}, error) {
		// v2에서 추가된 필드들
		data["invitationMethod"] = "direct" // 기본 초대 방법
		data["joinSource"] = "web"          // 기본 가입 소스

		// 권한 시스템 변경
		if role, ok := data["role"].(string); ok {
			data["permissions"] = getRolePermissions(role)
		}

		return data, nil
	})

	// MemberJoinedEvent v2 -> v3 업그레이드
	eu.RegisterUpgrader("MemberJoined", 2, func(data map[string]interface{}) (map[string]interface{}, error) {
		// v3에서 멤버 상태 추가
		data["memberStatus"] = "active"
		data["lastActivity"] = time.Now().Unix()

		// 초대 정보 구조 변경
		if invitedBy, ok := data["invitedBy"].(string); ok {
			data["invitation"] = map[string]interface{}{
				"invitedBy": invitedBy,
				"method":    data["invitationMethod"],
				"timestamp": data["joinedAt"],
			}
			delete(data, "invitedBy")
			delete(data, "invitationMethod")
		}

		return data, nil
	})

	// ResourceSharedEvent v1 -> v2 업그레이드
	eu.RegisterUpgrader("ResourceShared", 1, func(data map[string]interface{}) (map[string]interface{}, error) {
		// v2에서 리소스 카테고리 추가
		resourceType, ok := data["resourceType"].(string)
		if !ok {
			return data, fmt.Errorf("resourceType not found")
		}

		data["category"] = getResourceCategory(resourceType)
		data["rarity"] = getResourceRarity(resourceType)

		// 기여도 포인트 계산 추가
		if amount, ok := data["amount"].(float64); ok {
			data["contributionPoints"] = calculateContributionPoints(resourceType, int(amount))
		}

		return data, nil
	})
}

// 헬퍼 함수들
func getRolePermissions(role string) []string {
	rolePermissions := map[string][]string{
		"member":       {"view_guild", "share_resources"},
		"officer":      {"view_guild", "share_resources", "invite_members", "manage_resources"},
		"guild_master": {"view_guild", "share_resources", "invite_members", "manage_resources", "manage_guild", "kick_members"},
	}

	return rolePermissions[role]
}

func getResourceCategory(resourceType string) string {
	categories := map[string]string{
		"gold":   "currency",
		"silver": "currency",
		"wood":   "material",
		"stone":  "material",
		"iron":   "material",
		"food":   "consumable",
		"potion": "consumable",
	}

	if category, exists := categories[resourceType]; exists {
		return category
	}
	return "misc"
}

func getResourceRarity(resourceType string) string {
	rarities := map[string]string{
		"gold":   "rare",
		"silver": "common",
		"wood":   "common",
		"stone":  "common",
		"iron":   "uncommon",
		"food":   "common",
		"potion": "uncommon",
	}

	if rarity, exists := rarities[resourceType]; exists {
		return rarity
	}
	return "common"
}

func calculateContributionPoints(resourceType string, amount int) int {
	multipliers := map[string]int{
		"gold":   10,
		"silver": 1,
		"wood":   1,
		"stone":  1,
		"iron":   3,
		"food":   1,
		"potion": 5,
	}

	if multiplier, exists := multipliers[resourceType]; exists {
		return amount * multiplier
	}
	return amount
}
