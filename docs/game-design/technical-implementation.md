# 협력 타워 디펜스 기술적 구현 가이드

## 🏗️ 서버 아키텍처 연동

### 서버 역할 분담

#### 🛡️ GuardianApp (인증 서버)
```go
// 플레이어 인증 및 세션 관리
type AuthService struct {
    redis    *redis.Client
    jwtKey   []byte
}

// 게임 참여 전 인증 검증
func (s *AuthService) ValidateGameAccess(token string) (*Player, error)
```

#### 🏙️ TimeSquareApp (게임 서버)
```go
// 실시간 게임 로직 처리
type GameService struct {
    redis       *redis.Client
    pubsub      *redis.PubSub
    gameStates  map[string]*GameState
}

// 협력 액션 처리
func (s *GameService) ProcessCooperativeAction(gameID string, action *CoopAction) error
```

#### ⚡ CommandApp (운영 서버)
```go
// 게임 통계 및 모니터링
type StatsService struct {
    redis *redis.Client
}

// 협력 성과 분석
func (s *StatsService) AnalyzeTeamPerformance(gameID string) *TeamStats
```

## 📊 Redis 데이터 구조

### 게임 상태 관리
```redis
# 게임 세션 기본 정보
game:session:{gameId} = {
    "id": "game_12345",
    "status": "playing",
    "difficultyLevel": 45,
    "players": ["player1", "player2", "player3", "player4"],
    "playerRaces": {
        "player1": "human_alliance",
        "player2": "elven_kingdom",
        "player3": "dwarven_clan",
        "player4": "dragon_clan"
    },
    "environment": {
        "timeOfDay": "dusk",
        "weather": "storm",
        "terrain": "mountains",
        "magicLevel": "high",
        "specialEvents": ["meteor_shower"]
    },
    "currentWave": 5,
    "baseHealth": 80,
    "startTime": "2024-01-01T10:00:00Z",
    "teamResources": {
        "gold": 1500,
        "crystal": 200,
        "teamPoints": 50
    },
    "raceBuffs": {
        "human_alliance": {"cooperation_bonus": 1.2, "resource_efficiency": 1.15},
        "elven_kingdom": {"range_bonus": 1.2, "accuracy_bonus": 1.15},
        "dwarven_clan": {"defense_bonus": 1.3, "explosion_bonus": 1.4},
        "dragon_clan": {"fire_bonus": 1.6, "boss_damage": 1.25}
    },
    "environmentalEffects": {
        "storm_bonus": {"mechanical_empire": {"electric_attack": 1.4}},
        "mountain_bonus": {"dwarven_clan": {"all_abilities": 1.3}},
        "dusk_bonus": {"dragon_clan": {"all_abilities": 1.1}}
    }
}

# 플레이어별 상태
game:player:{gameId}:{playerId} = {
    "id": "player1",
    "race": "human_alliance",
    "role": "tanker",
    "position": {"x": 100, "y": 200},
    "resources": {"gold": 500, "crystal": 50},
    "towers": ["tower1", "tower2"],
    "raceAbilities": {
        "cooperation_bonus": 1.2,
        "resource_efficiency": 1.15,
        "all_tower_access": true
    },
    "environmentalEffects": {
        "current_buffs": ["dusk_cooperation", "mountain_stability"],
        "current_debuffs": ["storm_visibility"]
    },
    "stats": {
        "damageDealt": 15000,
        "cooperationScore": 85,
        "resourcesShared": 300,
        "raceSpecificActions": 12
    }
}

# 타워 정보
game:towers:{gameId} = {
    "tower1": {
        "id": "tower1",
        "type": "knights_fortress",
        "race": "human_alliance",
        "level": 2,
        "position": {"x": 150, "y": 250},
        "owner": "player1",
        "connectedTowers": ["tower2"],
        "baseStats": {"damage": 100, "range": 150, "defense": 200},
        "raceModifiers": {
            "cooperation_bonus": 1.2,
            "defense_aura": {"radius": 100, "bonus": 1.15}
        },
        "environmentalModifiers": {
            "mountain_terrain": {"stability": 1.1},
            "storm_weather": {"accuracy": 0.9}
        },
        "finalStats": {"damage": 120, "range": 150, "defense": 242},
        "specialAbilities": ["defense_aura", "resource_generation"]
    },
    "tower2": {
        "id": "tower2",
        "type": "ancient_tree",
        "race": "elven_kingdom",
        "level": 3,
        "position": {"x": 200, "y": 300},
        "owner": "player2",
        "connectedTowers": ["tower1"],
        "baseStats": {"damage": 80, "range": 200, "growth": 1},
        "raceModifiers": {
            "range_bonus": 1.2,
            "nature_affinity": 1.25
        },
        "environmentalModifiers": {
            "mountain_terrain": {"build_efficiency": 0.85},
            "dusk_time": {"damage": 1.25}
        },
        "finalStats": {"damage": 125, "range": 240, "growth": 2},
        "specialAbilities": ["growth_over_time", "nature_magic"]
    }
}

# 웨이브 상태
game:wave:{gameId} = {
    "currentWave": 5,
    "enemies": [
        {
            "id": "enemy1",
            "type": "tank",
            "health": 500,
            "position": {"x": 50, "y": 100},
            "path": ["point1", "point2", "point3"]
        }
    ],
    "nextWaveTime": "2024-01-01T10:05:00Z"
}
```

### 실시간 이벤트 채널
```redis
# 게임 이벤트 채널
PUBLISH events:game:{gameId} {
    "type": "tower_placed",
    "playerId": "player1",
    "data": {
        "towerId": "tower1",
        "position": {"x": 150, "y": 250},
        "type": "archer"
    },
    "timestamp": "2024-01-01T10:00:30Z"
}

# 협력 이벤트 채널
PUBLISH events:coop:{gameId} {
    "type": "resource_shared",
    "fromPlayer": "player1",
    "toPlayer": "player2",
    "amount": 100,
    "resourceType": "gold"
}

# 시스템 이벤트 채널
PUBLISH events:system:{gameId} {
    "type": "wave_started",
    "waveNumber": 6,
    "enemyCount": 15,
    "specialEnemies": ["boss_tank"]
}
```

## 🔄 실시간 협력 시스템

### SSE 이벤트 스트리밍
```go
type SSEHandler struct {
    redis  *redis.Client
    pubsub *redis.PubSub
}

func (h *SSEHandler) StreamGameEvents(w http.ResponseWriter, r *http.Request) {
    gameID := r.URL.Query().Get("gameId")
    playerID := r.URL.Query().Get("playerId")

    // SSE 헤더 설정
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Redis Pub/Sub 구독
    channel := fmt.Sprintf("events:game:%s", gameID)
    subscription := h.pubsub.Subscribe(channel)

    for {
        select {
        case msg := <-subscription.Channel():
            // 이벤트 필터링 및 전송
            event := h.filterEventForPlayer(msg.Payload, playerID)
            if event != nil {
                fmt.Fprintf(w, "data: %s\n\n", event)
                w.(http.Flusher).Flush()
            }
        case <-r.Context().Done():
            subscription.Close()
            return
        }
    }
}
```

### 협력 액션 처리
```go
type CooperativeAction struct {
    Type       string                 `json:"type"`
    PlayerID   string                 `json:"playerId"`
    GameID     string                 `json:"gameId"`
    Data       map[string]interface{} `json:"data"`
    Timestamp  time.Time              `json:"timestamp"`
}

func (s *GameService) ProcessCooperativeAction(action *CooperativeAction) error {
    switch action.Type {
    case "resource_share":
        return s.handleResourceShare(action)
    case "tower_build_request":
        return s.handleTowerBuildRequest(action)
    case "emergency_support":
        return s.handleEmergencySupport(action)
    case "ping_signal":
        return s.handlePingSignal(action)
    default:
        return fmt.Errorf("unknown cooperative action: %s", action.Type)
    }
}

func (s *GameService) handleResourceShare(action *CooperativeAction) error {
    fromPlayer := action.PlayerID
    toPlayer := action.Data["toPlayer"].(string)
    amount := int(action.Data["amount"].(float64))
    resourceType := action.Data["resourceType"].(string)

    // 자원 이전 로직
    if err := s.transferResource(action.GameID, fromPlayer, toPlayer, resourceType, amount); err != nil {
        return err
    }

    // 협력 이벤트 발행
    event := map[string]interface{}{
        "type": "resource_shared",
        "fromPlayer": fromPlayer,
        "toPlayer": toPlayer,
        "amount": amount,
        "resourceType": resourceType,
    }

    return s.publishCoopEvent(action.GameID, event)
}
```

## 🎮 게임 로직 구현

### 타워 시너지 시스템
```go
type TowerSynergy struct {
    TowerIDs    []string               `json:"towerIds"`
    SynergyType string                 `json:"synergyType"`
    Bonus       map[string]interface{} `json:"bonus"`
}

func (s *GameService) CalculateTowerSynergies(gameID string) ([]TowerSynergy, error) {
    towers, err := s.getTowers(gameID)
    if err != nil {
        return nil, err
    }

    var synergies []TowerSynergy

    // 인접 타워 시너지 계산
    for _, tower := range towers {
        adjacentTowers := s.findAdjacentTowers(tower, towers)
        if len(adjacentTowers) > 0 {
            synergy := s.calculateAdjacencyBonus(tower, adjacentTowers)
            synergies = append(synergies, synergy)
        }
    }

    // 속성 시너지 계산
    attributeSynergies := s.calculateAttributeSynergies(towers)
    synergies = append(synergies, attributeSynergies...)

    return synergies, nil
}
```

### 동적 난이도 조절
```go
type DifficultyAdjuster struct {
    redis *redis.Client
}

func (d *DifficultyAdjuster) AdjustWaveDifficulty(gameID string, waveNumber int) (*WaveConfig, error) {
    // 팀 성과 분석
    teamPerf, err := d.analyzeTeamPerformance(gameID)
    if err != nil {
        return nil, err
    }

    baseConfig := d.getBaseWaveConfig(waveNumber)

    // 성과에 따른 난이도 조절
    if teamPerf.CooperationScore > 80 {
        // 협력이 뛰어난 팀에게는 더 도전적인 웨이브
        baseConfig.EnemyCount = int(float64(baseConfig.EnemyCount) * 1.2)
        baseConfig.SpecialEnemies = append(baseConfig.SpecialEnemies, "elite_unit")
    } else if teamPerf.CooperationScore < 40 {
        // 협력이 부족한 팀에게는 난이도 완화
        baseConfig.EnemyCount = int(float64(baseConfig.EnemyCount) * 0.8)
        baseConfig.ResourceBonus = int(float64(baseConfig.ResourceBonus) * 1.3)
    }

    return baseConfig, nil
}
```

## 📈 성과 측정 및 분석

### 협력 점수 계산
```go
type CooperationMetrics struct {
    ResourceSharing    int     `json:"resourceSharing"`
    MutualSupport     int     `json:"mutualSupport"`
    StrategicAlignment float64 `json:"strategicAlignment"`
    CommunicationScore int     `json:"communicationScore"`
}

func (s *StatsService) CalculateCooperationScore(gameID string, playerID string) (int, error) {
    metrics, err := s.getCooperationMetrics(gameID, playerID)
    if err != nil {
        return 0, err
    }

    // 가중치 적용 점수 계산
    score := 0
    score += metrics.ResourceSharing * 2        // 자원 공유 (가중치 2)
    score += metrics.MutualSupport * 3          // 상호 지원 (가중치 3)
    score += int(metrics.StrategicAlignment * 25) // 전략적 일치 (가중치 25)
    score += metrics.CommunicationScore * 1     // 소통 점수 (가중치 1)

    // 0-100 범위로 정규화
    maxScore := 100
    if score > maxScore {
        score = maxScore
    }

    return score, nil
}
```

### 실시간 통계 업데이트
```go
func (s *StatsService) UpdateRealTimeStats(gameID string, event *GameEvent) error {
    switch event.Type {
    case "tower_placed":
        return s.updateTowerStats(gameID, event)
    case "enemy_killed":
        return s.updateCombatStats(gameID, event)
    case "resource_shared":
        return s.updateCooperationStats(gameID, event)
    case "wave_completed":
        return s.updateWaveStats(gameID, event)
    }

    return nil
}

func (s *StatsService) updateCooperationStats(gameID string, event *GameEvent) error {
    key := fmt.Sprintf("stats:cooperation:%s", gameID)

    // Redis Hash를 사용한 통계 업데이트
    pipe := s.redis.Pipeline()
    pipe.HIncrBy(key, "total_shares", 1)
    pipe.HIncrBy(key, fmt.Sprintf("player:%s:shares_given", event.PlayerID), 1)
    pipe.HIncrBy(key, fmt.Sprintf("player:%s:shares_received", event.Data["toPlayer"]), 1)
    pipe.Expire(key, 24*time.Hour) // 24시간 후 만료

    _, err := pipe.Exec()
    return err
}
```

## 🔧 최적화 및 성능

### 게임 상태 압축
```go
type CompressedGameState struct {
    Version   int                    `json:"v"`
    Timestamp int64                  `json:"t"`
    Delta     map[string]interface{} `json:"d"` // 변경된 부분만 포함
}

func (s *GameService) CompressGameState(gameID string, fullState *GameState) (*CompressedGameState, error) {
    lastState, err := s.getLastGameState(gameID)
    if err != nil {
        return nil, err
    }

    delta := s.calculateStateDelta(lastState, fullState)

    compressed := &CompressedGameState{
        Version:   fullState.Version,
        Timestamp: time.Now().Unix(),
        Delta:     delta,
    }

    return compressed, nil
}
```

### 이벤트 배치 처리
```go
func (s *GameService) ProcessEventBatch(gameID string, events []*GameEvent) error {
    // 이벤트를 타입별로 그룹화
    eventGroups := s.groupEventsByType(events)

    // Redis 파이프라인을 사용한 배치 처리
    pipe := s.redis.Pipeline()

    for eventType, eventList := range eventGroups {
        switch eventType {
        case "tower_action":
            s.batchProcessTowerActions(pipe, gameID, eventList)
        case "resource_action":
            s.batchProcessResourceActions(pipe, gameID, eventList)
        case "combat_action":
            s.batchProcessCombatActions(pipe, gameID, eventList)
        }
    }

    _, err := pipe.Exec()
    return err
}
```

---

이 기술적 구현 가이드는 협력 기반 실시간 타워 디펜스 게임의 서버 사이드 구현을 위한 상세한 방향을 제시합니다.
