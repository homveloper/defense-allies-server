// Package component registry provides registration and factory functionality
// for creating and managing component instances.
package component

import (
	"fmt"
	"reflect"
	"sync"
)

// ComponentFactory is a function that creates a new instance of a component
type ComponentFactory func(config map[string]interface{}) (AtomicComponent, error)

// ComponentRegistry manages the registration and creation of components
type ComponentRegistry struct {
	factories map[ComponentType]ComponentFactory
	metadata  map[ComponentType]ComponentMetadata
	mutex     sync.RWMutex
	
	// Component validation
	validators map[ComponentType]ComponentValidator
	
	// Component templates for quick creation
	templates map[string]ComponentTemplate
}

// ComponentValidator validates component configuration and state
type ComponentValidator interface {
	ValidateConfig(config map[string]interface{}) error
	ValidateState(component AtomicComponent) error
}

// ComponentTemplate provides a template for creating components with predefined configurations
type ComponentTemplate struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	ComponentType ComponentType         `json:"component_type"`
	DefaultConfig map[string]interface{} `json:"default_config"`
	RequiredInputs []string             `json:"required_inputs"`
	ProvidedOutputs []string            `json:"provided_outputs"`
	Category     ComponentCategory      `json:"category"`
	Tags         []string               `json:"tags"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author,omitempty"`
}

// NewComponentRegistry creates a new component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		factories:  make(map[ComponentType]ComponentFactory),
		metadata:   make(map[ComponentType]ComponentMetadata),
		validators: make(map[ComponentType]ComponentValidator),
		templates:  make(map[string]ComponentTemplate),
	}
}

// RegisterComponent registers a component factory with the registry
func (r *ComponentRegistry) RegisterComponent(
	componentType ComponentType,
	factory ComponentFactory,
	metadata ComponentMetadata,
	validator ComponentValidator,
) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Check if component type is already registered
	if _, exists := r.factories[componentType]; exists {
		return fmt.Errorf("component type %s is already registered", componentType)
	}
	
	// Validate the factory function
	if factory == nil {
		return fmt.Errorf("factory function cannot be nil for component type %s", componentType)
	}
	
	// Register the component
	r.factories[componentType] = factory
	r.metadata[componentType] = metadata
	if validator != nil {
		r.validators[componentType] = validator
	}
	
	return nil
}

// UnregisterComponent removes a component from the registry
func (r *ComponentRegistry) UnregisterComponent(componentType ComponentType) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.factories[componentType]; !exists {
		return fmt.Errorf("component type %s is not registered", componentType)
	}
	
	delete(r.factories, componentType)
	delete(r.metadata, componentType)
	delete(r.validators, componentType)
	
	return nil
}

// CreateComponent creates a new component instance
func (r *ComponentRegistry) CreateComponent(componentType ComponentType, config map[string]interface{}) (AtomicComponent, error) {
	r.mutex.RLock()
	factory, exists := r.factories[componentType]
	validator := r.validators[componentType]
	r.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("component type %s is not registered", componentType)
	}
	
	// Validate configuration if validator exists
	if validator != nil {
		if err := validator.ValidateConfig(config); err != nil {
			return nil, fmt.Errorf("invalid configuration for component type %s: %w", componentType, err)
		}
	}
	
	// Create the component
	component, err := factory(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create component of type %s: %w", componentType, err)
	}
	
	// Validate the created component
	if validator != nil {
		if err := validator.ValidateState(component); err != nil {
			return nil, fmt.Errorf("created component of type %s is invalid: %w", componentType, err)
		}
	}
	
	return component, nil
}

// GetComponentTypes returns all registered component types
func (r *ComponentRegistry) GetComponentTypes() []ComponentType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	types := make([]ComponentType, 0, len(r.factories))
	for componentType := range r.factories {
		types = append(types, componentType)
	}
	
	return types
}

// GetComponentMetadata returns metadata for a component type
func (r *ComponentRegistry) GetComponentMetadata(componentType ComponentType) (ComponentMetadata, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	metadata, exists := r.metadata[componentType]
	if !exists {
		return ComponentMetadata{}, fmt.Errorf("component type %s is not registered", componentType)
	}
	
	return metadata, nil
}

// GetComponentsByCategory returns all component types in a category
func (r *ComponentRegistry) GetComponentsByCategory(category ComponentCategory) []ComponentType {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var components []ComponentType
	for componentType, metadata := range r.metadata {
		if metadata.Category == category {
			components = append(components, componentType)
		}
	}
	
	return components
}

// IsRegistered checks if a component type is registered
func (r *ComponentRegistry) IsRegistered(componentType ComponentType) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	_, exists := r.factories[componentType]
	return exists
}

// RegisterTemplate registers a component template
func (r *ComponentRegistry) RegisterTemplate(name string, template ComponentTemplate) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.templates[name]; exists {
		return fmt.Errorf("template %s is already registered", name)
	}
	
	// Validate that the component type exists
	if _, exists := r.factories[template.ComponentType]; !exists {
		return fmt.Errorf("component type %s for template %s is not registered", template.ComponentType, name)
	}
	
	r.templates[name] = template
	return nil
}

// CreateFromTemplate creates a component from a template
func (r *ComponentRegistry) CreateFromTemplate(templateName string, overrides map[string]interface{}) (AtomicComponent, error) {
	r.mutex.RLock()
	template, exists := r.templates[templateName]
	r.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("template %s is not registered", templateName)
	}
	
	// Merge template config with overrides
	config := make(map[string]interface{})
	for k, v := range template.DefaultConfig {
		config[k] = v
	}
	for k, v := range overrides {
		config[k] = v
	}
	
	return r.CreateComponent(template.ComponentType, config)
}

// GetTemplates returns all registered templates
func (r *ComponentRegistry) GetTemplates() map[string]ComponentTemplate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	templates := make(map[string]ComponentTemplate)
	for name, template := range r.templates {
		templates[name] = template
	}
	
	return templates
}

// GetTemplatesByCategory returns templates in a specific category
func (r *ComponentRegistry) GetTemplatesByCategory(category ComponentCategory) []ComponentTemplate {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var templates []ComponentTemplate
	for _, template := range r.templates {
		if template.Category == category {
			templates = append(templates, template)
		}
	}
	
	return templates
}

// ComponentInfo provides detailed information about a registered component
type ComponentInfo struct {
	Type        ComponentType      `json:"type"`
	Metadata    ComponentMetadata  `json:"metadata"`
	HasValidator bool              `json:"has_validator"`
	Templates   []string          `json:"templates"`
	
	// Reflection information
	FactoryType string            `json:"factory_type"`
	
	// Usage statistics
	CreatedCount int64            `json:"created_count"`
	ErrorCount   int64            `json:"error_count"`
}

// GetComponentInfo returns detailed information about a component type
func (r *ComponentRegistry) GetComponentInfo(componentType ComponentType) (*ComponentInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	factory, exists := r.factories[componentType]
	if !exists {
		return nil, fmt.Errorf("component type %s is not registered", componentType)
	}
	
	metadata := r.metadata[componentType]
	_, hasValidator := r.validators[componentType]
	
	// Find templates that use this component type
	var templateNames []string
	for name, template := range r.templates {
		if template.ComponentType == componentType {
			templateNames = append(templateNames, name)
		}
	}
	
	info := &ComponentInfo{
		Type:         componentType,
		Metadata:     metadata,
		HasValidator: hasValidator,
		Templates:    templateNames,
		FactoryType:  reflect.TypeOf(factory).String(),
	}
	
	return info, nil
}

// GetAllComponentInfo returns information about all registered components
func (r *ComponentRegistry) GetAllComponentInfo() map[ComponentType]*ComponentInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	info := make(map[ComponentType]*ComponentInfo)
	for componentType := range r.factories {
		if componentInfo, err := r.GetComponentInfo(componentType); err == nil {
			info[componentType] = componentInfo
		}
	}
	
	return info
}

// ValidateRegistry validates the entire registry for consistency
func (r *ComponentRegistry) ValidateRegistry() []error {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var errors []error
	
	// Check for orphaned templates
	for name, template := range r.templates {
		if _, exists := r.factories[template.ComponentType]; !exists {
			errors = append(errors, fmt.Errorf("template %s references unregistered component type %s", name, template.ComponentType))
		}
	}
	
	// Check for missing metadata
	for componentType := range r.factories {
		if _, exists := r.metadata[componentType]; !exists {
			errors = append(errors, fmt.Errorf("component type %s has no metadata", componentType))
		}
	}
	
	return errors
}

// Global registry instance
var globalRegistry = NewComponentRegistry()

// Global functions for convenience

// RegisterComponent registers a component with the global registry
func RegisterComponent(componentType ComponentType, factory ComponentFactory, metadata ComponentMetadata, validator ComponentValidator) error {
	return globalRegistry.RegisterComponent(componentType, factory, metadata, validator)
}

// CreateComponent creates a component using the global registry
func CreateComponent(componentType ComponentType, config map[string]interface{}) (AtomicComponent, error) {
	return globalRegistry.CreateComponent(componentType, config)
}

// GetGlobalRegistry returns the global component registry
func GetGlobalRegistry() *ComponentRegistry {
	return globalRegistry
}
