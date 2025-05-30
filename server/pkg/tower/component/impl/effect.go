// Package impl provides effect component implementations
package impl

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/pkg/tower/component"
)

// EffectConfig holds configuration for effect components
type EffectConfig struct {
	EffectType    component.EffectType `json:"effect_type"`
	Duration      float64              `json:"duration"`
	Intensity     float64              `json:"intensity"`
	TickRate      float64              `json:"tick_rate"`      // For DoT effects (seconds between ticks)
	MaxStacks     int                  `json:"max_stacks"`     // Maximum number of stacks
	StackBehavior StackBehavior        `json:"stack_behavior"` // How stacks interact
	Dispellable   bool                 `json:"dispellable"`    // Can be removed by dispel effects
	Conditions    []EffectCondition    `json:"conditions"`     // Conditions for applying effect
	Modifiers     []EffectModifier     `json:"modifiers"`      // Additional effect modifiers
}

// StackBehavior defines how multiple instances of the same effect interact
type StackBehavior string

const (
	StackBehaviorReplace     StackBehavior = "replace"     // New effect replaces old one
	StackBehaviorRefresh     StackBehavior = "refresh"     // New effect refreshes duration
	StackBehaviorStack       StackBehavior = "stack"       // Effects stack (intensity adds up)
	StackBehaviorIndependent StackBehavior = "independent" // Effects are independent
)

// EffectCondition defines when an effect should be applied
type EffectCondition struct {
	Type     ConditionType `json:"type"`
	Value    interface{}   `json:"value"`
	Operator string        `json:"operator"`
}

// ConditionType defines what to check for effect application
type ConditionType string

const (
	ConditionTypeTargetHealth    ConditionType = "target_health"
	ConditionTypeTargetType      ConditionType = "target_type"
	ConditionTypeRandomChance    ConditionType = "random_chance"
	ConditionTypeDamageDealt     ConditionType = "damage_dealt"
	ConditionTypeTargetHasEffect ConditionType = "target_has_effect"
	ConditionTypeCustom          ConditionType = "custom"
)

// EffectModifier modifies effect properties based on conditions
type EffectModifier struct {
	Condition           string  `json:"condition"`
	DurationMultiplier  float64 `json:"duration_multiplier"`
	IntensityMultiplier float64 `json:"intensity_multiplier"`
	TickRateMultiplier  float64 `json:"tick_rate_multiplier"`
}

// BaseEffectComponent provides common effect functionality
type BaseEffectComponent struct {
	id       string
	config   EffectConfig
	metadata component.ComponentMetadata
}

// BurnEffectComponent applies burning damage over time
type BurnEffectComponent struct {
	BaseEffectComponent
}

// NewBurnEffectComponent creates a new burn effect component
func NewBurnEffectComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	effectConfig, err := parseEffectConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid effect config: %w", err)
	}

	// Override effect type to burn
	effectConfig.EffectType = component.EffectTypeBurn

	// Set default values for burn effect
	if effectConfig.Duration == 0 {
		effectConfig.Duration = 5.0 // 5 seconds
	}
	if effectConfig.TickRate == 0 {
		effectConfig.TickRate = 1.0 // 1 second between ticks
	}
	if effectConfig.MaxStacks == 0 {
		effectConfig.MaxStacks = 3 // Max 3 stacks
	}
	if effectConfig.StackBehavior == "" {
		effectConfig.StackBehavior = StackBehaviorStack
	}

	return &BurnEffectComponent{
		BaseEffectComponent: BaseEffectComponent{
			id:     generateComponentID("burn_effect"),
			config: effectConfig,
			metadata: component.ComponentMetadata{
				Name:        "Burn Effect",
				Description: "Applies burning damage over time to targets",
				Category:    component.CategoryEffect,
				Version:     "1.0",
			},
		},
	}, nil
}

func (bec *BurnEffectComponent) GetType() component.ComponentType {
	return component.ComponentTypeBurnEffect
}

func (bec *BurnEffectComponent) GetID() string {
	return bec.id
}

func (bec *BurnEffectComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	targets := bec.getTargetsFromInput(execCtx.InputData)
	if len(targets) == 0 {
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"effects": []component.Effect{},
			},
		}, nil
	}

	var effects []component.Effect

	for _, target := range targets {
		// Check conditions
		if !bec.shouldApplyEffect(target, execCtx) {
			continue
		}

		// Create burn effect
		effect := component.Effect{
			ID:        generateComponentID("burn"),
			Type:      bec.config.EffectType,
			Target:    target.GetID(),
			Duration:  bec.config.Duration,
			Intensity: bec.config.Intensity,
			Data: map[string]interface{}{
				"tick_rate":      bec.config.TickRate,
				"max_stacks":     bec.config.MaxStacks,
				"stack_behavior": bec.config.StackBehavior,
				"dispellable":    bec.config.Dispellable,
				"damage_type":    "fire",
			},
			Source:    bec.id,
			Timestamp: time.Now(),
		}

		effects = append(effects, effect)
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"effects": effects,
		},
		Events: []component.GameEvent{
			{
				ID:        generateComponentID("effects_applied"),
				Type:      component.EventTypeEffectApplied,
				Source:    bec.id,
				Data:      map[string]interface{}{"effect_count": len(effects)},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

func (bec *BurnEffectComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "targets",
			Type:        component.DataTypeTargets,
			Required:    true,
			Description: "Targets to apply burn effect to",
		},
		{
			Name:        "damage_dealt",
			Type:        component.DataTypeFloat,
			Required:    false,
			Description: "Amount of damage dealt (for intensity scaling)",
		},
	}
}

func (bec *BurnEffectComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "effects",
			Type:        component.DataTypeEffects,
			Description: "Burn effects to apply to targets",
		},
	}
}

func (bec *BurnEffectComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeEffects, input.Type) {
			return true
		}
	}
	return false
}

func (bec *BurnEffectComponent) Validate() error {
	if bec.config.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if bec.config.Intensity < 0 {
		return fmt.Errorf("intensity cannot be negative")
	}
	if bec.config.TickRate <= 0 {
		return fmt.Errorf("tick rate must be positive")
	}
	return nil
}

func (bec *BurnEffectComponent) Clone() component.AtomicComponent {
	clone := *bec
	clone.id = generateComponentID("burn_effect")
	return &clone
}

func (bec *BurnEffectComponent) GetMetadata() component.ComponentMetadata {
	return bec.metadata
}

// SlowEffectComponent applies movement speed reduction
type SlowEffectComponent struct {
	BaseEffectComponent
}

// NewSlowEffectComponent creates a new slow effect component
func NewSlowEffectComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	effectConfig, err := parseEffectConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid effect config: %w", err)
	}

	// Override effect type to slow
	effectConfig.EffectType = component.EffectTypeSlow

	// Set default values for slow effect
	if effectConfig.Duration == 0 {
		effectConfig.Duration = 3.0 // 3 seconds
	}
	if effectConfig.Intensity == 0 {
		effectConfig.Intensity = 0.5 // 50% speed reduction
	}
	if effectConfig.StackBehavior == "" {
		effectConfig.StackBehavior = StackBehaviorReplace
	}

	return &SlowEffectComponent{
		BaseEffectComponent: BaseEffectComponent{
			id:     generateComponentID("slow_effect"),
			config: effectConfig,
			metadata: component.ComponentMetadata{
				Name:        "Slow Effect",
				Description: "Reduces target movement speed",
				Category:    component.CategoryEffect,
				Version:     "1.0",
			},
		},
	}, nil
}

func (sec *SlowEffectComponent) GetType() component.ComponentType {
	return component.ComponentTypeSlowEffect
}

func (sec *SlowEffectComponent) GetID() string {
	return sec.id
}

func (sec *SlowEffectComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	targets := sec.getTargetsFromInput(execCtx.InputData)
	if len(targets) == 0 {
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"effects": []component.Effect{},
			},
		}, nil
	}

	var effects []component.Effect

	for _, target := range targets {
		if !sec.shouldApplyEffect(target, execCtx) {
			continue
		}

		effect := component.Effect{
			ID:        generateComponentID("slow"),
			Type:      sec.config.EffectType,
			Target:    target.GetID(),
			Duration:  sec.config.Duration,
			Intensity: sec.config.Intensity,
			Data: map[string]interface{}{
				"speed_reduction": sec.config.Intensity,
				"stack_behavior":  sec.config.StackBehavior,
				"dispellable":     sec.config.Dispellable,
			},
			Source:    sec.id,
			Timestamp: time.Now(),
		}

		effects = append(effects, effect)
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"effects": effects,
		},
	}, nil
}

func (sec *SlowEffectComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "targets",
			Type:        component.DataTypeTargets,
			Required:    true,
			Description: "Targets to apply slow effect to",
		},
	}
}

func (sec *SlowEffectComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "effects",
			Type:        component.DataTypeEffects,
			Description: "Slow effects to apply to targets",
		},
	}
}

func (sec *SlowEffectComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeEffects, input.Type) {
			return true
		}
	}
	return false
}

func (sec *SlowEffectComponent) Validate() error {
	if sec.config.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if sec.config.Intensity < 0 || sec.config.Intensity > 1 {
		return fmt.Errorf("slow intensity must be between 0 and 1")
	}
	return nil
}

func (sec *SlowEffectComponent) Clone() component.AtomicComponent {
	clone := *sec
	clone.id = generateComponentID("slow_effect")
	return &clone
}

func (sec *SlowEffectComponent) GetMetadata() component.ComponentMetadata {
	return sec.metadata
}

// Helper methods for BaseEffectComponent
func (bec *BaseEffectComponent) getTargetsFromInput(inputData map[string]interface{}) []component.Enemy {
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

func (bec *BaseEffectComponent) shouldApplyEffect(target component.Enemy, execCtx *component.ExecutionContext) bool {
	// Check all conditions
	for _, condition := range bec.config.Conditions {
		if !bec.evaluateCondition(condition, target, execCtx) {
			return false
		}
	}
	return true
}

func (bec *BaseEffectComponent) evaluateCondition(condition EffectCondition, target component.Enemy, execCtx *component.ExecutionContext) bool {
	switch condition.Type {
	case ConditionTypeRandomChance:
		if _, ok := condition.Value.(float64); ok {
			// TODO: Use proper random number generator
			return true // Placeholder - always apply for now
		}
	case ConditionTypeTargetType:
		if targetType, ok := condition.Value.(string); ok {
			return target.GetType() == targetType
		}
	case ConditionTypeTargetHealth:
		// TODO: Implement health condition checking
		return true
	default:
		return true
	}
	return false
}

// Helper function to parse effect configuration
func parseEffectConfig(config map[string]interface{}) (EffectConfig, error) {
	result := EffectConfig{
		Duration:      3.0,
		Intensity:     1.0,
		TickRate:      1.0,
		MaxStacks:     1,
		StackBehavior: StackBehaviorReplace,
		Dispellable:   true,
	}

	if val, ok := config["duration"].(float64); ok {
		result.Duration = val
	}

	if val, ok := config["intensity"].(float64); ok {
		result.Intensity = val
	}

	if val, ok := config["tick_rate"].(float64); ok {
		result.TickRate = val
	}

	if val, ok := config["max_stacks"].(float64); ok {
		result.MaxStacks = int(val)
	}

	if val, ok := config["stack_behavior"].(string); ok {
		result.StackBehavior = StackBehavior(val)
	}

	if val, ok := config["dispellable"].(bool); ok {
		result.Dispellable = val
	}

	return result, nil
}
