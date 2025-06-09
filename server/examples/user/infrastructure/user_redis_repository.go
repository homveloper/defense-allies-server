package infrastructure

import (
	"context"
	"cqrs"

	"defense-allies-server/examples/user/domain"

	"github.com/pkg/errors"
)

// UserRedisRepository implements Repository interface for User aggregates using Redis
type UserRedisRepository struct {
	stateStore *cqrsx.RedisStateStore
}

// NewUserRedisRepository creates a new UserRedisRepository
func NewUserRedisRepository(stateStore *cqrsx.RedisStateStore) *UserRedisRepository {
	return &UserRedisRepository{
		stateStore: stateStore,
	}
}

// Save saves a User aggregate
func (r *UserRedisRepository) Save(ctx context.Context, aggregate cqrs.AggregateRoot, expectedVersion int) error {
	user, ok := aggregate.(*domain.User)
	if !ok {
		return errors.Errorf("invalid aggregate type: expected *domain.User, got %T", aggregate)
	}

	// For simplicity, we'll convert the User to a BaseAggregate for storage
	// In a real implementation, you'd want to preserve the actual user data
	baseAggregate := cqrs.NewBaseAggregate(user.ID(), "User")
	baseAggregate.SetOriginalVersion(user.OriginalVersion())

	// Set the current version to match the user's version
	for i := baseAggregate.Version(); i < user.Version(); i++ {
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
	return aggregate.Version(), nil
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
	user, err := domain.NewUser(baseAggregate.ID(), "user@example.com", "User Name")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create user from base aggregate")
	}

	// Set the version information
	user.SetOriginalVersion(baseAggregate.OriginalVersion())

	// Clear changes since we're loading existing state
	user.ClearChanges()

	return user, nil
}
