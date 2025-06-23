package gamedata

import "sync"

// Repository 게임 데이터 레포지토리 인터페이스
type Repository interface {
	GetNewUserDefaults() *NewUserDefaults
	GetResourceInfo(resourceType string) *ResourceInfo
	GetAllResources() map[string]*ResourceInfo
	GetLeaderboardConfig() *LeaderboardConfig
	GetSessionConfig() *SessionConfig
}

// NewUserDefaults 신규 유저 기본값
type NewUserDefaults struct {
	Level     int
	Score     int64
	Resources map[string]int64
	Settings  map[string]string
}

// ResourceInfo 자원 정보
type ResourceInfo struct {
	Type         string
	DisplayName  string
	Description  string
	DefaultValue int64
	MaxValue     int64
	Category     string // currency, material, consumable 등
}

// LeaderboardConfig 리더보드 설정
type LeaderboardConfig struct {
	MaxEntries      int
	RefreshInterval int // seconds
	DurationDays    int // 활성 유저 기간
	Categories      []string
}

// SessionConfig 게임 세션 설정
type SessionConfig struct {
	MaxDuration      int // seconds
	IdleTimeout      int // seconds
	MaxConcurrent    int // 동시 세션 수
	HeartbeatInterval int // seconds
}

// InMemoryRepository 인메모리 게임 데이터 레포지토리
type InMemoryRepository struct {
	mu sync.RWMutex
}

// NewInMemoryRepository 새로운 인메모리 레포지토리 생성
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{}
}

// GetNewUserDefaults 신규 유저 기본값 반환
func (r *InMemoryRepository) GetNewUserDefaults() *NewUserDefaults {
	return &NewUserDefaults{
		Level: 1,
		Score: 0,
		Resources: map[string]int64{
			"gold":    1000,
			"crystal": 100,
			"energy":  50,
			"stamina": 100,
		},
		Settings: map[string]string{
			"music_volume": "0.7",
			"sfx_volume":   "0.8",
			"language":     "en",
			"tutorial":     "incomplete",
		},
	}
}

// GetResourceInfo 특정 자원 정보 반환
func (r *InMemoryRepository) GetResourceInfo(resourceType string) *ResourceInfo {
	resources := r.getAllResourcesMap()
	return resources[resourceType]
}

// GetAllResources 모든 자원 정보 반환
func (r *InMemoryRepository) GetAllResources() map[string]*ResourceInfo {
	return r.getAllResourcesMap()
}

// GetLeaderboardConfig 리더보드 설정 반환
func (r *InMemoryRepository) GetLeaderboardConfig() *LeaderboardConfig {
	return &LeaderboardConfig{
		MaxEntries:      100,
		RefreshInterval: 300, // 5분
		DurationDays:    30,  // 최근 30일 활성 유저
		Categories: []string{
			"score",
			"level",
			"tower_defense",
			"pvp",
		},
	}
}

// GetSessionConfig 세션 설정 반환
func (r *InMemoryRepository) GetSessionConfig() *SessionConfig {
	return &SessionConfig{
		MaxDuration:       3600, // 1시간
		IdleTimeout:       300,  // 5분
		MaxConcurrent:     1,    // 동시에 1개 세션만
		HeartbeatInterval: 30,   // 30초
	}
}

// getAllResourcesMap 모든 자원 정보 맵 (내부 헬퍼)
func (r *InMemoryRepository) getAllResourcesMap() map[string]*ResourceInfo {
	return map[string]*ResourceInfo{
		"gold": {
			Type:         "gold",
			DisplayName:  "Gold",
			Description:  "Basic currency for purchasing towers and upgrades",
			DefaultValue: 1000,
			MaxValue:     999999999,
			Category:     "currency",
		},
		"crystal": {
			Type:         "crystal",
			DisplayName:  "Crystal",
			Description:  "Premium currency for special items and boosts",
			DefaultValue: 100,
			MaxValue:     999999,
			Category:     "currency",
		},
		"energy": {
			Type:         "energy",
			DisplayName:  "Energy",
			Description:  "Required to start missions",
			DefaultValue: 50,
			MaxValue:     200,
			Category:     "consumable",
		},
		"stamina": {
			Type:         "stamina",
			DisplayName:  "Stamina",
			Description:  "Required for PvP battles",
			DefaultValue: 100,
			MaxValue:     100,
			Category:     "consumable",
		},
		"wood": {
			Type:         "wood",
			DisplayName:  "Wood",
			Description:  "Basic material for tower construction",
			DefaultValue: 0,
			MaxValue:     99999,
			Category:     "material",
		},
		"stone": {
			Type:         "stone",
			DisplayName:  "Stone",
			Description:  "Advanced material for tower upgrades",
			DefaultValue: 0,
			MaxValue:     99999,
			Category:     "material",
		},
		"iron": {
			Type:         "iron",
			DisplayName:  "Iron",
			Description:  "Rare material for special towers",
			DefaultValue: 0,
			MaxValue:     9999,
			Category:     "material",
		},
	}
}

// 싱글톤 인스턴스
var defaultRepository Repository = NewInMemoryRepository()

// GetRepository 기본 레포지토리 반환
func GetRepository() Repository {
	return defaultRepository
}