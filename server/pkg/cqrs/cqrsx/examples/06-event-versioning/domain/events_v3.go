package domain

import (
	"encoding/json"
	"time"

	"cqrs"
)

// V3 이벤트들 - 구조적 변경 버전

// PersonalInfo represents restructured personal information (V3)
type PersonalInfo struct {
	FullName    FullName  `json:"full_name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
	Nationality string    `json:"nationality"` // V3 addition
}

// FullName represents structured name information (V3)
type FullName struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"` // V3 addition
	LastName   string `json:"last_name"`
	Prefix     string `json:"prefix"` // V3 addition (Mr., Ms., Dr., etc.)
	Suffix     string `json:"suffix"` // V3 addition (Jr., Sr., III, etc.)
}

// ContactInfo represents restructured contact information (V3)
type ContactInfo struct {
	PrimaryEmail   string          `json:"primary_email"`
	SecondaryEmail string          `json:"secondary_email"` // V3 addition
	PhoneNumbers   []PhoneNumber   `json:"phone_numbers"`   // V3 addition
	Addresses      []Address       `json:"addresses"`       // V3 addition
	SocialProfiles []SocialProfile `json:"social_profiles"` // V3 addition
}

// PhoneNumber represents phone number information (V3)
type PhoneNumber struct {
	Type        string `json:"type"` // mobile, home, work
	CountryCode string `json:"country_code"`
	Number      string `json:"number"`
	IsPrimary   bool   `json:"is_primary"`
	IsVerified  bool   `json:"is_verified"`
}

// Address represents address information (V3)
type Address struct {
	Type       string `json:"type"` // home, work, billing, shipping
	Street1    string `json:"street1"`
	Street2    string `json:"street2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	IsPrimary  bool   `json:"is_primary"`
}

// SocialProfile represents social media profile (V3)
type SocialProfile struct {
	Platform string `json:"platform"` // twitter, linkedin, github, etc.
	Username string `json:"username"`
	URL      string `json:"url"`
}

// EventMetadata represents enhanced event metadata (V3)
type EventMetadata struct {
	Version       string                 `json:"version"`
	SchemaVersion string                 `json:"schema_version"`
	Source        string                 `json:"source"` // web, mobile, api, etc.
	CorrelationID string                 `json:"correlation_id"`
	CausationID   string                 `json:"causation_id"`
	UserAgent     string                 `json:"user_agent"`
	IPAddress     string                 `json:"ip_address"`
	Timestamp     time.Time              `json:"timestamp"`
	Custom        map[string]interface{} `json:"custom"`
}

// UserCreatedV3 represents the V3 version of user creation event
type UserCreatedV3 struct {
	*cqrs.BaseEventMessage
	UserID        string          `json:"user_id"`
	PersonalInfo  PersonalInfo    `json:"personal_info"`  // V3 restructure
	ContactInfo   ContactInfo     `json:"contact_info"`   // V3 restructure
	Preferences   UserPreferences `json:"preferences"`    // Kept from V2
	EventMetadata EventMetadata   `json:"event_metadata"` // V3 addition (renamed to avoid conflict)
}

// NewUserCreatedV3 creates a new V3 user created event
func NewUserCreatedV3(userID string, personalInfo PersonalInfo, contactInfo ContactInfo, preferences UserPreferences, metadata EventMetadata) *UserCreatedV3 {
	return &UserCreatedV3{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserCreated",
			userID,
			"User",
			1,
			map[string]interface{}{
				"version": "3.0",
				"schema":  "user_created_v3",
			},
		),
		UserID:        userID,
		PersonalInfo:  personalInfo,
		ContactInfo:   contactInfo,
		Preferences:   preferences,
		EventMetadata: metadata,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserCreatedV3) EventData() interface{} {
	return map[string]interface{}{
		"user_id":        e.UserID,
		"personal_info":  e.PersonalInfo,
		"contact_info":   e.ContactInfo,
		"preferences":    e.Preferences,
		"event_metadata": e.EventMetadata,
	}
}

// UserUpdatedV3 represents the V3 version of user update event
type UserUpdatedV3 struct {
	*cqrs.BaseEventMessage
	UserID        string          `json:"user_id"`
	PersonalInfo  PersonalInfo    `json:"personal_info"`  // V3 restructure
	ContactInfo   ContactInfo     `json:"contact_info"`   // V3 restructure
	Preferences   UserPreferences `json:"preferences"`    // Kept from V2
	Changes       []FieldChange   `json:"changes"`        // V3 addition
	EventMetadata EventMetadata   `json:"event_metadata"` // V3 addition (renamed to avoid conflict)
}

// FieldChange represents a specific field change (V3)
type FieldChange struct {
	FieldPath string      `json:"field_path"` // e.g., "personal_info.full_name.first_name"
	OldValue  interface{} `json:"old_value"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewUserUpdatedV3 creates a new V3 user updated event
func NewUserUpdatedV3(userID string, personalInfo PersonalInfo, contactInfo ContactInfo, preferences UserPreferences, changes []FieldChange, metadata EventMetadata) *UserUpdatedV3 {
	return &UserUpdatedV3{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserUpdated",
			userID,
			"User",
			1,
			map[string]interface{}{
				"version": "3.0",
				"schema":  "user_updated_v3",
			},
		),
		UserID:        userID,
		PersonalInfo:  personalInfo,
		ContactInfo:   contactInfo,
		Preferences:   preferences,
		Changes:       changes,
		EventMetadata: metadata,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserUpdatedV3) EventData() interface{} {
	return map[string]interface{}{
		"user_id":        e.UserID,
		"personal_info":  e.PersonalInfo,
		"contact_info":   e.ContactInfo,
		"preferences":    e.Preferences,
		"changes":        e.Changes,
		"event_metadata": e.EventMetadata,
	}
}

// UserDeletedV3 represents the V3 version of user deletion event
type UserDeletedV3 struct {
	*cqrs.BaseEventMessage
	UserID         string         `json:"user_id"`
	DeletionReason DeletionReason `json:"deletion_reason"` // V3 restructure
	EventMetadata  EventMetadata  `json:"event_metadata"`  // V3 addition (renamed to avoid conflict)
}

// DeletionReason represents structured deletion reason (V3)
type DeletionReason struct {
	Category    string                 `json:"category"` // user_request, admin_action, policy_violation, etc.
	Reason      string                 `json:"reason"`
	Details     string                 `json:"details"`
	RequestedBy string                 `json:"requested_by"`
	ApprovedBy  string                 `json:"approved_by"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
}

// NewUserDeletedV3 creates a new V3 user deleted event
func NewUserDeletedV3(userID string, deletionReason DeletionReason, metadata EventMetadata) *UserDeletedV3 {
	return &UserDeletedV3{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserDeleted",
			userID,
			"User",
			1,
			map[string]interface{}{
				"version": "3.0",
				"schema":  "user_deleted_v3",
			},
		),
		UserID:         userID,
		DeletionReason: deletionReason,
		EventMetadata:  metadata,
	}
}

// EventData returns the event data for serialization (implements EventMessage interface)
func (e *UserDeletedV3) EventData() interface{} {
	return map[string]interface{}{
		"user_id":         e.UserID,
		"deletion_reason": e.DeletionReason,
		"event_metadata":  e.EventMetadata,
	}
}

// V3EventFactory creates V3 events from raw data
type V3EventFactory struct{}

// CreateEvent creates a V3 event from event type and data
func (f *V3EventFactory) CreateEvent(eventType string, aggregateID string, eventData []byte, metadata map[string]interface{}) (cqrs.EventMessage, error) {
	switch eventType {
	case "UserCreated":
		return f.createUserCreatedV3(aggregateID, eventData, metadata)
	case "UserUpdated":
		return f.createUserUpdatedV3(aggregateID, eventData, metadata)
	case "UserDeleted":
		return f.createUserDeletedV3(aggregateID, eventData, metadata)
	default:
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "unknown event type: "+eventType, nil)
	}
}

// createUserCreatedV3 creates UserCreatedV3 from raw data
func (f *V3EventFactory) createUserCreatedV3(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserCreatedV3, error) {
	var data struct {
		UserID        string          `json:"user_id"`
		PersonalInfo  PersonalInfo    `json:"personal_info"`
		ContactInfo   ContactInfo     `json:"contact_info"`
		Preferences   UserPreferences `json:"preferences"`
		EventMetadata EventMetadata   `json:"event_metadata"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserCreatedV3", err)
	}

	event := &UserCreatedV3{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserCreated",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:        data.UserID,
		PersonalInfo:  data.PersonalInfo,
		ContactInfo:   data.ContactInfo,
		Preferences:   data.Preferences,
		EventMetadata: data.EventMetadata,
	}

	return event, nil
}

// createUserUpdatedV3 creates UserUpdatedV3 from raw data
func (f *V3EventFactory) createUserUpdatedV3(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserUpdatedV3, error) {
	var data struct {
		UserID        string          `json:"user_id"`
		PersonalInfo  PersonalInfo    `json:"personal_info"`
		ContactInfo   ContactInfo     `json:"contact_info"`
		Preferences   UserPreferences `json:"preferences"`
		Changes       []FieldChange   `json:"changes"`
		EventMetadata EventMetadata   `json:"event_metadata"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserUpdatedV3", err)
	}

	event := &UserUpdatedV3{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserUpdated",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:        data.UserID,
		PersonalInfo:  data.PersonalInfo,
		ContactInfo:   data.ContactInfo,
		Preferences:   data.Preferences,
		Changes:       data.Changes,
		EventMetadata: data.EventMetadata,
	}

	return event, nil
}

// createUserDeletedV3 creates UserDeletedV3 from raw data
func (f *V3EventFactory) createUserDeletedV3(aggregateID string, eventData []byte, metadata map[string]interface{}) (*UserDeletedV3, error) {
	var data struct {
		UserID         string         `json:"user_id"`
		DeletionReason DeletionReason `json:"deletion_reason"`
		EventMetadata  EventMetadata  `json:"event_metadata"`
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to unmarshal UserDeletedV3", err)
	}

	event := &UserDeletedV3{
		BaseEventMessage: cqrs.NewBaseEventMessage(
			"UserDeleted",
			aggregateID,
			"User",
			1,
			metadata,
		),
		UserID:         data.UserID,
		DeletionReason: data.DeletionReason,
		EventMetadata:  data.EventMetadata,
	}

	return event, nil
}

// GetVersion returns the version of V3 events
func (f *V3EventFactory) GetVersion() int {
	return 3
}

// GetSupportedEvents returns the list of supported event types
func (f *V3EventFactory) GetSupportedEvents() []string {
	return []string{"UserCreated", "UserUpdated", "UserDeleted"}
}
