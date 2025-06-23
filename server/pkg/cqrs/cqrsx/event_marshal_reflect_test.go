package cqrsx

import (
	"cqrs"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 새로운 EventRegistry 인터페이스: reflect.Value를 반환
type EventRegistryV2 interface {
	// 기존 메서드들
	RegisterEventDataType(eventType string, dataType reflect.Type) error
	GetDataType(eventType string) (reflect.Type, error)
	GetRegisteredEventTypes() []string
	IsRegistered(eventType string) bool

	CreateDataInstance(eventType string) (interface{}, error)

	// 새로운 메서드: reflect.Value를 반환
	CreateDataInstanceValue(eventType string) (reflect.Value, error)

	// 편의 메서드
	RegisterDataStruct(eventType string, dataStruct interface{}) error
}

// reflect.Value 기반 EventRegistry 구현
type ReflectValueEventRegistry struct {
	*InMemoryEventRegistry // 기존 구현 임베딩
}

func NewReflectValueEventRegistry() *ReflectValueEventRegistry {
	return &ReflectValueEventRegistry{
		InMemoryEventRegistry: NewInMemoryEventRegistry(),
	}
}

// CreateDataInstanceValue는 reflect.Value를 반환하는 새로운 메서드
func (r *ReflectValueEventRegistry) CreateDataInstanceValue(eventType string) (reflect.Value, error) {
	r.mu.RLock()
	dataType, exists := r.eventDataTypes[eventType]
	r.mu.RUnlock()

	if !exists {
		return reflect.Value{}, fmt.Errorf("unknown event type: %s", eventType)
	}

	// reflect.Value로 직접 새 인스턴스 생성
	return reflect.New(dataType), nil
}

// 기존 CreateDataInstance도 유지 (호환성을 위해)
func (r *ReflectValueEventRegistry) CreateDataInstance(eventType string) (interface{}, error) {
	value, err := r.CreateDataInstanceValue(eventType)
	if err != nil {
		return nil, err
	}
	return value.Interface(), nil
}

// reflect.Value 기반 UnmarshalEventBSON 구현
func UnmarshalEventBSON_ReflectValue(data []byte, registry EventRegistryV2) (cqrs.EventMessage, error) {
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

	// 2. reflect.Value로 인스턴스 생성
	instanceValue, err := registry.CreateDataInstance(typeExtractor.EventType)
	if err != nil {
		return nil, err
	}

	// 3. BSON을 map[string]interface{}로 언마샬링
	if err := bson.Unmarshal(data, instanceValue); err != nil {
		return nil, fmt.Errorf("failed to unmarshal BSON to raw map: %w", err)
	}

	// 5. interface{}로 변환 후 EventMessage 인터페이스 확인
	if eventMessage, ok := instanceValue.(cqrs.EventMessage); ok {
		return eventMessage, nil
	}

	return nil, fmt.Errorf("unmarshaled BSON event of type '%s' does not implement EventMessage interface", typeExtractor.EventType)
}

// mapRawDataToReflectValue는 reflect.Value를 직접 사용하여 매핑
func mapRawDataToReflectValue(rawData map[string]interface{}, targetValue reflect.Value) error {
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	return mapDataToValue(rawData, targetValue.Elem())
}

// mapDataToValue는 재귀적으로 데이터를 reflect.Value로 매핑
func mapDataToValue(data interface{}, targetValue reflect.Value) error {
	if !targetValue.CanSet() {
		return nil // 설정할 수 없는 필드는 건너뛰기
	}

	// 특별한 타입들을 먼저 처리
	if targetValue.Type() == reflect.TypeOf(time.Time{}) {
		if timeVal, ok := convertToTime(data); ok {
			targetValue.Set(reflect.ValueOf(timeVal))
			return nil
		}
	}

	if targetValue.Type() == reflect.TypeOf(primitive.ObjectID{}) {
		if objID, ok := data.(primitive.ObjectID); ok {
			targetValue.Set(reflect.ValueOf(objID))
			return nil
		}
	}

	switch targetValue.Kind() {
	case reflect.String:
		if str, ok := data.(string); ok {
			targetValue.SetString(str)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := convertToInt64(data); ok {
			targetValue.SetInt(num)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, ok := convertToUint64(data); ok {
			targetValue.SetUint(num)
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := convertToFloat64(data); ok {
			targetValue.SetFloat(num)
		}
	case reflect.Bool:
		if b, ok := data.(bool); ok {
			targetValue.SetBool(b)
		}
	case reflect.Interface:
		// interface{} 타입의 경우 bson.D를 map[string]interface{}로 변환
		convertedData := convertBsonDToMap(data)
		targetValue.Set(reflect.ValueOf(convertedData))
	case reflect.Map:
		return mapToMapValue(data, targetValue)
	case reflect.Slice:
		return mapToSliceValue(data, targetValue)
	case reflect.Struct:
		return mapToStructValue(data, targetValue)
	case reflect.Ptr:
		if targetValue.IsNil() {
			targetValue.Set(reflect.New(targetValue.Type().Elem()))
		}
		return mapDataToValue(data, targetValue.Elem())
	default:
		// 다른 타입들은 직접 설정 시도
		if reflect.TypeOf(data).AssignableTo(targetValue.Type()) {
			targetValue.Set(reflect.ValueOf(data))
		}
	}

	return nil
}

// convertBsonDToMap은 bson.D를 map[string]interface{}로 재귀적으로 변환
func convertBsonDToMap(data interface{}) interface{} {
	switch v := data.(type) {
	case primitive.D:
		// bson.D를 map[string]interface{}로 변환
		result := make(map[string]interface{})
		for _, elem := range v {
			result[elem.Key] = convertBsonDToMap(elem.Value)
		}
		return result
	case []interface{}:
		// 배열의 각 요소도 재귀적으로 변환
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertBsonDToMap(item)
		}
		return result
	case map[string]interface{}:
		// 맵의 각 값도 재귀적으로 변환
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = convertBsonDToMap(value)
		}
		return result
	default:
		return data
	}
}

// mapToMapValue는 맵 타입으로 매핑
func mapToMapValue(data interface{}, targetValue reflect.Value) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// bson.D인 경우 변환
		if bsonD, isBsonD := data.(primitive.D); isBsonD {
			dataMap = make(map[string]interface{})
			for _, elem := range bsonD {
				dataMap[elem.Key] = convertBsonDToMap(elem.Value)
			}
		} else {
			return nil
		}
	}

	if targetValue.IsNil() {
		targetValue.Set(reflect.MakeMap(targetValue.Type()))
	}

	mapValueType := targetValue.Type().Elem()

	for key, value := range dataMap {
		convertedValue := convertBsonDToMap(value)

		// 타입 호환성 확인 및 변환
		valueToSet := reflect.ValueOf(convertedValue)
		if valueToSet.IsValid() && valueToSet.Type().AssignableTo(mapValueType) {
			targetValue.SetMapIndex(reflect.ValueOf(key), valueToSet)
		} else if mapValueType.Kind() == reflect.Interface {
			// interface{} 타입인 경우 그대로 설정
			targetValue.SetMapIndex(reflect.ValueOf(key), valueToSet)
		} else {
			// 타입이 맞지 않는 경우 변환 시도
			if convertedVal, ok := convertValueToType(convertedValue, mapValueType); ok {
				targetValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(convertedVal))
			}
		}
	}

	return nil
}

// mapToSliceValue는 슬라이스 타입으로 매핑
func mapToSliceValue(data interface{}, targetValue reflect.Value) error {
	dataSlice, ok := data.([]interface{})
	if !ok {
		return nil
	}

	sliceType := targetValue.Type()
	elemType := sliceType.Elem()

	result := reflect.MakeSlice(sliceType, len(dataSlice), len(dataSlice))

	for i, item := range dataSlice {
		elemValue := result.Index(i)

		// 슬라이스 요소가 포인터 타입인 경우
		if elemType.Kind() == reflect.Ptr {
			newElem := reflect.New(elemType.Elem())
			if err := mapDataToValue(item, newElem.Elem()); err != nil {
				return err
			}
			elemValue.Set(newElem)
		} else {
			if err := mapDataToValue(item, elemValue); err != nil {
				return err
			}
		}
	}

	targetValue.Set(result)
	return nil
}

// mapToStructValue는 구조체 타입으로 매핑
func mapToStructValue(data interface{}, targetValue reflect.Value) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// bson.D인 경우 변환
		if bsonD, isBsonD := data.(primitive.D); isBsonD {
			dataMap = make(map[string]interface{})
			for _, elem := range bsonD {
				dataMap[elem.Key] = elem.Value
			}
		} else {
			return nil
		}
	}

	targetType := targetValue.Type()

	for i := 0; i < targetValue.NumField(); i++ {
		field := targetValue.Field(i)
		fieldType := targetType.Field(i)

		// 익스포트되지 않은 필드는 건너뛰기
		if !field.CanSet() {
			continue
		}

		// bson 태그에서 필드 이름 가져오기
		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		// 태그 파싱 (예: "fieldName,omitempty" -> "fieldName")
		fieldName := bsonTag
		if commaIdx := findChar(bsonTag, ','); commaIdx != -1 {
			fieldName = bsonTag[:commaIdx]
		}

		// inline 태그 처리
		if fieldName == "" && findString(bsonTag, "inline") != -1 {
			// inline인 경우 현재 데이터를 그대로 재귀 매핑
			if err := mapDataToValue(data, field); err != nil {
				return err
			}
			continue
		}

		// 일반 필드명인 경우 데이터에서 해당 필드 값 찾기
		if fieldName != "" {
			if value, exists := dataMap[fieldName]; exists {
				if err := mapDataToValue(value, field); err != nil {
					return err
				}
			}
		} else {
			// fieldName이 빈 경우, 필드 이름을 그대로 사용
			fieldRealName := fieldType.Name
			if value, exists := dataMap[fieldRealName]; exists {
				if err := mapDataToValue(value, field); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// 헬퍼 함수들

func convertToInt64(data interface{}) (int64, bool) {
	switch v := data.(type) {
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case float32:
		return int64(v), true
	default:
		return 0, false
	}
}

func convertToUint64(data interface{}) (uint64, bool) {
	switch v := data.(type) {
	case uint:
		return uint64(v), true
	case uint32:
		return uint64(v), true
	case uint64:
		return v, true
	case int:
		if v >= 0 {
			return uint64(v), true
		}
	case int64:
		if v >= 0 {
			return uint64(v), true
		}
	default:
		return 0, false
	}
	return 0, false
}

func convertToFloat64(data interface{}) (float64, bool) {
	switch v := data.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func convertToTime(data interface{}) (time.Time, bool) {
	switch v := data.(type) {
	case time.Time:
		return v, true
	case int64:
		// Unix timestamp (milliseconds)
		return time.Unix(0, v*int64(time.Millisecond)), true
	case float64:
		// Unix timestamp (seconds)
		return time.Unix(int64(v), 0), true
	case primitive.DateTime:
		return v.Time(), true
	default:
		return time.Time{}, false
	}
}

// convertValueToType은 값을 대상 타입으로 변환
func convertValueToType(value interface{}, targetType reflect.Type) (interface{}, bool) {
	if value == nil {
		return nil, true
	}

	valueType := reflect.TypeOf(value)
	if valueType.AssignableTo(targetType) {
		return value, true
	}

	// 특별한 타입 변환들
	switch targetType {
	case reflect.TypeOf(time.Time{}):
		if timeVal, ok := convertToTime(value); ok {
			return timeVal, true
		}
	case reflect.TypeOf(primitive.ObjectID{}):
		if objID, ok := value.(primitive.ObjectID); ok {
			return objID, true
		}
	}

	// 숫자 타입 변환
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if num, ok := convertToInt64(value); ok {
			return reflect.ValueOf(num).Convert(targetType).Interface(), true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num, ok := convertToUint64(value); ok {
			return reflect.ValueOf(num).Convert(targetType).Interface(), true
		}
	case reflect.Float32, reflect.Float64:
		if num, ok := convertToFloat64(value); ok {
			return reflect.ValueOf(num).Convert(targetType).Interface(), true
		}
	case reflect.String:
		if str, ok := value.(string); ok {
			return str, true
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			return b, true
		}
	}

	return nil, false
}

func findChar(s string, c rune) int {
	for i, char := range s {
		if char == c {
			return i
		}
	}
	return -1
}

func findString(s, substr string) int {
	// 간단한 문자열 검색
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// 테스트용 레지스트리 설정
func setupReflectValueTestRegistry() EventRegistryV2 {
	registry := NewReflectValueEventRegistry()

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

// 테스트: reflect.Value 기반 해결책 검증
func TestUnmarshalEventBSON_ReflectValue_Solution(t *testing.T) {
	registry := setupReflectValueTestRegistry()

	t.Log("=== reflect.Value 기반 BSON 언마샬링 테스트 ===")

	// 1. 복잡한 중첩 구조의 이벤트 생성
	originalEvent := ComplexNestedEventMessage{
		BaseEventMessage:       cqrs.NewBaseEventMessage("ComplexNestedEvent"),
		ComplexNestedEventData: createComplexNestedEventData(),
	}

	// 2. BSON으로 마샬링
	bsonData, err := MarshalEventBSON(originalEvent)
	require.NoError(t, err, "Failed to marshal complex event to BSON")

	t.Logf("BSON data size: %d bytes", len(bsonData))

	// 3. reflect.Value 기반 방식으로 언마샬링
	unmarshaledEvent, err := UnmarshalEventBSON_ReflectValue(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON_ReflectValue should not fail")

	t.Logf("reflect.Value 방식 결과 타입: %T", unmarshaledEvent)

	// 4. 결과 검증
	complexEvent, ok := unmarshaledEvent.(*ComplexNestedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *ComplexNestedEventMessage")

	// 5. bson.D 문제가 해결되었는지 확인
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

	// 6. 데이터 무결성 확인
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
}

// 성능 비교 테스트: JSON 변환 vs reflect.Value
func TestPerformanceComparison_JSON_vs_ReflectValue(t *testing.T) {
	jsonRegistry := setupBsonDTestRegistry()
	reflectRegistry := setupReflectValueTestRegistry()

	// 복잡한 이벤트로 테스트
	originalEvent := ComplexNestedEventMessage{
		BaseEventMessage:       cqrs.NewBaseEventMessage("ComplexNestedEvent"),
		ComplexNestedEventData: createComplexNestedEventData(),
	}

	bsonData, err := MarshalEventBSON(originalEvent)
	require.NoError(t, err, "Failed to marshal event")

	t.Log("=== 성능 비교 테스트 ===")

	// JSON 변환 방식 성능 측정
	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := UnmarshalEventBSON(bsonData, jsonRegistry)
		require.NoError(t, err)
	}
	jsonDuration := time.Since(start)

	// reflect.Value 방식 성능 측정
	start = time.Now()
	for i := 0; i < 1000; i++ {
		_, err := UnmarshalEventBSON_ReflectValue(bsonData, reflectRegistry)
		require.NoError(t, err)
	}
	reflectDuration := time.Since(start)

	t.Logf("JSON 변환 방식 (1000회): %v", jsonDuration)
	t.Logf("reflect.Value 방식 (1000회): %v", reflectDuration)
	t.Logf("성능 개선: %.2fx", float64(jsonDuration)/float64(reflectDuration))

	if reflectDuration < jsonDuration {
		t.Log("✓ reflect.Value 방식이 더 빠릅니다!")
	} else {
		t.Log("⚠ JSON 변환 방식이 더 빠릅니다.")
	}
}

// 벤치마크 테스트
func BenchmarkUnmarshalEventBSON_Methods_Comparison(b *testing.B) {
	jsonRegistry := setupBsonDTestRegistry()
	reflectRegistry := setupReflectValueTestRegistry()

	originalEvent := ComplexNestedEventMessage{
		BaseEventMessage:       cqrs.NewBaseEventMessage("ComplexNestedEvent"),
		ComplexNestedEventData: createComplexNestedEventData(),
	}

	bsonData, err := MarshalEventBSON(originalEvent)
	if err != nil {
		b.Fatalf("Failed to marshal event: %v", err)
	}

	b.Run("JSON_Method", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := UnmarshalEventBSON(bsonData, jsonRegistry)
			if err != nil {
				b.Fatalf("UnmarshalEventBSON failed: %v", err)
			}
		}
	})

	b.Run("ReflectValue_Method", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := UnmarshalEventBSON_ReflectValue(bsonData, reflectRegistry)
			if err != nil {
				b.Fatalf("UnmarshalEventBSON_ReflectValue failed: %v", err)
			}
		}
	})
}

// 간단한 이벤트 테스트
func TestUnmarshalEventBSON_ReflectValue_SimpleEvent(t *testing.T) {
	registry := setupReflectValueTestRegistry()

	// 간단한 이벤트 생성
	originalEvent := TransportCreatedEventMessage{
		BaseEventMessage: cqrs.NewBaseEventMessage("TransportCreated"),
		TransportCreatedEventData: TransportCreatedEventData{
			StartedAt: time.Now().UTC(),
		},
	}

	// BSON으로 마샬링
	bsonData, err := MarshalEventBSON(originalEvent)
	require.NoError(t, err, "Failed to marshal simple event to BSON")

	// reflect.Value 방식으로 언마샬링
	unmarshaledEvent, err := UnmarshalEventBSON_ReflectValue(bsonData, registry)
	require.NoError(t, err, "UnmarshalEventBSON_ReflectValue should not fail for simple event")

	// 결과 검증
	transportEvent, ok := unmarshaledEvent.(*TransportCreatedEventMessage)
	require.True(t, ok, "Unmarshaled event should be *TransportCreatedEventMessage")

	// 데이터 무결성 확인
	assert.Equal(t, "TransportCreated", transportEvent.EventType())
	assert.WithinDuration(t, originalEvent.StartedAt, transportEvent.StartedAt, time.Minute)

	t.Log("✓ 간단한 이벤트에서도 reflect.Value 방식이 정상 동작함")
}
