# Read Models & Projections Example

이 예제는 CQRS 패턴의 핵심인 Read Model과 Projection을 구현하는 방법을 보여줍니다.

## 📖 학습 목표

- Read Model과 Write Model의 분리
- Event-driven Projection 구현
- 다양한 View 생성 및 최적화
- TTL을 활용한 캐시 관리

## 🏗️ 아키텍처

```
Write Side (Command)          Read Side (Query)
├── User Aggregate           ├── UserView
├── Order Aggregate          ├── OrderSummaryView
└── Product Aggregate        ├── CustomerOrderHistoryView
                            ├── ProductPopularityView
                            └── DashboardView

Event Flow:
Events → Event Handlers → Read Models → MongoDB Read Store
```

## 📁 파일 구조

```
04-read-models/
├── README.md
├── main.go                        # 메인 데모 프로그램
├── domain/
│   ├── user.go                   # User Aggregate
│   ├── order.go                  # Order Aggregate
│   ├── product.go                # Product Aggregate
│   └── events.go                 # 도메인 이벤트들
├── readmodels/
│   ├── user_view.go              # 사용자 뷰
│   ├── order_summary_view.go     # 주문 요약 뷰
│   ├── customer_history_view.go  # 고객 주문 이력 뷰
│   ├── product_popularity_view.go # 상품 인기도 뷰
│   └── dashboard_view.go         # 대시보드 뷰
├── projections/
│   ├── user_projection.go        # 사용자 프로젝션
│   ├── order_projection.go       # 주문 프로젝션
│   ├── analytics_projection.go   # 분석 프로젝션
│   └── projection_manager.go     # 프로젝션 관리자
├── infrastructure/
│   ├── read_store_factory.go     # Read Store 팩토리
│   └── event_handlers.go         # 이벤트 핸들러들
└── demo/
    ├── scenarios.go              # 데모 시나리오
    └── query_examples.go         # 쿼리 예제들
```

## 🚀 실행 방법

### 1. MongoDB 실행
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

### 2. 예제 실행
```bash
cd 04-read-models
go run main.go
```

### 3. 대화형 데모
```
Commands:
  # Write Operations (Commands)
  create-user <name> <email>           - 사용자 생성
  create-order <user-id> <product-id>  - 주문 생성
  add-product <name> <price>           - 상품 추가
  complete-order <order-id>            - 주문 완료
  
  # Read Operations (Queries)
  user <user-id>                       - 사용자 조회
  user-orders <user-id>                - 사용자 주문 이력
  order-summary <order-id>             - 주문 요약
  popular-products                     - 인기 상품 목록
  dashboard                            - 대시보드 데이터
  
  # Analytics
  sales-report <date>                  - 매출 리포트
  customer-stats                       - 고객 통계
  product-analytics                    - 상품 분석
  
  # Management
  rebuild-projections                  - 프로젝션 재구축
  projection-status                    - 프로젝션 상태
  clear-cache                          - 캐시 삭제
  help                                 - 도움말
  exit                                 - 종료
```

## 💡 핵심 개념

### 1. Read Model 정의
```go
// 사용자 뷰 - 기본 정보 + 통계
type UserView struct {
    *cqrs.BaseReadModel
    Name           string    `json:"name"`
    Email          string    `json:"email"`
    TotalOrders    int       `json:"total_orders"`
    TotalSpent     decimal.Decimal `json:"total_spent"`
    LastOrderDate  *time.Time `json:"last_order_date"`
    IsVIP          bool      `json:"is_vip"`
    CreatedAt      time.Time `json:"created_at"`
}

// 주문 요약 뷰 - 복잡한 계산 결과 저장
type OrderSummaryView struct {
    *cqrs.BaseReadModel
    CustomerID     string          `json:"customer_id"`
    CustomerName   string          `json:"customer_name"`
    Items          []OrderItemView `json:"items"`
    SubTotal       decimal.Decimal `json:"sub_total"`
    TaxAmount      decimal.Decimal `json:"tax_amount"`
    DiscountAmount decimal.Decimal `json:"discount_amount"`
    TotalAmount    decimal.Decimal `json:"total_amount"`
    Status         string          `json:"status"`
    OrderDate      time.Time       `json:"order_date"`
}
```

### 2. Event-driven Projection
```go
type UserProjection struct {
    readStore cqrsx.ReadStore
}

func (p *UserProjection) Handle(ctx context.Context, event cqrs.EventMessage) error {
    switch e := event.EventData().(type) {
    case *UserCreated:
        return p.handleUserCreated(ctx, e)
    case *OrderCompleted:
        return p.handleOrderCompleted(ctx, e)
    case *UserUpdated:
        return p.handleUserUpdated(ctx, e)
    }
    return nil
}

func (p *UserProjection) handleUserCreated(ctx context.Context, event *UserCreated) error {
    userView := &UserView{
        BaseReadModel: cqrs.NewBaseReadModel(event.UserID, "UserView", nil),
        Name:          event.Name,
        Email:         event.Email,
        TotalOrders:   0,
        TotalSpent:    decimal.Zero,
        IsVIP:         false,
        CreatedAt:     time.Now(),
    }
    
    return p.readStore.Save(ctx, userView)
}

func (p *UserProjection) handleOrderCompleted(ctx context.Context, event *OrderCompleted) error {
    // 사용자 뷰 업데이트
    userView, err := p.readStore.GetByID(ctx, event.CustomerID, "UserView")
    if err != nil {
        return err
    }
    
    if uv, ok := userView.(*UserView); ok {
        uv.TotalOrders++
        uv.TotalSpent = uv.TotalSpent.Add(event.TotalAmount)
        uv.LastOrderDate = &event.CompletedAt
        uv.IsVIP = uv.TotalSpent.GreaterThan(decimal.NewFromInt(1000))
        
        return p.readStore.Save(ctx, uv)
    }
    
    return nil
}
```

### 3. 복잡한 쿼리 최적화
```go
// 고객 주문 이력 뷰 - 비정규화된 데이터
type CustomerOrderHistoryView struct {
    *cqrs.BaseReadModel
    CustomerID    string                    `json:"customer_id"`
    CustomerName  string                    `json:"customer_name"`
    Orders        []OrderHistoryItem        `json:"orders"`
    TotalOrders   int                       `json:"total_orders"`
    TotalSpent    decimal.Decimal          `json:"total_spent"`
    AverageOrder  decimal.Decimal          `json:"average_order"`
    LastOrderDate time.Time                `json:"last_order_date"`
    Tags          []string                  `json:"tags"` // VIP, Frequent, etc.
}

// MongoDB 쿼리 최적화
func (rs *MongoReadStore) GetCustomerOrderHistory(ctx context.Context, customerID string) (*CustomerOrderHistoryView, error) {
    criteria := cqrs.QueryCriteria{
        Filters: map[string]interface{}{
            "customer_id": customerID,
            "model_type":  "CustomerOrderHistoryView",
        },
    }
    
    results, err := rs.Query(ctx, criteria)
    if err != nil {
        return nil, err
    }
    
    if len(results) > 0 {
        if view, ok := results[0].(*CustomerOrderHistoryView); ok {
            return view, nil
        }
    }
    
    return nil, cqrs.NewCQRSError("NOT_FOUND", "customer order history not found", nil)
}
```

### 4. TTL을 활용한 캐시 관리
```go
// 대시보드 뷰 - 5분 TTL
type DashboardView struct {
    *cqrs.BaseReadModel
    TotalUsers       int             `json:"total_users"`
    TotalOrders      int             `json:"total_orders"`
    TodayRevenue     decimal.Decimal `json:"today_revenue"`
    PopularProducts  []ProductStats  `json:"popular_products"`
    RecentOrders     []OrderSummary  `json:"recent_orders"`
    GeneratedAt      time.Time       `json:"generated_at"`
}

func (dv *DashboardView) GetTTL() time.Duration {
    return 5 * time.Minute // 5분 후 자동 삭제
}

// 캐시 갱신 로직
func (p *DashboardProjection) RefreshDashboard(ctx context.Context) error {
    dashboard := &DashboardView{
        BaseReadModel:   cqrs.NewBaseReadModel("dashboard", "DashboardView", nil),
        TotalUsers:      p.getTotalUsers(ctx),
        TotalOrders:     p.getTotalOrders(ctx),
        TodayRevenue:    p.getTodayRevenue(ctx),
        PopularProducts: p.getPopularProducts(ctx),
        RecentOrders:    p.getRecentOrders(ctx),
        GeneratedAt:     time.Now(),
    }
    
    return p.readStore.Save(ctx, dashboard)
}
```

## 🔍 데모 시나리오

### 시나리오 1: 기본 CQRS 플로우
1. 사용자 생성 → UserView 자동 생성
2. 상품 추가 → ProductView 자동 생성
3. 주문 생성 → OrderSummaryView 생성
4. 주문 완료 → 모든 관련 뷰 업데이트

### 시나리오 2: 복잡한 분석 뷰
1. 여러 주문 생성
2. 고객 주문 이력 뷰 확인
3. 상품 인기도 분석
4. 대시보드 데이터 확인

### 시나리오 3: 프로젝션 재구축
1. 기존 데이터 생성
2. 프로젝션 로직 변경
3. 전체 프로젝션 재구축
4. 새로운 뷰 데이터 확인

## 📊 MongoDB 컬렉션 구조

### read_models 컬렉션
```json
{
  "_id": ObjectId("..."),
  "model_id": "user-123",
  "model_type": "UserView",
  "data": {
    "name": "John Doe",
    "email": "john@example.com",
    "total_orders": 5,
    "total_spent": "1250.00",
    "is_vip": true
  },
  "version": 3,
  "updated_at": ISODate("2024-01-01T12:00:00Z"),
  "ttl": null
}
```

### 인덱스 최적화
```javascript
// 기본 인덱스
db.read_models.createIndex({ "model_id": 1, "model_type": 1 }, { unique: true })

// 쿼리 최적화 인덱스
db.read_models.createIndex({ "model_type": 1, "updated_at": -1 })
db.read_models.createIndex({ "data.customer_id": 1 })
db.read_models.createIndex({ "data.is_vip": 1 })
db.read_models.createIndex({ "data.order_date": -1 })

// TTL 인덱스
db.read_models.createIndex({ "ttl": 1 }, { expireAfterSeconds: 0 })
```

## 🧪 테스트

```bash
# 기본 테스트
go test ./...

# 프로젝션 테스트
go test -run TestProjections ./...

# 성능 테스트
go test -bench=BenchmarkQuery ./...

# 통합 테스트
go test -tags=integration ./...
```

## ⚙️ 고급 기능

### 1. 프로젝션 상태 관리
```go
type ProjectionState struct {
    Name           string    `json:"name"`
    LastEventID    string    `json:"last_event_id"`
    LastProcessed  time.Time `json:"last_processed"`
    EventsProcessed int64    `json:"events_processed"`
    Status         string    `json:"status"` // running, stopped, error
    ErrorMessage   string    `json:"error_message,omitempty"`
}
```

### 2. 배치 업데이트
```go
func (p *OrderProjection) ProcessEventBatch(ctx context.Context, events []cqrs.EventMessage) error {
    updates := make(map[string]*OrderSummaryView)
    
    for _, event := range events {
        if view := p.processEvent(event); view != nil {
            updates[view.GetID()] = view
        }
    }
    
    // 배치로 저장
    return p.readStore.SaveBatch(ctx, updates)
}
```

### 3. 이벤트 재생 (Event Replay)
```go
func (pm *ProjectionManager) RebuildProjection(ctx context.Context, projectionName string) error {
    // 1. 기존 Read Model 삭제
    err := pm.clearProjectionData(ctx, projectionName)
    if err != nil {
        return err
    }
    
    // 2. 모든 이벤트 재생
    events, err := pm.eventStore.GetAllEvents(ctx)
    if err != nil {
        return err
    }
    
    projection := pm.getProjection(projectionName)
    for _, event := range events {
        err = projection.Handle(ctx, event)
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 🔗 다음 단계

1. [Saga Pattern](../05-saga-pattern/) - 복잡한 비즈니스 프로세스
2. [Performance](../07-performance/) - 성능 최적화
3. [Testing](../09-testing/) - 테스트 전략

## 💡 모범 사례

1. **단일 책임**: 각 Read Model은 특정 용도에 최적화
2. **비정규화**: 쿼리 성능을 위해 데이터 중복 허용
3. **이벤트 순서**: 이벤트 처리 순서 보장
4. **에러 처리**: 프로젝션 실패 시 재시도 로직
5. **모니터링**: 프로젝션 지연 및 오류 모니터링
6. **캐시 전략**: TTL과 수동 갱신의 적절한 조합
