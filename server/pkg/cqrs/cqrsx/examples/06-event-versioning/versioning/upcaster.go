package versioning

import (
	"fmt"
	"strings"
	"time"

	"cqrs"
	"cqrs/cqrsx/examples/06-event-versioning/domain"
)

// UserEventUpcaster handles upcasting User events to higher versions
type UserEventUpcaster struct {
	supportedPaths map[string]bool // Supported conversion paths
}

// NewUserEventUpcaster creates a new User event upcaster
func NewUserEventUpcaster() *UserEventUpcaster {
	supportedPaths := map[string]bool{
		"1->2": true, // V1 to V2
		"2->3": true, // V2 to V3
		"1->3": true, // V1 to V3 (via V2)
	}

	return &UserEventUpcaster{
		supportedPaths: supportedPaths,
	}
}

// UpcastToVersion upcasts an event to the target version
func (u *UserEventUpcaster) UpcastToVersion(event cqrs.EventMessage, targetVersion int) (cqrs.EventMessage, error) {
	// Detect current version from event
	currentVersion := u.detectEventVersion(event)

	if currentVersion == targetVersion {
		return event, nil
	}

	if currentVersion > targetVersion {
		return nil, fmt.Errorf("cannot upcast from higher version %d to lower version %d", currentVersion, targetVersion)
	}

	// Perform step-by-step upcasting
	currentEvent := event
	for currentVersion < targetVersion {
		nextVersion := currentVersion + 1

		var err error
		currentEvent, err = u.upcastToNextVersion(currentEvent, currentVersion, nextVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to upcast from v%d to v%d: %w", currentVersion, nextVersion, err)
		}

		currentVersion = nextVersion
	}

	return currentEvent, nil
}

// CanUpcast checks if upcasting is possible between versions
func (u *UserEventUpcaster) CanUpcast(fromVersion, toVersion int) bool {
	if fromVersion >= toVersion {
		return false
	}

	// Check if direct path exists
	path := fmt.Sprintf("%d->%d", fromVersion, toVersion)
	if u.supportedPaths[path] {
		return true
	}

	// Check if step-by-step path exists
	for version := fromVersion; version < toVersion; version++ {
		stepPath := fmt.Sprintf("%d->%d", version, version+1)
		if !u.supportedPaths[stepPath] {
			return false
		}
	}

	return true
}

// detectEventVersion detects the version of an event
func (u *UserEventUpcaster) detectEventVersion(event cqrs.EventMessage) int {
	// Check metadata for version info
	if metadata := event.Metadata(); metadata != nil {
		if versionStr, exists := metadata["version"]; exists {
			if version, ok := u.parseVersionString(versionStr); ok {
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
func (u *UserEventUpcaster) parseVersionString(versionValue interface{}) (int, bool) {
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

// upcastToNextVersion upcasts an event to the next version
func (u *UserEventUpcaster) upcastToNextVersion(event cqrs.EventMessage, fromVersion, toVersion int) (cqrs.EventMessage, error) {
	switch {
	case fromVersion == 1 && toVersion == 2:
		return u.upcastV1ToV2(event)
	case fromVersion == 2 && toVersion == 3:
		return u.upcastV2ToV3(event)
	default:
		return nil, fmt.Errorf("unsupported upcast path: v%d -> v%d", fromVersion, toVersion)
	}
}

// upcastV1ToV2 converts V1 events to V2
func (u *UserEventUpcaster) upcastV1ToV2(event cqrs.EventMessage) (cqrs.EventMessage, error) {
	switch v1Event := event.(type) {
	case *domain.UserCreatedV1:
		return u.upcastUserCreatedV1ToV2(v1Event), nil
	case *domain.UserUpdatedV1:
		return u.upcastUserUpdatedV1ToV2(v1Event), nil
	case *domain.UserDeletedV1:
		return u.upcastUserDeletedV1ToV2(v1Event), nil
	default:
		return nil, fmt.Errorf("unsupported V1 event type: %T", event)
	}
}

// upcastUserCreatedV1ToV2 converts UserCreatedV1 to UserCreatedV2
func (u *UserEventUpcaster) upcastUserCreatedV1ToV2(v1Event *domain.UserCreatedV1) *domain.UserCreatedV2 {
	// Extract name parts (simple heuristic)
	firstName, lastName := u.splitName(v1Event.Name)

	profile := domain.UserProfile{
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: time.Time{}, // Unknown, set to zero value
		Gender:      "not_specified",
		Avatar:      "",
		Bio:         "",
	}

	preferences := domain.DefaultUserPreferences()

	return domain.NewUserCreatedV2(
		v1Event.UserID,
		v1Event.Name,
		v1Event.Email,
		profile,
		preferences,
	)
}

// upcastUserUpdatedV1ToV2 converts UserUpdatedV1 to UserUpdatedV2
func (u *UserEventUpcaster) upcastUserUpdatedV1ToV2(v1Event *domain.UserUpdatedV1) *domain.UserUpdatedV2 {
	// Extract name parts
	firstName, lastName := u.splitName(v1Event.Name)

	profile := domain.UserProfile{
		FirstName:   firstName,
		LastName:    lastName,
		DateOfBirth: time.Time{},
		Gender:      "not_specified",
		Avatar:      "",
		Bio:         "",
	}

	preferences := domain.DefaultUserPreferences()

	return domain.NewUserUpdatedV2(
		v1Event.UserID,
		v1Event.Name,
		v1Event.Email,
		"system", // Default updater for migrated events
		profile,
		preferences,
	)
}

// upcastUserDeletedV1ToV2 converts UserDeletedV1 to UserDeletedV2
func (u *UserEventUpcaster) upcastUserDeletedV1ToV2(v1Event *domain.UserDeletedV1) *domain.UserDeletedV2 {
	metadata := map[string]interface{}{
		"migrated_from":  "v1",
		"migration_time": time.Now(),
	}

	return domain.NewUserDeletedV2(
		v1Event.UserID,
		v1Event.Reason,
		"system", // Default deleter for migrated events
		metadata,
	)
}

// upcastV2ToV3 converts V2 events to V3
func (u *UserEventUpcaster) upcastV2ToV3(event cqrs.EventMessage) (cqrs.EventMessage, error) {
	switch v2Event := event.(type) {
	case *domain.UserCreatedV2:
		return u.upcastUserCreatedV2ToV3(v2Event), nil
	case *domain.UserUpdatedV2:
		return u.upcastUserUpdatedV2ToV3(v2Event), nil
	case *domain.UserDeletedV2:
		return u.upcastUserDeletedV2ToV3(v2Event), nil
	default:
		return nil, fmt.Errorf("unsupported V2 event type: %T", event)
	}
}

// upcastUserCreatedV2ToV3 converts UserCreatedV2 to UserCreatedV3
func (u *UserEventUpcaster) upcastUserCreatedV2ToV3(v2Event *domain.UserCreatedV2) *domain.UserCreatedV3 {
	// Convert V2 profile to V3 personal info
	personalInfo := domain.PersonalInfo{
		FullName: domain.FullName{
			FirstName:  v2Event.Profile.FirstName,
			MiddleName: "", // Not available in V2
			LastName:   v2Event.Profile.LastName,
			Prefix:     "", // Not available in V2
			Suffix:     "", // Not available in V2
		},
		DateOfBirth: v2Event.Profile.DateOfBirth,
		Gender:      v2Event.Profile.Gender,
		Nationality: "", // Not available in V2
	}

	// Convert V2 email to V3 contact info
	contactInfo := domain.ContactInfo{
		PrimaryEmail:   v2Event.Email,
		SecondaryEmail: "",                       // Not available in V2
		PhoneNumbers:   []domain.PhoneNumber{},   // Not available in V2
		Addresses:      []domain.Address{},       // Not available in V2
		SocialProfiles: []domain.SocialProfile{}, // Not available in V2
	}

	// Create V3 metadata
	metadata := domain.EventMetadata{
		Version:       "3.0",
		SchemaVersion: "user_created_v3",
		Source:        "migration",
		CorrelationID: "",
		CausationID:   "",
		UserAgent:     "",
		IPAddress:     "",
		Timestamp:     time.Now(),
		Custom: map[string]interface{}{
			"migrated_from":       "v2",
			"original_created_at": v2Event.CreatedAt,
		},
	}

	return domain.NewUserCreatedV3(
		v2Event.UserID,
		personalInfo,
		contactInfo,
		v2Event.Preferences,
		metadata,
	)
}

// upcastUserUpdatedV2ToV3 converts UserUpdatedV2 to UserUpdatedV3
func (u *UserEventUpcaster) upcastUserUpdatedV2ToV3(v2Event *domain.UserUpdatedV2) *domain.UserUpdatedV3 {
	// Convert V2 profile to V3 personal info
	personalInfo := domain.PersonalInfo{
		FullName: domain.FullName{
			FirstName:  v2Event.Profile.FirstName,
			MiddleName: "",
			LastName:   v2Event.Profile.LastName,
			Prefix:     "",
			Suffix:     "",
		},
		DateOfBirth: v2Event.Profile.DateOfBirth,
		Gender:      v2Event.Profile.Gender,
		Nationality: "",
	}

	// Convert V2 email to V3 contact info
	contactInfo := domain.ContactInfo{
		PrimaryEmail:   v2Event.Email,
		SecondaryEmail: "",
		PhoneNumbers:   []domain.PhoneNumber{},
		Addresses:      []domain.Address{},
		SocialProfiles: []domain.SocialProfile{},
	}

	// Create field changes (simplified for migration)
	changes := []domain.FieldChange{
		{
			FieldPath: "migrated_from_v2",
			OldValue:  "v2_structure",
			NewValue:  "v3_structure",
			Timestamp: time.Now(),
		},
	}

	// Create V3 metadata
	metadata := domain.EventMetadata{
		Version:       "3.0",
		SchemaVersion: "user_updated_v3",
		Source:        "migration",
		CorrelationID: "",
		CausationID:   "",
		UserAgent:     "",
		IPAddress:     "",
		Timestamp:     time.Now(),
		Custom: map[string]interface{}{
			"migrated_from":       "v2",
			"original_updated_at": v2Event.UpdatedAt,
			"original_updated_by": v2Event.UpdatedBy,
		},
	}

	return domain.NewUserUpdatedV3(
		v2Event.UserID,
		personalInfo,
		contactInfo,
		v2Event.Preferences,
		changes,
		metadata,
	)
}

// upcastUserDeletedV2ToV3 converts UserDeletedV2 to UserDeletedV3
func (u *UserEventUpcaster) upcastUserDeletedV2ToV3(v2Event *domain.UserDeletedV2) *domain.UserDeletedV3 {
	// Convert V2 deletion info to V3 structured deletion reason
	deletionReason := domain.DeletionReason{
		Category:    "migrated", // Default category for migrated events
		Reason:      v2Event.Reason,
		Details:     "Migrated from V2 event",
		RequestedBy: v2Event.DeletedBy,
		ApprovedBy:  v2Event.DeletedBy, // Assume same person
		Timestamp:   v2Event.DeletedAt,
		Context:     v2Event.EventMetadata,
	}

	// Create V3 metadata
	metadata := domain.EventMetadata{
		Version:       "3.0",
		SchemaVersion: "user_deleted_v3",
		Source:        "migration",
		CorrelationID: "",
		CausationID:   "",
		UserAgent:     "",
		IPAddress:     "",
		Timestamp:     time.Now(),
		Custom: map[string]interface{}{
			"migrated_from":       "v2",
			"original_deleted_at": v2Event.DeletedAt,
		},
	}

	return domain.NewUserDeletedV3(
		v2Event.UserID,
		deletionReason,
		metadata,
	)
}

// splitName splits a full name into first and last name (simple heuristic)
func (u *UserEventUpcaster) splitName(fullName string) (firstName, lastName string) {
	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}
