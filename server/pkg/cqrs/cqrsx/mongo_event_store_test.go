package cqrsx

// import (
// 	"context"
// 	"cqrs"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// // TestMongoEventStore_BasicOperations tests basic event store operations
// func TestMongoEventStore_BasicOperations(t *testing.T) {
// 	// Skip if no MongoDB available
// 	if testing.Short() {
// 		t.Skip("Skipping MongoDB integration test")
// 	}

// 	ctx := context.Background()

// 	// Setup
// 	client, err := NewMongoClientManager(&MongoConfig{
// 		URI:                    "mongodb://localhost:27017",
// 		Database:               "cqrs_test",
// 		ConnectTimeout:         10 * time.Second,
// 		ServerSelectionTimeout: 5 * time.Second,
// 		MaxPoolSize:            10,
// 	})
// 	require.NoError(t, err)

// 	eventStore := NewMongoEventStore(client, "test_events")

// 	// Cleanup
// 	defer func() {
// 		client.GetDatabase().Drop(ctx)
// 		client.Close(ctx)
// 	}()

// 	aggregateID := "test-aggregate-1"
// 	aggregateType := "TestAggregate"

// 	t.Run("SaveAndLoadSingleEvent", func(t *testing.T) {
// 		// Create test event
// 		event := cqrs.NewBaseEventMessageWithConfig(
// 			"TestEvent",
// 			aggregateID,
// 			aggregateType,
// 			1,
// 			map[string]interface{}{
// 				"field1": "value1",
// 				"field2": 42,
// 			},
// 		)

// 		// Save event
// 		err := eventStore.SaveEvents(ctx, aggregateID, []cqrs.EventMessage{event}, 0)
// 		assert.NoError(t, err)

// 		// Load events
// 		events, err := eventStore.GetEventHistory(ctx, aggregateID, aggregateType, 1)
// 		assert.NoError(t, err)
// 		assert.Len(t, events, 1)

// 		loadedEvent := events[0]
// 		assert.Equal(t, "TestEvent", loadedEvent.EventType())
// 		assert.Equal(t, aggregateID, loadedEvent.ID())
// 		assert.Equal(t, aggregateType, loadedEvent.Type())
// 		assert.Equal(t, 1, loadedEvent.Version())
// 	})

// 	t.Run("SaveMultipleEventsSequentially", func(t *testing.T) {
// 		aggregateID2 := "test-aggregate-2"

// 		// Save first event
// 		event1 := cqrs.NewBaseEventMessageWithConfig("Event1", aggregateID2, aggregateType, 1, map[string]interface{}{"data": "data1"})
// 		err := eventStore.SaveEvents(ctx, aggregateID2, []cqrs.EventMessage{event1}, 0)
// 		assert.NoError(t, err)

// 		// Save second event
// 		event2 := cqrs.NewBaseEventMessageWithConfig("Event2", aggregateID2, aggregateType, 2, map[string]interface{}{"data": "data2"})
// 		err = eventStore.SaveEvents(ctx, aggregateID2, []cqrs.EventMessage{event2}, 1)
// 		assert.NoError(t, err)

// 		// Save third event
// 		event3 := cqrs.NewBaseEventMessageWithConfig("Event3", aggregateID2, aggregateType, 3, map[string]interface{}{"data": "data3"})
// 		err = eventStore.SaveEvents(ctx, aggregateID2, []cqrs.EventMessage{event3}, 2)
// 		assert.NoError(t, err)

// 		// Load all events
// 		events, err := eventStore.GetEventHistory(ctx, aggregateID2, aggregateType, 1)
// 		assert.NoError(t, err)
// 		assert.Len(t, events, 3)

// 		// Verify order and versions
// 		assert.Equal(t, "Event1", events[0].EventType())
// 		assert.Equal(t, 1, events[0].Version())
// 		assert.Equal(t, "Event2", events[1].EventType())
// 		assert.Equal(t, 2, events[1].Version())
// 		assert.Equal(t, "Event3", events[2].EventType())
// 		assert.Equal(t, 3, events[2].Version())
// 	})

// 	t.Run("GetLastEventVersion", func(t *testing.T) {
// 		aggregateID3 := "test-aggregate-3"

// 		// No events - should return 0
// 		version, err := eventStore.GetLastEventVersion(ctx, aggregateID3, aggregateType)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 0, version)

// 		// Save one event
// 		event1 := cqrs.NewBaseEventMessageWithConfig("Event1", aggregateID3, aggregateType, 1, map[string]interface{}{"data": "data1"})
// 		err = eventStore.SaveEvents(ctx, aggregateID3, []cqrs.EventMessage{event1}, 0)
// 		assert.NoError(t, err)

// 		// Should return 1
// 		version, err = eventStore.GetLastEventVersion(ctx, aggregateID3, aggregateType)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1, version)

// 		// Save another event
// 		event2 := cqrs.NewBaseEventMessageWithConfig("Event2", aggregateID3, aggregateType, 2, map[string]interface{}{"data": "data2"})
// 		err = eventStore.SaveEvents(ctx, aggregateID3, []cqrs.EventMessage{event2}, 1)
// 		assert.NoError(t, err)

// 		// Should return 2
// 		version, err = eventStore.GetLastEventVersion(ctx, aggregateID3, aggregateType)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 2, version)
// 	})

// 	t.Run("ConcurrencyControl", func(t *testing.T) {
// 		aggregateID4 := "test-aggregate-4"

// 		// Save first event
// 		event1 := cqrs.NewBaseEventMessageWithConfig("Event1", aggregateID4, aggregateType, 1, map[string]interface{}{"data": "data1"})
// 		err := eventStore.SaveEvents(ctx, aggregateID4, []cqrs.EventMessage{event1}, 0)
// 		assert.NoError(t, err)

// 		// Try to save with wrong expected version - should fail
// 		event2 := cqrs.NewBaseEventMessageWithConfig("Event2", aggregateID4, aggregateType, 2, map[string]interface{}{"data": "data2"})
// 		err = eventStore.SaveEvents(ctx, aggregateID4, []cqrs.EventMessage{event2}, 0) // Wrong expected version
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "concurrency conflict")

// 		// Save with correct expected version - should succeed
// 		err = eventStore.SaveEvents(ctx, aggregateID4, []cqrs.EventMessage{event2}, 1) // Correct expected version
// 		assert.NoError(t, err)

// 		// Verify final state
// 		version, err := eventStore.GetLastEventVersion(ctx, aggregateID4, aggregateType)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 2, version)
// 	})

// 	t.Run("LoadEventsFromVersion", func(t *testing.T) {
// 		aggregateID5 := "test-aggregate-5"

// 		// Save multiple events
// 		events := []cqrs.EventMessage{
// 			cqrs.NewBaseEventMessageWithConfig("Event1", aggregateID5, aggregateType, 1, map[string]interface{}{"data": "data1"}),
// 			cqrs.NewBaseEventMessageWithConfig("Event2", aggregateID5, aggregateType, 2, map[string]interface{}{"data": "data2"}),
// 			cqrs.NewBaseEventMessageWithConfig("Event3", aggregateID5, aggregateType, 3, map[string]interface{}{"data": "data3"}),
// 		}

// 		for i, event := range events {
// 			err := eventStore.SaveEvents(ctx, aggregateID5, []cqrs.EventMessage{event}, i)
// 			assert.NoError(t, err)
// 		}

// 		// Load from version 2
// 		loadedEvents, err := eventStore.GetEventHistory(ctx, aggregateID5, aggregateType, 2)
// 		assert.NoError(t, err)
// 		assert.Len(t, loadedEvents, 2) // Should get events 2 and 3

// 		assert.Equal(t, "Event2", loadedEvents[0].EventType())
// 		assert.Equal(t, 2, loadedEvents[0].Version())
// 		assert.Equal(t, "Event3", loadedEvents[1].EventType())
// 		assert.Equal(t, 3, loadedEvents[1].Version())
// 	})
// }

// // TestMongoEventStore_EdgeCases tests edge cases and error conditions
// func TestMongoEventStore_EdgeCases(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping MongoDB integration test")
// 	}

// 	ctx := context.Background()

// 	client, err := NewMongoClientManager(&MongoConfig{
// 		URI:                    "mongodb://localhost:27017",
// 		Database:               "cqrs_test_edge",
// 		ConnectTimeout:         10 * time.Second,
// 		ServerSelectionTimeout: 5 * time.Second,
// 		MaxPoolSize:            10,
// 	})
// 	require.NoError(t, err)

// 	eventStore := NewMongoEventStore(client, "test_events")

// 	defer func() {
// 		client.GetDatabase().Drop(ctx)
// 		client.Close(ctx)
// 	}()

// 	t.Run("EmptyAggregateID", func(t *testing.T) {
// 		event := cqrs.NewBaseEventMessageWithConfig("TestEvent", "", "TestAggregate", 1, map[string]interface{}{"data": "test"})
// 		err := eventStore.SaveEvents(ctx, "", []cqrs.EventMessage{event}, 0)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "aggregate ID cannot be empty")
// 	})

// 	t.Run("EmptyEventsList", func(t *testing.T) {
// 		err := eventStore.SaveEvents(ctx, "test-id", []cqrs.EventMessage{}, 0)
// 		assert.NoError(t, err) // Should be no-op
// 	})

// 	t.Run("NonExistentAggregate", func(t *testing.T) {
// 		events, err := eventStore.GetEventHistory(ctx, "non-existent", "TestAggregate", 1)
// 		assert.NoError(t, err)
// 		assert.Len(t, events, 0)

// 		version, err := eventStore.GetLastEventVersion(ctx, "non-existent", "TestAggregate")
// 		assert.NoError(t, err)
// 		assert.Equal(t, 0, version)
// 	})
// }
