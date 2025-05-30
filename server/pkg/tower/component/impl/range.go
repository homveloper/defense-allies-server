// Package impl provides range component implementations
package impl

import (
	"context"
	"fmt"
	"math"
	"time"

	"defense-allies-server/pkg/tower/component"
)

// RangeConfig holds configuration for range components
type RangeConfig struct {
	Range       float64     `json:"range"`
	Shape       RangeShape  `json:"shape"`
	Angle       float64     `json:"angle"`        // For cone shapes (degrees)
	Width       float64     `json:"width"`        // For line shapes
	InnerRadius float64     `json:"inner_radius"` // For donut shapes
	Falloff     FalloffType `json:"falloff"`      // Damage/effect falloff
	FalloffRate float64     `json:"falloff_rate"` // Rate of falloff
}

// RangeShape defines the shape of the range area
type RangeShape string

const (
	RangeShapeCircle    RangeShape = "circle"
	RangeShapeCone      RangeShape = "cone"
	RangeShapeLine      RangeShape = "line"
	RangeShapeRectangle RangeShape = "rectangle"
	RangeShapeDonut     RangeShape = "donut"
	RangeShapeCustom    RangeShape = "custom"
)

// FalloffType defines how effects diminish with distance
type FalloffType string

const (
	FalloffTypeNone        FalloffType = "none"        // No falloff
	FalloffTypeLinear      FalloffType = "linear"      // Linear decrease
	FalloffTypeQuadratic   FalloffType = "quadratic"   // Quadratic decrease
	FalloffTypeExponential FalloffType = "exponential" // Exponential decrease
	FalloffTypeStep        FalloffType = "step"        // Step function
)

// AreaInfo represents information about targets within an area
type AreaInfo struct {
	Position     component.Vector2 `json:"position"`
	Targets      []TargetInArea    `json:"targets"`
	TotalTargets int               `json:"total_targets"`
	AreaSize     float64           `json:"area_size"`
}

// TargetInArea represents a target within the area with distance information
type TargetInArea struct {
	Enemy      component.Enemy `json:"enemy"`
	Distance   float64         `json:"distance"`
	Multiplier float64         `json:"multiplier"` // Effect multiplier based on distance
	InRange    bool            `json:"in_range"`
}

// BaseRangeComponent provides common range functionality
type BaseRangeComponent struct {
	id       string
	config   RangeConfig
	metadata component.ComponentMetadata
}

// RangeCheckComponent checks if targets are within range
type RangeCheckComponent struct {
	BaseRangeComponent
}

// NewRangeCheckComponent creates a new range check component
func NewRangeCheckComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	rangeConfig, err := parseRangeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid range config: %w", err)
	}

	return &RangeCheckComponent{
		BaseRangeComponent: BaseRangeComponent{
			id:     generateComponentID("range_check"),
			config: rangeConfig,
			metadata: component.ComponentMetadata{
				Name:        "Range Check",
				Description: "Filters targets based on range and shape",
				Category:    component.CategoryRange,
				Version:     "1.0",
			},
		},
	}, nil
}

func (rcc *RangeCheckComponent) GetType() component.ComponentType {
	return component.ComponentTypeRangeCheck
}

func (rcc *RangeCheckComponent) GetID() string {
	return rcc.id
}

func (rcc *RangeCheckComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	// Get inputs
	sourcePos, ok := execCtx.InputData["source_position"].(component.Vector2)
	if !ok {
		return nil, fmt.Errorf("source_position is required")
	}

	targets := rcc.getTargetsFromInput(execCtx.InputData)
	if len(targets) == 0 {
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"targets_in_range": []component.Enemy{},
				"area_info":        AreaInfo{Position: sourcePos, Targets: []TargetInArea{}, TotalTargets: 0},
			},
		}, nil
	}

	// Filter targets by range and shape
	var targetsInRange []component.Enemy
	var targetsInArea []TargetInArea

	for _, target := range targets {
		targetPos := target.GetPosition()
		distance := calculateDistance(sourcePos, targetPos)

		inRange := rcc.isInRange(sourcePos, targetPos, distance)
		multiplier := rcc.calculateMultiplier(distance)

		targetInArea := TargetInArea{
			Enemy:      target,
			Distance:   distance,
			Multiplier: multiplier,
			InRange:    inRange,
		}

		targetsInArea = append(targetsInArea, targetInArea)

		if inRange {
			targetsInRange = append(targetsInRange, target)
		}
	}

	areaInfo := AreaInfo{
		Position:     sourcePos,
		Targets:      targetsInArea,
		TotalTargets: len(targetsInRange),
		AreaSize:     rcc.calculateAreaSize(),
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"targets_in_range": targetsInRange,
			"area_info":        areaInfo,
		},
	}, nil
}

func (rcc *RangeCheckComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "source_position",
			Type:        component.DataTypeVector2,
			Required:    true,
			Description: "Position to check range from",
		},
		{
			Name:        "targets",
			Type:        component.DataTypeTargets,
			Required:    true,
			Description: "Targets to check range for",
		},
		{
			Name:        "direction",
			Type:        component.DataTypeVector2,
			Required:    false,
			Description: "Direction for cone/line shapes",
		},
	}
}

func (rcc *RangeCheckComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "targets_in_range",
			Type:        component.DataTypeTargets,
			Description: "Targets within the specified range",
		},
		{
			Name:        "area_info",
			Type:        component.DataTypeObject,
			Description: "Detailed information about the area and targets",
		},
	}
}

func (rcc *RangeCheckComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeTargets, input.Type) ||
			component.IsCompatible(component.DataTypeObject, input.Type) {
			return true
		}
	}
	return false
}

func (rcc *RangeCheckComponent) Validate() error {
	if rcc.config.Range <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if rcc.config.Shape == RangeShapeCone && (rcc.config.Angle <= 0 || rcc.config.Angle > 360) {
		return fmt.Errorf("cone angle must be between 0 and 360 degrees")
	}
	if rcc.config.Shape == RangeShapeDonut && rcc.config.InnerRadius >= rcc.config.Range {
		return fmt.Errorf("inner radius must be less than outer radius")
	}
	return nil
}

func (rcc *RangeCheckComponent) Clone() component.AtomicComponent {
	clone := *rcc
	clone.id = generateComponentID("range_check")
	return &clone
}

func (rcc *RangeCheckComponent) GetMetadata() component.ComponentMetadata {
	return rcc.metadata
}

// AreaOfEffectComponent creates area effects
type AreaOfEffectComponent struct {
	BaseRangeComponent
}

// NewAreaOfEffectComponent creates a new area of effect component
func NewAreaOfEffectComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	rangeConfig, err := parseRangeConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid range config: %w", err)
	}

	// Default to circle shape for AoE
	if rangeConfig.Shape == "" {
		rangeConfig.Shape = RangeShapeCircle
	}

	return &AreaOfEffectComponent{
		BaseRangeComponent: BaseRangeComponent{
			id:     generateComponentID("area_of_effect"),
			config: rangeConfig,
			metadata: component.ComponentMetadata{
				Name:        "Area of Effect",
				Description: "Applies effects to targets in an area",
				Category:    component.CategoryRange,
				Version:     "1.0",
			},
		},
	}, nil
}

func (aoe *AreaOfEffectComponent) GetType() component.ComponentType {
	return component.ComponentTypeAreaOfEffect
}

func (aoe *AreaOfEffectComponent) GetID() string {
	return aoe.id
}

func (aoe *AreaOfEffectComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	// Get center position (could be target position or specified position)
	var centerPos component.Vector2

	if pos, ok := execCtx.InputData["center_position"].(component.Vector2); ok {
		centerPos = pos
	} else if target, ok := execCtx.InputData["center_target"].(component.Enemy); ok {
		centerPos = target.GetPosition()
	} else {
		return nil, fmt.Errorf("center_position or center_target is required")
	}

	// Get all available targets
	allTargets := aoe.getTargetsFromInput(execCtx.InputData)
	if len(allTargets) == 0 {
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"affected_targets": []component.Enemy{},
				"area_info":        AreaInfo{Position: centerPos, Targets: []TargetInArea{}, TotalTargets: 0},
			},
		}, nil
	}

	// Find targets in area
	var affectedTargets []component.Enemy
	var targetsInArea []TargetInArea

	for _, target := range allTargets {
		targetPos := target.GetPosition()
		distance := calculateDistance(centerPos, targetPos)

		inRange := aoe.isInRange(centerPos, targetPos, distance)
		multiplier := aoe.calculateMultiplier(distance)

		targetInArea := TargetInArea{
			Enemy:      target,
			Distance:   distance,
			Multiplier: multiplier,
			InRange:    inRange,
		}

		targetsInArea = append(targetsInArea, targetInArea)

		if inRange {
			affectedTargets = append(affectedTargets, target)
		}
	}

	areaInfo := AreaInfo{
		Position:     centerPos,
		Targets:      targetsInArea,
		TotalTargets: len(affectedTargets),
		AreaSize:     aoe.calculateAreaSize(),
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"affected_targets": affectedTargets,
			"area_info":        areaInfo,
		},
		Events: []component.GameEvent{
			{
				ID:        generateComponentID("area_effect"),
				Type:      component.EventTypeAbilityUsed,
				Source:    aoe.id,
				Data:      map[string]interface{}{"targets_affected": len(affectedTargets)},
				Timestamp: time.Now(),
			},
		},
	}, nil
}

func (aoe *AreaOfEffectComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "center_position",
			Type:        component.DataTypeVector2,
			Required:    false,
			Description: "Center position for the area effect",
		},
		{
			Name:        "center_target",
			Type:        component.DataTypeTarget,
			Required:    false,
			Description: "Target to center the area effect on",
		},
		{
			Name:        "available_targets",
			Type:        component.DataTypeTargets,
			Required:    false, // Made optional - will use ExecutionContext if not provided
			Description: "All available targets to check",
		},
	}
}

func (aoe *AreaOfEffectComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "affected_targets",
			Type:        component.DataTypeTargets,
			Description: "Targets affected by the area effect",
		},
		{
			Name:        "area_info",
			Type:        component.DataTypeObject,
			Description: "Information about the area and affected targets",
		},
	}
}

func (aoe *AreaOfEffectComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeTargets, input.Type) {
			return true
		}
	}
	return false
}

func (aoe *AreaOfEffectComponent) Validate() error {
	if aoe.config.Range <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if aoe.config.Shape == RangeShapeCone && (aoe.config.Angle <= 0 || aoe.config.Angle > 360) {
		return fmt.Errorf("cone angle must be between 0 and 360 degrees")
	}
	if aoe.config.Shape == RangeShapeDonut && aoe.config.InnerRadius >= aoe.config.Range {
		return fmt.Errorf("inner radius must be less than outer radius")
	}
	return nil
}

func (aoe *AreaOfEffectComponent) Clone() component.AtomicComponent {
	clone := *aoe
	clone.id = generateComponentID("area_of_effect")
	return &clone
}

func (aoe *AreaOfEffectComponent) GetMetadata() component.ComponentMetadata {
	return aoe.metadata
}

// Helper methods for BaseRangeComponent
func (brc *BaseRangeComponent) getTargetsFromInput(inputData map[string]interface{}) []component.Enemy {
	// Try different input names
	if targets, ok := inputData["targets"].([]component.Enemy); ok {
		return targets
	}
	if targets, ok := inputData["available_targets"].([]component.Enemy); ok {
		return targets
	}
	if target, ok := inputData["target"].(component.Enemy); ok {
		return []component.Enemy{target}
	}
	return nil
}

func (brc *BaseRangeComponent) isInRange(sourcePos, targetPos component.Vector2, distance float64) bool {
	switch brc.config.Shape {
	case RangeShapeCircle:
		return distance <= brc.config.Range
	case RangeShapeDonut:
		return distance <= brc.config.Range && distance >= brc.config.InnerRadius
	case RangeShapeCone:
		// TODO: Implement cone shape calculation
		return distance <= brc.config.Range
	case RangeShapeLine:
		// TODO: Implement line shape calculation
		return distance <= brc.config.Range
	case RangeShapeRectangle:
		// TODO: Implement rectangle shape calculation
		return distance <= brc.config.Range
	default:
		return distance <= brc.config.Range
	}
}

func (brc *BaseRangeComponent) calculateMultiplier(distance float64) float64 {
	if distance > brc.config.Range {
		return 0.0
	}

	switch brc.config.Falloff {
	case FalloffTypeNone:
		return 1.0
	case FalloffTypeLinear:
		return 1.0 - (distance/brc.config.Range)*brc.config.FalloffRate
	case FalloffTypeQuadratic:
		ratio := distance / brc.config.Range
		return 1.0 - ratio*ratio*brc.config.FalloffRate
	case FalloffTypeExponential:
		return math.Exp(-distance * brc.config.FalloffRate / brc.config.Range)
	default:
		return 1.0
	}
}

func (brc *BaseRangeComponent) calculateAreaSize() float64 {
	switch brc.config.Shape {
	case RangeShapeCircle:
		return math.Pi * brc.config.Range * brc.config.Range
	case RangeShapeDonut:
		outerArea := math.Pi * brc.config.Range * brc.config.Range
		innerArea := math.Pi * brc.config.InnerRadius * brc.config.InnerRadius
		return outerArea - innerArea
	case RangeShapeCone:
		return 0.5 * brc.config.Range * brc.config.Range * (brc.config.Angle * math.Pi / 180)
	case RangeShapeRectangle:
		return brc.config.Range * brc.config.Width
	default:
		return brc.config.Range * brc.config.Range
	}
}

// Helper function to parse range configuration
func parseRangeConfig(config map[string]interface{}) (RangeConfig, error) {
	result := RangeConfig{
		Range:       5.0,
		Shape:       RangeShapeCircle,
		Angle:       90.0,
		Width:       2.0,
		InnerRadius: 0.0,
		Falloff:     FalloffTypeNone,
		FalloffRate: 1.0,
	}

	if val, ok := config["range"].(float64); ok {
		result.Range = val
	}

	if val, ok := config["shape"].(string); ok {
		result.Shape = RangeShape(val)
	}

	if val, ok := config["angle"].(float64); ok {
		result.Angle = val
	}

	if val, ok := config["width"].(float64); ok {
		result.Width = val
	}

	if val, ok := config["inner_radius"].(float64); ok {
		result.InnerRadius = val
	}

	if val, ok := config["falloff"].(string); ok {
		result.Falloff = FalloffType(val)
	}

	if val, ok := config["falloff_rate"].(float64); ok {
		result.FalloffRate = val
	}

	return result, nil
}
