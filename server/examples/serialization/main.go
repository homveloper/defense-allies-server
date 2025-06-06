package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"defense-allies-server/examples/user/domain"
	"defense-allies-server/pkg/cqrs"

	"go.mongodb.org/mongo-driver/bson"
)

// UserAggregateData represents serializable user data
type UserAggregateData struct {
	*cqrs.BaseAggregate
	Email              string                   `json:"email" bson:"email"`
	Name               string                   `json:"name" bson:"name"`
	Status             string                   `json:"status" bson:"status"`
	LastLoginAt        *time.Time               `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	DeactivatedAt      *time.Time               `json:"deactivated_at,omitempty" bson:"deactivated_at,omitempty"`
	DeactivationReason string                   `json:"deactivation_reason,omitempty" bson:"deactivation_reason,omitempty"`
	Profile            map[string]interface{}   `json:"profile,omitempty" bson:"profile,omitempty"`
	Roles              []map[string]interface{} `json:"roles,omitempty" bson:"roles,omitempty"`
}

// UserSerializer handles User aggregate serialization
type UserSerializer struct{}

// SerializeUser converts User aggregate to serializable data
func (s *UserSerializer) SerializeUser(user *domain.User) (*UserAggregateData, error) {
	// Convert profile to map
	var profileData map[string]interface{}
	if user.GetProfile() != nil {
		profileData = map[string]interface{}{
			"first_name":   user.GetProfile().FirstName,
			"last_name":    user.GetProfile().LastName,
			"display_name": user.GetProfile().DisplayName,
			"bio":          user.GetProfile().Bio,
			"avatar":       user.GetProfile().Avatar,
			"phone_number": user.GetProfile().PhoneNumber,
			"address":      user.GetProfile().Address,
			"city":         user.GetProfile().City,
			"country":      user.GetProfile().Country,
			"postal_code":  user.GetProfile().PostalCode,
			"preferences":  user.GetProfile().Preferences,
		}
	}

	// Convert roles to slice of maps
	var rolesData []map[string]interface{}
	for _, role := range user.GetRoles() {
		roleData := map[string]interface{}{
			"type":        role.Type.String(),
			"assigned_by": role.AssignedBy,
			"assigned_at": role.AssignedAt,
		}
		if role.ExpiresAt != nil {
			roleData["expires_at"] = *role.ExpiresAt
		}
		rolesData = append(rolesData, roleData)
	}

	// Create base aggregate copy for serialization
	baseAggregate := cqrs.NewBaseAggregate(user.ID(), user.Type())
	baseAggregate.SetOriginalVersion(user.OriginalVersion())

	// Set current version to match user's version
	for i := baseAggregate.Version(); i < user.Version(); i++ {
		baseAggregate.IncrementVersion()
	}

	baseAggregate.SetCreatedAt(user.CreatedAt())
	baseAggregate.SetUpdatedAt(user.UpdatedAt())
	baseAggregate.SetDeleted(user.Deleted())

	return &UserAggregateData{
		BaseAggregate:      baseAggregate,
		Email:              user.Email(),
		Name:               user.Name(),
		Status:             user.Status().String(),
		LastLoginAt:        user.LastLoginAt(),
		DeactivatedAt:      user.DeactivatedAt(),
		DeactivationReason: user.DeactivationReason(),
		Profile:            profileData,
		Roles:              rolesData,
	}, nil
}

// DeserializeUser converts serializable data back to User aggregate
func (s *UserSerializer) DeserializeUser(data *UserAggregateData) (*domain.User, error) {
	// Create user with basic information
	user, err := domain.NewUser(data.ID(), data.Email, data.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Set base aggregate properties
	user.SetOriginalVersion(data.OriginalVersion())

	// Set current version to match serialized version
	for i := user.Version(); i < data.Version(); i++ {
		user.IncrementVersion()
	}

	user.SetCreatedAt(data.CreatedAt())
	user.SetUpdatedAt(data.UpdatedAt())
	user.SetDeleted(data.IsDeleted())

	// Clear any changes from creation
	user.ClearChanges()

	return user, nil
}

// JSONUserRepository demonstrates JSON-based user storage
type JSONUserRepository struct {
	serializer *UserSerializer
	storage    map[string][]byte // In-memory storage for demo
}

// NewJSONUserRepository creates a new JSON-based repository
func NewJSONUserRepository() *JSONUserRepository {
	return &JSONUserRepository{
		serializer: &UserSerializer{},
		storage:    make(map[string][]byte),
	}
}

// Save saves a user aggregate as JSON
func (r *JSONUserRepository) Save(user *domain.User) error {
	// Serialize user to data structure
	userData, err := r.serializer.SerializeUser(user)
	if err != nil {
		return fmt.Errorf("failed to serialize user: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user to JSON: %w", err)
	}

	// Store in memory (in real implementation, this would be database/file)
	r.storage[user.ID()] = jsonData

	fmt.Printf("✅ Saved user %s as JSON (%d bytes)\n", user.ID(), len(jsonData))
	return nil
}

// Load loads a user aggregate from JSON
func (r *JSONUserRepository) Load(userID string) (*domain.User, error) {
	// Get JSON data from storage
	jsonData, exists := r.storage[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	// Unmarshal JSON to data structure
	var userData UserAggregateData
	if err := json.Unmarshal(jsonData, &userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user JSON: %w", err)
	}

	// Deserialize to user aggregate
	user, err := r.serializer.DeserializeUser(&userData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize user: %w", err)
	}

	fmt.Printf("✅ Loaded user %s from JSON\n", userID)
	return user, nil
}

// BSONUserRepository demonstrates BSON-based user storage
type BSONUserRepository struct {
	serializer *UserSerializer
	storage    map[string][]byte // In-memory storage for demo
}

// NewBSONUserRepository creates a new BSON-based repository
func NewBSONUserRepository() *BSONUserRepository {
	return &BSONUserRepository{
		serializer: &UserSerializer{},
		storage:    make(map[string][]byte),
	}
}

// Save saves a user aggregate as BSON
func (r *BSONUserRepository) Save(user *domain.User) error {
	// Serialize user to data structure
	userData, err := r.serializer.SerializeUser(user)
	if err != nil {
		return fmt.Errorf("failed to serialize user: %w", err)
	}

	// Convert to BSON
	bsonData, err := bson.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user to BSON: %w", err)
	}

	// Store in memory (in real implementation, this would be MongoDB)
	r.storage[user.ID()] = bsonData

	fmt.Printf("✅ Saved user %s as BSON (%d bytes)\n", user.ID(), len(bsonData))
	return nil
}

// Load loads a user aggregate from BSON
func (r *BSONUserRepository) Load(userID string) (*domain.User, error) {
	// Get BSON data from storage
	bsonData, exists := r.storage[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	// Unmarshal BSON to data structure
	var userData UserAggregateData
	if err := bson.Unmarshal(bsonData, &userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user BSON: %w", err)
	}

	// Deserialize to user aggregate
	user, err := r.serializer.DeserializeUser(&userData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize user: %w", err)
	}

	fmt.Printf("✅ Loaded user %s from BSON\n", userID)
	return user, nil
}

func main() {
	fmt.Println("🚀 Defense Allies Aggregate Serialization Example")
	fmt.Println("=================================================")

	// 1. Create a user aggregate
	user, err := domain.NewUser("user-123", "alice@example.com", "Alice Smith")
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	// Modify user state
	if err := user.ChangeEmail("alice.smith@example.com"); err != nil {
		log.Fatalf("Failed to change email: %v", err)
	}

	if err := user.UpdateDisplayName("Alice S."); err != nil {
		log.Fatalf("Failed to update display name: %v", err)
	}

	fmt.Printf("📝 Created user: %s (%s) - Version: %d\n",
		user.Name(), user.Email(), user.Version())

	// 2. Test JSON serialization
	fmt.Println("\n=== JSON Serialization Test ===")
	jsonRepo := NewJSONUserRepository()

	if err := jsonRepo.Save(user); err != nil {
		log.Fatalf("Failed to save user as JSON: %v", err)
	}

	loadedUserJSON, err := jsonRepo.Load(user.ID())
	if err != nil {
		log.Fatalf("Failed to load user from JSON: %v", err)
	}

	fmt.Printf("📋 Loaded user: %s (%s) - Version: %d\n",
		loadedUserJSON.Name(), loadedUserJSON.Email(), loadedUserJSON.Version())

	// 3. Test BSON serialization
	fmt.Println("\n=== BSON Serialization Test ===")
	bsonRepo := NewBSONUserRepository()

	if err := bsonRepo.Save(user); err != nil {
		log.Fatalf("Failed to save user as BSON: %v", err)
	}

	loadedUserBSON, err := bsonRepo.Load(user.ID())
	if err != nil {
		log.Fatalf("Failed to load user from BSON: %v", err)
	}

	fmt.Printf("📋 Loaded user: %s (%s) - Version: %d\n",
		loadedUserBSON.Name(), loadedUserBSON.Email(), loadedUserBSON.Version())

	// 4. Test BaseAggregate serialization
	fmt.Println("\n=== BaseAggregate Serialization Test ===")
	example := &cqrs.SerializationExample{}
	if err := example.RunAllDemos(); err != nil {
		log.Fatalf("Failed to run serialization demos: %v", err)
	}

	fmt.Println("\n🎉 All serialization tests completed successfully!")
}
