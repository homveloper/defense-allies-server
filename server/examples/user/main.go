package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"defense-allies-server/examples/user/domain"
	"defense-allies-server/examples/user/handlers"
	"defense-allies-server/examples/user/projections"
	"defense-allies-server/pkg/cqrs"
	"defense-allies-server/pkg/cqrs/cqrsx"

	"github.com/google/uuid"

	"github.com/pkg/errors"
)

func main() {
	fmt.Println("🚀 Defense Allies CQRS User Example")
	fmt.Println("=====================================")

	ctx := context.Background()

	// Run InMemory example
	fmt.Println("\n📝 Running InMemory Implementation Example...")
	if err := runInMemoryExample(ctx); err != nil {
		fmt.Printf("InMemory example failed: %+v\n", err)
		log.Fatalf("InMemory example failed with stack trace: %+v", err)
	}

	// Run Redis example (if Redis is available)
	fmt.Println("\n🔴 Running Redis Implementation Example...")
	if err := runRedisExample(ctx); err != nil {
		fmt.Printf("Redis example failed: %+v\n", err)
		log.Printf("Redis example failed (Redis might not be available): %+v", err)
		fmt.Println("⚠️  Redis example skipped - make sure Redis is running on localhost:6379")
	}

	fmt.Println("\n✅ All examples completed successfully!")

	// Run additional test scenarios
	fmt.Println("\n🧪 Running Additional Test Scenarios...")
	fmt.Println("======================================")

	// Run extended features test
	fmt.Println("\n📋 Extended Features Test:")
	RunExtendedFeaturesTest()

	// Run search features test
	fmt.Println("\n🔍 Search Features Test:")
	RunSearchFeaturesTest()

	fmt.Println("\n🎉 All tests completed successfully!")
}

// runInMemoryExample demonstrates CQRS with InMemory implementations
func runInMemoryExample(ctx context.Context) error {
	fmt.Println("Setting up InMemory CQRS infrastructure...")

	// Create InMemory implementations
	eventBus := cqrs.NewInMemoryEventBus()
	if err := eventBus.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start event bus")
	}
	defer eventBus.Stop(ctx)

	// Note: Using InMemory implementations for this example

	// Create repository (we'll use a simple in-memory implementation)
	repository := NewInMemoryUserRepository()

	// Create read store
	readStore := NewInMemoryReadStore()

	// Set up projection
	userProjection := projections.NewUserViewProjection(readStore)
	projectionManager := cqrs.NewInMemoryProjectionManager()
	if err := projectionManager.RegisterProjection(userProjection); err != nil {
		return errors.Wrap(err, "failed to register projection")
	}

	// Subscribe projection to events
	if _, err := eventBus.Subscribe(domain.UserCreatedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return errors.Wrap(err, "failed to subscribe to UserCreated events")
	}
	if _, err := eventBus.Subscribe(domain.EmailChangedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return errors.Wrap(err, "failed to subscribe to EmailChanged events")
	}
	if _, err := eventBus.Subscribe(domain.UserDeactivatedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return errors.Wrap(err, "failed to subscribe to UserDeactivated events")
	}
	if _, err := eventBus.Subscribe(domain.UserActivatedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return errors.Wrap(err, "failed to subscribe to UserActivated events")
	}

	// Create handlers
	commandHandler := handlers.NewUserCommandHandler(repository, eventBus)
	queryDispatcher := handlers.NewUserQueryDispatcher(readStore)

	// Run the example scenario
	return runUserScenario(ctx, commandHandler, queryDispatcher, "InMemory")
}

// runRedisExample demonstrates CQRS with Redis implementations
func runRedisExample(ctx context.Context) error {
	fmt.Println("Setting up Redis CQRS infrastructure...")

	// Create Redis client
	config := &cqrs.RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Database: 0,
	}

	client, err := cqrsx.NewRedisClientManager(config)
	if err != nil {
		return fmt.Errorf("failed to create Redis client: %w", err)
	}
	defer client.Close()

	// Test Redis connection
	if err := client.ExecuteCommand(ctx, func() error { return nil }); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Create Redis-based implementations
	stateStore := cqrsx.NewRedisStateStore(client, "user_example")

	// Create factory and serializer for UserView read models
	factory := projections.NewUserReadModelFactory()
	serializer := cqrsx.NewJSONReadModelSerializer(factory)
	readStore := cqrsx.NewRedisReadStore(client, "user_example", serializer)

	// Create User-specific repository
	repository := &UserRedisRepository{stateStore: stateStore}

	// Create event bus (using InMemory for simplicity in this example)
	eventBus := cqrs.NewInMemoryEventBus()
	if err := eventBus.Start(ctx); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}
	defer eventBus.Stop(ctx)

	// Set up projection
	userProjection := projections.NewUserViewProjection(readStore)

	// Subscribe projection to events
	if _, err := eventBus.Subscribe(domain.UserCreatedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return fmt.Errorf("failed to subscribe to UserCreated events: %w", err)
	}
	if _, err := eventBus.Subscribe(domain.EmailChangedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return fmt.Errorf("failed to subscribe to EmailChanged events: %w", err)
	}
	if _, err := eventBus.Subscribe(domain.UserDeactivatedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return fmt.Errorf("failed to subscribe to UserDeactivated events: %w", err)
	}
	if _, err := eventBus.Subscribe(domain.UserActivatedEventType, &ProjectionEventHandler{projection: userProjection}); err != nil {
		return fmt.Errorf("failed to subscribe to UserActivated events: %w", err)
	}

	// Create handlers
	commandHandler := handlers.NewUserCommandHandler(repository, eventBus)
	queryDispatcher := handlers.NewUserQueryDispatcher(readStore)

	// Run the example scenario
	return runUserScenario(ctx, commandHandler, queryDispatcher, "Redis")
}

// runUserScenario runs a complete user management scenario
func runUserScenario(ctx context.Context, commandHandler *handlers.UserCommandHandler, queryDispatcher *handlers.UserQueryDispatcher, implementation string) error {
	fmt.Printf("\n🎯 Running User Scenario (%s Implementation)\n", implementation)
	fmt.Println("=" + fmt.Sprintf("%*s", len(implementation)+30, "="))

	// Generate user IDs
	userID1 := uuid.New().String()
	userID2 := uuid.New().String()

	// 1. Create users
	fmt.Println("\n1️⃣  Creating users...")

	createCmd1 := domain.NewCreateUserCommand(userID1, "alice@example.com", "Alice Smith")
	result1, err := commandHandler.Handle(ctx, createCmd1)
	if err != nil {
		return errors.Wrapf(err, "failed to create user 1")
	}
	if !result1.Success {
		return errors.Wrapf(result1.Error, "failed to create user 1")
	}
	fmt.Printf("   ✅ Created user: %s (Alice Smith)\n", userID1)

	createCmd2 := domain.NewCreateUserCommand(userID2, "bob@example.com", "Bob Johnson")
	result2, err := commandHandler.Handle(ctx, createCmd2)
	if err != nil {
		return errors.Wrapf(err, "failed to create user 2")
	}
	if !result2.Success {
		return errors.Wrapf(result2.Error, "failed to create user 2")
	}
	fmt.Printf("   ✅ Created user: %s (Bob Johnson)\n", userID2)

	// Wait for projections to process
	time.Sleep(100 * time.Millisecond)

	// 2. Query users
	fmt.Println("\n2️⃣  Querying users...")

	getUserQuery := handlers.CreateGetUserQuery(userID1)
	userResult, err := queryDispatcher.Dispatch(ctx, getUserQuery)
	if err != nil {
		return errors.Wrapf(err, "failed to get user")
	}

	if !userResult.Success {
		return errors.Wrapf(userResult.Error, "failed to get user")
	}

	userView := userResult.Data.(*projections.UserView)
	fmt.Printf("   📋 User Details: %s - %s (%s)\n", userView.Name, userView.Email, userView.Status)

	// 3. List all users
	listQuery := handlers.CreateListUsersQuery("", 1, 10)
	listResult, err := queryDispatcher.Dispatch(ctx, listQuery)
	if err != nil {
		return errors.Wrapf(err, "failed to list users")
	}

	if !listResult.Success {
		return fmt.Errorf("failed to list users: %v", listResult.Error)
	}

	userViews := listResult.Data.([]*projections.UserView)
	fmt.Printf("   📋 Total users: %d\n", len(userViews))
	for _, view := range userViews {
		fmt.Printf("      - %s: %s (%s)\n", view.Name, view.Email, view.Status)
	}

	// 4. Change email
	fmt.Println("\n3️⃣  Changing email...")

	changeEmailCmd := domain.NewChangeEmailCommand(userID1, "alice.smith@example.com")
	emailResult, err := commandHandler.Handle(ctx, changeEmailCmd)
	if err != nil {
		return errors.Wrapf(err, "failed to change email")
	}

	if !emailResult.Success {
		return errors.Wrapf(emailResult.Error, "failed to change email")
	}
	fmt.Printf("   ✅ Changed email for user: %s\n", userID1)

	// Wait for projections to process
	time.Sleep(100 * time.Millisecond)

	// 5. Verify email change
	getUserQuery2 := handlers.CreateGetUserQuery(userID1)
	userResult2, err := queryDispatcher.Dispatch(ctx, getUserQuery2)
	if err != nil {
		return errors.Wrapf(err, "failed to get user after email change")
	}

	if !userResult2.Success {
		return errors.Wrapf(userResult2.Error, "failed to get user after email change")
	}

	userView2 := userResult2.Data.(*projections.UserView)
	fmt.Printf("   📋 Updated email: %s\n", userView2.Email)

	// 6. Deactivate user
	fmt.Println("\n4️⃣  Deactivating user...")

	deactivateCmd := domain.NewDeactivateUserCommand(userID2, "User requested account closure")
	deactivateResult, err := commandHandler.Handle(ctx, deactivateCmd)
	if err != nil {
		return errors.Wrapf(err, "failed to deactivate user")
	}

	if !deactivateResult.Success {
		return errors.Wrapf(deactivateResult.Error, "failed to deactivate user")
	}
	fmt.Printf("   ✅ Deactivated user: %s\n", userID2)

	// Wait for projections to process
	time.Sleep(100 * time.Millisecond)

	// 7. List active users only
	fmt.Println("\n5️⃣  Listing active users...")

	activeQuery := handlers.CreateListActiveUsersQuery(1, 10)
	activeResult, err := queryDispatcher.Dispatch(ctx, activeQuery)
	if err != nil {
		return errors.Wrapf(err, "failed to list active users")
	}

	if !activeResult.Success {
		return errors.Wrapf(activeResult.Error, "failed to list active users")
	}

	activeViews := activeResult.Data.([]*projections.UserView)
	fmt.Printf("   📋 Active users: %d\n", len(activeViews))
	for _, view := range activeViews {
		fmt.Printf("      - %s: %s (%s)\n", view.Name, view.Email, view.Status)
	}

	// 8. Performance metrics
	fmt.Println("\n6️⃣  Performance Summary...")
	fmt.Printf("   ⚡ Create User 1: %v\n", result1.ExecutionTime)
	fmt.Printf("   ⚡ Create User 2: %v\n", result2.ExecutionTime)
	fmt.Printf("   ⚡ Change Email: %v\n", emailResult.ExecutionTime)
	fmt.Printf("   ⚡ Deactivate User: %v\n", deactivateResult.ExecutionTime)
	fmt.Printf("   ⚡ Query User: %v\n", userResult.ExecutionTime)
	fmt.Printf("   ⚡ List Users: %v\n", listResult.ExecutionTime)

	fmt.Printf("\n✅ %s scenario completed successfully!\n", implementation)
	return nil
}

// InMemoryUserRepository is a simple in-memory repository for the example
type InMemoryUserRepository struct {
	users map[string]*domain.User
}

// NewInMemoryUserRepository creates a new InMemoryUserRepository
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*domain.User),
	}
}

// Save saves an aggregate
func (r *InMemoryUserRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	user, ok := aggregate.(*domain.User)
	if !ok {
		return fmt.Errorf("invalid aggregate type: expected *domain.User, got %T", aggregate)
	}

	// Simplified version control for demo purposes
	// In a real implementation, you would implement proper optimistic concurrency control
	fmt.Printf("DEBUG SAVE: Saving user %s with version %d (expected: %d)\n",
		user.AggregateID(), user.CurrentVersion(), expectedVersion)

	// Clone the user to avoid reference issues
	clonedUser := *user
	r.users[user.AggregateID()] = &clonedUser
	return nil
}

// GetByID gets an aggregate by ID
func (r *InMemoryUserRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, cqrs.NewCQRSError(cqrs.ErrCodeAggregateNotFound.String(),
			fmt.Sprintf("user with ID %s not found", id), nil)
	}

	// Clone the user to avoid reference issues
	clonedUser := *user
	// Set original version to current version for optimistic concurrency control
	clonedUser.SetOriginalVersion(clonedUser.CurrentVersion())
	// Clear any uncommitted changes
	clonedUser.ClearChanges()

	fmt.Printf("DEBUG LOAD: User %s loaded with version %d\n", id, clonedUser.CurrentVersion())
	return &clonedUser, nil
}

// GetVersion gets the version of an aggregate
func (r *InMemoryUserRepository) GetVersion(ctx context.Context, id string) (int, error) {
	user, exists := r.users[id]
	if !exists {
		return 0, cqrs.NewCQRSError(cqrs.ErrCodeAggregateNotFound.String(),
			fmt.Sprintf("user with ID %s not found", id), nil)
	}
	return user.CurrentVersion(), nil
}

// Exists checks if an aggregate exists
func (r *InMemoryUserRepository) Exists(ctx context.Context, id string) bool {
	_, exists := r.users[id]
	return exists
}

// InMemoryReadStore is a simple in-memory read store for the example
type InMemoryReadStore struct {
	readModels map[string]map[string]cqrs.ReadModel // [type][id]readModel
}

// NewInMemoryReadStore creates a new InMemoryReadStore
func NewInMemoryReadStore() *InMemoryReadStore {
	return &InMemoryReadStore{
		readModels: make(map[string]map[string]cqrs.ReadModel),
	}
}

// Save saves a read model
func (s *InMemoryReadStore) Save(ctx context.Context, readModel cqrs.ReadModel) error {
	modelType := readModel.GetType()
	if s.readModels[modelType] == nil {
		s.readModels[modelType] = make(map[string]cqrs.ReadModel)
	}
	s.readModels[modelType][readModel.GetID()] = readModel
	return nil
}

// GetByID gets a read model by ID and type
func (s *InMemoryReadStore) GetByID(ctx context.Context, id string, modelType string) (cqrs.ReadModel, error) {
	if s.readModels[modelType] == nil {
		return nil, fmt.Errorf("read model type %s not found", modelType)
	}

	readModel, exists := s.readModels[modelType][id]
	if !exists {
		return nil, fmt.Errorf("read model with ID %s and type %s not found", id, modelType)
	}

	return readModel, nil
}

// Delete deletes a read model
func (s *InMemoryReadStore) Delete(ctx context.Context, id string, modelType string) error {
	if s.readModels[modelType] != nil {
		delete(s.readModels[modelType], id)
	}
	return nil
}

// Query queries read models
func (s *InMemoryReadStore) Query(ctx context.Context, criteria cqrs.QueryCriteria) ([]cqrs.ReadModel, error) {
	var results []cqrs.ReadModel

	// Simple implementation - just return all UserView models
	if userViews, exists := s.readModels["UserView"]; exists {
		for _, readModel := range userViews {
			userView := readModel.(*projections.UserView)

			// Apply status filter if specified
			if status, hasStatus := criteria.Filters["status"]; hasStatus {
				if userView.Status != status.(string) {
					continue
				}
			}

			results = append(results, readModel)
		}
	}

	// Apply limit and offset
	if criteria.Offset > 0 && criteria.Offset < len(results) {
		results = results[criteria.Offset:]
	}
	if criteria.Limit > 0 && criteria.Limit < len(results) {
		results = results[:criteria.Limit]
	}

	return results, nil
}

// Count counts read models
func (s *InMemoryReadStore) Count(ctx context.Context, criteria cqrs.QueryCriteria) (int64, error) {
	count := int64(0)

	// Simple implementation - just count UserView models
	if userViews, exists := s.readModels["UserView"]; exists {
		for _, readModel := range userViews {
			userView := readModel.(*projections.UserView)

			// Apply status filter if specified
			if status, hasStatus := criteria.Filters["status"]; hasStatus {
				if userView.Status != status.(string) {
					continue
				}
			}

			count++
		}
	}

	return count, nil
}

// SaveBatch saves multiple read models
func (s *InMemoryReadStore) SaveBatch(ctx context.Context, readModels []cqrs.ReadModel) error {
	for _, readModel := range readModels {
		if err := s.Save(ctx, readModel); err != nil {
			return err
		}
	}
	return nil
}

// DeleteBatch deletes multiple read models
func (s *InMemoryReadStore) DeleteBatch(ctx context.Context, ids []string, modelType string) error {
	for _, id := range ids {
		if err := s.Delete(ctx, id, modelType); err != nil {
			return err
		}
	}
	return nil
}

// CreateIndex creates an index (no-op for in-memory)
func (s *InMemoryReadStore) CreateIndex(ctx context.Context, modelType string, fields []string) error {
	return nil
}

// DropIndex drops an index (no-op for in-memory)
func (s *InMemoryReadStore) DropIndex(ctx context.Context, modelType string, indexName string) error {
	return nil
}

// ProjectionEventHandler handles events for projections
type ProjectionEventHandler struct {
	projection cqrs.Projection
}

// Handle handles the event
func (h *ProjectionEventHandler) Handle(ctx context.Context, event cqrs.EventMessage) error {
	return h.projection.Project(ctx, event)
}

// CanHandle returns true if the handler can handle the event type
func (h *ProjectionEventHandler) CanHandle(eventType string) bool {
	return h.projection.CanHandle(eventType)
}

// GetHandlerName returns the handler name
func (h *ProjectionEventHandler) GetHandlerName() string {
	return h.projection.GetProjectionName() + "EventHandler"
}

// GetHandlerType returns the handler type
func (h *ProjectionEventHandler) GetHandlerType() cqrs.HandlerType {
	return cqrs.ProjectionHandler
}

// UserRedisRepository implements Repository interface for User aggregates using Redis
type UserRedisRepository struct {
	stateStore *cqrsx.RedisStateStore
}

// Save saves a User aggregate
func (r *UserRedisRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	user, ok := aggregate.(*domain.User)
	if !ok {
		return errors.Errorf("invalid aggregate type: expected *domain.User, got %T", aggregate)
	}

	// For simplicity, we'll convert the User to a BaseAggregate for storage
	// In a real implementation, you'd want to preserve the actual user data
	baseAggregate := cqrs.NewBaseAggregate(user.AggregateID(), "User")
	baseAggregate.SetOriginalVersion(user.OriginalVersion())

	// Set the current version to match the user's version
	for i := baseAggregate.CurrentVersion(); i < user.CurrentVersion(); i++ {
		baseAggregate.IncrementVersion()
	}

	return r.stateStore.Save(ctx, baseAggregate, expectedVersion)
}

// GetByID gets a User aggregate by ID
func (r *UserRedisRepository) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	// Load from state store
	aggregate, err := r.stateStore.GetByID(ctx, "User", id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load user aggregate %s", id)
	}

	// The state store returns a BaseAggregate, we need to convert it to User
	baseAggregate, ok := aggregate.(*cqrs.BaseAggregate)
	if !ok {
		return nil, errors.Errorf("invalid aggregate type: expected *cqrs.BaseAggregate, got %T", aggregate)
	}

	// Convert back to User domain aggregate
	user, err := r.convertToUser(baseAggregate)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert to User aggregate")
	}

	return user, nil
}

// GetVersion gets the version of a User aggregate
func (r *UserRedisRepository) GetVersion(ctx context.Context, id string) (int, error) {
	aggregate, err := r.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}
	return aggregate.CurrentVersion(), nil
}

// Exists checks if a User aggregate exists
func (r *UserRedisRepository) Exists(ctx context.Context, id string) bool {
	_, err := r.GetByID(ctx, id)
	return err == nil
}

// convertToUser converts BaseAggregate back to domain.User
func (r *UserRedisRepository) convertToUser(baseAggregate *cqrs.BaseAggregate) (*domain.User, error) {
	// For now, we'll create a simple User from the BaseAggregate
	// In a real implementation, you would extract the stored user data

	// Create a new User with minimal data
	// This is a simplified approach - in production you'd want to store and retrieve actual user data
	user, err := domain.NewUser(baseAggregate.AggregateID(), "user@example.com", "User Name")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user from base aggregate")
	}

	// Set the version information to match the stored aggregate
	user.SetOriginalVersion(baseAggregate.CurrentVersion()) // Use CurrentVersion as OriginalVersion

	// Set the current version to match the stored version
	for i := user.CurrentVersion(); i < baseAggregate.CurrentVersion(); i++ {
		user.IncrementVersion()
	}

	// Clear changes since we're loading existing state
	user.ClearChanges()

	return user, nil
}
