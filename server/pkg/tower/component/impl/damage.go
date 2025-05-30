// Package impl provides damage component implementations
package impl

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"defense-allies-server/pkg/tower/component"
)

// DamageConfig holds configuration for damage components
type DamageConfig struct {
	BaseDamage         float64          `json:"base_damage"`
	DamageType         DamageType       `json:"damage_type"`
	CriticalChance     float64          `json:"critical_chance"`
	CriticalMultiplier float64          `json:"critical_multiplier"`
	Variance           float64          `json:"variance"` // Random damage variance (0.0 to 1.0)
	Scaling            DamageScaling    `json:"scaling"`
	Modifiers          []DamageModifier `json:"modifiers"`
	IgnoreArmor        bool             `json:"ignore_armor"`
	ArmorPenetration   float64          `json:"armor_penetration"`
}

// DamageType represents different types of damage
type DamageType string

const (
	DamageTypePhysical DamageType = "physical"
	DamageTypeFire     DamageType = "fire"
	DamageTypeIce      DamageType = "ice"
	DamageTypeElectric DamageType = "electric"
	DamageTypePoison   DamageType = "poison"
	DamageTypeMagical  DamageType = "magical"
	DamageTypeTrue     DamageType = "true"
	DamageTypePercent  DamageType = "percent"
)

// DamageScaling defines how damage scales with various factors
type DamageScaling struct {
	PowerMatrix bool    `json:"power_matrix"` // Use tower's power matrix
	TowerLevel  float64 `json:"tower_level"`  // Damage per tower level
	GameTime    float64 `json:"game_time"`    // Damage per minute
	EnemyHealth float64 `json:"enemy_health"` // Damage based on enemy health %
	Distance    float64 `json:"distance"`     // Damage based on distance
}

// DamageModifier applies conditional damage modifications
type DamageModifier struct {
	Condition  string  `json:"condition"`  // Condition expression
	Multiplier float64 `json:"multiplier"` // Damage multiplier when condition is true
	Additive   float64 `json:"additive"`   // Flat damage bonus when condition is true
	Priority   int     `json:"priority"`   // Application order
}

// DamageResult contains the result of damage calculation
type DamageResult struct {
	FinalDamage    float64                `json:"final_damage"`
	BaseDamage     float64                `json:"base_damage"`
	DamageType     DamageType             `json:"damage_type"`
	IsCritical     bool                   `json:"is_critical"`
	Modifiers      []string               `json:"modifiers"`     // Applied modifier names
	Effectiveness  float64                `json:"effectiveness"` // Damage effectiveness (0.0 to 2.0+)
	ArmorReduction float64                `json:"armor_reduction"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// BaseDamageComponent provides common damage functionality
type BaseDamageComponent struct {
	id       string
	config   DamageConfig
	metadata component.ComponentMetadata
	rng      *rand.Rand
}

// BasicDamageComponent deals basic damage
type BasicDamageComponent struct {
	BaseDamageComponent
}

// NewBasicDamageComponent creates a new basic damage component
func NewBasicDamageComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	damageConfig, err := parseDamageConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid damage config: %w", err)
	}

	return &BasicDamageComponent{
		BaseDamageComponent: BaseDamageComponent{
			id:     generateComponentID("basic_damage"),
			config: damageConfig,
			metadata: component.ComponentMetadata{
				Name:        "Basic Damage",
				Description: "Deals basic damage to targets",
				Category:    component.CategoryDamage,
				Version:     "1.0",
			},
			rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		},
	}, nil
}

func (bdc *BasicDamageComponent) GetType() component.ComponentType {
	return component.ComponentTypeBasicDamage
}

func (bdc *BasicDamageComponent) GetID() string {
	return bdc.id
}

func (bdc *BasicDamageComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	// Get targets from input
	targets := bdc.getTargetsFromInput(execCtx.InputData)
	if len(targets) == 0 {
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"damage_events": []component.Effect{},
			},
		}, nil
	}

	var damageEvents []component.Effect

	// Calculate damage for each target
	for _, target := range targets {
		damageResult := bdc.calculateDamage(target, execCtx)

		// Create damage effect
		effect := component.Effect{
			ID:        generateComponentID("damage_effect"),
			Type:      component.EffectTypeDamage,
			Target:    target.GetID(),
			Intensity: damageResult.FinalDamage,
			Data: map[string]interface{}{
				"damage_type":     damageResult.DamageType,
				"is_critical":     damageResult.IsCritical,
				"effectiveness":   damageResult.Effectiveness,
				"armor_reduction": damageResult.ArmorReduction,
				"base_damage":     damageResult.BaseDamage,
			},
			Source:    bdc.id,
			Timestamp: time.Now(),
		}

		damageEvents = append(damageEvents, effect)
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"damage_events": damageEvents,
		},
		Events: []component.GameEvent{
			{
				ID:        generateComponentID("damage_dealt"),
				Type:      component.EventTypeDamageDealt,
				Source:    bdc.id,
				Data:      map[string]interface{}{"damage_count": len(damageEvents)},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

func (bdc *BasicDamageComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "targets",
			Type:        component.DataTypeTargets,
			Required:    false, // Made optional - can work without targets for support towers
			Description: "Targets to deal damage to",
		},
		{
			Name:        "power_matrix",
			Type:        component.DataTypeMatrix,
			Required:    false,
			Description: "Tower's power matrix for scaling",
		},
		{
			Name:         "tower_level",
			Type:         component.DataTypeInt,
			Required:     false,
			Description:  "Tower level for scaling",
			DefaultValue: 1,
		},
	}
}

func (bdc *BasicDamageComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "damage_events",
			Type:        component.DataTypeEffects,
			Description: "Damage effects to apply to targets",
		},
	}
}

func (bdc *BasicDamageComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeEffects, input.Type) {
			return true
		}
	}
	return false
}

func (bdc *BasicDamageComponent) Validate() error {
	if bdc.config.BaseDamage < 0 {
		return fmt.Errorf("base damage cannot be negative")
	}
	if bdc.config.CriticalChance < 0 || bdc.config.CriticalChance > 1 {
		return fmt.Errorf("critical chance must be between 0 and 1")
	}
	if bdc.config.CriticalMultiplier < 1 {
		return fmt.Errorf("critical multiplier must be >= 1")
	}
	return nil
}

func (bdc *BasicDamageComponent) Clone() component.AtomicComponent {
	clone := *bdc
	clone.id = generateComponentID("basic_damage")
	clone.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &clone
}

func (bdc *BasicDamageComponent) GetMetadata() component.ComponentMetadata {
	return bdc.metadata
}

// FireDamageComponent deals fire damage with burning effects
type FireDamageComponent struct {
	BaseDamageComponent
	burnDuration float64
	burnDPS      float64
}

// NewFireDamageComponent creates a new fire damage component
func NewFireDamageComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	damageConfig, err := parseDamageConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid damage config: %w", err)
	}

	// Override damage type to fire
	damageConfig.DamageType = DamageTypeFire

	burnDuration := 3.0
	if val, ok := config["burn_duration"].(float64); ok {
		burnDuration = val
	}

	burnDPS := damageConfig.BaseDamage * 0.2 // 20% of base damage per second
	if val, ok := config["burn_dps"].(float64); ok {
		burnDPS = val
	}

	return &FireDamageComponent{
		BaseDamageComponent: BaseDamageComponent{
			id:     generateComponentID("fire_damage"),
			config: damageConfig,
			metadata: component.ComponentMetadata{
				Name:        "Fire Damage",
				Description: "Deals fire damage and applies burning effect",
				Category:    component.CategoryDamage,
				Version:     "1.0",
			},
			rng: rand.New(rand.NewSource(time.Now().UnixNano())),
		},
		burnDuration: burnDuration,
		burnDPS:      burnDPS,
	}, nil
}

func (fdc *FireDamageComponent) GetType() component.ComponentType {
	return component.ComponentTypeFireDamage
}

func (fdc *FireDamageComponent) GetID() string {
	return fdc.id
}

func (fdc *FireDamageComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	// Get targets from input
	targets := fdc.getTargetsFromInput(execCtx.InputData)
	if len(targets) == 0 {
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"damage_events": []component.Effect{},
				"burn_effects":  []component.Effect{},
			},
		}, nil
	}

	var damageEvents []component.Effect

	// Calculate damage for each target
	for _, target := range targets {
		damageResult := fdc.calculateDamage(target, execCtx)

		// Create damage effect
		effect := component.Effect{
			ID:        generateComponentID("damage_effect"),
			Type:      component.EffectTypeDamage,
			Target:    target.GetID(),
			Intensity: damageResult.FinalDamage,
			Data: map[string]interface{}{
				"damage_type":     damageResult.DamageType,
				"is_critical":     damageResult.IsCritical,
				"effectiveness":   damageResult.Effectiveness,
				"armor_reduction": damageResult.ArmorReduction,
				"base_damage":     damageResult.BaseDamage,
			},
			Source:    fdc.id,
			Timestamp: time.Now(),
		}

		damageEvents = append(damageEvents, effect)
	}

	// Add burn effects
	var burnEffects []component.Effect

	for _, target := range targets {
		// Create burn effect
		burnEffect := component.Effect{
			ID:        generateComponentID("burn_effect"),
			Type:      component.EffectTypeBurn,
			Target:    target.GetID(),
			Duration:  fdc.burnDuration,
			Intensity: fdc.burnDPS,
			Data: map[string]interface{}{
				"damage_per_second": fdc.burnDPS,
				"damage_type":       DamageTypeFire,
			},
			Source:    fdc.id,
			Timestamp: time.Now(),
		}

		burnEffects = append(burnEffects, burnEffect)
	}

	// Combine damage and burn effects
	allEffects := append(damageEvents, burnEffects...)

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"damage_events": allEffects,
			"burn_effects":  burnEffects,
		},
		Events: []component.GameEvent{
			{
				ID:        generateComponentID("fire_damage_dealt"),
				Type:      component.EventTypeDamageDealt,
				Source:    fdc.id,
				Data:      map[string]interface{}{"damage_count": len(damageEvents), "burn_count": len(burnEffects)},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

func (fdc *FireDamageComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "targets",
			Type:        component.DataTypeTargets,
			Required:    true,
			Description: "Targets to deal fire damage to",
		},
		{
			Name:        "power_matrix",
			Type:        component.DataTypeMatrix,
			Required:    false,
			Description: "Tower's power matrix for scaling",
		},
		{
			Name:         "tower_level",
			Type:         component.DataTypeInt,
			Required:     false,
			Description:  "Tower level for scaling",
			DefaultValue: 1,
		},
	}
}

func (fdc *FireDamageComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "damage_events",
			Type:        component.DataTypeEffects,
			Description: "Fire damage effects to apply to targets",
		},
		{
			Name:        "burn_effects",
			Type:        component.DataTypeEffects,
			Description: "Burn effects to apply to targets",
		},
	}
}

func (fdc *FireDamageComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeEffects, input.Type) {
			return true
		}
	}
	return false
}

func (fdc *FireDamageComponent) Validate() error {
	if fdc.config.BaseDamage < 0 {
		return fmt.Errorf("base damage cannot be negative")
	}
	if fdc.config.CriticalChance < 0 || fdc.config.CriticalChance > 1 {
		return fmt.Errorf("critical chance must be between 0 and 1")
	}
	if fdc.config.CriticalMultiplier < 1 {
		return fmt.Errorf("critical multiplier must be >= 1")
	}
	if fdc.burnDuration <= 0 {
		return fmt.Errorf("burn duration must be positive")
	}
	if fdc.burnDPS < 0 {
		return fmt.Errorf("burn DPS cannot be negative")
	}
	return nil
}

func (fdc *FireDamageComponent) Clone() component.AtomicComponent {
	clone := *fdc
	clone.id = generateComponentID("fire_damage")
	clone.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	return &clone
}

func (fdc *FireDamageComponent) GetMetadata() component.ComponentMetadata {
	return fdc.metadata
}

// Helper methods for BaseDamageComponent
func (bdc *BaseDamageComponent) getTargetsFromInput(inputData map[string]interface{}) []component.Enemy {
	// Try to get targets as array
	if targets, ok := inputData["targets"].([]component.Enemy); ok {
		return targets
	}

	// Try to get single target
	if target, ok := inputData["target"].(component.Enemy); ok {
		return []component.Enemy{target}
	}

	return nil
}

func (bdc *BaseDamageComponent) calculateDamage(target component.Enemy, execCtx *component.ExecutionContext) DamageResult {
	result := DamageResult{
		BaseDamage:     bdc.config.BaseDamage,
		DamageType:     bdc.config.DamageType,
		IsCritical:     false,
		Effectiveness:  1.0,
		ArmorReduction: 0.0,
		Metadata:       make(map[string]interface{}),
	}

	// Apply scaling
	scaledDamage := bdc.applyScaling(result.BaseDamage, target, execCtx)

	// Apply variance
	if bdc.config.Variance > 0 {
		variance := bdc.config.Variance * scaledDamage
		scaledDamage += (bdc.rng.Float64()*2 - 1) * variance
	}

	// Check for critical hit
	if bdc.rng.Float64() < bdc.config.CriticalChance {
		result.IsCritical = true
		scaledDamage *= bdc.config.CriticalMultiplier
	}

	// Apply modifiers
	scaledDamage = bdc.applyModifiers(scaledDamage, target, execCtx)

	// Apply armor reduction (if not ignored)
	if !bdc.config.IgnoreArmor {
		scaledDamage = bdc.applyArmorReduction(scaledDamage, target)
	}

	result.FinalDamage = math.Max(0, scaledDamage) // Ensure non-negative damage

	return result
}

func (bdc *BaseDamageComponent) applyScaling(baseDamage float64, target component.Enemy, execCtx *component.ExecutionContext) float64 {
	damage := baseDamage

	// Power matrix scaling
	if bdc.config.Scaling.PowerMatrix {
		if matrix, ok := execCtx.InputData["power_matrix"].(component.Matrix); ok {
			// Use matrix[0][0] as damage multiplier (offensive individual power)
			if len(matrix.Data) > 0 && len(matrix.Data[0]) > 0 {
				damage *= matrix.Data[0][0]
			}
		}
	}

	// Tower level scaling
	if bdc.config.Scaling.TowerLevel > 0 {
		if level, ok := execCtx.InputData["tower_level"].(int); ok {
			damage += float64(level-1) * bdc.config.Scaling.TowerLevel
		}
	}

	// Game time scaling
	if bdc.config.Scaling.GameTime > 0 {
		gameMinutes := execCtx.GameTime.Sub(time.Time{}).Minutes()
		damage += gameMinutes * bdc.config.Scaling.GameTime
	}

	return damage
}

func (bdc *BaseDamageComponent) applyModifiers(damage float64, target component.Enemy, execCtx *component.ExecutionContext) float64 {
	// TODO: Implement condition evaluation for modifiers
	// For now, just return the damage unchanged
	return damage
}

func (bdc *BaseDamageComponent) applyArmorReduction(damage float64, target component.Enemy) float64 {
	// TODO: Get target armor value and apply reduction formula
	// For now, just return damage unchanged
	return damage
}

// Helper function to parse damage configuration
func parseDamageConfig(config map[string]interface{}) (DamageConfig, error) {
	result := DamageConfig{
		BaseDamage:         100.0,
		DamageType:         DamageTypePhysical,
		CriticalChance:     0.1,
		CriticalMultiplier: 2.0,
		Variance:           0.1,
		IgnoreArmor:        false,
		ArmorPenetration:   0.0,
	}

	if val, ok := config["base_damage"].(float64); ok {
		result.BaseDamage = val
	}

	if val, ok := config["damage_type"].(string); ok {
		result.DamageType = DamageType(val)
	}

	if val, ok := config["critical_chance"].(float64); ok {
		result.CriticalChance = val
	}

	if val, ok := config["critical_multiplier"].(float64); ok {
		result.CriticalMultiplier = val
	}

	if val, ok := config["variance"].(float64); ok {
		result.Variance = val
	}

	if val, ok := config["ignore_armor"].(bool); ok {
		result.IgnoreArmor = val
	}

	if val, ok := config["armor_penetration"].(float64); ok {
		result.ArmorPenetration = val
	}

	return result, nil
}
