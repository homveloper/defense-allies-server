# Cargo Transport System - Event Sourcing CQRS Example

이 예제는 Defense Allies CQRS 인프라를 사용한 이벤트 소싱 기반의 화물 운송 시스템입니다.

## 📦 도메인 모델

### **화물 (Cargo)**
- 여러 이송품을 운송하는 컨테이너
- 출발지에서 목적지로 운송
- 이송품 적재/하차 관리
- 운송 상태 추적

### **이송품 (Shipment)**
- 화물에 포함되는 개별 물품
- 무게, 부피, 종류 등의 속성
- 적재/하차 시간 추적

## 🎯 주요 기능

### **화물 관리**
- 화물 생성 및 등록
- 이송품 적재 (LoadShipment)
- 이송품 하차 (UnloadShipment)
- 운송 시작 (StartTransport)
- 운송 완료 (CompleteTransport)

### **운송 추적**
- 실시간 화물 위치 추적
- 운송 상태 모니터링
- 이송품별 처리 시간 계산

## 🏗️ CQRS 아키텍처

### **Command Side (Write)**
```
Commands → CommandHandlers → Aggregates → Events → EventStore
```

### **Query Side (Read)**
```
Events → EventHandlers → Projections → ReadModels → Queries
```

## 📁 프로젝트 구조

```
server/examples/cargo/
├── domain/
│   ├── cargo_aggregate.go          # 화물 Aggregate
│   ├── shipment.go                 # 이송품 Value Object
│   └── events/
│       ├── cargo_created.go        # 화물 생성 이벤트
│       ├── shipment_loaded.go      # 이송품 적재 이벤트
│       ├── shipment_unloaded.go    # 이송품 하차 이벤트
│       ├── transport_started.go    # 운송 시작 이벤트
│       └── transport_completed.go  # 운송 완료 이벤트
├── application/
│   ├── commands/
│   │   ├── create_cargo.go         # 화물 생성 커맨드
│   │   ├── load_shipment.go        # 이송품 적재 커맨드
│   │   ├── unload_shipment.go      # 이송품 하차 커맨드
│   │   ├── start_transport.go      # 운송 시작 커맨드
│   │   └── complete_transport.go   # 운송 완료 커맨드
│   └── handlers/
│       └── cargo_command_handler.go # 화물 커맨드 핸들러
├── infrastructure/
│   ├── projections/
│   │   ├── cargo_summary.go        # 화물 요약 프로젝션
│   │   └── transport_tracking.go   # 운송 추적 프로젝션
│   └── queries/
│       ├── get_cargo_details.go    # 화물 상세 조회
│       └── get_transport_status.go # 운송 상태 조회
├── main.go                         # 메인 실행 파일
└── README.md                       # 이 파일
```

## 🚀 실행 방법

```bash
cd server/examples/cargo
go run main.go
```

## 📊 이벤트 소싱 플로우

1. **화물 생성**: `CreateCargoCommand` → `CargoCreatedEvent`
2. **이송품 적재**: `LoadShipmentCommand` → `ShipmentLoadedEvent`
3. **운송 시작**: `StartTransportCommand` → `TransportStartedEvent`
4. **이송품 하차**: `UnloadShipmentCommand` → `ShipmentUnloadedEvent`
5. **운송 완료**: `CompleteTransportCommand` → `TransportCompletedEvent`

각 이벤트는 EventStore에 저장되고, Projection을 통해 ReadModel이 업데이트됩니다.

## 🎮 Defense Allies CQRS 활용

이 예제는 다음 Defense Allies CQRS 컴포넌트들을 활용합니다:

- `DomainEventMessage` - 도메인 이벤트 정의
- `AggregateRoot` - 화물 Aggregate 구현
- `CommandHandler` - 커맨드 처리
- `EventHandler` - 이벤트 처리 및 프로젝션
- `Repository` - 이벤트 소싱 저장소
- `QueryDispatcher` - 쿼리 처리

## 📈 확장 가능성

- 다중 화물 운송 최적화
- 실시간 위치 추적 (GPS)
- 화물 보험 및 손해 처리
- 운송 비용 계산
- 고객 알림 시스템
