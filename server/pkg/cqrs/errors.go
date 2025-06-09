package cqrs

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

// Common CQRS errors
var (
	// Aggregate errors
	ErrAggregateNotFound    = errors.New("aggregate not found")
	ErrInvalidAggregateID   = errors.New("invalid aggregate ID")
	ErrInvalidAggregateType = errors.New("invalid aggregate type")
	ErrInvalidVersion       = errors.New("invalid version")
	ErrConcurrencyConflict  = errors.New("concurrency conflict")

	// Command errors
	ErrInvalidCommand          = errors.New("invalid command")
	ErrCommandHandlerNotFound  = errors.New("command handler not found")
	ErrCommandValidationFailed = errors.New("command validation failed")

	// Query errors
	ErrInvalidQuery          = errors.New("invalid query")
	ErrQueryHandlerNotFound  = errors.New("query handler not found")
	ErrQueryValidationFailed = errors.New("query validation failed")

	// Event errors
	ErrInvalidEvent          = errors.New("invalid event")
	ErrEventHandlerNotFound  = errors.New("event handler not found")
	ErrEventValidationFailed = errors.New("event validation failed")

	// Snapshot errors
	ErrSnapshotNotFound         = errors.New("snapshot not found")
	ErrInvalidSnapshotData      = errors.New("invalid snapshot data")
	ErrSnapshotValidationFailed = errors.New("snapshot validation failed")

	// Repository errors
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrSaveAggregate      = errors.New("failed to save aggregate")
	ErrLoadAggregate      = errors.New("failed to load aggregate")

	// Event Store errors
	ErrEventStoreNotFound = errors.New("event store not found")
	ErrSaveEvents         = errors.New("failed to save events")
	ErrLoadEvents         = errors.New("failed to load events")

	// Event Bus errors
	ErrEventBusNotFound = errors.New("event bus not found")
	ErrPublishEvent     = errors.New("failed to publish event")
	ErrSubscribeEvent   = errors.New("failed to subscribe to event")

	// Serialization errors
	ErrSerializationFailed   = errors.New("serialization failed")
	ErrDeserializationFailed = errors.New("deserialization failed")
	ErrUnsupportedFormat     = errors.New("unsupported serialization format")
)

// CQRSError represents a CQRS-specific error with additional context
type CQRSError struct {
	Code    string
	Message string
	Cause   error
	Context map[string]interface{}
}

func (e *CQRSError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *CQRSError) Unwrap() error {
	return e.Cause
}

// NewCQRSError creates a new CQRSError
func NewCQRSError(code, message string, cause error) *CQRSError {
	return &CQRSError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *CQRSError) WithContext(key string, value interface{}) *CQRSError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// ErrorCode represents CQRS error codes
type ErrorCode int

const (
	ErrCodeAggregateNotFound ErrorCode = iota
	ErrCodeInvalidAggregate
	ErrCodeConcurrencyConflict
	ErrCodeCommandValidation
	ErrCodeQueryValidation
	ErrCodeEventValidation
	ErrCodeSerializationError
	ErrCodeRepositoryError
	ErrCodeEventStoreError
	ErrCodeEventBusError
	ErrCodeSnapshotValidationFailed
	ErrCodeStateStoreError
	ErrCodeSnapshotStoreError
	ErrCodeReadStoreError
	ErrCodeSnapshotNotFound
	ErrCodeReadModelNotFound
	ErrCodeValidationError
	ErrCodeNotFoundError
)

func (ec ErrorCode) String() string {
	switch ec {
	case ErrCodeAggregateNotFound:
		return "AGGREGATE_NOT_FOUND"
	case ErrCodeInvalidAggregate:
		return "INVALID_AGGREGATE"
	case ErrCodeConcurrencyConflict:
		return "CONCURRENCY_CONFLICT"
	case ErrCodeCommandValidation:
		return "COMMAND_VALIDATION"
	case ErrCodeQueryValidation:
		return "QUERY_VALIDATION"
	case ErrCodeEventValidation:
		return "EVENT_VALIDATION"
	case ErrCodeSerializationError:
		return "SERIALIZATION_ERROR"
	case ErrCodeRepositoryError:
		return "REPOSITORY_ERROR"
	case ErrCodeEventStoreError:
		return "EVENT_STORE_ERROR"
	case ErrCodeEventBusError:
		return "EVENT_BUS_ERROR"
	case ErrCodeSnapshotValidationFailed:
		return "SNAPSHOT_VALIDATION_FAILED"
	case ErrCodeStateStoreError:
		return "STATE_STORE_ERROR"
	case ErrCodeSnapshotStoreError:
		return "SNAPSHOT_STORE_ERROR"
	case ErrCodeReadStoreError:
		return "READ_STORE_ERROR"
	case ErrCodeSnapshotNotFound:
		return "SNAPSHOT_NOT_FOUND"
	case ErrCodeReadModelNotFound:
		return "READ_MODEL_NOT_FOUND"
	case ErrCodeValidationError:
		return "VALIDATION_ERROR"
	case ErrCodeNotFoundError:
		return "NOT_FOUND_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// IsNotFoundError checks if an error is a "not found" type error
func IsNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for CQRS error codes
	if cqrsErr, ok := err.(*CQRSError); ok {
		switch cqrsErr.Code {
		case ErrCodeAggregateNotFound.String(),
			ErrCodeSnapshotNotFound.String(),
			ErrCodeReadModelNotFound.String():
			return true
		}
	}

	// Check for standard errors
	if errors.Is(err, ErrAggregateNotFound) {
		return true
	}

	return false
}

// Helper function for checksum calculation
func calculateDataChecksum(aggregateID, aggregateType string, version int, data interface{}) string {
	input := fmt.Sprintf("%s:%s:%d:%v", aggregateID, aggregateType, version, data)
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}
