# Basic Event Sourcing Example

이 예제는 가장 기본적인 Event Sourcing 패턴을 보여줍니다.

## 📖 학습 목표

- Event Sourcing의 기본 개념 이해
- Aggregate와 Event의 관계
- MongoDB Event Store 기본 사용법
- 이벤트 기반 상태 복원

## 🏗️ 아키텍처

```
User Aggregate
├── UserCreated Event
├── UserUpdated Event
└── UserDeleted Event

MongoDB Collections
├── events (이벤트 저장)
└── snapshots (스냅샷 저장)
```

## 📁 파일 구조

```
01-basic-event-sourcing/
├── README.md
├── main.go                 # 메인 데모 프로그램
├── domain/
│   ├── user.go            # User Aggregate
│   └── events.go          # User 관련 이벤트들
├── infrastructure/
│   └── config.go          # MongoDB 설정
└── demo/
    └── scenarios.go       # 데모 시나리오들
```

## 🚀 실행 방법

### 1. MongoDB 실행
```bash
# Docker 사용
docker run -d -p 27017:27017 --name mongodb mongo:latest

# 또는 로컬 MongoDB 실행
mongod
```

### 2. 예제 실행
```bash
cd 01-basic-event-sourcing
go run main.go
```

### 3. 대화형 데모
프로그램 실행 후 다음 명령어들을 사용할 수 있습니다:

```
Commands:
  create <name> <email>     - 새 사용자 생성
  update <id> <name>        - 사용자 이름 업데이트  
  delete <id>               - 사용자 삭제
  get <id>                  - 사용자 조회
  history <id>              - 이벤트 히스토리 조회
  list                      - 모든 사용자 목록
  clear                     - 모든 데이터 삭제
  help                      - 도움말
  exit                      - 종료
```

## 💡 핵심 개념

### 1. Aggregate (User)
```go
type User struct {
    *cqrs.BaseAggregate
    Name    string
    Email   string
    IsActive bool
}

// 비즈니스 로직
func (u *User) UpdateName(newName string) error {
    if newName == "" {
        return errors.New("name cannot be empty")
    }
    
    event := &UserUpdated{
        UserID:   u.ID(),
        OldName:  u.Name,
        NewName:  newName,
    }
    
    u.TrackChange(event)
    return nil
}
```

### 2. Events
```go
type UserCreated struct {
    UserID string
    Name   string
    Email  string
}

type UserUpdated struct {
    UserID  string
    OldName string
    NewName string
}
```

### 3. Event Store 사용
```go
// 이벤트 저장
events := user.GetUncommittedChanges()
err := eventStore.SaveEvents(ctx, user.ID(), user.Type(), events, user.Version())

// 이벤트 복원
events, err := eventStore.GetEventHistory(ctx, userID, "User")
user := domain.NewUser()
for _, event := range events {
    user.Apply(event)
}
```

## 🔍 데모 시나리오

### 시나리오 1: 기본 CRUD 작업
1. 사용자 생성
2. 이름 업데이트
3. 사용자 조회
4. 이벤트 히스토리 확인

### 시나리오 2: 이벤트 기반 복원
1. 여러 이벤트 생성
2. 메모리에서 Aggregate 제거
3. 이벤트로부터 상태 복원
4. 복원된 상태 확인

### 시나리오 3: 동시성 처리
1. 같은 사용자에 대한 동시 업데이트
2. 버전 충돌 처리
3. 낙관적 잠금 동작 확인

## 📊 MongoDB 컬렉션 구조

### events 컬렉션
```json
{
  "_id": ObjectId("..."),
  "event_id": "uuid-string",
  "event_type": "UserCreated",
  "aggregate_id": "user-uuid",
  "aggregate_type": "User",
  "event_version": 1,
  "event_data": "{\"user_id\":\"...\",\"name\":\"John\",\"email\":\"john@example.com\"}",
  "metadata": {},
  "timestamp": ISODate("2024-01-01T00:00:00Z")
}
```

### 인덱스
- `{aggregate_id: 1, event_version: 1}` (unique)
- `{aggregate_id: 1, timestamp: 1}`
- `{event_type: 1}`

## 🧪 테스트

```bash
# 단위 테스트 실행
go test ./...

# 통합 테스트 실행 (MongoDB 필요)
go test -tags=integration ./...
```

## 🔗 다음 단계

이 예제를 완료한 후 다음 예제들을 확인해보세요:

1. [Custom Collection Names](../02-custom-collections/) - 컬렉션 명 커스터마이징
2. [Snapshots](../03-snapshots/) - 스냅샷을 활용한 성능 최적화
3. [Read Models](../04-read-models/) - Read Model과 Projection 패턴

## 🐛 문제 해결

### MongoDB 연결 실패
```bash
# MongoDB 상태 확인
docker ps | grep mongo

# 로그 확인
docker logs mongodb
```

### 포트 충돌
기본 포트 27017이 사용 중인 경우:
```bash
# 다른 포트 사용
docker run -d -p 27018:27017 --name mongodb mongo:latest
```

config.go에서 URI 수정:
```go
URI: "mongodb://localhost:27018"
```
