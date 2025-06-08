package redisstream

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"defense-allies-server/pkg/cqrs"
)

// TestEventSerializer tests the EventSerializer interface
func TestEventSerializer_JSON(t *testing.T) {
	serializer := NewJSONEventSerializer()

	t.Run("should serialize and deserialize event correctly", func(t *testing.T) {
		// Create test event
		baseOptions := cqrs.Options().
			WithAggregateID("test-aggregate-123").
			WithAggregateType("TestAggregate").
			WithVersion(1).
			WithTimestamp(time.Date(2025, 6, 8, 12, 0, 0, 0, time.UTC))

		originalEvent := cqrs.NewBaseDomainEventMessage(
			"TestEventSerialized",
			map[string]interface{}{
				"name":     "Test User",
				"email":    "test@example.com",
				"age":      30,
				"active":   true,
				"tags":     []string{"tag1", "tag2"},
				"metadata": map[string]interface{}{"source": "test"},
			},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// Serialize
		data, err := serializer.Serialize(originalEvent)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Deserialize
		deserializedEvent, err := serializer.Deserialize(data)
		require.NoError(t, err)

		// Verify basic fields
		assert.Equal(t, originalEvent.EventID(), deserializedEvent.EventID())
		assert.Equal(t, originalEvent.EventType(), deserializedEvent.EventType())
		assert.Equal(t, originalEvent.ID(), deserializedEvent.ID())
		assert.Equal(t, originalEvent.Type(), deserializedEvent.Type())
		assert.Equal(t, originalEvent.Version(), deserializedEvent.Version())
		assert.True(t, originalEvent.Timestamp().Equal(deserializedEvent.Timestamp()))

		// Verify event data
		originalData := originalEvent.EventData().(map[string]interface{})
		deserializedData := deserializedEvent.EventData().(map[string]interface{})

		assert.Equal(t, originalData["name"], deserializedData["name"])
		assert.Equal(t, originalData["email"], deserializedData["email"])
		assert.Equal(t, float64(30), deserializedData["age"]) // JSON numbers are float64
		assert.Equal(t, originalData["active"], deserializedData["active"])
	})

	t.Run("should handle complex nested data", func(t *testing.T) {
		complexData := map[string]interface{}{
			"user": map[string]interface{}{
				"profile": map[string]interface{}{
					"personal": map[string]interface{}{
						"name": "John Doe",
						"age":  25,
					},
					"settings": map[string]interface{}{
						"theme":         "dark",
						"notifications": true,
						"preferences":   []string{"email", "sms"},
					},
				},
			},
			"timestamp": time.Now().Unix(),
			"scores":    []float64{1.1, 2.2, 3.3},
		}

		baseOptions := cqrs.Options().
			WithAggregateID("complex-123").
			WithAggregateType("ComplexAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"ComplexEventType",
			complexData,
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		// Serialize and deserialize
		data, err := serializer.Serialize(event)
		require.NoError(t, err)

		deserializedEvent, err := serializer.Deserialize(data)
		require.NoError(t, err)

		// Verify complex nested structure
		deserializedData := deserializedEvent.EventData().(map[string]interface{})
		user := deserializedData["user"].(map[string]interface{})
		profile := user["profile"].(map[string]interface{})
		personal := profile["personal"].(map[string]interface{})

		assert.Equal(t, "John Doe", personal["name"])
		assert.Equal(t, float64(25), personal["age"])
	})

	t.Run("should handle domain event metadata", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("domain-123").
			WithAggregateType("DomainAggregate")

		domainOptions := &cqrs.BaseDomainEventMessageOptions{}
		domainOptions.IssuerID = &[]string{"user-456"}[0]
		domainOptions.IssuerType = &[]cqrs.IssuerType{cqrs.UserIssuer}[0]
		domainOptions.CausationID = &[]string{"cmd-789"}[0]
		domainOptions.CorrelationID = &[]string{"corr-101"}[0]
		domainOptions.Category = &[]cqrs.EventCategory{cqrs.DomainEvent}[0]
		domainOptions.Priority = &[]cqrs.EventPriority{cqrs.PriorityHigh}[0]

		event := cqrs.NewBaseDomainEventMessage(
			"DomainEventType",
			map[string]interface{}{"action": "domain_action"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
			domainOptions,
		)

		// Serialize and deserialize
		data, err := serializer.Serialize(event)
		require.NoError(t, err)

		deserializedEvent, err := serializer.Deserialize(data)
		require.NoError(t, err)

		// Verify domain event interface
		if domainEvent, ok := deserializedEvent.(cqrs.DomainEventMessage); ok {
			assert.Equal(t, "user-456", domainEvent.IssuerID())
			assert.Equal(t, cqrs.UserIssuer, domainEvent.IssuerType())
			assert.Equal(t, "cmd-789", domainEvent.CausationID())
			assert.Equal(t, "corr-101", domainEvent.CorrelationID())
			assert.Equal(t, cqrs.DomainEvent, domainEvent.GetEventCategory())
			assert.Equal(t, cqrs.PriorityHigh, domainEvent.GetPriority())
		} else {
			t.Fatal("Expected DomainEventMessage interface")
		}
	})

	t.Run("should handle invalid data gracefully", func(t *testing.T) {
		// Test with invalid JSON
		_, err := serializer.Deserialize([]byte("invalid json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to deserialize")

		// Test with incomplete data
		incompleteJSON := `{"event_type": "Test"}`
		_, err = serializer.Deserialize([]byte(incompleteJSON))
		assert.Error(t, err)
	})

	t.Run("should validate serialization format", func(t *testing.T) {
		baseOptions := cqrs.Options().
			WithAggregateID("validate-123").
			WithAggregateType("ValidateAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"ValidateEventType",
			map[string]interface{}{"test": "data"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		data, err := serializer.Serialize(event)
		require.NoError(t, err)

		// Verify it's valid JSON
		var jsonData map[string]interface{}
		err = json.Unmarshal(data, &jsonData)
		require.NoError(t, err)

		// Verify required fields exist
		assert.Contains(t, jsonData, "event_id")
		assert.Contains(t, jsonData, "event_type")
		assert.Contains(t, jsonData, "aggregate_id")
		assert.Contains(t, jsonData, "aggregate_type")
		assert.Contains(t, jsonData, "version")
		assert.Contains(t, jsonData, "event_data")
		assert.Contains(t, jsonData, "timestamp")
	})
}

// TestEventSerializer_Protobuf tests Protobuf serialization (if implemented)
func TestEventSerializer_Protobuf(t *testing.T) {
	t.Skip("Protobuf serialization not implemented yet")

	// Future implementation:
	// serializer := NewProtobufEventSerializer()
	// ... similar tests as JSON but with protobuf specifics
}

// TestEventSerializer_Avro tests Avro serialization (if implemented)
func TestEventSerializer_Avro(t *testing.T) {
	t.Skip("Avro serialization not implemented yet")

	// Future implementation:
	// serializer := NewAvroEventSerializer()
	// ... similar tests as JSON but with avro specifics
}

// TestSerializationRegistry tests the serialization registry
func TestSerializationRegistry(t *testing.T) {
	registry := NewSerializationRegistry()

	t.Run("should register and retrieve serializers", func(t *testing.T) {
		jsonSerializer := NewJSONEventSerializer()

		err := registry.Register(SerializationFormatJSON, jsonSerializer)
		assert.NoError(t, err)

		retrievedSerializer, err := registry.Get(SerializationFormatJSON)
		assert.NoError(t, err)
		assert.Equal(t, jsonSerializer, retrievedSerializer)
	})

	t.Run("should return error for unknown format", func(t *testing.T) {
		_, err := registry.Get(SerializationFormat("unknown"))
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSerializationFormatNotSupported)
	})

	t.Run("should not allow duplicate registration", func(t *testing.T) {
		jsonSerializer1 := NewJSONEventSerializer()
		jsonSerializer2 := NewJSONEventSerializer()

		err := registry.Register(SerializationFormatJSON, jsonSerializer1)
		assert.NoError(t, err)

		err = registry.Register(SerializationFormatJSON, jsonSerializer2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSerializationFormatAlreadyRegistered)
	})

	t.Run("should list supported formats", func(t *testing.T) {
		registry := NewSerializationRegistry()

		// Initially empty
		formats := registry.SupportedFormats()
		assert.Empty(t, formats)

		// After registration
		jsonSerializer := NewJSONEventSerializer()
		registry.Register(SerializationFormatJSON, jsonSerializer)

		formats = registry.SupportedFormats()
		assert.Len(t, formats, 1)
		assert.Contains(t, formats, SerializationFormatJSON)
	})
}

// TestSerializationPerformance tests serialization performance
func TestSerializationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	serializer := NewJSONEventSerializer()

	// Create a large event for testing
	largeData := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("field_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	baseOptions := cqrs.Options().
		WithAggregateID("perf-test-123").
		WithAggregateType("PerfTestAggregate")

	event := cqrs.NewBaseDomainEventMessage(
		"PerformanceTestEvent",
		largeData,
		[]*cqrs.BaseEventMessageOptions{baseOptions},
	)

	t.Run("serialization performance", func(t *testing.T) {
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			_, err := serializer.Serialize(event)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		t.Logf("Average serialization time: %v", avgDuration)

		// Should be under 1ms for reasonable performance
		assert.Less(t, avgDuration, time.Millisecond, "Serialization too slow")
	})

	t.Run("deserialization performance", func(t *testing.T) {
		// First serialize the event
		data, err := serializer.Serialize(event)
		require.NoError(t, err)

		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			_, err := serializer.Deserialize(data)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		avgDuration := duration / time.Duration(iterations)

		t.Logf("Average deserialization time: %v", avgDuration)

		// Should be under 1ms for reasonable performance
		assert.Less(t, avgDuration, time.Millisecond, "Deserialization too slow")
	})
}

// TestSerializationCompatibility tests backward compatibility
func TestSerializationCompatibility(t *testing.T) {
	t.Run("should handle version upgrades", func(t *testing.T) {
		// This test would verify that newer versions can read older serialized data
		// For now, we'll just verify the current version works

		serializer := NewJSONEventSerializer()

		baseOptions := cqrs.Options().
			WithAggregateID("compat-123").
			WithAggregateType("CompatAggregate")

		event := cqrs.NewBaseDomainEventMessage(
			"CompatibilityTest",
			map[string]interface{}{"version": "1.0"},
			[]*cqrs.BaseEventMessageOptions{baseOptions},
		)

		data, err := serializer.Serialize(event)
		require.NoError(t, err)

		deserializedEvent, err := serializer.Deserialize(data)
		require.NoError(t, err)

		assert.Equal(t, event.EventType(), deserializedEvent.EventType())
	})
}
