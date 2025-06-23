package cqrsx

import (
	"cqrs"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 문제 해결을 위한 개선된 UnmarshalEventBSON 함수
func UnmarshalEventBSON_Fixed(data []byte, registry EventRegistry) (cqrs.EventMessage, error) {
	// 1. eventType 필드 추출
	var typeExtractor struct {
		EventType string `bson:"eventType"`
	}
	if err := bson.Unmarshal(data, &typeExtractor); err != nil {
		return nil, fmt.Errorf("failed to extract eventType from BSON: %w", err)
	}
	if typeExtractor.EventType == "" {
		return nil, fmt.Errorf("eventType is missing in the BSON data")
	}

	// 2. 레지스트리를 사용하여 인스턴스 생성
	instance, err := registry.CreateDataInstance(typeExtractor.EventType)
	if err != nil {
		return nil, err
	}

	// 3. 해결책: BSON을 먼저 map[string]interface{}로 언마샬링
	var rawData map[string]interface{}
	if err := bson.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BSON to raw map: %w", err)
	}

	// 4. JSON을 통해 강타입 구조체로 변환 (bson.D 문제 우회)
	jsonBytes, err := json.Marshal(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw data to JSON: %w", err)
	}

	if err := json.Unmarshal(jsonBytes, &instance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to typed instance: %w", err)
	}

	// 5. 인터페이스 만족 확인 및 반환
	if eventMessage, ok := instance.(cqrs.EventMessage); ok {
		return eventMessage, nil
	}

	return nil, fmt.Errorf("unmarshaled BSON event of type '%s' does not implement EventMessage interface", typeExtractor.EventType)
}

// 테스트: 개선된 UnmarshalEventBSON 함수 검증
func TestUnmarshalEventBSON_Fixed_Solution(t *testing.T) {
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

	// 3. 기존 방식으로 언마샬링 - bson.D 문제 재현
	t.Log("=== 기존 방식 테스트 (문제 재현) ===")
	instance, err := registry.CreateDataInstance("ComplexNestedEvent")
	require.NoError(t, err, "Failed to create data instance")

	// 직접 bson.Unmarshal - 여기서 bson.D 문제 발생
	err = bson.Unmarshal(bsonData, &instance)
	require.NoError(t, err, "Direct bson.Unmarshal should not fail")

	t.Logf("기존 방식 결과 타입: %T", instance)

	// bson.D로 변환되었는지 확인
	if _, isBsonD := instance.(primitive.D); isBsonD {
		t.Log("✓ 기존 방식에서 bson.D 문제 재현됨")
	} else {
		t.Error("기존 방식에서 bson.D 문제가 재현되지 않음")
	}

	// 4. 개선된 방식으로 언마샬링 - 문제 해결
	t.Log("=== 개선된 방식 테스트 (문제 해결) ===")
	unmarshaledEvent, err := UnmarshalEventBSON_Fixed(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON_Fixed should not fail")

	t.Logf("개선된 방식 결과 타입: %T", unmarshaledEvent)

	// 5. 결과 검증
	complexEvent, ok := unmarshaledEvent.(*ComplexNestedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *ComplexNestedEventMessage")

	// 6. bson.D 문제가 해결되었는지 확인
	t.Log("=== bson.D 문제 해결 확인 ===")

	hasNoBsonDProblem := true

	if containsBsonD(complexEvent.PlayerInfo) {
		hasNoBsonDProblem = false
		t.Error("PlayerInfo still contains bson.D")
	} else {
		t.Log("✓ PlayerInfo: bson.D 문제 해결됨")
	}

	if containsBsonD(complexEvent.GameConfig) {
		hasNoBsonDProblem = false
		t.Error("GameConfig still contains bson.D")
	} else {
		t.Log("✓ GameConfig: bson.D 문제 해결됨")
	}

	if containsBsonD(complexEvent.EventMeta) {
		hasNoBsonDProblem = false
		t.Error("EventMeta still contains bson.D")
	} else {
		t.Log("✓ EventMeta: bson.D 문제 해결됨")
	}

	for i, item := range complexEvent.Inventory {
		if containsBsonD(item) {
			hasNoBsonDProblem = false
			t.Errorf("Inventory[%d] still contains bson.D", i)
		}
	}

	if hasNoBsonDProblem {
		t.Log("✓ 모든 필드에서 bson.D 문제가 해결되었습니다!")
	}

	// 7. 데이터 무결성 확인
	t.Log("=== 데이터 무결성 확인 ===")

	// PlayerInfo 확인
	if playerInfo, ok := complexEvent.PlayerInfo["playerID"]; ok {
		assert.Equal(t, "player123", playerInfo, "PlayerID should be preserved")
		t.Log("✓ PlayerInfo 데이터 무결성 확인됨")
	}

	// GameConfig 확인
	if gameConfig, ok := complexEvent.GameConfig["difficulty"]; ok {
		assert.Equal(t, "hard", gameConfig, "Difficulty should be preserved")
		t.Log("✓ GameConfig 데이터 무결성 확인됨")
	}

	// EventMeta 확인
	if eventMeta, ok := complexEvent.EventMeta["source"]; ok {
		assert.Equal(t, "game_server", eventMeta, "Source should be preserved")
		t.Log("✓ EventMeta 데이터 무결성 확인됨")
	}

	// Inventory 확인
	if len(complexEvent.Inventory) > 0 {
		if firstItem, ok := complexEvent.Inventory[0]["slotID"]; ok {
			// JSON으로 변환하면서 숫자가 float64가 될 수 있음
			if slotID, ok := firstItem.(float64); ok {
				assert.Equal(t, float64(1), slotID, "First inventory slotID should be 1")
				t.Log("✓ Inventory 데이터 무결성 확인됨")
			}
		}
	}

	// 8. 성능 비교 (선택적)
	t.Log("=== 성능 비교 ===")

	// 기존 방식 성능 측정 (실패하지만 시간 측정)
	start := time.Now()
	for i := 0; i < 100; i++ {
		instance, _ := registry.CreateDataInstance("ComplexNestedEvent")
		bson.Unmarshal(bsonData, &instance)
	}
	oldDuration := time.Since(start)

	// 개선된 방식 성능 측정
	start = time.Now()
	for i := 0; i < 100; i++ {
		UnmarshalEventBSON_Fixed(bsonData, registry)
	}
	newDuration := time.Since(start)

	t.Logf("기존 방식 (100회): %v", oldDuration)
	t.Logf("개선된 방식 (100회): %v", newDuration)
	t.Logf("성능 비율: %.2fx", float64(newDuration)/float64(oldDuration))
}

// 테스트: 간단한 이벤트에서도 해결책이 동작하는지 확인
func TestUnmarshalEventBSON_Fixed_SimpleEvent(t *testing.T) {
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
	require.NoError(t, err, "Failed to marshal simple event to BSON")

	// 3. 개선된 방식으로 언마샬링
	unmarshaledEvent, err := UnmarshalEventBSON_Fixed(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON_Fixed should not fail for simple event")

	// 4. 결과 검증
	transportEvent, ok := unmarshaledEvent.(*TransportCreatedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *TransportCreatedEventMessage")

	// 5. 데이터 무결성 확인
	assert.Equal(t, "TransportCreated", transportEvent.EventType())
	assert.WithinDuration(t, originalEvent.StartedAt, transportEvent.StartedAt, time.Second)

	t.Log("✓ 간단한 이벤트에서도 개선된 방식이 정상 동작함")
}

// 테스트: 다양한 데이터 타입에 대한 bson.D 문제 확인
func TestBsonD_VariousDataTypes(t *testing.T) {
	// 다양한 데이터 타입이 포함된 복잡한 구조
	testData := map[string]interface{}{
		"stringField": "test_string",
		"intField":    42,
		"floatField":  3.14159,
		"boolField":   true,
		"arrayField": []interface{}{
			"string_in_array",
			123,
			map[string]interface{}{
				"nested_in_array": "value",
			},
		},
		"mapField": map[string]interface{}{
			"nested_string": "nested_value",
			"nested_int":    999,
			"deeply_nested": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": "deep_value",
				},
			},
		},
		"complexArray": []interface{}{
			map[string]interface{}{
				"item1": map[string]interface{}{
					"property": "value1",
				},
			},
			map[string]interface{}{
				"item2": map[string]interface{}{
					"property": "value2",
				},
			},
		},
	}

	// BSON으로 마샬링 후 다시 언마샬링
	bsonData, err := bson.Marshal(testData)
	require.NoError(t, err, "Failed to marshal test data")

	var unmarshaled map[string]interface{}
	err = bson.Unmarshal(bsonData, &unmarshaled)
	require.NoError(t, err, "Failed to unmarshal test data")

	// bson.D 문제 확인
	t.Log("=== 다양한 데이터 타입에서 bson.D 문제 확인 ===")
	checkForBsonDInMap(t, "TestData", unmarshaled)

	// JSON을 통한 변환으로 문제 해결 확인
	jsonBytes, err := json.Marshal(unmarshaled)
	require.NoError(t, err, "Failed to marshal to JSON")

	var fixedData map[string]interface{}
	err = json.Unmarshal(jsonBytes, &fixedData)
	require.NoError(t, err, "Failed to unmarshal from JSON")

	t.Log("=== JSON 변환 후 bson.D 문제 해결 확인 ===")
	hasNoBsonD := !containsBsonD(fixedData)
	if hasNoBsonD {
		t.Log("✓ JSON 변환을 통해 bson.D 문제가 해결됨")
	} else {
		t.Error("JSON 변환 후에도 bson.D 문제가 남아있음")
	}
}
