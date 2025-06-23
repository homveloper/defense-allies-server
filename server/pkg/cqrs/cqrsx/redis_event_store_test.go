package cqrsx

// import (
// 	"context"
// 	"cqrs"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// )

// func TestJSONEventSerializer_Serialize(t *testing.T) {
// 	// Arrange
// 	serializer := &JSONEventMarshaler{}
// 	event := cqrs.NewBaseEventMessageWithConfig("TestEvent", "test-id", "TestAggregate", 1, "test data")
// 	event.AddMetadata("key1", "value1")

// 	// Act
// 	data, err := serializer.Serialize(event)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, data)
// }

// func TestJSONEventSerializer_Deserialize(t *testing.T) {
// 	// Arrange
// 	serializer := &JSONEventMarshaler{}
// 	originalEvent := cqrs.NewBaseEventMessageWithConfig("TestEvent", "test-id", "TestAggregate", 1, "test data")
// 	originalEvent.AddMetadata("key1", "value1")

// 	// Serialize first
// 	data, err := serializer.Serialize(originalEvent)
// 	assert.NoError(t, err)

// 	// Act
// 	deserializedEvent, err := serializer.Deserialize(data)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, deserializedEvent)
// 	assert.Equal(t, originalEvent.EventType(), deserializedEvent.EventType())
// 	assert.Equal(t, originalEvent.ID(), deserializedEvent.ID())
// 	assert.Equal(t, originalEvent.Type(), deserializedEvent.Type())
// 	assert.Equal(t, originalEvent.Version(), deserializedEvent.Version())
// 	assert.Equal(t, originalEvent.EventData(), deserializedEvent.EventData())

// 	// Check metadata
// 	metadata := deserializedEvent.Metadata()
// 	assert.Equal(t, "value1", metadata["key1"])
// }

// func TestJSONEventSerializer_SerializeDeserialize_RoundTrip(t *testing.T) {
// 	// Arrange
// 	serializer := &JSONEventMarshaler{}
// 	originalEvent := cqrs.NewBaseEventMessageWithConfig("UserCreated", "user-123", "User", 1, map[string]interface{}{
// 		"name":  "John Doe",
// 		"email": "john@example.com",
// 	})
// 	originalEvent.AddMetadata("userId", "user-123")
// 	originalEvent.AddMetadata("timestamp", time.Now().Unix())

// 	// Act - Serialize
// 	data, err := serializer.Serialize(originalEvent)
// 	assert.NoError(t, err)

// 	// Act - Deserialize
// 	deserializedEvent, err := serializer.Deserialize(data)
// 	assert.NoError(t, err)

// 	// Assert
// 	assert.Equal(t, originalEvent.EventID(), deserializedEvent.EventID())
// 	assert.Equal(t, originalEvent.EventType(), deserializedEvent.EventType())
// 	assert.Equal(t, originalEvent.ID(), deserializedEvent.ID())
// 	assert.Equal(t, originalEvent.Type(), deserializedEvent.Type())
// 	assert.Equal(t, originalEvent.Version(), deserializedEvent.Version())

// 	// Compare event data (as JSON objects)
// 	originalData := originalEvent.EventData().(map[string]interface{})
// 	deserializedData := deserializedEvent.EventData().(map[string]interface{})
// 	assert.Equal(t, originalData["name"], deserializedData["name"])
// 	assert.Equal(t, originalData["email"], deserializedData["email"])

// 	// Compare metadata
// 	originalMetadata := originalEvent.Metadata()
// 	deserializedMetadata := deserializedEvent.Metadata()
// 	assert.Equal(t, originalMetadata["userId"], deserializedMetadata["userId"])
// }

// func TestJSONEventSerializer_Deserialize_InvalidJSON(t *testing.T) {
// 	// Arrange
// 	serializer := &JSONEventMarshaler{}
// 	invalidJSON := []byte("invalid json")

// 	// Act
// 	event, err := serializer.Deserialize(invalidJSON)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Nil(t, event)
// }

// func TestNewRedisEventStore(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	keyPrefix := "test"

// 	// Act
// 	eventStore := NewRedisEventStore(client, keyPrefix)

// 	// Assert
// 	assert.NotNil(t, eventStore)
// 	assert.Equal(t, client, eventStore.client)
// 	assert.Equal(t, keyPrefix, eventStore.keyBuilder.GetPrefix())
// 	assert.NotNil(t, eventStore.serializer)
// }

// // Note: The following tests would require a running Redis instance
// // In a real test environment, you would either:
// // 1. Use a test Redis instance
// // 2. Use Redis mocks
// // 3. Use integration test tags to run only when Redis is available

// func TestRedisEventStore_SaveEvents_EmptyEvents(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	err = eventStore.SaveEvents(context.Background(), "test-id", []cqrs.EventMessage{}, 0)

// 	// Assert
// 	assert.NoError(t, err) // Should not error for empty events
// }

// func TestRedisEventStore_SaveEvents_EmptyAggregateID(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")
// 	event := cqrs.NewBaseEventMessageWithConfig("TestEvent", "test-id", "TestAggregate", 1, "test data")

// 	// Act
// 	err = eventStore.SaveEvents(context.Background(), "", []cqrs.EventMessage{event}, 0)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "aggregate ID cannot be empty")
// }

// func TestRedisEventStore_GetEventHistory_EmptyAggregateID(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	events, err := eventStore.GetEventHistory(context.Background(), "", "TestAggregate", 0)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Nil(t, events)
// 	assert.Contains(t, err.Error(), "aggregate ID cannot be empty")
// }

// func TestRedisEventStore_GetEventHistory_EmptyAggregateType(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	events, err := eventStore.GetEventHistory(context.Background(), "test-id", "", 0)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Nil(t, events)
// 	assert.Contains(t, err.Error(), "aggregate type cannot be empty")
// }

// func TestRedisEventStore_GetLastEventVersion_EmptyAggregateID(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	version, err := eventStore.GetLastEventVersion(context.Background(), "", "TestAggregate")

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Equal(t, 0, version)
// 	assert.Contains(t, err.Error(), "aggregate ID cannot be empty")
// }

// func TestRedisEventStore_GetLastEventVersion_EmptyAggregateType(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	version, err := eventStore.GetLastEventVersion(context.Background(), "test-id", "")

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Equal(t, 0, version)
// 	assert.Contains(t, err.Error(), "aggregate type cannot be empty")
// }

// func TestRedisEventStore_CompactEvents_EmptyAggregateID(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	err = eventStore.CompactEvents(context.Background(), "", "TestAggregate", 5)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "aggregate ID cannot be empty")
// }

// func TestRedisEventStore_CompactEvents_EmptyAggregateType(t *testing.T) {
// 	// Arrange
// 	config := &RedisConfig{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
// 	client, err := NewRedisClientManager(config)
// 	assert.NoError(t, err)
// 	defer client.Close()

// 	eventStore := NewRedisEventStore(client, "test")

// 	// Act
// 	err = eventStore.CompactEvents(context.Background(), "test-id", "", 5)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "aggregate type cannot be empty")
// }
