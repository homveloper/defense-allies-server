// Package factory provides tower factory functionality for creating
// tower instances from JSON definitions in the Defense Allies system.
package factory

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"defense-allies-server/pkg/tower/assembly"
	"defense-allies-server/pkg/tower/component"
	"defense-allies-server/pkg/tower/component/impl"
	"defense-allies-server/pkg/tower/definition"
)

// TowerFactory manages tower definitions and creates tower instances
type TowerFactory struct {
	definitions map[string]*definition.TowerDefinition
	assemblies  map[string]*assembly.ComponentAssembly
	engine      *assembly.AssemblyEngine
	mutex       sync.RWMutex

	// Configuration
	dataPath   string
	autoReload bool
	lastReload time.Time
}

// FactoryConfig provides configuration for the tower factory
type FactoryConfig struct {
	DataPath   string `json:"data_path"`
	AutoReload bool   `json:"auto_reload"`
}

// DefaultFactoryConfig returns a default factory configuration
func DefaultFactoryConfig() FactoryConfig {
	return FactoryConfig{
		DataPath:   "data/towers",
		AutoReload: true,
	}
}

// NewTowerFactory creates a new tower factory
func NewTowerFactory(config FactoryConfig) (*TowerFactory, error) {
	// Initialize component registry (only if not already done)
	// This prevents duplicate registration errors in tests
	if len(impl.GetRegisteredComponentTypes()) == 0 {
		if err := impl.RegisterAllComponents(); err != nil {
			return nil, fmt.Errorf("failed to register components: %w", err)
		}
	}

	// Create assembly engine
	engineConfig := assembly.DefaultEngineConfig()
	engine := assembly.NewAssemblyEngine(engineConfig)

	factory := &TowerFactory{
		definitions: make(map[string]*definition.TowerDefinition),
		assemblies:  make(map[string]*assembly.ComponentAssembly),
		engine:      engine,
		dataPath:    config.DataPath,
		autoReload:  config.AutoReload,
		lastReload:  time.Now(),
	}

	// Load initial definitions
	if err := factory.LoadDefinitions(); err != nil {
		return nil, fmt.Errorf("failed to load initial definitions: %w", err)
	}

	return factory, nil
}

// LoadDefinitions loads all tower definitions from the data directory
func (tf *TowerFactory) LoadDefinitions() error {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	// Clear existing definitions
	tf.definitions = make(map[string]*definition.TowerDefinition)
	tf.assemblies = make(map[string]*assembly.ComponentAssembly)

	// Walk through the data directory
	err := filepath.WalkDir(tf.dataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process JSON files
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Load the definition
		if err := tf.loadDefinitionFile(path); err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk data directory: %w", err)
	}

	tf.lastReload = time.Now()
	return nil
}

// loadDefinitionFile loads a single tower definition file
func (tf *TowerFactory) loadDefinitionFile(filePath string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the definition
	towerDef, err := definition.FromJSON(data)
	if err != nil {
		return fmt.Errorf("failed to parse definition: %w", err)
	}

	// Create assembly from definition
	assemblyDef, err := tf.createAssemblyFromDefinition(towerDef)
	if err != nil {
		return fmt.Errorf("failed to create assembly: %w", err)
	}

	// Register assembly with engine
	if err := tf.engine.RegisterAssembly(assemblyDef); err != nil {
		// Add more detailed error information
		return fmt.Errorf("failed to register assembly for tower %s: %w\nAssembly errors: %v\nAssembly warnings: %v",
			towerDef.ID, err, assemblyDef.Errors, assemblyDef.Warnings)
	}

	// Store the definition and assembly
	tf.definitions[towerDef.ID] = towerDef
	tf.assemblies[towerDef.ID] = assemblyDef

	return nil
}

// createAssemblyFromDefinition creates a component assembly from a tower definition
func (tf *TowerFactory) createAssemblyFromDefinition(towerDef *definition.TowerDefinition) (*assembly.ComponentAssembly, error) {
	// Create new assembly
	assemblyDef := assembly.NewComponentAssembly(towerDef.Name + " Assembly")
	assemblyDef.Description = fmt.Sprintf("Component assembly for %s", towerDef.Name)
	assemblyDef.Metadata.Category = towerDef.Category
	assemblyDef.Metadata.Race = towerDef.Race
	assemblyDef.Metadata.Tags = towerDef.Tags

	// Create components
	componentMap := make(map[string]component.AtomicComponent)
	for _, compDef := range towerDef.Assembly.Components {
		comp, err := impl.CreateComponentByType(component.ComponentType(compDef.Type), compDef.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create component %s: %w", compDef.ID, err)
		}

		// Store with the definition ID for connection mapping
		componentMap[compDef.ID] = comp

		// Add to assembly with the actual component ID
		if err := assemblyDef.AddComponent(comp); err != nil {
			return nil, fmt.Errorf("failed to add component %s: %w", compDef.ID, err)
		}
	}

	// Create connections
	for _, connDef := range towerDef.Assembly.Connections {
		fromComp := componentMap[connDef.FromComponent]
		toComp := componentMap[connDef.ToComponent]

		if fromComp == nil {
			return nil, fmt.Errorf("from component %s not found", connDef.FromComponent)
		}
		if toComp == nil {
			return nil, fmt.Errorf("to component %s not found", connDef.ToComponent)
		}

		// Convert string priority to Priority type
		var priority component.Priority
		switch connDef.Priority {
		case "lowest":
			priority = component.PriorityLowest
		case "low":
			priority = component.PriorityLow
		case "high":
			priority = component.PriorityHigh
		case "highest":
			priority = component.PriorityHighest
		default:
			priority = component.PriorityNormal
		}

		conn := component.ComponentConnection{
			ID:            generateConnectionID(),
			FromComponent: fromComp.GetID(),
			FromOutput:    connDef.FromOutput,
			ToComponent:   toComp.GetID(),
			ToInput:       connDef.ToInput,
			Type:          component.ConnectionType(connDef.Type),
			Enabled:       connDef.Enabled,
			Priority:      priority,
		}

		if err := assemblyDef.AddConnection(conn); err != nil {
			return nil, fmt.Errorf("failed to add connection %s: %w", connDef.ID, err)
		}
	}

	return assemblyDef, nil
}

// GetDefinition returns a tower definition by ID
func (tf *TowerFactory) GetDefinition(towerID string) (*definition.TowerDefinition, error) {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	if tf.autoReload {
		tf.checkAndReload()
	}

	def, exists := tf.definitions[towerID]
	if !exists {
		return nil, fmt.Errorf("tower definition %s not found", towerID)
	}

	return def, nil
}

// GetAllDefinitions returns all loaded tower definitions
func (tf *TowerFactory) GetAllDefinitions() map[string]*definition.TowerDefinition {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	if tf.autoReload {
		tf.checkAndReload()
	}

	// Return a copy to prevent external modification
	result := make(map[string]*definition.TowerDefinition)
	for id, def := range tf.definitions {
		result[id] = def
	}

	return result
}

// GetDefinitionsByRace returns all tower definitions for a specific race
func (tf *TowerFactory) GetDefinitionsByRace(race string) []*definition.TowerDefinition {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	if tf.autoReload {
		tf.checkAndReload()
	}

	var result []*definition.TowerDefinition
	for _, def := range tf.definitions {
		if def.Race == race {
			result = append(result, def)
		}
	}

	return result
}

// GetDefinitionsByCategory returns all tower definitions for a specific category
func (tf *TowerFactory) GetDefinitionsByCategory(category string) []*definition.TowerDefinition {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	if tf.autoReload {
		tf.checkAndReload()
	}

	var result []*definition.TowerDefinition
	for _, def := range tf.definitions {
		if def.Category == category {
			result = append(result, def)
		}
	}

	return result
}

// CreateTowerInstance creates a new tower instance from a definition
func (tf *TowerFactory) CreateTowerInstance(towerID, ownerID string, position component.Vector2) (*definition.TowerInstance, error) {
	// Get the definition
	towerDef, err := tf.GetDefinition(towerID)
	if err != nil {
		return nil, err
	}

	// Create the instance
	instance := definition.NewTowerInstance(towerID, ownerID, position)

	// Copy stats from definition (level 1)
	instance.CurrentStats = towerDef.Stats
	instance.Health = towerDef.Stats.Health

	// Set assembly ID
	if assemblyDef, exists := tf.assemblies[towerID]; exists {
		instance.AssemblyID = assemblyDef.ID
	}

	return instance, nil
}

// ExecuteTowerAction executes a tower's action (attack, ability, etc.)
func (tf *TowerFactory) ExecuteTowerAction(instance *definition.TowerInstance, execCtx *component.ExecutionContext) (*assembly.AssemblyResult, error) {
	if instance.AssemblyID == "" {
		return nil, fmt.Errorf("tower instance has no assembly ID")
	}

	// Execute the assembly
	return tf.engine.ExecuteAssembly(instance.AssemblyID, execCtx)
}

// UpgradeTower upgrades a tower instance to the next level
func (tf *TowerFactory) UpgradeTower(instance *definition.TowerInstance) error {
	// Get the definition
	towerDef, err := tf.GetDefinition(instance.DefinitionID)
	if err != nil {
		return err
	}

	// Check if upgrade is possible
	if instance.Level >= towerDef.Scaling.MaxLevel {
		return fmt.Errorf("tower is already at maximum level")
	}

	// Upgrade the tower
	instance.Level++

	// Update stats based on scaling
	instance.CurrentStats.Damage = towerDef.Stats.Damage + (towerDef.Scaling.DamagePerLevel * float64(instance.Level-1))
	instance.CurrentStats.Range = towerDef.Stats.Range + (towerDef.Scaling.RangePerLevel * float64(instance.Level-1))
	instance.CurrentStats.Health = towerDef.Stats.Health + (towerDef.Scaling.HealthPerLevel * float64(instance.Level-1))

	// Update health to new maximum
	instance.Health = instance.CurrentStats.Health
	instance.UpdatedAt = time.Now()

	return nil
}

// checkAndReload checks if definitions need to be reloaded
func (tf *TowerFactory) checkAndReload() {
	// Simple time-based reload check (in production, use file watchers)
	if time.Since(tf.lastReload) > time.Minute*5 {
		tf.LoadDefinitions()
	}
}

// GetFactoryStats returns statistics about the factory
func (tf *TowerFactory) GetFactoryStats() FactoryStats {
	tf.mutex.RLock()
	defer tf.mutex.RUnlock()

	stats := FactoryStats{
		TotalDefinitions:  len(tf.definitions),
		RaceBreakdown:     make(map[string]int),
		CategoryBreakdown: make(map[string]int),
		LastReload:        tf.lastReload,
	}

	for _, def := range tf.definitions {
		stats.RaceBreakdown[def.Race]++
		stats.CategoryBreakdown[def.Category]++
	}

	return stats
}

// FactoryStats provides statistics about the tower factory
type FactoryStats struct {
	TotalDefinitions  int            `json:"total_definitions"`
	RaceBreakdown     map[string]int `json:"race_breakdown"`
	CategoryBreakdown map[string]int `json:"category_breakdown"`
	LastReload        time.Time      `json:"last_reload"`
}

// Helper functions
func generateConnectionID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}
