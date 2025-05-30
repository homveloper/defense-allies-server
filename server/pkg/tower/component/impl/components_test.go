package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"defense-allies-server/pkg/tower/component"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Register components once for all tests
func init() {
	RegisterAllComponents()
}

// Mock implementations for testing
type MockEnemy struct {
	id        string
	position  component.Vector2
	enemyType string
}

func (me *MockEnemy) GetID() string {
	return me.id
}

func (me *MockEnemy) GetPosition() component.Vector2 {
	return me.position
}

func (me *MockEnemy) GetType() string {
	return me.enemyType
}

func TestSingleTargetComponent(t *testing.T) {
	// Create single target component
	config := map[string]interface{}{
		"range":    10.0,
		"priority": "closest",
	}

	comp, err := NewSingleTargetComponent(config)
	require.NoError(t, err, "Failed to create single target component")

	// Test component properties
	assert.Equal(t, component.ComponentTypeSingleTarget, comp.GetType())
	assert.NotEmpty(t, comp.GetID(), "Component ID should not be empty")

	// Test validation
	assert.NoError(t, comp.Validate(), "Component validation should pass")

	// Test inputs and outputs
	inputs := comp.GetInputs()
	assert.Len(t, inputs, 2, "Should have 2 inputs")

	outputs := comp.GetOutputs()
	assert.Len(t, outputs, 1, "Should have 1 output")

	// Test execution
	enemies := []component.Enemy{
		&MockEnemy{id: "enemy1", position: component.Vector2{X: 5, Y: 0}, enemyType: "basic"},
		&MockEnemy{id: "enemy2", position: component.Vector2{X: 15, Y: 0}, enemyType: "basic"}, // Out of range
		&MockEnemy{id: "enemy3", position: component.Vector2{X: 2, Y: 1}, enemyType: "basic"},  // Distance = √5 ≈ 2.24 (closest)
	}

	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"tower_position":    component.Vector2{X: 0, Y: 0},
			"available_enemies": enemies,
		},
	}

	result, err := comp.Execute(context.Background(), execCtx)
	require.NoError(t, err, "Component execution should succeed")
	assert.True(t, result.Success, "Execution should be successful")

	target, ok := result.Outputs["target"]
	assert.True(t, ok, "Should have target output")

	if target != nil {
		targetEnemy, ok := target.(component.Enemy)
		require.True(t, ok, "Target should be an Enemy")
		assert.Equal(t, "enemy3", targetEnemy.GetID(), "Should target closest enemy (enemy3)")
	}
}

func TestBasicDamageComponent(t *testing.T) {
	// Create basic damage component
	config := map[string]interface{}{
		"base_damage":         100.0,
		"critical_chance":     0.0, // No crit for predictable testing
		"critical_multiplier": 2.0,
		"variance":            0.0, // No variance for predictable testing
	}

	comp, err := NewBasicDamageComponent(config)
	require.NoError(t, err, "Failed to create basic damage component")

	// Test component properties
	assert.Equal(t, component.ComponentTypeBasicDamage, comp.GetType())

	// Test execution
	targets := []component.Enemy{
		&MockEnemy{id: "enemy1", position: component.Vector2{X: 0, Y: 0}, enemyType: "basic"},
		&MockEnemy{id: "enemy2", position: component.Vector2{X: 5, Y: 0}, enemyType: "basic"},
	}

	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"targets": targets,
		},
	}

	result, err := comp.Execute(context.Background(), execCtx)
	require.NoError(t, err, "Component execution should succeed")
	assert.True(t, result.Success, "Execution should be successful")

	damageEvents, ok := result.Outputs["damage_events"].([]component.Effect)
	require.True(t, ok, "Should have damage_events output")
	assert.Len(t, damageEvents, 2, "Should have 2 damage events")

	// Check damage values
	for _, event := range damageEvents {
		assert.Equal(t, component.EffectTypeDamage, event.Type, "Should be damage effect")
		assert.Equal(t, 100.0, event.Intensity, "Should deal 100 damage")
	}
}

func TestFireDamageComponent(t *testing.T) {
	// Create fire damage component
	config := map[string]interface{}{
		"base_damage":   120.0,
		"burn_duration": 3.0,
		"burn_dps":      25.0,
	}

	comp, err := NewFireDamageComponent(config)
	require.NoError(t, err, "Failed to create fire damage component")

	// Test component properties
	assert.Equal(t, component.ComponentTypeFireDamage, comp.GetType())

	// Test execution
	targets := []component.Enemy{
		&MockEnemy{id: "enemy1", position: component.Vector2{X: 0, Y: 0}, enemyType: "basic"},
	}

	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"targets": targets,
		},
	}

	result, err := comp.Execute(context.Background(), execCtx)
	require.NoError(t, err, "Component execution should succeed")
	assert.True(t, result.Success, "Execution should be successful")

	// Should have both damage and burn effects
	_, ok := result.Outputs["damage_events"].([]component.Effect)
	assert.True(t, ok, "Should have damage_events output")

	burnEffects, ok := result.Outputs["burn_effects"].([]component.Effect)
	require.True(t, ok, "Should have burn_effects output")
	assert.Len(t, burnEffects, 1, "Should have 1 burn effect")

	// Check burn effect properties
	burnEffect := burnEffects[0]
	assert.Equal(t, component.EffectTypeBurn, burnEffect.Type, "Should be burn effect")
	assert.Equal(t, 3.0, burnEffect.Duration, "Should have 3 second duration")
	assert.Equal(t, 25.0, burnEffect.Intensity, "Should have 25 DPS")
}

func TestRangeCheckComponent(t *testing.T) {
	// Create range check component
	config := map[string]interface{}{
		"range": 10.0,
		"shape": "circle",
	}

	comp, err := NewRangeCheckComponent(config)
	require.NoError(t, err, "Failed to create range check component")

	// Test component properties
	assert.Equal(t, component.ComponentTypeRangeCheck, comp.GetType())

	// Test execution
	targets := []component.Enemy{
		&MockEnemy{id: "enemy1", position: component.Vector2{X: 5, Y: 0}, enemyType: "basic"},  // In range
		&MockEnemy{id: "enemy2", position: component.Vector2{X: 15, Y: 0}, enemyType: "basic"}, // Out of range
		&MockEnemy{id: "enemy3", position: component.Vector2{X: 0, Y: 8}, enemyType: "basic"},  // In range
	}

	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"source_position": component.Vector2{X: 0, Y: 0},
			"targets":         targets,
		},
	}

	result, err := comp.Execute(context.Background(), execCtx)
	require.NoError(t, err, "Component execution should succeed")
	assert.True(t, result.Success, "Execution should be successful")

	targetsInRange, ok := result.Outputs["targets_in_range"].([]component.Enemy)
	require.True(t, ok, "Should have targets_in_range output")
	assert.Len(t, targetsInRange, 2, "Should have 2 targets in range")

	// Check area info
	areaInfo, ok := result.Outputs["area_info"].(AreaInfo)
	require.True(t, ok, "Should have area_info output")
	assert.Equal(t, 2, areaInfo.TotalTargets, "Should have 2 total targets")
}

func TestComponentRegistry(t *testing.T) {
	// Components should already be registered globally
	// Just test that they exist

	// Test getting registered types
	types := GetRegisteredComponentTypes()
	assert.NotEmpty(t, types, "Should have registered component types")

	// Test creating component by type
	config := map[string]interface{}{
		"range":    8.0,
		"priority": "closest",
	}

	comp, err := CreateComponentByType(component.ComponentTypeSingleTarget, config)
	require.NoError(t, err, "Should create component by type")
	assert.Equal(t, component.ComponentTypeSingleTarget, comp.GetType())

	// Test getting component info
	info, err := GetComponentInfo(component.ComponentTypeSingleTarget)
	require.NoError(t, err, "Should get component info")
	assert.Equal(t, component.ComponentTypeSingleTarget, info.Type)
	assert.NotEmpty(t, info.Metadata.Name, "Should have component metadata name")
}

func TestComponentCloning(t *testing.T) {
	// Create original component
	config := map[string]interface{}{
		"base_damage": 150.0,
	}

	original, err := NewBasicDamageComponent(config)
	require.NoError(t, err, "Should create original component")

	// Clone the component
	clone := original.Clone()

	// Test that clone is different instance
	assert.NotEqual(t, original.GetID(), clone.GetID(), "Clone should have different ID")
	assert.Equal(t, original.GetType(), clone.GetType(), "Clone should have same type")

	// Test that both work independently
	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"targets": []component.Enemy{
				&MockEnemy{id: "enemy1", position: component.Vector2{X: 0, Y: 0}, enemyType: "basic"},
			},
		},
	}

	result1, err1 := original.Execute(context.Background(), execCtx)
	result2, err2 := clone.Execute(context.Background(), execCtx)

	assert.NoError(t, err1, "Original should execute successfully")
	assert.NoError(t, err2, "Clone should execute successfully")
	assert.True(t, result1.Success, "Original execution should be successful")
	assert.True(t, result2.Success, "Clone execution should be successful")
}

// Benchmark tests
func BenchmarkSingleTargetExecution(b *testing.B) {
	config := map[string]interface{}{
		"range":    10.0,
		"priority": "closest",
	}

	comp, _ := NewSingleTargetComponent(config)

	enemies := make([]component.Enemy, 100)
	for i := 0; i < 100; i++ {
		enemies[i] = &MockEnemy{
			id:        fmt.Sprintf("enemy%d", i),
			position:  component.Vector2{X: float64(i % 20), Y: float64(i / 20)},
			enemyType: "basic",
		}
	}

	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"tower_position":    component.Vector2{X: 0, Y: 0},
			"available_enemies": enemies,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Execute(context.Background(), execCtx)
	}
}

func BenchmarkDamageCalculation(b *testing.B) {
	config := map[string]interface{}{
		"base_damage": 100.0,
	}

	comp, _ := NewBasicDamageComponent(config)

	targets := []component.Enemy{
		&MockEnemy{id: "enemy1", position: component.Vector2{X: 0, Y: 0}, enemyType: "basic"},
	}

	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"targets": targets,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Execute(context.Background(), execCtx)
	}
}
