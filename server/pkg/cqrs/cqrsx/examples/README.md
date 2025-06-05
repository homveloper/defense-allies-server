# CQRS Infrastructure Examples

이 폴더는 `cqrsx` 패키지를 사용한 MongoDB 기반 Event Sourcing과 CQRS 패턴의 실제 구현 예제들을 제공합니다.

## 📋 예제 목록

### 1. Basic Event Sourcing (`01-basic-event-sourcing/`)
**목적**: 가장 기본적인 이벤트 소싱 패턴 구현
- 단일 Aggregate (User) 구현
- 기본 이벤트 생성, 저장, 복원
- MongoDB Event Store 기본 사용법
- 기본 컬렉션 명 사용 (`events`, `snapshots`)

**포함 내용**:
- User Aggregate 구현
- UserCreated, UserUpdated 이벤트
- 기본 Event Store 설정
- 간단한 CLI 데모

### 2. Custom Collection Names (`02-custom-collections/`)
**목적**: 컬렉션 명 커스터마이징 방법 시연
- Prefix 사용 예제
- 완전 커스텀 컬렉션 명 사용
- 멀티 테넌트 환경 시뮬레이션

**포함 내용**:
- 여러 테넌트별 컬렉션 분리
- 환경별 컬렉션 명 설정 (dev, staging, prod)
- 설정 파일 기반 컬렉션 명 관리

### 3. Snapshots (`03-snapshots/`)
**목적**: 스냅샷 기능을 활용한 성능 최적화
- 스냅샷 생성 및 복원
- 커스텀 스냅샷 직렬화
- 스냅샷 정책 구현

**포함 내용**:
- Order Aggregate (복잡한 상태)
- 자동 스냅샷 생성 로직
- 스냅샷 기반 빠른 복원
- 성능 비교 데모

### 4. Read Models & Projections (`04-read-models/`)
**목적**: Read Model과 Projection 패턴 구현
- 이벤트 기반 Read Model 업데이트
- 다양한 View 생성
- MongoDB Read Store 활용

**포함 내용**:
- UserView, OrderSummaryView 구현
- Event Handler를 통한 자동 업데이트
- 복잡한 쿼리 최적화
- TTL을 활용한 캐시 관리

### 5. Multi-Aggregate Saga (`05-saga-pattern/`)
**목적**: 여러 Aggregate 간의 복잡한 비즈니스 프로세스
- Saga 패턴 구현
- 분산 트랜잭션 시뮬레이션
- 보상 트랜잭션 (Compensation)

**포함 내용**:
- Order, Payment, Inventory Aggregates
- OrderProcessingSaga 구현
- 실패 시나리오 및 롤백
- 이벤트 기반 상태 머신

### 6. Event Versioning (`06-event-versioning/`)
**목적**: 이벤트 스키마 진화 및 버전 관리
- 이벤트 스키마 변경 처리
- 하위 호환성 유지
- 마이그레이션 전략

**포함 내용**:
- 이벤트 V1, V2 구현
- 업캐스팅/다운캐스팅
- 점진적 마이그레이션
- 버전별 직렬화

### 7. Performance Optimization (`07-performance/`)
**목적**: 대용량 데이터 처리 및 성능 최적화
- 배치 처리
- 인덱스 최적화
- 메모리 효율성

**포함 내용**:
- 대량 이벤트 배치 저장
- 스트리밍 이벤트 처리
- MongoDB 인덱스 전략
- 성능 측정 및 모니터링

### 8. Event Store Patterns (`08-event-store-patterns/`)
**목적**: 고급 Event Store 패턴들
- 이벤트 스트림 분할
- 아카이빙 전략
- 이벤트 압축

**포함 내용**:
- 시간 기반 파티셔닝
- 콜드 스토리지 이동
- 이벤트 중복 제거
- 스토리지 최적화

### 9. Testing Strategies (`09-testing/`)
**목적**: Event Sourcing 시스템의 테스트 전략
- 단위 테스트 패턴
- 통합 테스트
- 이벤트 기반 테스트

**포함 내용**:
- Given-When-Then 패턴
- 이벤트 스토어 모킹
- 시간 여행 테스트
- 성능 테스트

### 10. Microservices Integration (`10-microservices/`)
**목적**: 마이크로서비스 환경에서의 Event Sourcing
- 서비스 간 이벤트 공유
- 이벤트 발행/구독
- 분산 시스템 패턴

**포함 내용**:
- 여러 서비스 시뮬레이션
- 이벤트 버스 구현
- 서비스 간 통신
- 장애 격리

## 🚀 시작하기

### 전제 조건
- Go 1.21+
- MongoDB 4.4+
- Docker (선택사항)

### 환경 설정
```bash
# MongoDB 실행 (Docker 사용 시)
docker run -d -p 27017:27017 --name mongodb mongo:latest

# 또는 로컬 MongoDB 설치 후 실행
mongod --dbpath /your/data/path
```

### 예제 실행
각 예제 폴더로 이동하여 README.md의 지시사항을 따르세요:

```bash
cd 01-basic-event-sourcing
go run main.go
```

## 📚 학습 순서 권장사항

1. **초보자**: 01 → 02 → 03 → 04
2. **중급자**: 05 → 06 → 07 → 09
3. **고급자**: 08 → 10

## 🔧 공통 유틸리티

각 예제에서 공통으로 사용하는 유틸리티들:

### MongoDB 설정
```go
// 기본 설정
config := &cqrsx.MongoConfig{
    URI:      "mongodb://localhost:27017",
    Database: "cqrs_examples",
}

// 개발 환경용 (prefix 사용)
client, err := cqrsx.NewMongoClientManagerWithPrefix(config, "dev")

// 프로덕션 환경용 (커스텀 컬렉션)
customNames := &cqrsx.CollectionNames{
    Events:     "production_events",
    Snapshots:  "production_snapshots", 
    ReadModels: "production_read_models",
}
client, err := cqrsx.NewMongoClientManagerWithCollections(config, "", customNames)
```

### 이벤트 직렬화
```go
// JSON 직렬화 (기본)
serializer := &cqrsx.JSONEventSerializer{}

// 압축 직렬화 (고성능)
serializer := &cqrsx.CompactEventSerializer{}
```

## 🤝 기여하기

새로운 예제나 개선사항이 있으시면 언제든 기여해주세요!

1. 새로운 예제 폴더 생성
2. README.md와 완전한 코드 예제 작성
3. 테스트 코드 포함
4. 실행 가능한 데모 제공

## 📖 추가 자료

- [CQRS 패턴 가이드](../README.md)
- [MongoDB Event Store 문서](../mongo_event_store.go)
- [Event Sourcing 모범 사례](https://docs.microsoft.com/en-us/azure/architecture/patterns/event-sourcing)

## 🎯 예제별 난이도

| 예제                    | 난이도 | 소요시간 | 전제지식            |
| ----------------------- | ------ | -------- | ------------------- |
| 01-basic-event-sourcing | ⭐      | 30분     | Go 기초             |
| 02-custom-collections   | ⭐⭐     | 45분     | MongoDB 기초        |
| 03-snapshots            | ⭐⭐     | 1시간    | Event Sourcing 개념 |
| 04-read-models          | ⭐⭐⭐    | 1.5시간  | CQRS 패턴           |
| 05-saga-pattern         | ⭐⭐⭐⭐   | 2시간    | 분산 시스템         |
| 06-event-versioning     | ⭐⭐⭐    | 1.5시간  | 스키마 진화         |
| 07-performance          | ⭐⭐⭐⭐   | 2시간    | 성능 최적화         |
| 08-event-store-patterns | ⭐⭐⭐⭐⭐  | 3시간    | 고급 패턴           |
| 09-testing              | ⭐⭐⭐    | 1시간    | 테스트 전략         |
| 10-microservices        | ⭐⭐⭐⭐⭐  | 3시간    | 마이크로서비스      |
