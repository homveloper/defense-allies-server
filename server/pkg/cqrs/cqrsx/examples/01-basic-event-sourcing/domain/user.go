package domain

import (
	"defense-allies-server/pkg/cqrs"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// User Aggregate - 사용자 도메인 모델
type User struct {
	*cqrs.BaseAggregate
	name      string
	email     string
	isActive  bool
	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

// NewUser 새로운 User Aggregate 생성
func NewUser() *User {
	return &User{
		BaseAggregate: cqrs.NewBaseAggregate("", "User"),
		isActive:      false,
	}
}

// NewUserWithID ID를 지정하여 User Aggregate 생성
func NewUserWithID(id string) *User {
	return &User{
		BaseAggregate: cqrs.NewBaseAggregate(id, "User"),
		isActive:      false,
	}
}

// CreateUser 새 사용자 생성 - 비즈니스 로직
func (u *User) CreateUser(id, name, email string) error {
	// 비즈니스 규칙 검증
	if err := u.validateUserCreation(id, name, email); err != nil {
		return err
	}

	// 이미 생성된 사용자인지 확인
	if u.Version() > 0 {
		return errors.New("user already exists")
	}

	// Aggregate ID 설정 (BaseAggregate의 id 필드 직접 설정)
	u.BaseAggregate = cqrs.NewBaseAggregate(id, "User")

	// 이벤트 생성 및 추적
	event := CreateUserCreatedEvent(id, name, email)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		id,
		"User",
		1,
		event,
	)

	u.TrackChange(eventMessage)
	return nil
}

// UpdateUser 사용자 정보 업데이트
func (u *User) UpdateUser(newName, newEmail string) error {
	// 사용자가 존재하는지 확인
	if u.Version() == 0 {
		return errors.New("user does not exist")
	}

	// 삭제된 사용자인지 확인
	if u.IsDeleted() {
		return errors.New("cannot update deleted user")
	}

	// 변경사항이 있는지 확인
	if u.name == newName && u.email == newEmail {
		return errors.New("no changes detected")
	}

	// 새로운 값 검증
	if err := u.validateUserUpdate(newName, newEmail); err != nil {
		return err
	}

	// 이벤트 생성 및 추적
	event := CreateUserUpdatedEvent(u.ID(), u.name, newName, u.email, newEmail)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		u.ID(),
		"User",
		u.Version()+1,
		event,
	)

	u.TrackChange(eventMessage)
	return nil
}

// DeleteUser 사용자 삭제
func (u *User) DeleteUser(reason string) error {
	// 사용자가 존재하는지 확인
	if u.Version() == 0 {
		return errors.New("user does not exist")
	}

	// 이미 삭제된 사용자인지 확인
	if u.IsDeleted() {
		return errors.New("user already deleted")
	}

	// 이벤트 생성 및 추적
	event := CreateUserDeletedEvent(u.ID(), u.name, u.email, reason)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		u.ID(),
		"User",
		u.Version()+1,
		event,
	)

	u.TrackChange(eventMessage)
	return nil
}

// ActivateUser 사용자 활성화
func (u *User) ActivateUser(activatedBy string) error {
	if u.Version() == 0 {
		return errors.New("user does not exist")
	}

	if u.IsDeleted() {
		return errors.New("cannot activate deleted user")
	}

	if u.isActive {
		return errors.New("user already active")
	}

	// 이벤트 생성 및 추적
	event := CreateUserActivatedEvent(u.ID(), activatedBy)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		u.ID(),
		"User",
		u.Version()+1,
		event,
	)

	u.TrackChange(eventMessage)
	return nil
}

// DeactivateUser 사용자 비활성화
func (u *User) DeactivateUser(deactivatedBy, reason string) error {
	if u.Version() == 0 {
		return errors.New("user does not exist")
	}

	if u.IsDeleted() {
		return errors.New("cannot deactivate deleted user")
	}

	if !u.isActive {
		return errors.New("user already inactive")
	}

	// 이벤트 생성 및 추적
	event := CreateUserDeactivatedEvent(u.ID(), deactivatedBy, reason)
	eventMessage := cqrs.NewBaseEventMessage(
		event.EventType(),
		u.ID(),
		"User",
		u.Version()+1,
		event,
	)

	u.TrackChange(eventMessage)
	return nil
}

// Apply 이벤트를 적용하여 상태 변경
func (u *User) ApplyEvent(event cqrs.EventMessage) error {
	// BaseAggregate의 Apply 메서드 호출 (버전 관리)
	if err := u.BaseAggregate.ApplyEvent(event); err != nil {
		return err
	}

	// BSON에서 역직렬화된 데이터를 구체적인 이벤트 타입으로 변환
	eventData, err := u.convertEventData(event.EventType(), event.EventData())
	if err != nil {
		return fmt.Errorf("failed to convert event data for %s: %w", event.EventType(), err)
	}

	switch e := eventData.(type) {
	case *UserCreated:
		return u.applyUserCreated(e)
	case *UserUpdated:
		return u.applyUserUpdated(e)
	case *UserDeleted:
		return u.applyUserDeleted(e)
	case *UserActivated:
		return u.applyUserActivated(e)
	case *UserDeactivated:
		return u.applyUserDeactivated(e)
	default:
		return fmt.Errorf("unknown event type: %T", e)
	}
}

// 이벤트 적용 메서드들
func (u *User) applyUserCreated(event *UserCreated) error {
	// ID는 이미 NewUserWithID에서 설정되었거나 BaseAggregate에서 관리됨
	u.name = event.Name
	u.email = event.Email
	u.isActive = false
	u.createdAt = event.CreatedAt
	u.updatedAt = event.CreatedAt
	// BaseAggregate에서 버전 관리하므로 IncrementVersion() 호출 제거
	return nil
}

func (u *User) applyUserUpdated(event *UserUpdated) error {
	u.name = event.NewName
	u.email = event.NewEmail
	u.updatedAt = event.UpdatedAt
	return nil
}

func (u *User) applyUserDeleted(event *UserDeleted) error {
	now := event.DeletedAt
	u.deletedAt = &now
	u.updatedAt = event.DeletedAt
	u.MarkAsDeleted()
	return nil
}

func (u *User) applyUserActivated(event *UserActivated) error {
	u.isActive = true
	u.updatedAt = event.ActivatedAt
	return nil
}

func (u *User) applyUserDeactivated(event *UserDeactivated) error {
	u.isActive = false
	u.updatedAt = event.DeactivatedAt
	return nil
}

// Getter 메서드들
func (u *User) Name() string {
	return u.name
}

func (u *User) Email() string {
	return u.email
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) DeletedAt() *time.Time {
	return u.deletedAt
}

// Version returns the current version (convenience method)
func (u *User) Version() int {
	return u.Version()
}

// ID returns the aggregate ID (convenience method)
func (u *User) ID() string {
	return u.ID()
}

// Type returns the aggregate type (convenience method)
func (u *User) Type() string {
	return u.Type()
}

// GetUncommittedChanges returns uncommitted changes (convenience method)
func (u *User) GetUncommittedChanges() []cqrs.EventMessage {
	return u.GetChanges()
}

// SetID sets the aggregate ID (convenience method)
func (u *User) SetID(id string) {
	// For this example, we'll use reflection or direct field access
	// In a real implementation, BaseAggregate should provide this method
	if u.BaseAggregate != nil {
		// We'll set it during event application
	}
}

// 검증 메서드들
func (u *User) validateUserCreation(id, name, email string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("user ID cannot be empty")
	}

	if strings.TrimSpace(name) == "" {
		return errors.New("name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}

	if err := u.validateEmail(email); err != nil {
		return err
	}

	return nil
}

func (u *User) validateUserUpdate(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}

	if err := u.validateEmail(email); err != nil {
		return err
	}

	return nil
}

func (u *User) validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("email cannot be empty")
	}

	// 간단한 이메일 형식 검증
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	if len(email) > 255 {
		return errors.New("email cannot exceed 255 characters")
	}

	return nil
}

// String 사용자 정보를 문자열로 반환
func (u *User) String() string {
	status := "inactive"
	if u.isActive {
		status = "active"
	}
	if u.IsDeleted() {
		status = "deleted"
	}

	return fmt.Sprintf("User{ID: %s, Name: %s, Email: %s, Status: %s, Version: %d}",
		u.ID(), u.name, u.email, status, u.Version())
}

// convertEventData BSON 데이터를 구체적인 이벤트 타입으로 변환
func (u *User) convertEventData(eventType string, data interface{}) (interface{}, error) {
	// BSON 바이트로 변환
	bsonData, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	// 이벤트 타입에 따라 구체적인 구조체로 역직렬화
	switch eventType {
	case "UserCreated":
		var event UserCreated
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserCreated: %w", err)
		}
		return &event, nil

	case "UserUpdated":
		var event UserUpdated
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserUpdated: %w", err)
		}
		return &event, nil

	case "UserDeleted":
		var event UserDeleted
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserDeleted: %w", err)
		}
		return &event, nil

	case "UserActivated":
		var event UserActivated
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserActivated: %w", err)
		}
		return &event, nil

	case "UserDeactivated":
		var event UserDeactivated
		if err := bson.Unmarshal(bsonData, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UserDeactivated: %w", err)
		}
		return &event, nil

	default:
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}
}
