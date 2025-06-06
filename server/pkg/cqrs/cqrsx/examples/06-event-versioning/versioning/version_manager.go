package versioning

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"defense-allies-server/pkg/cqrs"
)

// EventVersionManager manages event versioning and conversion
type EventVersionManager interface {
	// DetectVersion detects the version of an event from its data
	DetectVersion(eventData []byte, metadata map[string]interface{}) (int, error)

	// ConvertToVersion converts an event to a specific version
	ConvertToVersion(event cqrs.EventMessage, targetVersion int) (cqrs.EventMessage, error)

	// GetLatestVersion returns the latest supported version
	GetLatestVersion() int

	// GetSupportedVersions returns all supported versions
	GetSupportedVersions() []int

	// IsVersionSupported checks if a version is supported
	IsVersionSupported(version int) bool
}

// UserEventVersionManager implements version management for User events
type UserEventVersionManager struct {
	upcaster   EventUpcaster
	downcaster EventDowncaster
	factories  map[int]EventFactory
}

// EventFactory creates events from raw data for a specific version
type EventFactory interface {
	CreateEvent(eventType string, aggregateID string, eventData []byte, metadata map[string]interface{}) (cqrs.EventMessage, error)
	GetVersion() int
	GetSupportedEvents() []string
}

// EventUpcaster converts events to higher versions
type EventUpcaster interface {
	UpcastToVersion(event cqrs.EventMessage, targetVersion int) (cqrs.EventMessage, error)
	CanUpcast(fromVersion, toVersion int) bool
}

// EventDowncaster converts events to lower versions
type EventDowncaster interface {
	DowncastToVersion(event cqrs.EventMessage, targetVersion int) (cqrs.EventMessage, error)
	CanDowncast(fromVersion, toVersion int) bool
}

// NewUserEventVersionManager creates a new version manager for User events
func NewUserEventVersionManager(upcaster EventUpcaster, downcaster EventDowncaster, factories map[int]EventFactory) *UserEventVersionManager {
	return &UserEventVersionManager{
		upcaster:   upcaster,
		downcaster: downcaster,
		factories:  factories,
	}
}

// DetectVersion detects the version of an event from its data
func (vm *UserEventVersionManager) DetectVersion(eventData []byte, metadata map[string]interface{}) (int, error) {
	// 1. Check metadata for version information
	if version, exists := vm.getVersionFromMetadata(metadata); exists {
		return version, nil
	}

	// 2. Analyze event data structure to infer version
	return vm.inferVersionFromStructure(eventData)
}

// getVersionFromMetadata extracts version from event metadata
func (vm *UserEventVersionManager) getVersionFromMetadata(metadata map[string]interface{}) (int, bool) {
	if metadata == nil {
		return 0, false
	}

	// Check for explicit version field
	if versionStr, exists := metadata["version"]; exists {
		if version, err := vm.parseVersion(versionStr); err == nil {
			return version, true
		}
	}

	// Check for schema field that might contain version info
	if schema, exists := metadata["schema"]; exists {
		if schemaStr, ok := schema.(string); ok {
			if version := vm.extractVersionFromSchema(schemaStr); version > 0 {
				return version, true
			}
		}
	}

	return 0, false
}

// parseVersion parses version from various formats
func (vm *UserEventVersionManager) parseVersion(versionValue interface{}) (int, error) {
	switch v := versionValue.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		// Handle formats like "1.0", "2.0", "v1", "version_2"
		v = strings.ToLower(v)
		v = strings.TrimPrefix(v, "v")
		v = strings.TrimPrefix(v, "version_")

		if dotIndex := strings.Index(v, "."); dotIndex != -1 {
			v = v[:dotIndex] // Take only the major version
		}

		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("unsupported version format: %T", versionValue)
	}
}

// extractVersionFromSchema extracts version from schema string
func (vm *UserEventVersionManager) extractVersionFromSchema(schema string) int {
	// Handle patterns like "user_created_v1", "user_updated_v2"
	parts := strings.Split(schema, "_")
	for _, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) > 1 {
			if version, err := strconv.Atoi(part[1:]); err == nil {
				return version
			}
		}
	}
	return 0
}

// inferVersionFromStructure analyzes event structure to infer version
func (vm *UserEventVersionManager) inferVersionFromStructure(eventData []byte) (int, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(eventData, &data); err != nil {
		return 0, fmt.Errorf("failed to parse event data: %w", err)
	}

	// V3 indicators: personal_info, contact_info structures
	if _, hasPersonalInfo := data["personal_info"]; hasPersonalInfo {
		if _, hasContactInfo := data["contact_info"]; hasContactInfo {
			return 3, nil
		}
	}

	// V2 indicators: profile, preferences, created_at/updated_at
	if _, hasProfile := data["profile"]; hasProfile {
		if _, hasPreferences := data["preferences"]; hasPreferences {
			return 2, nil
		}
	}

	// V1 indicators: only basic fields (user_id, name, email)
	if vm.hasOnlyBasicFields(data) {
		return 1, nil
	}

	// Default to V1 if structure is unclear
	return 1, nil
}

// hasOnlyBasicFields checks if data contains only V1 basic fields
func (vm *UserEventVersionManager) hasOnlyBasicFields(data map[string]interface{}) bool {
	basicFields := map[string]bool{
		"user_id":    true,
		"name":       true,
		"email":      true,
		"deleted_at": true,
		"reason":     true,
	}

	for key := range data {
		if !basicFields[key] {
			return false
		}
	}

	return true
}

// ConvertToVersion converts an event to a specific version
func (vm *UserEventVersionManager) ConvertToVersion(event cqrs.EventMessage, targetVersion int) (cqrs.EventMessage, error) {
	if !vm.IsVersionSupported(targetVersion) {
		return nil, fmt.Errorf("unsupported target version: %d", targetVersion)
	}

	// Detect current version
	eventData, err := json.Marshal(event.EventData())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	currentVersion, err := vm.DetectVersion(eventData, event.Metadata())
	if err != nil {
		return nil, fmt.Errorf("failed to detect current version: %w", err)
	}

	// No conversion needed
	if currentVersion == targetVersion {
		return event, nil
	}

	// Upcast to higher version
	if currentVersion < targetVersion {
		if !vm.upcaster.CanUpcast(currentVersion, targetVersion) {
			return nil, fmt.Errorf("cannot upcast from version %d to %d", currentVersion, targetVersion)
		}
		return vm.upcaster.UpcastToVersion(event, targetVersion)
	}

	// Downcast to lower version
	if !vm.downcaster.CanDowncast(currentVersion, targetVersion) {
		return nil, fmt.Errorf("cannot downcast from version %d to %d", currentVersion, targetVersion)
	}
	return vm.downcaster.DowncastToVersion(event, targetVersion)
}

// GetLatestVersion returns the latest supported version
func (vm *UserEventVersionManager) GetLatestVersion() int {
	maxVersion := 0
	for version := range vm.factories {
		if version > maxVersion {
			maxVersion = version
		}
	}
	return maxVersion
}

// GetSupportedVersions returns all supported versions
func (vm *UserEventVersionManager) GetSupportedVersions() []int {
	versions := make([]int, 0, len(vm.factories))
	for version := range vm.factories {
		versions = append(versions, version)
	}
	return versions
}

// IsVersionSupported checks if a version is supported
func (vm *UserEventVersionManager) IsVersionSupported(version int) bool {
	_, exists := vm.factories[version]
	return exists
}

// CreateEventFromRawData creates an event from raw data using the appropriate factory
func (vm *UserEventVersionManager) CreateEventFromRawData(eventType string, aggregateID string, eventData []byte, metadata map[string]interface{}) (cqrs.EventMessage, error) {
	// Detect version
	version, err := vm.DetectVersion(eventData, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to detect version: %w", err)
	}

	// Get appropriate factory
	factory, exists := vm.factories[version]
	if !exists {
		return nil, fmt.Errorf("no factory available for version %d", version)
	}

	// Create event
	return factory.CreateEvent(eventType, aggregateID, eventData, metadata)
}

// VersionInfo represents version information
type VersionInfo struct {
	Version         int      `json:"version"`
	SupportedEvents []string `json:"supported_events"`
	IsLatest        bool     `json:"is_latest"`
	CanUpcastTo     []int    `json:"can_upcast_to"`
	CanDowncastTo   []int    `json:"can_downcast_to"`
}

// GetVersionInfo returns detailed information about a version
func (vm *UserEventVersionManager) GetVersionInfo(version int) (*VersionInfo, error) {
	if !vm.IsVersionSupported(version) {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	factory := vm.factories[version]
	latestVersion := vm.GetLatestVersion()

	// Determine upcast targets
	var canUpcastTo []int
	for targetVersion := range vm.factories {
		if targetVersion > version && vm.upcaster.CanUpcast(version, targetVersion) {
			canUpcastTo = append(canUpcastTo, targetVersion)
		}
	}

	// Determine downcast targets
	var canDowncastTo []int
	for targetVersion := range vm.factories {
		if targetVersion < version && vm.downcaster.CanDowncast(version, targetVersion) {
			canDowncastTo = append(canDowncastTo, targetVersion)
		}
	}

	return &VersionInfo{
		Version:         version,
		SupportedEvents: factory.GetSupportedEvents(),
		IsLatest:        version == latestVersion,
		CanUpcastTo:     canUpcastTo,
		CanDowncastTo:   canDowncastTo,
	}, nil
}
