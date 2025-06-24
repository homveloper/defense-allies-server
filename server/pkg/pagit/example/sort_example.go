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

func runSortExample() {
	// Setup Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx := context.Background()

	// Clear previous data
	client.Del(ctx, "users:sorted:*", "users:*")

	// Seed data for sorting examples
	seedSortUsers(ctx, client)

	fmt.Println("=== Sorting Examples ===")

	// Example 1: Default sorting (ID DESC)
	fmt.Println("1. Default sorting (ID DESC):")
	adapter := pagitredis.NewOffsetAdapter(client, "users:sorted:id", pagitredis.UnmarshalJSON[User])
	req := pagit.OffsetRequest{Page: 1, PageSize: 5}
	resp, err := pagit.Paginate(ctx, req, adapter)
	if err != nil {
		log.Fatal(err)
	}
	printUsers(resp.Items)

	// Example 2: Sort by name ascending
	fmt.Println("\n2. Sort by name ascending:")
	nameSort := pagit.NewSort().Asc("name")
	adapter = pagitredis.NewOffsetAdapterWithSort(client, "users:sorted:name", nameSort, pagitredis.UnmarshalJSON[User])
	resp, err = pagit.Paginate(ctx, req, adapter)
	if err != nil {
		log.Fatal(err)
	}
	printUsers(resp.Items)

	// Example 3: Sort by created_at descending (latest first)
	fmt.Println("\n3. Sort by created_at descending:")
	timeSort := pagit.NewSort().Desc("created_at")
	adapter = pagitredis.NewOffsetAdapterWithSort(client, "users:sorted:created_at", timeSort, pagitredis.UnmarshalJSON[User])
	resp, err = pagit.Paginate(ctx, req, adapter)
	if err != nil {
		log.Fatal(err)
	}
	printUsers(resp.Items)

	// Example 4: Multi-field adapter with dynamic sorting
	fmt.Println("\n4. Multi-field adapter - sort by email ascending:")
	emailSort := pagit.NewSort().Asc("email")
	multiAdapter := pagitredis.NewMultiFieldOffsetAdapter(client, "users:sorted:", emailSort, pagitredis.UnmarshalJSON[User])
	resp, err = pagit.Paginate(ctx, req, multiAdapter)
	if err != nil {
		log.Fatal(err)
	}
	printUsers(resp.Items)

	// Example 5: Change sort on existing adapter
	fmt.Println("\n5. Change sort to ID ascending:")
	idAscSort := pagit.NewSort().Asc("id")
	newAdapter := multiAdapter.WithSort(idAscSort)
	resp, err = pagit.Paginate(ctx, req, newAdapter)
	if err != nil {
		log.Fatal(err)
	}
	printUsers(resp.Items)

	// Example 6: Cursor-based with sorting
	fmt.Println("\n6. Cursor-based pagination with name descending:")
	nameDescSort := pagit.NewSort().Desc("name")
	cursorAdapter := pagitredis.NewCursorAdapterWithSort(client, "users:sorted:name", nameDescSort, pagitredis.UnmarshalJSON[User])
	cursorReq := pagit.CursorRequest{PageSize: 3}
	cursorResp, err := pagit.PaginateCursor(ctx, cursorReq, cursorAdapter)
	if err != nil {
		log.Fatal(err)
	}
	printUsers(cursorResp.Items)
}

func seedSortUsers(ctx context.Context, client *redis.Client) {
	users := []User{
		{ID: "001", Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now().Add(-5 * time.Hour)},
		{ID: "002", Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now().Add(-4 * time.Hour)},
		{ID: "003", Name: "Charlie", Email: "charlie@example.com", CreatedAt: time.Now().Add(-3 * time.Hour)},
		{ID: "004", Name: "Diana", Email: "diana@example.com", CreatedAt: time.Now().Add(-2 * time.Hour)},
		{ID: "005", Name: "Eve", Email: "eve@example.com", CreatedAt: time.Now().Add(-1 * time.Hour)},
		{ID: "006", Name: "Frank", Email: "frank@example.com", CreatedAt: time.Now().Add(-30 * time.Minute)},
		{ID: "007", Name: "Grace", Email: "grace@example.com", CreatedAt: time.Now().Add(-15 * time.Minute)},
		{ID: "008", Name: "Henry", Email: "henry@example.com", CreatedAt: time.Now().Add(-5 * time.Minute)},
	}

	for _, user := range users {
		data, _ := json.Marshal(user)

		// Create sorted sets for different fields
		// ID sorted set
		client.ZAdd(ctx, "users:sorted:id", redis.Z{
			Score:  float64(user.CreatedAt.Unix()), // Use timestamp as score for ID sorting
			Member: string(data),
		})

		// Name sorted set (alphabetical)
		nameScore := float64(user.Name[0]) // Simple alphabetical scoring
		client.ZAdd(ctx, "users:sorted:name", redis.Z{
			Score:  nameScore,
			Member: string(data),
		})

		// Created_at sorted set
		client.ZAdd(ctx, "users:sorted:created_at", redis.Z{
			Score:  float64(user.CreatedAt.Unix()),
			Member: string(data),
		})

		// Email sorted set (alphabetical)
		emailScore := float64(user.Email[0]) // Simple alphabetical scoring
		client.ZAdd(ctx, "users:sorted:email", redis.Z{
			Score:  emailScore,
			Member: string(data),
		})
	}
}

func printUsers(users []User) {
	for _, user := range users {
		fmt.Printf("  - %s: %s (%s) [%s]\n",
			user.ID, user.Name, user.Email, user.CreatedAt.Format("15:04:05"))
	}
}
