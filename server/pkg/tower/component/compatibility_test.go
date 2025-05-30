package component

import (
	"encoding/json"
	"testing"
)

func TestTypeCompatibilityMatrix(t *testing.T) {
	matrix := NewTypeCompatibilityMatrix()

	// Test exact match
	if !matrix.IsCompatible(DataTypeInt, DataTypeInt) {
		t.Error("Expected exact type match to be compatible")
	}

	// Test Any type compatibility
	if !matrix.IsCompatible(DataTypeAny, DataTypeString) {
		t.Error("Expected Any type to be compatible with String")
	}

	if !matrix.IsCompatible(DataTypeFloat, DataTypeAny) {
		t.Error("Expected Float to be compatible with Any type")
	}

	// Test default rules
	if !matrix.IsCompatible(DataTypeInt, DataTypeFloat) {
		t.Error("Expected Int to be compatible with Float")
	}

	if !matrix.IsCompatible(DataTypeTarget, DataTypeTargets) {
		t.Error("Expected Target to be compatible with Targets")
	}
}

func TestAddCompatibilityRule(t *testing.T) {
	matrix := NewTypeCompatibilityMatrix()

	// Add a custom rule
	rule := TypeCompatibilityRule{
		FromType:      DataTypeString,
		ToType:        DataTypeInt,
		Bidirectional: false,
		Cost:          5,
		Transform:     "parseInt(value)",
	}

	matrix.AddRule(rule)

	// Test the custom rule
	if !matrix.IsCompatible(DataTypeString, DataTypeInt) {
		t.Error("Expected custom rule to make String compatible with Int")
	}

	// Test that reverse is not automatically added (bidirectional = false)
	if matrix.IsCompatible(DataTypeInt, DataTypeString) {
		t.Error("Expected reverse compatibility to not exist when bidirectional = false")
	}

	// Test bidirectional rule
	bidirectionalRule := TypeCompatibilityRule{
		FromType:      DataTypeBool,
		ToType:        DataTypeString,
		Bidirectional: true,
		Cost:          2,
		Transform:     "string(value)",
	}

	matrix.AddRule(bidirectionalRule)

	if !matrix.IsCompatible(DataTypeBool, DataTypeString) {
		t.Error("Expected Bool to be compatible with String")
	}

	if !matrix.IsCompatible(DataTypeString, DataTypeBool) {
		t.Error("Expected String to be compatible with Bool (bidirectional)")
	}
}

func TestGetCompatibilityRule(t *testing.T) {
	matrix := NewTypeCompatibilityMatrix()

	// Test getting a default rule
	rule, exists := matrix.GetCompatibilityRule(DataTypeInt, DataTypeFloat)
	if !exists {
		t.Error("Expected to find compatibility rule for Int -> Float")
	}

	if rule.Cost != 1 {
		t.Errorf("Expected cost to be 1, got %d", rule.Cost)
	}

	if rule.Transform != "float(value)" {
		t.Errorf("Expected transform to be 'float(value)', got '%s'", rule.Transform)
	}
}

func TestGetCompatibleTypes(t *testing.T) {
	matrix := NewTypeCompatibilityMatrix()

	compatibleTypes := matrix.GetCompatibleTypes(DataTypeTarget)

	// Should include at least: target (self), any, targets, enemy, enemies
	expectedTypes := map[DataType]bool{
		DataTypeTarget:  true,
		DataTypeAny:     true,
		DataTypeTargets: true,
		DataTypeEnemy:   true,
		DataTypeEnemies: true,
	}

	found := make(map[DataType]bool)
	for _, dataType := range compatibleTypes {
		found[dataType] = true
	}

	for expectedType := range expectedTypes {
		if !found[expectedType] {
			t.Errorf("Expected %s to be compatible with Target", expectedType)
		}
	}
}

func TestAdvancedCompatibilityMatrix(t *testing.T) {
	matrix := NewAdvancedCompatibilityMatrix()

	// Test loading from config
	config := CreateDefaultCompatibilityConfig()
	err := matrix.LoadFromConfig(config)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test category functionality
	numericTypes, err := matrix.GetCategoryTypes("numeric")
	if err != nil {
		t.Fatalf("Failed to get numeric types: %v", err)
	}

	expectedNumericTypes := []DataType{DataTypeInt, DataTypeFloat}
	if len(numericTypes) != len(expectedNumericTypes) {
		t.Errorf("Expected %d numeric types, got %d", len(expectedNumericTypes), len(numericTypes))
	}

	// Test preset loading
	err = matrix.LoadPreset("basic")
	if err != nil {
		t.Fatalf("Failed to load basic preset: %v", err)
	}

	// Test type category lookup
	category, found := matrix.GetTypeCategory(DataTypeInt)
	if !found {
		t.Error("Expected to find category for Int type")
	}

	if category != "numeric" {
		t.Errorf("Expected Int to be in 'numeric' category, got '%s'", category)
	}
}

func TestCompatibilityPath(t *testing.T) {
	matrix := NewAdvancedCompatibilityMatrix()

	// Add some rules to create a path
	matrix.AddRule(TypeCompatibilityRule{
		FromType: DataTypeString, ToType: DataTypeInt,
		Cost: 3, Transform: "parseInt(value)",
	})

	matrix.AddRule(TypeCompatibilityRule{
		FromType: DataTypeInt, ToType: DataTypeFloat,
		Cost: 1, Transform: "float(value)",
	})

	// Test direct path
	path, cost, found := matrix.FindCompatibilityPath(DataTypeInt, DataTypeFloat)
	if !found {
		t.Error("Expected to find direct path from Int to Float")
	}

	if len(path) != 1 {
		t.Errorf("Expected path length 1, got %d", len(path))
	}

	if cost != 1 {
		t.Errorf("Expected cost 1, got %d", cost)
	}

	// Test indirect path
	path, cost, found = matrix.FindCompatibilityPath(DataTypeString, DataTypeFloat)
	if !found {
		t.Error("Expected to find indirect path from String to Float")
	}

	if len(path) != 2 {
		t.Errorf("Expected path length 2, got %d", len(path))
	}

	if cost != 4 { // 3 + 1
		t.Errorf("Expected total cost 4, got %d", cost)
	}
}

func TestValidateConfiguration(t *testing.T) {
	matrix := NewAdvancedCompatibilityMatrix()

	// Add a circular dependency
	matrix.AddRule(TypeCompatibilityRule{
		FromType: DataTypeString, ToType: DataTypeInt,
		Cost: 1,
	})
	matrix.AddRule(TypeCompatibilityRule{
		FromType: DataTypeInt, ToType: DataTypeString,
		Cost: 1,
	})

	errors := matrix.ValidateConfiguration()

	// Should detect circular dependency
	hasCircularError := false
	for _, err := range errors {
		if err.Error() == "circular dependency detected starting from type string" ||
			err.Error() == "circular dependency detected starting from type int" {
			hasCircularError = true
			break
		}
	}

	if !hasCircularError {
		t.Error("Expected to detect circular dependency")
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Test global IsCompatible function
	if !IsCompatible(DataTypeInt, DataTypeFloat) {
		t.Error("Expected global IsCompatible to work")
	}

	// Test adding global rule
	AddCompatibilityRule(TypeCompatibilityRule{
		FromType: DataTypeString, ToType: DataTypeVector2,
		Cost: 10, Transform: "parseVector(value)",
	})

	if !IsCompatible(DataTypeString, DataTypeVector2) {
		t.Error("Expected global rule to be applied")
	}

	// Test getting compatible types
	compatibleTypes := GetCompatibleTypes(DataTypeString)
	found := false
	for _, dataType := range compatibleTypes {
		if dataType == DataTypeVector2 {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected Vector2 to be in compatible types for String")
	}
}

func TestJSONExport(t *testing.T) {
	matrix := NewAdvancedCompatibilityMatrix()

	// Load default config
	config := CreateDefaultCompatibilityConfig()
	err := matrix.LoadFromConfig(config)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Export to JSON
	jsonData, err := matrix.ExportToJSON()
	if err != nil {
		t.Fatalf("Failed to export to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON export")
	}

	// Verify it's valid JSON by trying to parse it back
	var exportedConfig CompatibilityConfig
	err = json.Unmarshal(jsonData, &exportedConfig)
	if err != nil {
		t.Fatalf("Failed to parse exported JSON: %v", err)
	}

	if exportedConfig.Version == "" {
		t.Error("Expected exported config to have version")
	}
}

// Benchmark tests
func BenchmarkIsCompatible(b *testing.B) {
	matrix := NewTypeCompatibilityMatrix()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matrix.IsCompatible(DataTypeTarget, DataTypeEnemy)
	}
}

func BenchmarkFindCompatibilityPath(b *testing.B) {
	matrix := NewAdvancedCompatibilityMatrix()

	// Add some rules
	matrix.AddRule(TypeCompatibilityRule{FromType: DataTypeString, ToType: DataTypeInt, Cost: 1})
	matrix.AddRule(TypeCompatibilityRule{FromType: DataTypeInt, ToType: DataTypeFloat, Cost: 1})
	matrix.AddRule(TypeCompatibilityRule{FromType: DataTypeFloat, ToType: DataTypeVector2, Cost: 1})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		matrix.FindCompatibilityPath(DataTypeString, DataTypeVector2)
	}
}
