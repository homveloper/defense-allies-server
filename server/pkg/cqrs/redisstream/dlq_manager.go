package redisstream

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"defense-allies-server/pkg/cqrs"
)

// DLQManager manages Dead Letter Queue operations
type DLQManager interface {
	// DLQ stream management
	GetDLQStreamName(originalStream string) string
	GetDLQConsumerGroupName(originalGroup string) string

	// Failure detection and handling
	ShouldMoveToDLQ(event cqrs.EventMessage) bool
	EnrichEventForDLQ(event cqrs.EventMessage, error *ProcessingError) cqrs.EventMessage

	// Statistics and monitoring
	RecordDLQEvent(streamName, handlerName, reason string)
	RecordProcessedEvent(streamName string)
	GetDLQStatistics() *DLQStatistics
	GetDLQRate(streamName string) float64
	GetOverallDLQRate() float64
	GetTopErrorReasons(limit int) []*ErrorReasonStats

	// Configuration
	IsDLQEnabled() bool
}

// ProcessingError represents an error that occurred during event processing
type ProcessingError struct {
	Error      string
	Handler    string
	Timestamp  time.Time
	RetryCount int
	StreamName string
	MessageID  string
	StackTrace string
}

// DLQStatistics contains statistics about DLQ usage
type DLQStatistics struct {
	TotalDLQEvents  int64
	EventsByStream  map[string]int64
	EventsByHandler map[string]int64
	EventsByReason  map[string]int64
	ProcessedEvents map[string]int64 // For calculating rates
	LastUpdated     time.Time
}

// ErrorReasonStats contains statistics for a specific error reason
type ErrorReasonStats struct {
	Reason string
	Count  int64
	Rate   float64
}

// dlqManager implements DLQManager
type dlqManager struct {
	config *RedisStreamConfig
	stats  *DLQStatistics
	mu     sync.RWMutex
}

// NewDLQManager creates a new DLQ manager
func NewDLQManager(config *RedisStreamConfig) (DLQManager, error) {
	if config == nil {
		return nil, ErrConfigInvalid("config cannot be nil")
	}

	return &dlqManager{
		config: config,
		stats: &DLQStatistics{
			EventsByStream:  make(map[string]int64),
			EventsByHandler: make(map[string]int64),
			EventsByReason:  make(map[string]int64),
			ProcessedEvents: make(map[string]int64),
			LastUpdated:     time.Now(),
		},
	}, nil
}

// GetDLQStreamName generates DLQ stream name from original stream
func (m *dlqManager) GetDLQStreamName(originalStream string) string {
	if !m.config.Stream.DLQEnabled {
		return ""
	}

	if originalStream == "" {
		return ""
	}

	return fmt.Sprintf("%s%s%s", originalStream, m.config.Stream.NamespaceDelim, m.config.Stream.DLQSuffix)
}

// GetDLQConsumerGroupName generates DLQ consumer group name
func (m *dlqManager) GetDLQConsumerGroupName(originalGroup string) string {
	if !m.config.Stream.DLQEnabled {
		return ""
	}

	if originalGroup == "" {
		return ""
	}

	return fmt.Sprintf("%s_%s", originalGroup, m.config.Stream.DLQSuffix)
}

// ShouldMoveToDLQ determines if an event should be moved to DLQ
func (m *dlqManager) ShouldMoveToDLQ(event cqrs.EventMessage) bool {
	if !m.config.Stream.DLQEnabled {
		return false
	}

	metadata := event.Metadata()
	if metadata == nil {
		return false
	}

	// Check retry count
	retryCountRaw, hasRetryCount := metadata["retry_count"]
	maxRetriesRaw, hasMaxRetries := metadata["max_retries"]

	if !hasRetryCount || !hasMaxRetries {
		return false
	}

	retryCount, err := m.convertToInt(retryCountRaw)
	if err != nil {
		return false
	}

	maxRetries, err := m.convertToInt(maxRetriesRaw)
	if err != nil {
		// Use config default if metadata is invalid
		maxRetries = m.config.Retry.MaxAttempts
	}

	// Move to DLQ if retry count >= max retries
	return retryCount >= maxRetries
}

// EnrichEventForDLQ adds DLQ-specific metadata to an event
func (m *dlqManager) EnrichEventForDLQ(event cqrs.EventMessage, processingError *ProcessingError) cqrs.EventMessage {
	// Clone the event to avoid modifying the original
	enrichedEvent := event.Clone()

	// Get existing metadata or create new
	metadata := enrichedEvent.Metadata()
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Add DLQ-specific metadata
	dlqMetadata := map[string]interface{}{
		"dlq_reason":              "max_retries_exceeded",
		"dlq_timestamp":           time.Now().Format(time.RFC3339Nano),
		"dlq_original_stream":     processingError.StreamName,
		"dlq_original_handler":    processingError.Handler,
		"dlq_retry_count":         processingError.RetryCount,
		"dlq_original_error":      processingError.Error,
		"dlq_original_message_id": processingError.MessageID,
	}

	if processingError.StackTrace != "" {
		dlqMetadata["dlq_stack_trace"] = processingError.StackTrace
	}

	// Merge DLQ metadata with existing metadata
	for key, value := range dlqMetadata {
		metadata[key] = value
	}

	// Create new event with enriched metadata
	if _, ok := enrichedEvent.(*cqrs.BaseEventMessage); ok {
		// For BaseEventMessage, we need to recreate with new metadata
		baseOptions := cqrs.Options().
			WithEventID(enrichedEvent.EventID()).
			WithAggregateID(enrichedEvent.ID()).
			WithAggregateType(enrichedEvent.Type()).
			WithVersion(enrichedEvent.Version()).
			WithTimestamp(enrichedEvent.Timestamp()).
			WithMetadata(metadata)

		if domainEvent, isDomain := enrichedEvent.(cqrs.DomainEventMessage); isDomain {
			// Preserve domain event properties
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
				enrichedEvent.EventType(),
				enrichedEvent.EventData(),
				[]*cqrs.BaseEventMessageOptions{baseOptions},
				domainOptions,
			)
		}

		return cqrs.NewBaseEventMessage(
			enrichedEvent.EventType(),
			enrichedEvent.EventData(),
			baseOptions,
		)
	}

	// For other event types, try to add metadata directly (this is a fallback)
	if baseEvent, ok := enrichedEvent.(*cqrs.BaseEventMessage); ok {
		for key, value := range dlqMetadata {
			baseEvent.AddMetadata(key, value)
		}
	}

	return enrichedEvent
}

// RecordDLQEvent records a DLQ event for statistics
func (m *dlqManager) RecordDLQEvent(streamName, handlerName, reason string) {
	if !m.config.Stream.DLQEnabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.TotalDLQEvents++
	m.stats.EventsByStream[streamName]++
	m.stats.EventsByHandler[handlerName]++
	m.stats.EventsByReason[reason]++
	m.stats.LastUpdated = time.Now()
}

// RecordProcessedEvent records a successfully processed event
func (m *dlqManager) RecordProcessedEvent(streamName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats.ProcessedEvents[streamName]++
	m.stats.LastUpdated = time.Now()
}

// GetDLQStatistics returns current DLQ statistics
func (m *dlqManager) GetDLQStatistics() *DLQStatistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy to avoid data races
	return &DLQStatistics{
		TotalDLQEvents:  m.stats.TotalDLQEvents,
		EventsByStream:  m.copyInt64Map(m.stats.EventsByStream),
		EventsByHandler: m.copyInt64Map(m.stats.EventsByHandler),
		EventsByReason:  m.copyInt64Map(m.stats.EventsByReason),
		ProcessedEvents: m.copyInt64Map(m.stats.ProcessedEvents),
		LastUpdated:     m.stats.LastUpdated,
	}
}

// GetDLQRate returns the DLQ rate for a specific stream
func (m *dlqManager) GetDLQRate(streamName string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	dlqEvents := m.stats.EventsByStream[streamName]
	processedEvents := m.stats.ProcessedEvents[streamName]
	totalEvents := dlqEvents + processedEvents

	if totalEvents == 0 {
		return 0.0
	}

	return float64(dlqEvents) / float64(totalEvents)
}

// GetOverallDLQRate returns the overall DLQ rate across all streams
func (m *dlqManager) GetOverallDLQRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var totalDLQEvents int64
	var totalProcessedEvents int64

	for _, count := range m.stats.EventsByStream {
		totalDLQEvents += count
	}

	for _, count := range m.stats.ProcessedEvents {
		totalProcessedEvents += count
	}

	totalEvents := totalDLQEvents + totalProcessedEvents
	if totalEvents == 0 {
		return 0.0
	}

	return float64(totalDLQEvents) / float64(totalEvents)
}

// GetTopErrorReasons returns the top error reasons by frequency
func (m *dlqManager) GetTopErrorReasons(limit int) []*ErrorReasonStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var reasons []*ErrorReasonStats
	for reason, count := range m.stats.EventsByReason {
		rate := float64(count) / float64(m.stats.TotalDLQEvents)
		reasons = append(reasons, &ErrorReasonStats{
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

// IsDLQEnabled returns whether DLQ is enabled
func (m *dlqManager) IsDLQEnabled() bool {
	return m.config.Stream.DLQEnabled
}

// Helper methods

func (m *dlqManager) convertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

func (m *dlqManager) copyInt64Map(original map[string]int64) map[string]int64 {
	copy := make(map[string]int64)
	for key, value := range original {
		copy[key] = value
	}
	return copy
}

// DLQReprocessor handles reprocessing of events from DLQ
type DLQReprocessor interface {
	// Reprocess events from DLQ
	ReprocessDLQEvents(streamName string, limit int) (*ReprocessResult, error)
	ReprocessEventByID(streamName, eventID string) error

	// DLQ management
	PurgeDLQStream(streamName string, olderThan time.Duration) (*PurgeResult, error)
	GetDLQStreamInfo(streamName string) (*DLQStreamInfo, error)
}

// ReprocessResult contains the result of DLQ reprocessing
type ReprocessResult struct {
	TotalEvents      int
	SuccessfulEvents int
	FailedEvents     int
	Errors           []error
}

// PurgeResult contains the result of DLQ purging
type PurgeResult struct {
	DeletedEvents int64
	OldestDeleted time.Time
	NewestDeleted time.Time
}

// DLQStreamInfo contains information about a DLQ stream
type DLQStreamInfo struct {
	StreamName     string
	EventCount     int64
	OldestEvent    time.Time
	NewestEvent    time.Time
	ConsumerGroups []string
}

// dlqReprocessor implements DLQReprocessor
type dlqReprocessor struct {
	config     *RedisStreamConfig
	dlqManager DLQManager
}

// NewDLQReprocessor creates a new DLQ reprocessor
func NewDLQReprocessor(config *RedisStreamConfig, dlqManager DLQManager) DLQReprocessor {
	return &dlqReprocessor{
		config:     config,
		dlqManager: dlqManager,
	}
}

// ReprocessDLQEvents reprocesses events from DLQ (placeholder implementation)
func (r *dlqReprocessor) ReprocessDLQEvents(streamName string, limit int) (*ReprocessResult, error) {
	// This would be implemented with actual Redis operations
	// For now, return a placeholder result
	return &ReprocessResult{
		TotalEvents:      0,
		SuccessfulEvents: 0,
		FailedEvents:     0,
		Errors:           nil,
	}, nil
}

// ReprocessEventByID reprocesses a specific event by ID (placeholder implementation)
func (r *dlqReprocessor) ReprocessEventByID(streamName, eventID string) error {
	// This would be implemented with actual Redis operations
	return nil
}

// PurgeDLQStream purges old events from DLQ (placeholder implementation)
func (r *dlqReprocessor) PurgeDLQStream(streamName string, olderThan time.Duration) (*PurgeResult, error) {
	// This would be implemented with actual Redis operations
	return &PurgeResult{
		DeletedEvents: 0,
		OldestDeleted: time.Time{},
		NewestDeleted: time.Time{},
	}, nil
}

// GetDLQStreamInfo gets information about a DLQ stream (placeholder implementation)
func (r *dlqReprocessor) GetDLQStreamInfo(streamName string) (*DLQStreamInfo, error) {
	// This would be implemented with actual Redis operations
	return &DLQStreamInfo{
		StreamName:     streamName,
		EventCount:     0,
		OldestEvent:    time.Time{},
		NewestEvent:    time.Time{},
		ConsumerGroups: []string{},
	}, nil
}
