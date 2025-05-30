// Package component provides the core interfaces and types for the modular tower system.
// This package implements a LEGO-style component system where tower abilities can be
// composed from atomic components that can be connected together.
package component

import (
	"context"
	"time"
)

// AtomicComponent represents the smallest unit of tower functionality.
// Components can be connected together like LEGO blocks to create complex tower behaviors.
type AtomicComponent interface {
	// GetType returns the component type identifier
	GetType() ComponentType
	
	// GetID returns the unique identifier for this component instance
	GetID() string
	
	// Execute processes the component logic with the given context
	Execute(ctx context.Context, execCtx *ExecutionContext) (*ComponentResult, error)
	
	// GetInputs returns the list of input ports this component accepts
	GetInputs() []ComponentInput
	
	// GetOutputs returns the list of output ports this component provides
	GetOutputs() []ComponentOutput
	
	// CanConnectTo checks if this component can connect to another component
	CanConnectTo(other AtomicComponent) bool
	
	// Validate checks if the component configuration is valid
	Validate() error
	
	// Clone creates a deep copy of this component
	Clone() AtomicComponent
	
	// GetMetadata returns component metadata for UI and documentation
	GetMetadata() ComponentMetadata
}

// ComponentInput represents an input port that accepts data from other components
type ComponentInput struct {
	Name        string      `json:"name"`
	Type        DataType    `json:"type"`
	Required    bool        `json:"required"`
	Description string      `json:"description"`
	DefaultValue interface{} `json:"default_value,omitempty"`
}

// ComponentOutput represents an output port that provides data to other components
type ComponentOutput struct {
	Name        string   `json:"name"`
	Type        DataType `json:"type"`
	Description string   `json:"description"`
}

// ComponentResult contains the output data from component execution
type ComponentResult struct {
	Success    bool                   `json:"success"`
	Outputs    map[string]interface{} `json:"outputs"`
	Effects    []Effect               `json:"effects,omitempty"`
	Events     []GameEvent            `json:"events,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionContext provides the runtime context for component execution
type ExecutionContext struct {
	// Game state
	GameTime    time.Time              `json:"game_time"`
	DeltaTime   float64                `json:"delta_time"`
	GameState   GameState              `json:"game_state"`
	
	// Tower context
	TowerID     string                 `json:"tower_id"`
	TowerPos    Vector2                `json:"tower_position"`
	OwnerID     string                 `json:"owner_id"`
	
	// Input data from connected components
	InputData   map[string]interface{} `json:"input_data"`
	
	// Environment context
	Environment EnvironmentState       `json:"environment"`
	
	// Matrix context for balancing
	PowerMatrix Matrix                 `json:"power_matrix"`
	
	// Execution metadata
	ExecutionID string                 `json:"execution_id"`
	TraceID     string                 `json:"trace_id"`
}

// ComponentMetadata provides information about the component for UI and documentation
type ComponentMetadata struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Category     ComponentCategory `json:"category"`
	Version      string            `json:"version"`
	Author       string            `json:"author,omitempty"`
	Tags         []string          `json:"tags,omitempty"`
	Icon         string            `json:"icon,omitempty"`
	Color        string            `json:"color,omitempty"`
	
	// UI hints
	UIHints      UIHints           `json:"ui_hints,omitempty"`
	
	// Documentation
	Examples     []ComponentExample `json:"examples,omitempty"`
	Documentation string           `json:"documentation,omitempty"`
}

// UIHints provides hints for UI rendering
type UIHints struct {
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Resizable   bool   `json:"resizable,omitempty"`
	Collapsible bool   `json:"collapsible,omitempty"`
	PreferredPosition string `json:"preferred_position,omitempty"` // "top", "bottom", "left", "right"
}

// ComponentExample provides usage examples for documentation
type ComponentExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	Inputs      map[string]interface{} `json:"inputs"`
	Expected    map[string]interface{} `json:"expected"`
}

// Effect represents a game effect that can be applied to targets
type Effect struct {
	ID          string                 `json:"id"`
	Type        EffectType             `json:"type"`
	Target      string                 `json:"target"`
	Duration    float64                `json:"duration,omitempty"`
	Intensity   float64                `json:"intensity,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
}

// GameEvent represents an event that occurred during component execution
type GameEvent struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// GameState represents the current state of the game
type GameState struct {
	Towers      map[string]Tower       `json:"towers"`
	Enemies     map[string]Enemy       `json:"enemies"`
	Projectiles map[string]Projectile  `json:"projectiles"`
	Players     map[string]Player      `json:"players"`
	Wave        WaveState              `json:"wave"`
	Resources   map[string]float64     `json:"resources"`
}

// EnvironmentState represents the current environment conditions
type EnvironmentState struct {
	Time        TimeState              `json:"time"`
	Weather     WeatherState           `json:"weather"`
	Terrain     TerrainState           `json:"terrain"`
	Events      []EnvironmentEvent     `json:"events"`
	Modifiers   map[string]float64     `json:"modifiers"`
}

// Vector2 represents a 2D position or vector
type Vector2 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Matrix represents a mathematical matrix for balancing calculations
type Matrix struct {
	Rows int         `json:"rows"`
	Cols int         `json:"cols"`
	Data [][]float64 `json:"data"`
}

// Basic game entity interfaces (to be expanded in other packages)
type Tower interface {
	GetID() string
	GetPosition() Vector2
	GetOwner() string
}

type Enemy interface {
	GetID() string
	GetPosition() Vector2
	GetType() string
}

type Projectile interface {
	GetID() string
	GetPosition() Vector2
	GetTarget() string
}

type Player interface {
	GetID() string
	GetName() string
	GetRace() string
}

// State structures (to be expanded in other packages)
type WaveState struct {
	Number    int     `json:"number"`
	Progress  float64 `json:"progress"`
	Remaining int     `json:"remaining"`
}

type TimeState struct {
	Period    string  `json:"period"`    // "day", "night", "dawn", "dusk"
	Intensity float64 `json:"intensity"` // 0.0 to 1.0
}

type WeatherState struct {
	Type      string  `json:"type"`      // "clear", "rain", "storm", "snow"
	Intensity float64 `json:"intensity"` // 0.0 to 1.0
}

type TerrainState struct {
	Type      string  `json:"type"`      // "forest", "mountain", "desert", "swamp"
	Elevation float64 `json:"elevation"` // relative elevation
}

type EnvironmentEvent struct {
	Type      string                 `json:"type"`
	Duration  float64                `json:"duration"`
	Intensity float64                `json:"intensity"`
	Data      map[string]interface{} `json:"data"`
}
