// serializer.go - 이벤트 직렬화 구현
package cqrsx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JSONEventSerializer는 JSON 기반 이벤트 직렬화기입니다
type JSONEventSerializer struct {
	eventRegistry map[EventType]reflect.Type
	mu            sync.RWMutex
}

// SerializableEvent는 직렬화 가능한 이벤트 구조체입니다
type SerializableEvent struct {
	AggregateID uuid.UUID   `json:"aggregateId"`
	EventType   EventType   `json:"eventType"`
	Data        interface{} `json:"data"`
	Version     int         `json:"version"`
	Timestamp   time.Time   `json:"timestamp"`
	Metadata    Metadata    `json:"metadata"`
}

// NewJSONEventSerializer는 새로운 JSON 직렬화기를 생성합니다
func NewJSONEventSerializer() *JSONEventSerializer {
	serializer := &JSONEventSerializer{
		eventRegistry: make(map[EventType]reflect.Type),
	}

	// 기본 이벤트 타입들 등록
	serializer.registerDefaultEventTypes()

	return serializer
}

// RegisterEventType은 이벤트 타입을 등록합니다
func (s *JSONEventSerializer) RegisterEventType(eventType EventType, dataType interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventRegistry[eventType] = reflect.TypeOf(dataType)
}

// Serialize는 이벤트를 JSON으로 직렬화합니다
func (s *JSONEventSerializer) Serialize(event Event) ([]byte, error) {
	serializableEvent := SerializableEvent{
		AggregateID: event.AggregateID(),
		EventType:   event.EventType(),
		Data:        event.Data(),
		Version:     event.Version(),
		Timestamp:   event.Timestamp(),
		Metadata:    event.Metadata(),
	}

	return json.Marshal(serializableEvent)
}

// Deserialize는 JSON을 이벤트로 역직렬화합니다
func (s *JSONEventSerializer) Deserialize(data []byte, eventType EventType) (Event, error) {
	var serializableEvent SerializableEvent
	if err := json.Unmarshal(data, &serializableEvent); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// 등록된 타입으로 데이터 변환
	eventData, err := s.deserializeEventData(serializableEvent.Data, eventType)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize event data: %w", err)
	}

	return &BaseEvent{
		aggregateID: serializableEvent.AggregateID,
		eventType:   serializableEvent.EventType,
		data:        eventData,
		version:     serializableEvent.Version,
		timestamp:   serializableEvent.Timestamp,
		metadata:    serializableEvent.Metadata,
	}, nil
}

// Private methods

func (s *JSONEventSerializer) registerDefaultEventTypes() {
	// 게임 길드 시스템 기본 이벤트 타입들
	s.RegisterEventType("GuildCreated", GuildCreatedEvent{})
	s.RegisterEventType("MemberJoined", MemberJoinedEvent{})
	s.RegisterEventType("MemberLeft", MemberLeftEvent{})
	s.RegisterEventType("ResourceShared", ResourceSharedEvent{})
	s.RegisterEventType("GuildLevelUp", GuildLevelUpEvent{})
	s.RegisterEventType("RoleChanged", RoleChangedEvent{})
}

func (s *JSONEventSerializer) deserializeEventData(data interface{}, eventType EventType) (EventData, error) {
	s.mu.RLock()
	targetType, exists := s.eventRegistry[eventType]
	s.mu.RUnlock()

	if !exists {
		// 등록되지 않은 타입은 원본 데이터 그대로 반환
		return data, nil
	}

	// JSON 재직렬화 후 타입 변환
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// 새 인스턴스 생성
	newInstance := reflect.New(targetType).Interface()
	if err := json.Unmarshal(jsonData, newInstance); err != nil {
		return nil, err
	}

	return reflect.ValueOf(newInstance).Elem().Interface(), nil
}

// 게임 길드 시스템 이벤트 타입들

// GuildCreatedEvent는 길드 생성 이벤트입니다
type GuildCreatedEvent struct {
	GuildID    uuid.UUID `json:"guildId"`
	FounderID  uuid.UUID `json:"founderId"`
	Name       string    `json:"name"`
	MaxMembers int       `json:"maxMembers"`
	GameType   string    `json:"gameType"`
	CreatedAt  time.Time `json:"createdAt"`
}

// MemberJoinedEvent는 멤버 가입 이벤트입니다
type MemberJoinedEvent struct {
	GuildID          uuid.UUID `json:"guildId"`
	MemberID         uuid.UUID `json:"memberId"`
	Role             string    `json:"role"`
	InvitedBy        uuid.UUID `json:"invitedBy"`
	JoinedAt         time.Time `json:"joinedAt"`
	InvitationMethod string    `json:"invitationMethod"`
}

// MemberLeftEvent는 멤버 탈퇴 이벤트입니다
type MemberLeftEvent struct {
	GuildID  uuid.UUID  `json:"guildId"`
	MemberID uuid.UUID  `json:"memberId"`
	Reason   string     `json:"reason"`
	LeftAt   time.Time  `json:"leftAt"`
	KickedBy *uuid.UUID `json:"kickedBy,omitempty"`
}

// ResourceSharedEvent는 자원 공유 이벤트입니다
type ResourceSharedEvent struct {
	GuildID       uuid.UUID `json:"guildId"`
	ContributorID uuid.UUID `json:"contributorId"`
	ResourceType  string    `json:"resourceType"`
	Amount        int       `json:"amount"`
	TotalBalance  int       `json:"totalBalance"`
	SharedAt      time.Time `json:"sharedAt"`
}

// GuildLevelUpEvent는 길드 레벨업 이벤트입니다
type GuildLevelUpEvent struct {
	GuildID    uuid.UUID `json:"guildId"`
	OldLevel   int       `json:"oldLevel"`
	NewLevel   int       `json:"newLevel"`
	Experience int       `json:"experience"`
	LevelUpAt  time.Time `json:"levelUpAt"`
	Rewards    []string  `json:"rewards"`
}

// RoleChangedEvent는 역할 변경 이벤트입니다
type RoleChangedEvent struct {
	GuildID   uuid.UUID `json:"guildId"`
	MemberID  uuid.UUID `json:"memberId"`
	OldRole   string    `json:"oldRole"`
	NewRole   string    `json:"newRole"`
	ChangedBy uuid.UUID `json:"changedBy"`
	ChangedAt time.Time `json:"changedAt"`
	Reason    string    `json:"reason"`
}

// BinaryEventSerializer는 바이너리 직렬화기입니다 (성능 최적화용)
type BinaryEventSerializer struct {
	// 바이너리 직렬화 구현 (protobuf, msgpack 등 활용)
	// 게임에서 높은 성능이 필요한 경우 사용
}

// CompressedEventSerializer는 압축 직렬화기입니다
type CompressedEventSerializer struct {
	underlying EventSerializer
	// gzip, lz4 등 압축 알고리즘 적용
}

// VersionedEventSerializer는 버전 관리 직렬화기입니다
type VersionedEventSerializer struct {
	serializers    map[int]EventSerializer
	currentVersion int
}

// NewVersionedEventSerializer는 버전 관리 직렬화기를 생성합니다
func NewVersionedEventSerializer() *VersionedEventSerializer {
	return &VersionedEventSerializer{
		serializers:    make(map[int]EventSerializer),
		currentVersion: 1,
	}
}

// AddVersion은 새 버전의 직렬화기를 추가합니다
func (v *VersionedEventSerializer) AddVersion(version int, serializer EventSerializer) {
	v.serializers[version] = serializer
	if version > v.currentVersion {
		v.currentVersion = version
	}
}

// Serialize는 현재 버전으로 직렬화합니다
func (v *VersionedEventSerializer) Serialize(event Event) ([]byte, error) {
	serializer := v.serializers[v.currentVersion]
	if serializer == nil {
		return nil, fmt.Errorf("no serializer for version %d", v.currentVersion)
	}

	// 버전 정보를 포함한 직렬화
	data, err := serializer.Serialize(event)
	if err != nil {
		return nil, err
	}

	// 버전 헤더 추가
	versionedData := append([]byte{byte(v.currentVersion)}, data...)
	return versionedData, nil
}

// Deserialize는 버전에 맞는 역직렬화를 수행합니다
func (v *VersionedEventSerializer) Deserialize(data []byte, eventType EventType) (Event, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("invalid versioned data")
	}

	version := int(data[0])
	serializer := v.serializers[version]
	if serializer == nil {
		return nil, fmt.Errorf("no serializer for version %d", version)
	}

	return serializer.Deserialize(data[1:], eventType)
}
