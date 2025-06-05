package cqrsx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMongoClientManager_NewMongoClientManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *MongoConfig
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "empty URI",
			config: &MongoConfig{
				URI:      "",
				Database: "test",
			},
			expectError: true,
		},
		{
			name: "empty database",
			config: &MongoConfig{
				URI:      "mongodb://localhost:27017",
				Database: "",
			},
			expectError: true,
		},
		{
			name: "invalid URI",
			config: &MongoConfig{
				URI:      "invalid-uri",
				Database: "test",
			},
			expectError: true,
		},
		{
			name: "valid config",
			config: &MongoConfig{
				URI:      "mongodb://localhost:27017",
				Database: "test",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewMongoClientManager(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)
			} else {
				// Note: This test will fail if MongoDB is not running
				// In a real test environment, you would use a test container
				if err != nil {
					t.Skipf("MongoDB not available: %v", err)
				}
				assert.NoError(t, err)
				assert.NotNil(t, client)

				if client != nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					client.Close(ctx)
				}
			}
		})
	}
}

func TestMongoClientManager_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *MongoConfig
		expectError bool
	}{
		{
			name: "negative max pool size",
			config: &MongoConfig{
				URI:         "mongodb://localhost:27017",
				Database:    "test",
				MaxPoolSize: -1,
			},
			expectError: true,
		},
		{
			name: "negative connect timeout",
			config: &MongoConfig{
				URI:            "mongodb://localhost:27017",
				Database:       "test",
				ConnectTimeout: -1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "valid config with defaults",
			config: &MongoConfig{
				URI:      "mongodb://localhost:27017",
				Database: "test",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMongoConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMongoClientManager_ApplyDefaults(t *testing.T) {
	config := &MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "test",
	}

	applyMongoConfigDefaults(config)

	assert.Equal(t, 100, config.MaxPoolSize)
	assert.Equal(t, 10*time.Second, config.ConnectTimeout)
	assert.Equal(t, 30*time.Second, config.SocketTimeout)
	assert.Equal(t, 30*time.Second, config.ServerSelectionTimeout)
}

func TestMongoClientManager_Metrics(t *testing.T) {
	config := &MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "test",
	}

	client, err := NewMongoClientManager(config)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	require.NotNil(t, client)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client.Close(ctx)
	}()

	// Test initial metrics
	metrics := client.GetMetrics()
	assert.Equal(t, int64(0), metrics.CommandCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.Equal(t, "test", metrics.DatabaseName)

	// Test command execution with metrics
	ctx := context.Background()
	err = client.ExecuteCommand(ctx, func() error {
		return nil // Successful command
	})
	assert.NoError(t, err)

	metrics = client.GetMetrics()
	assert.Equal(t, int64(1), metrics.CommandCount)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.True(t, metrics.AverageLatency > 0)

	// Test command execution with error
	err = client.ExecuteCommand(ctx, func() error {
		return assert.AnError
	})
	assert.Error(t, err)

	metrics = client.GetMetrics()
	assert.Equal(t, int64(2), metrics.CommandCount)
	assert.Equal(t, int64(1), metrics.ErrorCount)
}

func TestMongoClientManager_GetCollection(t *testing.T) {
	config := &MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "test",
	}

	client, err := NewMongoClientManager(config)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	require.NotNil(t, client)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client.Close(ctx)
	}()

	collection := client.GetCollection("test_collection")
	assert.NotNil(t, collection)
	assert.Equal(t, "test_collection", collection.Name())
	assert.Equal(t, "test", collection.Database().Name())
}

func TestMongoClientManager_InitializeEventSourcingSchema(t *testing.T) {
	config := &MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "test_event_sourcing",
	}

	client, err := NewMongoClientManager(config)
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	require.NotNil(t, client)

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Clean up test database
		client.GetDatabase().Drop(ctx)
		client.Close(ctx)
	}()

	ctx := context.Background()
	err = client.InitializeEventSourcingSchema(ctx)
	assert.NoError(t, err)

	// Verify collections exist
	collections, err := client.GetDatabase().ListCollectionNames(ctx, map[string]interface{}{})
	assert.NoError(t, err)

	expectedCollections := []string{"events", "snapshots", "read_models"}
	for _, expected := range expectedCollections {
		assert.Contains(t, collections, expected)
	}

	// Verify indexes exist
	eventsCollection := client.GetCollection("events")
	indexes := eventsCollection.Indexes()
	cursor, err := indexes.List(ctx)
	assert.NoError(t, err)
	defer cursor.Close(ctx)

	var indexCount int
	for cursor.Next(ctx) {
		indexCount++
	}

	// Should have at least the default _id index plus our custom indexes
	assert.True(t, indexCount > 1)
}

// Integration test helper
func getTestMongoConfig() *MongoConfig {
	return &MongoConfig{
		URI:                    "mongodb://localhost:27017",
		Database:               "test_cqrs",
		MaxPoolSize:            10,
		ConnectTimeout:         5 * time.Second,
		SocketTimeout:          10 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
	}
}

// Test helper to create a test MongoDB client
func createTestMongoClient(t *testing.T) *MongoClientManager {
	config := getTestMongoConfig()
	client, err := NewMongoClientManager(config)
	if err != nil {
		t.Skipf("MongoDB not available for integration tests: %v", err)
	}
	require.NotNil(t, client)

	// Initialize schema
	ctx := context.Background()
	err = client.InitializeEventSourcingSchema(ctx)
	require.NoError(t, err)

	return client
}

// Test helper to clean up test MongoDB client
func cleanupTestMongoClient(t *testing.T, client *MongoClientManager) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Drop test database
	client.GetDatabase().Drop(ctx)

	// Close connection
	client.Close(ctx)
}

func TestMongoClientManager_CollectionNames(t *testing.T) {
	config := &MongoConfig{
		URI:      "mongodb://localhost:27017",
		Database: "test_collection_names",
	}

	t.Run("default collection names", func(t *testing.T) {
		manager, err := NewMongoClientManager(config)
		if err != nil {
			t.Skipf("MongoDB not available: %v", err)
		}
		defer cleanupTestMongoClient(t, manager)

		names := manager.GetCollectionNames()
		assert.Equal(t, "events", names.Events)
		assert.Equal(t, "snapshots", names.Snapshots)
		assert.Equal(t, "read_models", names.ReadModels)

		assert.Equal(t, "events", manager.GetCollectionName("events"))
		assert.Equal(t, "snapshots", manager.GetCollectionName("snapshots"))
		assert.Equal(t, "read_models", manager.GetCollectionName("read_models"))
	})

	t.Run("with prefix", func(t *testing.T) {
		manager, err := NewMongoClientManagerWithPrefix(config, "myapp")
		if err != nil {
			t.Skipf("MongoDB not available: %v", err)
		}
		defer cleanupTestMongoClient(t, manager)

		names := manager.GetCollectionNames()
		assert.Equal(t, "myapp_events", names.Events)
		assert.Equal(t, "myapp_snapshots", names.Snapshots)
		assert.Equal(t, "myapp_read_models", names.ReadModels)

		assert.Equal(t, "myapp_events", manager.GetCollectionName("events"))
		assert.Equal(t, "myapp_snapshots", manager.GetCollectionName("snapshots"))
		assert.Equal(t, "myapp_read_models", manager.GetCollectionName("read_models"))

		// Custom collection should also get prefix
		assert.Equal(t, "myapp_custom", manager.GetCollectionName("custom"))
	})

	t.Run("with custom collection names", func(t *testing.T) {
		customNames := &CollectionNames{
			Events:     "my_events",
			Snapshots:  "my_snapshots",
			ReadModels: "my_read_models",
		}

		manager, err := NewMongoClientManagerWithCollections(config, "", customNames)
		if err != nil {
			t.Skipf("MongoDB not available: %v", err)
		}
		defer cleanupTestMongoClient(t, manager)

		names := manager.GetCollectionNames()
		assert.Equal(t, "my_events", names.Events)
		assert.Equal(t, "my_snapshots", names.Snapshots)
		assert.Equal(t, "my_read_models", names.ReadModels)
	})

	t.Run("with prefix and custom collection names", func(t *testing.T) {
		customNames := &CollectionNames{
			Events:     "events",
			Snapshots:  "snapshots",
			ReadModels: "read_models",
		}

		manager, err := NewMongoClientManagerWithCollections(config, "myapp", customNames)
		if err != nil {
			t.Skipf("MongoDB not available: %v", err)
		}
		defer cleanupTestMongoClient(t, manager)

		names := manager.GetCollectionNames()
		assert.Equal(t, "myapp_events", names.Events)
		assert.Equal(t, "myapp_snapshots", names.Snapshots)
		assert.Equal(t, "myapp_read_models", names.ReadModels)
	})
}
