// Package impl provides concrete implementations of atomic components
// for the Defense Allies tower system.
package impl

import (
	"context"
	"fmt"
	"math"
	"sort"

	"defense-allies-server/pkg/tower/component"

	"github.com/google/uuid"
)

// BaseTargetingComponent provides common functionality for all targeting components
type BaseTargetingComponent struct {
	id       string
	config   TargetingConfig
	metadata component.ComponentMetadata
}

// TargetingConfig holds configuration for targeting components
type TargetingConfig struct {
	Range        float64           `json:"range"`
	Priority     TargetingPriority `json:"priority"`
	MaxTargets   int               `json:"max_targets"`
	TargetTypes  []string          `json:"target_types"`
	ExcludeTypes []string          `json:"exclude_types"`
	RequireLOS   bool              `json:"require_line_of_sight"`
	Filters      []TargetingFilter `json:"filters"`
}

// TargetingPriority defines how targets should be prioritized
type TargetingPriority string

const (
	PriorityClosest   TargetingPriority = "closest"
	PriorityFarthest  TargetingPriority = "farthest"
	PriorityWeakest   TargetingPriority = "weakest"
	PriorityStrongest TargetingPriority = "strongest"
	PriorityFirst     TargetingPriority = "first"
	PriorityLast      TargetingPriority = "last"
	PriorityRandom    TargetingPriority = "random"
	PriorityCustom    TargetingPriority = "custom"
)

// TargetingFilter provides additional filtering criteria
type TargetingFilter struct {
	Type      FilterType  `json:"type"`
	Value     interface{} `json:"value"`
	Operator  string      `json:"operator"`  // "eq", "gt", "lt", "gte", "lte", "contains"
	Condition string      `json:"condition"` // "and", "or", "not"
}

// FilterType defines what property to filter on
type FilterType string

const (
	FilterTypeHealth    FilterType = "health"
	FilterTypeHealthPct FilterType = "health_percent"
	FilterTypeSpeed     FilterType = "speed"
	FilterTypeArmor     FilterType = "armor"
	FilterTypeSize      FilterType = "size"
	FilterTypeDistance  FilterType = "distance"
	FilterTypeEnemyType FilterType = "enemy_type"
	FilterTypeHasEffect FilterType = "has_effect"
	FilterTypeCustom    FilterType = "custom"
)

// TargetInfo represents information about a potential target
type TargetInfo struct {
	Enemy    component.Enemy `json:"enemy"`
	Distance float64         `json:"distance"`
	Priority float64         `json:"priority"`
	Valid    bool            `json:"valid"`
	Reason   string          `json:"reason,omitempty"`
}

// SingleTargetComponent targets a single enemy
type SingleTargetComponent struct {
	BaseTargetingComponent
}

// NewSingleTargetComponent creates a new single target component
func NewSingleTargetComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	targetingConfig, err := parseTargetingConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid targeting config: %w", err)
	}

	// Ensure max_targets is 1 for single targeting
	targetingConfig.MaxTargets = 1

	return &SingleTargetComponent{
		BaseTargetingComponent: BaseTargetingComponent{
			id:     generateComponentID("single_target"),
			config: targetingConfig,
			metadata: component.ComponentMetadata{
				Name:        "Single Target",
				Description: "Targets a single enemy based on priority",
				Category:    component.CategoryTargeting,
				Version:     "1.0",
			},
		},
	}, nil
}

func (stc *SingleTargetComponent) GetType() component.ComponentType {
	return component.ComponentTypeSingleTarget
}

func (stc *SingleTargetComponent) GetID() string {
	return stc.id
}

func (stc *SingleTargetComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	targets := stc.findTargets(execCtx)

	var selectedTarget component.Enemy
	if len(targets) > 0 {
		selectedTarget = targets[0].Enemy
	}

	result := &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"target": selectedTarget,
		},
	}

	if selectedTarget == nil {
		result.Outputs["target"] = nil
	}

	return result, nil
}

func (stc *SingleTargetComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "tower_position",
			Type:        component.DataTypeVector2,
			Required:    false, // Made optional - will use ExecutionContext.TowerPos if not provided
			Description: "Position of the tower for range calculation",
		},
		{
			Name:        "available_enemies",
			Type:        component.DataTypeEnemies,
			Required:    false, // Made optional - will use ExecutionContext.InputData if not provided
			Description: "List of available enemies to target",
		},
	}
}

func (stc *SingleTargetComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "target",
			Type:        component.DataTypeTarget,
			Description: "Selected target enemy",
		},
	}
}

func (stc *SingleTargetComponent) CanConnectTo(other component.AtomicComponent) bool {
	// Can connect to any component that accepts target input
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeTarget, input.Type) {
			return true
		}
	}
	return false
}

func (stc *SingleTargetComponent) Validate() error {
	if stc.config.Range <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if stc.config.MaxTargets != 1 {
		return fmt.Errorf("single target component must have max_targets = 1")
	}
	return nil
}

func (stc *SingleTargetComponent) Clone() component.AtomicComponent {
	clone := *stc
	clone.id = generateComponentID("single_target")
	return &clone
}

func (stc *SingleTargetComponent) GetMetadata() component.ComponentMetadata {
	return stc.metadata
}

// MultiTargetComponent targets multiple enemies
type MultiTargetComponent struct {
	BaseTargetingComponent
}

// NewMultiTargetComponent creates a new multi target component
func NewMultiTargetComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	targetingConfig, err := parseTargetingConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid targeting config: %w", err)
	}

	// Default max targets if not specified
	if targetingConfig.MaxTargets <= 0 {
		targetingConfig.MaxTargets = 3
	}

	return &MultiTargetComponent{
		BaseTargetingComponent: BaseTargetingComponent{
			id:     generateComponentID("multi_target"),
			config: targetingConfig,
			metadata: component.ComponentMetadata{
				Name:        "Multi Target",
				Description: "Targets multiple enemies based on priority",
				Category:    component.CategoryTargeting,
				Version:     "1.0",
			},
		},
	}, nil
}

func (mtc *MultiTargetComponent) GetType() component.ComponentType {
	return component.ComponentTypeMultiTarget
}

func (mtc *MultiTargetComponent) GetID() string {
	return mtc.id
}

func (mtc *MultiTargetComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	targets := mtc.findTargets(execCtx)

	// Limit to max targets
	if len(targets) > mtc.config.MaxTargets {
		targets = targets[:mtc.config.MaxTargets]
	}

	// Extract enemies from target info
	enemies := make([]component.Enemy, len(targets))
	for i, target := range targets {
		enemies[i] = target.Enemy
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"targets": enemies,
		},
	}, nil
}

func (mtc *MultiTargetComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "tower_position",
			Type:        component.DataTypeVector2,
			Required:    true,
			Description: "Position of the tower for range calculation",
		},
		{
			Name:        "available_enemies",
			Type:        component.DataTypeEnemies,
			Required:    true,
			Description: "List of available enemies to target",
		},
	}
}

func (mtc *MultiTargetComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "targets",
			Type:        component.DataTypeTargets,
			Description: "List of selected target enemies",
		},
	}
}

func (mtc *MultiTargetComponent) CanConnectTo(other component.AtomicComponent) bool {
	for _, input := range other.GetInputs() {
		if component.IsCompatible(component.DataTypeTargets, input.Type) {
			return true
		}
	}
	return false
}

func (mtc *MultiTargetComponent) Validate() error {
	if mtc.config.Range <= 0 {
		return fmt.Errorf("range must be positive")
	}
	if mtc.config.MaxTargets <= 0 {
		return fmt.Errorf("max_targets must be positive")
	}
	return nil
}

func (mtc *MultiTargetComponent) Clone() component.AtomicComponent {
	clone := *mtc
	clone.id = generateComponentID("multi_target")
	return &clone
}

func (mtc *MultiTargetComponent) GetMetadata() component.ComponentMetadata {
	return mtc.metadata
}

// Common targeting logic for all targeting components
func (btc *BaseTargetingComponent) findTargets(execCtx *component.ExecutionContext) []TargetInfo {
	// Get tower position (from input or ExecutionContext)
	var towerPos component.Vector2
	if pos, ok := execCtx.InputData["tower_position"].(component.Vector2); ok {
		towerPos = pos
	} else {
		towerPos = execCtx.TowerPos // Use ExecutionContext position as fallback
	}

	// Get enemies (from input or ExecutionContext)
	var enemies []component.Enemy
	if enemyList, ok := execCtx.InputData["available_enemies"].([]component.Enemy); ok {
		enemies = enemyList
	} else if enemyList, ok := execCtx.InputData["enemies"].([]component.Enemy); ok {
		enemies = enemyList
	} else {
		return nil // No enemies available
	}

	var candidates []TargetInfo

	// Filter and evaluate each enemy
	for _, enemy := range enemies {
		targetInfo := TargetInfo{
			Enemy:    enemy,
			Distance: calculateDistance(towerPos, enemy.GetPosition()),
			Valid:    true,
		}

		// Range check
		if targetInfo.Distance > btc.config.Range {
			targetInfo.Valid = false
			targetInfo.Reason = "out_of_range"
			continue
		}

		// Type filtering
		if !btc.isValidTargetType(enemy) {
			targetInfo.Valid = false
			targetInfo.Reason = "invalid_type"
			continue
		}

		// Apply custom filters
		if !btc.passesFilters(enemy, targetInfo.Distance) {
			targetInfo.Valid = false
			targetInfo.Reason = "failed_filter"
			continue
		}

		// Line of sight check
		if btc.config.RequireLOS && !btc.hasLineOfSight(towerPos, enemy.GetPosition(), execCtx) {
			targetInfo.Valid = false
			targetInfo.Reason = "no_line_of_sight"
			continue
		}

		// Calculate priority
		targetInfo.Priority = btc.calculatePriority(enemy, targetInfo.Distance)

		if targetInfo.Valid {
			candidates = append(candidates, targetInfo)
		}
	}

	// Sort by priority
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Priority > candidates[j].Priority
	})

	return candidates
}

// Helper functions
func parseTargetingConfig(config map[string]interface{}) (TargetingConfig, error) {
	result := TargetingConfig{
		Range:      8.0,
		Priority:   PriorityClosest,
		MaxTargets: 1,
		RequireLOS: false,
	}

	if val, ok := config["range"].(float64); ok {
		result.Range = val
	}

	if val, ok := config["priority"].(string); ok {
		result.Priority = TargetingPriority(val)
	}

	if val, ok := config["max_targets"].(float64); ok {
		result.MaxTargets = int(val)
	}

	if val, ok := config["require_line_of_sight"].(bool); ok {
		result.RequireLOS = val
	}

	// Parse target types
	if val, ok := config["target_types"].([]interface{}); ok {
		for _, t := range val {
			if str, ok := t.(string); ok {
				result.TargetTypes = append(result.TargetTypes, str)
			}
		}
	}

	return result, nil
}

func generateComponentID(prefix string) string {
	// Use UUIDv7 for time-ordered unique IDs
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to UUIDv4 if v7 fails
		return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
	}
	return fmt.Sprintf("%s_%s", prefix, id.String())
}

func calculateDistance(pos1, pos2 component.Vector2) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func (btc *BaseTargetingComponent) isValidTargetType(enemy component.Enemy) bool {
	enemyType := enemy.GetType()

	// Check exclude list first
	for _, excludeType := range btc.config.ExcludeTypes {
		if enemyType == excludeType {
			return false
		}
	}

	// If no target types specified, accept all (except excluded)
	if len(btc.config.TargetTypes) == 0 {
		return true
	}

	// Check if enemy type is in allowed list
	for _, allowedType := range btc.config.TargetTypes {
		if enemyType == allowedType {
			return true
		}
	}

	return false
}

func (btc *BaseTargetingComponent) passesFilters(enemy component.Enemy, distance float64) bool {
	// TODO: Implement filter logic based on TargetingFilter
	// This would check health, speed, armor, etc.
	return true
}

func (btc *BaseTargetingComponent) hasLineOfSight(from, to component.Vector2, execCtx *component.ExecutionContext) bool {
	// TODO: Implement line of sight calculation
	// This would check for obstacles between tower and target
	return true
}

func (btc *BaseTargetingComponent) calculatePriority(enemy component.Enemy, distance float64) float64 {
	switch btc.config.Priority {
	case PriorityClosest:
		return 10000.0 - distance // Closer = higher priority (increased base value)
	case PriorityFarthest:
		return distance // Farther = higher priority
	case PriorityWeakest:
		// TODO: Get enemy health and return inverse
		return 1000.0 // Placeholder
	case PriorityStrongest:
		// TODO: Get enemy health
		return 1.0 // Placeholder
	default:
		return 1.0
	}
}
