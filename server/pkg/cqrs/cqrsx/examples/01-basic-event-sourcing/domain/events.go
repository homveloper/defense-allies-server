package domain

import (
	"time"
)

// UserCreated 이벤트 - 사용자가 생성되었을 때 발생
type UserCreated struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// EventType implements cqrs.DomainEvent
func (e *UserCreated) EventType() string {
	return "UserCreated"
}

// string implements cqrs.DomainEvent
func (e *UserCreated) ID() string {
	return e.UserID
}

// EventData implements cqrs.DomainEvent
func (e *UserCreated) EventData() interface{} {
	return e
}

// UserUpdated 이벤트 - 사용자 정보가 업데이트되었을 때 발생
type UserUpdated struct {
	UserID    string    `json:"user_id"`
	OldName   string    `json:"old_name"`
	NewName   string    `json:"new_name"`
	OldEmail  string    `json:"old_email"`
	NewEmail  string    `json:"new_email"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EventType implements cqrs.DomainEvent
func (e *UserUpdated) EventType() string {
	return "UserUpdated"
}

// string implements cqrs.DomainEvent
func (e *UserUpdated) ID() string {
	return e.UserID
}

// EventData implements cqrs.DomainEvent
func (e *UserUpdated) EventData() interface{} {
	return e
}

// UserDeleted 이벤트 - 사용자가 삭제되었을 때 발생
type UserDeleted struct {
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
	Reason    string    `json:"reason,omitempty"`
}

// EventType implements cqrs.DomainEvent
func (e *UserDeleted) EventType() string {
	return "UserDeleted"
}

// string implements cqrs.DomainEvent
func (e *UserDeleted) ID() string {
	return e.UserID
}

// EventData implements cqrs.DomainEvent
func (e *UserDeleted) EventData() interface{} {
	return e
}

// UserActivated 이벤트 - 사용자가 활성화되었을 때 발생
type UserActivated struct {
	UserID      string    `json:"user_id"`
	ActivatedAt time.Time `json:"activated_at"`
	ActivatedBy string    `json:"activated_by,omitempty"`
}

// EventType implements cqrs.DomainEvent
func (e *UserActivated) EventType() string {
	return "UserActivated"
}

// string implements cqrs.DomainEvent
func (e *UserActivated) ID() string {
	return e.UserID
}

// EventData implements cqrs.DomainEvent
func (e *UserActivated) EventData() interface{} {
	return e
}

// UserDeactivated 이벤트 - 사용자가 비활성화되었을 때 발생
type UserDeactivated struct {
	UserID        string    `json:"user_id"`
	DeactivatedAt time.Time `json:"deactivated_at"`
	DeactivatedBy string    `json:"deactivated_by,omitempty"`
	Reason        string    `json:"reason,omitempty"`
}

// EventType implements cqrs.DomainEvent
func (e *UserDeactivated) EventType() string {
	return "UserDeactivated"
}

// string implements cqrs.DomainEvent
func (e *UserDeactivated) ID() string {
	return e.UserID
}

// EventData implements cqrs.DomainEvent
func (e *UserDeactivated) EventData() interface{} {
	return e
}

// CreateUserCreatedEvent 헬퍼 함수 - UserCreated 이벤트 생성
func CreateUserCreatedEvent(userID, name, email string) *UserCreated {
	return &UserCreated{
		UserID:    userID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}
}

// CreateUserUpdatedEvent 헬퍼 함수 - UserUpdated 이벤트 생성
func CreateUserUpdatedEvent(userID, oldName, newName, oldEmail, newEmail string) *UserUpdated {
	return &UserUpdated{
		UserID:    userID,
		OldName:   oldName,
		NewName:   newName,
		OldEmail:  oldEmail,
		NewEmail:  newEmail,
		UpdatedAt: time.Now(),
	}
}

// CreateUserDeletedEvent 헬퍼 함수 - UserDeleted 이벤트 생성
func CreateUserDeletedEvent(userID, name, email, reason string) *UserDeleted {
	return &UserDeleted{
		UserID:    userID,
		Name:      name,
		Email:     email,
		DeletedAt: time.Now(),
		Reason:    reason,
	}
}

// CreateUserActivatedEvent 헬퍼 함수 - UserActivated 이벤트 생성
func CreateUserActivatedEvent(userID, activatedBy string) *UserActivated {
	return &UserActivated{
		UserID:      userID,
		ActivatedAt: time.Now(),
		ActivatedBy: activatedBy,
	}
}

// CreateUserDeactivatedEvent 헬퍼 함수 - UserDeactivated 이벤트 생성
func CreateUserDeactivatedEvent(userID, deactivatedBy, reason string) *UserDeactivated {
	return &UserDeactivated{
		UserID:        userID,
		DeactivatedAt: time.Now(),
		DeactivatedBy: deactivatedBy,
		Reason:        reason,
	}
}

// EventFactory 함수들 - 이벤트 타입별 팩토리 등록용
func init() {
	// 이벤트 팩토리 등록 (필요시 사용)
	// cqrs.RegisterEventType("UserCreated", func(data interface{}) (cqrs.EventMessage, error) {
	//     // 구현 필요시 추가
	// })
}
