package redisstream

import (
	"errors"
	"fmt"
)

// Common Redis Stream EventBus errors
var (
	ErrStreamNotFound            = errors.New("stream not found")
	ErrConsumerNotFound          = errors.New("consumer not found")
	ErrConnectionFailed          = errors.New("redis connection failed")
	ErrEventBusNotRunning        = errors.New("event bus is not running")
	ErrEventBusAlreadyRunning    = errors.New("event bus is already running")
	ErrInvalidConfiguration      = errors.New("invalid configuration")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrPublishFailed             = errors.New("event publish failed")
	ErrHandlerRegistrationFailed = errors.New("handler registration failed")
	ErrMaxRetriesExceeded        = errors.New("maximum retry attempts exceeded")
	ErrTimeoutExceeded           = errors.New("operation timeout exceeded")
	ErrCircuitBreakerOpen        = errors.New("circuit breaker is open")

	// Serialization errors
	ErrSerializationFailed                  = errors.New("event serialization failed")
	ErrDeserializationFailed                = errors.New("event deserialization failed")
	ErrUnsupportedFormat                    = errors.New("unsupported serialization format")
	ErrSerializerNotFound                   = errors.New("serializer not found")
	ErrSerializerNotSet                     = errors.New("serializer not set")
	ErrSerializationFormatAlreadyRegistered = errors.New("serialization format already registered")
	ErrSerializationFormatNotSupported      = errors.New("serialization format not supported")

	// Priority stream errors
	ErrPriorityStreamDisabled = errors.New("priority streams are disabled")
	ErrInvalidPriority        = errors.New("invalid event priority")

	// DLQ errors
	ErrDLQDisabled           = errors.New("dead letter queue is disabled")
	ErrDLQOperationFailed    = errors.New("DLQ operation failed")
	ErrDLQReprocessingFailed = errors.New("DLQ reprocessing failed")

	// Retry policy errors
	ErrRetryPolicyInvalid = errors.New("invalid retry policy")

	// Circuit breaker errors
	ErrCircuitBreakerInvalid = errors.New("invalid circuit breaker configuration")

	// Health check errors
	ErrHealthCheckFailed   = errors.New("health check failed")
	ErrHealthCheckTimeout  = errors.New("health check timeout")
	ErrHealthCheckDisabled = errors.New("health checks are disabled")
)

// ErrConfigInvalid creates a new configuration validation error
func ErrConfigInvalid(reason string) error {
	return fmt.Errorf("%w: %s", ErrInvalidConfiguration, reason)
}

// ErrStreamOperation creates a new stream operation error
func ErrStreamOperation(operation string, cause error) error {
	return fmt.Errorf("stream operation '%s' failed: %w", operation, cause)
}

// ErrConsumerOperation creates a new consumer operation error
func ErrConsumerOperation(consumerID string, operation string, cause error) error {
	return fmt.Errorf("consumer '%s' operation '%s' failed: %w", consumerID, operation, cause)
}

// ErrEventProcessing creates a new event processing error
func ErrEventProcessing(eventID string, cause error) error {
	return fmt.Errorf("event processing failed for event '%s': %w", eventID, cause)
}

// ErrRetryExhausted creates a new retry exhausted error
func ErrRetryExhausted(eventID string, attempts int) error {
	return fmt.Errorf("%w: event '%s' failed after %d attempts", ErrMaxRetriesExceeded, eventID, attempts)
}

// ErrDLQOperation creates a new DLQ operation error
func ErrDLQOperation(operation string, cause error) error {
	return fmt.Errorf("%w: operation '%s' failed: %v", ErrDLQOperationFailed, operation, cause)
}

// ErrPriorityOperation creates a new priority operation error
func ErrPriorityOperation(operation string, cause error) error {
	return fmt.Errorf("priority operation '%s' failed: %w", operation, cause)
}

// ErrCircuitBreakerOperation creates a new circuit breaker operation error
func ErrCircuitBreakerOperation(serviceName string, operation string, cause error) error {
	return fmt.Errorf("circuit breaker '%s' operation '%s' failed: %w", serviceName, operation, cause)
}

// ErrHealthCheckOperation creates a new health check operation error
func ErrHealthCheckOperation(checkName string, cause error) error {
	return fmt.Errorf("%w: check '%s' failed: %v", ErrHealthCheckFailed, checkName, cause)
}

// ErrSerializationOperation creates a new serialization operation error
func ErrSerializationOperation(format string, operation string, cause error) error {
	return fmt.Errorf("serialization format '%s' operation '%s' failed: %w", format, operation, cause)
}
