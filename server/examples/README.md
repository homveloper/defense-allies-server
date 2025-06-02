# Defense Allies CQRS Examples

이 디렉토리는 Defense Allies CQRS 프레임워크의 사용법을 보여주는 예제들을 포함합니다.

## 예제 목록

### 1. User Aggregate 예제 (`user/`)

User 관리 시스템을 통해 CQRS 패턴의 전체 플로우를 보여줍니다.

**구현된 기능**:
- ✅ User 생성 (CreateUser Command)
- ✅ 이메일 변경 (ChangeEmail Command)  
- ✅ 사용자 비활성화 (DeactivateUser Command)
- ✅ 사용자 조회 (GetUser Query)
- ✅ 사용자 목록 조회 (ListUsers Query)

**검증하는 내용**:
- Command → Aggregate → Events → Projections → ReadModel 플로우
- Redis vs InMemory 구현체 비교
- 동시성 제어 (Optimistic Concurrency Control)
- Event Sourcing vs State-based 저장 방식

**실행 방법**:
```bash
cd server/examples/user
go run main.go
```

## 디렉토리 구조

```
examples/
├── user/                    # User Aggregate 예제
│   ├── domain/             # Domain Layer
│   │   ├── user.go         # User Aggregate
│   │   ├── events.go       # User Domain Events
│   │   └── commands.go     # User Commands
│   ├── handlers/           # Command/Query Handlers
│   │   ├── command_handlers.go
│   │   └── query_handlers.go
│   ├── projections/        # Read Model Projections
│   │   └── user_view.go
│   └── main.go            # 예제 실행 파일
└── README.md              # 이 파일
```

## 학습 목표

1. **CQRS 패턴 이해**: Command와 Query의 분리
2. **Event Sourcing**: 이벤트 기반 상태 관리
3. **Aggregate 설계**: 도메인 로직 캡슐화
4. **Projection**: Read Model 생성 및 관리
5. **저장소 전략**: Redis vs InMemory 비교

## 다음 단계

이 예제를 통해 CQRS 패턴을 이해한 후, 실제 Defense Allies 게임 도메인을 구현할 수 있습니다:

- Authentication Domain
- Game Domain (Tower, Player, Match 등)
- Real-time Event Processing
