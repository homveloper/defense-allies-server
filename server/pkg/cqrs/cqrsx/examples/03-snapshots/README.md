# Snapshots Example

이 예제는 스냅샷을 활용한 Event Sourcing 성능 최적화 방법을 보여줍니다.

## 📖 학습 목표

- 스냅샷의 개념과 필요성 이해
- 커스텀 스냅샷 직렬화 구현
- 스냅샷 정책 및 전략 수립
- 성능 최적화 효과 측정

## 🏗️ 아키텍처

```
Order Aggregate (복잡한 상태)
├── OrderCreated Event
├── ItemAdded Event (여러 개)
├── ItemRemoved Event
├── DiscountApplied Event
├── ShippingUpdated Event
└── OrderCompleted Event

Snapshot Strategy
├── 이벤트 10개마다 스냅샷 생성
├── 복원 시 최신 스냅샷 + 이후 이벤트
└── 성능 비교 (스냅샷 vs 전체 이벤트 재생)
```

## 📁 파일 구조

```
03-snapshots/
├── README.md
├── main.go                     # 메인 데모 프로그램
├── domain/
│   ├── order.go               # Order Aggregate (복잡한 상태)
│   ├── events.go              # Order 관련 이벤트들
│   └── snapshot.go            # 커스텀 스냅샷 구현
├── infrastructure/
│   ├── snapshot_policy.go     # 스냅샷 정책
│   └── performance_monitor.go # 성능 측정
└── demo/
    ├── performance_demo.go    # 성능 비교 데모
    └── scenarios.go           # 다양한 시나리오
```

## 🚀 실행 방법

### 1. MongoDB 실행
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

### 2. 예제 실행
```bash
cd 03-snapshots
go run main.go
```

### 3. 대화형 데모
```
Commands:
  create <customer-id>              - 새 주문 생성
  add-item <order-id> <name> <qty>  - 상품 추가
  remove-item <order-id> <name>     - 상품 제거
  apply-discount <order-id> <rate>  - 할인 적용
  complete <order-id>               - 주문 완료
  get <order-id>                    - 주문 조회
  snapshot <order-id>               - 수동 스냅샷 생성
  restore-test <order-id>           - 복원 성능 테스트
  benchmark                         - 성능 벤치마크
  policy <events|time|size>         - 스냅샷 정책 변경
  stats                             - 통계 정보
  clear                             - 모든 데이터 삭제
  help                              - 도움말
  exit                              - 종료
```

## 💡 핵심 개념

### 1. 복잡한 Aggregate (Order)
```go
type Order struct {
    *cqrs.BaseAggregate
    CustomerID    string
    Items         map[string]*OrderItem
    TotalAmount   decimal.Decimal
    DiscountRate  decimal.Decimal
    ShippingInfo  *ShippingInfo
    Status        OrderStatus
    CreatedAt     time.Time
    CompletedAt   *time.Time
}

type OrderItem struct {
    Name     string
    Quantity int
    Price    decimal.Decimal
    Total    decimal.Decimal
}
```

### 2. 커스텀 스냅샷 구현
```go
type OrderSnapshot struct {
    AggregateID   string                 `json:"aggregate_id"`
    Version       int                    `json:"version"`
    CustomerID    string                 `json:"customer_id"`
    Items         map[string]*OrderItem  `json:"items"`
    TotalAmount   string                 `json:"total_amount"` // decimal as string
    DiscountRate  string                 `json:"discount_rate"`
    ShippingInfo  *ShippingInfo         `json:"shipping_info"`
    Status        string                 `json:"status"`
    CreatedAt     time.Time             `json:"created_at"`
    CompletedAt   *time.Time            `json:"completed_at"`
    Timestamp     time.Time             `json:"timestamp"`
}

func (o *Order) CreateSnapshot() (cqrs.Snapshot, error) {
    return &OrderSnapshot{
        AggregateID:  o.ID(),
        Version:      o.Version(),
        CustomerID:   o.CustomerID,
        Items:        o.Items,
        TotalAmount:  o.TotalAmount.String(),
        DiscountRate: o.DiscountRate.String(),
        ShippingInfo: o.ShippingInfo,
        Status:       string(o.Status),
        CreatedAt:    o.CreatedAt,
        CompletedAt:  o.CompletedAt,
        Timestamp:    time.Now(),
    }, nil
}
```

### 3. 스냅샷 정책
```go
type SnapshotPolicy interface {
    ShouldCreateSnapshot(aggregate cqrs.Aggregate) bool
}

// 이벤트 개수 기반
type EventCountPolicy struct {
    EventThreshold int
}

func (p *EventCountPolicy) ShouldCreateSnapshot(aggregate cqrs.Aggregate) bool {
    return aggregate.Version() > 0 && aggregate.Version()%p.EventThreshold == 0
}

// 시간 기반
type TimeBasedPolicy struct {
    TimeThreshold time.Duration
    lastSnapshot  map[string]time.Time
}

// 크기 기반
type SizeBasedPolicy struct {
    SizeThreshold int64
}
```

### 4. 성능 최적화된 복원
```go
func (r *OrderRepository) GetByID(ctx context.Context, id string) (*Order, error) {
    // 1. 최신 스냅샷 조회
    snapshot, err := r.snapshotStore.GetSnapshot(ctx, id, "Order")
    
    var order *Order
    var fromVersion int
    
    if err == nil && snapshot != nil {
        // 스냅샷에서 복원
        order, err = r.restoreFromSnapshot(snapshot)
        if err != nil {
            return nil, err
        }
        fromVersion = order.Version() + 1
    } else {
        // 새 인스턴스 생성
        order = domain.NewOrder()
        fromVersion = 1
    }
    
    // 2. 스냅샷 이후 이벤트들만 조회
    events, err := r.eventStore.GetEventHistory(ctx, id, "Order", fromVersion)
    if err != nil {
        return nil, err
    }
    
    // 3. 이벤트 적용
    for _, event := range events {
        order.Apply(event)
    }
    
    return order, nil
}
```

## 🔍 데모 시나리오

### 시나리오 1: 기본 스냅샷 생성
1. 주문 생성
2. 여러 상품 추가 (10개 이상)
3. 자동 스냅샷 생성 확인
4. 스냅샷에서 복원 테스트

### 시나리오 2: 성능 비교
1. 대량 이벤트 생성 (100개+)
2. 스냅샷 없이 복원 시간 측정
3. 스냅샷 생성 후 복원 시간 측정
4. 성능 개선 효과 확인

### 시나리오 3: 다양한 스냅샷 정책
1. 이벤트 개수 기반 정책 테스트
2. 시간 기반 정책 테스트
3. 크기 기반 정책 테스트
4. 정책별 효과 비교

## 📊 성능 측정 결과 예시

```
Performance Benchmark Results:
================================

Order ID: order-123
Total Events: 150

Without Snapshot:
- Restoration Time: 45.2ms
- Events Processed: 150
- Memory Usage: 2.1MB

With Snapshot (every 10 events):
- Snapshot Load Time: 2.1ms
- Events Processed: 5 (from snapshot)
- Total Restoration Time: 7.3ms
- Memory Usage: 0.8MB
- Performance Improvement: 83.8%

Snapshot Storage:
- Snapshot Size: 1.2KB
- Compression Ratio: 65%
- Storage Overhead: 0.8%
```

## 🧪 테스트

```bash
# 기본 테스트
go test ./...

# 성능 테스트
go test -bench=. ./...

# 스냅샷 정책 테스트
go test -run TestSnapshotPolicy ./...

# 메모리 사용량 테스트
go test -memprofile=mem.prof ./...
```

## ⚙️ 설정 옵션

### 스냅샷 정책 설정
```go
type SnapshotConfig struct {
    Policy           string        `json:"policy"`           // "events", "time", "size"
    EventThreshold   int           `json:"event_threshold"`  // 이벤트 개수
    TimeThreshold    time.Duration `json:"time_threshold"`   // 시간 간격
    SizeThreshold    int64         `json:"size_threshold"`   // 크기 임계값
    CompressionLevel int           `json:"compression"`      // 압축 레벨
    RetentionDays    int           `json:"retention_days"`   // 보관 기간
}
```

### MongoDB 인덱스 최적화
```javascript
// snapshots 컬렉션 인덱스
db.snapshots.createIndex(
    { "aggregate_id": 1, "aggregate_type": 1 }, 
    { unique: true }
)
db.snapshots.createIndex(
    { "timestamp": -1 }
)
db.snapshots.createIndex(
    { "ttl": 1 }, 
    { expireAfterSeconds: 0 }
)
```

## 🔧 고급 기능

### 1. 스냅샷 압축
```go
type CompressedSnapshot struct {
    Data         []byte    `json:"data"`
    Compression  string    `json:"compression"`
    OriginalSize int64     `json:"original_size"`
}

func CompressSnapshot(snapshot cqrs.Snapshot) (*CompressedSnapshot, error) {
    data, err := json.Marshal(snapshot)
    if err != nil {
        return nil, err
    }
    
    compressed := gzip.Compress(data)
    return &CompressedSnapshot{
        Data:         compressed,
        Compression:  "gzip",
        OriginalSize: int64(len(data)),
    }, nil
}
```

### 2. 스냅샷 검증
```go
func ValidateSnapshot(snapshot cqrs.Snapshot, events []cqrs.EventMessage) error {
    // 스냅샷에서 복원한 상태와 이벤트 재생 결과 비교
    // 데이터 무결성 검증
}
```

### 3. 스냅샷 마이그레이션
```go
func MigrateSnapshots(oldVersion, newVersion int) error {
    // 스냅샷 스키마 변경 시 마이그레이션
    // 버전별 호환성 처리
}
```

## 🔗 다음 단계

1. [Read Models](../04-read-models/) - Read Model과 Projection
2. [Performance](../07-performance/) - 고급 성능 최적화
3. [Event Store Patterns](../08-event-store-patterns/) - 고급 패턴

## 💡 모범 사례

1. **적절한 스냅샷 주기**: 너무 자주 생성하면 저장소 부담, 너무 드물면 성능 저하
2. **압축 활용**: 큰 Aggregate의 경우 압축으로 저장 공간 절약
3. **비동기 생성**: 스냅샷 생성을 비동기로 처리하여 응답 시간 개선
4. **검증 로직**: 스냅샷 무결성 검증 로직 포함
5. **모니터링**: 스냅샷 생성 빈도와 성능 지표 모니터링
