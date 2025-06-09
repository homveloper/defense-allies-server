package redisstream

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"cqrs"
)

// PriorityStreamManager manages priority-based event stream routing
type PriorityStreamManager interface {
	// Stream naming and routing
	GetStreamName(priority cqrs.EventPriority, category cqrs.EventCategory, partitionKey string) string
	GetConsumerGroupName(priority cqrs.EventPriority, serviceName string, handlerType cqrs.HandlerType) string
	GetRoutingInfo(event cqrs.EventMessage) *StreamRoutingInfo

	// Stream management
	GetStreamsByPriority(category cqrs.EventCategory, partitionKey string) []*StreamInfo
	GetStreamsWithMinPriority(category cqrs.EventCategory, partitionKey string, minPriority cqrs.EventPriority) []*StreamInfo

	// Consumer management
	GetConsumerConfigurations(category cqrs.EventCategory, partitionKey string, handlerType cqrs.HandlerType) []*ConsumerConfiguration

	// Metrics and monitoring
	RecordPublishedEvent(priority cqrs.EventPriority, streamName string)
	RecordProcessedEvent(priority cqrs.EventPriority, streamName string, latency time.Duration)
	GetPriorityMetrics() map[cqrs.EventPriority]*PriorityMetrics
	GetPriorityRatios() map[cqrs.EventPriority]float64

	// Configuration
	IsPriorityEnabled() bool
}

// StreamRoutingInfo contains routing information for an event
type StreamRoutingInfo struct {
	StreamName   string
	Priority     cqrs.EventPriority
	Category     cqrs.EventCategory
	PartitionKey string
}

// StreamInfo contains information about a priority stream
type StreamInfo struct {
	StreamName   string
	Priority     cqrs.EventPriority
	Category     cqrs.EventCategory
	PartitionKey string
}

// ConsumerConfiguration contains consumer setup information
type ConsumerConfiguration struct {
	StreamName    string
	ConsumerGroup string
	Priority      cqrs.EventPriority
	Category      cqrs.EventCategory
	PartitionKey  string
	HandlerType   cqrs.HandlerType
}

// PriorityMetrics tracks metrics for a specific priority level
type PriorityMetrics struct {
	PublishedEvents int64
	ProcessedEvents int64
	FailedEvents    int64
	AverageLatency  time.Duration
	LastEventTime   time.Time

	// Internal for latency calculation
	totalLatency time.Duration
	latencyCount int64
}

// priorityStreamManager implements PriorityStreamManager
type priorityStreamManager struct {
	config  *RedisStreamConfig
	metrics map[cqrs.EventPriority]*PriorityMetrics
	mu      sync.RWMutex
}

// NewPriorityStreamManager creates a new priority stream manager
func NewPriorityStreamManager(config *RedisStreamConfig) (PriorityStreamManager, error) {
	if config == nil {
		return nil, ErrConfigInvalid("config cannot be nil")
	}

	manager := &priorityStreamManager{
		config:  config,
		metrics: make(map[cqrs.EventPriority]*PriorityMetrics),
	}

	// Initialize metrics for all priority levels
	for _, priority := range []cqrs.EventPriority{cqrs.PriorityCritical, cqrs.PriorityHigh, cqrs.PriorityNormal, cqrs.PriorityLow} {
		manager.metrics[priority] = &PriorityMetrics{}
	}

	return manager, nil
}

// GetStreamName generates stream name based on priority and configuration
func (m *priorityStreamManager) GetStreamName(priority cqrs.EventPriority, category cqrs.EventCategory, partitionKey string) string {
	if partitionKey == "" {
		partitionKey = "default"
	}

	// If priority streams are disabled, return simple stream name
	if !m.config.Stream.EnablePriorityStreams {
		return fmt.Sprintf("%s%s%s%s%s",
			m.config.Stream.StreamPrefix,
			m.config.Stream.NamespaceDelim,
			category.String(),
			m.config.Stream.NamespaceDelim,
			partitionKey,
		)
	}

	// Include priority in stream name
	return fmt.Sprintf("%s%s%s%s%s%s%s",
		m.config.Stream.StreamPrefix,
		m.config.Stream.NamespaceDelim,
		category.String(),
		m.config.Stream.NamespaceDelim,
		priority.String(),
		m.config.Stream.NamespaceDelim,
		partitionKey,
	)
}

// GetConsumerGroupName generates consumer group name
func (m *priorityStreamManager) GetConsumerGroupName(priority cqrs.EventPriority, serviceName string, handlerType cqrs.HandlerType) string {
	if !m.config.Stream.EnablePriorityStreams {
		return fmt.Sprintf("%s_%s_cg", serviceName, handlerType.String())
	}

	return fmt.Sprintf("%s_%s_%s_cg", serviceName, handlerType.String(), priority.String())
}

// GetRoutingInfo determines routing information for an event
func (m *priorityStreamManager) GetRoutingInfo(event cqrs.EventMessage) *StreamRoutingInfo {
	priority := cqrs.PriorityNormal // Default priority
	category := cqrs.DomainEvent    // Default category
	partitionKey := event.Type()    // Use aggregate type as partition key

	// Extract priority and category from domain events
	if domainEvent, ok := event.(cqrs.DomainEventMessage); ok {
		priority = domainEvent.GetPriority()
		category = domainEvent.GetEventCategory()
	}

	streamName := m.GetStreamName(priority, category, partitionKey)

	return &StreamRoutingInfo{
		StreamName:   streamName,
		Priority:     priority,
		Category:     category,
		PartitionKey: partitionKey,
	}
}

// GetStreamsByPriority returns streams ordered by priority (highest first)
func (m *priorityStreamManager) GetStreamsByPriority(category cqrs.EventCategory, partitionKey string) []*StreamInfo {
	if !m.config.Stream.EnablePriorityStreams {
		// Return single stream when priority is disabled
		return []*StreamInfo{
			{
				StreamName:   m.GetStreamName(cqrs.PriorityNormal, category, partitionKey),
				Priority:     cqrs.PriorityNormal,
				Category:     category,
				PartitionKey: partitionKey,
			},
		}
	}

	priorities := []cqrs.EventPriority{cqrs.PriorityCritical, cqrs.PriorityHigh, cqrs.PriorityNormal, cqrs.PriorityLow}
	streams := make([]*StreamInfo, len(priorities))

	for i, priority := range priorities {
		streams[i] = &StreamInfo{
			StreamName:   m.GetStreamName(priority, category, partitionKey),
			Priority:     priority,
			Category:     category,
			PartitionKey: partitionKey,
		}
	}

	return streams
}

// GetStreamsWithMinPriority returns streams with priority >= minPriority
func (m *priorityStreamManager) GetStreamsWithMinPriority(category cqrs.EventCategory, partitionKey string, minPriority cqrs.EventPriority) []*StreamInfo {
	allStreams := m.GetStreamsByPriority(category, partitionKey)

	var filteredStreams []*StreamInfo
	for _, stream := range allStreams {
		if stream.Priority >= minPriority {
			filteredStreams = append(filteredStreams, stream)
		}
	}

	return filteredStreams
}

// GetConsumerConfigurations returns consumer configurations for all relevant streams
func (m *priorityStreamManager) GetConsumerConfigurations(category cqrs.EventCategory, partitionKey string, handlerType cqrs.HandlerType) []*ConsumerConfiguration {
	streams := m.GetStreamsByPriority(category, partitionKey)
	configs := make([]*ConsumerConfiguration, len(streams))

	for i, stream := range streams {
		configs[i] = &ConsumerConfiguration{
			StreamName:    stream.StreamName,
			ConsumerGroup: m.GetConsumerGroupName(stream.Priority, m.config.Consumer.ServiceName, handlerType),
			Priority:      stream.Priority,
			Category:      category,
			PartitionKey:  partitionKey,
			HandlerType:   handlerType,
		}
	}

	return configs
}

// RecordPublishedEvent records a published event for metrics
func (m *priorityStreamManager) RecordPublishedEvent(priority cqrs.EventPriority, streamName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.metrics[priority]; exists {
		metrics.PublishedEvents++
		metrics.LastEventTime = time.Now()
	}
}

// RecordProcessedEvent records a processed event with latency for metrics
func (m *priorityStreamManager) RecordProcessedEvent(priority cqrs.EventPriority, streamName string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if metrics, exists := m.metrics[priority]; exists {
		metrics.ProcessedEvents++
		metrics.totalLatency += latency
		metrics.latencyCount++

		// Calculate average latency
		if metrics.latencyCount > 0 {
			metrics.AverageLatency = metrics.totalLatency / time.Duration(metrics.latencyCount)
		}
	}
}

// GetPriorityMetrics returns current metrics for all priorities
func (m *priorityStreamManager) GetPriorityMetrics() map[cqrs.EventPriority]*PriorityMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a deep copy to avoid data races
	result := make(map[cqrs.EventPriority]*PriorityMetrics)
	for priority, metrics := range m.metrics {
		result[priority] = &PriorityMetrics{
			PublishedEvents: metrics.PublishedEvents,
			ProcessedEvents: metrics.ProcessedEvents,
			FailedEvents:    metrics.FailedEvents,
			AverageLatency:  metrics.AverageLatency,
			LastEventTime:   metrics.LastEventTime,
		}
	}

	return result
}

// GetPriorityRatios returns the ratio of events per priority
func (m *priorityStreamManager) GetPriorityRatios() map[cqrs.EventPriority]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ratios := make(map[cqrs.EventPriority]float64)

	// Calculate total events
	var totalEvents int64
	for _, metrics := range m.metrics {
		totalEvents += metrics.PublishedEvents
	}

	// Calculate ratios
	if totalEvents > 0 {
		for priority, metrics := range m.metrics {
			ratios[priority] = float64(metrics.PublishedEvents) / float64(totalEvents)
		}
	} else {
		// No events, all ratios are 0
		for priority := range m.metrics {
			ratios[priority] = 0.0
		}
	}

	return ratios
}

// IsPriorityEnabled returns whether priority streams are enabled
func (m *priorityStreamManager) IsPriorityEnabled() bool {
	return m.config.Stream.EnablePriorityStreams
}

// Helper function to sort priorities (highest first)
func sortPrioritiesByImportance(priorities []cqrs.EventPriority) {
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i] > priorities[j] // Higher priority values come first
	})
}

// PriorityAwareStreamSelection provides logic for selecting streams based on priority
type PriorityAwareStreamSelection struct {
	manager PriorityStreamManager
}

// NewPriorityAwareStreamSelection creates a new priority-aware stream selector
func NewPriorityAwareStreamSelection(manager PriorityStreamManager) *PriorityAwareStreamSelection {
	return &PriorityAwareStreamSelection{
		manager: manager,
	}
}

// SelectOptimalStream selects the best stream for consuming based on current load
func (s *PriorityAwareStreamSelection) SelectOptimalStream(category cqrs.EventCategory, partitionKey string, currentLoad map[cqrs.EventPriority]int) *StreamInfo {
	streams := s.manager.GetStreamsByPriority(category, partitionKey)

	if len(streams) == 0 {
		return nil
	}

	// If priority is disabled, return the single stream
	if !s.manager.IsPriorityEnabled() {
		return streams[0]
	}

	// Select stream with highest priority that has manageable load
	for _, stream := range streams {
		load, exists := currentLoad[stream.Priority]
		if !exists {
			load = 0
		}

		// Simple load balancing: prefer high priority with low load
		// In a real implementation, this could be more sophisticated
		if load < 100 { // Arbitrary threshold
			return stream
		}
	}

	// If all streams are loaded, still prefer highest priority
	return streams[0]
}
