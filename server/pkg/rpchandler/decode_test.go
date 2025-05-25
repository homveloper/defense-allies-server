package rpchandler

import (
	"reflect"
	"testing"
)

// TestDecodeJSONArrayToTypes DecodeJSONArrayToTypes 함수 테스트
func TestDecodeJSONArrayToTypes(t *testing.T) {
	t.Run("BasicTypes", func(t *testing.T) {
		// 기본 타입들 테스트
		jsonArray := `[42, "hello", true, 3.14]`
		paramTypes := []reflect.Type{
			reflect.TypeOf(int(0)),
			reflect.TypeOf(""),
			reflect.TypeOf(bool(false)),
			reflect.TypeOf(float64(0)),
		}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode JSON array: %v", err)
		}
		
		if len(values) != 4 {
			t.Fatalf("Expected 4 values, got %d", len(values))
		}
		
		// int 값 확인
		if values[0].Interface().(int) != 42 {
			t.Errorf("Expected int 42, got %v", values[0].Interface())
		}
		
		// string 값 확인
		if values[1].Interface().(string) != "hello" {
			t.Errorf("Expected string 'hello', got %v", values[1].Interface())
		}
		
		// bool 값 확인
		if values[2].Interface().(bool) != true {
			t.Errorf("Expected bool true, got %v", values[2].Interface())
		}
		
		// float64 값 확인
		if values[3].Interface().(float64) != 3.14 {
			t.Errorf("Expected float64 3.14, got %v", values[3].Interface())
		}
	})
	
	t.Run("PointerTypes", func(t *testing.T) {
		// 포인터 타입들 테스트
		jsonArray := `[100, "world"]`
		paramTypes := []reflect.Type{
			reflect.TypeOf((*int)(nil)),
			reflect.TypeOf((*string)(nil)),
		}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode JSON array: %v", err)
		}
		
		if len(values) != 2 {
			t.Fatalf("Expected 2 values, got %d", len(values))
		}
		
		// *int 값 확인
		intPtr := values[0].Interface().(*int)
		if *intPtr != 100 {
			t.Errorf("Expected *int 100, got %v", *intPtr)
		}
		
		// *string 값 확인
		strPtr := values[1].Interface().(*string)
		if *strPtr != "world" {
			t.Errorf("Expected *string 'world', got %v", *strPtr)
		}
	})
	
	t.Run("StructType", func(t *testing.T) {
		// 구조체 타입 테스트
		type TestStruct struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		
		jsonArray := `[{"id": 123, "name": "test"}]`
		paramTypes := []reflect.Type{
			reflect.TypeOf(TestStruct{}),
		}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode JSON array: %v", err)
		}
		
		if len(values) != 1 {
			t.Fatalf("Expected 1 value, got %d", len(values))
		}
		
		// 구조체 값 확인
		testStruct := values[0].Interface().(TestStruct)
		if testStruct.ID != 123 {
			t.Errorf("Expected ID 123, got %v", testStruct.ID)
		}
		if testStruct.Name != "test" {
			t.Errorf("Expected Name 'test', got %v", testStruct.Name)
		}
	})
	
	t.Run("PointerToStruct", func(t *testing.T) {
		// 구조체 포인터 타입 테스트
		type TestStruct struct {
			Value int `json:"value"`
		}
		
		jsonArray := `[{"value": 456}]`
		paramTypes := []reflect.Type{
			reflect.TypeOf((*TestStruct)(nil)),
		}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode JSON array: %v", err)
		}
		
		if len(values) != 1 {
			t.Fatalf("Expected 1 value, got %d", len(values))
		}
		
		// 구조체 포인터 값 확인
		testStructPtr := values[0].Interface().(*TestStruct)
		if testStructPtr.Value != 456 {
			t.Errorf("Expected Value 456, got %v", testStructPtr.Value)
		}
	})
	
	t.Run("MixedTypes", func(t *testing.T) {
		// 혼합 타입들 테스트
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		
		jsonArray := `[42, "hello", {"name": "John", "age": 30}, true]`
		paramTypes := []reflect.Type{
			reflect.TypeOf(int(0)),
			reflect.TypeOf(""),
			reflect.TypeOf(Person{}),
			reflect.TypeOf(bool(false)),
		}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode JSON array: %v", err)
		}
		
		if len(values) != 4 {
			t.Fatalf("Expected 4 values, got %d", len(values))
		}
		
		// 각 값 확인
		if values[0].Interface().(int) != 42 {
			t.Errorf("Expected int 42, got %v", values[0].Interface())
		}
		
		if values[1].Interface().(string) != "hello" {
			t.Errorf("Expected string 'hello', got %v", values[1].Interface())
		}
		
		person := values[2].Interface().(Person)
		if person.Name != "John" || person.Age != 30 {
			t.Errorf("Expected Person{Name: 'John', Age: 30}, got %+v", person)
		}
		
		if values[3].Interface().(bool) != true {
			t.Errorf("Expected bool true, got %v", values[3].Interface())
		}
	})
	
	t.Run("EmptyArray", func(t *testing.T) {
		// 빈 배열 테스트
		jsonArray := `[]`
		paramTypes := []reflect.Type{}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode empty JSON array: %v", err)
		}
		
		if len(values) != 0 {
			t.Fatalf("Expected 0 values, got %d", len(values))
		}
	})
	
	t.Run("NullValues", func(t *testing.T) {
		// null 값들 테스트
		jsonArray := `[null, null]`
		paramTypes := []reflect.Type{
			reflect.TypeOf((*int)(nil)),
			reflect.TypeOf((*string)(nil)),
		}
		
		values, err := DecodeJSONArrayToTypes(jsonArray, paramTypes)
		if err != nil {
			t.Fatalf("Failed to decode JSON array with nulls: %v", err)
		}
		
		if len(values) != 2 {
			t.Fatalf("Expected 2 values, got %d", len(values))
		}
		
		// null 포인터 확인
		if !values[0].IsNil() {
			t.Error("Expected first value to be nil")
		}
		
		if !values[1].IsNil() {
			t.Error("Expected second value to be nil")
		}
	})
	
	t.Run("ErrorCases", func(t *testing.T) {
		// 에러 케이스들
		testCases := []struct {
			name      string
			jsonArray string
			types     []reflect.Type
		}{
			{
				name:      "InvalidJSON",
				jsonArray: `[invalid json`,
				types:     []reflect.Type{reflect.TypeOf(int(0))},
			},
			{
				name:      "TypeMismatch",
				jsonArray: `["string"]`,
				types:     []reflect.Type{reflect.TypeOf(int(0))},
			},
			{
				name:      "ArrayLengthMismatch",
				jsonArray: `[1, 2, 3]`,
				types:     []reflect.Type{reflect.TypeOf(int(0)), reflect.TypeOf(int(0))},
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := DecodeJSONArrayToTypes(tc.jsonArray, tc.types)
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tc.name)
				}
				t.Logf("Expected error for %s: %v", tc.name, err)
			})
		}
	})
}
