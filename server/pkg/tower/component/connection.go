// Package component connection provides the connection system for linking components together.
package component

import (
	"fmt"
	"sync"
	"time"
)

// ComponentConnection represents a connection between two component ports
type ComponentConnection struct {
	ID            string         `json:"id"`
	FromComponent string         `json:"from_component"`
	FromOutput    string         `json:"from_output"`
	ToComponent   string         `json:"to_component"`
	ToInput       string         `json:"to_input"`
	Type          ConnectionType `json:"type"`
	Enabled       bool           `json:"enabled"`

	// Connection metadata
	CreatedAt   time.Time `json:"created_at"`
	CreatedBy   string    `json:"created_by,omitempty"`
	Description string    `json:"description,omitempty"`

	// Connection properties
	Priority  Priority `json:"priority"`
	Condition string   `json:"condition,omitempty"` // Expression for conditional connections
	Transform string   `json:"transform,omitempty"` // Data transformation expression

	// Validation and constraints
	Constraints []ConnectionConstraint `json:"constraints,omitempty"`

	// Runtime data
	LastUsed   time.Time `json:"last_used,omitempty"`
	UseCount   int64     `json:"use_count"`
	ErrorCount int64     `json:"error_count"`
	LastError  string    `json:"last_error,omitempty"`
}

// ConnectionConstraint defines validation rules for connections
type ConnectionConstraint struct {
	Type     ConstraintType `json:"type"`
	Value    interface{}    `json:"value"`
	Message  string         `json:"message"`
	Severity Severity       `json:"severity"`
}

// ConstraintType represents different types of connection constraints
type ConstraintType string

const (
	ConstraintTypeDataType ConstraintType = "data_type" // Must match specific data type
	ConstraintTypeRange    ConstraintType = "range"     // Value must be within range
	ConstraintTypeEnum     ConstraintType = "enum"      // Value must be one of specified values
	ConstraintTypePattern  ConstraintType = "pattern"   // String must match regex pattern
	ConstraintTypeRequired ConstraintType = "required"  // Input must have a value
	ConstraintTypeUnique   ConstraintType = "unique"    // Value must be unique
	ConstraintTypeCustom   ConstraintType = "custom"    // Custom validation function
)

// Severity represents the severity level of constraint violations
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// ConnectionPort represents a connection point on a component
type ConnectionPort struct {
	ComponentID string   `json:"component_id"`
	PortName    string   `json:"port_name"`
	PortType    PortType `json:"port_type"` // input or output
	DataType    DataType `json:"data_type"`
	Position    Vector2  `json:"position"` // UI position for visual editor
}

// PortType represents whether a port is an input or output
type PortType string

const (
	PortTypeInput  PortType = "input"
	PortTypeOutput PortType = "output"
)

// ConnectionGraph represents the complete graph of component connections
type ConnectionGraph struct {
	Connections []ComponentConnection `json:"connections"`
	Components  []string              `json:"components"` // Component IDs in the graph

	// Graph metadata
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Execution metadata
	EntryPoints []string `json:"entry_points"` // Components with no inputs
	ExitPoints  []string `json:"exit_points"`  // Components with no outputs

	// Validation state
	IsValid  bool                `json:"is_valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error in the connection graph
type ValidationError struct {
	Type         string    `json:"type"`
	Message      string    `json:"message"`
	ComponentID  string    `json:"component_id,omitempty"`
	ConnectionID string    `json:"connection_id,omitempty"`
	Severity     Severity  `json:"severity"`
	Timestamp    time.Time `json:"timestamp"`
}

// ValidationWarning represents a validation warning in the connection graph
type ValidationWarning struct {
	Type         string    `json:"type"`
	Message      string    `json:"message"`
	ComponentID  string    `json:"component_id,omitempty"`
	ConnectionID string    `json:"connection_id,omitempty"`
	Suggestion   string    `json:"suggestion,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// ConnectionValidator provides validation for component connections
type ConnectionValidator interface {
	// ValidateConnection checks if a single connection is valid
	ValidateConnection(conn ComponentConnection, components map[string]AtomicComponent) error

	// ValidateGraph checks if the entire connection graph is valid
	ValidateGraph(graph ConnectionGraph, components map[string]AtomicComponent) []ValidationError

	// CanConnect checks if two ports can be connected
	CanConnect(from, to ConnectionPort, components map[string]AtomicComponent) bool

	// GetConnectionSuggestions provides suggestions for connecting components
	GetConnectionSuggestions(componentID string, components map[string]AtomicComponent) []ConnectionSuggestion
}

// ConnectionSuggestion represents a suggested connection between components
type ConnectionSuggestion struct {
	From        ConnectionPort `json:"from"`
	To          ConnectionPort `json:"to"`
	Confidence  float64        `json:"confidence"` // 0.0 to 1.0
	Reason      string         `json:"reason"`
	Type        ConnectionType `json:"type"`
	AutoConnect bool           `json:"auto_connect"` // Whether this can be auto-connected
}

// ConnectionManager manages component connections and their lifecycle
type ConnectionManager interface {
	// CreateConnection creates a new connection between components
	CreateConnection(from, to ConnectionPort, connType ConnectionType) (*ComponentConnection, error)

	// RemoveConnection removes an existing connection
	RemoveConnection(connectionID string) error

	// GetConnections returns all connections for a component
	GetConnections(componentID string) []ComponentConnection

	// GetInputConnections returns all input connections for a component
	GetInputConnections(componentID string) []ComponentConnection

	// GetOutputConnections returns all output connections for a component
	GetOutputConnections(componentID string) []ComponentConnection

	// UpdateConnection updates an existing connection
	UpdateConnection(connectionID string, updates map[string]interface{}) error

	// EnableConnection enables a connection
	EnableConnection(connectionID string) error

	// DisableConnection disables a connection
	DisableConnection(connectionID string) error

	// ValidateConnections validates all connections
	ValidateConnections() []ValidationError
}

// DataTransformer handles data transformation between connected components
type DataTransformer interface {
	// Transform converts data from one format to another
	Transform(data interface{}, fromType, toType DataType, expression string) (interface{}, error)

	// CanTransform checks if data can be transformed between types
	CanTransform(fromType, toType DataType) bool

	// GetTransformExpression generates a transform expression for common conversions
	GetTransformExpression(fromType, toType DataType) string
}

// ConnectionMetrics tracks metrics for connection performance and usage
type ConnectionMetrics struct {
	ConnectionID   string        `json:"connection_id"`
	UseCount       int64         `json:"use_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	LastUsed       time.Time     `json:"last_used"`
	DataThroughput int64         `json:"data_throughput"` // bytes transferred

	// Performance metrics
	MinLatency  time.Duration `json:"min_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	SuccessRate float64       `json:"success_rate"`

	// Error tracking
	LastError    string   `json:"last_error,omitempty"`
	ErrorHistory []string `json:"error_history,omitempty"`
}

// ConnectionEvent represents events that occur during connection lifecycle
type ConnectionEvent struct {
	Type         ConnectionEventType `json:"type"`
	ConnectionID string              `json:"connection_id"`
	ComponentID  string              `json:"component_id,omitempty"`
	Data         interface{}         `json:"data,omitempty"`
	Timestamp    time.Time           `json:"timestamp"`
	Error        string              `json:"error,omitempty"`
}

// ConnectionEventType represents different types of connection events
type ConnectionEventType string

const (
	ConnectionEventCreated   ConnectionEventType = "created"
	ConnectionEventRemoved   ConnectionEventType = "removed"
	ConnectionEventEnabled   ConnectionEventType = "enabled"
	ConnectionEventDisabled  ConnectionEventType = "disabled"
	ConnectionEventDataFlow  ConnectionEventType = "data_flow"
	ConnectionEventError     ConnectionEventType = "error"
	ConnectionEventValidated ConnectionEventType = "validated"
)

// Utility functions for working with connections

// NewComponentConnection creates a new component connection
func NewComponentConnection(from, to ConnectionPort, connType ConnectionType) *ComponentConnection {
	return &ComponentConnection{
		ID:            generateConnectionID(from, to),
		FromComponent: from.ComponentID,
		FromOutput:    from.PortName,
		ToComponent:   to.ComponentID,
		ToInput:       to.PortName,
		Type:          connType,
		Enabled:       true,
		CreatedAt:     time.Now(),
		Priority:      PriorityNormal,
		UseCount:      0,
		ErrorCount:    0,
	}
}

// generateConnectionID generates a unique ID for a connection
func generateConnectionID(from, to ConnectionPort) string {
	return fmt.Sprintf("%s.%s->%s.%s", from.ComponentID, from.PortName, to.ComponentID, to.PortName)
}

// TypeCompatibilityRule represents a rule for type compatibility
type TypeCompatibilityRule struct {
	FromType      DataType `json:"from_type"`
	ToType        DataType `json:"to_type"`
	Bidirectional bool     `json:"bidirectional"` // If true, rule applies both ways
	Cost          int      `json:"cost"`          // Cost of conversion (0 = free, higher = more expensive)
	Transform     string   `json:"transform"`     // Optional transformation expression
}

// TypeCompatibilityMatrix manages type compatibility rules
type TypeCompatibilityMatrix struct {
	rules map[DataType]map[DataType]*TypeCompatibilityRule
	mutex sync.RWMutex
}

// NewTypeCompatibilityMatrix creates a new type compatibility matrix
func NewTypeCompatibilityMatrix() *TypeCompatibilityMatrix {
	matrix := &TypeCompatibilityMatrix{
		rules: make(map[DataType]map[DataType]*TypeCompatibilityRule),
	}

	// Initialize with default rules
	matrix.initializeDefaultRules()

	return matrix
}

// AddRule adds a compatibility rule to the matrix
func (tcm *TypeCompatibilityMatrix) AddRule(rule TypeCompatibilityRule) {
	tcm.mutex.Lock()
	defer tcm.mutex.Unlock()

	// Ensure maps exist
	if tcm.rules[rule.FromType] == nil {
		tcm.rules[rule.FromType] = make(map[DataType]*TypeCompatibilityRule)
	}

	tcm.rules[rule.FromType][rule.ToType] = &rule

	// Add reverse rule if bidirectional
	if rule.Bidirectional {
		if tcm.rules[rule.ToType] == nil {
			tcm.rules[rule.ToType] = make(map[DataType]*TypeCompatibilityRule)
		}
		reverseRule := rule
		reverseRule.FromType = rule.ToType
		reverseRule.ToType = rule.FromType
		tcm.rules[rule.ToType][rule.FromType] = &reverseRule
	}
}

// IsCompatible checks if two data types are compatible
func (tcm *TypeCompatibilityMatrix) IsCompatible(fromType, toType DataType) bool {
	tcm.mutex.RLock()
	defer tcm.mutex.RUnlock()

	// Exact match
	if fromType == toType {
		return true
	}

	// Any type is compatible with everything
	if fromType == DataTypeAny || toType == DataTypeAny {
		return true
	}

	// Check direct rule
	if fromRules, exists := tcm.rules[fromType]; exists {
		if _, ruleExists := fromRules[toType]; ruleExists {
			return true
		}
	}

	return false
}

// GetCompatibilityRule returns the compatibility rule between two types
func (tcm *TypeCompatibilityMatrix) GetCompatibilityRule(fromType, toType DataType) (*TypeCompatibilityRule, bool) {
	tcm.mutex.RLock()
	defer tcm.mutex.RUnlock()

	if fromRules, exists := tcm.rules[fromType]; exists {
		if rule, ruleExists := fromRules[toType]; ruleExists {
			return rule, true
		}
	}

	return nil, false
}

// GetCompatibleTypes returns all types that are compatible with the given type
func (tcm *TypeCompatibilityMatrix) GetCompatibleTypes(dataType DataType) []DataType {
	tcm.mutex.RLock()
	defer tcm.mutex.RUnlock()

	var compatible []DataType

	// Add exact match
	compatible = append(compatible, dataType)

	// Add Any type (always compatible)
	if dataType != DataTypeAny {
		compatible = append(compatible, DataTypeAny)
	}

	// Add types from rules
	if rules, exists := tcm.rules[dataType]; exists {
		for toType := range rules {
			compatible = append(compatible, toType)
		}
	}

	// Add types that can convert TO this type
	for fromType, rules := range tcm.rules {
		if _, exists := rules[dataType]; exists {
			compatible = append(compatible, fromType)
		}
	}

	return compatible
}

// initializeDefaultRules sets up the default type compatibility rules
func (tcm *TypeCompatibilityMatrix) initializeDefaultRules() {
	// Numeric conversions
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeInt, ToType: DataTypeFloat,
		Bidirectional: true, Cost: 1, Transform: "float(value)",
	})

	// Target type conversions
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeTarget, ToType: DataTypeTargets,
		Cost: 0, Transform: "[value]",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeTargets, ToType: DataTypeTarget,
		Cost: 1, Transform: "value[0]",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeTarget, ToType: DataTypeEnemy,
		Cost: 0, Transform: "value",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEnemy, ToType: DataTypeTarget,
		Cost: 0, Transform: "value",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEnemies, ToType: DataTypeTargets,
		Bidirectional: true, Cost: 0, Transform: "value",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEnemy, ToType: DataTypeEnemies,
		Cost: 0, Transform: "[value]",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEnemies, ToType: DataTypeEnemy,
		Cost: 1, Transform: "value[0]",
	})

	// Tower type conversions
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeTower, ToType: DataTypeTowers,
		Cost: 0, Transform: "[value]",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeTowers, ToType: DataTypeTower,
		Cost: 1, Transform: "value[0]",
	})

	// Effect conversions
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEffect, ToType: DataTypeEffects,
		Cost: 0, Transform: "[value]",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEffects, ToType: DataTypeEffect,
		Cost: 1, Transform: "value[0]",
	})

	// Event conversions
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEvent, ToType: DataTypeEvents,
		Cost: 0, Transform: "[value]",
	})
	tcm.AddRule(TypeCompatibilityRule{
		FromType: DataTypeEvents, ToType: DataTypeEvent,
		Cost: 1, Transform: "value[0]",
	})

	// String conversions (most types can be converted to string)
	basicTypes := []DataType{DataTypeInt, DataTypeFloat, DataTypeBool, DataTypeVector2}
	for _, dataType := range basicTypes {
		tcm.AddRule(TypeCompatibilityRule{
			FromType: dataType, ToType: DataTypeString,
			Cost: 2, Transform: "string(value)",
		})
	}
}

// Global compatibility matrix instance
var globalCompatibilityMatrix = NewTypeCompatibilityMatrix()

// IsCompatible checks if two data types are compatible using the global matrix
func IsCompatible(fromType, toType DataType) bool {
	return globalCompatibilityMatrix.IsCompatible(fromType, toType)
}

// GetCompatibilityRule returns the compatibility rule between two types using the global matrix
func GetCompatibilityRule(fromType, toType DataType) (*TypeCompatibilityRule, bool) {
	return globalCompatibilityMatrix.GetCompatibilityRule(fromType, toType)
}

// AddCompatibilityRule adds a new compatibility rule to the global matrix
func AddCompatibilityRule(rule TypeCompatibilityRule) {
	globalCompatibilityMatrix.AddRule(rule)
}

// GetCompatibleTypes returns all types compatible with the given type using the global matrix
func GetCompatibleTypes(dataType DataType) []DataType {
	return globalCompatibilityMatrix.GetCompatibleTypes(dataType)
}
