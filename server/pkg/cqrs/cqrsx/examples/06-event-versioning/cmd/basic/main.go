package main

import (
	"fmt"
	"log"
	"time"

	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/06-event-versioning/domain"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/06-event-versioning/versioning"
)

func main() {
	fmt.Println("🎯 Event Versioning Example - Basic Demo")
	fmt.Println("========================================")

	// Initialize version manager
	versionManager := setupVersionManager()

	// Demo 1: Version Detection
	fmt.Println("\n📋 Demo 1: Event Version Detection")
	demoVersionDetection(versionManager)

	// Demo 2: Upcasting (V1 -> V2 -> V3)
	fmt.Println("\n⬆️  Demo 2: Event Upcasting")
	demoUpcasting(versionManager)

	// Demo 3: Downcasting (V3 -> V2 -> V1)
	fmt.Println("\n⬇️  Demo 3: Event Downcasting")
	demoDowncasting(versionManager)

	// Demo 4: Version Compatibility
	fmt.Println("\n🔄 Demo 4: Version Compatibility")
	demoVersionCompatibility(versionManager)

	fmt.Println("\n✅ Event Versioning Demo completed successfully!")
}

// setupVersionManager initializes the version manager with all components
func setupVersionManager() *versioning.UserEventVersionManager {
	// Create event factories for each version
	factories := map[int]versioning.EventFactory{
		1: &domain.V1EventFactory{},
		2: &domain.V2EventFactory{},
		3: &domain.V3EventFactory{},
	}

	// Create upcaster and downcaster
	upcaster := versioning.NewUserEventUpcaster()
	downcaster := versioning.NewUserEventDowncaster()

	// Create version manager
	return versioning.NewUserEventVersionManager(upcaster, downcaster, factories)
}

// demoVersionDetection demonstrates automatic version detection
func demoVersionDetection(vm *versioning.UserEventVersionManager) {

	// Create events of different versions
	v1Event := domain.NewUserCreatedV1("user-1", "John Doe", "john@example.com")
	v2Event := domain.NewUserCreatedV2(
		"user-2",
		"Jane Smith",
		"jane@example.com",
		domain.UserProfile{
			FirstName: "Jane",
			LastName:  "Smith",
			Gender:    "female",
		},
		domain.DefaultUserPreferences(),
	)
	v3Event := domain.NewUserCreatedV3(
		"user-3",
		domain.PersonalInfo{
			FullName: domain.FullName{
				FirstName: "Bob",
				LastName:  "Johnson",
			},
			Gender: "male",
		},
		domain.ContactInfo{
			PrimaryEmail: "bob@example.com",
		},
		domain.DefaultUserPreferences(),
		domain.EventMetadata{
			Version: "3.0",
			Source:  "web",
		},
	)

	events := []cqrs.EventMessage{v1Event, v2Event, v3Event}

	for i, event := range events {
		eventData, _ := event.EventData().(map[string]interface{})
		eventBytes, _ := marshalEventData(eventData)

		version, err := vm.DetectVersion(eventBytes, event.Metadata())
		if err != nil {
			log.Printf("❌ Failed to detect version for event %d: %v", i+1, err)
			continue
		}

		fmt.Printf("   Event %d: Detected version %d ✅\n", i+1, version)
		fmt.Printf("     Type: %s\n", event.EventType())
		fmt.Printf("     Metadata: %v\n", event.Metadata())
	}
}

// demoUpcasting demonstrates converting events to higher versions
func demoUpcasting(vm *versioning.UserEventVersionManager) {
	// Create a V1 event
	v1Event := domain.NewUserCreatedV1("user-upcast", "Alice Cooper", "alice@example.com")

	fmt.Printf("   Original V1 Event:\n")
	printEventDetails(v1Event)

	// Upcast V1 -> V2
	v2Event, err := vm.ConvertToVersion(v1Event, 2)
	if err != nil {
		log.Printf("❌ Failed to upcast V1 to V2: %v", err)
		return
	}

	fmt.Printf("\n   Upcasted to V2:\n")
	printEventDetails(v2Event)

	// Upcast V2 -> V3
	v3Event, err := vm.ConvertToVersion(v2Event, 3)
	if err != nil {
		log.Printf("❌ Failed to upcast V2 to V3: %v", err)
		return
	}

	fmt.Printf("\n   Upcasted to V3:\n")
	printEventDetails(v3Event)

	// Direct upcast V1 -> V3
	v3DirectEvent, err := vm.ConvertToVersion(v1Event, 3)
	if err != nil {
		log.Printf("❌ Failed to directly upcast V1 to V3: %v", err)
		return
	}

	fmt.Printf("\n   Direct V1 -> V3 Upcast:\n")
	printEventDetails(v3DirectEvent)
}

// demoDowncasting demonstrates converting events to lower versions
func demoDowncasting(vm *versioning.UserEventVersionManager) {
	// Create a V3 event with rich data
	v3Event := domain.NewUserCreatedV3(
		"user-downcast",
		domain.PersonalInfo{
			FullName: domain.FullName{
				Prefix:     "Dr.",
				FirstName:  "Sarah",
				MiddleName: "Elizabeth",
				LastName:   "Wilson",
				Suffix:     "PhD",
			},
			DateOfBirth: time.Date(1985, 5, 15, 0, 0, 0, 0, time.UTC),
			Gender:      "female",
			Nationality: "US",
		},
		domain.ContactInfo{
			PrimaryEmail:   "sarah.wilson@university.edu",
			SecondaryEmail: "sarah@personal.com",
			PhoneNumbers: []domain.PhoneNumber{
				{Type: "mobile", Number: "+1-555-0123", IsPrimary: true},
			},
		},
		domain.UserPreferences{
			Language: "en",
			Theme:    "dark",
		},
		domain.EventMetadata{
			Version: "3.0",
			Source:  "admin_panel",
		},
	)

	fmt.Printf("   Original V3 Event:\n")
	printEventDetails(v3Event)

	// Downcast V3 -> V2
	v2Event, err := vm.ConvertToVersion(v3Event, 2)
	if err != nil {
		log.Printf("❌ Failed to downcast V3 to V2: %v", err)
		return
	}

	fmt.Printf("\n   Downcasted to V2:\n")
	printEventDetails(v2Event)

	// Downcast V2 -> V1
	v1Event, err := vm.ConvertToVersion(v2Event, 1)
	if err != nil {
		log.Printf("❌ Failed to downcast V2 to V1: %v", err)
		return
	}

	fmt.Printf("\n   Downcasted to V1:\n")
	printEventDetails(v1Event)

	// Direct downcast V3 -> V1
	v1DirectEvent, err := vm.ConvertToVersion(v3Event, 1)
	if err != nil {
		log.Printf("❌ Failed to directly downcast V3 to V1: %v", err)
		return
	}

	fmt.Printf("\n   Direct V3 -> V1 Downcast:\n")
	printEventDetails(v1DirectEvent)
}

// demoVersionCompatibility demonstrates version compatibility checks
func demoVersionCompatibility(vm *versioning.UserEventVersionManager) {
	fmt.Printf("   Supported Versions: %v\n", vm.GetSupportedVersions())
	fmt.Printf("   Latest Version: %d\n", vm.GetLatestVersion())

	// Test version compatibility
	versions := []int{1, 2, 3, 4}
	for _, version := range versions {
		supported := vm.IsVersionSupported(version)
		status := "❌"
		if supported {
			status = "✅"
		}
		fmt.Printf("   Version %d supported: %s\n", version, status)

		if supported {
			info, err := vm.GetVersionInfo(version)
			if err == nil {
				fmt.Printf("     - Events: %v\n", info.SupportedEvents)
				fmt.Printf("     - Can upcast to: %v\n", info.CanUpcastTo)
				fmt.Printf("     - Can downcast to: %v\n", info.CanDowncastTo)
				fmt.Printf("     - Is latest: %t\n", info.IsLatest)
			}
		}
	}
}

// printEventDetails prints detailed information about an event
func printEventDetails(event cqrs.EventMessage) {
	fmt.Printf("     Type: %s\n", event.EventType())
	fmt.Printf("     ID: %s\n", event.EventID())
	fmt.Printf("     Aggregate: %s\n", event.AggregateID())
	fmt.Printf("     Metadata: %v\n", event.Metadata())

	eventData := event.EventData()
	if data, ok := eventData.(map[string]interface{}); ok {
		fmt.Printf("     Data keys: %v\n", getMapKeys(data))

		// Show some sample data
		if userID, exists := data["user_id"]; exists {
			fmt.Printf("     User ID: %v\n", userID)
		}
		if name, exists := data["name"]; exists {
			fmt.Printf("     Name: %v\n", name)
		}
		if email, exists := data["email"]; exists {
			fmt.Printf("     Email: %v\n", email)
		}
	}
}

// getMapKeys returns the keys of a map
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// marshalEventData marshals event data to bytes (simplified)
func marshalEventData(data interface{}) ([]byte, error) {
	// In a real implementation, this would use proper JSON marshaling
	// For demo purposes, we'll create a simple representation
	if dataMap, ok := data.(map[string]interface{}); ok {
		result := "{"
		first := true
		for key, value := range dataMap {
			if !first {
				result += ","
			}
			result += fmt.Sprintf(`"%s":"%v"`, key, value)
			first = false
		}
		result += "}"
		return []byte(result), nil
	}
	return []byte("{}"), nil
}
