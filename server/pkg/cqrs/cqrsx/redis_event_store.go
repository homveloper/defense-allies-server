package cqrsx

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisEventStore implements event storage using Redis
type RedisEventStore struct {
	client     *RedisClientManager
	keyBuilder *RedisKeyBuilder
	serializer EventSerializer
}

// Note: EventSerializer interface and implementations are now in event_serializer.go

// NewRedisEventStore creates a new Redis event store
func NewRedisEventStore(client *RedisClientManager, keyPrefix string) *RedisEventStore {
	return &RedisEventStore{
		client:     client,
		keyBuilder: NewRedisKeyBuilder(keyPrefix),
		serializer: &JSONEventSerializer{},
	}
}

// SaveEvents saves events to Redis
func (es *RedisEventStore) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	if aggregateID == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	// Get aggregate type from first event
	aggregateType := events[0].Type()
	eventKey := es.keyBuilder.EventKey(aggregateType, aggregateID)
	metadataKey := es.keyBuilder.MetadataKey(aggregateType, aggregateID)

	return es.client.ExecuteCommand(ctx, func() error {
		pipe := es.client.GetClient().Pipeline()

		// Check current version for optimistic concurrency control
		currentVersionCmd := pipe.HGet(ctx, metadataKey, "version")

		// Execute pipeline to get current version
		_, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to get current version", err)
		}

		// Verify expected version
		if currentVersionCmd.Val() != "" {
			currentVersion, err := strconv.Atoi(currentVersionCmd.Val())
			if err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "invalid current version", err)
			}
			if currentVersion != expectedVersion {
				return cqrs.NewCQRSError(cqrs.ErrCodeConcurrencyConflict.String(),
					fmt.Sprintf("expected version %d, but current version is %d", expectedVersion, currentVersion),
					cqrs.ErrConcurrencyConflict)
			}
		} else if expectedVersion != 0 {
			return cqrs.NewCQRSError(cqrs.ErrCodeConcurrencyConflict.String(),
				fmt.Sprintf("expected version %d, but aggregate does not exist", expectedVersion),
				cqrs.ErrConcurrencyConflict)
		}

		// Create new pipeline for saving events
		pipe = es.client.GetClient().Pipeline()

		// Serialize and save each event
		for _, event := range events {
			eventData, err := es.serializer.Serialize(event)
			if err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeSerializationError.String(), "failed to serialize event", err)
			}

			// Add event to list
			pipe.RPush(ctx, eventKey, eventData)
		}

		// Update metadata
		lastEvent := events[len(events)-1]
		pipe.HMSet(ctx, metadataKey, map[string]interface{}{
			"version":      lastEvent.Version(),
			"last_updated": time.Now().Unix(),
			"event_count":  len(events),
		})

		// Set expiration for metadata (optional)
		pipe.Expire(ctx, metadataKey, 24*time.Hour)

		// Execute pipeline
		_, err = pipe.Exec(ctx)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to save events", err)
		}

		return nil
	})
}

// GetEventHistory retrieves event history for an aggregate
func (es *RedisEventStore) GetEventHistory(ctx context.Context, aggregateID string, aggregateType string, fromVersion int) ([]cqrs.EventMessage, error) {
	if aggregateID == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	eventKey := es.keyBuilder.EventKey(aggregateType, aggregateID)

	var events []cqrs.EventMessage

	err := es.client.ExecuteCommand(ctx, func() error {
		// Get all events from Redis list
		eventData, err := es.client.GetClient().LRange(ctx, eventKey, 0, -1).Result()
		if err != nil {
			if err == redis.Nil {
				return nil // No events found
			}
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to get events", err)
		}

		// Deserialize events
		for _, data := range eventData {
			event, err := es.serializer.Deserialize([]byte(data))
			if err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeSerializationError.String(), "failed to deserialize event", err)
			}

			// Filter by version if specified
			if fromVersion > 0 && event.Version() < fromVersion {
				continue
			}

			events = append(events, event)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetLastEventVersion gets the last event version for an aggregate
func (es *RedisEventStore) GetLastEventVersion(ctx context.Context, aggregateID string, aggregateType string) (int, error) {
	if aggregateID == "" {
		return 0, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return 0, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	metadataKey := es.keyBuilder.MetadataKey(aggregateType, aggregateID)

	var version int

	err := es.client.ExecuteCommand(ctx, func() error {
		versionStr, err := es.client.GetClient().HGet(ctx, metadataKey, "version").Result()
		if err != nil {
			if err == redis.Nil {
				version = 0
				return nil
			}
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to get version", err)
		}

		version, err = strconv.Atoi(versionStr)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "invalid version format", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return version, nil
}

// CompactEvents removes old events before a specific version
func (es *RedisEventStore) CompactEvents(ctx context.Context, aggregateID string, aggregateType string, beforeVersion int) error {
	if aggregateID == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}
	if aggregateType == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	eventKey := es.keyBuilder.EventKey(aggregateType, aggregateID)

	return es.client.ExecuteCommand(ctx, func() error {
		// Get all events
		eventData, err := es.client.GetClient().LRange(ctx, eventKey, 0, -1).Result()
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to get events for compaction", err)
		}

		// Find events to keep
		var eventsToKeep []string
		for _, data := range eventData {
			event, err := es.serializer.Deserialize([]byte(data))
			if err != nil {
				continue // Skip invalid events
			}

			if event.Version() >= beforeVersion {
				eventsToKeep = append(eventsToKeep, data)
			}
		}

		// Replace the list with compacted events
		pipe := es.client.GetClient().Pipeline()
		pipe.Del(ctx, eventKey)
		if len(eventsToKeep) > 0 {
			// Convert []string to []interface{}
			args := make([]interface{}, len(eventsToKeep))
			for i, event := range eventsToKeep {
				args[i] = event
			}
			pipe.RPush(ctx, eventKey, args...)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "failed to compact events", err)
		}

		return nil
	})
}

// Note: JSONEventSerializer implementation is now in event_serializer.go
