// Package definition provides tower definition structures and JSON schema
// for the Defense Allies tower system.
package definition

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"defense-allies-server/pkg/tower/component"
)

// TowerDefinition represents a complete tower definition that can be instantiated
type TowerDefinition struct {
	// Basic Information
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	
	// Classification
	Race        string   `json:"race"`        // "human_alliance", "elven_kingdom", etc.
	Category    string   `json:"category"`    // "offensive", "defensive", "support", "special"
	Tags        []string `json:"tags"`        // ["basic", "fire", "area"], etc.
	
	// Game Balance
	Cost        TowerCost        `json:"cost"`
	Stats       TowerStats       `json:"stats"`
	Scaling     TowerScaling     `json:"scaling"`
	Restrictions TowerRestrictions `json:"restrictions"`
	
	// Visual & Audio
	Appearance  TowerAppearance  `json:"appearance"`
	
	// Component Assembly
	Assembly    AssemblyDefinition `json:"assembly"`
	
	// Metadata
	Author      string            `json:"author,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TowerCost defines the cost to build and upgrade a tower
type TowerCost struct {
	BuildCost    ResourceCost            `json:"build_cost"`
	UpgradeCosts map[int]ResourceCost    `json:"upgrade_costs"` // Level -> Cost
	SellValue    float64                 `json:"sell_value"`    // Percentage of total investment
}

// ResourceCost represents the cost in various resources
type ResourceCost struct {
	Gold     int `json:"gold"`
	Wood     int `json:"wood,omitempty"`
	Stone    int `json:"stone,omitempty"`
	Iron     int `json:"iron,omitempty"`
	Gems     int `json:"gems,omitempty"`
	Mana     int `json:"mana,omitempty"`
	Souls    int `json:"souls,omitempty"`
	Energy   int `json:"energy,omitempty"`
}

// TowerStats defines the base statistics of a tower
type TowerStats struct {
	// Combat Stats
	Damage       float64 `json:"damage"`
	Range        float64 `json:"range"`
	AttackSpeed  float64 `json:"attack_speed"`  // Attacks per second
	CritChance   float64 `json:"crit_chance"`   // 0.0 to 1.0
	CritMultiplier float64 `json:"crit_multiplier"`
	
	// Defensive Stats
	Health       float64 `json:"health,omitempty"`
	Armor        float64 `json:"armor,omitempty"`
	MagicResist  float64 `json:"magic_resist,omitempty"`
	
	// Special Stats
	Accuracy     float64 `json:"accuracy,omitempty"`     // 0.0 to 1.0
	Penetration  float64 `json:"penetration,omitempty"`  // Armor penetration
	Splash       float64 `json:"splash,omitempty"`       // Splash damage radius
	
	// Resource Stats
	ManaCost     float64 `json:"mana_cost,omitempty"`    // Mana per attack
	Cooldown     float64 `json:"cooldown,omitempty"`     // Ability cooldown
}

// TowerScaling defines how tower stats scale with level and other factors
type TowerScaling struct {
	// Per Level Scaling
	DamagePerLevel    float64 `json:"damage_per_level"`
	RangePerLevel     float64 `json:"range_per_level"`
	HealthPerLevel    float64 `json:"health_per_level"`
	
	// Matrix Scaling (from power matrix system)
	UsesPowerMatrix   bool    `json:"uses_power_matrix"`
	MatrixMultiplier  float64 `json:"matrix_multiplier"`
	
	// Time Scaling (gets stronger over time)
	TimeScaling       float64 `json:"time_scaling,omitempty"`
	
	// Kill Scaling (gets stronger with kills)
	KillScaling       float64 `json:"kill_scaling,omitempty"`
	
	// Maximum Level
	MaxLevel          int     `json:"max_level"`
}

// TowerRestrictions defines placement and usage restrictions
type TowerRestrictions struct {
	// Placement Restrictions
	TerrainTypes     []string `json:"terrain_types,omitempty"`     // ["ground", "water", "air"]
	PlacementRules   []string `json:"placement_rules,omitempty"`   // ["near_water", "not_near_tower"]
	MinDistance      float64  `json:"min_distance,omitempty"`      // Minimum distance from other towers
	MaxCount         int      `json:"max_count,omitempty"`         // Maximum number per player
	
	// Usage Restrictions
	RequiredLevel    int      `json:"required_level,omitempty"`    // Player level requirement
	RequiredTech     []string `json:"required_tech,omitempty"`     // Technology requirements
	RequiredBuilding []string `json:"required_building,omitempty"` // Building requirements
	
	// Race Restrictions
	AllowedRaces     []string `json:"allowed_races,omitempty"`     // Which races can build this
	ForbiddenRaces   []string `json:"forbidden_races,omitempty"`   // Which races cannot build this
}

// TowerAppearance defines visual and audio properties
type TowerAppearance struct {
	// 3D Model
	ModelPath        string            `json:"model_path"`
	TexturePath      string            `json:"texture_path"`
	AnimationSet     string            `json:"animation_set"`
	Scale            float64           `json:"scale"`
	
	// Effects
	MuzzleFlash      string            `json:"muzzle_flash,omitempty"`
	ProjectileEffect string            `json:"projectile_effect,omitempty"`
	HitEffect        string            `json:"hit_effect,omitempty"`
	AuraEffect       string            `json:"aura_effect,omitempty"`
	
	// Audio
	AttackSound      string            `json:"attack_sound,omitempty"`
	HitSound         string            `json:"hit_sound,omitempty"`
	BuildSound       string            `json:"build_sound,omitempty"`
	UpgradeSound     string            `json:"upgrade_sound,omitempty"`
	
	// UI
	IconPath         string            `json:"icon_path"`
	PortraitPath     string            `json:"portrait_path,omitempty"`
	
	// Colors
	PrimaryColor     string            `json:"primary_color,omitempty"`   // Hex color
	SecondaryColor   string            `json:"secondary_color,omitempty"` // Hex color
	
	// Custom Properties
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

// AssemblyDefinition defines how components are assembled for this tower
type AssemblyDefinition struct {
	// Component Definitions
	Components  []ComponentDefinition  `json:"components"`
	Connections []ConnectionDefinition `json:"connections"`
	
	// Assembly Metadata
	EntryPoints []string              `json:"entry_points,omitempty"`
	ExitPoints  []string              `json:"exit_points,omitempty"`
	
	// Validation Rules
	ValidationRules []string          `json:"validation_rules,omitempty"`
	
	// Performance Hints
	ExpectedTargets    int     `json:"expected_targets,omitempty"`
	ExecutionFrequency float64 `json:"execution_frequency,omitempty"`
}

// ComponentDefinition defines a component within the assembly
type ComponentDefinition struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // component.ComponentType
	Config      map[string]interface{} `json:"config"`
	Position    ComponentPosition      `json:"position,omitempty"` // For visual editor
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectionDefinition defines a connection between components
type ConnectionDefinition struct {
	ID            string `json:"id"`
	FromComponent string `json:"from_component"`
	FromOutput    string `json:"from_output"`
	ToComponent   string `json:"to_component"`
	ToInput       string `json:"to_input"`
	Type          string `json:"type"`     // connection type
	Enabled       bool   `json:"enabled"`
	Priority      string `json:"priority,omitempty"`
}

// ComponentPosition defines position for visual editor
type ComponentPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z,omitempty"`
}

// TowerInstance represents a runtime instance of a tower
type TowerInstance struct {
	// Instance Information
	InstanceID   string    `json:"instance_id"`
	DefinitionID string    `json:"definition_id"`
	OwnerID      string    `json:"owner_id"`
	
	// Game State
	Level        int       `json:"level"`
	Experience   float64   `json:"experience"`
	Position     component.Vector2 `json:"position"`
	Rotation     float64   `json:"rotation"`
	
	// Runtime Stats (modified by upgrades, buffs, etc.)
	CurrentStats TowerStats `json:"current_stats"`
	
	// Status
	Health       float64   `json:"health"`
	Mana         float64   `json:"mana,omitempty"`
	IsActive     bool      `json:"is_active"`
	IsBusy       bool      `json:"is_busy"`
	
	// Combat State
	LastAttackTime time.Time `json:"last_attack_time"`
	CurrentTarget  string    `json:"current_target,omitempty"`
	KillCount      int       `json:"kill_count"`
	DamageDealt    float64   `json:"damage_dealt"`
	
	// Buffs and Debuffs
	ActiveEffects []component.Effect `json:"active_effects,omitempty"`
	
	// Assembly Instance
	AssemblyID   string    `json:"assembly_id"`
	
	// Timestamps
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewTowerDefinition creates a new tower definition with default values
func NewTowerDefinition(name, race string) *TowerDefinition {
	return &TowerDefinition{
		ID:          generateTowerID(),
		Name:        name,
		Version:     "1.0",
		Race:        race,
		Category:    "offensive",
		Tags:        []string{},
		Cost:        TowerCost{SellValue: 0.7},
		Stats:       TowerStats{},
		Scaling:     TowerScaling{MaxLevel: 10, UsesPowerMatrix: true, MatrixMultiplier: 1.0},
		Restrictions: TowerRestrictions{TerrainTypes: []string{"ground"}},
		Appearance:  TowerAppearance{Scale: 1.0},
		Assembly:    AssemblyDefinition{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

// NewTowerInstance creates a new tower instance from a definition
func NewTowerInstance(definitionID, ownerID string, position component.Vector2) *TowerInstance {
	return &TowerInstance{
		InstanceID:     generateTowerInstanceID(),
		DefinitionID:   definitionID,
		OwnerID:        ownerID,
		Level:          1,
		Experience:     0,
		Position:       position,
		Rotation:       0,
		IsActive:       true,
		IsBusy:         false,
		KillCount:      0,
		DamageDealt:    0,
		ActiveEffects:  []component.Effect{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

// Validate validates the tower definition
func (td *TowerDefinition) Validate() error {
	if td.Name == "" {
		return fmt.Errorf("tower name cannot be empty")
	}
	if td.Race == "" {
		return fmt.Errorf("tower race cannot be empty")
	}
	if td.Cost.BuildCost.Gold <= 0 {
		return fmt.Errorf("build cost must be positive")
	}
	if td.Stats.Damage < 0 {
		return fmt.Errorf("damage cannot be negative")
	}
	if td.Stats.Range <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if td.Scaling.MaxLevel <= 0 {
		return fmt.Errorf("max level must be positive")
	}
	return nil
}

// ToJSON converts the tower definition to JSON
func (td *TowerDefinition) ToJSON() ([]byte, error) {
	return json.MarshalIndent(td, "", "  ")
}

// FromJSON creates a tower definition from JSON
func FromJSON(data []byte) (*TowerDefinition, error) {
	var td TowerDefinition
	if err := json.Unmarshal(data, &td); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tower definition: %w", err)
	}
	
	if err := td.Validate(); err != nil {
		return nil, fmt.Errorf("invalid tower definition: %w", err)
	}
	
	return &td, nil
}

// Helper functions
func generateTowerID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}

func generateTowerInstanceID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}
