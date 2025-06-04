# Guild Management System - CQRS Event Sourcing Example

이 예제는 Defense Allies CQRS 인프라를 사용한 길드 관리 시스템입니다. 실시간 협력 게임에서 필요한 길드 기능들을 CQRS 패턴으로 구현합니다.

## 🏰 도메인 모델

### **길드 (Guild)**
- 플레이어들이 모여 협력하는 조직
- 길드마스터와 길드원으로 구성된 계층 구조
- 역할 기반 권한 시스템
- 공동 자원 관리 (광산, 채굴, 운송)

### **길드원 (Guild Member)**
- 길드에 소속된 플레이어
- 역할: 길드마스터, 부길드마스터, 일반 길드원
- 권한: 초대, 추방, 채굴, 운송 등
- 상태: 활성, 비활성, 탈퇴

### **광산 (Mine)**
- 길드가 소유하는 채굴 지역
- 여러 종류의 광물 보유
- 일꾼 배치를 통한 채굴 시스템
- 채굴량과 효율성 관리

### **운송 (Transport)**
- 광산에서 길드로 광물 운송
- 운송 중 다른 길드의 침공 가능
- 방어 시간과 약탈 메커니즘
- 실시간 상태 추적

## 🎯 주요 기능

### **길드 관리**
- 길드 생성 및 해체
- 길드 정보 수정 (이름, 설명, 공지사항)
- 길드 검색 및 랭킹 시스템
- 길드 채팅 시스템

### **회원 관리**
- 가입 신청 및 승인/거절
- 길드원 초대 시스템
- 역할 및 권한 관리
- 길드원 추방 및 탈퇴
- 길드원 상태 갱신

### **광산 및 채굴**
- 광산 발견 및 점령
- 일꾼 배치 및 관리
- 채굴 효율성 최적화
- 자원 생산량 추적

### **운송 및 침공**
- 광물 운송 시작
- 운송 경로 및 시간 관리
- 다른 길드 운송 침공
- 방어 및 약탈 시스템

## 🏗️ CQRS 아키텍처

### **Command Side (Write)**
```
Commands → CommandHandlers → Aggregates → Events → EventStore
```

### **Query Side (Read)**
```
Events → EventHandlers → Projections → ReadModels → Queries
```

### **실시간 처리**
```
Events → EventBus → SSE/WebSocket → Client Updates
```

## 📁 프로젝트 구조

```
server/examples/guild/
├── domain/
│   ├── guild_aggregate.go          # 길드 Aggregate
│   ├── member.go                   # 길드원 Value Object
│   ├── mine.go                     # 광산 Value Object
│   ├── transport.go                # 운송 Value Object
│   ├── role.go                     # 역할 및 권한 정의
│   └── events/
│       ├── guild_created.go        # 길드 생성 이벤트
│       ├── member_joined.go        # 회원 가입 이벤트
│       ├── member_left.go          # 회원 탈퇴 이벤트
│       ├── member_promoted.go      # 회원 승진 이벤트
│       ├── mine_discovered.go      # 광산 발견 이벤트
│       ├── mining_started.go       # 채굴 시작 이벤트
│       ├── transport_started.go    # 운송 시작 이벤트
│       ├── transport_attacked.go   # 운송 침공 이벤트
│       └── transport_completed.go  # 운송 완료 이벤트
├── application/
│   ├── commands/
│   │   ├── guild_commands.go       # 길드 관련 커맨드
│   │   ├── member_commands.go      # 회원 관리 커맨드
│   │   ├── mining_commands.go      # 채굴 관련 커맨드
│   │   └── transport_commands.go   # 운송 관련 커맨드
│   └── handlers/
│       ├── guild_command_handler.go    # 길드 커맨드 핸들러
│       └── guild_query_handler.go      # 길드 쿼리 핸들러
├── infrastructure/
│   ├── projections/
│   │   ├── guild_summary.go        # 길드 요약 프로젝션
│   │   ├── member_list.go          # 길드원 목록 프로젝션
│   │   ├── guild_ranking.go        # 길드 랭킹 프로젝션
│   │   ├── mining_status.go        # 채굴 상태 프로젝션
│   │   └── transport_tracking.go   # 운송 추적 프로젝션
│   ├── queries/
│   │   ├── search_guilds.go        # 길드 검색 쿼리
│   │   ├── get_guild_details.go    # 길드 상세 조회
│   │   ├── get_member_list.go      # 길드원 목록 조회
│   │   ├── get_guild_ranking.go    # 길드 랭킹 조회
│   │   └── get_transport_status.go # 운송 상태 조회
│   └── chat/
│       ├── guild_chat_handler.go   # 길드 채팅 핸들러
│       └── chat_projection.go      # 채팅 프로젝션
├── main.go                         # 메인 실행 파일
└── README.md                       # 이 파일
```

## 🚀 실행 방법

```bash
cd server/examples/guild
go run main.go
```

## 📊 이벤트 소싱 플로우

### **길드 생성 플로우**
1. `CreateGuildCommand` → `GuildCreatedEvent`
2. `InviteMemberCommand` → `MemberInvitedEvent`
3. `AcceptInvitationCommand` → `MemberJoinedEvent`

### **채굴 플로우**
1. `DiscoverMineCommand` → `MineDiscoveredEvent`
2. `StartMiningCommand` → `MiningStartedEvent`
3. `AssignWorkerCommand` → `WorkerAssignedEvent`

### **운송 및 침공 플로우**
1. `StartTransportCommand` → `TransportStartedEvent`
2. `AttackTransportCommand` → `TransportAttackedEvent`
3. `DefendTransportCommand` → `TransportDefendedEvent`
4. `CompleteTransportCommand` → `TransportCompletedEvent`

각 이벤트는 EventStore에 저장되고, Projection을 통해 ReadModel이 업데이트됩니다.

## 🎮 Defense Allies CQRS 활용

이 예제는 다음 Defense Allies CQRS 컴포넌트들을 활용합니다:

- `AggregateRoot` - 길드 Aggregate 구현
- `DomainEventMessage` - 도메인 이벤트 정의
- `CommandHandler` - 커맨드 처리
- `EventHandler` - 이벤트 처리 및 프로젝션
- `EventSourcedRepository` - 이벤트 소싱 저장소
- `QueryDispatcher` - 쿼리 처리
- `EventBus` - 실시간 이벤트 전파

## 📈 확장 가능성

- 길드 연합 시스템
- 길드 전쟁 및 토너먼트
- 고급 권한 관리 시스템
- 길드 상점 및 경제 시스템
- 길드 업적 및 보상 시스템
- 크로스 서버 길드 시스템

## 🔧 기술적 특징

- **Event Sourcing**: 모든 길드 활동의 완전한 이력 추적
- **CQRS**: 읽기/쓰기 최적화된 분리 아키텍처
- **실시간 업데이트**: EventBus를 통한 즉시 상태 동기화
- **확장성**: Redis 기반 분산 처리 지원
- **동시성 제어**: Optimistic Concurrency Control
- **복원력**: 이벤트 재생을 통한 상태 복구

이 예제를 통해 복잡한 멀티플레이어 게임의 길드 시스템을 CQRS 패턴으로 구현하는 방법을 학습할 수 있습니다.
