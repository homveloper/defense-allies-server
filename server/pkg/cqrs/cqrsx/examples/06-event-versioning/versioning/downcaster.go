package versioning

import (
	"fmt"
	"strings"

	"cqrs"
	"cqrs/cqrsx/examples/06-event-versioning/domain"
)

// UserEventDowncaster handles downcasting User events to lower versions
type UserEventDowncaster struct {
	supportedPaths map[string]bool // Supported conversion paths
}

// NewUserEventDowncaster creates a new User event downcaster
func NewUserEventDowncaster() *UserEventDowncaster {
	supportedPaths := map[string]bool{
		"2->1": true, // V2 to V1
		"3->2": true, // V3 to V2
		"3->1": true, // V3 to V1 (via V2)
	}

	return &UserEventDowncaster{
		supportedPaths: supportedPaths,
	}
}

// DowncastToVersion downcasts an event to the target version
func (d *UserEventDowncaster) DowncastToVersion(event cqrs.EventMessage, targetVersion int) (cqrs.EventMessage, error) {
	// Detect current version from event
	currentVersion := d.detectEventVersion(event)

	if currentVersion == targetVersion {
		return event, nil
	}

	if currentVersion < targetVersion {
		return nil, fmt.Errorf("cannot downcast from lower version %d to higher version %d", currentVersion, targetVersion)
	}

	// Perform step-by-step downcasting
	currentEvent := event
	for currentVersion > targetVersion {
		nextVersion := currentVersion - 1

		var err error
		currentEvent, err = d.downcastToNextVersion(currentEvent, currentVersion, nextVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to downcast from v%d to v%d: %w", currentVersion, nextVersion, err)
		}

		currentVersion = nextVersion
	}

	return currentEvent, nil
}

// CanDowncast checks if downcasting is possible between versions
func (d *UserEventDowncaster) CanDowncast(fromVersion, toVersion int) bool {
	if fromVersion <= toVersion {
		return false
	}

	// Check if direct path exists
	path := fmt.Sprintf("%d->%d", fromVersion, toVersion)
	if d.supportedPaths[path] {
		return true
	}

	// Check if step-by-step path exists
	for version := fromVersion; version > toVersion; version-- {
		stepPath := fmt.Sprintf("%d->%d", version, version-1)
		if !d.supportedPaths[stepPath] {
			return false
		}
	}

	return true
}

// detectEventVersion detects the version of an event
func (d *UserEventDowncaster) detectEventVersion(event cqrs.EventMessage) int {
	// Check metadata for version info
	if metadata := event.Metadata(); metadata != nil {
		if versionStr, exists := metadata["version"]; exists {
			if version, ok := d.parseVersionString(versionStr); ok {
				return version
			}
		}
	}

	// Fallback: analyze event type
	switch event.(type) {
	case *domain.UserCreatedV1, *domain.UserUpdatedV1, *domain.UserDeletedV1:
		return 1
	case *domain.UserCreatedV2, *domain.UserUpdatedV2, *domain.UserDeletedV2:
		return 2
	case *domain.UserCreatedV3, *domain.UserUpdatedV3, *domain.UserDeletedV3:
		return 3
	default:
		return 1 // Default to V1
	}
}

// parseVersionString parses version from string
func (d *UserEventDowncaster) parseVersionString(versionValue interface{}) (int, bool) {
	switch v := versionValue.(type) {
	case string:
		switch v {
		case "1.0":
			return 1, true
		case "2.0":
			return 2, true
		case "3.0":
			return 3, true
		}
	case int:
		return v, true
	case float64:
		return int(v), true
	}
	return 0, false
}

// downcastToNextVersion downcasts an event to the next lower version
func (d *UserEventDowncaster) downcastToNextVersion(event cqrs.EventMessage, fromVersion, toVersion int) (cqrs.EventMessage, error) {
	switch {
	case fromVersion == 2 && toVersion == 1:
		return d.downcastV2ToV1(event)
	case fromVersion == 3 && toVersion == 2:
		return d.downcastV3ToV2(event)
	default:
		return nil, fmt.Errorf("unsupported downcast path: v%d -> v%d", fromVersion, toVersion)
	}
}

// downcastV2ToV1 converts V2 events to V1
func (d *UserEventDowncaster) downcastV2ToV1(event cqrs.EventMessage) (cqrs.EventMessage, error) {
	switch v2Event := event.(type) {
	case *domain.UserCreatedV2:
		return d.downcastUserCreatedV2ToV1(v2Event), nil
	case *domain.UserUpdatedV2:
		return d.downcastUserUpdatedV2ToV1(v2Event), nil
	case *domain.UserDeletedV2:
		return d.downcastUserDeletedV2ToV1(v2Event), nil
	default:
		return nil, fmt.Errorf("unsupported V2 event type: %T", event)
	}
}

// downcastUserCreatedV2ToV1 converts UserCreatedV2 to UserCreatedV1
func (d *UserEventDowncaster) downcastUserCreatedV2ToV1(v2Event *domain.UserCreatedV2) *domain.UserCreatedV1 {
	return domain.NewUserCreatedV1(
		v2Event.UserID,
		v2Event.Name,
		v2Event.Email,
		// Profile, Preferences, CreatedAt fields are dropped
	)
}

// downcastUserUpdatedV2ToV1 converts UserUpdatedV2 to UserUpdatedV1
func (d *UserEventDowncaster) downcastUserUpdatedV2ToV1(v2Event *domain.UserUpdatedV2) *domain.UserUpdatedV1 {
	return domain.NewUserUpdatedV1(
		v2Event.UserID,
		v2Event.Name,
		v2Event.Email,
		// Profile, Preferences, UpdatedAt, UpdatedBy fields are dropped
	)
}

// downcastUserDeletedV2ToV1 converts UserDeletedV2 to UserDeletedV1
func (d *UserEventDowncaster) downcastUserDeletedV2ToV1(v2Event *domain.UserDeletedV2) *domain.UserDeletedV1 {
	return domain.NewUserDeletedV1(
		v2Event.UserID,
		v2Event.Reason,
		// DeletedBy, EventMetadata fields are dropped
		// DeletedAt is preserved in the constructor
	)
}

// downcastV3ToV2 converts V3 events to V2
func (d *UserEventDowncaster) downcastV3ToV2(event cqrs.EventMessage) (cqrs.EventMessage, error) {
	switch v3Event := event.(type) {
	case *domain.UserCreatedV3:
		return d.downcastUserCreatedV3ToV2(v3Event), nil
	case *domain.UserUpdatedV3:
		return d.downcastUserUpdatedV3ToV2(v3Event), nil
	case *domain.UserDeletedV3:
		return d.downcastUserDeletedV3ToV2(v3Event), nil
	default:
		return nil, fmt.Errorf("unsupported V3 event type: %T", event)
	}
}

// downcastUserCreatedV3ToV2 converts UserCreatedV3 to UserCreatedV2
func (d *UserEventDowncaster) downcastUserCreatedV3ToV2(v3Event *domain.UserCreatedV3) *domain.UserCreatedV2 {
	// Convert V3 personal info to V2 profile
	profile := domain.UserProfile{
		FirstName:   v3Event.PersonalInfo.FullName.FirstName,
		LastName:    v3Event.PersonalInfo.FullName.LastName,
		DateOfBirth: v3Event.PersonalInfo.DateOfBirth,
		Gender:      v3Event.PersonalInfo.Gender,
		Avatar:      "", // Not available in V3 structure
		Bio:         "", // Not available in V3 structure
	}

	// Extract primary email from V3 contact info
	email := v3Event.ContactInfo.PrimaryEmail
	if email == "" && len(v3Event.ContactInfo.PhoneNumbers) > 0 {
		// Fallback: use secondary email if primary is empty
		email = v3Event.ContactInfo.SecondaryEmail
	}

	// Construct name from V3 full name
	name := d.constructFullName(v3Event.PersonalInfo.FullName)

	return domain.NewUserCreatedV2(
		v3Event.UserID,
		name,
		email,
		profile,
		v3Event.Preferences,
		// EventMetadata fields are dropped, CreatedAt is set to current time
	)
}

// downcastUserUpdatedV3ToV2 converts UserUpdatedV3 to UserUpdatedV2
func (d *UserEventDowncaster) downcastUserUpdatedV3ToV2(v3Event *domain.UserUpdatedV3) *domain.UserUpdatedV2 {
	// Convert V3 personal info to V2 profile
	profile := domain.UserProfile{
		FirstName:   v3Event.PersonalInfo.FullName.FirstName,
		LastName:    v3Event.PersonalInfo.FullName.LastName,
		DateOfBirth: v3Event.PersonalInfo.DateOfBirth,
		Gender:      v3Event.PersonalInfo.Gender,
		Avatar:      "",
		Bio:         "",
	}

	// Extract primary email from V3 contact info
	email := v3Event.ContactInfo.PrimaryEmail
	if email == "" {
		email = v3Event.ContactInfo.SecondaryEmail
	}

	// Construct name from V3 full name
	name := d.constructFullName(v3Event.PersonalInfo.FullName)

	// Extract updater from V3 metadata or changes
	updatedBy := "system" // Default
	if v3Event.EventMetadata.Source != "" {
		updatedBy = v3Event.EventMetadata.Source
	}

	return domain.NewUserUpdatedV2(
		v3Event.UserID,
		name,
		email,
		updatedBy,
		profile,
		v3Event.Preferences,
		// Changes and detailed metadata are dropped
	)
}

// downcastUserDeletedV3ToV2 converts UserDeletedV3 to UserDeletedV2
func (d *UserEventDowncaster) downcastUserDeletedV3ToV2(v3Event *domain.UserDeletedV3) *domain.UserDeletedV2 {
	// Extract simple metadata from V3 structured deletion reason
	metadata := map[string]interface{}{
		"category":      v3Event.DeletionReason.Category,
		"details":       v3Event.DeletionReason.Details,
		"approved_by":   v3Event.DeletionReason.ApprovedBy,
		"downcast_from": "v3",
	}

	// Merge V3 context into metadata
	for key, value := range v3Event.DeletionReason.Context {
		metadata[key] = value
	}

	return domain.NewUserDeletedV2(
		v3Event.UserID,
		v3Event.DeletionReason.Reason,
		v3Event.DeletionReason.RequestedBy,
		metadata,
	)
}

// constructFullName constructs a full name from V3 FullName structure
func (d *UserEventDowncaster) constructFullName(fullName domain.FullName) string {
	parts := []string{}

	if fullName.Prefix != "" {
		parts = append(parts, fullName.Prefix)
	}
	if fullName.FirstName != "" {
		parts = append(parts, fullName.FirstName)
	}
	if fullName.MiddleName != "" {
		parts = append(parts, fullName.MiddleName)
	}
	if fullName.LastName != "" {
		parts = append(parts, fullName.LastName)
	}
	if fullName.Suffix != "" {
		parts = append(parts, fullName.Suffix)
	}

	return strings.Join(parts, " ")
}
