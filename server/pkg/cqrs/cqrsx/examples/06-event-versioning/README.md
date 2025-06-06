# Event Versioning Example

이 예제는 Event Sourcing에서 이벤트 스키마의 진화와 버전 관리를 다룹니다.

## 🎯 목적

- 이벤트 스키마 변경 처리
- 하위 호환성 유지
- 마이그레이션 전략 구현
- 버전별 직렬화/역직렬화

## 📋 시나리오

### 1. 초기 버전 (V1)
```go
// UserCreatedV1 - 기본 사용자 정보만 포함
type UserCreatedV1 struct {
    UserID   string `json:"user_id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
}
```

### 2. 확장 버전 (V2)
```go
// UserCreatedV2 - 추가 필드 포함
type UserCreatedV2 struct {
    UserID      string            `json:"user_id"`
    Name        string            `json:"name"`
    Email       string            `json:"email"`
    Profile     UserProfile       `json:"profile"`      // 새로운 필드
    Preferences UserPreferences   `json:"preferences"`  // 새로운 필드
    CreatedAt   time.Time         `json:"created_at"`   // 새로운 필드
}
```

### 3. 구조 변경 버전 (V3)
```go
// UserCreatedV3 - 구조적 변경
type UserCreatedV3 struct {
    UserID      string            `json:"user_id"`
    PersonalInfo PersonalInfo     `json:"personal_info"` // 구조 변경
    ContactInfo  ContactInfo      `json:"contact_info"`  // 구조 변경
    Metadata     EventMetadata    `json:"metadata"`      // 새로운 구조
}
```

## 🔧 핵심 기능

### 1. Event Version Manager
- 이벤트 버전 감지
- 자동 업캐스팅/다운캐스팅
- 버전별 직렬화 전략

### 2. Migration Strategies
- **Forward Migration**: V1 → V2 → V3
- **Backward Compatibility**: V3 → V2 → V1
- **Lazy Migration**: 읽을 때만 변환
- **Batch Migration**: 전체 이벤트 일괄 변환

### 3. Schema Evolution Patterns
- **Additive Changes**: 필드 추가
- **Structural Changes**: 구조 변경
- **Breaking Changes**: 호환성 없는 변경

## 🚀 실행 방법

```bash
# 기본 실행
go run cmd/basic/main.go

# 마이그레이션 실행
go run cmd/migration/main.go

# 성능 테스트
go run cmd/performance/main.go
```

## 📁 폴더 구조

```
06-event-versioning/
├── README.md
├── cmd/
│   ├── basic/           # 기본 버전 관리 데모
│   ├── migration/       # 마이그레이션 데모
│   └── performance/     # 성능 비교 데모
├── domain/
│   ├── user.go         # User Aggregate
│   ├── events_v1.go    # V1 이벤트들
│   ├── events_v2.go    # V2 이벤트들
│   └── events_v3.go    # V3 이벤트들
├── versioning/
│   ├── version_manager.go    # 버전 관리자
│   ├── upcaster.go          # 업캐스팅 로직
│   ├── downcaster.go        # 다운캐스팅 로직
│   └── migration.go         # 마이그레이션 전략
└── infrastructure/
    ├── versioned_event_store.go  # 버전 지원 Event Store
    ├── serializers.go           # 버전별 직렬화
    └── repositories.go          # Repository 구현
```

## 🎓 학습 포인트

1. **이벤트 스키마 진화의 어려움**
2. **하위 호환성 유지 전략**
3. **성능과 호환성의 트레이드오프**
4. **점진적 마이그레이션 방법**
5. **버전 관리 모범 사례**

## 🔍 주요 패턴

### 1. Upcasting Pattern
```go
func (u *UserEventUpcaster) UpcastV1ToV2(v1Event *UserCreatedV1) *UserCreatedV2 {
    return &UserCreatedV2{
        UserID:      v1Event.UserID,
        Name:        v1Event.Name,
        Email:       v1Event.Email,
        Profile:     DefaultUserProfile(),     // 기본값 설정
        Preferences: DefaultUserPreferences(), // 기본값 설정
        CreatedAt:   time.Now(),              // 현재 시간으로 설정
    }
}
```

### 2. Downcasting Pattern
```go
func (d *UserEventDowncaster) DowncastV2ToV1(v2Event *UserCreatedV2) *UserCreatedV1 {
    return &UserCreatedV1{
        UserID: v2Event.UserID,
        Name:   v2Event.Name,
        Email:  v2Event.Email,
        // Profile, Preferences, CreatedAt 필드는 제거
    }
}
```

### 3. Version Detection
```go
func (vm *VersionManager) DetectVersion(eventData []byte) (int, error) {
    // JSON 구조 분석을 통한 버전 감지
    // 또는 메타데이터의 version 필드 확인
}
```

## 📊 성능 비교

| 전략                | 읽기 성능 | 쓰기 성능 | 저장 공간 | 호환성 |
| ------------------- | --------- | --------- | --------- | ------ |
| Lazy Migration      | 느림      | 빠름      | 적음      | 높음   |
| Eager Migration     | 빠름      | 느림      | 많음      | 높음   |
| Version Branching   | 빠름      | 빠름      | 많음      | 중간   |
| Schema Evolution    | 중간      | 중간      | 중간      | 높음   |

## 🚨 주의사항

1. **Breaking Changes 최소화**
2. **버전 정보 메타데이터 포함**
3. **테스트 커버리지 확보**
4. **점진적 배포 전략**
5. **롤백 계획 수립**
