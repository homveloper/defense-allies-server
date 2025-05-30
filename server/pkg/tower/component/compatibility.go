// Package component compatibility provides advanced type compatibility management
// including JSON-based rule configuration and dynamic rule loading.
package component

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// CompatibilityConfig represents the JSON configuration for type compatibility rules
type CompatibilityConfig struct {
	Version     string                  `json:"version"`
	Description string                  `json:"description"`
	Rules       []TypeCompatibilityRule `json:"rules"`
	Categories  []CompatibilityCategory `json:"categories"`
	Presets     []CompatibilityPreset   `json:"presets"`
}

// CompatibilityCategory groups related data types for easier rule management
type CompatibilityCategory struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Types       []DataType `json:"types"`

	// Category-wide rules
	InternalCompatible bool `json:"internal_compatible"` // All types in category are compatible with each other
	ConversionCost     int  `json:"conversion_cost"`     // Default cost for conversions within category
}

// CompatibilityPreset provides predefined sets of compatibility rules
type CompatibilityPreset struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Rules       []TypeCompatibilityRule `json:"rules"`
	Categories  []string                `json:"categories"` // Category names to include
}

// AdvancedCompatibilityMatrix extends the basic matrix with advanced features
type AdvancedCompatibilityMatrix struct {
	*TypeCompatibilityMatrix
	categories map[string]CompatibilityCategory
	presets    map[string]CompatibilityPreset
	config     *CompatibilityConfig
}

// NewAdvancedCompatibilityMatrix creates a new advanced compatibility matrix
func NewAdvancedCompatibilityMatrix() *AdvancedCompatibilityMatrix {
	return &AdvancedCompatibilityMatrix{
		TypeCompatibilityMatrix: NewTypeCompatibilityMatrix(),
		categories:              make(map[string]CompatibilityCategory),
		presets:                 make(map[string]CompatibilityPreset),
	}
}

// LoadFromJSON loads compatibility rules from a JSON file
func (acm *AdvancedCompatibilityMatrix) LoadFromJSON(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read compatibility config file %s: %w", filename, err)
	}

	var config CompatibilityConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse compatibility config: %w", err)
	}

	return acm.LoadFromConfig(config)
}

// LoadFromConfig loads compatibility rules from a configuration object
func (acm *AdvancedCompatibilityMatrix) LoadFromConfig(config CompatibilityConfig) error {
	acm.config = &config

	// Load categories first
	for _, category := range config.Categories {
		acm.categories[category.Name] = category

		// Add internal compatibility rules if specified
		if category.InternalCompatible {
			acm.addCategoryInternalRules(category)
		}
	}

	// Load presets
	for _, preset := range config.Presets {
		acm.presets[preset.Name] = preset
	}

	// Load individual rules
	for _, rule := range config.Rules {
		acm.AddRule(rule)
	}

	return nil
}

// LoadPreset applies a specific preset to the matrix
func (acm *AdvancedCompatibilityMatrix) LoadPreset(presetName string) error {
	preset, exists := acm.presets[presetName]
	if !exists {
		return fmt.Errorf("preset %s not found", presetName)
	}

	// Apply preset rules
	for _, rule := range preset.Rules {
		acm.AddRule(rule)
	}

	// Apply category rules
	for _, categoryName := range preset.Categories {
		if category, exists := acm.categories[categoryName]; exists {
			if category.InternalCompatible {
				acm.addCategoryInternalRules(category)
			}
		}
	}

	return nil
}

// addCategoryInternalRules adds compatibility rules within a category
func (acm *AdvancedCompatibilityMatrix) addCategoryInternalRules(category CompatibilityCategory) {
	for i, fromType := range category.Types {
		for j, toType := range category.Types {
			if i != j { // Don't add self-compatibility (already handled by exact match)
				rule := TypeCompatibilityRule{
					FromType:      fromType,
					ToType:        toType,
					Bidirectional: false, // Add both directions explicitly
					Cost:          category.ConversionCost,
					Transform:     "value", // Default pass-through transform
				}
				acm.AddRule(rule)
			}
		}
	}
}

// GetCategoryTypes returns all types in a category
func (acm *AdvancedCompatibilityMatrix) GetCategoryTypes(categoryName string) ([]DataType, error) {
	category, exists := acm.categories[categoryName]
	if !exists {
		return nil, fmt.Errorf("category %s not found", categoryName)
	}

	return category.Types, nil
}

// GetTypeCategory returns the category that contains the given type
func (acm *AdvancedCompatibilityMatrix) GetTypeCategory(dataType DataType) (string, bool) {
	for name, category := range acm.categories {
		for _, categoryType := range category.Types {
			if categoryType == dataType {
				return name, true
			}
		}
	}
	return "", false
}

// FindCompatibilityPath finds a path of conversions between two types
func (acm *AdvancedCompatibilityMatrix) FindCompatibilityPath(fromType, toType DataType) ([]TypeCompatibilityRule, int, bool) {
	// Direct compatibility
	if rule, exists := acm.GetCompatibilityRule(fromType, toType); exists {
		return []TypeCompatibilityRule{*rule}, rule.Cost, true
	}

	// Try to find a path through intermediate types
	visited := make(map[DataType]bool)
	return acm.findPathRecursive(fromType, toType, visited, []TypeCompatibilityRule{}, 0, 3) // Max depth of 3
}

// findPathRecursive recursively searches for a compatibility path
func (acm *AdvancedCompatibilityMatrix) findPathRecursive(
	current, target DataType,
	visited map[DataType]bool,
	path []TypeCompatibilityRule,
	totalCost int,
	maxDepth int,
) ([]TypeCompatibilityRule, int, bool) {
	if maxDepth <= 0 {
		return nil, 0, false
	}

	if visited[current] {
		return nil, 0, false
	}

	visited[current] = true
	defer func() { visited[current] = false }()

	// Check all possible next steps
	if rules, exists := acm.rules[current]; exists {
		for nextType, rule := range rules {
			newPath := append(path, *rule)
			newCost := totalCost + rule.Cost

			if nextType == target {
				return newPath, newCost, true
			}

			// Recursive search
			if finalPath, finalCost, found := acm.findPathRecursive(nextType, target, visited, newPath, newCost, maxDepth-1); found {
				return finalPath, finalCost, true
			}
		}
	}

	return nil, 0, false
}

// ExportToJSON exports the current compatibility matrix to JSON
func (acm *AdvancedCompatibilityMatrix) ExportToJSON() ([]byte, error) {
	config := CompatibilityConfig{
		Version:     "1.0",
		Description: "Exported compatibility configuration",
		Rules:       acm.getAllRules(),
		Categories:  acm.getAllCategories(),
		Presets:     acm.getAllPresets(),
	}

	return json.MarshalIndent(config, "", "  ")
}

// getAllRules returns all rules in the matrix
func (acm *AdvancedCompatibilityMatrix) getAllRules() []TypeCompatibilityRule {
	var rules []TypeCompatibilityRule

	for _, fromRules := range acm.rules {
		for _, rule := range fromRules {
			rules = append(rules, *rule)
		}
	}

	return rules
}

// getAllCategories returns all categories
func (acm *AdvancedCompatibilityMatrix) getAllCategories() []CompatibilityCategory {
	var categories []CompatibilityCategory

	for _, category := range acm.categories {
		categories = append(categories, category)
	}

	return categories
}

// getAllPresets returns all presets
func (acm *AdvancedCompatibilityMatrix) getAllPresets() []CompatibilityPreset {
	var presets []CompatibilityPreset

	for _, preset := range acm.presets {
		presets = append(presets, preset)
	}

	return presets
}

// ValidateConfiguration validates the compatibility configuration
func (acm *AdvancedCompatibilityMatrix) ValidateConfiguration() []error {
	var errors []error

	// Check for circular dependencies
	for fromType := range acm.rules {
		if acm.hasCircularDependency(fromType, make(map[DataType]bool)) {
			errors = append(errors, fmt.Errorf("circular dependency detected starting from type %s", fromType))
		}
	}

	// Check for orphaned types in categories
	for categoryName, category := range acm.categories {
		for _, dataType := range category.Types {
			if !acm.hasAnyRule(dataType) {
				errors = append(errors, fmt.Errorf("type %s in category %s has no compatibility rules", dataType, categoryName))
			}
		}
	}

	return errors
}

// hasCircularDependency checks for circular dependencies in compatibility rules
func (acm *AdvancedCompatibilityMatrix) hasCircularDependency(startType DataType, visited map[DataType]bool) bool {
	if visited[startType] {
		return true
	}

	visited[startType] = true
	defer func() { visited[startType] = false }()

	if rules, exists := acm.rules[startType]; exists {
		for toType := range rules {
			if acm.hasCircularDependency(toType, visited) {
				return true
			}
		}
	}

	return false
}

// hasAnyRule checks if a type has any compatibility rules
func (acm *AdvancedCompatibilityMatrix) hasAnyRule(dataType DataType) bool {
	// Check if type appears as source
	if _, exists := acm.rules[dataType]; exists {
		return true
	}

	// Check if type appears as target
	for _, rules := range acm.rules {
		if _, exists := rules[dataType]; exists {
			return true
		}
	}

	return false
}

// LoadCompatibilityFromDirectory loads all JSON files from a directory
func LoadCompatibilityFromDirectory(directory string) (*AdvancedCompatibilityMatrix, error) {
	matrix := NewAdvancedCompatibilityMatrix()

	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list JSON files in directory %s: %w", directory, err)
	}

	for _, file := range files {
		if err := matrix.LoadFromJSON(file); err != nil {
			return nil, fmt.Errorf("failed to load compatibility file %s: %w", file, err)
		}
	}

	return matrix, nil
}

// CreateDefaultCompatibilityConfig creates a default compatibility configuration
func CreateDefaultCompatibilityConfig() CompatibilityConfig {
	return CompatibilityConfig{
		Version:     "1.0",
		Description: "Default Defense Allies type compatibility configuration",
		Categories: []CompatibilityCategory{
			{
				Name:               "numeric",
				Description:        "Numeric data types",
				Types:              []DataType{DataTypeInt, DataTypeFloat},
				InternalCompatible: true,
				ConversionCost:     1,
			},
			{
				Name:               "entities",
				Description:        "Game entity types",
				Types:              []DataType{DataTypeTarget, DataTypeEnemy, DataTypeTower, DataTypePlayer},
				InternalCompatible: false,
				ConversionCost:     0,
			},
			{
				Name:               "collections",
				Description:        "Collection types",
				Types:              []DataType{DataTypeTargets, DataTypeEnemies, DataTypeTowers, DataTypePlayers},
				InternalCompatible: true,
				ConversionCost:     1,
			},
		},
		Presets: []CompatibilityPreset{
			{
				Name:        "basic",
				Description: "Basic compatibility rules for simple tower systems",
				Categories:  []string{"numeric", "entities"},
			},
			{
				Name:        "advanced",
				Description: "Advanced compatibility rules for complex tower systems",
				Categories:  []string{"numeric", "entities", "collections"},
			},
		},
	}
}

// Global advanced compatibility matrix
var globalAdvancedMatrix = NewAdvancedCompatibilityMatrix()

// InitializeCompatibilityFromConfig initializes the global matrix from config
func InitializeCompatibilityFromConfig(config CompatibilityConfig) error {
	return globalAdvancedMatrix.LoadFromConfig(config)
}

// InitializeCompatibilityFromFile initializes the global matrix from a JSON file
func InitializeCompatibilityFromFile(filename string) error {
	return globalAdvancedMatrix.LoadFromJSON(filename)
}
