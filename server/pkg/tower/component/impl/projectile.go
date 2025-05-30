// Package impl provides projectile component implementations
package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"defense-allies-server/pkg/tower/component"
)

// ProjectileComponent represents a projectile that travels to targets
type ProjectileComponent struct {
	id       string
	metadata component.ComponentMetadata
	config   ProjectileConfig
}

// ProjectileConfig defines the configuration for projectile components
type ProjectileConfig struct {
	Speed           float64 `json:"speed"`            // Units per second
	Lifetime        float64 `json:"lifetime"`         // Maximum lifetime in seconds
	Gravity         float64 `json:"gravity"`          // Gravity effect (0 = no gravity)
	Homing          bool    `json:"homing"`           // Whether projectile homes to target
	HomingStrength  float64 `json:"homing_strength"`  // How strongly it homes (0-1)
	Pierce          bool    `json:"pierce"`           // Whether projectile pierces through targets
	MaxPierceCount  int     `json:"max_pierce_count"` // Maximum number of targets to pierce
	EffectPath      string  `json:"effect_path"`      // Visual effect path
	ExplosionEffect string  `json:"explosion_effect"` // Explosion effect on impact
	ExplosionRadius float64 `json:"explosion_radius"` // Explosion radius
}

// NewProjectileComponent creates a new projectile component
func NewProjectileComponent(config map[string]interface{}) (component.AtomicComponent, error) {
	// Parse configuration
	projectileConfig := ProjectileConfig{
		Speed:          15.0, // Default values
		Lifetime:       2.0,
		Gravity:        0.0,
		Homing:         false,
		HomingStrength: 0.0,
		Pierce:         false,
		MaxPierceCount: 0,
	}

	// Override with provided config
	if speed, ok := config["speed"].(float64); ok {
		projectileConfig.Speed = speed
	}
	if lifetime, ok := config["lifetime"].(float64); ok {
		projectileConfig.Lifetime = lifetime
	}
	if gravity, ok := config["gravity"].(float64); ok {
		projectileConfig.Gravity = gravity
	}
	if homing, ok := config["homing"].(bool); ok {
		projectileConfig.Homing = homing
	}
	if homingStrength, ok := config["homing_strength"].(float64); ok {
		projectileConfig.HomingStrength = homingStrength
	}
	if pierce, ok := config["pierce"].(bool); ok {
		projectileConfig.Pierce = pierce
	}
	if maxPierceCount, ok := config["max_pierce_count"].(float64); ok {
		projectileConfig.MaxPierceCount = int(maxPierceCount)
	}
	if effectPath, ok := config["effect_path"].(string); ok {
		projectileConfig.EffectPath = effectPath
	}
	if explosionEffect, ok := config["explosion_effect"].(string); ok {
		projectileConfig.ExplosionEffect = explosionEffect
	}
	if explosionRadius, ok := config["explosion_radius"].(float64); ok {
		projectileConfig.ExplosionRadius = explosionRadius
	}

	return &ProjectileComponent{
		id:     generateProjectileID(),
		config: projectileConfig,
		metadata: component.ComponentMetadata{
			Name:        "Projectile",
			Description: "Launches projectiles that travel to targets",
			Category:    component.CategoryProjectile,
			Version:     "1.0",
			Tags:        []string{"projectile", "visual", "travel"},
		},
	}, nil
}

// GetType returns the component type
func (pc *ProjectileComponent) GetType() component.ComponentType {
	return component.ComponentType("projectile")
}

// GetID returns the component ID
func (pc *ProjectileComponent) GetID() string {
	return pc.id
}

// Execute executes the projectile component
func (pc *ProjectileComponent) Execute(ctx context.Context, execCtx *component.ExecutionContext) (*component.ComponentResult, error) {
	// Get damage events from input
	damageEvents, ok := execCtx.InputData["damage_events"].([]component.Effect)
	if !ok {
		// No damage events to process
		return &component.ComponentResult{
			Success: true,
			Outputs: map[string]interface{}{
				"projectiles": []ProjectileInstance{},
			},
			Effects: []component.Effect{},
			Events:  []component.GameEvent{},
		}, nil
	}

	// Create projectile instances for each damage event
	var projectiles []ProjectileInstance
	var events []component.GameEvent

	for _, damageEvent := range damageEvents {
		projectile := ProjectileInstance{
			ID:              generateProjectileInstanceID(),
			Config:          pc.config,
			SourcePosition:  execCtx.TowerPos,
			TargetID:        damageEvent.Target,
			DamageEvent:     damageEvent,
			CreatedAt:       time.Now(),
			TravelTime:      pc.calculateTravelTime(execCtx.TowerPos, damageEvent.Target),
			IsActive:        true,
		}

		projectiles = append(projectiles, projectile)

		// Create projectile launch event
		events = append(events, component.GameEvent{
			ID:        generateProjectileEventID(),
			Type:      component.EventType("projectile_launched"),
			Source:    pc.id,
			Target:    damageEvent.Target,
			Data: map[string]interface{}{
				"projectile_id": projectile.ID,
				"speed":         pc.config.Speed,
				"effect_path":   pc.config.EffectPath,
			},
			Timestamp: time.Now(),
		})
	}

	return &component.ComponentResult{
		Success: true,
		Outputs: map[string]interface{}{
			"projectiles": projectiles,
		},
		Effects: damageEvents, // Pass through damage events
		Events:  events,
	}, nil
}

// GetInputs returns the component inputs
func (pc *ProjectileComponent) GetInputs() []component.ComponentInput {
	return []component.ComponentInput{
		{
			Name:        "damage_events",
			Type:        component.DataTypeEffects,
			Required:    true,
			Description: "Damage events to create projectiles for",
		},
	}
}

// GetOutputs returns the component outputs
func (pc *ProjectileComponent) GetOutputs() []component.ComponentOutput {
	return []component.ComponentOutput{
		{
			Name:        "projectiles",
			Type:        component.DataTypeArray,
			Description: "Created projectile instances",
		},
	}
}

// CanConnectTo checks if this component can connect to another
func (pc *ProjectileComponent) CanConnectTo(other component.AtomicComponent) bool {
	// Projectiles are typically end-of-chain components
	return false
}

// Validate validates the component configuration
func (pc *ProjectileComponent) Validate() error {
	if pc.config.Speed <= 0 {
		return fmt.Errorf("speed must be positive")
	}
	if pc.config.Lifetime <= 0 {
		return fmt.Errorf("lifetime must be positive")
	}
	if pc.config.HomingStrength < 0 || pc.config.HomingStrength > 1 {
		return fmt.Errorf("homing strength must be between 0 and 1")
	}
	if pc.config.MaxPierceCount < 0 {
		return fmt.Errorf("max pierce count cannot be negative")
	}
	return nil
}

// Clone creates a copy of this component
func (pc *ProjectileComponent) Clone() component.AtomicComponent {
	return &ProjectileComponent{
		id:       generateProjectileID(),
		config:   pc.config,
		metadata: pc.metadata,
	}
}

// GetMetadata returns component metadata
func (pc *ProjectileComponent) GetMetadata() component.ComponentMetadata {
	return pc.metadata
}

// calculateTravelTime calculates how long the projectile will take to reach target
func (pc *ProjectileComponent) calculateTravelTime(sourcePos component.Vector2, targetID string) float64 {
	// For now, return a default travel time
	// In a real implementation, this would calculate based on target position and projectile speed
	return 1.0 / pc.config.Speed
}

// ProjectileInstance represents a runtime instance of a projectile
type ProjectileInstance struct {
	ID              string                `json:"id"`
	Config          ProjectileConfig      `json:"config"`
	SourcePosition  component.Vector2     `json:"source_position"`
	TargetID        string                `json:"target_id"`
	DamageEvent     component.Effect      `json:"damage_event"`
	CreatedAt       time.Time             `json:"created_at"`
	TravelTime      float64               `json:"travel_time"`
	IsActive        bool                  `json:"is_active"`
	CurrentPosition component.Vector2     `json:"current_position"`
}

// Helper functions
func generateProjectileID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Sprintf("projectile_%s", uuid.New().String())
	}
	return fmt.Sprintf("projectile_%s", id.String())
}

func generateProjectileInstanceID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Sprintf("proj_inst_%s", uuid.New().String())
	}
	return fmt.Sprintf("proj_inst_%s", id.String())
}

func generateProjectileEventID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Sprintf("proj_event_%s", uuid.New().String())
	}
	return fmt.Sprintf("proj_event_%s", id.String())
}
