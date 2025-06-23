# 통합 인증 서비스 구현 계획

## 📋 프로젝트 개요

Defense Allies 타임스퀘어 앱에 **통합 인증 서비스**를 구현하여 게스트 인증을 시작으로 Apple, Google 등 다양한 인증 제공자를 지원하는 확장 가능한 시스템을 구축합니다.

## 🎯 Phase 1: 게스트 인증 구현 목표

### 구현 범위
- **Guest Provider**: Device ID 기반 게스트 인증 구현
- **Game Account Service**: 게임 계정 생성 및 관리
- **Authentication Service**: 인증 오케스트레이션 서비스  
- **TimeSquare Integration**: 타임스퀘어 앱에 인증 엔드포인트 추가
- **JWT Session Management**: 세션 토큰 생성 및 관리

### 성공 기준
- ✅ Device ID로 즉시 게스트 로그인 가능
- ✅ 게임 계정 자동 생성 및 고유 게임 ID 발급
- ✅ JWT 기반 세션 관리
- ✅ 기존 게스트 계정 재로그인 지원
- ✅ 타임스퀘어 앱과 완전 통합

## 🏗️ 구현 아키텍처

### 디렉토리 구조
```
server/
├── pkg/
│   └── gameauth/              # 통합 인증 서비스 패키지
│       ├── domain/
│       │   ├── gameaccount/   # 게임 계정 도메인
│       │   │   ├── aggregate.go
│       │   │   ├── events.go
│       │   │   └── repository.go
│       │   ├── authsession/   # 인증 세션 도메인
│       │   │   ├── aggregate.go
│       │   │   ├── events.go
│       │   │   └── repository.go
│       │   └── common/
│       │       └── types.go   # 공통 타입 정의
│       ├── application/
│       │   ├── auth/          # 인증 애플리케이션 서비스
│       │   │   ├── service.go
│       │   │   ├── commands.go
│       │   │   └── handlers.go
│       │   ├── providers/     # 인증 제공자들
│       │   │   ├── interfaces.go
│       │   │   ├── guest/
│       │   │   │   └── provider.go
│       │   │   └── registry.go
│       │   └── gameaccount/   # 게임 계정 서비스
│       │       ├── service.go
│       │       ├── commands.go
│       │       └── handlers.go
│       ├── infrastructure/
│       │   ├── repositories/  # Repository 구현체
│       │   │   ├── redis_gameaccount_repo.go
│       │   │   └── redis_authsession_repo.go
│       │   ├── jwt/           # JWT 토큰 관리
│       │   │   ├── manager.go
│       │   │   └── claims.go
│       │   └── uuid/          # ID 생성기
│       │       └── generator.go
│       └── api/               # HTTP API 핸들러
│           ├── handlers.go
│           ├── middleware.go
│           └── routes.go
└── serverapp/
    └── timesquare/
        ├── app.go             # 기존 타임스퀘어 앱
        └── auth_integration.go # 인증 서비스 통합
```

## 📝 단계별 구현 계획

### Step 1: 도메인 모델 구현 (1-2일)

#### GameAccount Aggregate
```go
// 구현 목표
type GameAccount struct {
    ID          string                    // 게임 계정 고유 ID
    Username    string                    // 게임 내 사용자명
    DisplayName string                    // 표시명
    Status      GameAccountStatus         // Active, Suspended, Deleted
    
    // 연결된 인증 제공자들
    AuthProviders map[ProviderType]AuthProviderInfo
    
    // 메타데이터 (기기 정보 등)
    Metadata    GameAccountMetadata
    
    // 기본 필드
    CreatedAt   time.Time
    UpdatedAt   time.Time
    LastLoginAt *time.Time
}
```

#### AuthSession Aggregate
```go
// 구현 목표
type AuthSession struct {
    ID              string           // 세션 고유 ID
    GameAccountID   string           // 게임 계정 ID
    ProviderType    ProviderType     // 사용된 인증 제공자
    SessionToken    string           // JWT 토큰
    RefreshToken    string           // 갱신 토큰
    Status          SessionStatus    // Active, Expired, Revoked
    
    CreatedAt       time.Time
    ExpiresAt       time.Time
    LastActivityAt  time.Time
    
    ClientInfo      ClientInfo       // 클라이언트 정보
}
```

#### 구현 작업
- [ ] `pkg/gameauth/domain/gameaccount/aggregate.go` 구현
- [ ] `pkg/gameauth/domain/authsession/aggregate.go` 구현
- [ ] `pkg/gameauth/domain/common/types.go` 공통 타입 정의
- [ ] Domain Events 정의 (GameAccountCreated, AuthSessionStarted 등)
- [ ] Repository 인터페이스 정의

### Step 2: Guest Provider 구현 (1일)

#### Guest Provider 인터페이스
```go
// 구현 목표
type AuthProvider interface {
    ProviderType() ProviderType
    Authenticate(ctx context.Context, credentials interface{}) (*AuthResult, error)
    GenerateGameID(ctx context.Context, externalID string) (string, error)
    ValidateCredentials(credentials interface{}) error
}

type GuestProvider struct {
    idGenerator IDGenerator
    repository  GameAccountRepository
}
```

#### 구현 작업
- [ ] `pkg/gameauth/application/providers/interfaces.go` 인터페이스 정의
- [ ] `pkg/gameauth/application/providers/guest/provider.go` Guest Provider 구현
- [ ] Device ID 검증 로직 구현
- [ ] 게임 ID 생성 규칙 구현 (예: `guest_${hash(device_id)}`)
- [ ] Provider Registry 구현

### Step 3: Game Account Service 구현 (1-2일)

#### Game Account Service
```go
// 구현 목표
type GameAccountService struct {
    repository    GameAccountRepository
    eventBus      EventBus
    idGenerator   IDGenerator
}

// 주요 메서드
func (s *GameAccountService) CreateAccount(ctx context.Context, cmd CreateGameAccountCommand) (*GameAccount, error)
func (s *GameAccountService) LoadAccount(ctx context.Context, gameID string) (*GameAccount, error)
func (s *GameAccountService) LinkProvider(ctx context.Context, cmd LinkProviderCommand) error
```

#### 구현 작업
- [ ] `pkg/gameauth/application/gameaccount/service.go` 서비스 구현
- [ ] `pkg/gameauth/application/gameaccount/commands.go` 커맨드 정의
- [ ] `pkg/gameauth/application/gameaccount/handlers.go` 커맨드 핸들러 구현
- [ ] CQRS 패턴 적용 (Command/Query 분리)
- [ ] Event 발행 로직 구현

### Step 4: Authentication Service 구현 (1-2일)

#### Authentication Service (오케스트레이션)
```go
// 구현 목표
type AuthenticationService struct {
    providerRegistry   ProviderRegistry
    gameAccountService GameAccountService
    sessionManager     SessionManager
    jwtManager        JWTManager
}

// 주요 메서드
func (s *AuthenticationService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
func (s *AuthenticationService) LinkProvider(ctx context.Context, req LinkProviderRequest) error
func (s *AuthenticationService) RefreshSession(ctx context.Context, refreshToken string) (*LoginResponse, error)
```

#### 구현 작업
- [ ] `pkg/gameauth/application/auth/service.go` 메인 인증 서비스 구현
- [ ] Provider 선택 및 위임 로직 구현
- [ ] 게임 계정 생성/로드 오케스트레이션
- [ ] 세션 생성 및 JWT 토큰 발급
- [ ] 에러 처리 및 로깅

### Step 5: Infrastructure 구현 (2일)

#### Redis Repository 구현
```go
// 구현 목표
type RedisGameAccountRepository struct {
    client     *cqrsx.RedisClientManager
    serializer Serializer
}

type RedisAuthSessionRepository struct {
    client     *cqrsx.RedisClientManager
    serializer Serializer
}
```

#### JWT Manager 구현
```go
// 구현 목표
type JWTManager struct {
    secretKey    []byte
    issuer       string
    defaultTTL   time.Duration
}

type GameAuthClaims struct {
    GameAccountID    string   `json:"game_account_id"`
    Username         string   `json:"username"`
    ProviderType     string   `json:"provider_type"`
    LinkedProviders  []string `json:"linked_providers"`
    Permissions      []string `json:"permissions"`
    IsGuestOnly      bool     `json:"is_guest_only"`
    jwt.RegisteredClaims
}
```

#### 구현 작업
- [ ] `pkg/gameauth/infrastructure/repositories/` Redis Repository 구현
- [ ] `pkg/gameauth/infrastructure/jwt/` JWT 토큰 관리 구현
- [ ] `pkg/gameauth/infrastructure/uuid/` ID 생성기 구현
- [ ] Redis 키 네이밍 규칙 정의
- [ ] 직렬화/역직렬화 로직 구현

### Step 6: HTTP API 구현 (1일)

#### REST API 엔드포인트
```go
// 구현 목표
POST /api/v1/auth/login/guest
GET  /api/v1/auth/session/refresh  
GET  /api/v1/account/profile
POST /api/v1/auth/logout
```

#### 구현 작업
- [ ] `pkg/gameauth/api/handlers.go` HTTP 핸들러 구현
- [ ] `pkg/gameauth/api/routes.go` 라우팅 설정
- [ ] `pkg/gameauth/api/middleware.go` 인증 미들웨어 구현
- [ ] Request/Response 구조체 정의
- [ ] 입력 검증 및 에러 처리

### Step 7: TimeSquare 앱 통합 (1일)

#### TimeSquare 앱에 인증 서비스 통합
```go
// 구현 목표
type TimeSquareApp struct {
    // 기존 필드들...
    authService *gameauth.AuthenticationService
}

func (app *TimeSquareApp) setupAuthRoutes() {
    // 인증 관련 라우트 설정
}
```

#### 구현 작업
- [ ] `serverapp/timesquare/auth_integration.go` 통합 모듈 구현
- [ ] 기존 TimeSquare 앱에 인증 서비스 주입
- [ ] 인증 관련 라우트 추가
- [ ] 설정 파일 업데이트
- [ ] 의존성 주입 설정

### Step 8: 테스트 구현 (1-2일)

#### 테스트 범위
- **Unit Tests**: 각 컴포넌트 단위 테스트
- **Integration Tests**: Redis 통합 테스트
- **API Tests**: HTTP API 엔드투엔드 테스트
- **Load Tests**: 성능 테스트

#### 구현 작업
- [ ] Guest Provider 단위 테스트
- [ ] Game Account Service 테스트
- [ ] Authentication Service 테스트
- [ ] Redis Repository 통합 테스트
- [ ] HTTP API 테스트
- [ ] 성능 벤치마크 테스트

## 🛠️ 기술 스택 및 의존성

### 새로 추가할 의존성
```go
// go.mod 추가 필요
require (
    github.com/golang-jwt/jwt/v5 v5.2.0     // JWT 토큰 관리
    github.com/google/uuid v1.6.0           // UUID 생성 (이미 있음)
    golang.org/x/crypto v0.17.0             // 암호화 유틸리티
)
```

### 기존 활용 가능한 컴포넌트
- **Redis Client**: 기존 `pkg/cqrs/cqrsx/redis_client.go` 활용
- **CQRS Framework**: 기존 `pkg/cqrs/` 패키지 활용
- **Event Bus**: 기존 Event Bus 시스템 활용
- **Serialization**: 기존 JSON 직렬화 활용

## 📊 개발 일정

### 총 개발 기간: 7-10일

| 단계 | 작업 내용 | 예상 시간 | 의존성 |
|------|----------|----------|--------|
| Step 1 | 도메인 모델 구현 | 1-2일 | - |
| Step 2 | Guest Provider 구현 | 1일 | Step 1 |
| Step 3 | Game Account Service | 1-2일 | Step 1, 2 |
| Step 4 | Authentication Service | 1-2일 | Step 2, 3 |
| Step 5 | Infrastructure 구현 | 2일 | Step 1-4 |
| Step 6 | HTTP API 구현 | 1일 | Step 4, 5 |
| Step 7 | TimeSquare 통합 | 1일 | Step 6 |
| Step 8 | 테스트 구현 | 1-2일 | 모든 단계 |

### 마일스톤
- **Week 1 End**: 도메인 모델 + Provider 구현 완료
- **Week 2 Mid**: 인증 서비스 + Infrastructure 완료
- **Week 2 End**: TimeSquare 통합 + 테스트 완료

## 🔍 검증 계획

### 기능 검증
1. **Guest 로그인 테스트**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login/guest \
        -H "Content-Type: application/json" \
        -d '{"device_id":"test-device-123","device_info":{"platform":"iOS 17.0"}}'
   ```

2. **세션 검증 테스트**
   ```bash
   curl -X GET http://localhost:8080/api/v1/account/profile \
        -H "Authorization: Bearer {jwt_token}"
   ```

3. **재로그인 테스트**
   - 동일한 device_id로 다시 로그인 시 기존 계정 반환 확인

### 성능 검증
- **목표 응답시간**: Guest 로그인 < 100ms
- **동시 접속**: 1,000명 동시 로그인 처리
- **메모리 사용량**: 안정적인 메모리 사용 패턴

### 보안 검증
- JWT 토큰 유효성 검증
- Device ID 중복 처리 검증
- 세션 만료 처리 검증

## 🚀 배포 준비

### 설정 파일 업데이트
```yaml
# configs/timesquare.yaml
auth:
  jwt:
    secret_key: "${JWT_SECRET_KEY}"
    issuer: "defense-allies-timesquare"
    access_token_ttl: "1h"
    refresh_token_ttl: "30d"
  
  guest:
    enabled: true
    game_id_prefix: "guest_"
    username_prefix: "Guest_"

redis:
  # 기존 Redis 설정 활용
```

### 환경 변수
```bash
export JWT_SECRET_KEY="your-secret-key-here"
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
```

### 모니터링 지표
- 인증 성공/실패 비율
- 평균 응답 시간
- 동시 세션 수
- 에러 발생률

이 계획을 바탕으로 단계별로 구현을 진행하면 안정적이고 확장 가능한 통합 인증 서비스를 구축할 수 있습니다.