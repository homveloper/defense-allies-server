// Package impl provides component registration and initialization
package impl

import (
	"fmt"

	"defense-allies-server/pkg/tower/component"
)

// RegisterAllComponents registers all implemented components with the global registry
func RegisterAllComponents() error {
	// Register targeting components
	if err := registerTargetingComponents(); err != nil {
		return fmt.Errorf("failed to register targeting components: %w", err)
	}

	// Register damage components
	if err := registerDamageComponents(); err != nil {
		return fmt.Errorf("failed to register damage components: %w", err)
	}

	// Register effect components
	if err := registerEffectComponents(); err != nil {
		return fmt.Errorf("failed to register effect components: %w", err)
	}

	// Register range components
	if err := registerRangeComponents(); err != nil {
		return fmt.Errorf("failed to register range components: %w", err)
	}

	// Register projectile components
	if err := registerProjectileComponents(); err != nil {
		return fmt.Errorf("failed to register projectile components: %w", err)
	}

	return nil
}

// registerTargetingComponents registers all targeting components
func registerTargetingComponents() error {
	// Single Target Component
	err := component.RegisterComponent(
		component.ComponentTypeSingleTarget,
		NewSingleTargetComponent,
		component.ComponentMetadata{
			Name:        "Single Target",
			Description: "Targets a single enemy based on priority rules",
			Category:    component.CategoryTargeting,
			Version:     "1.0",
			Tags:        []string{"targeting", "single", "priority"},
			Examples: []component.ComponentExample{
				{
					Name:        "Basic single targeting",
					Description: "Target closest enemy within range",
					Config: map[string]interface{}{
						"range":    8.0,
						"priority": "closest",
					},
					Expected: map[string]interface{}{
						"output_type": "single_target",
					},
				},
			},
		},
		&BasicTargetingValidator{},
	)
	if err != nil {
		return err
	}

	// Multi Target Component
	err = component.RegisterComponent(
		component.ComponentTypeMultiTarget,
		NewMultiTargetComponent,
		component.ComponentMetadata{
			Name:        "Multi Target",
			Description: "Targets multiple enemies based on priority rules",
			Category:    component.CategoryTargeting,
			Version:     "1.0",
			Tags:        []string{"targeting", "multi", "priority"},
			Examples: []component.ComponentExample{
				{
					Name:        "Multi target closest",
					Description: "Target up to 3 closest enemies",
					Config: map[string]interface{}{
						"range":       10.0,
						"priority":    "closest",
						"max_targets": 3,
					},
					Expected: map[string]interface{}{
						"output_type": "multiple_targets",
					},
				},
			},
		},
		&BasicTargetingValidator{},
	)
	if err != nil {
		return err
	}

	return nil
}

// registerDamageComponents registers all damage components
func registerDamageComponents() error {
	// Basic Damage Component
	err := component.RegisterComponent(
		component.ComponentTypeBasicDamage,
		NewBasicDamageComponent,
		component.ComponentMetadata{
			Name:        "Basic Damage",
			Description: "Deals basic damage to targets with critical hit support",
			Category:    component.CategoryDamage,
			Version:     "1.0",
			Tags:        []string{"damage", "basic", "critical"},
			Examples: []component.ComponentExample{
				{
					Name:        "Standard damage",
					Description: "100 damage with 10% crit chance",
					Config: map[string]interface{}{
						"base_damage":         100.0,
						"critical_chance":     0.1,
						"critical_multiplier": 2.0,
					},
					Expected: map[string]interface{}{
						"damage_range": "100-200",
					},
				},
			},
		},
		&BasicDamageValidator{},
	)
	if err != nil {
		return err
	}

	// Fire Damage Component
	err = component.RegisterComponent(
		component.ComponentTypeFireDamage,
		NewFireDamageComponent,
		component.ComponentMetadata{
			Name:        "Fire Damage",
			Description: "Deals fire damage and applies burning effect over time",
			Category:    component.CategoryDamage,
			Version:     "1.0",
			Tags:        []string{"damage", "fire", "burn", "dot"},
			Examples: []component.ComponentExample{
				{
					Name:        "Fire attack with burn",
					Description: "120 fire damage + 3 second burn",
					Config: map[string]interface{}{
						"base_damage":   120.0,
						"damage_type":   "fire",
						"burn_duration": 3.0,
						"burn_dps":      25.0,
					},
					Expected: map[string]interface{}{
						"total_damage": "195", // 120 + (25*3)
					},
				},
			},
		},
		&BasicDamageValidator{},
	)
	if err != nil {
		return err
	}

	return nil
}

// registerEffectComponents registers all effect components
func registerEffectComponents() error {
	// Burn Effect Component
	err := component.RegisterComponent(
		component.ComponentTypeBurnEffect,
		NewBurnEffectComponent,
		component.ComponentMetadata{
			Name:        "Burn Effect",
			Description: "Applies burning damage over time to targets",
			Category:    component.CategoryEffect,
			Version:     "1.0",
			Tags:        []string{"effect", "burn", "dot", "fire"},
			Examples: []component.ComponentExample{
				{
					Name:        "Standard burn",
					Description: "5 second burn dealing 20 DPS",
					Config: map[string]interface{}{
						"duration":  5.0,
						"intensity": 20.0,
						"tick_rate": 1.0,
					},
					Expected: map[string]interface{}{
						"total_damage": "100", // 20*5
					},
				},
			},
		},
		&BasicEffectValidator{},
	)
	if err != nil {
		return err
	}

	// Slow Effect Component
	err = component.RegisterComponent(
		component.ComponentTypeSlowEffect,
		NewSlowEffectComponent,
		component.ComponentMetadata{
			Name:        "Slow Effect",
			Description: "Reduces target movement speed for a duration",
			Category:    component.CategoryEffect,
			Version:     "1.0",
			Tags:        []string{"effect", "slow", "debuff", "movement"},
			Examples: []component.ComponentExample{
				{
					Name:        "Standard slow",
					Description: "50% speed reduction for 3 seconds",
					Config: map[string]interface{}{
						"duration":  3.0,
						"intensity": 0.5,
					},
					Expected: map[string]interface{}{
						"speed_reduction": "50%",
					},
				},
			},
		},
		&BasicEffectValidator{},
	)
	if err != nil {
		return err
	}

	return nil
}

// registerRangeComponents registers all range components
func registerRangeComponents() error {
	// Range Check Component
	err := component.RegisterComponent(
		component.ComponentTypeRangeCheck,
		NewRangeCheckComponent,
		component.ComponentMetadata{
			Name:        "Range Check",
			Description: "Filters targets based on range and shape constraints",
			Category:    component.CategoryRange,
			Version:     "1.0",
			Tags:        []string{"range", "filter", "shape", "area"},
			Examples: []component.ComponentExample{
				{
					Name:        "Circular range",
					Description: "8 unit radius circle",
					Config: map[string]interface{}{
						"range": 8.0,
						"shape": "circle",
					},
					Expected: map[string]interface{}{
						"area": "201.06", // π * 8²
					},
				},
			},
		},
		&BasicRangeValidator{},
	)
	if err != nil {
		return err
	}

	// Area of Effect Component
	err = component.RegisterComponent(
		component.ComponentTypeAreaOfEffect,
		NewAreaOfEffectComponent,
		component.ComponentMetadata{
			Name:        "Area of Effect",
			Description: "Applies effects to all targets within a specified area",
			Category:    component.CategoryRange,
			Version:     "1.0",
			Tags:        []string{"range", "area", "aoe", "splash"},
			Examples: []component.ComponentExample{
				{
					Name:        "Explosion area",
					Description: "5 unit radius explosion with linear falloff",
					Config: map[string]interface{}{
						"range":        5.0,
						"shape":        "circle",
						"falloff":      "linear",
						"falloff_rate": 0.8,
					},
					Expected: map[string]interface{}{
						"max_targets": "unlimited",
					},
				},
			},
		},
		&BasicRangeValidator{},
	)
	if err != nil {
		return err
	}

	return nil
}

// registerProjectileComponents registers all projectile components
func registerProjectileComponents() error {
	// Basic Projectile Component
	err := component.RegisterComponent(
		component.ComponentType("projectile"),
		NewProjectileComponent,
		component.ComponentMetadata{
			Name:        "Projectile",
			Description: "Launches projectiles that travel to targets and deal damage",
			Category:    component.CategoryProjectile,
			Version:     "1.0",
			Tags:        []string{"projectile", "travel", "visual"},
			Examples: []component.ComponentExample{
				{
					Name:        "Basic arrow",
					Description: "Simple arrow projectile with medium speed",
					Config: map[string]interface{}{
						"speed":    15.0,
						"lifetime": 2.0,
						"gravity":  0.0,
						"homing":   false,
					},
					Expected: map[string]interface{}{
						"travel_time": "variable",
					},
				},
			},
		},
		&BasicProjectileValidator{},
	)
	if err != nil {
		return err
	}

	return nil
}

// Component validators
type BasicTargetingValidator struct{}

func (btv *BasicTargetingValidator) ValidateConfig(config map[string]interface{}) error {
	if range_, ok := config["range"].(float64); ok && range_ <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if maxTargets, ok := config["max_targets"].(float64); ok && maxTargets <= 0 {
		return fmt.Errorf("max_targets must be positive")
	}
	return nil
}

func (btv *BasicTargetingValidator) ValidateState(comp component.AtomicComponent) error {
	return comp.Validate()
}

type BasicDamageValidator struct{}

func (bdv *BasicDamageValidator) ValidateConfig(config map[string]interface{}) error {
	if damage, ok := config["base_damage"].(float64); ok && damage < 0 {
		return fmt.Errorf("base_damage cannot be negative")
	}
	if crit, ok := config["critical_chance"].(float64); ok && (crit < 0 || crit > 1) {
		return fmt.Errorf("critical_chance must be between 0 and 1")
	}
	return nil
}

func (bdv *BasicDamageValidator) ValidateState(comp component.AtomicComponent) error {
	return comp.Validate()
}

type BasicEffectValidator struct{}

func (bev *BasicEffectValidator) ValidateConfig(config map[string]interface{}) error {
	if duration, ok := config["duration"].(float64); ok && duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if intensity, ok := config["intensity"].(float64); ok && intensity < 0 {
		return fmt.Errorf("intensity cannot be negative")
	}
	return nil
}

func (bev *BasicEffectValidator) ValidateState(comp component.AtomicComponent) error {
	return comp.Validate()
}

type BasicRangeValidator struct{}

func (brv *BasicRangeValidator) ValidateConfig(config map[string]interface{}) error {
	if range_, ok := config["range"].(float64); ok && range_ <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if angle, ok := config["angle"].(float64); ok && (angle <= 0 || angle > 360) {
		return fmt.Errorf("angle must be between 0 and 360 degrees")
	}
	return nil
}

func (brv *BasicRangeValidator) ValidateState(comp component.AtomicComponent) error {
	return comp.Validate()
}

type BasicProjectileValidator struct{}

func (bpv *BasicProjectileValidator) ValidateConfig(config map[string]interface{}) error {
	if speed, ok := config["speed"].(float64); ok && speed <= 0 {
		return fmt.Errorf("speed must be positive")
	}
	if lifetime, ok := config["lifetime"].(float64); ok && lifetime <= 0 {
		return fmt.Errorf("lifetime must be positive")
	}
	return nil
}

func (bpv *BasicProjectileValidator) ValidateState(comp component.AtomicComponent) error {
	return comp.Validate()
}

// GetRegisteredComponentTypes returns all registered component types
func GetRegisteredComponentTypes() []component.ComponentType {
	return component.GetGlobalRegistry().GetComponentTypes()
}

// CreateComponentByType creates a component by its type with the given config
func CreateComponentByType(componentType component.ComponentType, config map[string]interface{}) (component.AtomicComponent, error) {
	return component.CreateComponent(componentType, config)
}

// GetComponentInfo returns information about a registered component type
func GetComponentInfo(componentType component.ComponentType) (*component.ComponentInfo, error) {
	return component.GetGlobalRegistry().GetComponentInfo(componentType)
}
