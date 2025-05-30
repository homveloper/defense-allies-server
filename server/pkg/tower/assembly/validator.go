// Package assembly validator provides validation logic for component assemblies
package assembly

import (
	"context"
	"fmt"
	"time"

	"defense-allies-server/pkg/tower/component"
)

// AssemblyValidator validates component assemblies for correctness and completeness
type AssemblyValidator struct {
	rules []ValidationRule
}

// ValidationRule represents a validation rule that can be applied to assemblies
type ValidationRule interface {
	Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue
	GetName() string
	GetDescription() string
	GetSeverity() Severity
}

// ValidationIssue represents an issue found during validation
type ValidationIssue struct {
	Type        string    `json:"type"`
	Severity    Severity  `json:"severity"`
	Message     string    `json:"message"`
	ComponentID string    `json:"component_id,omitempty"`
	ConnectionID string   `json:"connection_id,omitempty"`
	Suggestion  string    `json:"suggestion,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewAssemblyValidator creates a new assembly validator with default rules
func NewAssemblyValidator() *AssemblyValidator {
	validator := &AssemblyValidator{
		rules: []ValidationRule{},
	}
	
	// Add default validation rules
	validator.AddRule(&ComponentExistenceRule{})
	validator.AddRule(&ConnectionValidityRule{})
	validator.AddRule(&CircularDependencyRule{})
	validator.AddRule(&RequiredInputRule{})
	validator.AddRule(&TypeCompatibilityRule{})
	validator.AddRule(&OrphanedComponentRule{})
	validator.AddRule(&EntryPointRule{})
	validator.AddRule(&ExitPointRule{})
	
	return validator
}

// AddRule adds a validation rule to the validator
func (av *AssemblyValidator) AddRule(rule ValidationRule) {
	av.rules = append(av.rules, rule)
}

// RemoveRule removes a validation rule by name
func (av *AssemblyValidator) RemoveRule(ruleName string) {
	for i, rule := range av.rules {
		if rule.GetName() == ruleName {
			av.rules = append(av.rules[:i], av.rules[i+1:]...)
			break
		}
	}
}

// ValidateAssembly validates an assembly using all registered rules
func (av *AssemblyValidator) ValidateAssembly(ctx context.Context, assembly *ComponentAssembly) error {
	var allIssues []ValidationIssue
	
	// Run all validation rules
	for _, rule := range av.rules {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			issues := rule.Validate(ctx, assembly)
			allIssues = append(allIssues, issues...)
		}
	}
	
	// Categorize issues into errors and warnings
	var errors []AssemblyError
	var warnings []AssemblyWarning
	
	for _, issue := range allIssues {
		if issue.Severity == SeverityError || issue.Severity == SeverityCritical {
			errors = append(errors, AssemblyError{
				Type:         ErrorType(issue.Type),
				Message:      issue.Message,
				ComponentID:  issue.ComponentID,
				ConnectionID: issue.ConnectionID,
				Severity:     issue.Severity,
				Timestamp:    issue.Timestamp,
			})
		} else {
			warnings = append(warnings, AssemblyWarning{
				Type:         WarningType(issue.Type),
				Message:      issue.Message,
				ComponentID:  issue.ComponentID,
				ConnectionID: issue.ConnectionID,
				Suggestion:   issue.Suggestion,
				Timestamp:    issue.Timestamp,
			})
		}
	}
	
	// Update assembly validation state
	assembly.Errors = errors
	assembly.Warnings = warnings
	assembly.IsValid = len(errors) == 0
	
	// Calculate execution order if valid
	if assembly.IsValid {
		executionOrder, err := av.calculateExecutionOrder(assembly)
		if err != nil {
			assembly.IsValid = false
			assembly.Errors = append(assembly.Errors, AssemblyError{
				Type:      ErrorTypeCircularDependency,
				Message:   fmt.Sprintf("Failed to calculate execution order: %v", err),
				Severity:  SeverityError,
				Timestamp: time.Now(),
			})
		} else {
			assembly.ExecutionOrder = executionOrder
			assembly.EntryPoints = av.findEntryPoints(assembly)
			assembly.ExitPoints = av.findExitPoints(assembly)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("assembly validation failed with %d errors", len(errors))
	}
	
	return nil
}

// calculateExecutionOrder calculates the topological order for component execution
func (av *AssemblyValidator) calculateExecutionOrder(assembly *ComponentAssembly) ([]string, error) {
	// Build dependency graph
	inDegree := make(map[string]int)
	dependencies := make(map[string][]string)
	
	// Initialize all components with in-degree 0
	for componentID := range assembly.Components {
		inDegree[componentID] = 0
		dependencies[componentID] = []string{}
	}
	
	// Calculate in-degrees based on connections
	for _, conn := range assembly.Connections {
		inDegree[conn.ToComponent]++
		dependencies[conn.FromComponent] = append(dependencies[conn.FromComponent], conn.ToComponent)
	}
	
	// Kahn's algorithm for topological sorting
	var queue []string
	var result []string
	
	// Find all nodes with in-degree 0
	for componentID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, componentID)
		}
	}
	
	// Process queue
	for len(queue) > 0 {
		// Remove a node from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		// For each dependent of current node
		for _, dependent := range dependencies[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}
	
	// Check for circular dependencies
	if len(result) != len(assembly.Components) {
		return nil, fmt.Errorf("circular dependency detected")
	}
	
	return result, nil
}

// findEntryPoints finds components with no input connections
func (av *AssemblyValidator) findEntryPoints(assembly *ComponentAssembly) []string {
	var entryPoints []string
	
	for componentID := range assembly.Components {
		hasInput := false
		for _, conn := range assembly.Connections {
			if conn.ToComponent == componentID {
				hasInput = true
				break
			}
		}
		if !hasInput {
			entryPoints = append(entryPoints, componentID)
		}
	}
	
	return entryPoints
}

// findExitPoints finds components with no output connections
func (av *AssemblyValidator) findExitPoints(assembly *ComponentAssembly) []string {
	var exitPoints []string
	
	for componentID := range assembly.Components {
		hasOutput := false
		for _, conn := range assembly.Connections {
			if conn.FromComponent == componentID {
				hasOutput = true
				break
			}
		}
		if !hasOutput {
			exitPoints = append(exitPoints, componentID)
		}
	}
	
	return exitPoints
}

// Validation Rules Implementation

// ComponentExistenceRule validates that all referenced components exist
type ComponentExistenceRule struct{}

func (cer *ComponentExistenceRule) GetName() string { return "component_existence" }
func (cer *ComponentExistenceRule) GetDescription() string { return "Validates that all referenced components exist" }
func (cer *ComponentExistenceRule) GetSeverity() Severity { return SeverityError }

func (cer *ComponentExistenceRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	for _, conn := range assembly.Connections {
		if _, exists := assembly.Components[conn.FromComponent]; !exists {
			issues = append(issues, ValidationIssue{
				Type:         "missing_component",
				Severity:     SeverityError,
				Message:      fmt.Sprintf("From component %s does not exist", conn.FromComponent),
				ConnectionID: conn.ID,
				Timestamp:    time.Now(),
			})
		}
		
		if _, exists := assembly.Components[conn.ToComponent]; !exists {
			issues = append(issues, ValidationIssue{
				Type:         "missing_component",
				Severity:     SeverityError,
				Message:      fmt.Sprintf("To component %s does not exist", conn.ToComponent),
				ConnectionID: conn.ID,
				Timestamp:    time.Now(),
			})
		}
	}
	
	return issues
}

// ConnectionValidityRule validates that connections are valid
type ConnectionValidityRule struct{}

func (cvr *ConnectionValidityRule) GetName() string { return "connection_validity" }
func (cvr *ConnectionValidityRule) GetDescription() string { return "Validates that connections are valid" }
func (cvr *ConnectionValidityRule) GetSeverity() Severity { return SeverityError }

func (cvr *ConnectionValidityRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	for _, conn := range assembly.Connections {
		fromComp, fromExists := assembly.Components[conn.FromComponent]
		toComp, toExists := assembly.Components[conn.ToComponent]
		
		if fromExists && toExists {
			if !fromComp.CanConnectTo(toComp) {
				issues = append(issues, ValidationIssue{
					Type:         "invalid_connection",
					Severity:     SeverityError,
					Message:      fmt.Sprintf("Component %s cannot connect to %s", conn.FromComponent, conn.ToComponent),
					ConnectionID: conn.ID,
					Suggestion:   "Check component compatibility and data types",
					Timestamp:    time.Now(),
				})
			}
		}
	}
	
	return issues
}

// CircularDependencyRule detects circular dependencies
type CircularDependencyRule struct{}

func (cdr *CircularDependencyRule) GetName() string { return "circular_dependency" }
func (cdr *CircularDependencyRule) GetDescription() string { return "Detects circular dependencies" }
func (cdr *CircularDependencyRule) GetSeverity() Severity { return SeverityError }

func (cdr *CircularDependencyRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	// Build adjacency list
	graph := make(map[string][]string)
	for componentID := range assembly.Components {
		graph[componentID] = []string{}
	}
	
	for _, conn := range assembly.Connections {
		graph[conn.FromComponent] = append(graph[conn.FromComponent], conn.ToComponent)
	}
	
	// Check for cycles starting from each component
	for componentID := range assembly.Components {
		if !visited[componentID] {
			if cdr.hasCycle(componentID, graph, visited, recStack) {
				issues = append(issues, ValidationIssue{
					Type:        "circular_dependency",
					Severity:    SeverityError,
					Message:     fmt.Sprintf("Circular dependency detected involving component %s", componentID),
					ComponentID: componentID,
					Suggestion:  "Remove or reorganize connections to eliminate the cycle",
					Timestamp:   time.Now(),
				})
				break // Only report one cycle to avoid spam
			}
		}
	}
	
	return issues
}

func (cdr *CircularDependencyRule) hasCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true
	
	for _, neighbor := range graph[node] {
		if !visited[neighbor] && cdr.hasCycle(neighbor, graph, visited, recStack) {
			return true
		} else if recStack[neighbor] {
			return true
		}
	}
	
	recStack[node] = false
	return false
}

// RequiredInputRule validates that all required inputs are connected
type RequiredInputRule struct{}

func (rir *RequiredInputRule) GetName() string { return "required_input" }
func (rir *RequiredInputRule) GetDescription() string { return "Validates that all required inputs are connected" }
func (rir *RequiredInputRule) GetSeverity() Severity { return SeverityError }

func (rir *RequiredInputRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	for componentID, comp := range assembly.Components {
		inputs := comp.GetInputs()
		
		for _, input := range inputs {
			if input.Required {
				connected := false
				for _, conn := range assembly.Connections {
					if conn.ToComponent == componentID && conn.ToInput == input.Name {
						connected = true
						break
					}
				}
				
				if !connected {
					issues = append(issues, ValidationIssue{
						Type:        "missing_required_input",
						Severity:    SeverityError,
						Message:     fmt.Sprintf("Required input '%s' of component %s is not connected", input.Name, componentID),
						ComponentID: componentID,
						Suggestion:  fmt.Sprintf("Connect a compatible output to the '%s' input", input.Name),
						Timestamp:   time.Now(),
					})
				}
			}
		}
	}
	
	return issues
}

// TypeCompatibilityRule validates type compatibility between connections
type TypeCompatibilityRule struct{}

func (tcr *TypeCompatibilityRule) GetName() string { return "type_compatibility" }
func (tcr *TypeCompatibilityRule) GetDescription() string { return "Validates type compatibility between connections" }
func (tcr *TypeCompatibilityRule) GetSeverity() Severity { return SeverityError }

func (tcr *TypeCompatibilityRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	for _, conn := range assembly.Connections {
		fromComp, fromExists := assembly.Components[conn.FromComponent]
		toComp, toExists := assembly.Components[conn.ToComponent]
		
		if fromExists && toExists {
			// Find the output and input types
			var fromType, toType component.DataType
			
			for _, output := range fromComp.GetOutputs() {
				if output.Name == conn.FromOutput {
					fromType = output.Type
					break
				}
			}
			
			for _, input := range toComp.GetInputs() {
				if input.Name == conn.ToInput {
					toType = input.Type
					break
				}
			}
			
			if fromType != "" && toType != "" && !component.IsCompatible(fromType, toType) {
				issues = append(issues, ValidationIssue{
					Type:         "incompatible_types",
					Severity:     SeverityError,
					Message:      fmt.Sprintf("Incompatible types: %s -> %s", fromType, toType),
					ConnectionID: conn.ID,
					Suggestion:   "Use a compatible data type or add a type converter component",
					Timestamp:    time.Now(),
				})
			}
		}
	}
	
	return issues
}

// OrphanedComponentRule detects components with no connections
type OrphanedComponentRule struct{}

func (ocr *OrphanedComponentRule) GetName() string { return "orphaned_component" }
func (ocr *OrphanedComponentRule) GetDescription() string { return "Detects components with no connections" }
func (ocr *OrphanedComponentRule) GetSeverity() Severity { return SeverityWarning }

func (ocr *OrphanedComponentRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	for componentID := range assembly.Components {
		hasConnection := false
		for _, conn := range assembly.Connections {
			if conn.FromComponent == componentID || conn.ToComponent == componentID {
				hasConnection = true
				break
			}
		}
		
		if !hasConnection {
			issues = append(issues, ValidationIssue{
				Type:        "orphaned_component",
				Severity:    SeverityWarning,
				Message:     fmt.Sprintf("Component %s has no connections", componentID),
				ComponentID: componentID,
				Suggestion:  "Connect this component to other components or remove it",
				Timestamp:   time.Now(),
			})
		}
	}
	
	return issues
}

// EntryPointRule validates that there is at least one entry point
type EntryPointRule struct{}

func (epr *EntryPointRule) GetName() string { return "entry_point" }
func (epr *EntryPointRule) GetDescription() string { return "Validates that there is at least one entry point" }
func (epr *EntryPointRule) GetSeverity() Severity { return SeverityWarning }

func (epr *EntryPointRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	entryPoints := 0
	for componentID := range assembly.Components {
		hasInput := false
		for _, conn := range assembly.Connections {
			if conn.ToComponent == componentID {
				hasInput = true
				break
			}
		}
		if !hasInput {
			entryPoints++
		}
	}
	
	if entryPoints == 0 && len(assembly.Components) > 0 {
		issues = append(issues, ValidationIssue{
			Type:       "no_entry_point",
			Severity:   SeverityWarning,
			Message:    "Assembly has no entry points (components with no inputs)",
			Suggestion: "Ensure at least one component can receive external input",
			Timestamp:  time.Now(),
		})
	}
	
	return issues
}

// ExitPointRule validates that there is at least one exit point
type ExitPointRule struct{}

func (expr *ExitPointRule) GetName() string { return "exit_point" }
func (expr *ExitPointRule) GetDescription() string { return "Validates that there is at least one exit point" }
func (expr *ExitPointRule) GetSeverity() Severity { return SeverityWarning }

func (expr *ExitPointRule) Validate(ctx context.Context, assembly *ComponentAssembly) []ValidationIssue {
	var issues []ValidationIssue
	
	exitPoints := 0
	for componentID := range assembly.Components {
		hasOutput := false
		for _, conn := range assembly.Connections {
			if conn.FromComponent == componentID {
				hasOutput = true
				break
			}
		}
		if !hasOutput {
			exitPoints++
		}
	}
	
	if exitPoints == 0 && len(assembly.Components) > 0 {
		issues = append(issues, ValidationIssue{
			Type:       "no_exit_point",
			Severity:   SeverityWarning,
			Message:    "Assembly has no exit points (components with no outputs)",
			Suggestion: "Ensure at least one component produces external output",
			Timestamp:  time.Now(),
		})
	}
	
	return issues
}
