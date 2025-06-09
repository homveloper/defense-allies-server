package cqrsx

import (
	"context"
	"cqrs"
	"crypto/sha256"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoSnapshotStore implements AdvancedSnapshotStore interface using MongoDB
// Uses standard Event Sourcing snapshot schema that developers don't need to design
type MongoSnapshotStore struct {
	client         *MongoClientManager
	collectionName string
	serializer     SnapshotSerializer
}

// MongoSnapshotDocument represents the enhanced Event Sourcing snapshot schema in MongoDB
// This is a pre-designed schema that developers don't need to worry about
type MongoSnapshotDocument struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty"`
	string        cqrs.string            `bson:"aggregate_id"`   // Aggregate identifier
	AggregateType string                 `bson:"aggregate_type"` // Type of aggregate
	SnapshotData  bson.Raw               `bson:"snapshot_data"`  // Serialized aggregate state
	Version       int                    `bson:"version"`        // Version at which snapshot was taken
	Timestamp     time.Time              `bson:"timestamp"`      // When snapshot was created
	Size          int64                  `bson:"size"`           // Size of snapshot data in bytes
	ContentType   string                 `bson:"content_type"`   // Content type (application/json, application/bson)
	Compression   string                 `bson:"compression"`    // Compression type (none, gzip)
	Checksum      string                 `bson:"checksum"`       // Data integrity checksum
	Metadata      map[string]interface{} `bson:"metadata"`       // Additional metadata
}

// SnapshotSerializer interface for snapshot serialization
// This is what developers need to implement for their specific aggregates
type SnapshotSerializer interface {
	SerializeSnapshot(aggregate cqrs.AggregateRoot) ([]byte, error)
	DeserializeSnapshot(data []byte, aggregateType string) (cqrs.AggregateRoot, error)
}

// MongoSnapshotData implements SnapshotData interface for MongoDB documents
type MongoSnapshotData struct {
	doc *MongoSnapshotDocument
}

func (s *MongoSnapshotData) ID() cqrs.string {
	return s.doc.string
}

func (s *MongoSnapshotData) Type() string {
	return s.doc.AggregateType
}

func (s *MongoSnapshotData) Version() int {
	return s.doc.Version
}

func (s *MongoSnapshotData) Data() []byte {
	return []byte(s.doc.SnapshotData)
}

func (s *MongoSnapshotData) Timestamp() time.Time {
	return s.doc.Timestamp
}

func (s *MongoSnapshotData) Metadata() map[string]interface{} {
	return s.doc.Metadata
}

func (s *MongoSnapshotData) Size() int64 {
	return s.doc.Size
}

func (s *MongoSnapshotData) ContentType() string {
	return s.doc.ContentType
}

func (s *MongoSnapshotData) Compression() string {
	return s.doc.Compression
}

// NewMongoSnapshotStore creates a new MongoDB snapshot store with standard schema
func NewMongoSnapshotStore(client *MongoClientManager, collectionName string) *MongoSnapshotStore {
	if collectionName == "" {
		collectionName = "snapshots" // Standard collection name
	}

	return &MongoSnapshotStore{
		client:         client,
		collectionName: collectionName,
		serializer:     NewJSONSnapshotSerializer(false),
	}
}

// NewMongoSnapshotStoreWithSerializer creates a new MongoDB snapshot store with custom serializer
func NewMongoSnapshotStoreWithSerializer(client *MongoClientManager, collectionName string, serializer SnapshotSerializer) *MongoSnapshotStore {
	if collectionName == "" {
		collectionName = "snapshots"
	}

	return &MongoSnapshotStore{
		client:         client,
		collectionName: collectionName,
		serializer:     serializer,
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

		// Calculate metadata
		metadata := map[string]interface{}{
			"created_by": "MongoSnapshotStore",
		}

		// Add serializer metadata if available
		if advancedSerializer, ok := ss.serializer.(AdvancedSnapshotSerializer); ok {
			metadata["content_type"] = advancedSerializer.GetContentType()
			metadata["compression"] = advancedSerializer.GetCompressionType()
		}

		// Create enhanced snapshot document
		doc := MongoSnapshotDocument{
			string:        aggregate.ID(),
			AggregateType: aggregate.Type(),
			SnapshotData:  bson.Raw(snapshotData),
			Version:       aggregate.Version(),
			Timestamp:     time.Now(),
			Size:          int64(len(snapshotData)),
			ContentType:   getContentType(ss.serializer),
			Compression:   getCompressionType(ss.serializer),
			Checksum:      calculateChecksum(snapshotData),
			Metadata:      metadata,
		}

		// Upsert snapshot (replace existing snapshot for this aggregate)
		filter := bson.M{
			"aggregate_id":   aggregate.ID(),
			"aggregate_type": aggregate.Type(),
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

// Helper functions for enhanced snapshot features

// getContentType extracts content type from serializer
func getContentType(serializer SnapshotSerializer) string {
	if advancedSerializer, ok := serializer.(AdvancedSnapshotSerializer); ok {
		return advancedSerializer.GetContentType()
	}
	return "application/json" // default
}

// getCompressionType extracts compression type from serializer
func getCompressionType(serializer SnapshotSerializer) string {
	if advancedSerializer, ok := serializer.(AdvancedSnapshotSerializer); ok {
		return advancedSerializer.GetCompressionType()
	}
	return "none" // default
}

// calculateChecksum calculates SHA256 checksum for data integrity
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// AdvancedSnapshotStore interface implementation

// GetSnapshot gets the latest snapshot up to maxVersion
func (ss *MongoSnapshotStore) GetSnapshot(ctx context.Context, aggregateID string, maxVersion int) (SnapshotData, error) {
	if aggregateID == "" {
		return nil, &SnapshotError{
			Code:      ErrCodeSnapshotNotFound,
			Message:   "aggregate ID cannot be empty",
			Operation: "GetSnapshot",
		}
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_id": aggregateID,
		"version":      bson.M{"$lte": maxVersion},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "version", Value: -1}})

	var doc MongoSnapshotDocument
	err := collection.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &SnapshotError{
				Code:      ErrCodeSnapshotNotFound,
				Message:   fmt.Sprintf("snapshot not found for aggregate %s with version <= %d", aggregateID, maxVersion),
				Operation: "GetSnapshot",
			}
		}
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to get snapshot",
			Operation: "GetSnapshot",
			Cause:     err,
		}
	}

	return &MongoSnapshotData{doc: &doc}, nil
}

// GetSnapshotByVersion gets a snapshot by specific version
func (ss *MongoSnapshotStore) GetSnapshotByVersion(ctx context.Context, aggregateID string, version int) (SnapshotData, error) {
	if aggregateID == "" {
		return nil, &SnapshotError{
			Code:      ErrCodeSnapshotNotFound,
			Message:   "aggregate ID cannot be empty",
			Operation: "GetSnapshotByVersion",
		}
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{
		"aggregate_id": aggregateID,
		"version":      version,
	}

	var doc MongoSnapshotDocument
	err := collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, &SnapshotError{
				Code:      ErrCodeSnapshotNotFound,
				Message:   fmt.Sprintf("snapshot not found for aggregate %s version %d", aggregateID, version),
				Operation: "GetSnapshotByVersion",
			}
		}
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to get snapshot by version",
			Operation: "GetSnapshotByVersion",
			Cause:     err,
		}
	}

	return &MongoSnapshotData{doc: &doc}, nil
}

// DeleteOldSnapshots deletes old snapshots, keeping only the latest keepCount
func (ss *MongoSnapshotStore) DeleteOldSnapshots(ctx context.Context, aggregateID string, keepCount int) error {
	if aggregateID == "" {
		return &SnapshotError{
			Code:      ErrCodeSnapshotNotFound,
			Message:   "aggregate ID cannot be empty",
			Operation: "DeleteOldSnapshots",
		}
	}

	if keepCount <= 0 {
		return nil // nothing to delete
	}

	collection := ss.client.GetCollection(ss.collectionName)

	// Find all snapshots for the aggregate, sorted by version descending
	filter := bson.M{"aggregate_id": aggregateID}
	opts := options.Find().SetSort(bson.D{{Key: "version", Value: -1}}).SetSkip(int64(keepCount))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to find old snapshots",
			Operation: "DeleteOldSnapshots",
			Cause:     err,
		}
	}
	defer cursor.Close(ctx)

	var versionsToDelete []int
	for cursor.Next(ctx) {
		var doc MongoSnapshotDocument
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		versionsToDelete = append(versionsToDelete, doc.Version)
	}

	if len(versionsToDelete) == 0 {
		return nil // no old snapshots to delete
	}

	// Delete old snapshots
	deleteFilter := bson.M{
		"aggregate_id": aggregateID,
		"version":      bson.M{"$in": versionsToDelete},
	}

	_, err = collection.DeleteMany(ctx, deleteFilter)
	if err != nil {
		return &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to delete old snapshots",
			Operation: "DeleteOldSnapshots",
			Cause:     err,
		}
	}

	return nil
}

// ListSnapshotsForAggregate lists all snapshots for an aggregate (AdvancedSnapshotStore interface)
func (ss *MongoSnapshotStore) ListSnapshotsForAggregate(ctx context.Context, aggregateID string) ([]SnapshotData, error) {
	if aggregateID == "" {
		return nil, &SnapshotError{
			Code:      ErrCodeSnapshotNotFound,
			Message:   "aggregate ID cannot be empty",
			Operation: "ListSnapshots",
		}
	}

	collection := ss.client.GetCollection(ss.collectionName)

	filter := bson.M{"aggregate_id": aggregateID}
	opts := options.Find().SetSort(bson.D{{Key: "version", Value: -1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to list snapshots",
			Operation: "ListSnapshots",
			Cause:     err,
		}
	}
	defer cursor.Close(ctx)

	var snapshots []SnapshotData
	for cursor.Next(ctx) {
		var doc MongoSnapshotDocument
		if err := cursor.Decode(&doc); err != nil {
			continue // Skip failed decodes
		}
		snapshots = append(snapshots, &MongoSnapshotData{doc: &doc})
	}

	return snapshots, cursor.Err()
}

// GetSnapshotStats returns statistics about snapshots
func (ss *MongoSnapshotStore) GetSnapshotStats(ctx context.Context) (map[string]interface{}, error) {
	collection := ss.client.GetCollection(ss.collectionName)

	// Count total snapshots
	totalCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to count snapshots",
			Operation: "GetSnapshotStats",
			Cause:     err,
		}
	}

	// Aggregate statistics
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":              "$aggregate_type",
				"count":            bson.M{"$sum": 1},
				"total_size":       bson.M{"$sum": "$size"},
				"avg_size":         bson.M{"$avg": "$size"},
				"latest_timestamp": bson.M{"$max": "$timestamp"},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to aggregate snapshot stats",
			Operation: "GetSnapshotStats",
			Cause:     err,
		}
	}
	defer cursor.Close(ctx)

	var typeStats []map[string]interface{}
	if err := cursor.All(ctx, &typeStats); err != nil {
		return nil, &SnapshotError{
			Code:      ErrCodeStorageFailed,
			Message:   "failed to decode snapshot stats",
			Operation: "GetSnapshotStats",
			Cause:     err,
		}
	}

	stats := map[string]interface{}{
		"total_snapshots": totalCount,
		"by_type":         typeStats,
		"generated_at":    time.Now(),
	}

	return stats, nil
}
