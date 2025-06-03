package cqrs

import "time"

// EventMessage interface compatible with go.cqrs framework
type EventMessage interface {
	// Basic event information
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	Version() int

	// Event data (serialization handled separately)
	EventData() interface{}

	// Metadata
	Metadata() map[string]interface{}
	Timestamp() time.Time
}

// EventCategory represents different types of events
type EventCategory int

const (
	UserAction EventCategory = iota
	SystemEvent
	IntegrationEvent
	DomainEvent
)

func (ec EventCategory) String() string {
	switch ec {
	case UserAction:
		return "user_action"
	case SystemEvent:
		return "system_event"
	case IntegrationEvent:
		return "integration_event"
	case DomainEvent:
		return "domain_event"
	default:
		return "unknown"
	}
}

// EventPriority represents event processing priority
type EventPriority int

const (
	Low EventPriority = iota
	Normal
	High
	Critical
)

func (ep EventPriority) String() string {
	switch ep {
	case Low:
		return "low"
	case Normal:
		return "normal"
	case High:
		return "high"
	case Critical:
		return "critical"
	default:
		return "unknown"
	}
}

// IssuerType represents the type of entity that issued the event
type IssuerType int

const (
	UserIssuer      IssuerType = iota // Regular user/player
	SystemIssuer                      // System/AI/Game engine
	AdminIssuer                       // Administrator
	ServiceIssuer                     // External service
	SchedulerIssuer                   // Scheduler/Cron job
)

func (it IssuerType) String() string {
	switch it {
	case UserIssuer:
		return "user"
	case SystemIssuer:
		return "system"
	case AdminIssuer:
		return "admin"
	case ServiceIssuer:
		return "service"
	case SchedulerIssuer:
		return "scheduler"
	default:
		return "unknown"
	}
}

// DomainEventMessage extends EventMessage with Defense Allies specific features
type DomainEventMessage interface {
	EventMessage

	// Event issuer information (improved from UserID)
	IssuerID() string       // ID of the entity that issued this event
	IssuerType() IssuerType // Type of the issuer (user, system, admin, etc.)

	// Additional domain information
	CausationID() string   // ID of the command that caused this event
	CorrelationID() string // Correlation tracking ID

	// Event classification
	GetEventCategory() EventCategory // Event category
	GetPriority() EventPriority      // Event priority

	// Validation and security
	ValidateEvent() error // Event validation
	GetChecksum() string  // Integrity verification
}
