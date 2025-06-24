package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/defense-allies/pagit"
	pagitredis "github.com/defense-allies/pagit/adapters/pagit-redis"
)

// User represents a simple user model
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func runOffsetExample() {
	// Setup Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	// Clear previous data
	client.Del(ctx, "users:offset", "users:*")

	// Seed some data
	seedOffsetUsers(ctx, client)

	// Create offset adapter
	adapter := pagitredis.NewOffsetAdapter(client, "users:offset", pagitredis.UnmarshalJSON[User])

	// Example 1: First page
	fmt.Println("=== Offset-based Pagination ===")
	req := pagit.OffsetRequest{
		Page:     1,
		PageSize: 5,
	}

	resp, err := pagit.Paginate(ctx, req, adapter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Page %d/%d (Total: %d items)\n", resp.Page, resp.TotalPages, resp.Total)
	fmt.Printf("Has Next: %v, Has Prev: %v\n", resp.HasNext, resp.HasPrev)
	fmt.Println("Items:")
	for _, user := range resp.Items {
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}

	// Example 2: Second page
	req.Page = 2
	resp, err = pagit.Paginate(ctx, req, adapter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nPage %d/%d\n", resp.Page, resp.TotalPages)
	fmt.Println("Items:")
	for _, user := range resp.Items {
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}

	// Example 3: Custom options
	opts := pagit.OffsetOptions{
		DefaultPageSize: 3,
		MaxPageSize:     10,
		MinPageSize:     1,
	}

	req = pagit.OffsetRequest{Page: 1, PageSize: 15} // Will be capped to 10
	resp, err = pagit.Paginate(ctx, req, adapter, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nWith custom options - Requested: 15, Got: %d (capped)\n", resp.PageSize)
}

func seedOffsetUsers(ctx context.Context, client *redis.Client) {
	// Create 20 users
	for i := 1; i <= 20; i++ {
		user := User{
			ID:        fmt.Sprintf("user-%03d", i),
			Name:      fmt.Sprintf("User %d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Hour),
		}

		// Store in sorted set with timestamp as score
		data, _ := json.Marshal(user)
		score := float64(user.CreatedAt.Unix())
		client.ZAdd(ctx, "users:offset", redis.Z{
			Score:  score,
			Member: string(data),
		})
	}
}
