package factory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"defense-allies-server/pkg/tower/component"
)

const testDataPath = "../../../../data/towers"

func TestTowerFactory_LoadDefinitions(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err, "Should create tower factory")

	// Check that definitions were loaded
	stats := factory.GetFactoryStats()
	assert.Greater(t, stats.TotalDefinitions, 0, "Should load at least one definition")

	// Check Human Alliance towers
	humanTowers := factory.GetDefinitionsByRace("human_alliance")
	assert.GreaterOrEqual(t, len(humanTowers), 4, "Should have at least 4 Human Alliance towers")

	// Verify specific towers exist
	archerTower, err := factory.GetDefinition("human_archer_tower")
	require.NoError(t, err, "Should find Archer Tower")
	assert.Equal(t, "Archer Tower", archerTower.Name)
	assert.Equal(t, "human_alliance", archerTower.Race)
	assert.Equal(t, "offensive", archerTower.Category)

	cannonTower, err := factory.GetDefinition("human_cannon_tower")
	require.NoError(t, err, "Should find Cannon Tower")
	assert.Equal(t, "Cannon Tower", cannonTower.Name)
	assert.Contains(t, cannonTower.Tags, "area_damage")

	mageTower, err := factory.GetDefinition("human_mage_tower")
	require.NoError(t, err, "Should find Mage Tower")
	assert.Equal(t, "Mage Tower", mageTower.Name)
	assert.Contains(t, mageTower.Tags, "magic")

	barracks, err := factory.GetDefinition("human_barracks")
	require.NoError(t, err, "Should find Barracks")
	assert.Equal(t, "Barracks", barracks.Name)
	assert.Equal(t, "support", barracks.Category)
}

func TestTowerFactory_GetDefinitionsByCategory(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	// Get offensive towers
	offensiveTowers := factory.GetDefinitionsByCategory("offensive")
	assert.GreaterOrEqual(t, len(offensiveTowers), 3, "Should have at least 3 offensive towers")

	// Get support towers
	supportTowers := factory.GetDefinitionsByCategory("support")
	assert.GreaterOrEqual(t, len(supportTowers), 1, "Should have at least 1 support tower")

	// Verify categories
	for _, tower := range offensiveTowers {
		assert.Equal(t, "offensive", tower.Category)
	}

	for _, tower := range supportTowers {
		assert.Equal(t, "support", tower.Category)
	}
}

func TestTowerFactory_CreateTowerInstance(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	// Create an Archer Tower instance
	position := component.Vector2{X: 10, Y: 20}
	instance, err := factory.CreateTowerInstance("human_archer_tower", "player123", position)
	require.NoError(t, err, "Should create tower instance")

	// Verify instance properties
	assert.NotEmpty(t, instance.InstanceID)
	assert.Equal(t, "human_archer_tower", instance.DefinitionID)
	assert.Equal(t, "player123", instance.OwnerID)
	assert.Equal(t, position, instance.Position)
	assert.Equal(t, 1, instance.Level)
	assert.Equal(t, 0.0, instance.Experience)
	assert.True(t, instance.IsActive)
	assert.False(t, instance.IsBusy)
	assert.NotEmpty(t, instance.AssemblyID)

	// Verify stats were copied
	assert.Equal(t, 45.0, instance.CurrentStats.Damage)
	assert.Equal(t, 8.0, instance.CurrentStats.Range)
	assert.Equal(t, 1.2, instance.CurrentStats.AttackSpeed)
}

func TestTowerFactory_UpgradeTower(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	// Create a tower instance
	position := component.Vector2{X: 0, Y: 0}
	instance, err := factory.CreateTowerInstance("human_archer_tower", "player123", position)
	require.NoError(t, err)

	// Store original stats
	originalDamage := instance.CurrentStats.Damage
	originalRange := instance.CurrentStats.Range
	originalLevel := instance.Level

	// Upgrade the tower
	err = factory.UpgradeTower(instance)
	require.NoError(t, err, "Should upgrade tower successfully")

	// Verify upgrade
	assert.Equal(t, originalLevel+1, instance.Level)
	assert.Greater(t, instance.CurrentStats.Damage, originalDamage, "Damage should increase")
	assert.Greater(t, instance.CurrentStats.Range, originalRange, "Range should increase")

	// Verify specific scaling (8.0 damage per level, 0.2 range per level)
	expectedDamage := originalDamage + 8.0
	expectedRange := originalRange + 0.2
	assert.Equal(t, expectedDamage, instance.CurrentStats.Damage)
	assert.Equal(t, expectedRange, instance.CurrentStats.Range)
}

func TestTowerFactory_UpgradeTowerMaxLevel(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	// Create a tower instance
	position := component.Vector2{X: 0, Y: 0}
	instance, err := factory.CreateTowerInstance("human_archer_tower", "player123", position)
	require.NoError(t, err)

	// Upgrade to max level (10)
	for i := 1; i < 10; i++ {
		err = factory.UpgradeTower(instance)
		require.NoError(t, err, "Should upgrade to level %d", i+1)
	}

	assert.Equal(t, 10, instance.Level)

	// Try to upgrade beyond max level
	err = factory.UpgradeTower(instance)
	assert.Error(t, err, "Should not upgrade beyond max level")
	assert.Contains(t, err.Error(), "maximum level")
}

func TestTowerFactory_ExecuteTowerAction(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	// Create a tower instance
	position := component.Vector2{X: 0, Y: 0}
	instance, err := factory.CreateTowerInstance("human_archer_tower", "player123", position)
	require.NoError(t, err)

	// Create execution context
	execCtx := &component.ExecutionContext{
		GameTime:  time.Now(),
		DeltaTime: 0.016,
		TowerID:   instance.InstanceID,
		TowerPos:  instance.Position,
		OwnerID:   instance.OwnerID,
		InputData: map[string]interface{}{
			"tower_position":    instance.Position,
			"available_enemies": []component.Enemy{}, // Empty for testing
		},
		ExecutionID: "test_execution",
	}

	// Execute tower action
	result, err := factory.ExecuteTowerAction(instance, execCtx)
	require.NoError(t, err, "Should execute tower action")
	assert.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestTowerFactory_GetFactoryStats(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	stats := factory.GetFactoryStats()

	// Verify stats structure
	assert.Greater(t, stats.TotalDefinitions, 0)
	assert.NotEmpty(t, stats.RaceBreakdown)
	assert.NotEmpty(t, stats.CategoryBreakdown)
	assert.False(t, stats.LastReload.IsZero())

	// Verify Human Alliance count
	humanCount, exists := stats.RaceBreakdown["human_alliance"]
	assert.True(t, exists, "Should have Human Alliance towers")
	assert.GreaterOrEqual(t, humanCount, 4, "Should have at least 4 Human Alliance towers")

	// Verify category breakdown
	offensiveCount, exists := stats.CategoryBreakdown["offensive"]
	assert.True(t, exists, "Should have offensive towers")
	assert.GreaterOrEqual(t, offensiveCount, 3, "Should have at least 3 offensive towers")

	supportCount, exists := stats.CategoryBreakdown["support"]
	assert.True(t, exists, "Should have support towers")
	assert.GreaterOrEqual(t, supportCount, 1, "Should have at least 1 support tower")
}

func TestTowerFactory_NonExistentTower(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	// Try to get non-existent tower
	_, err = factory.GetDefinition("non_existent_tower")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Try to create instance of non-existent tower
	position := component.Vector2{X: 0, Y: 0}
	_, err = factory.CreateTowerInstance("non_existent_tower", "player123", position)
	assert.Error(t, err)
}

func TestTowerFactory_GetAllDefinitions(t *testing.T) {
	config := FactoryConfig{
		DataPath:   testDataPath,
		AutoReload: false,
	}

	factory, err := NewTowerFactory(config)
	require.NoError(t, err)

	allDefs := factory.GetAllDefinitions()
	assert.Greater(t, len(allDefs), 0, "Should have definitions")

	// Verify specific towers exist
	_, exists := allDefs["human_archer_tower"]
	assert.True(t, exists, "Should have Archer Tower")

	_, exists = allDefs["human_cannon_tower"]
	assert.True(t, exists, "Should have Cannon Tower")

	_, exists = allDefs["human_mage_tower"]
	assert.True(t, exists, "Should have Mage Tower")

	_, exists = allDefs["human_barracks"]
	assert.True(t, exists, "Should have Barracks")
}
