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

func runCursorExample() {
	// Setup Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	// Clear previous data
	client.Del(ctx, "users:cursor", "users:*")

	// Seed some data
	seedCursorUsers(ctx, client)

	// Create cursor adapter
	adapter := pagitredis.NewCursorAdapter(client, "users:cursor", pagitredis.UnmarshalJSON[User])

	// Example 1: First page
	fmt.Println("=== Cursor-based Pagination ===")
	req := pagit.CursorRequest{
		PageSize: 5,
	}

	resp, err := pagit.PaginateCursor(ctx, req, adapter)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retrieved %d items\n", len(resp.Items))
	fmt.Printf("Has Next: %v\n", resp.HasNext)
	fmt.Printf("Next Cursor: %s\n", resp.NextCursor)
	fmt.Println("Items:")
	for _, user := range resp.Items {
		fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
	}

	// Example 2: Next page using cursor
	if resp.HasNext {
		req.Cursor = resp.NextCursor
		resp, err = pagit.PaginateCursor(ctx, req, adapter)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\nNext page: %d items\n", len(resp.Items))
		fmt.Printf("Has Next: %v\n", resp.HasNext)
		fmt.Println("Items:")
		for _, user := range resp.Items {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	}

	// Example 3: Custom options
	opts := pagit.CursorOptions{
		DefaultPageSize: 3,
		MaxPageSize:     8,
		MinPageSize:     1,
	}

	req = pagit.CursorRequest{PageSize: 12} // Will be capped to 8
	resp, err = pagit.PaginateCursor(ctx, req, adapter, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nWith custom options - Requested: 12, Got: %d (capped)\n", resp.PageSize)
}

func seedCursorUsers(ctx context.Context, client *redis.Client) {
	// Create 15 users
	for i := 1; i <= 15; i++ {
		user := User{
			ID:        fmt.Sprintf("cursor-user-%03d", i),
			Name:      fmt.Sprintf("Cursor User %d", i),
			Email:     fmt.Sprintf("cursor%d@example.com", i),
			CreatedAt: time.Now().Add(-time.Duration(i) * time.Minute),
		}

		// Store in sorted set with timestamp as score
		data, _ := json.Marshal(user)
		score := float64(user.CreatedAt.Unix())
		client.ZAdd(ctx, "users:cursor", redis.Z{
			Score:  score,
			Member: string(data),
		})
	}
}