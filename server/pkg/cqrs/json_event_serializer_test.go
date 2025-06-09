package cqrs

import (
	"encoding/json"
	"testing"
	"time"
)

// Test event data structures
type UserCreatedEventData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type UserUpdatedEventData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type InvalidEventData struct {
	Channel chan int `json:"channel"` // channels cannot be marshaled to JSON
}

func TestEventDataRegistry_RegisterEventData(t *testing.T) {
	registry := NewEventDataRegistry()

	// Test successful registration
	err := registry.RegisterEventData("UserCreated", UserCreatedEventData{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test registration with invalid data (should fail JSON validation)
	err = registry.RegisterEventData("InvalidEvent", InvalidEventData{})
	if err == nil {
		t.Error("Expected error for invalid event data, got nil")
	}

	// Test registration with empty event type
	err = registry.RegisterEventData("", UserCreatedEventData{})
	if err == nil {
		t.Error("Expected error for empty event type, got nil")
	}

	// Test registration with nil event data
	err = registry.RegisterEventData("NilEvent", nil)
	if err == nil {
		t.Error("Expected error for nil event data, got nil")
	}
}

func TestEventDataRegistry_GetEventDataType(t *testing.T) {
	registry := NewEventDataRegistry()

	// Register event data
	err := registry.RegisterEventData("UserCreated", UserCreatedEventData{})
	if err != nil {
		t.Fatalf("Failed to register event data: %v", err)
	}

	// Test getting registered type
	dataType, err := registry.GetEventDataType("UserCreated")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if dataType == nil {
		t.Error("Expected data type, got nil")
	}

	// Test getting unregistered type
	_, err = registry.GetEventDataType("UnregisteredEvent")
	if err == nil {
		t.Error("Expected error for unregistered event type, got nil")
	}
}

func TestEventDataRegistry_IsRegistered(t *testing.T) {
	registry := NewEventDataRegistry()

	// Register event data
	err := registry.RegisterEventData("UserCreated", UserCreatedEventData{})
	if err != nil {
		t.Fatalf("Failed to register event data: %v", err)
	}

	// Test registered type
	if !registry.IsRegistered("UserCreated") {
		t.Error("Expected UserCreated to be registered")
	}

	// Test unregistered type
	if registry.IsRegistered("UnregisteredEvent") {
		t.Error("Expected UnregisteredEvent to not be registered")
	}
}

func TestEventDataRegistry_CreateEventDataInstance(t *testing.T) {
	registry := NewEventDataRegistry()

	// Register event data
	err := registry.RegisterEventData("UserCreated", UserCreatedEventData{})
	if err != nil {
		t.Fatalf("Failed to register event data: %v", err)
	}

	// Test creating instance
	instance, err := registry.CreateEventDataInstance("UserCreated")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if instance == nil {
		t.Error("Expected instance, got nil")
	}

	// Verify instance type (should be a pointer)
	if _, ok := instance.(*UserCreatedEventData); !ok {
		t.Errorf("Expected *UserCreatedEventData, got %T", instance)
	}

	// Test creating instance for unregistered type
	_, err = registry.CreateEventDataInstance("UnregisteredEvent")
	if err == nil {
		t.Error("Expected error for unregistered event type, got nil")
	}
}

func TestJSONEventSerializer_Serialize(t *testing.T) {
	registry := NewEventDataRegistry()
	serializer := NewJSONEventSerializer(registry)

	// Create test event
	eventData := UserCreatedEventData{
		UserID:   "user-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	event := NewBaseEventMessage("UserCreated", eventData)
	event.SetEventID("event-123")
	event.SetTimestamp(time.Now())

	// Test serialization
	data, err := serializer.Serialize(event)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected serialized data, got empty")
	}

	// Verify JSON structure
	var serializedEvent SerializableEventMessage
	err = json.Unmarshal(data, &serializedEvent)
	if err != nil {
		t.Errorf("Failed to unmarshal serialized data: %v", err)
	}

	if serializedEvent.EventType != "UserCreated" {
		t.Errorf("Expected event type 'UserCreated', got %s", serializedEvent.EventType)
	}
}

func TestJSONEventSerializer_Deserialize(t *testing.T) {
	registry := NewEventDataRegistry()

	// Register event data
	err := registry.RegisterEventData("UserCreated", UserCreatedEventData{})
	if err != nil {
		t.Fatalf("Failed to register event data: %v", err)
	}

	serializer := NewJSONEventSerializer(registry)

	// Create test event
	originalEventData := UserCreatedEventData{
		UserID:   "user-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	originalEvent := NewBaseEventMessage("UserCreated", originalEventData)
	originalEvent.SetEventID("event-123")
	originalEvent.SetTimestamp(time.Now())

	// Serialize
	data, err := serializer.Serialize(originalEvent)
	if err != nil {
		t.Fatalf("Failed to serialize event: %v", err)
	}

	// Deserialize
	deserializedEvent, err := serializer.Deserialize(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if deserializedEvent == nil {
		t.Error("Expected deserialized event, got nil")
	}

	// Verify event properties
	if deserializedEvent.EventType() != "UserCreated" {
		t.Errorf("Expected event type 'UserCreated', got %s", deserializedEvent.EventType())
	}

	if deserializedEvent.EventID() != "event-123" {
		t.Errorf("Expected event ID 'event-123', got %s", deserializedEvent.EventID())
	}

	// Verify event data (should be pointer type)
	eventData, ok := deserializedEvent.EventData().(*UserCreatedEventData)
	if !ok {
		t.Errorf("Expected *UserCreatedEventData, got %T", deserializedEvent.EventData())
	} else {
		if eventData.UserID != "user-123" {
			t.Errorf("Expected UserID 'user-123', got %s", eventData.UserID)
		}
		if eventData.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got %s", eventData.Username)
		}
		if eventData.Email != "test@example.com" {
			t.Errorf("Expected Email 'test@example.com', got %s", eventData.Email)
		}
	}
}

func TestJSONEventSerializer_DeserializeUnregisteredType(t *testing.T) {
	registry := NewEventDataRegistry()
	serializer := NewJSONEventSerializer(registry)

	// Create test event with unregistered type
	eventData := map[string]interface{}{
		"user_id":  "user-123",
		"username": "testuser",
		"email":    "test@example.com",
	}

	event := NewBaseEventMessage("UnregisteredEvent", eventData)
	event.SetEventID("event-123")
	event.SetTimestamp(time.Now())

	// Serialize
	data, err := serializer.Serialize(event)
	if err != nil {
		t.Fatalf("Failed to serialize event: %v", err)
	}

	// Deserialize (should work even for unregistered types)
	deserializedEvent, err := serializer.Deserialize(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if deserializedEvent == nil {
		t.Error("Expected deserialized event, got nil")
	}

	// Verify event properties
	if deserializedEvent.EventType() != "UnregisteredEvent" {
		t.Errorf("Expected event type 'UnregisteredEvent', got %s", deserializedEvent.EventType())
	}
}

func TestJSONEventSerializer_ValidateEventData(t *testing.T) {
	serializer := NewJSONEventSerializer(nil)

	// Test valid event data
	validData := UserCreatedEventData{
		UserID:   "user-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	err := serializer.ValidateEventData(validData)
	if err != nil {
		t.Errorf("Expected no error for valid data, got %v", err)
	}

	// Test invalid event data
	invalidData := InvalidEventData{
		Channel: make(chan int),
	}

	err = serializer.ValidateEventData(invalidData)
	if err == nil {
		t.Error("Expected error for invalid data, got nil")
	}
}
