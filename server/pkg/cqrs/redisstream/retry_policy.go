package redisstream

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// RetryPolicyManager manages retry policies and retry logic
type RetryPolicyManager interface {
	// Retry decision
	ShouldRetry(event cqrs.EventMessage, error *ProcessingError) bool
	CalculateDelay(policy *RetryPolicy, attempt int) time.Duration

	// Event enrichment
	EnrichEventForRetry(event cqrs.EventMessage, error *ProcessingError) cqrs.EventMessage

	// Policy management
	GetDefaultRetryPolicy() *RetryPolicy
	SetHandlerRetryPolicy(handlerName string, policy *RetryPolicy) error
	GetHandlerRetryPolicy(handlerName string) *RetryPolicy
	SetEventTypeRetryPolicy(eventType string, policy *RetryPolicy) error
	GetEventTypeRetryPolicy(eventType string) *RetryPolicy

	// Statistics and monitoring
	RecordRetryAttempt(streamName, handlerName string, attempt int, reason string)
	RecordRetrySuccess(streamName, handlerName string, finalAttempt int)
	RecordRetryExhausted(streamName, handlerName string, finalAttempt int, reason string)
	GetRetryStatistics() *RetryStatistics
	GetRetrySuccessRate(streamName string) float64
	GetOverallRetrySuccessRate() float64
	GetTopRetryReasons(limit int) []*RetryReasonStats
}

// RetryPolicy defines retry behavior for event processing
type RetryPolicy struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffType   cqrs.BackoffType
	BackoffFactor float64
}

// BackoffType represents retry backoff strategies (moved to main types)
// type BackoffType int  // Already defined in event_bus.go

// RetryStatistics contains statistics about retry operations
type RetryStatistics struct {
	TotalRetryAttempts int64
	SuccessfulRetries  int64
	ExhaustedRetries   int64
	RetriesByStream    map[string]int64
	RetriesByHandler   map[string]int64
	RetriesByReason    map[string]int64
	LastUpdated        time.Time
}

// RetryReasonStats contains statistics for a specific retry reason
type RetryReasonStats struct {
	Reason string
	Count  int64
	Rate   float64
}

// retryPolicyManager implements RetryPolicyManager
type retryPolicyManager struct {
	config            *RedisStreamConfig
	defaultPolicy     *RetryPolicy
	handlerPolicies   map[string]*RetryPolicy
	eventTypePolicies map[string]*RetryPolicy
	stats             *RetryStatistics
	mu                sync.RWMutex
}

// NewRetryPolicyManager creates a new retry policy manager
func NewRetryPolicyManager(config *RedisStreamConfig) (RetryPolicyManager, error) {
	if config == nil {
		return nil, ErrConfigInvalid("config cannot be nil")
	}

	if err := validateRetryConfig(&config.Retry); err != nil {
		return nil, err
	}

	defaultPolicy := &RetryPolicy{
		MaxAttempts:   config.Retry.MaxAttempts,
		InitialDelay:  config.Retry.InitialDelay,
		MaxDelay:      config.Retry.MaxDelay,
		BackoffType:   parseBackoffType(config.Retry.BackoffType),
		BackoffFactor: config.Retry.BackoffFactor,
	}

	return &retryPolicyManager{
		config:            config,
		defaultPolicy:     defaultPolicy,
		handlerPolicies:   make(map[string]*RetryPolicy),
		eventTypePolicies: make(map[string]*RetryPolicy),
		stats: &RetryStatistics{
			RetriesByStream:  make(map[string]int64),
			RetriesByHandler: make(map[string]int64),
			RetriesByReason:  make(map[string]int64),
			LastUpdated:      time.Now(),
		},
	}, nil
}

// ShouldRetry determines if an event should be retried
func (m *retryPolicyManager) ShouldRetry(event cqrs.EventMessage, error *ProcessingError) bool {
	// Check if error is retryable
	if !m.isRetryableError(error) {
		return false
	}

	// Get retry count from event metadata
	retryCount := m.getRetryCount(event)

	// Get applicable retry policy
	policy := m.getApplicablePolicy(event, error.Handler)

	// Check if we've exhausted retries
	return retryCount < policy.MaxAttempts
}

// CalculateDelay calculates the delay for a retry attempt
func (m *retryPolicyManager) CalculateDelay(policy *RetryPolicy, attempt int) time.Duration {
	if attempt <= 0 {
		return policy.InitialDelay
	}

	var delay time.Duration

	switch policy.BackoffType {
	case cqrs.FixedBackoff:
		delay = policy.InitialDelay

	case cqrs.ExponentialBackoff:
		// delay = initialDelay * (backoffFactor ^ (attempt - 1))
		multiplier := math.Pow(policy.BackoffFactor, float64(attempt-1))
		delay = time.Duration(float64(policy.InitialDelay) * multiplier)

	case cqrs.LinearBackoff:
		// delay = initialDelay + (initialDelay * backoffFactor * (attempt - 1))
		additive := time.Duration(float64(policy.InitialDelay) * policy.BackoffFactor * float64(attempt-1))
		delay = policy.InitialDelay + additive

	default:
		delay = policy.InitialDelay
	}

	// Cap at max delay
	if delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}

	return delay
}

// EnrichEventForRetry adds retry metadata to an event
func (m *retryPolicyManager) EnrichEventForRetry(event cqrs.EventMessage, error *ProcessingError) cqrs.EventMessage {
	// Clone the event
	enrichedEvent := event.Clone()

	// Get current retry count
	currentRetryCount := m.getRetryCount(event)
	newRetryCount := currentRetryCount + 1

	// Get applicable policy
	policy := m.getApplicablePolicy(event, error.Handler)

	// Prepare retry metadata
	retryMetadata := map[string]interface{}{
		"retry_count":          newRetryCount,
		"max_retries":          policy.MaxAttempts,
		"last_error":           error.Error,
		"last_retry_timestamp": time.Now().Format(time.RFC3339Nano),
		"retry_handler":        error.Handler,
	}

	// Add first failure timestamp if this is the first retry
	if currentRetryCount == 0 {
		retryMetadata["first_failure"] = error.Timestamp.Format(time.RFC3339Nano)
	}

	// Update retry history
	retryHistory := m.getRetryHistory(event)
	retryHistory = append(retryHistory, map[string]interface{}{
		"attempt":   newRetryCount,
		"error":     error.Error,
		"timestamp": error.Timestamp.Format(time.RFC3339Nano),
		"handler":   error.Handler,
	})
	retryMetadata["retry_history"] = retryHistory

	// Get existing metadata
	metadata := enrichedEvent.Metadata()
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Merge retry metadata
	for key, value := range retryMetadata {
		metadata[key] = value
	}

	// Create new event with enriched metadata
	return m.createEventWithMetadata(enrichedEvent, metadata)
}

// GetDefaultRetryPolicy returns the default retry policy
func (m *retryPolicyManager) GetDefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:   m.defaultPolicy.MaxAttempts,
		InitialDelay:  m.defaultPolicy.InitialDelay,
		MaxDelay:      m.defaultPolicy.MaxDelay,
		BackoffType:   m.defaultPolicy.BackoffType,
		BackoffFactor: m.defaultPolicy.BackoffFactor,
	}
}

// SetHandlerRetryPolicy sets a custom retry policy for a specific handler
func (m *retryPolicyManager) SetHandlerRetryPolicy(handlerName string, policy *RetryPolicy) error {
	if handlerName == "" {
		return fmt.Errorf("handler name cannot be empty")
	}
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}
	if err := validateRetryPolicy(policy); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to avoid mutation
	m.handlerPolicies[handlerName] = &RetryPolicy{
		MaxAttempts:   policy.MaxAttempts,
		InitialDelay:  policy.InitialDelay,
		MaxDelay:      policy.MaxDelay,
		BackoffType:   policy.BackoffType,
		BackoffFactor: policy.BackoffFactor,
	}

	return nil
}

// GetHandlerRetryPolicy gets the retry policy for a specific handler
func (m *retryPolicyManager) GetHandlerRetryPolicy(handlerName string) *RetryPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if policy, exists := m.handlerPolicies[handlerName]; exists {
		// Return a copy to prevent mutation
		return &RetryPolicy{
			MaxAttempts:   policy.MaxAttempts,
			InitialDelay:  policy.InitialDelay,
			MaxDelay:      policy.MaxDelay,
			BackoffType:   policy.BackoffType,
			BackoffFactor: policy.BackoffFactor,
		}
	}

	return m.GetDefaultRetryPolicy()
}

// SetEventTypeRetryPolicy sets a custom retry policy for a specific event type
func (m *retryPolicyManager) SetEventTypeRetryPolicy(eventType string, policy *RetryPolicy) error {
	if eventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}
	if err := validateRetryPolicy(policy); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventTypePolicies[eventType] = &RetryPolicy{
		MaxAttempts:   policy.MaxAttempts,
		InitialDelay:  policy.InitialDelay,
		MaxDelay:      policy.MaxDelay,
		BackoffType:   policy.BackoffType,
		BackoffFactor: policy.BackoffFactor,
	}

	return nil
}

// GetEventTypeRetryPolicy gets the retry policy for a specific event type
func (m *retryPolicyManager) GetEventTypeRetryPolicy(eventType string) *RetryPolicy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if policy, exists := m.eventTypePolicies[eventType]; exists {
		return &RetryPolicy{
			MaxAttempts:   policy.MaxAttempts,
			InitialDelay:  policy.InitialDelay,
			MaxDelay:      policy.MaxDelay,
			BackoffType:   policy.BackoffType,
			BackoffFactor: policy.BackoffFactor,
		}
	}

	return m.GetDefaultRetryPolicy()
}

// RecordRetryAttempt records a retry attempt for statistics
func (m *retryPolicyManager) RecordRetryAttempt(streamName, handlerName string, attempt int, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalRetryAttempts++
	m.stats.RetriesByStream[streamName]++
	m.stats.RetriesByHandler[handlerName]++
	m.stats.RetriesByReason[reason]++
	m.stats.LastUpdated = time.Now()
}

// RecordRetrySuccess records a successful retry
func (m *retryPolicyManager) RecordRetrySuccess(streamName, handlerName string, finalAttempt int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.SuccessfulRetries++
	m.stats.LastUpdated = time.Now()
}

// RecordRetryExhausted records when retries are exhausted
func (m *retryPolicyManager) RecordRetryExhausted(streamName, handlerName string, finalAttempt int, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.ExhaustedRetries++
	m.stats.LastUpdated = time.Now()
}

// GetRetryStatistics returns current retry statistics
func (m *retryPolicyManager) GetRetryStatistics() *RetryStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &RetryStatistics{
		TotalRetryAttempts: m.stats.TotalRetryAttempts,
		SuccessfulRetries:  m.stats.SuccessfulRetries,
		ExhaustedRetries:   m.stats.ExhaustedRetries,
		RetriesByStream:    m.copyInt64Map(m.stats.RetriesByStream),
		RetriesByHandler:   m.copyInt64Map(m.stats.RetriesByHandler),
		RetriesByReason:    m.copyInt64Map(m.stats.RetriesByReason),
		LastUpdated:        m.stats.LastUpdated,
	}
}

// GetRetrySuccessRate returns the retry success rate for a specific stream
func (m *retryPolicyManager) GetRetrySuccessRate(streamName string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Note: This is a simplified calculation
	// In a real implementation, you'd track success/failure per stream
	totalResolved := m.stats.SuccessfulRetries + m.stats.ExhaustedRetries
	if totalResolved == 0 {
		return 0.0
	}

	return float64(m.stats.SuccessfulRetries) / float64(totalResolved)
}

// GetOverallRetrySuccessRate returns the overall retry success rate
func (m *retryPolicyManager) GetOverallRetrySuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalResolved := m.stats.SuccessfulRetries + m.stats.ExhaustedRetries
	if totalResolved == 0 {
		return 0.0
	}

	return float64(m.stats.SuccessfulRetries) / float64(totalResolved)
}

// GetTopRetryReasons returns the top retry reasons by frequency
func (m *retryPolicyManager) GetTopRetryReasons(limit int) []*RetryReasonStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var reasons []*RetryReasonStats
	for reason, count := range m.stats.RetriesByReason {
		rate := float64(count) / float64(m.stats.TotalRetryAttempts)
		reasons = append(reasons, &RetryReasonStats{
			Reason: reason,
			Count:  count,
			Rate:   rate,
		})
	}

	// Sort by count (descending)
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i].Count > reasons[j].Count
	})

	// Limit results
	if limit > 0 && len(reasons) > limit {
		reasons = reasons[:limit]
	}

	return reasons
}

// Helper methods

func (m *retryPolicyManager) isRetryableError(error *ProcessingError) bool {
	errorMsg := strings.ToLower(error.Error)

	// Non-retryable errors (validation, business logic, etc.)
	nonRetryablePatterns := []string{
		"validation",
		"invalid",
		"malformed",
		"unauthorized",
		"forbidden",
		"not found",
		"conflict",
		"duplicate",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return false
		}
	}

	// Retryable errors (network, timeouts, temporary issues)
	retryablePatterns := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"unavailable",
		"overloaded",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errorMsg, pattern) {
			return true
		}
	}

	// Default to retryable for unknown errors
	return true
}

func (m *retryPolicyManager) getRetryCount(event cqrs.EventMessage) int {
	metadata := event.Metadata()
	if metadata == nil {
		return 0
	}

	retryCountRaw, exists := metadata["retry_count"]
	if !exists {
		return 0
	}

	retryCount, err := m.convertToInt(retryCountRaw)
	if err != nil {
		return 0
	}

	return retryCount
}

func (m *retryPolicyManager) getRetryHistory(event cqrs.EventMessage) []map[string]interface{} {
	metadata := event.Metadata()
	if metadata == nil {
		return []map[string]interface{}{}
	}

	historyRaw, exists := metadata["retry_history"]
	if !exists {
		return []map[string]interface{}{}
	}

	history, ok := historyRaw.([]map[string]interface{})
	if !ok {
		return []map[string]interface{}{}
	}

	return history
}

func (m *retryPolicyManager) getApplicablePolicy(event cqrs.EventMessage, handlerName string) *RetryPolicy {
	// Check for handler-specific policy first
	if policy := m.GetHandlerRetryPolicy(handlerName); policy != nil {
		// Only return if it's not the default policy
		defaultPolicy := m.GetDefaultRetryPolicy()
		if policy.MaxAttempts != defaultPolicy.MaxAttempts ||
			policy.InitialDelay != defaultPolicy.InitialDelay ||
			policy.BackoffType != defaultPolicy.BackoffType {
			return policy
		}
	}

	// Check for event-type-specific policy
	if policy := m.GetEventTypeRetryPolicy(event.EventType()); policy != nil {
		defaultPolicy := m.GetDefaultRetryPolicy()
		if policy.MaxAttempts != defaultPolicy.MaxAttempts ||
			policy.InitialDelay != defaultPolicy.InitialDelay ||
			policy.BackoffType != defaultPolicy.BackoffType {
			return policy
		}
	}

	// Fall back to default policy
	return m.GetDefaultRetryPolicy()
}

func (m *retryPolicyManager) createEventWithMetadata(originalEvent cqrs.EventMessage, metadata map[string]interface{}) cqrs.EventMessage {
	// Similar to DLQ manager's approach
	baseOptions := cqrs.Options().
		WithEventID(originalEvent.EventID()).
		WithAggregateID(originalEvent.ID()).
		WithAggregateType(originalEvent.Type()).
		WithVersion(originalEvent.Version()).
		WithTimestamp(originalEvent.Timestamp()).
		WithMetadata(metadata)

	if domainEvent, ok := originalEvent.(cqrs.DomainEventMessage); ok {
		domainOptions := &cqrs.BaseDomainEventMessageOptions{}

		issuerID := domainEvent.IssuerID()
		if issuerID != "" {
			domainOptions.IssuerID = &issuerID
		}

		issuerType := domainEvent.IssuerType()
		domainOptions.IssuerType = &issuerType

		causationID := domainEvent.CausationID()
		if causationID != "" {
			domainOptions.CausationID = &causationID
		}

		correlationID := domainEvent.CorrelationID()
		if correlationID != "" {
			domainOptions.CorrelationID = &correlationID
		}

		category := domainEvent.GetEventCategory()
		domainOptions.Category = &category

		priority := domainEvent.GetPriority()
		domainOptions.Priority = &priority

		return cqrs.NewBaseDomainEventMessage(
			originalEvent.EventType(),
			originalEvent.EventData(),
			[]*cqrs.BaseEventMessageOptions{baseOptions},
			domainOptions,
		)
	}

	return cqrs.NewBaseEventMessage(
		originalEvent.EventType(),
		originalEvent.EventData(),
		baseOptions,
	)
}

func (m *retryPolicyManager) convertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

func (m *retryPolicyManager) copyInt64Map(original map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for key, value := range original {
		copy[key] = value
	}
	return copy
}

// Validation functions

func validateRetryConfig(config *RetryConfig) error {
	if config.MaxAttempts <= 0 {
		return fmt.Errorf("max attempts must be greater than 0")
	}
	if config.InitialDelay <= 0 {
		return fmt.Errorf("initial delay must be greater than 0")
	}
	if config.MaxDelay < config.InitialDelay {
		return fmt.Errorf("max delay must be greater than or equal to initial delay")
	}
	if config.BackoffFactor <= 0 {
		return fmt.Errorf("backoff factor must be greater than 0")
	}
	return nil
}

func validateRetryPolicy(policy *RetryPolicy) error {
	if policy.MaxAttempts <= 0 {
		return fmt.Errorf("max attempts must be greater than 0")
	}
	if policy.InitialDelay <= 0 {
		return fmt.Errorf("initial delay must be greater than 0")
	}
	if policy.MaxDelay < policy.InitialDelay {
		return fmt.Errorf("max delay must be greater than or equal to initial delay")
	}
	if policy.BackoffFactor <= 0 {
		return fmt.Errorf("backoff factor must be greater than 0")
	}
	return nil
}

func parseBackoffType(backoffType string) cqrs.BackoffType {
	switch strings.ToLower(backoffType) {
	case "fixed":
		return cqrs.FixedBackoff
	case "exponential":
		return cqrs.ExponentialBackoff
	case "linear":
		return cqrs.LinearBackoff
	default:
		return cqrs.FixedBackoff
	}
}
