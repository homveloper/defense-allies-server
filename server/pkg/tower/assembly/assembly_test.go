package assembly

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"defense-allies-server/pkg/tower/component"
	"defense-allies-server/pkg/tower/component/impl"
)

// Initialize components for testing
func init() {
	impl.RegisterAllComponents()
}

// TestComponent is a simple component for testing with no required inputs
type TestComponent struct {
	id       string
	metadata component.ComponentMetadata
}

func NewTestComponent() *TestComponent {
	return &TestComponent{
		id: fmt.Sprintf("test_component_%d", time.Now().UnixNano()),
		metadata: component.ComponentMetadata{
			Name:        "Test Component",
			Description: "A simple test component with no required inputs",
			Category:    "test",
			Version:     "1.0",
		},
	}
}

func (tc *TestComponent) GetType() component.ComponentType {
	return "test_component"
}

func (tc *TestComponent) GetID() string {
	return tc.id
}

func (tc *TestComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"test_output": "test_value",
		},
		Effects: []component.Effect{},
		Events:  []component.GameEvent{},
	}, nil
}

func (tc *TestComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "optional_input",
			Type:        component.DataTypeString,
			Required:    false,
			Description: "Optional test input",
		},
	}
}

func (tc *TestComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "test_output",
			Type:        component.DataTypeString,
			Description: "Test output",
		},
	}
}

func (tc *TestComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeString, input.Type) {
			return true
		}
	}
	return false
}

func (tc *TestComponent) Validate() error {
	return nil
}

func (tc *TestComponent) Clone() component.AtomicComponent {
	return &TestComponent{
		id:       fmt.Sprintf("test_component_%d", time.Now().UnixNano()),
		metadata: tc.metadata,
	}
}

func (tc *TestComponent) GetMetadata() component.ComponentMetadata {
	return tc.metadata
}

func TestComponentAssembly_Basic(t *testing.T) {
	// Create a new assembly
	assembly := NewComponentAssembly("Test Assembly")

	assert.Equal(t, "Test Assembly", assembly.Name)
	assert.NotEmpty(t, assembly.ID)
	assert.False(t, assembly.IsValid)
	assert.Empty(t, assembly.Components)
	assert.Empty(t, assembly.Connections)
}

func TestComponentAssembly_AddComponent(t *testing.T) {
	assembly := NewComponentAssembly("Test Assembly")

	// Create a test component
	config := map[string]interface{}{
		"range":    10.0,
		"priority": "closest",
	}

	comp, err := impl.NewSingleTargetComponent(config)
	require.NoError(t, err)

	// Add component to assembly
	err = assembly.AddComponent(comp)
	require.NoError(t, err)

	assert.Len(t, assembly.Components, 1)
	assert.Contains(t, assembly.Components, comp.GetID())
	assert.False(t, assembly.IsValid) // Should need revalidation

	// Try to add the same component again (should fail)
	err = assembly.AddComponent(comp)
	assert.Error(t, err)
}

func TestComponentAssembly_RemoveComponent(t *testing.T) {
	assembly := NewComponentAssembly("Test Assembly")

	// Create and add a component
	config := map[string]interface{}{
		"base_damage": 100.0,
	}

	comp, err := impl.NewBasicDamageComponent(config)
	require.NoError(t, err)

	err = assembly.AddComponent(comp)
	require.NoError(t, err)

	componentID := comp.GetID()

	// Remove the component
	err = assembly.RemoveComponent(componentID)
	require.NoError(t, err)

	assert.Empty(t, assembly.Components)

	// Try to remove non-existent component
	err = assembly.RemoveComponent("non-existent")
	assert.Error(t, err)
}

func TestComponentAssembly_AddConnection(t *testing.T) {
	assembly := NewComponentAssembly("Test Assembly")

	// Create targeting component
	targetingConfig := map[string]interface{}{
		"range":    10.0,
		"priority": "closest",
	}
	targeting, err := impl.NewSingleTargetComponent(targetingConfig)
	require.NoError(t, err)

	// Create damage component
	damageConfig := map[string]interface{}{
		"base_damage": 100.0,
	}
	damage, err := impl.NewBasicDamageComponent(damageConfig)
	require.NoError(t, err)

	// Add components
	err = assembly.AddComponent(targeting)
	require.NoError(t, err)
	err = assembly.AddComponent(damage)
	require.NoError(t, err)

	// Create connection
	conn := component.ComponentConnection{
		ID:            "test_connection",
		FromComponent: targeting.GetID(),
		FromOutput:    "target",
		ToComponent:   damage.GetID(),
		ToInput:       "targets",
		Type:          component.ConnectionTypeSequential,
		Enabled:       true,
		Priority:      component.PriorityNormal,
	}

	// Add connection
	err = assembly.AddConnection(conn)
	require.NoError(t, err)

	assert.Len(t, assembly.Connections, 1)
	assert.False(t, assembly.IsValid) // Should need revalidation
}

func TestAssemblyValidator_Basic(t *testing.T) {
	validator := NewAssemblyValidator()
	assembly := NewComponentAssembly("Test Assembly")

	// Empty assembly should be valid but with warnings
	err := validator.ValidateAssembly(context.Background(), assembly)
	assert.NoError(t, err)
	assert.True(t, assembly.IsValid)
}

func TestAssemblyValidator_WithComponents(t *testing.T) {
	validator := NewAssemblyValidator()
	assembly := NewComponentAssembly("Test Assembly")

	// Create a simple test component (no required inputs)
	testComp := NewTestComponent()

	err := assembly.AddComponent(testComp)
	require.NoError(t, err)

	// Validate assembly with single component
	err = validator.ValidateAssembly(context.Background(), assembly)
	if err != nil {
		t.Logf("Validation errors: %v", assembly.Errors)
		t.Logf("Validation warnings: %v", assembly.Warnings)
	}
	assert.NoError(t, err)
	assert.True(t, assembly.IsValid)
	assert.NotEmpty(t, assembly.Warnings) // Should have orphaned component warnings
	assert.NotEmpty(t, assembly.ExecutionOrder)
	assert.NotEmpty(t, assembly.EntryPoints)
	assert.NotEmpty(t, assembly.ExitPoints)
}

func TestAssemblyEngine_Basic(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewAssemblyEngine(config)

	assert.NotNil(t, engine)

	// Test empty engine
	assemblies := engine.ListAssemblies()
	assert.Empty(t, assemblies)

	stats := engine.GetEngineStats()
	assert.Equal(t, 0, stats.RegisteredAssemblies)
}

func TestAssemblyEngine_RegisterAssembly(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewAssemblyEngine(config)

	// Create a simple assembly
	assembly := CreateSimpleAssembly("Test Assembly")

	// Register assembly
	err := engine.RegisterAssembly(assembly)
	require.NoError(t, err)

	// Check that assembly is registered
	assemblies := engine.ListAssemblies()
	assert.Len(t, assemblies, 1)

	// Try to register the same assembly again
	err = engine.RegisterAssembly(assembly)
	assert.Error(t, err)

	// Get assembly by ID
	retrieved, err := engine.GetAssembly(assembly.ID)
	require.NoError(t, err)
	assert.Equal(t, assembly.ID, retrieved.ID)
}

func TestAssemblyEngine_ExecuteAssembly(t *testing.T) {
	config := DefaultEngineConfig()
	engine := NewAssemblyEngine(config)

	// Create and register a simple assembly
	assembly := CreateSimpleAssembly("Test Assembly")

	// Add a simple test component
	testComp := NewTestComponent()

	err := assembly.AddComponent(testComp)
	require.NoError(t, err)

	err = engine.RegisterAssembly(assembly)
	require.NoError(t, err)

	// Create execution context
	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		TowerID:   "test_tower",
		TowerPos:  component.Vector2{X: 0, Y: 0},
		OwnerID:   "test_player",
		InputData: map[string]interface{}{
			"targets": []component.Enemy{}, // Empty targets for testing
		},
		ExecutionID: "test_execution",
	}

	// Execute assembly
	result, err := engine.ExecuteAssembly(assembly.ID, execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestAssemblyClone(t *testing.T) {
	assembly := NewComponentAssembly("Original Assembly")
	assembly.Description = "Original description"

	// Add a test component
	comp := NewTestComponent()

	err := assembly.AddComponent(comp)
	require.NoError(t, err)

	// Clone the assembly
	clone := assembly.Clone()

	assert.NotEqual(t, assembly.ID, clone.ID, "Clone should have different assembly ID")
	assert.Equal(t, "Original Assembly (Copy)", clone.Name)
	assert.Equal(t, assembly.Description, clone.Description)
	assert.Len(t, clone.Components, 1)
	assert.False(t, clone.IsValid) // Clone should need revalidation

	// Verify components are actually cloned (different instances)
	originalComp := assembly.Components[comp.GetID()]
	var clonedComp component.AtomicComponent
	for _, c := range clone.Components {
		clonedComp = c
		break
	}

	// Note: Component cloning creates new IDs, so they should be different
	// But the original component ID might still be the same in the clone's map
	assert.Equal(t, originalComp.GetType(), clonedComp.GetType())

	// The clone should have different component instances even if map keys are same
	assert.NotSame(t, originalComp, clonedComp)
}

func TestAssemblyStats(t *testing.T) {
	assembly := NewComponentAssembly("Test Assembly")

	// Initial stats
	stats := assembly.GetStats()
	assert.Equal(t, 0, stats.ComponentCount)
	assert.Equal(t, 0, stats.ConnectionCount)
	assert.False(t, stats.IsValid)

	// Add components
	config1 := map[string]interface{}{"range": 10.0, "priority": "closest"}
	comp1, err := impl.NewSingleTargetComponent(config1)
	require.NoError(t, err)

	config2 := map[string]interface{}{"base_damage": 100.0}
	comp2, err := impl.NewBasicDamageComponent(config2)
	require.NoError(t, err)

	err = assembly.AddComponent(comp1)
	require.NoError(t, err)
	err = assembly.AddComponent(comp2)
	require.NoError(t, err)

	// Add connection
	conn := component.ComponentConnection{
		ID:            "test_connection",
		FromComponent: comp1.GetID(),
		FromOutput:    "target",
		ToComponent:   comp2.GetID(),
		ToInput:       "targets",
		Type:          component.ConnectionTypeSequential,
		Enabled:       true,
	}

	err = assembly.AddConnection(conn)
	require.NoError(t, err)

	// Updated stats
	stats = assembly.GetStats()
	assert.Equal(t, 2, stats.ComponentCount)
	assert.Equal(t, 1, stats.ConnectionCount)
	assert.Greater(t, stats.Complexity, 0.0)
}

func TestRealAssemblyExecutor(t *testing.T) {
	config := DefaultExecutorConfig()
	executor := NewRealAssemblyExecutor(config)

	// Create a simple assembly with one component
	assembly := NewComponentAssembly("Test Assembly")

	testComp := NewTestComponent()

	err := assembly.AddComponent(testComp)
	require.NoError(t, err)

	// Validate assembly
	validator := NewAssemblyValidator()
	err = validator.ValidateAssembly(context.Background(), assembly)
	require.NoError(t, err)

	// Create execution context
	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		InputData: map[string]interface{}{
			"targets": []component.Enemy{}, // Empty for testing
		},
	}

	// Execute assembly
	result, err := executor.ExecuteAssembly(context.Background(), assembly, execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.GreaterOrEqual(t, result.ExecutionTime, time.Duration(0), "Execution time should be non-negative")
}
