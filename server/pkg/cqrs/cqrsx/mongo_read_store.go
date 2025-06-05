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

// MongoReadStore implements ReadStore interface using MongoDB
// Uses standard CQRS read model schema that developers don't need to design
type MongoReadStore struct {
	client         *MongoClientManager
	collectionName string
	serializer     ReadModelSerializer
}

// MongoReadModelDocument represents the standard CQRS read model schema in MongoDB
// This is a pre-designed schema that developers don't need to worry about
type MongoReadModelDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ModelID   string             `bson:"model_id"`      // Read model identifier
	ModelType string             `bson:"model_type"`    // Type of read model
	Data      bson.Raw           `bson:"data"`          // Serialized read model data
	Version   int                `bson:"version"`       // Version for optimistic updates
	CreatedAt time.Time          `bson:"created_at"`    // When read model was created
	UpdatedAt time.Time          `bson:"updated_at"`    // When read model was last updated
	TTL       *time.Time         `bson:"ttl,omitempty"` // Time-to-live for automatic expiration
}

// ReadModelSerializer interface for read model serialization
type ReadModelSerializer interface {
	SerializeReadModel(model cqrs.ReadModel) ([]byte, error)
	DeserializeReadModel(data []byte, modelType string) (cqrs.ReadModel, error)
}

// JSONReadModelSerializer implements ReadModelSerializer using JSON
type JSONReadModelSerializer struct{}

// SerializeReadModel serializes a read model to JSON bytes
func (s *JSONReadModelSerializer) SerializeReadModel(model cqrs.ReadModel) ([]byte, error) {
	return cqrs.SerializeToJSON(model)
}

// DeserializeReadModel deserializes JSON bytes to a read model
func (s *JSONReadModelSerializer) DeserializeReadModel(data []byte, modelType string) (cqrs.ReadModel, error) {
	// Get model type for deserialization
	modelTypeReflect, err := cqrs.GetReadModelType(modelType)
	if err != nil {
		return nil, err
	}

	result, err := cqrs.DeserializeFromJSON(data, modelTypeReflect)
	if err != nil {
		return nil, err
	}

	// Type assertion to ReadModel
	if readModel, ok := result.(cqrs.ReadModel); ok {
		return readModel, nil
	}

	return nil, fmt.Errorf("deserialized object is not a ReadModel")
}

// NewMongoReadStore creates a new MongoDB read store with standard schema
func NewMongoReadStore(client *MongoClientManager, collectionName string) *MongoReadStore {
	if collectionName == "" {
		collectionName = "read_models" // Standard collection name
	}

	return &MongoReadStore{
		client:         client,
		collectionName: collectionName,
		serializer:     &JSONReadModelSerializer{},
	}
}

// Save saves a read model to MongoDB using standard CQRS pattern
func (rs *MongoReadStore) Save(ctx context.Context, readModel cqrs.ReadModel) error {
	if readModel == nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(), "read model cannot be nil", nil)
	}

	collection := rs.client.GetCollection(rs.collectionName)

	return rs.client.ExecuteCommand(ctx, func() error {
		// Serialize read model
		data, err := rs.serializer.SerializeReadModel(readModel)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to serialize read model: %v", err), err)
		}

		// Create standard read model document
		now := time.Now()
		doc := MongoReadModelDocument{
			ModelID:   readModel.GetID(),
			ModelType: readModel.GetType(),
			Data:      bson.Raw(data),
			Version:   readModel.GetVersion(),
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Set TTL if specified (check if readModel has TTL method)
		if ttlModel, ok := readModel.(interface{ GetTTL() time.Duration }); ok {
			if ttl := ttlModel.GetTTL(); ttl > 0 {
				expiresAt := now.Add(ttl)
				doc.TTL = &expiresAt
			}
		}

		// Upsert document using standard CQRS pattern
		filter := bson.M{
			"model_id":   readModel.GetID(),
			"model_type": readModel.GetType(),
		}

		update := bson.M{
			"$set": doc,
			"$setOnInsert": bson.M{
				"created_at": now,
			},
		}

		opts := options.Update().SetUpsert(true)
		_, err = collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to save read model: %v", err), err)
		}

		return nil
	})
}

// GetByID retrieves a read model by ID and type using standard CQRS query
func (rs *MongoReadStore) GetByID(ctx context.Context, id string, modelType string) (cqrs.ReadModel, error) {
	if id == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(), "model ID cannot be empty", nil)
	}

	if modelType == "" {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(), "model type cannot be empty", nil)
	}

	collection := rs.client.GetCollection(rs.collectionName)
	var readModel cqrs.ReadModel

	err := rs.client.ExecuteCommand(ctx, func() error {
		// Standard CQRS read model query
		filter := bson.M{
			"model_id":   id,
			"model_type": modelType,
		}

		var doc MongoReadModelDocument
		err := collection.FindOne(ctx, filter).Decode(&doc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return cqrs.NewCQRSError(cqrs.ErrCodeReadModelNotFound.String(),
					fmt.Sprintf("read model not found: %s/%s", modelType, id), nil)
			}
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to find read model: %v", err), err)
		}

		// Deserialize read model
		readModel, err = rs.serializer.DeserializeReadModel([]byte(doc.Data), modelType)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to deserialize read model: %v", err), err)
		}

		return nil
	})

	return readModel, err
}

// Delete removes a read model from MongoDB using standard CQRS pattern
func (rs *MongoReadStore) Delete(ctx context.Context, id string, modelType string) error {
	if id == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(), "model ID cannot be empty", nil)
	}

	if modelType == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(), "model type cannot be empty", nil)
	}

	collection := rs.client.GetCollection(rs.collectionName)

	return rs.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{
			"model_id":   id,
			"model_type": modelType,
		}

		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to delete read model: %v", err), err)
		}

		if result.DeletedCount == 0 {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadModelNotFound.String(),
				fmt.Sprintf("read model not found: %s/%s", modelType, id), nil)
		}

		return nil
	})
}

// Query executes a query against read models using standard CQRS pattern
func (rs *MongoReadStore) Query(ctx context.Context, criteria cqrs.QueryCriteria) ([]cqrs.ReadModel, error) {
	collection := rs.client.GetCollection(rs.collectionName)
	var readModels []cqrs.ReadModel

	err := rs.client.ExecuteCommand(ctx, func() error {
		// Build MongoDB filter from criteria
		filter := rs.buildMongoFilter(criteria)

		// Build options
		opts := options.Find()

		// Apply sorting
		if criteria.SortBy != "" {
			direction := 1
			if criteria.SortOrder == cqrs.Descending {
				direction = -1
			}
			sortDoc := bson.D{{Key: criteria.SortBy, Value: direction}}
			opts.SetSort(sortDoc)
		}

		// Apply pagination
		if criteria.Limit > 0 {
			opts.SetLimit(int64(criteria.Limit))
		}
		if criteria.Offset > 0 {
			opts.SetSkip(int64(criteria.Offset))
		}

		// Execute query
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to execute query: %v", err), err)
		}
		defer cursor.Close(ctx)

		// Decode results
		for cursor.Next(ctx) {
			var doc MongoReadModelDocument
			if err := cursor.Decode(&doc); err != nil {
				return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
					fmt.Sprintf("failed to decode read model document: %v", err), err)
			}

			// Deserialize read model
			readModel, err := rs.serializer.DeserializeReadModel([]byte(doc.Data), doc.ModelType)
			if err != nil {
				continue // Skip failed deserializations
			}

			readModels = append(readModels, readModel)
		}

		if err := cursor.Err(); err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("cursor error: %v", err), err)
		}

		return nil
	})

	return readModels, err
}

// Count counts read models matching the criteria
func (rs *MongoReadStore) Count(ctx context.Context, criteria cqrs.QueryCriteria) (int64, error) {
	collection := rs.client.GetCollection(rs.collectionName)
	var count int64

	err := rs.client.ExecuteCommand(ctx, func() error {
		filter := rs.buildMongoFilter(criteria)

		var err error
		count, err = collection.CountDocuments(ctx, filter)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to count documents: %v", err), err)
		}

		return nil
	})

	return count, err
}

// SaveBatch saves multiple read models in a single operation
func (rs *MongoReadStore) SaveBatch(ctx context.Context, readModels []cqrs.ReadModel) error {
	if len(readModels) == 0 {
		return nil
	}

	collection := rs.client.GetCollection(rs.collectionName)

	return rs.client.ExecuteCommand(ctx, func() error {
		// Prepare bulk operations
		var operations []mongo.WriteModel

		for _, readModel := range readModels {
			if readModel == nil {
				continue
			}

			// Serialize read model
			data, err := rs.serializer.SerializeReadModel(readModel)
			if err != nil {
				continue // Skip failed serializations
			}

			// Create document
			now := time.Now()
			doc := MongoReadModelDocument{
				ModelID:   readModel.GetID(),
				ModelType: readModel.GetType(),
				Data:      bson.Raw(data),
				Version:   readModel.GetVersion(),
				CreatedAt: now,
				UpdatedAt: now,
			}

			// Set TTL if specified (check if readModel has TTL method)
			if ttlModel, ok := readModel.(interface{ GetTTL() time.Duration }); ok {
				if ttl := ttlModel.GetTTL(); ttl > 0 {
					expiresAt := now.Add(ttl)
					doc.TTL = &expiresAt
				}
			}

			// Create upsert operation
			filter := bson.M{
				"model_id":   readModel.GetID(),
				"model_type": readModel.GetType(),
			}

			update := bson.M{
				"$set": doc,
				"$setOnInsert": bson.M{
					"created_at": now,
				},
			}

			operation := mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true)

			operations = append(operations, operation)
		}

		if len(operations) == 0 {
			return nil
		}

		// Execute bulk write
		_, err := collection.BulkWrite(ctx, operations)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to save read models batch: %v", err), err)
		}

		return nil
	})
}

// DeleteBatch deletes multiple read models in a single operation
func (rs *MongoReadStore) DeleteBatch(ctx context.Context, ids []string, modelType string) error {
	if len(ids) == 0 {
		return nil
	}

	if modelType == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(), "model type cannot be empty", nil)
	}

	collection := rs.client.GetCollection(rs.collectionName)

	return rs.client.ExecuteCommand(ctx, func() error {
		filter := bson.M{
			"model_id":   bson.M{"$in": ids},
			"model_type": modelType,
		}

		_, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			return cqrs.NewCQRSError(cqrs.ErrCodeReadStoreError.String(),
				fmt.Sprintf("failed to delete read models batch: %v", err), err)
		}

		return nil
	})
}

// buildMongoFilter builds MongoDB filter from query criteria
func (rs *MongoReadStore) buildMongoFilter(criteria cqrs.QueryCriteria) bson.M {
	filter := bson.M{}

	// Add field filters
	for field, value := range criteria.Filters {
		filter[field] = value
	}

	return filter
}
