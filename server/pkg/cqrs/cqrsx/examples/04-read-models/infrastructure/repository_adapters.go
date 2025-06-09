package infrastructure

import (
	"context"
	"cqrs"
	"cqrs/cqrsx"
	"cqrs/cqrsx/examples/04-read-models/domain"
	"fmt"
)

// UserRepositoryAdapter adapts cqrsx repository to domain UserRepository interface
type UserRepositoryAdapter struct {
	repo *cqrsx.MongoEventSourcedRepository
}

// NewUserRepositoryAdapter creates a new user repository adapter
func NewUserRepositoryAdapter(repo *cqrsx.MongoEventSourcedRepository) domain.UserRepository {
	return &UserRepositoryAdapter{
		repo: repo,
	}
}

// Save saves a user aggregate with automatic version management
func (a *UserRepositoryAdapter) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	// 자동 버전 관리: 적절한 expectedVersion을 자동으로 결정
	autoExpectedVersion := a.determineExpectedVersion(ctx, aggregate, expectedVersion)
	return a.repo.Save(ctx, aggregate, autoExpectedVersion)
}

// determineExpectedVersion automatically determines the correct expected version
func (a *UserRepositoryAdapter) determineExpectedVersion(ctx context.Context, aggregate cqrs.AggregateRoot, providedVersion int) int {
	// 1. 새로운 Aggregate 감지 (OriginalVersion=0이고 변경사항이 있는 경우)
	if aggregate.OriginalVersion() == 0 && len(aggregate.GetChanges()) > 0 {
		return -1 // 새로운 Aggregate는 -1 사용
	}

	// 2. 기존 Aggregate인 경우 OriginalVersion 사용
	if aggregate.OriginalVersion() > 0 {
		return aggregate.OriginalVersion()
	}

	// 3. 제공된 버전이 유효한 경우 사용
	if providedVersion >= 0 {
		return providedVersion
	}

	// 4. 기본값: 새로운 Aggregate로 간주
	return -1
}

// GetByID loads a user aggregate by ID with automatic version setup
func (a *UserRepositoryAdapter) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	// 1. 실제 로드
	aggregate, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 로드 후 버전 상태 자동 설정
	a.setupVersionAfterLoad(aggregate)

	return aggregate, nil
}

// setupVersionAfterLoad sets up version state after loading an aggregate
func (a *UserRepositoryAdapter) setupVersionAfterLoad(aggregate cqrs.AggregateRoot) {
	// 로드된 버전을 OriginalVersion으로 설정
	currentVersion := aggregate.Version()
	aggregate.SetOriginalVersion(currentVersion)

	// 변경사항 정리 (로드된 상태이므로 변경사항 없음)
	aggregate.ClearChanges()
}

// GetVersion gets the version of an aggregate
func (a *UserRepositoryAdapter) GetVersion(ctx context.Context, id string) (int, error) {
	return a.repo.GetVersion(ctx, id)
}

// Exists checks if an aggregate exists
func (a *UserRepositoryAdapter) Exists(ctx context.Context, id string) bool {
	return a.repo.Exists(ctx, id)
}

// SaveEvents saves events for an aggregate (not directly supported, use Save instead)
func (a *UserRepositoryAdapter) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	// This is handled internally by the Save method
	return fmt.Errorf("SaveEvents not directly supported, use Save method instead")
}

// GetEventHistory gets event history for an aggregate (not directly supported)
func (a *UserRepositoryAdapter) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]cqrs.EventMessage, error) {
	// This would require direct access to the event store
	return nil, fmt.Errorf("GetEventHistory not directly supported")
}

// GetEventStream gets event stream for an aggregate (not directly supported)
func (a *UserRepositoryAdapter) GetEventStream(ctx context.Context, aggregateID string) (<-chan cqrs.EventMessage, error) {
	// This would require streaming support
	return nil, fmt.Errorf("GetEventStream not directly supported")
}

// SaveSnapshot saves a snapshot
func (a *UserRepositoryAdapter) SaveSnapshot(ctx context.Context, snapshot cqrs.SnapshotData) error {
	// Convert SnapshotData to AggregateRoot if possible
	if aggregate, ok := snapshot.(cqrs.AggregateRoot); ok {
		return a.repo.SaveSnapshot(ctx, aggregate)
	}
	return fmt.Errorf("snapshot must be an AggregateRoot")
}

// GetSnapshot gets a snapshot (not directly supported)
func (a *UserRepositoryAdapter) GetSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	// This would require direct access to the snapshot store
	return nil, fmt.Errorf("GetSnapshot not directly supported")
}

// DeleteSnapshot deletes a snapshot (not directly supported)
func (a *UserRepositoryAdapter) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	// This would require direct access to the snapshot store
	return fmt.Errorf("DeleteSnapshot not directly supported")
}

// GetLastEventVersion gets the last event version
func (a *UserRepositoryAdapter) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	// Use GetVersion as a fallback
	return a.repo.GetVersion(ctx, aggregateID)
}

// CompactEvents compacts events (not directly supported)
func (a *UserRepositoryAdapter) CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error {
	// This would require direct access to the event store
	return fmt.Errorf("CompactEvents not directly supported")
}

// FindByEmail finds a user by email address
func (a *UserRepositoryAdapter) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the database directly
	// For now, we'll return nil to indicate not found
	return nil, fmt.Errorf("user not found with email: %s", email)
}

// GetUserStats gets user statistics
func (a *UserRepositoryAdapter) GetUserStats(ctx context.Context, userID string) (*domain.UserStats, error) {
	// This is a simplified implementation
	// In a real implementation, you would calculate stats from the database
	return nil, fmt.Errorf("GetUserStats not implemented")
}

// OrderRepositoryAdapter adapts cqrsx repository to domain OrderRepository interface
type OrderRepositoryAdapter struct {
	repo *cqrsx.MongoHybridRepository
}

// NewOrderRepositoryAdapter creates a new order repository adapter
func NewOrderRepositoryAdapter(repo *cqrsx.MongoHybridRepository) domain.OrderRepository {
	return &OrderRepositoryAdapter{
		repo: repo,
	}
}

// Save saves an order aggregate with automatic version management
func (a *OrderRepositoryAdapter) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	// 자동 버전 관리: 적절한 expectedVersion을 자동으로 결정
	autoExpectedVersion := a.determineExpectedVersion(ctx, aggregate, expectedVersion)
	return a.repo.Save(ctx, aggregate, autoExpectedVersion)
}

// GetByID loads an order aggregate by ID with automatic version setup
func (a *OrderRepositoryAdapter) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	// 1. 실제 로드
	aggregate, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 로드 후 버전 상태 자동 설정
	a.setupVersionAfterLoad(aggregate)

	return aggregate, nil
}

// determineExpectedVersion automatically determines the correct expected version
func (a *OrderRepositoryAdapter) determineExpectedVersion(ctx context.Context, aggregate cqrs.AggregateRoot, providedVersion int) int {
	// 1. 새로운 Aggregate 감지
	if aggregate.OriginalVersion() == 0 && len(aggregate.GetChanges()) > 0 {
		return -1 // 새로운 Aggregate는 -1 사용
	}

	// 2. 기존 Aggregate인 경우 OriginalVersion 사용
	if aggregate.OriginalVersion() > 0 {
		return aggregate.OriginalVersion()
	}

	// 3. 제공된 버전이 유효한 경우 사용
	if providedVersion >= 0 {
		return providedVersion
	}

	// 4. 기본값: 새로운 Aggregate로 간주
	return -1
}

// setupVersionAfterLoad sets up version state after loading an aggregate
func (a *OrderRepositoryAdapter) setupVersionAfterLoad(aggregate cqrs.AggregateRoot) {
	// 로드된 버전을 OriginalVersion으로 설정
	currentVersion := aggregate.Version()
	aggregate.SetOriginalVersion(currentVersion)

	// 변경사항 정리 (로드된 상태이므로 변경사항 없음)
	aggregate.ClearChanges()
}

// GetVersion gets the version of an aggregate
func (a *OrderRepositoryAdapter) GetVersion(ctx context.Context, id string) (int, error) {
	return a.repo.GetVersion(ctx, id)
}

// Exists checks if an aggregate exists
func (a *OrderRepositoryAdapter) Exists(ctx context.Context, id string) bool {
	return a.repo.Exists(ctx, id)
}

// SaveEvents saves events for an aggregate (not directly supported)
func (a *OrderRepositoryAdapter) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	return fmt.Errorf("SaveEvents not directly supported, use Save method instead")
}

// GetEventHistory gets event history for an aggregate (not directly supported)
func (a *OrderRepositoryAdapter) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]cqrs.EventMessage, error) {
	return nil, fmt.Errorf("GetEventHistory not directly supported")
}

// GetEventStream gets event stream for an aggregate (not directly supported)
func (a *OrderRepositoryAdapter) GetEventStream(ctx context.Context, aggregateID string) (<-chan cqrs.EventMessage, error) {
	return nil, fmt.Errorf("GetEventStream not directly supported")
}

// SaveSnapshot saves a snapshot
func (a *OrderRepositoryAdapter) SaveSnapshot(ctx context.Context, snapshot cqrs.SnapshotData) error {
	if aggregate, ok := snapshot.(cqrs.AggregateRoot); ok {
		return a.repo.SaveSnapshot(ctx, aggregate)
	}
	return fmt.Errorf("snapshot must be an AggregateRoot")
}

// GetSnapshot gets a snapshot (not directly supported)
func (a *OrderRepositoryAdapter) GetSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	return nil, fmt.Errorf("GetSnapshot not directly supported")
}

// DeleteSnapshot deletes a snapshot (not directly supported)
func (a *OrderRepositoryAdapter) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	return fmt.Errorf("DeleteSnapshot not directly supported")
}

// GetLastEventVersion gets the last event version
func (a *OrderRepositoryAdapter) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	return a.repo.GetVersion(ctx, aggregateID)
}

// CompactEvents compacts events (not directly supported)
func (a *OrderRepositoryAdapter) CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error {
	return fmt.Errorf("CompactEvents not directly supported")
}

// Create creates a new order (use Save instead)
func (a *OrderRepositoryAdapter) Create(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	return a.repo.Save(ctx, aggregate, 0)
}

// Update updates an order (use Save instead)
func (a *OrderRepositoryAdapter) Update(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	return a.repo.Save(ctx, aggregate, aggregate.Version())
}

// Delete deletes an order (not directly supported)
func (a *OrderRepositoryAdapter) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("Delete not directly supported")
}

// FindBy finds orders by criteria (not directly supported)
func (a *OrderRepositoryAdapter) FindBy(ctx context.Context, criteria cqrs.QueryCriteria) ([]cqrs.AggregateRoot, error) {
	return nil, fmt.Errorf("FindBy not directly supported")
}

// Count counts orders by criteria (not directly supported)
func (a *OrderRepositoryAdapter) Count(ctx context.Context, criteria cqrs.QueryCriteria) (int64, error) {
	return 0, fmt.Errorf("Count not directly supported")
}

// SaveBatch saves multiple orders (not directly supported)
func (a *OrderRepositoryAdapter) SaveBatch(ctx context.Context, aggregates []cqrs.AggregateRoot) error {
	return fmt.Errorf("SaveBatch not directly supported")
}

// DeleteBatch deletes multiple orders (not directly supported)
func (a *OrderRepositoryAdapter) DeleteBatch(ctx context.Context, ids []string) error {
	return fmt.Errorf("DeleteBatch not directly supported")
}

// FindByCustomerID finds orders by customer ID
func (a *OrderRepositoryAdapter) FindByCustomerID(ctx context.Context, customerID string) ([]*domain.Order, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the database directly
	return []*domain.Order{}, nil
}

// FindByStatus finds orders by status
func (a *OrderRepositoryAdapter) FindByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the database directly
	return []*domain.Order{}, nil
}

// GetOrderStats gets order statistics
func (a *OrderRepositoryAdapter) GetOrderStats(ctx context.Context, orderID string) (*domain.OrderStats, error) {
	// This is a simplified implementation
	// In a real implementation, you would calculate stats from the database
	return nil, fmt.Errorf("GetOrderStats not implemented")
}

// ProductRepositoryAdapter adapts cqrsx repository to domain ProductRepository interface
type ProductRepositoryAdapter struct {
	repo *cqrsx.MongoStateBasedRepository
}

// NewProductRepositoryAdapter creates a new product repository adapter
func NewProductRepositoryAdapter(repo *cqrsx.MongoStateBasedRepository) domain.ProductRepository {
	return &ProductRepositoryAdapter{
		repo: repo,
	}
}

// Save saves a product aggregate with automatic version management
func (a *ProductRepositoryAdapter) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	// 자동 버전 관리: 적절한 expectedVersion을 자동으로 결정
	autoExpectedVersion := a.determineExpectedVersion(ctx, aggregate, expectedVersion)
	return a.repo.Save(ctx, aggregate, autoExpectedVersion)
}

// GetByID loads a product aggregate by ID with automatic version setup
func (a *ProductRepositoryAdapter) GetByID(ctx context.Context, id string) (cqrs.AggregateRoot, error) {
	// 1. 실제 로드
	aggregate, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. 로드 후 버전 상태 자동 설정
	a.setupVersionAfterLoad(aggregate)

	return aggregate, nil
}

// determineExpectedVersion automatically determines the correct expected version
func (a *ProductRepositoryAdapter) determineExpectedVersion(ctx context.Context, aggregate cqrs.AggregateRoot, providedVersion int) int {
	// 1. 새로운 Aggregate 감지
	if aggregate.OriginalVersion() == 0 && len(aggregate.GetChanges()) > 0 {
		return -1 // 새로운 Aggregate는 -1 사용
	}

	// 2. 기존 Aggregate인 경우 OriginalVersion 사용
	if aggregate.OriginalVersion() > 0 {
		return aggregate.OriginalVersion()
	}

	// 3. 제공된 버전이 유효한 경우 사용
	if providedVersion >= 0 {
		return providedVersion
	}

	// 4. 기본값: 새로운 Aggregate로 간주
	return -1
}

// setupVersionAfterLoad sets up version state after loading an aggregate
func (a *ProductRepositoryAdapter) setupVersionAfterLoad(aggregate cqrs.AggregateRoot) {
	// 로드된 버전을 OriginalVersion으로 설정
	currentVersion := aggregate.Version()
	aggregate.SetOriginalVersion(currentVersion)

	// 변경사항 정리 (로드된 상태이므로 변경사항 없음)
	aggregate.ClearChanges()
}

// GetVersion gets the version of an aggregate
func (a *ProductRepositoryAdapter) GetVersion(ctx context.Context, id string) (int, error) {
	return a.repo.GetVersion(ctx, id)
}

// Exists checks if an aggregate exists
func (a *ProductRepositoryAdapter) Exists(ctx context.Context, id string) bool {
	return a.repo.Exists(ctx, id)
}

// Create creates a new product (use Save instead)
func (a *ProductRepositoryAdapter) Create(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	return a.repo.Save(ctx, aggregate, 0)
}

// Update updates a product (use Save instead)
func (a *ProductRepositoryAdapter) Update(ctx context.Context, aggregate cqrs.AggregateRoot) error {
	return a.repo.Save(ctx, aggregate, aggregate.Version())
}

// Delete deletes a product (not directly supported)
func (a *ProductRepositoryAdapter) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("Delete not directly supported")
}

// FindBy finds products by criteria (not directly supported)
func (a *ProductRepositoryAdapter) FindBy(ctx context.Context, criteria cqrs.QueryCriteria) ([]cqrs.AggregateRoot, error) {
	return nil, fmt.Errorf("FindBy not directly supported")
}

// Count counts products by criteria (not directly supported)
func (a *ProductRepositoryAdapter) Count(ctx context.Context, criteria cqrs.QueryCriteria) (int64, error) {
	return 0, fmt.Errorf("Count not directly supported")
}

// SaveBatch saves multiple products (not directly supported)
func (a *ProductRepositoryAdapter) SaveBatch(ctx context.Context, aggregates []cqrs.AggregateRoot) error {
	return fmt.Errorf("SaveBatch not directly supported")
}

// DeleteBatch deletes multiple products (not directly supported)
func (a *ProductRepositoryAdapter) DeleteBatch(ctx context.Context, ids []string) error {
	return fmt.Errorf("DeleteBatch not directly supported")
}

// FindByCategory finds products by category
func (a *ProductRepositoryAdapter) FindByCategory(ctx context.Context, category string) ([]*domain.Product, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the database directly
	return []*domain.Product{}, nil
}

// FindActiveProducts finds all active products
func (a *ProductRepositoryAdapter) FindActiveProducts(ctx context.Context) ([]*domain.Product, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the database directly
	return []*domain.Product{}, nil
}

// GetProductStats gets product statistics
func (a *ProductRepositoryAdapter) GetProductStats(ctx context.Context, productID string) (*domain.ProductStats, error) {
	// This is a simplified implementation
	// In a real implementation, you would calculate stats from the database
	return nil, fmt.Errorf("GetProductStats not implemented")
}

// SaveEvents saves events for an aggregate (not directly supported)
func (a *ProductRepositoryAdapter) SaveEvents(ctx context.Context, aggregateID string, events []cqrs.EventMessage, expectedVersion int) error {
	return fmt.Errorf("SaveEvents not directly supported, use Save method instead")
}

// GetEventHistory gets event history for an aggregate (not directly supported)
func (a *ProductRepositoryAdapter) GetEventHistory(ctx context.Context, aggregateID string, fromVersion int) ([]cqrs.EventMessage, error) {
	return nil, fmt.Errorf("GetEventHistory not directly supported")
}

// GetEventStream gets event stream for an aggregate (not directly supported)
func (a *ProductRepositoryAdapter) GetEventStream(ctx context.Context, aggregateID string) (<-chan cqrs.EventMessage, error) {
	return nil, fmt.Errorf("GetEventStream not directly supported")
}

// SaveSnapshot saves a snapshot (not directly supported)
func (a *ProductRepositoryAdapter) SaveSnapshot(ctx context.Context, snapshot cqrs.SnapshotData) error {
	return fmt.Errorf("SaveSnapshot not directly supported")
}

// GetSnapshot gets a snapshot (not directly supported)
func (a *ProductRepositoryAdapter) GetSnapshot(ctx context.Context, aggregateID string) (cqrs.SnapshotData, error) {
	return nil, fmt.Errorf("GetSnapshot not directly supported")
}

// DeleteSnapshot deletes a snapshot (not directly supported)
func (a *ProductRepositoryAdapter) DeleteSnapshot(ctx context.Context, aggregateID string) error {
	return fmt.Errorf("DeleteSnapshot not directly supported")
}

// GetLastEventVersion gets the last event version
func (a *ProductRepositoryAdapter) GetLastEventVersion(ctx context.Context, aggregateID string) (int, error) {
	return a.repo.GetVersion(ctx, aggregateID)
}

// CompactEvents is a placeholder to satisfy the interface
func (a *ProductRepositoryAdapter) CompactEvents(ctx context.Context, aggregateID string, beforeVersion int) error {
	// State-based repositories don't typically have events to compact
	return nil
}
