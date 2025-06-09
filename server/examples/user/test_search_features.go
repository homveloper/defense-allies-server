package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cqrs"
	"defense-allies-server/examples/user/domain"
	"defense-allies-server/examples/user/projections"
	"defense-allies-server/examples/user/queries"

	"github.com/google/uuid"
)

// RunSearchFeaturesTest runs the search features test
func RunSearchFeaturesTest() {
	fmt.Println("🔍 Testing User Search and Filtering Features")
	fmt.Println("==============================================")

	ctx := context.Background()

	// Setup in-memory infrastructure
	readStore := cqrs.NewInMemoryReadStore()
	queryDispatcher := cqrs.NewInMemoryQueryDispatcher()

	// Register query handler
	searchHandler := queries.NewUserSearchHandler(readStore)
	queryDispatcher.RegisterHandler("SearchUsers", searchHandler)
	queryDispatcher.RegisterHandler("GetUserByID", searchHandler)
	queryDispatcher.RegisterHandler("GetUsersByRole", searchHandler)

	// Create test users with different profiles and roles
	testUsers := createTestUsers()

	// Create projections and populate read store
	fmt.Println("📝 Creating test users and projections...")
	for i, user := range testUsers {
		userView := createUserViewFromAggregate(user)

		// Add some variety to the test data
		switch i {
		case 0: // Alice - Admin from San Francisco
			userView.FirstName = "Alice"
			userView.LastName = "Smith"
			userView.DisplayName = "Alice S."
			userView.City = "San Francisco"
			userView.Country = "USA"
			userView.Bio = "Software engineer and team lead"
			userView.Roles = []string{"user", "admin"}
			userView.Permissions = []string{"*"}

		case 1: // Bob - Player from New York
			userView.FirstName = "Bob"
			userView.LastName = "Johnson"
			userView.DisplayName = "BobJ"
			userView.City = "New York"
			userView.Country = "USA"
			userView.Bio = "Passionate gamer and beta tester"
			userView.Roles = []string{"user", "player", "beta_tester"}
			userView.Permissions = []string{"user.read", "game.play", "game.test"}

		case 2: // Carol - Moderator from London
			userView.FirstName = "Carol"
			userView.LastName = "Williams"
			userView.DisplayName = "Carol W."
			userView.City = "London"
			userView.Country = "UK"
			userView.Bio = "Community moderator and content manager"
			userView.Roles = []string{"user", "moderator"}
			userView.Permissions = []string{"user.read", "content.moderate"}

		case 3: // David - Developer from Tokyo
			userView.FirstName = "David"
			userView.LastName = "Brown"
			userView.DisplayName = "DevDavid"
			userView.City = "Tokyo"
			userView.Country = "Japan"
			userView.Bio = "Game developer and system architect"
			userView.Roles = []string{"user", "developer"}
			userView.Permissions = []string{"user.*", "game.*", "system.*"}

		case 4: // Eve - Support from Berlin
			userView.FirstName = "Eve"
			userView.LastName = "Davis"
			userView.DisplayName = "EveSupport"
			userView.City = "Berlin"
			userView.Country = "Germany"
			userView.Bio = "Customer support specialist"
			userView.Roles = []string{"user", "support"}
			userView.Permissions = []string{"user.read", "ticket.manage"}
		}

		// Update searchable text
		userView.UpdateSearchableText()

		// Save to read store
		err := readStore.Save(ctx, userView)
		if err != nil {
			log.Fatalf("Failed to save user view: %v", err)
		}

		fmt.Printf("   ✅ Created %s (%s) - %s from %s\n",
			userView.DisplayName, userView.Email, userView.Roles[len(userView.Roles)-1], userView.City)
	}

	// Test 1: Basic text search
	fmt.Println("\n🔍 Test 1: Basic Text Search")
	fmt.Println("============================")

	searchQuery := queries.NewSearchUsersQuery().
		WithSearchText("alice").
		WithPagination(0, 10)

	result, err := queryDispatcher.Dispatch(ctx, searchQuery)
	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	searchResult := result.Data.(*queries.UserSearchResult)
	fmt.Printf("🔍 Search for 'alice': Found %d users\n", searchResult.TotalCount)
	for _, user := range searchResult.Users {
		fmt.Printf("   - %s (%s) from %s\n", user.DisplayName, user.Email, user.City)
	}

	// Test 2: Role-based filtering
	fmt.Println("\n👥 Test 2: Role-based Filtering")
	fmt.Println("===============================")

	roleQuery := queries.NewGetUsersByRoleQuery("admin").
		WithPagination(0, 10)

	result, err = queryDispatcher.Dispatch(ctx, roleQuery)
	if err != nil {
		log.Fatalf("Role query failed: %v", err)
	}

	roleResult := result.Data.(*queries.UserSearchResult)
	fmt.Printf("👑 Users with 'admin' role: Found %d users\n", roleResult.TotalCount)
	for _, user := range roleResult.Users {
		fmt.Printf("   - %s (%s) - Roles: %v\n", user.DisplayName, user.Email, user.Roles)
	}

	// Test 3: Location-based filtering
	fmt.Println("\n🌍 Test 3: Location-based Filtering")
	fmt.Println("===================================")

	locationQuery := queries.NewSearchUsersQuery().
		WithCountries("USA").
		WithPagination(0, 10)

	result, err = queryDispatcher.Dispatch(ctx, locationQuery)
	if err != nil {
		log.Fatalf("Location query failed: %v", err)
	}

	locationResult := result.Data.(*queries.UserSearchResult)
	fmt.Printf("🇺🇸 Users from USA: Found %d users\n", locationResult.TotalCount)
	for _, user := range locationResult.Users {
		fmt.Printf("   - %s from %s, %s\n", user.DisplayName, user.City, user.Country)
	}

	// Test 4: Multiple role filtering
	fmt.Println("\n🎭 Test 4: Multiple Role Filtering")
	fmt.Println("==================================")

	multiRoleQuery := queries.NewSearchUsersQuery().
		WithRoles("player", "beta_tester").
		WithPagination(0, 10)

	result, err = queryDispatcher.Dispatch(ctx, multiRoleQuery)
	if err != nil {
		log.Fatalf("Multi-role query failed: %v", err)
	}

	multiRoleResult := result.Data.(*queries.UserSearchResult)
	fmt.Printf("🎮 Users with 'player' or 'beta_tester' roles: Found %d users\n", multiRoleResult.TotalCount)
	for _, user := range multiRoleResult.Users {
		fmt.Printf("   - %s - Roles: %v\n", user.DisplayName, user.Roles)
	}

	// Test 5: Complex search with multiple filters
	fmt.Println("\n🔧 Test 5: Complex Multi-Filter Search")
	fmt.Println("=====================================")

	complexQuery := queries.NewSearchUsersQuery().
		WithSearchText("developer").
		WithStatus("active").
		WithCountries("Japan", "Germany").
		WithSorting("display_name", "asc").
		WithPagination(0, 10)

	result, err = queryDispatcher.Dispatch(ctx, complexQuery)
	if err != nil {
		log.Fatalf("Complex query failed: %v", err)
	}

	complexResult := result.Data.(*queries.UserSearchResult)
	fmt.Printf("🔍 Complex search (text:'developer', status:'active', countries:['Japan','Germany']): Found %d users\n", complexResult.TotalCount)
	for _, user := range complexResult.Users {
		fmt.Printf("   - %s (%s) from %s, %s - Roles: %v\n",
			user.DisplayName, user.Email, user.City, user.Country, user.Roles)
	}

	// Test 6: Get user by ID
	fmt.Println("\n🆔 Test 6: Get User by ID")
	fmt.Println("=========================")

	if len(testUsers) > 0 {
		userID := testUsers[0].ID()
		getUserQuery := queries.NewGetUserByIDQuery(userID)

		result, err = queryDispatcher.Dispatch(ctx, getUserQuery)
		if err != nil {
			log.Fatalf("Get user by ID failed: %v", err)
		}

		userView := result.Data.(*projections.UserView)
		fmt.Printf("👤 User details for ID %s:\n", userID)
		fmt.Printf("   Name: %s\n", userView.GetFullName())
		fmt.Printf("   Email: %s\n", userView.Email)
		fmt.Printf("   Display Name: %s\n", userView.DisplayName)
		fmt.Printf("   Bio: %s\n", userView.Bio)
		fmt.Printf("   Location: %s, %s\n", userView.City, userView.Country)
		fmt.Printf("   Roles: %v\n", userView.Roles)
		fmt.Printf("   Status: %s\n", userView.Status)
	}

	// Test 7: Pagination test
	fmt.Println("\n📄 Test 7: Pagination")
	fmt.Println("=====================")

	// Get first page
	pageQuery := queries.NewSearchUsersQuery().
		WithPagination(0, 2).
		WithSorting("display_name", "asc")

	result, err = queryDispatcher.Dispatch(ctx, pageQuery)
	if err != nil {
		log.Fatalf("Pagination query failed: %v", err)
	}

	pageResult := result.Data.(*queries.UserSearchResult)
	fmt.Printf("📄 Page 1 (limit: 2): Found %d total users, showing %d\n",
		pageResult.TotalCount, len(pageResult.Users))
	for i, user := range pageResult.Users {
		fmt.Printf("   %d. %s (%s)\n", i+1, user.DisplayName, user.Email)
	}
	fmt.Printf("   Has more: %t\n", pageResult.HasMore)

	// Get second page
	pageQuery2 := queries.NewSearchUsersQuery().
		WithPagination(2, 2).
		WithSorting("display_name", "asc")

	result, err = queryDispatcher.Dispatch(ctx, pageQuery2)
	if err != nil {
		log.Fatalf("Pagination query 2 failed: %v", err)
	}

	pageResult2 := result.Data.(*queries.UserSearchResult)
	fmt.Printf("📄 Page 2 (limit: 2): Showing %d users\n", len(pageResult2.Users))
	for i, user := range pageResult2.Users {
		fmt.Printf("   %d. %s (%s)\n", i+3, user.DisplayName, user.Email)
	}
	fmt.Printf("   Has more: %t\n", pageResult2.HasMore)

	fmt.Println("\n🎉 All search and filtering tests completed successfully!")
}

// createTestUsers creates test users for demonstration
func createTestUsers() []*domain.User {
	var users []*domain.User

	emails := []string{
		"alice.smith@example.com",
		"bob.johnson@example.com",
		"carol.williams@example.com",
		"david.brown@example.com",
		"eve.davis@example.com",
	}

	names := []string{
		"Alice Smith",
		"Bob Johnson",
		"Carol Williams",
		"David Brown",
		"Eve Davis",
	}

	for i := 0; i < len(emails); i++ {
		userID := uuid.New().String()
		user, err := domain.NewUser(userID, emails[i], names[i])
		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		users = append(users, user)
	}

	return users
}

// createUserViewFromAggregate creates a UserView from a User aggregate
func createUserViewFromAggregate(user *domain.User) *projections.UserView {
	userView := projections.NewUserView(user.ID())
	userView.Email = user.Email()
	userView.Name = user.Name()
	userView.Status = user.Status().String()
	userView.CreatedAt = time.Now()
	userView.UpdatedAt = time.Now()

	if user.LastLoginAt() != nil {
		userView.LastLoginAt = user.LastLoginAt()
	}

	if user.DeactivatedAt() != nil {
		userView.DeactivatedAt = user.DeactivatedAt()
		userView.DeactivationReason = user.DeactivationReason()
	}

	return userView
}
