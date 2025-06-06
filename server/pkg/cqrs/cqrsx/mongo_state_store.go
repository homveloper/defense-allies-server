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

// MongoStateStore implements state-based storage using MongoDB
// Uses standard state storage schema that developers don't need to design
type MongoStateStore struct {
	client         *MongoClientManager
	collectionName string
	serializer     AggregateSerializer
}

// MongoAggregateDocument represents the standard state storage schema in MongoDB
// This is a pre-designed schema that developers don't need to worry about
type MongoAggregateDocument struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	AggregateID   string             `bson:"aggregate_id"`   // Aggregate identifier
	AggregateType string             `bson:"aggregate_type"` // Type of aggregate
	Data          bson.Raw           `bson:"data"`           // Serialized aggregate state
	Version       int                `bson:"version"`        // Current version
	CreatedAt     time.Time          `bson:"created_at"`     // When aggregate was created
	UpdatedAt     time.Time          `bson:"updated_at"`     // When aggregate was last updated
	Deleted       bool               `bson:"deleted"`        // Soft delete flag
}

// AggregateSerializer interface for aggregate serialization
type AggregateSerializer interface {
	SerializeAggregate(aggregate cqrs.AggregateRoot) ([]byte, error)
	DeserializeAggregate(data []byte, aggregateType string) (cqrs.AggregateRoot, error)
}

// JSONAggregateSerializer implements AggregateSerializer using JSON
type JSONAggregateSerializer struct{}

// SerializeAggregate serializes an aggregate to JSON bytes
func (s *JSONAggregateSerializer) SerializeAggregate(aggregate cqrs.AggregateRoot) ([]byte, error) {
	return cqrs.SerializeToJSON(aggregate)
}

// DeserializeAggregate deserializes JSON bytes to an aggregate
func (s *JSONAggregateSerializer) DeserializeAggregate(data []byte, aggregateType string) (cqrs.AggregateRoot, error) {
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

// NewMongoStateStore creates a new MongoDB state store with standard schema
func NewMongoStateStore(client *MongoClientManager, collectionName string) *MongoStateStore {
	if collectionName == "" {
		collectionName = "aggregates" // Standard collection name
	}

	return &MongoStateStore{
		client:         client,
		collectionName: collectionName,
		serializer:     &JSONAggregateSerializer{},
	}
}

// SaveAggregate saves an aggregate to MongoDB using standard state storage pattern
func (ss *MongoStateStore) SaveAggregate(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	if aggregate == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate cannot be nil", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	return ss.client.ExecuteCommand(ctx, func() error {
		// Check version for optimistic concurrency control
		if expectedVersion >= 0 {
			currentVersion, err := ss.GetAggregateVersion(ctx, aggregate.AggregateID(), aggregate.AggregateType())
			if err != nil && !cqrs.IsNotFoundError(err) {
				return err
			}

			if currentVersion != expectedVersion {
				return cqrs.NewCQRSError(cqrs.ErrCodeConcurrencyConflict.String(),
					fmt.Sprintf("concurrency conflict: expected version %d, got %d", expectedVersion, currentVersion), nil)
			}
		}

		// Serialize aggregate
		data, err := ss.serializer.SerializeAggregate(aggregate)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("failed to serialize aggregate: %v", err), err)
		}

		// Create standard state document
		now := time.Now()
		doc := MongoAggregateDocument{
			AggregateID:   aggregate.AggregateID(),
			AggregateType: aggregate.AggregateType(),
			Data:          bson.Raw(data),
			Version:       aggregate.CurrentVersion(),
			CreatedAt:     aggregate.CreatedAt(),
			UpdatedAt:     now,
			Deleted:       aggregate.IsDeleted(),
		}

		// Upsert document using standard pattern
		filter := bson.M{
			"aggregate_id":   aggregate.AggregateID(),
			"aggregate_type": aggregate.AggregateType(),
		}

		update := bson.M{
			"$set": doc,
			"$setOnInsert": bson.M{
				"created_at": aggregate.CreatedAt(),
			},
		}

		opts := options.Update().SetUpsert(true)
		_, err = collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("failed to save aggregate: %v", err), err)
		}

		return nil
	})
}

// LoadAggregate loads an aggregate from MongoDB using standard state storage pattern
func (ss *MongoStateStore) LoadAggregate(ctx context.Context, aggregateID, aggregateType string) (cqrs.AggregateRoot, error) {
	if aggregateID == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)
	var aggregate cqrs.AggregateRoot

	err := ss.client.ExecuteCommand(ctx, func() error {
		// Standard state storage query
		filter := bson.M{
			"aggregate_id":   aggregateID,
			"aggregate_type": aggregateType,
			"deleted":        false,
		}

		var doc MongoAggregateDocument
		err := collection.FindOne(ctx, filter).Decode(&doc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return cqrs.NewCQRSError(cqrs.ErrCodeAggregateNotFound.String(),
					fmt.Sprintf("aggregate not found: %s/%s", aggregateType, aggregateID), nil)
			}
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("failed to find aggregate: %v", err), err)
		}

		// Deserialize aggregate
		aggregate, err = ss.serializer.DeserializeAggregate([]byte(doc.Data), aggregateType)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("failed to deserialize aggregate: %v", err), err)
		}

		return nil
	})

	return aggregate, err
}

// GetAggregateVersion gets the version of an aggregate
func (ss *MongoStateStore) GetAggregateVersion(ctx context.Context, aggregateID, aggregateType string) (int, error) {
	if aggregateID == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_id":   aggregateID,
		"aggregate_type": aggregateType,
	}

	// Only select the version field for efficiency
	opts := options.FindOne().SetProjection(bson.M{"version": 1})

	var doc struct {
		Version int `bson:"version"`
	}

	err := collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// EventStore와 일치하도록 존재하지 않는 Aggregate는 버전 0 반환
			return 0, nil
		}
		return -1, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
			fmt.Sprintf("failed to get aggregate version: %v", err), err)
	}

	return doc.Version, nil
}

// AggregateExists checks if an aggregate exists
func (ss *MongoStateStore) AggregateExists(ctx context.Context, aggregateID, aggregateType string) bool {
	if aggregateID == "" || aggregateType == "" {
		return false
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_id":   aggregateID,
		"aggregate_type": aggregateType,
		"deleted":        false,
	}

	count, err := collection.CountDocuments(ctx, filter)
	return err == nil && count > 0
}

// DeleteAggregate marks an aggregate as deleted (soft delete)
func (ss *MongoStateStore) DeleteAggregate(ctx context.Context, aggregateID, aggregateType string) error {
	if aggregateID == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate ID cannot be empty", nil)
	}

	if aggregateType == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	return ss.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{
			"aggregate_id":   aggregateID,
			"aggregate_type": aggregateType,
		}

		update := bson.M{
			"$set": bson.M{
				"deleted":    true,
				"updated_at": time.Now(),
			},
		}

		result, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("failed to delete aggregate: %v", err), err)
		}

		if result.MatchedCount == 0 {
			return cqrs.NewCQRSError(cqrs.ErrCodeAggregateNotFound.String(),
				fmt.Sprintf("aggregate not found: %s/%s", aggregateType, aggregateID), nil)
		}

		return nil
	})
}

// ListAggregates lists aggregates by type with pagination
func (ss *MongoStateStore) ListAggregates(ctx context.Context, aggregateType string, limit, offset int) ([]cqrs.AggregateRoot, error) {
	if aggregateType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)
	var aggregates []cqrs.AggregateRoot

	err := ss.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{
			"aggregate_type": aggregateType,
			"deleted":        false,
		}

		opts := options.Find().
			SetSort(bson.D{{Key: "updated_at", Value: -1}}).
			SetLimit(int64(limit)).
			SetSkip(int64(offset))

		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("failed to find aggregates: %v", err), err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc MongoAggregateDocument
			if err := cursor.Decode(&doc); err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
					fmt.Sprintf("failed to decode aggregate document: %v", err), err)
			}

			// Deserialize aggregate
			aggregate, err := ss.serializer.DeserializeAggregate([]byte(doc.Data), aggregateType)
			if err != nil {
				continue // Skip failed deserializations
			}

			aggregates = append(aggregates, aggregate)
		}

		if err := cursor.Err(); err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
				fmt.Sprintf("cursor error: %v", err), err)
		}

		return nil
	})

	return aggregates, err
}

// CountAggregates counts aggregates by type
func (ss *MongoStateStore) CountAggregates(ctx context.Context, aggregateType string) (int64, error) {
	if aggregateType == "" {
		return 0, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(), "aggregate type cannot be empty", nil)
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_type": aggregateType,
		"deleted":        false,
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
			fmt.Sprintf("failed to count aggregates: %v", err), err)
	}

	return count, nil
}

// CreateIndexes creates necessary indexes for the state store
func (ss *MongoStateStore) CreateIndexes(ctx context.Context) error {
	collection := ss.client.GetCollection(ss.collectionName)

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "aggregate_id", Value: 1},
				{Key: "aggregate_type", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_aggregate_id_type"),
		},
		{
			Keys: bson.D{
				{Key: "aggregate_type", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("idx_type_updated"),
		},
		{
			Keys:    bson.D{{Key: "deleted", Value: 1}},
			Options: options.Index().SetName("idx_deleted"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeStateStoreError.String(),
			fmt.Sprintf("failed to create indexes: %v", err), err)
	}

	return nil
}
