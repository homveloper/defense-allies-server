package cqrsx

import (
	"cqrs"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransportCreatedEventData - 문서에서 제공된 이벤트 스키마1
type TransportCreatedEventData struct {
	StartedAt time.Time `bson:"startedAt" json:"startedAt"`
}

type TransportCreatedEventMessage struct {
	*cqrs.BaseEventMessage
	TransportCreatedEventData `bson:",inline" json:",inline"`
}

func (e TransportCreatedEventMessage) EventType() string {
	return "TransportCreated"
}

func (e TransportCreatedEventMessage) EventData() interface{} {
	return e
}

// 복잡한 중첩 구조를 가진 이벤트 (bson.D 문제를 재현하기 위함)
type ComplexNestedEventData struct {
	PlayerInfo map[string]interface{}   `bson:"playerInfo" json:"playerInfo"`
	GameConfig map[string]interface{}   `bson:"gameConfig" json:"gameConfig"`
	Inventory  []map[string]interface{} `bson:"inventory" json:"inventory"`
	EventMeta  map[string]interface{}   `bson:"eventMeta" json:"eventMeta"`
	Timestamps map[string]time.Time     `bson:"timestamps" json:"timestamps"`
}

type ComplexNestedEventMessage struct {
	*cqrs.BaseEventMessage
	ComplexNestedEventData `bson:",inline" json:",inline"`
}

func (e ComplexNestedEventMessage) EventType() string {
	return "ComplexNestedEvent"
}

func (e ComplexNestedEventMessage) EventData() interface{} {
	return e.ComplexNestedEventData
}

// 테스트용 레지스트리 설정
func setupBsonDTestRegistry() EventRegistry {
	registry := NewInMemoryEventRegistry()

	// TransportCreatedEventMessage 등록
	err := registry.RegisterDataStruct("TransportCreated", &TransportCreatedEventMessage{})
	if err != nil {
		panic(err)
	}

	// ComplexNestedEventMessage 등록
	err = registry.RegisterDataStruct("ComplexNestedEvent", &ComplexNestedEventMessage{})
	if err != nil {
		panic(err)
	}

	return registry
}

// 복잡한 중첩 데이터 생성 (bson.D 문제를 유발할 수 있는 구조)
func createComplexNestedEventData() ComplexNestedEventData {
	return ComplexNestedEventData{
		PlayerInfo: map[string]interface{}{
			"playerID": "player123",
			"level":    50,
			"stats": map[string]interface{}{
				"attack":  100,
				"defense": 80,
				"skills": []interface{}{
					map[string]interface{}{
						"name":     "Fireball",
						"level":    5,
						"cooldown": 3.5,
					},
					map[string]interface{}{
						"name":    "Shield",
						"level":   3,
						"effects": []string{"defense", "immunity"},
					},
				},
			},
		},
		GameConfig: map[string]interface{}{
			"difficulty": "hard",
			"rules": map[string]interface{}{
				"pvp":       true,
				"timeLimit": 3600,
				"rewards": map[string]interface{}{
					"experience": 1000,
					"items": []interface{}{
						map[string]interface{}{
							"itemID":   "sword001",
							"quantity": 1,
							"properties": map[string]interface{}{
								"damage":       150,
								"durability":   100,
								"enchantments": []string{"fire", "sharpness"},
							},
						},
					},
				},
			},
		},
		Inventory: []map[string]interface{}{
			{
				"slotID": 1,
				"item": map[string]interface{}{
					"id":   "potion001",
					"name": "Health Potion",
					"effects": []interface{}{
						map[string]interface{}{
							"type":  "heal",
							"value": 50,
						},
					},
				},
			},
			{
				"slotID": 2,
				"item": map[string]interface{}{
					"id":   "armor001",
					"name": "Steel Armor",
					"stats": map[string]interface{}{
						"defense": 25,
						"weight":  10.5,
					},
				},
			},
		},
		EventMeta: map[string]interface{}{
			"source":  "game_server",
			"version": "1.2.3",
			"debug": map[string]interface{}{
				"trace_id":   "abc123",
				"session_id": "session456",
				"performance": map[string]interface{}{
					"processing_time_ms": 123.45,
					"memory_usage_mb":    256.78,
				},
			},
		},
		Timestamps: map[string]time.Time{
			"created_at": time.Now().UTC(),
			"updated_at": time.Now().UTC().Add(5 * time.Minute),
		},
	}
}

// 테스트 1: 기본 이벤트에서 bson.D 문제가 발생하지 않는지 확인
func TestUnmarshalEventBSON_SimpleEvent_NoBsonDProblem(t *testing.T) {
	registry := setupBsonDTestRegistry()

	// 1. 간단한 이벤트 생성
	originalEvent := TransportCreatedEventMessage{
		BaseEventMessage: cqrs.NewBaseEventMessage("TransportCreated"),
		TransportCreatedEventData: TransportCreatedEventData{
			StartedAt: time.Now().UTC(),
		},
	}

	// 2. BSON으로 마샬링
	bsonData, err := MarshalEventBSON(originalEvent)
	require.NoError(t, err, "Failed to marshal event to BSON")

	// 3. CreateDataInstance로 강타입 인스턴스 생성
	instance, err := registry.CreateDataInstance("TransportCreated")
	require.NoError(t, err, "Failed to create data instance")

	// 타입 확인
	assert.IsType(t, &TransportCreatedEventMessage{}, instance, "Created instance should be *TransportCreatedEventMessage")

	// 4. 직접 bson.Unmarshal 테스트 (이 과정에서 문제가 발생할 수 있음)
	err = bson.Unmarshal(bsonData, &instance)
	require.NoError(t, err, "Direct bson.Unmarshal should not fail")

	// 5. 타입이 여전히 올바른지 확인
	transportEvent, ok := instance.(*TransportCreatedEventMessage)
	require.True(t, ok, "Instance should still be *TransportCreatedEventMessage after unmarshal")

	// 6. 데이터가 올바르게 복원되었는지 확인
	assert.Equal(t, "TransportCreated", transportEvent.EventType())
	assert.WithinDuration(t, originalEvent.StartedAt, transportEvent.StartedAt, time.Second)

	// 7. UnmarshalEventBSON 함수로도 테스트
	unmarshaledEvent, err := UnmarshalEventBSON(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON should not fail")

	transportEvent2, ok := unmarshaledEvent.(*TransportCreatedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *TransportCreatedEventMessage")
	assert.Equal(t, "TransportCreated", transportEvent2.EventType())
}

// 테스트 2: 복잡한 중첩 구조에서 bson.D 문제 재현 및 감지
func TestUnmarshalEventBSON_ComplexEvent_BsonDProblemDetection(t *testing.T) {
	registry := setupBsonDTestRegistry()

	// 1. 복잡한 중첩 구조의 이벤트 생성
	originalEvent := ComplexNestedEventMessage{
		BaseEventMessage:       cqrs.NewBaseEventMessage("ComplexNestedEvent"),
		ComplexNestedEventData: createComplexNestedEventData(),
	}

	// 2. BSON으로 마샬링
	bsonData, err := MarshalEventBSON(originalEvent)
	require.NoError(t, err, "Failed to marshal complex event to BSON")

	t.Logf("BSON data size: %d bytes", len(bsonData))

	// 3. 원시 BSON 데이터 검사
	var rawData map[string]interface{}
	err = bson.Unmarshal(bsonData, &rawData)
	require.NoError(t, err, "Failed to unmarshal BSON to raw map")

	t.Logf("Raw BSON keys: %v", getMapKeys(rawData))

	// 4. CreateDataInstance로 강타입 인스턴스 생성
	instance, err := registry.CreateDataInstance("ComplexNestedEvent")
	require.NoError(t, err, "Failed to create data instance")

	// 타입 확인
	assert.IsType(t, &ComplexNestedEventMessage{}, instance, "Created instance should be *ComplexNestedEventMessage")

	// 5. 직접 bson.Unmarshal 테스트 - 여기서 bson.D 문제가 발생할 수 있음
	err = bson.Unmarshal(bsonData, &instance)
	require.NoError(t, err, "Direct bson.Unmarshal should not fail")

	// 6. bson.D 타입 변환 문제 감지
	t.Logf("After unmarshal - instance type: %T", instance)
	complexEvent, ok := instance.(*ComplexNestedEventMessage)
	if !ok {
		t.Logf("Failed to cast to *ComplexNestedEventMessage. Actual type: %T", instance)
		t.Logf("Instance value: %+v", instance)
		t.FailNow()
	}

	// PlayerInfo 필드에서 bson.D 문제 확인
	t.Logf("PlayerInfo type: %T", complexEvent.PlayerInfo)
	checkForBsonDInMap(t, "PlayerInfo", complexEvent.PlayerInfo)

	// GameConfig 필드에서 bson.D 문제 확인
	t.Logf("GameConfig type: %T", complexEvent.GameConfig)
	checkForBsonDInMap(t, "GameConfig", complexEvent.GameConfig)

	// Inventory 필드에서 bson.D 문제 확인
	t.Logf("Inventory length: %d", len(complexEvent.Inventory))
	for i, item := range complexEvent.Inventory {
		t.Logf("Inventory[%d] type: %T", i, item)
		checkForBsonDInMap(t, fmt.Sprintf("Inventory[%d]", i), item)
	}

	// Metadata 필드에서 bson.D 문제 확인
	t.Logf("EventMeta type: %T", complexEvent.EventMeta)
	checkForBsonDInMap(t, "EventMeta", complexEvent.EventMeta)

	// 7. UnmarshalEventBSON 함수로도 테스트
	unmarshaledEvent, err := UnmarshalEventBSON(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON should not fail")

	complexEvent2, ok := unmarshaledEvent.(*ComplexNestedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *ComplexNestedEventMessage")

	// UnmarshalEventBSON으로 생성된 인스턴스에서도 bson.D 문제 확인
	t.Log("=== Checking UnmarshalEventBSON result ===")
	checkForBsonDInMap(t, "UnmarshalEventBSON.PlayerInfo", complexEvent2.PlayerInfo)
	checkForBsonDInMap(t, "UnmarshalEventBSON.GameConfig", complexEvent2.GameConfig)
	checkForBsonDInMap(t, "UnmarshalEventBSON.EventMeta", complexEvent2.EventMeta)
}

// 테스트 3: 순수 bson.D 문제 재현 테스트
func TestBsonD_DirectReproduction(t *testing.T) {
	// 1. 중첩된 맵 구조 생성
	complexData := map[string]interface{}{
		"eventType": "ComplexNestedEvent",
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "deep_value",
				"array": []interface{}{
					map[string]interface{}{
						"item": "value1",
						"nested": map[string]interface{}{
							"property": "nested_value",
						},
					},
				},
			},
		},
	}

	// 2. BSON으로 마샬링
	bsonData, err := bson.Marshal(complexData)
	require.NoError(t, err, "Failed to marshal complex data")

	// 3. 다시 맵으로 언마샬링 - 여기서 bson.D 변환 발생
	var unmarshaled map[string]interface{}
	err = bson.Unmarshal(bsonData, &unmarshaled)
	require.NoError(t, err, "Failed to unmarshal to map")

	// 4. bson.D 타입 감지
	t.Log("=== Direct bson.D reproduction test ===")
	checkForBsonDInMap(t, "DirectTest", unmarshaled)
}

// 테스트 4: Registry를 통한 강타입 인스턴스 생성 후 언마샬링에서 bson.D 문제
func TestRegistry_CreateDataInstance_BsonD_Problem(t *testing.T) {
	registry := setupBsonDTestRegistry()

	// 1. 복잡한 BSON 데이터 직접 생성
	bsonData := createComplexBSONData(t)

	// 2. Registry를 통해 강타입 인스턴스 생성
	instance, err := registry.CreateDataInstance("ComplexNestedEvent")
	require.NoError(t, err, "Failed to create data instance")

	t.Logf("Created instance type: %T", instance)

	// 3. 생성된 인스턴스에 BSON 데이터 언마샬링
	err = bson.Unmarshal(bsonData, &instance)
	require.NoError(t, err, "Failed to unmarshal BSON to typed instance")

	// 4. 인스턴스 타입이 여전히 올바른지 확인
	complexEvent, ok := instance.(*ComplexNestedEventMessage)
	require.True(t, ok, "Instance should be *ComplexNestedEventMessage")

	// 5. bson.D 문제 확인
	t.Log("=== Registry instance bson.D check ===")
	checkForBsonDInMap(t, "Registry.PlayerInfo", complexEvent.PlayerInfo)
	checkForBsonDInMap(t, "Registry.GameConfig", complexEvent.GameConfig)
	checkForBsonDInMap(t, "Registry.EventMeta", complexEvent.EventMeta)
}

// 헬퍼 함수들

// checkForBsonDInMap은 맵에서 bson.D 타입을 재귀적으로 찾아서 테스트 실패시킴
func checkForBsonDInMap(t *testing.T, fieldPath string, data interface{}) {
	switch v := data.(type) {
	case primitive.D:
		t.Errorf("BSON.D PROBLEM DETECTED: %s contains bson.D type: %+v", fieldPath, v)
	case map[string]interface{}:
		for key, value := range v {
			checkForBsonDInMap(t, fmt.Sprintf("%s.%s", fieldPath, key), value)
		}
	case []interface{}:
		for i, item := range v {
			checkForBsonDInMap(t, fmt.Sprintf("%s[%d]", fieldPath, i), item)
		}
	case []map[string]interface{}:
		for i, item := range v {
			checkForBsonDInMap(t, fmt.Sprintf("%s[%d]", fieldPath, i), item)
		}
	default:
		// 다른 타입은 무시 (primitive 타입, time.Time 등)
	}
}

// getMapKeys는 맵의 키들을 반환 (디버깅용)
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// createComplexBSONData는 복잡한 중첩 구조의 BSON 데이터를 생성
func createComplexBSONData(t *testing.T) []byte {
	complexData := map[string]interface{}{
		"eventType": "ComplexNestedEvent",
		"eventId":   "test-event-id",
		"id":        primitive.NewObjectID(),
		"playerInfo": map[string]interface{}{
			"playerID": "player123",
			"stats": map[string]interface{}{
				"level": 50,
				"skills": []interface{}{
					map[string]interface{}{
						"name": "Fireball",
						"properties": map[string]interface{}{
							"damage": 100,
							"type":   "fire",
						},
					},
				},
			},
		},
		"gameConfig": map[string]interface{}{
			"rules": map[string]interface{}{
				"pvp": true,
				"rewards": map[string]interface{}{
					"items": []interface{}{
						map[string]interface{}{
							"id": "item001",
							"properties": map[string]interface{}{
								"damage": 150,
							},
						},
					},
				},
			},
		},
		"eventMeta": map[string]interface{}{
			"debug": map[string]interface{}{
				"performance": map[string]interface{}{
					"processing_time_ms": 123.45,
				},
			},
		},
	}

	bsonData, err := bson.Marshal(complexData)
	require.NoError(t, err, "Failed to create complex BSON data")

	return bsonData
}

// 테스트 5: 문제 해결 방안 검증 테스트
func TestUnmarshalEventBSON_Solution_Verification(t *testing.T) {
	registry := setupBsonDTestRegistry()

	// 1. 복잡한 이벤트 생성 및 마샬링
	originalEvent := ComplexNestedEventMessage{
		BaseEventMessage:       cqrs.NewBaseEventMessage("ComplexNestedEvent"),
		ComplexNestedEventData: createComplexNestedEventData(),
	}

	bsonData, err := MarshalEventBSON(originalEvent)
	require.NoError(t, err, "Failed to marshal event")

	// 2. 현재 UnmarshalEventBSON 함수 테스트
	unmarshaledEvent, err := UnmarshalEventBSON(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON failed")

	complexEvent, ok := unmarshaledEvent.(*ComplexNestedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *ComplexNestedEventMessage")

	// 3. 현재 구현에서 bson.D 문제가 있는지 확인
	hasBsonDProblem := false
	defer func() {
		if hasBsonDProblem {
			t.Log("현재 구현에서 bson.D 문제가 발견되었습니다.")
			t.Log("해결 방안:")
			t.Log("1. bson.Unmarshal 전에 BSON 데이터를 map[string]interface{}로 먼저 언마샬링")
			t.Log("2. 그 다음 JSON을 통해 강타입 구조체로 변환")
			t.Log("3. 또는 bson 태그에서 bson.D 변환을 방지하는 옵션 사용")
		} else {
			t.Log("현재 구현에서 bson.D 문제가 발견되지 않았습니다.")
		}
	}()

	// bson.D 문제 검사
	if containsBsonD(complexEvent.PlayerInfo) {
		hasBsonDProblem = true
		t.Error("PlayerInfo contains bson.D")
	}

	if containsBsonD(complexEvent.GameConfig) {
		hasBsonDProblem = true
		t.Error("GameConfig contains bson.D")
	}

	if containsBsonD(complexEvent.EventMeta) {
		hasBsonDProblem = true
		t.Error("EventMeta contains bson.D")
	}

	for i, item := range complexEvent.Inventory {
		if containsBsonD(item) {
			hasBsonDProblem = true
			t.Errorf("Inventory[%d] contains bson.D", i)
		}
	}
}

// containsBsonD는 데이터 구조에 bson.D가 포함되어 있는지 확인
func containsBsonD(data interface{}) bool {
	switch v := data.(type) {
	case primitive.D:
		return true
	case map[string]interface{}:
		for _, value := range v {
			if containsBsonD(value) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if containsBsonD(item) {
				return true
			}
		}
	case []map[string]interface{}:
		for _, item := range v {
			if containsBsonD(item) {
				return true
			}
		}
	}
	return false
}
