package cqrsx

import (
	"context"
	"defense-allies-server/pkg/cqrs"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoEventStore implements EventStore interface using MongoDB
// Uses standard Event Sourcing schema that developers don't need to design
type MongoEventStore struct {
	client         *MongoClientManager
	collectionName string
	serializer     EventSerializer
}

// MongoEventDocument represents the standard Event Sourcing document schema in MongoDB
// This is a pre-designed schema that developers don't need to worry about
type MongoEventDocument struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty"`
	AggregateID   string                 `bson:"aggregate_id"`       // Aggregate identifier
	AggregateType string                 `bson:"aggregate_type"`     // Type of aggregate (User, Order, etc.)
	EventID       string                 `bson:"event_id"`           // Unique event identifier
	EventType     string                 `bson:"event_type"`         // Type of event (UserCreated, OrderPlaced, etc.)
	EventData     bson.Raw               `bson:"event_data"`         // Serialized event payload
	EventVersion  int                    `bson:"event_version"`      // Version for optimistic concurrency control
	Timestamp     time.Time              `bson:"timestamp"`          // When the event occurred
	Metadata      map[string]interface{} `bson:"metadata,omitempty"` // Additional metadata
}

// NewMongoEventStore creates a new MongoDB event store with standard schema
func NewMongoEventStore(client *MongoClientManager, collectionName string) *MongoEventStore {
	if collectionName == "" {
		collectionName = "events" // Standard collection name
	}

	return &MongoEventStore{
		client:         client,
		collectionName: collectionName,
		serializer:     &BSONEventSerializer{},
	}
}

// SaveEvents saves events to MongoDB using the standard Event Sourcing pattern
func (es *MongoEventStore) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	if aggregateID == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	collection := es.client.GetCollection(es.collectionName)

	return es.client.ExecuteCommand(ctx, func() error {
		// Start a session for transaction
		session, err := es.client.GetClient().StartSession()
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
				fmt.Sprintf("failed to start MongoDB session: %v", err), err)
		}
		defer session.EndSession(ctx)

		// Execute transaction
		_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			// Check current version for optimistic concurrency control
			if expectedVersion >= 0 {
				currentVersion, err := es.getLastEventVersion(sessCtx, aggregateID, events[0].AggregateType())
				if err != nil {
					return nil, err
				}

				if currentVersion != expectedVersion {
					return nil, cqrs.NewCQRSError(cqrs.ErrCodeConcurrencyConflict.String(),
						fmt.Sprintf("concurrency conflict: expected version %d, got %d", expectedVersion, currentVersion), nil)
				}
			}

			// Convert events to MongoDB documents using standard schema
			documents := make([]interface{}, len(events))
			for i, event := range events {
				// Serialize event data properly
				var eventDataBytes []byte
				var err error

				// Handle different data types properly
				eventData := event.EventData()
				if eventData != nil {
					eventDataBytes, err = bson.Marshal(eventData)
					if err != nil {
						return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
							fmt.Sprintf("failed to serialize event data: %v", err), err)
					}
				} else {
					// Handle nil event data
					eventDataBytes, _ = bson.Marshal(bson.M{})
				}

				// Standard Event Sourcing document structure - no duplication
				doc := MongoEventDocument{
					AggregateID:   aggregateID,
					AggregateType: event.AggregateType(),
					EventID:       event.EventID(),
					EventType:     event.EventType(),
					EventData:     bson.Raw(eventDataBytes), // Only store the actual event data
					EventVersion:  expectedVersion + i + 1,
					Timestamp:     event.Timestamp(),
					Metadata:      event.Metadata(),
				}

				documents[i] = doc
			}

			// Insert events atomically
			_, err = collection.InsertMany(sessCtx, documents)
			if err != nil {
				return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
					fmt.Sprintf("failed to insert events: %v", err), err)
			}

			return nil, nil
		})

		return err
	})
}

// LoadEvents loads events from MongoDB using standard Event Sourcing queries
func (es *MongoEventStore) LoadEvents(ctx context.Context, aggregateID, aggregateType string, fromVersion, toVersion int) ([]cqrs.EventMessage, error) {
	if aggregateID == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := es.client.GetCollection(es.collectionName)
	var events []cqrs.EventMessage

	err := es.client.ExecuteCommand(ctx, func() error {
		// Build filter using standard Event Sourcing query pattern
		filter := bson.M{
			"aggregate_id":   aggregateID,
			"aggregate_type": aggregateType,
		}

		if fromVersion > 0 || toVersion > 0 {
			versionFilter := bson.M{}
			if fromVersion > 0 {
				versionFilter["$gte"] = fromVersion
			}
			if toVersion > 0 {
				versionFilter["$lte"] = toVersion
			}
			filter["event_version"] = versionFilter
		}

		// Sort by event version (standard Event Sourcing ordering)
		opts := options.Find().SetSort(bson.D{{Key: "event_version", Value: 1}})

		// Find events
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
				fmt.Sprintf("failed to find events: %v", err), err)
		}
		defer cursor.Close(ctx)

		// Decode events
		for cursor.Next(ctx) {
			var doc MongoEventDocument
			if err := cursor.Decode(&doc); err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
					fmt.Sprintf("failed to decode event document: %v", err), err)
			}

			// Deserialize only the event data
			var eventData interface{}
			if err := bson.Unmarshal(doc.EventData, &eventData); err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
					fmt.Sprintf("failed to deserialize event data: %v", err), err)
			}

			// Reconstruct the event message
			event := cqrs.NewBaseEventMessage(
				doc.EventType,
				doc.AggregateID,
				doc.AggregateType,
				doc.EventVersion,
				eventData,
			)

			event.SetEventID(doc.EventID)
			event.SetTimestamp(doc.Timestamp)

			// Set metadata
			for key, value := range doc.Metadata {
				event.AddMetadata(key, value)
			}

			events = append(events, event)
		}

		if err := cursor.Err(); err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
				fmt.Sprintf("cursor error: %v", err), err)
		}

		return nil
	})

	return events, err
}

// GetEventHistory retrieves event history for an aggregate (standard Event Sourcing operation)
func (es *MongoEventStore) GetEventHistory(ctx context.Context, aggregateID, aggregateType string, fromVersion int) ([]cqrs.EventMessage, error) {
	if aggregateID == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	return es.LoadEvents(ctx, aggregateID, aggregateType, fromVersion, 0)
}

// GetLastEventVersion gets the last event version for an aggregate (standard Event Sourcing query)
func (es *MongoEventStore) GetLastEventVersion(ctx context.Context, aggregateID, aggregateType string) (int, error) {
	if aggregateID == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	return es.getLastEventVersion(ctx, aggregateID, aggregateType)
}

// getLastEventVersion internal method to get last event version
func (es *MongoEventStore) getLastEventVersion(ctx context.Context, aggregateID, aggregateType string) (int, error) {
	collection := es.client.GetCollection(es.collectionName)

	// Standard Event Sourcing query to get latest version
	filter := bson.M{
		"aggregate_id":   aggregateID,
		"aggregate_type": aggregateType,
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "event_version", Value: -1}})

	var doc MongoEventDocument
	err := collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, nil // No events found, start from version 0
		}
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
			fmt.Sprintf("failed to get last event version: %v", err), err)
	}

	return doc.EventVersion, nil
}

// CompactEvents removes old events (standard Event Sourcing maintenance operation)
func (es *MongoEventStore) CompactEvents(ctx context.Context, aggregateID, aggregateType string, beforeVersion int) error {
	if aggregateID == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := es.client.GetCollection(es.collectionName)

	return es.client.ExecuteCommand(ctx, func() error {
		// Standard Event Sourcing compaction query
		filter := bson.M{
			"aggregate_id":   aggregateID,
			"aggregate_type": aggregateType,
			"event_version":  bson.M{"$lt": beforeVersion},
		}

		_, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
				fmt.Sprintf("failed to compact events: %v", err), err)
		}

		return nil
	})
}

// GetEventsByType gets events by event type (useful for projections)
func (es *MongoEventStore) GetEventsByType(ctx context.Context, eventType string, fromTimestamp time.Time, limit int) ([]cqrs.EventMessage, error) {
	if eventType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(), "event type cannot be empty", nil)
	}

	collection := es.client.GetCollection(es.collectionName)
	var events []cqrs.EventMessage

	err := es.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{"event_type": eventType}

		if !fromTimestamp.IsZero() {
			filter["timestamp"] = bson.M{"$gte": fromTimestamp}
		}

		opts := options.Find().
			SetSort(bson.D{{Key: "timestamp", Value: 1}}).
			SetLimit(int64(limit))

		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeEventStoreError.String(),
				fmt.Sprintf("failed to find events by type: %v", err), err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc MongoEventDocument
			if err := cursor.Decode(&doc); err != nil {
				continue // Skip failed decodes
			}

			// Deserialize only the event data
			var eventData interface{}
			if err := bson.Unmarshal(doc.EventData, &eventData); err != nil {
				continue // Skip failed deserializations
			}

			// Reconstruct the event message
			event := cqrs.NewBaseEventMessage(
				doc.EventType,
				doc.AggregateID,
				doc.AggregateType,
				doc.EventVersion,
				eventData,
			)

			event.SetEventID(doc.EventID)
			event.SetTimestamp(doc.Timestamp)

			// Set metadata
			for key, value := range doc.Metadata {
				event.AddMetadata(key, value)
			}

			events = append(events, event)
		}

		return cursor.Err()
	})

	return events, err
}
