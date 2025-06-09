package main

import (
	"fmt"

	"cqrs/cqrsx"
	"cqrs/cqrsx/examples/03-snapshots/domain"

	"github.com/shopspring/decimal"
)

// Serializers demo demonstrates different serializers
func main() {
	fmt.Println("🔧 Snapshot Serializers Demo")
	fmt.Println("=============================")

	factory := cqrsx.NewSnapshotSerializerFactory()

	serializers := []struct {
		name        string
		serType     string
		compression string
		options     map[string]interface{}
	}{
		{"JSON", "json", "none", map[string]interface{}{"pretty_print": false}},
		{"Pretty JSON", "json", "none", map[string]interface{}{"pretty_print": true}},
		{"Compressed JSON", "json", "gzip", map[string]interface{}{"pretty_print": false}},
		{"BSON", "bson", "none", map[string]interface{}{}},
		{"Compressed BSON", "bson", "gzip", map[string]interface{}{}},
	}

	order := createSampleOrder()
	addItemToOrder(order, "Serialization Test", decimal.NewFromInt(200), 1)

	fmt.Printf("Testing serializers with order v%d:\n\n", order.Version())

	for _, config := range serializers {
		serializer, err := factory.CreateSerializer(config.serType, config.compression, config.options)
		if err != nil {
			fmt.Printf("%-20s: ERROR - %v\n", config.name, err)
			continue
		}

		data, err := serializer.SerializeSnapshot(order)
		if err != nil {
			fmt.Printf("%-20s: SERIALIZE ERROR - %v\n", config.name, err)
			continue
		}

		fmt.Printf("%-20s: %d bytes (%s, %s)\n",
			config.name, len(data), serializer.GetContentType(), serializer.GetCompressionType())
	}

	fmt.Println("\n" + repeatString("=", 50))
	fmt.Println("Detailed Serialization Comparison:")
	fmt.Println(repeatString("=", 50))

	// Create a more complex order for better comparison
	complexOrder := createComplexOrder()

	jsonSerializer := cqrsx.NewJSONSnapshotSerializer(false)
	prettyJsonSerializer := cqrsx.NewJSONSnapshotSerializer(true)
	compressedJsonSerializer := cqrsx.NewCompressedJSONSnapshotSerializer("gzip", false)

	// Test JSON serialization
	jsonData, _ := jsonSerializer.SerializeSnapshot(complexOrder)
	prettyJsonData, _ := prettyJsonSerializer.SerializeSnapshot(complexOrder)
	compressedJsonData, _ := compressedJsonSerializer.SerializeSnapshot(complexOrder)

	fmt.Printf("\nComplex Order Serialization Results:\n")
	fmt.Printf("%-20s: %d bytes\n", "JSON (compact)", len(jsonData))
	fmt.Printf("%-20s: %d bytes\n", "JSON (pretty)", len(prettyJsonData))
	fmt.Printf("%-20s: %d bytes\n", "JSON (compressed)", len(compressedJsonData))

	compressionRatio := float64(len(compressedJsonData)) / float64(len(jsonData)) * 100
	fmt.Printf("Compression ratio: %.1f%%\n", compressionRatio)

	fmt.Println("\n" + repeatString("=", 50))
	fmt.Println("Serializer Features:")
	fmt.Println(repeatString("=", 50))

	features := map[string][]string{
		"JSON":            {"Human readable", "Wide support", "Larger size"},
		"Pretty JSON":     {"Human readable", "Formatted", "Largest size"},
		"Compressed JSON": {"Smaller size", "Faster transfer", "CPU overhead"},
		"BSON":            {"Binary format", "Type preservation", "MongoDB native"},
		"Compressed BSON": {"Smallest size", "Binary + compression", "Highest CPU overhead"},
	}

	for serializerName, featureList := range features {
		fmt.Printf("\n%s:\n", serializerName)
		for _, feature := range featureList {
			fmt.Printf("  • %s\n", feature)
		}
	}

	fmt.Println("\n✅ Serializers demo completed!")
}

// Helper functions

func createSampleOrder() *domain.Order {
	order := domain.NewOrder()
	// Create order with customer ID
	order.CreateOrder(order.ID(), "customer-123", decimal.Zero)
	return order
}

func createComplexOrder() *domain.Order {
	order := domain.NewOrder()
	order.CreateOrder(order.ID(), "customer-complex-test", decimal.NewFromInt(15))

	// Add multiple items
	order.AddItem("laptop-001", "Gaming Laptop", 1, decimal.NewFromInt(1500))
	order.AddItem("mouse-001", "Gaming Mouse", 2, decimal.NewFromInt(75))
	order.AddItem("keyboard-001", "Mechanical Keyboard", 1, decimal.NewFromInt(120))
	order.AddItem("monitor-001", "4K Monitor", 2, decimal.NewFromInt(400))
	order.AddItem("headset-001", "Gaming Headset", 1, decimal.NewFromInt(200))

	// Apply discount
	order.ApplyDiscount(decimal.NewFromFloat(0.15), "Bulk purchase discount")

	// Confirm order
	order.ConfirmOrder()

	return order
}

func addItemToOrder(order *domain.Order, name string, price decimal.Decimal, quantity int) {
	order.AddItem(fmt.Sprintf("prod-%s", name), name, quantity, price)
}

// String repeat helper
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
