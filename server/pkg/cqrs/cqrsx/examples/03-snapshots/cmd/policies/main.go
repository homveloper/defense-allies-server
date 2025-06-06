package main

import (
	"fmt"
	"time"

	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/domain"

	"github.com/shopspring/decimal"
)

// Policies demo demonstrates different snapshot policies
func main() {
	fmt.Println("📋 Snapshot Policies Demo")
	fmt.Println("==========================")

	policies := []cqrsx.SnapshotPolicy{
		cqrsx.NewEventCountPolicy(5),
		cqrsx.NewVersionBasedPolicy(3),
		cqrsx.NewTimeBasedPolicy(1 * time.Second),
		cqrsx.NewCompositePolicy("OR",
			cqrsx.NewEventCountPolicy(10),
			cqrsx.NewVersionBasedPolicy(5),
		),
		cqrsx.NewAdaptivePolicy(8, 0.7),
		cqrsx.NewAlwaysPolicy(),
		cqrsx.NewNeverPolicy(),
	}

	order := createSampleOrder()
	addItemToOrder(order, "Test Item", decimal.NewFromInt(100), 1)

	fmt.Printf("Testing policies with order v%d, 5 events:\n\n", order.Version())

	for _, policy := range policies {
		shouldCreate := policy.ShouldCreateSnapshot(order, 5)
		fmt.Printf("%-20s: %v (interval: %d)\n",
			policy.GetPolicyName(), shouldCreate, policy.GetSnapshotInterval())
	}

	fmt.Println("\n" + repeatString("=", 50))
	fmt.Println("Testing Composite Policy Details:")
	fmt.Println(repeatString("=", 50))

	// Test composite policy in detail
	eventCountPolicy := cqrsx.NewEventCountPolicy(3)
	versionPolicy := cqrsx.NewVersionBasedPolicy(2)

	andPolicy := cqrsx.NewCompositePolicy("AND", eventCountPolicy, versionPolicy)
	orPolicy := cqrsx.NewCompositePolicy("OR", eventCountPolicy, versionPolicy)

	fmt.Printf("\nOrder version: %d, Event count: 6\n", order.Version())
	fmt.Printf("EventCountPolicy(3): %v\n", eventCountPolicy.ShouldCreateSnapshot(order, 6))
	fmt.Printf("VersionBasedPolicy(2): %v\n", versionPolicy.ShouldCreateSnapshot(order, 6))
	fmt.Printf("CompositePolicy(AND): %v\n", andPolicy.ShouldCreateSnapshot(order, 6))
	fmt.Printf("CompositePolicy(OR): %v\n", orPolicy.ShouldCreateSnapshot(order, 6))

	fmt.Println("\n" + repeatString("=", 50))
	fmt.Println("Testing Time-Based Policy:")
	fmt.Println(repeatString("=", 50))

	timePolicy := cqrsx.NewTimeBasedPolicy(2 * time.Second)

	fmt.Printf("First check: %v\n", timePolicy.ShouldCreateSnapshot(order, 1))
	fmt.Printf("Immediate second check: %v\n", timePolicy.ShouldCreateSnapshot(order, 1))

	fmt.Println("Waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	fmt.Printf("After 3 seconds: %v\n", timePolicy.ShouldCreateSnapshot(order, 1))

	fmt.Println("\n" + repeatString("=", 50))
	fmt.Println("Testing Adaptive Policy:")
	fmt.Println(repeatString("=", 50))

	adaptivePolicy := cqrsx.NewAdaptivePolicy(5, 0.8)

	// Simulate performance data
	adaptivePolicy.UpdatePerformanceMetrics(order.ID(), 50*time.Millisecond, 10)
	fmt.Printf("With fast restore (50ms): %v\n", adaptivePolicy.ShouldCreateSnapshot(order, 5))

	adaptivePolicy.UpdatePerformanceMetrics(order.ID(), 200*time.Millisecond, 10)
	fmt.Printf("With slow restore (200ms): %v\n", adaptivePolicy.ShouldCreateSnapshot(order, 4))

	fmt.Println("\n✅ Policies demo completed!")
}

// Helper functions

func createSampleOrder() *domain.Order {
	order := domain.NewOrder()
	// Create order with customer ID
	order.CreateOrder(order.ID(), "customer-123", decimal.Zero)
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
