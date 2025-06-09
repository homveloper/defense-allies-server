package cqrsx

import (
	"context"
	"cqrs"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

// MongoClientManager manages MongoDB connections for CQRS infrastructure
type MongoClientManager struct {
	client           *mongo.Client
	database         *mongo.Database
	config           *MongoConfig
	metrics          *MongoMetrics
	collectionPrefix string           // Prefix for all collection names
	collectionNames  *CollectionNames // Configurable collection names
}

// CollectionNames holds configurable collection names for CQRS infrastructure
type CollectionNames struct {
	Events     string `json:"events"`      // Events collection name
	Snapshots  string `json:"snapshots"`   // Snapshots collection name
	ReadModels string `json:"read_models"` // Read models collection name
}

// MongoMetrics represents MongoDB performance metrics
type MongoMetrics struct {
	ConnectionCount int64
	CommandCount    int64
	ErrorCount      int64
	AverageLatency  time.Duration
	LastCommandTime time.Time
	DatabaseName    string
}

// NewMongoClientManager creates a new MongoDB client manager with default collection names
func NewMongoClientManager(config *MongoConfig) (*MongoClientManager, error) {
	return NewMongoClientManagerWithPrefix(config, "")
}

// NewMongoClientManagerWithPrefix creates a new MongoDB client manager with collection prefix
func NewMongoClientManagerWithPrefix(config *MongoConfig, prefix string) (*MongoClientManager, error) {
	return NewMongoClientManagerWithCollections(config, prefix, nil)
}

// NewMongoClientManagerWithCollections creates a new MongoDB client manager with custom collection names
func NewMongoClientManagerWithCollections(config *MongoConfig, prefix string, collectionNames *CollectionNames) (*MongoClientManager, error) {
	if config == nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB config cannot be nil", nil)
	}

	if err := validateMongoConfig(config); err != nil {
		return nil, err
	}

	// Apply defaults
	applyMongoConfigDefaults(config)

	// Set default collection names if not provided
	if collectionNames == nil {
		collectionNames = getDefaultCollectionNames(prefix)
	} else {
		// Apply prefix to provided collection names if they don't already have it
		if prefix != "" {
			collectionNames = applyPrefixToCollectionNames(collectionNames, prefix)
		}
	}

	// Create MongoDB client options
	clientOptions := options.Client().ApplyURI(config.URI)

	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
			Password: config.Password,
		})
	}

	if config.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(config.MaxPoolSize))
	}

	if config.ConnectTimeout > 0 {
		clientOptions.SetConnectTimeout(config.ConnectTimeout)
	}

	if config.SocketTimeout > 0 {
		clientOptions.SetSocketTimeout(config.SocketTimeout)
	}

	if config.ServerSelectionTimeout > 0 {
		clientOptions.SetServerSelectionTimeout(config.ServerSelectionTimeout)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("failed to connect to MongoDB: %v", err), err)
	}

	// Test the connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		client.Disconnect(ctx)
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("failed to ping MongoDB: %v", err), err)
	}

	// Get database
	database := client.Database(config.Database)

	return &MongoClientManager{
		client:           client,
		database:         database,
		config:           config,
		metrics:          &MongoMetrics{DatabaseName: config.Database},
		collectionPrefix: prefix,
		collectionNames:  collectionNames,
	}, nil
}

// GetClient returns the MongoDB client
func (mm *MongoClientManager) GetClient() *mongo.Client {
	return mm.client
}

// GetDatabase returns the MongoDB database
func (mm *MongoClientManager) GetDatabase() *mongo.Database {
	return mm.database
}

// GetCollection returns a MongoDB collection
func (mm *MongoClientManager) GetCollection(name string) *mongo.Collection {
	return mm.database.Collection(name)
}

// Close closes the MongoDB connection
func (mm *MongoClientManager) Close(ctx context.Context) error {
	if mm.client != nil {
		return mm.client.Disconnect(ctx)
	}
	return nil
}

// GetMetrics returns current MongoDB metrics
func (mm *MongoClientManager) GetMetrics() *MongoMetrics {
	// Return a copy of metrics
	return &MongoMetrics{
		ConnectionCount: mm.metrics.ConnectionCount,
		CommandCount:    mm.metrics.CommandCount,
		ErrorCount:      mm.metrics.ErrorCount,
		AverageLatency:  mm.metrics.AverageLatency,
		LastCommandTime: mm.metrics.LastCommandTime,
		DatabaseName:    mm.metrics.DatabaseName,
	}
}

// ExecuteCommand executes a MongoDB command with metrics tracking
func (mm *MongoClientManager) ExecuteCommand(ctx context.Context, cmd func() error) error {
	start := time.Now()

	err := cmd()

	mm.updateMetrics(time.Since(start), err)

	return err
}

// updateMetrics updates performance metrics
func (mm *MongoClientManager) updateMetrics(duration time.Duration, err error) {
	mm.metrics.CommandCount++
	mm.metrics.LastCommandTime = time.Now()

	if err != nil {
		mm.metrics.ErrorCount++
	}

	// Update average latency (simple moving average)
	if mm.metrics.CommandCount == 1 {
		mm.metrics.AverageLatency = duration
	} else {
		mm.metrics.AverageLatency = time.Duration(
			(int64(mm.metrics.AverageLatency) + int64(duration)) / 2,
		)
	}
}

// validateMongoConfig validates MongoDB configuration
func validateMongoConfig(config *MongoConfig) error {
	if config.URI == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB URI is required", nil)
	}

	// Parse and validate URI
	if _, err := connstring.ParseAndValidate(config.URI); err != nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("invalid MongoDB URI: %v", err), err)
	}

	if config.Database == "" {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB database name is required", nil)
	}

	if config.MaxPoolSize < 0 {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB max pool size cannot be negative", nil)
	}

	if config.ConnectTimeout < 0 {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB connect timeout cannot be negative", nil)
	}

	if config.SocketTimeout < 0 {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB socket timeout cannot be negative", nil)
	}

	if config.ServerSelectionTimeout < 0 {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(), "MongoDB server selection timeout cannot be negative", nil)
	}

	return nil
}

// applyMongoConfigDefaults applies default values to MongoDB configuration
func applyMongoConfigDefaults(config *MongoConfig) {
	if config.MaxPoolSize == 0 {
		config.MaxPoolSize = 100
	}

	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}

	if config.SocketTimeout == 0 {
		config.SocketTimeout = 30 * time.Second
	}

	if config.ServerSelectionTimeout == 0 {
		config.ServerSelectionTimeout = 30 * time.Second
	}
}

// InitializeEventSourcingSchema creates the standard Event Sourcing collections and indexes
func (mm *MongoClientManager) InitializeEventSourcingSchema(ctx context.Context) error {
	// Create events collection with schema validation
	if err := mm.createEventsCollection(ctx); err != nil {
		return err
	}

	// Create snapshots collection with schema validation
	if err := mm.createSnapshotsCollection(ctx); err != nil {
		return err
	}

	// Create read models collection
	if err := mm.createReadModelsCollection(ctx); err != nil {
		return err
	}

	return nil
}

// createEventsCollection creates the events collection with proper schema and indexes
func (mm *MongoClientManager) createEventsCollection(ctx context.Context) error {
	// Events collection schema is handled by the application layer
	// Just create indexes for performance
	collection := mm.GetCollection(mm.collectionNames.Events)

	// Create indexes for events collection
	// This is the standard Event Sourcing schema that developers don't need to worry about
	return mm.createEventsIndexes(ctx, collection)
}

// createSnapshotsCollection creates the snapshots collection
func (mm *MongoClientManager) createSnapshotsCollection(ctx context.Context) error {
	collection := mm.GetCollection(mm.collectionNames.Snapshots)
	return mm.createSnapshotsIndexes(ctx, collection)
}

// createReadModelsCollection creates the read models collection
func (mm *MongoClientManager) createReadModelsCollection(ctx context.Context) error {
	collection := mm.GetCollection(mm.collectionNames.ReadModels)
	return mm.createReadModelsIndexes(ctx, collection)
}

// createEventsIndexes creates standard Event Sourcing indexes for events collection
func (mm *MongoClientManager) createEventsIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "aggregate_id", Value: 1},
				{Key: "event_version", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_aggregate_version"),
		},
		{
			Keys: bson.D{
				{Key: "aggregate_id", Value: 1},
				{Key: "timestamp", Value: 1},
			},
			Options: options.Index().SetName("idx_aggregate_timestamp"),
		},
		{
			Keys: bson.D{
				{Key: "aggregate_type", Value: 1},
				{Key: "timestamp", Value: 1},
			},
			Options: options.Index().SetName("idx_type_timestamp"),
		},
		{
			Keys: bson.D{
				{Key: "event_type", Value: 1},
			},
			Options: options.Index().SetName("idx_event_type"),
		},
		{
			Keys: bson.D{
				{Key: "event_id", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_event_id"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("failed to create events indexes: %v", err), err)
	}

	return nil
}

// createSnapshotsIndexes creates indexes for snapshots collection
func (mm *MongoClientManager) createSnapshotsIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "aggregate_id", Value: 1},
				{Key: "aggregate_type", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_aggregate_snapshot"),
		},
		{
			Keys: bson.D{
				{Key: "aggregate_type", Value: 1},
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName("idx_type_timestamp_desc"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("failed to create snapshots indexes: %v", err), err)
	}

	return nil
}

// createReadModelsIndexes creates indexes for read models collection
func (mm *MongoClientManager) createReadModelsIndexes(ctx context.Context, collection *mongo.Collection) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "model_id", Value: 1},
				{Key: "model_type", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("idx_model_id_type"),
		},
		{
			Keys: bson.D{
				{Key: "model_type", Value: 1},
				{Key: "updated_at", Value: -1},
			},
			Options: options.Index().SetName("idx_type_updated"),
		},
		{
			Keys: bson.D{
				{Key: "ttl", Value: 1},
			},
			Options: options.Index().SetExpireAfterSeconds(0).SetName("idx_ttl"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return cqrs.NewCQRSError(cqrs.ErrCodeRepositoryError.String(),
			fmt.Sprintf("failed to create read models indexes: %v", err), err)
	}

	return nil
}

// getDefaultCollectionNames returns default collection names with optional prefix
func getDefaultCollectionNames(prefix string) *CollectionNames {
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	return &CollectionNames{
		Events:     prefix + "events",
		Snapshots:  prefix + "snapshots",
		ReadModels: prefix + "read_models",
	}
}

// applyPrefixToCollectionNames applies prefix to existing collection names if they don't already have it
func applyPrefixToCollectionNames(names *CollectionNames, prefix string) *CollectionNames {
	if prefix == "" {
		return names
	}

	if !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	result := &CollectionNames{
		Events:     names.Events,
		Snapshots:  names.Snapshots,
		ReadModels: names.ReadModels,
	}

	// Only add prefix if not already present
	if !strings.HasPrefix(result.Events, prefix) {
		result.Events = prefix + result.Events
	}
	if !strings.HasPrefix(result.Snapshots, prefix) {
		result.Snapshots = prefix + result.Snapshots
	}
	if !strings.HasPrefix(result.ReadModels, prefix) {
		result.ReadModels = prefix + result.ReadModels
	}

	return result
}

// GetCollectionName returns the configured collection name for the given type
func (mm *MongoClientManager) GetCollectionName(collectionType string) string {
	switch collectionType {
	case "events":
		return mm.collectionNames.Events
	case "snapshots":
		return mm.collectionNames.Snapshots
	case "read_models":
		return mm.collectionNames.ReadModels
	default:
		// For custom collections, apply prefix if configured
		if mm.collectionPrefix != "" {
			prefix := mm.collectionPrefix
			if !strings.HasSuffix(prefix, "_") {
				prefix += "_"
			}
			return prefix + collectionType
		}
		return collectionType
	}
}

// GetCollectionNames returns a copy of the current collection names configuration
func (mm *MongoClientManager) GetCollectionNames() *CollectionNames {
	return &CollectionNames{
		Events:     mm.collectionNames.Events,
		Snapshots:  mm.collectionNames.Snapshots,
		ReadModels: mm.collectionNames.ReadModels,
	}
}
