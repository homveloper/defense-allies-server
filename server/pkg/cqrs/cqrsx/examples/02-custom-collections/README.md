# Custom Collection Names Example

이 예제는 MongoDB 컬렉션 명을 커스터마이징하는 다양한 방법을 보여줍니다.

## 📖 학습 목표

- 컬렉션 명 prefix 사용법
- 완전 커스텀 컬렉션 명 설정
- 멀티 테넌트 환경 구현
- 환경별 컬렉션 분리 전략

## 🏗️ 아키텍처

```
Multi-Tenant Architecture
├── Tenant A
│   ├── tenant_a_events
│   ├── tenant_a_snapshots
│   └── tenant_a_read_models
├── Tenant B  
│   ├── tenant_b_events
│   ├── tenant_b_snapshots
│   └── tenant_b_read_models
└── Environment Separation
    ├── dev_* collections
    ├── staging_* collections
    └── prod_* collections
```

## 📁 파일 구조

```
02-custom-collections/
├── README.md
├── main.go                    # 메인 데모 프로그램
├── config/
│   ├── environments.go        # 환경별 설정
│   └── tenants.go            # 테넌트 설정
├── domain/
│   ├── product.go            # Product Aggregate
│   └── events.go             # Product 관련 이벤트들
├── infrastructure/
│   ├── tenant_manager.go     # 테넌트 관리
│   └── collection_factory.go # 컬렉션 팩토리
└── demo/
    └── multi_tenant_demo.go  # 멀티 테넌트 데모
```

## 🚀 실행 방법

### 1. MongoDB 실행
```bash
docker run -d -p 27017:27017 --name mongodb mongo:latest
```

### 2. 예제 실행
```bash
cd 02-custom-collections
go run main.go
```

### 3. 대화형 데모
```
Commands:
  env <dev|staging|prod>           - 환경 전환
  tenant <tenant-id>               - 테넌트 전환
  create <name> <price>            - 제품 생성
  update <id> <price>              - 제품 가격 업데이트
  get <id>                         - 제품 조회
  list                             - 현재 테넌트의 모든 제품
  collections                      - 현재 컬렉션 명 확인
  switch-demo                      - 데모 모드 전환
  clear                            - 현재 테넌트 데이터 삭제
  clear-all                        - 모든 데이터 삭제
  help                             - 도움말
  exit                             - 종료
```

## 💡 핵심 개념

### 1. Prefix 기반 컬렉션 명
```go
// 개발 환경
devClient, err := cqrsx.NewMongoClientManagerWithPrefix(config, "dev")
// 결과: dev_events, dev_snapshots, dev_read_models

// 테넌트별 분리
tenantClient, err := cqrsx.NewMongoClientManagerWithPrefix(config, "tenant_123")
// 결과: tenant_123_events, tenant_123_snapshots, tenant_123_read_models
```

### 2. 완전 커스텀 컬렉션 명
```go
customNames := &cqrsx.CollectionNames{
    Events:     "company_events_2024",
    Snapshots:  "company_snapshots_2024",
    ReadModels: "company_views_2024",
}

client, err := cqrsx.NewMongoClientManagerWithCollections(config, "", customNames)
```

### 3. 환경별 설정
```go
type EnvironmentConfig struct {
    Name            string
    CollectionNames *cqrsx.CollectionNames
    Database        string
}

var environments = map[string]*EnvironmentConfig{
    "dev": {
        Name:     "development",
        Database: "cqrs_dev",
        CollectionNames: &cqrsx.CollectionNames{
            Events:     "dev_events",
            Snapshots:  "dev_snapshots", 
            ReadModels: "dev_read_models",
        },
    },
    "prod": {
        Name:     "production",
        Database: "cqrs_production",
        CollectionNames: &cqrsx.CollectionNames{
            Events:     "production_events",
            Snapshots:  "production_snapshots",
            ReadModels: "production_read_models", 
        },
    },
}
```

### 4. 테넌트 관리자
```go
type TenantManager struct {
    clients map[string]*cqrsx.MongoClientManager
    config  *cqrsx.MongoConfig
}

func (tm *TenantManager) GetClient(tenantID string) (*cqrsx.MongoClientManager, error) {
    if client, exists := tm.clients[tenantID]; exists {
        return client, nil
    }
    
    // 새 테넌트 클라이언트 생성
    client, err := cqrsx.NewMongoClientManagerWithPrefix(tm.config, tenantID)
    if err != nil {
        return nil, err
    }
    
    tm.clients[tenantID] = client
    return client, nil
}
```

## 🔍 데모 시나리오

### 시나리오 1: 환경별 분리
1. 개발 환경에서 제품 생성
2. 스테이징 환경으로 전환
3. 같은 ID로 다른 제품 생성
4. 각 환경의 데이터 독립성 확인

### 시나리오 2: 멀티 테넌트
1. 테넌트 A에서 제품 생성
2. 테넌트 B로 전환
3. 같은 제품명으로 다른 제품 생성
4. 테넌트별 데이터 격리 확인

### 시나리오 3: 컬렉션 명 확인
1. 다양한 설정으로 클라이언트 생성
2. 실제 MongoDB 컬렉션 명 확인
3. 인덱스 생성 확인

## 📊 MongoDB 컬렉션 예시

### 개발 환경
```
cqrs_dev database:
├── dev_events
├── dev_snapshots
└── dev_read_models
```

### 프로덕션 환경
```
cqrs_production database:
├── production_events
├── production_snapshots
└── production_read_models
```

### 멀티 테넌트
```
cqrs_examples database:
├── tenant_abc_events
├── tenant_abc_snapshots
├── tenant_abc_read_models
├── tenant_xyz_events
├── tenant_xyz_snapshots
└── tenant_xyz_read_models
```

## ⚙️ 설정 파일 예시

### config.yaml
```yaml
environments:
  development:
    database: "cqrs_dev"
    collections:
      events: "dev_events"
      snapshots: "dev_snapshots"
      read_models: "dev_read_models"
  
  production:
    database: "cqrs_prod"
    collections:
      events: "events"
      snapshots: "snapshots"
      read_models: "read_models"

tenants:
  default_prefix: "tenant_"
  isolation_level: "collection"  # collection | database
```

## 🧪 테스트

```bash
# 기본 테스트
go test ./...

# 멀티 테넌트 테스트
go test -run TestMultiTenant ./...

# 환경별 테스트
go test -run TestEnvironments ./...
```

## 🔧 고급 사용법

### 1. 동적 컬렉션 명 생성
```go
func GenerateCollectionNames(tenantID, environment string) *cqrsx.CollectionNames {
    prefix := fmt.Sprintf("%s_%s", environment, tenantID)
    return &cqrsx.CollectionNames{
        Events:     prefix + "_events",
        Snapshots:  prefix + "_snapshots",
        ReadModels: prefix + "_read_models",
    }
}
```

### 2. 컬렉션 마이그레이션
```go
func MigrateCollections(oldClient, newClient *cqrsx.MongoClientManager) error {
    // 기존 컬렉션에서 새 컬렉션으로 데이터 이동
    // 인덱스 재생성
    // 검증
}
```

### 3. 컬렉션 모니터링
```go
func MonitorCollections(client *cqrsx.MongoClientManager) {
    names := client.GetCollectionNames()
    log.Printf("Monitoring collections: %+v", names)
    
    // 컬렉션 크기, 인덱스 상태 등 모니터링
}
```

## 🔗 다음 단계

1. [Snapshots](../03-snapshots/) - 스냅샷 기능 활용
2. [Read Models](../04-read-models/) - Read Model 패턴
3. [Performance](../07-performance/) - 성능 최적화

## 💡 모범 사례

1. **환경별 분리**: 개발/스테이징/프로덕션 환경별로 다른 데이터베이스 사용
2. **테넌트 격리**: 민감한 데이터의 경우 데이터베이스 레벨 분리 고려
3. **네이밍 컨벤션**: 일관된 명명 규칙 사용
4. **모니터링**: 컬렉션별 성능 및 크기 모니터링
5. **백업**: 테넌트별 백업 전략 수립
