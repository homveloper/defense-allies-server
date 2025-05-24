# Redis 데이터 구조 설계

## 🔴 **Redis 키 패턴**

### 1. **유저 데이터**
```
user:{user_id}  (Hash)
├── id: string
├── username: string  
├── email: string
├── created_at: RFC3339 timestamp
├── updated_at: RFC3339 timestamp
├── last_login: RFC3339 timestamp
└── game_data: JSON string
```

**예시:**
```redis
HGETALL user:auth0|12345
1) "id"
2) "auth0|12345"
3) "username" 
4) "player123"
5) "email"
6) "player@example.com"
7) "created_at"
8) "2024-01-15T10:30:00Z"
9) "updated_at"
10) "2024-01-15T15:45:00Z"
11) "last_login"
12) "2024-01-15T15:45:00Z"
13) "game_data"
14) "{\"level\":5,\"score\":1500,\"resources\":{\"gold\":2500}}"
```

### 2. **인덱스 구조**

#### 유저 목록 인덱스
```
users:index  (Set)
- 모든 유저 ID 저장
```

#### 이메일 인덱스
```
users:email_index  (Hash)
email -> user_id 매핑
```

#### 유저명 인덱스  
```
users:username_index  (Hash)
username -> user_id 매핑
```

#### 마지막 로그인 시간 인덱스
```
users:last_login  (Sorted Set)
score: unix timestamp
member: user_id
```

### 3. **유저 역할**
```
user:roles:{user_id}  (Set)
- 유저의 역할들 저장
```

### 4. **게임 세션 (선택적)**
```
session:{session_id}  (Hash)
├── id: string
├── user_id: string
├── server_instance: string
├── created_at: RFC3339 timestamp
├── expires_at: RFC3339 timestamp
├── last_activity: RFC3339 timestamp
├── ip_address: string
└── user_agent: string
```

## 🔐 **보안 및 분산 환경 고려사항**

### 1. **원자적 작업 보장**
```go
// Redis Pipeline 사용으로 원자성 보장
pipe := redis.Pipeline()
pipe.HMSet(ctx, userKey, userData)
pipe.SAdd(ctx, UserIndexKey, userID)
pipe.ZAdd(ctx, UserLastLoginKey, redis.Z{Score: timestamp, Member: userID})
_, err := pipe.Exec(ctx)
```

### 2. **중복 생성 방지**
- Redis의 원자적 연산 활용
- SET IF NOT EXISTS 패턴 사용
- 이메일/유저명 유니크 제약 인덱스로 보장

### 3. **데이터 일관성**
```go
// 트랜잭션 스타일 업데이트
pipe := redis.Pipeline()
pipe.HMSet(ctx, userKey, updates)
pipe.ZAdd(ctx, indexKey, score)
_, err := pipe.Exec(ctx)
```

### 4. **성능 최적화**

#### 배치 조회
```go
// 파이프라인으로 여러 유저 동시 조회
pipe := redis.Pipeline()
for _, userID := range userIDs {
    pipe.HGetAll(ctx, UserKeyPrefix + userID)
}
results, _ := pipe.Exec(ctx)
```

#### 캐싱 전략
```go
// 자주 조회되는 데이터 TTL 설정
redis.Set(ctx, "cache:user:"+userID, userData, 5*time.Minute)
```

## 🚀 **사용 예시**

### 1. **신규 유저 생성**
```bash
# 1. 유저 데이터 저장
HMSET user:auth0|12345 id "auth0|12345" username "player123" email "player@example.com" created_at "2024-01-15T10:30:00Z"

# 2. 인덱스 업데이트
SADD users:index "auth0|12345"
HSET users:email_index "player@example.com" "auth0|12345"
HSET users:username_index "player123" "auth0|12345"
ZADD users:last_login 1705315800 "auth0|12345"

# 3. 역할 설정
SADD user:roles:auth0|12345 "player" "premium"
```

### 2. **유저 조회**
```bash
# ID로 조회
HGETALL user:auth0|12345

# 이메일로 조회
HGET users:email_index "player@example.com"
HGETALL user:auth0|12345

# 최근 로그인 유저들
ZREVRANGE users:last_login 0 99 WITHSCORES
```

### 3. **게임 데이터 업데이트**
```bash
# 게임 데이터만 업데이트
HSET user:auth0|12345 game_data "{\"level\":6,\"score\":2000}" updated_at "2024-01-15T16:00:00Z"
```

## 📊 **모니터링 쿼리**

### 활성 유저 수
```bash
SCARD users:index
```

### 최근 24시간 로그인 유저
```bash
ZCOUNT users:last_login $(date -d "24 hours ago" +%s) $(date +%s)
```

### 메모리 사용량 확인
```bash
MEMORY USAGE user:auth0|12345
```

## 🔧 **백업 및 복구**

### 데이터 백업
```bash
# RDB 스냅샷
BGSAVE

# AOF 백업
BGREWRITEAOF
```

### 데이터 마이그레이션
```bash
# 키 패턴별 백업
redis-cli --scan --pattern "user:*" | xargs redis-cli DUMP
```

## ⚠️ **주의사항**

1. **메모리 관리**: Redis는 인메모리 DB이므로 메모리 사용량 모니터링 필수
2. **영구 저장**: AOF 또는 RDB 설정으로 데이터 영구 보존
3. **키 만료**: 불필요한 세션 데이터는 TTL 설정
4. **인덱스 관리**: 이메일/유저명 변경 시 인덱스 동기화 필수
5. **분산 환경**: Redis Cluster 또는 Sentinel 구성 고려
