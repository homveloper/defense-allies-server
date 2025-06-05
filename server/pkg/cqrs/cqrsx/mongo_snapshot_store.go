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

// MongoSnapshotStore implements SnapshotStore interface using MongoDB
// Uses standard Event Sourcing snapshot schema that developers don't need to design
type MongoSnapshotStore struct {
	client         *MongoClientManager
	collectionName string
	serializer     SnapshotSerializer
}

// MongoSnapshotDocument represents the standard Event Sourcing snapshot schema in MongoDB
// This is a pre-designed schema that developers don't need to worry about
type MongoSnapshotDocument struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	AggregateID   string             `bson:"aggregate_id"`   // Aggregate identifier
	AggregateType string             `bson:"aggregate_type"` // Type of aggregate
	SnapshotData  bson.Raw           `bson:"snapshot_data"`  // Serialized aggregate state
	Version       int                `bson:"version"`        // Version at which snapshot was taken
	Timestamp     time.Time          `bson:"timestamp"`      // When snapshot was created
}

// SnapshotSerializer interface for snapshot serialization
// This is what developers need to implement for their specific aggregates
type SnapshotSerializer interface {
	SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error)
	DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error)
}

// JSONSnapshotSerializer provides a default JSON-based snapshot serializer
// Developers can use this or implement their own for custom serialization
type JSONSnapshotSerializer struct{}

// SerializeSnapshot serializes an aggregate to JSON bytes
func (s *JSONSnapshotSerializer) SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error) {
	return cqrs.SerializeToJSON(aggregate)
}

// DeserializeSnapshot deserializes JSON bytes to an aggregate
func (s *JSONSnapshotSerializer) DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error) {
	// Create aggregate instance
	aggregate, err := cqrs.CreateAggregateInstance(aggregateType, "")
	if err != nil {
		return nil, err
	}

	// Deserialize data into aggregate
	err = cqrs.DeserializeFromJSONInto(data, aggregate)
	if err != nil {
		return nil, err
	}

	return aggregate, nil
}

// NewMongoSnapshotStore creates a new MongoDB snapshot store with standard schema
func NewMongoSnapshotStore(client *MongoClientManager, collectionName string) *MongoSnapshotStore {
	if collectionName == "" {
		collectionName = "snapshots" // Standard collection name
	}

	return &MongoSnapshotStore{
		client:         client,
		collectionName: collectionName,
		serializer:     &JSONSnapshotSerializer{},
	}
}

// SetSerializer allows developers to set custom snapshot serialization logic
func (ss *MongoSnapshotStore) SetSerializer(serializer SnapshotSerializer) {
	ss.serializer = serializer
}

// SaveSnapshot saves an aggregate snapshot using standard Event Sourcing pattern
func (ss *MongoSnapshotStore) SaveSnapshot(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	if aggregate == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate cannot be nil", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	return ss.client.ExecuteCommand(ctx, func() error {
		// Serialize aggregate using developer-provided or default serializer
		snapshotData, err := ss.serializer.SerializeSnapshot(aggregate)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
				fmt.Sprintf("failed to serialize snapshot: %v", err), err)
		}

		// Create standard snapshot document
		doc := MongoSnapshotDocument{
			AggregateID:   aggregate.AggregateID(),
			AggregateType: aggregate.AggregateType(),
			SnapshotData:  bson.Raw(snapshotData),
			Version:       aggregate.CurrentVersion(),
			Timestamp:     time.Now(),
		}

		// Upsert snapshot (replace existing snapshot for this aggregate)
		filter := bson.M{
			"aggregate_id":   aggregate.AggregateID(),
			"aggregate_type": aggregate.AggregateType(),
		}

		update := bson.M{"$set": doc}
		opts := options.Update().SetUpsert(true)

		_, err = collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
				fmt.Sprintf("failed to save snapshot: %v", err), err)
		}

		return nil
	})
}

// LoadSnapshot loads an aggregate snapshot using standard Event Sourcing pattern
func (ss *MongoSnapshotStore) LoadSnapshot(ctx context.Context, aggregateID, aggregateType string) (cqrs.AggregateRoot, error) {
	if aggregateID == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)
	var aggregate cqrs.AggregateRoot

	err := ss.client.ExecuteCommand(ctx, func() error {
		// Standard Event Sourcing snapshot query
		filter := bson.M{
			"aggregate_id":   aggregateID,
			"aggregate_type": aggregateType,
		}

		var doc MongoSnapshotDocument
		err := collection.FindOne(ctx, filter).Decode(&doc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotNotFound.String(),
					fmt.Sprintf("snapshot not found: %s/%s", aggregateType, aggregateID), nil)
			}
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
				fmt.Sprintf("failed to find snapshot: %v", err), err)
		}

		// Deserialize snapshot using developer-provided or default serializer
		aggregate, err = ss.serializer.DeserializeSnapshot([]byte(doc.SnapshotData), aggregateType)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
				fmt.Sprintf("failed to deserialize snapshot: %v", err), err)
		}

		return nil
	})

	return aggregate, err
}

// GetSnapshotVersion gets the version of the latest snapshot
func (ss *MongoSnapshotStore) GetSnapshotVersion(ctx context.Context, aggregateID, aggregateType string) (int, error) {
	if aggregateID == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_id":   aggregateID,
		"aggregate_type": aggregateType,
	}

	// Only select the version field
	opts := options.FindOne().SetProjection(bson.M{"version": 1})

	var doc struct {
		Version int `bson:"version"`
	}

	err := collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return -1, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotNotFound.String(),
				fmt.Sprintf("snapshot not found: %s/%s", aggregateType, aggregateID), nil)
		}
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
			fmt.Sprintf("failed to get snapshot version: %v", err), err)
	}

	return doc.Version, nil
}

// DeleteSnapshot deletes a snapshot
func (ss *MongoSnapshotStore) DeleteSnapshot(ctx context.Context, aggregateID, aggregateType string) error {
	if aggregateID == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	return ss.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{
			"aggregate_id":   aggregateID,
			"aggregate_type": aggregateType,
		}

		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
				fmt.Sprintf("failed to delete snapshot: %v", err), err)
		}

		if result.DeletedCount == 0 {
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotNotFound.String(),
				fmt.Sprintf("snapshot not found: %s/%s", aggregateType, aggregateID), nil)
		}

		return nil
	})
}

// SnapshotExists checks if a snapshot exists for the aggregate
func (ss *MongoSnapshotStore) SnapshotExists(ctx context.Context, aggregateID, aggregateType string) bool {
	if aggregateID == "" || aggregateType == "" {
		return false
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_id":   aggregateID,
		"aggregate_type": aggregateType,
	}

	count, err := collection.CountDocuments(ctx, filter)
	return err == nil && count > 0
}

// ListSnapshots lists snapshots by aggregate type with pagination
func (ss *MongoSnapshotStore) ListSnapshots(ctx context.Context, aggregateType string, limit, offset int) ([]cqrs.AggregateRoot, error) {
	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)
	var aggregates []cqrs.AggregateRoot

	err := ss.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{"aggregate_type": aggregateType}

		opts := options.Find().
			SetSort(bson.D{{Key: "timestamp", Value: -1}}).
			SetLimit(int64(limit)).
			SetSkip(int64(offset))

		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeSnapshotStoreError.String(),
				fmt.Sprintf("failed to find snapshots: %v", err), err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc MongoSnapshotDocument
			if err := cursor.Decode(&doc); err != nil {
				continue // Skip failed decodes
			}

			// Deserialize snapshot
			aggregate, err := ss.serializer.DeserializeSnapshot([]byte(doc.SnapshotData), aggregateType)
			if err != nil {
				continue // Skip failed deserializations
			}

			aggregates = append(aggregates, aggregate)
		}

		return cursor.Err()
	})

	return aggregates, err
}
