// Package assembly provides the component assembly system for creating
// complex tower behaviors by connecting atomic components together.
package assembly

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/tower/component"

	"github.com/google/uuid"
)

// ComponentAssembly represents a collection of connected components that form a complete tower behavior
type ComponentAssembly struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`

	// Components and connections
	Components  map[string]component.AtomicComponent `json:"components"`
	Connections []component.ComponentConnection      `json:"connections"`

	// Assembly metadata
	Metadata AssemblyMetadata `json:"metadata"`

	// Execution information
	EntryPoints []string `json:"entry_points"` // Components with no inputs
	ExitPoints  []string `json:"exit_points"`  // Components with no outputs

	// Validation state
	IsValid  bool              `json:"is_valid"`
	Errors   []AssemblyError   `json:"errors,omitempty"`
	Warnings []AssemblyWarning `json:"warnings,omitempty"`

	// Runtime state
	ExecutionOrder []string `json:"execution_order"` // Topologically sorted component IDs

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AssemblyMetadata provides additional information about the assembly
type AssemblyMetadata struct {
	Author      string           `json:"author,omitempty"`
	Tags        []string         `json:"tags,omitempty"`
	Category    string           `json:"category,omitempty"`
	Difficulty  int              `json:"difficulty,omitempty"` // 1-10 complexity rating
	Performance PerformanceHints `json:"performance,omitempty"`

	// Tower-specific metadata
	Race      string `json:"race,omitempty"`
	TowerType string `json:"tower_type,omitempty"`
	MinLevel  int    `json:"min_level,omitempty"`
	MaxLevel  int    `json:"max_level,omitempty"`

	// Custom properties
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// PerformanceHints provides hints for optimizing assembly execution
type PerformanceHints struct {
	ExpectedTargets    int     `json:"expected_targets"`    // Expected number of targets
	ExecutionFrequency float64 `json:"execution_frequency"` // Expected executions per second
	MemoryUsage        string  `json:"memory_usage"`        // "low", "medium", "high"
	CPUIntensity       string  `json:"cpu_intensity"`       // "low", "medium", "high"
	Cacheable          bool    `json:"cacheable"`           // Whether results can be cached
}

// AssemblyError represents an error in the assembly
type AssemblyError struct {
	Type         ErrorType              `json:"type"`
	Message      string                 `json:"message"`
	ComponentID  string                 `json:"component_id,omitempty"`
	ConnectionID string                 `json:"connection_id,omitempty"`
	Severity     Severity               `json:"severity"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// AssemblyWarning represents a warning in the assembly
type AssemblyWarning struct {
	Type         WarningType `json:"type"`
	Message      string      `json:"message"`
	ComponentID  string      `json:"component_id,omitempty"`
	ConnectionID string      `json:"connection_id,omitempty"`
	Suggestion   string      `json:"suggestion,omitempty"`
	Timestamp    time.Time   `json:"timestamp"`
}

// ErrorType represents different types of assembly errors
type ErrorType string

const (
	ErrorTypeInvalidConnection    ErrorType = "invalid_connection"
	ErrorTypeMissingComponent     ErrorType = "missing_component"
	ErrorTypeCircularDependency   ErrorType = "circular_dependency"
	ErrorTypeIncompatibleTypes    ErrorType = "incompatible_types"
	ErrorTypeValidationFailed     ErrorType = "validation_failed"
	ErrorTypeMissingRequiredInput ErrorType = "missing_required_input"
	ErrorTypeOrphanedComponent    ErrorType = "orphaned_component"
	ErrorTypeInvalidConfiguration ErrorType = "invalid_configuration"
)

// WarningType represents different types of assembly warnings
type WarningType string

const (
	WarningTypePerformance          WarningType = "performance"
	WarningTypeUnusedOutput         WarningType = "unused_output"
	WarningTypeSuboptimalPath       WarningType = "suboptimal_path"
	WarningTypeDeprecatedComponent  WarningType = "deprecated_component"
	WarningTypeMissingOptionalInput WarningType = "missing_optional_input"
	WarningTypeComplexity           WarningType = "complexity"
)

// Severity represents the severity level of errors and warnings
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// AssemblyBuilder provides a fluent interface for building component assemblies
type AssemblyBuilder struct {
	assembly *ComponentAssembly
}

// NewComponentAssembly creates a new empty component assembly
func NewComponentAssembly(name string) *ComponentAssembly {
	return &ComponentAssembly{
		ID:          generateAssemblyID(),
		Name:        name,
		Version:     "1.0",
		Components:  make(map[string]component.AtomicComponent),
		Connections: []component.ComponentConnection{},
		Metadata:    AssemblyMetadata{},
		EntryPoints: []string{},
		ExitPoints:  []string{},
		IsValid:     false,
		Errors:      []AssemblyError{},
		Warnings:    []AssemblyWarning{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// NewAssemblyBuilder creates a new assembly builder
func NewAssemblyBuilder(name string) *AssemblyBuilder {
	return &AssemblyBuilder{
		assembly: NewComponentAssembly(name),
	}
}

// AddComponent adds a component to the assembly
func (ca *ComponentAssembly) AddComponent(comp component.AtomicComponent) error {
	if comp == nil {
		return fmt.Errorf("component cannot be nil")
	}

	componentID := comp.GetID()
	if componentID == "" {
		return fmt.Errorf("component must have a valid ID")
	}

	if _, exists := ca.Components[componentID]; exists {
		return fmt.Errorf("component with ID %s already exists", componentID)
	}

	// Validate the component
	if err := comp.Validate(); err != nil {
		return fmt.Errorf("component validation failed: %w", err)
	}

	ca.Components[componentID] = comp
	ca.UpdatedAt = time.Now()
	ca.IsValid = false // Mark as needing revalidation

	return nil
}

// RemoveComponent removes a component from the assembly
func (ca *ComponentAssembly) RemoveComponent(componentID string) error {
	if _, exists := ca.Components[componentID]; !exists {
		return fmt.Errorf("component with ID %s does not exist", componentID)
	}

	// Remove all connections involving this component
	var newConnections []component.ComponentConnection
	for _, conn := range ca.Connections {
		if conn.FromComponent != componentID && conn.ToComponent != componentID {
			newConnections = append(newConnections, conn)
		}
	}
	ca.Connections = newConnections

	// Remove the component
	delete(ca.Components, componentID)
	ca.UpdatedAt = time.Now()
	ca.IsValid = false

	return nil
}

// AddConnection adds a connection between components
func (ca *ComponentAssembly) AddConnection(conn component.ComponentConnection) error {
	// Validate that both components exist
	fromComp, fromExists := ca.Components[conn.FromComponent]
	toComp, toExists := ca.Components[conn.ToComponent]

	if !fromExists {
		return fmt.Errorf("from component %s does not exist", conn.FromComponent)
	}
	if !toExists {
		return fmt.Errorf("to component %s does not exist", conn.ToComponent)
	}

	// Validate that the connection is valid
	if !fromComp.CanConnectTo(toComp) {
		return fmt.Errorf("component %s cannot connect to %s", conn.FromComponent, conn.ToComponent)
	}

	// Check for duplicate connections
	for _, existing := range ca.Connections {
		if existing.FromComponent == conn.FromComponent &&
			existing.FromOutput == conn.FromOutput &&
			existing.ToComponent == conn.ToComponent &&
			existing.ToInput == conn.ToInput {
			return fmt.Errorf("connection already exists")
		}
	}

	ca.Connections = append(ca.Connections, conn)
	ca.UpdatedAt = time.Now()
	ca.IsValid = false

	return nil
}

// RemoveConnection removes a connection from the assembly
func (ca *ComponentAssembly) RemoveConnection(connectionID string) error {
	for i, conn := range ca.Connections {
		if conn.ID == connectionID {
			ca.Connections = append(ca.Connections[:i], ca.Connections[i+1:]...)
			ca.UpdatedAt = time.Now()
			ca.IsValid = false
			return nil
		}
	}
	return fmt.Errorf("connection with ID %s does not exist", connectionID)
}

// GetComponent returns a component by ID
func (ca *ComponentAssembly) GetComponent(componentID string) (component.AtomicComponent, bool) {
	comp, exists := ca.Components[componentID]
	return comp, exists
}

// GetConnections returns all connections for a component
func (ca *ComponentAssembly) GetConnections(componentID string) []component.ComponentConnection {
	var connections []component.ComponentConnection
	for _, conn := range ca.Connections {
		if conn.FromComponent == componentID || conn.ToComponent == componentID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// GetInputConnections returns all input connections for a component
func (ca *ComponentAssembly) GetInputConnections(componentID string) []component.ComponentConnection {
	var connections []component.ComponentConnection
	for _, conn := range ca.Connections {
		if conn.ToComponent == componentID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// GetOutputConnections returns all output connections for a component
func (ca *ComponentAssembly) GetOutputConnections(componentID string) []component.ComponentConnection {
	var connections []component.ComponentConnection
	for _, conn := range ca.Connections {
		if conn.FromComponent == componentID {
			connections = append(connections, conn)
		}
	}
	return connections
}

// Clone creates a deep copy of the assembly
func (ca *ComponentAssembly) Clone() *ComponentAssembly {
	clone := &ComponentAssembly{
		ID:          generateAssemblyID(),
		Name:        ca.Name + " (Copy)",
		Description: ca.Description,
		Version:     ca.Version,
		Components:  make(map[string]component.AtomicComponent),
		Connections: []component.ComponentConnection{},
		Metadata:    ca.Metadata,
		EntryPoints: []string{},
		ExitPoints:  []string{},
		IsValid:     false, // Clone needs revalidation
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Map old component IDs to new component IDs
	idMapping := make(map[string]string)

	// Clone components with new IDs
	for oldID, comp := range ca.Components {
		clonedComp := comp.Clone()
		newID := clonedComp.GetID()
		clone.Components[newID] = clonedComp
		idMapping[oldID] = newID
	}

	// Clone connections with updated component IDs
	for _, conn := range ca.Connections {
		newConn := component.ComponentConnection{
			ID:            generateAssemblyID(), // Reuse assembly ID generator
			FromComponent: idMapping[conn.FromComponent],
			FromOutput:    conn.FromOutput,
			ToComponent:   idMapping[conn.ToComponent],
			ToInput:       conn.ToInput,
			Type:          conn.Type,
			Enabled:       conn.Enabled,
			Priority:      conn.Priority,
		}
		clone.Connections = append(clone.Connections, newConn)
	}

	// Update entry and exit points with new IDs
	for _, oldID := range ca.EntryPoints {
		if newID, exists := idMapping[oldID]; exists {
			clone.EntryPoints = append(clone.EntryPoints, newID)
		}
	}

	for _, oldID := range ca.ExitPoints {
		if newID, exists := idMapping[oldID]; exists {
			clone.ExitPoints = append(clone.ExitPoints, newID)
		}
	}

	return clone
}

// GetStats returns statistics about the assembly
func (ca *ComponentAssembly) GetStats() AssemblyStats {
	return AssemblyStats{
		ComponentCount:  len(ca.Components),
		ConnectionCount: len(ca.Connections),
		EntryPointCount: len(ca.EntryPoints),
		ExitPointCount:  len(ca.ExitPoints),
		ErrorCount:      len(ca.Errors),
		WarningCount:    len(ca.Warnings),
		IsValid:         ca.IsValid,
		Complexity:      ca.calculateComplexity(),
	}
}

// AssemblyStats provides statistics about an assembly
type AssemblyStats struct {
	ComponentCount  int     `json:"component_count"`
	ConnectionCount int     `json:"connection_count"`
	EntryPointCount int     `json:"entry_point_count"`
	ExitPointCount  int     `json:"exit_point_count"`
	ErrorCount      int     `json:"error_count"`
	WarningCount    int     `json:"warning_count"`
	IsValid         bool    `json:"is_valid"`
	Complexity      float64 `json:"complexity"` // 0.0 to 10.0
}

// Helper functions
func generateAssemblyID() string {
	// Use UUIDv7 for time-ordered unique IDs
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to UUIDv4 if v7 fails
		return uuid.New().String()
	}
	return id.String()
}

func (ca *ComponentAssembly) calculateComplexity() float64 {
	// Simple complexity calculation based on components and connections
	componentWeight := float64(len(ca.Components)) * 0.5
	connectionWeight := float64(len(ca.Connections)) * 0.3

	// Add penalty for circular dependencies and complex paths
	complexity := componentWeight + connectionWeight

	// Normalize to 0-10 scale
	if complexity > 10 {
		complexity = 10
	}

	return complexity
}
