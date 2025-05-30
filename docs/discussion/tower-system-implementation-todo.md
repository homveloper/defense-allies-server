# Defense Allies 타워 시스템 구현 TODO 리스트

## 📋 프로젝트 개요
- **목표**: 레고 스타일 모듈형 타워 시스템 구현
- **예상 기간**: 6-8주
- **우선순위**: MVP → 확장 → 최적화

## � 현재 진행 상황 (2025-05-30)
**✅ Phase 1 완료**: 핵심 기반 시스템 구현 완료
- 모든 기본 컴포넌트 시스템 구현 및 테스트 통과
- JSON 기반 타워 정의 시스템 완료
- 컴포넌트 조립 엔진 및 실행 시스템 완료
- 타워 팩토리 시스템 완료 (로딩, 생성, 업그레이드, 실행)
- 4개 Human Alliance 타워 프리셋 완성

## �🎯 Phase 1: 핵심 기반 시스템 (2주)

### Week 1: 기본 컴포넌트 시스템

#### Day 1-2: 컴포넌트 인터페이스 설계
- [x] `AtomicComponent` 인터페이스 정의
- [x] `ComponentInput/Output` 타입 시스템 설계
- [x] `ComponentConnection` 구조체 정의
- [x] 기본 컴포넌트 타입 열거형 정의

**파일 생성 목록:**
```
pkg/tower/component/
├── interfaces.go          # 기본 인터페이스들
├── types.go              # 컴포넌트 타입 정의
├── connection.go         # 연결 시스템
└── registry.go           # 컴포넌트 레지스트리
```

#### Day 3-4: 기본 컴포넌트 구현
- [x] `TargetingComponent` 구현 (단일/다중/영역)
- [x] `DamageComponent` 구현 (기본/원소/특수)
- [x] `EffectComponent` 구현 (상태효과/버프/디버프)
- [x] `RangeComponent` 구현 (사거리/영역)

**파일 생성 목록:**
```
pkg/tower/component/impl/
├── targeting.go          # 타겟팅 컴포넌트들
├── damage.go            # 데미지 컴포넌트들
├── effect.go            # 효과 컴포넌트들
└── range.go             # 범위 컴포넌트들
```

#### Day 5-7: 컴포넌트 조립 엔진
- [x] `ComponentAssembly` 구조체 구현
- [x] `AssemblyEngine` 핵심 로직 구현
- [x] 컴포넌트 연결 유효성 검증
- [x] 실행 순서 계산 (토폴로지 정렬)

**파일 생성 목록:**
```
pkg/tower/assembly/
├── engine.go            # 조립 엔진 메인
├── validator.go         # 연결 유효성 검증
├── executor.go          # 실행 순서 계산
└── assembly.go          # 어셈블리 구조체
```

### Week 2: 데이터 시스템 및 기본 실행

#### Day 8-10: JSON 데이터 시스템
- [x] 타워 정의 JSON 스키마 설계
- [x] JSON → 컴포넌트 변환 로직
- [x] 데이터 로더 및 캐시 시스템
- [x] 기본 프리셋 데이터 작성

**파일 생성 목록:**
```
pkg/tower/data/
├── schema.go            # JSON 스키마 정의
├── loader.go            # 데이터 로더
├── cache.go             # 캐시 시스템
└── presets.go           # 프리셋 관리

data/towers/
├── presets/             # 기본 프리셋들
│   ├── basic_archer.json
│   ├── flame_tower.json
│   └── ice_tower.json
└── components/          # 컴포넌트 정의들
    ├── targeting.json
    ├── damage.json
    └── effects.json
```

#### Day 11-14: 기본 실행 엔진
- [ ] `TowerBehavior` 실행 시스템
- [ ] `ExecutionContext` 구현
- [ ] 기본 전투 루프 구현
- [ ] 단위 테스트 작성

**파일 생성 목록:**
```
pkg/tower/execution/
├── behavior.go          # 타워 행동 시스템
├── context.go           # 실행 컨텍스트
├── combat.go            # 전투 루프
└── executor.go          # 실행기

pkg/tower/execution/test/
├── behavior_test.go
├── combat_test.go
└── integration_test.go
```

## 🚀 Phase 2: 매트릭스 시스템 및 확장 (2주)

### Week 3: 매트릭스 기반 밸런싱

#### Day 15-17: 매트릭스 엔진 구현
- [ ] `Matrix` 기본 구조체 및 연산
- [ ] `MatrixOperation` 인터페이스 및 구현체들
- [ ] `MatrixEngine` 핵심 로직
- [ ] 종족별 기본 매트릭스 정의

**파일 생성 목록:**
```
pkg/tower/matrix/
├── matrix.go            # 매트릭스 기본 구조
├── operations.go        # 매트릭스 연산들
├── engine.go            # 매트릭스 엔진
└── race_matrices.go     # 종족별 매트릭스

pkg/math/
├── matrix.go            # 범용 매트릭스 유틸
└── operations.go        # 수학 연산 유틸
```

#### Day 18-21: 환경 시스템 통합
- [ ] 환경 효과 매트릭스 시스템
- [ ] 시너지 계산 매트릭스 연산
- [ ] 동적 밸런싱 시스템
- [ ] 매트릭스 기반 컴포넌트 수정

**파일 생성 목록:**
```
pkg/tower/environment/
├── effects.go           # 환경 효과 시스템
├── synergy.go           # 시너지 계산
└── balance.go           # 동적 밸런싱

data/environment/
├── weather_effects.json
├── terrain_effects.json
└── synergy_rules.json
```

### Week 4: 고급 컴포넌트 및 프리셋

#### Day 22-24: 고급 컴포넌트 구현
- [ ] `ConditionalComponent` (조건부 로직)
- [ ] `ChainComponent` (연쇄 공격)
- [x] `ProjectileComponent` (투사체 시스템)
- [ ] `SynergyComponent` (시너지 효과)

**파일 생성 목록:**
```
pkg/tower/component/impl/
├── conditional.go       # 조건부 컴포넌트
├── chain.go            # 연쇄 공격
├── projectile.go       # 투사체 시스템
└── synergy.go          # 시너지 컴포넌트
```

#### Day 25-28: 프리셋 및 템플릿 시스템
- [ ] 프리셋 자동 생성 시스템
- [ ] 템플릿 엔진 구현
- [ ] 18개 종족별 기본 타워 프리셋 작성
- [ ] 템플릿 기반 변형 생성 시스템

**파일 생성 목록:**
```
pkg/tower/template/
├── engine.go            # 템플릿 엔진
├── generator.go         # 자동 생성기
└── validator.go         # 템플릿 검증

data/towers/presets/races/
├── human_alliance/      # 휴먼 연합 타워들
├── elven_kingdom/       # 엘프 왕국 타워들
├── dwarven_clan/        # 드워프 클랜 타워들
└── ... (18개 종족)
```

## 🎮 Phase 3: 서버 통합 및 최적화 (2주)

### Week 5: 서버 시스템 통합

#### Day 29-31: TimeSquare 서버 통합
- [ ] 타워 시스템을 TimeSquare 서버에 통합
- [ ] 서버 기반 전투 처리 시스템
- [ ] 클라이언트 상태 동기화
- [ ] SSE 기반 실시간 업데이트

**파일 생성 목록:**
```
services/timesquare/tower/
├── service.go           # 타워 서비스
├── combat.go            # 서버 전투 시스템
├── sync.go              # 상태 동기화
└── events.go            # 이벤트 처리

services/timesquare/tower/handler/
├── tower_handler.go     # 타워 관련 핸들러
├── combat_handler.go    # 전투 핸들러
└── assembly_handler.go  # 조립 핸들러
```

#### Day 32-35: 성능 최적화
- [ ] 컴포넌트 풀링 시스템
- [ ] 공간 분할 최적화
- [ ] 배치 처리 시스템
- [ ] 메모리 사용량 최적화

**파일 생성 목록:**
```
pkg/tower/optimization/
├── pooling.go           # 객체 풀링
├── spatial.go           # 공간 분할
├── batch.go             # 배치 처리
└── memory.go            # 메모리 최적화

pkg/tower/benchmark/
├── component_bench_test.go
├── assembly_bench_test.go
└── combat_bench_test.go
```

### Week 6: 레고 스타일 에디터 기반

#### Day 36-38: 에디터 API 설계
- [ ] 타워 빌더 REST API
- [ ] 컴포넌트 팔레트 API
- [ ] 실시간 미리보기 API
- [ ] 조립 검증 API

**파일 생성 목록:**
```
services/timesquare/tower/api/
├── builder.go           # 빌더 API
├── palette.go           # 팔레트 API
├── preview.go           # 미리보기 API
└── validation.go        # 검증 API

api/tower/
├── builder.proto        # gRPC 정의 (선택사항)
└── openapi.yaml         # REST API 문서
```

#### Day 39-42: 고급 기능 구현
- [ ] 스마트 자동 연결 시스템
- [ ] 런타임 조합 변경
- [ ] 타워 저장/불러오기
- [ ] 커뮤니티 공유 시스템

**파일 생성 목록:**
```
pkg/tower/editor/
├── smart_connect.go     # 스마트 연결
├── runtime.go           # 런타임 수정
├── storage.go           # 저장/불러오기
└── sharing.go           # 공유 시스템
```

## 🧪 Phase 4: 테스트 및 문서화 (1-2주)

### Week 7: 종합 테스트

#### Day 43-45: 통합 테스트
- [ ] 전체 시스템 통합 테스트
- [ ] 성능 벤치마크 테스트
- [ ] 부하 테스트 (동시 접속자)
- [ ] 메모리 누수 테스트

#### Day 46-49: 문서화 및 예제
- [ ] API 문서 작성
- [ ] 사용자 가이드 작성
- [ ] 개발자 문서 작성
- [ ] 예제 타워 조합 작성

**파일 생성 목록:**
```
docs/tower-system/
├── api-reference.md     # API 레퍼런스
├── user-guide.md        # 사용자 가이드
├── developer-guide.md   # 개발자 가이드
└── examples/            # 예제들
    ├── basic-towers.md
    ├── advanced-combos.md
    └── custom-components.md
```

## 📊 구현 우선순위 매트릭스

| 기능 | 중요도 | 복잡도 | 우선순위 | 예상 시간 |
|------|--------|--------|----------|-----------|
| 기본 컴포넌트 시스템 | 높음 | 중간 | 1 | 1주 |
| JSON 데이터 시스템 | 높음 | 낮음 | 2 | 3일 |
| 매트릭스 엔진 | 높음 | 높음 | 3 | 1주 |
| 서버 통합 | 높음 | 중간 | 4 | 4일 |
| 성능 최적화 | 중간 | 높음 | 5 | 4일 |
| 에디터 API | 중간 | 중간 | 6 | 1주 |
| 고급 기능 | 낮음 | 높음 | 7 | 1주 |

## 🎯 마일스톤

### Milestone 1 (Week 2 완료)
- ✅ 기본 컴포넌트 시스템 동작
- ✅ JSON으로 타워 정의 가능
- ✅ 간단한 타워 조합 실행

### Milestone 2 (Week 4 완료)
- ✅ 매트릭스 기반 밸런싱 동작
- ✅ 18개 종족 기본 타워 완성
- ✅ 환경 효과 시스템 동작

### Milestone 3 (Week 6 완료)
- ✅ 서버 통합 완료
- ✅ 실시간 전투 시스템 동작
- ✅ 기본 에디터 API 완성

### Milestone 4 (Week 7 완료)
- ✅ 전체 시스템 안정화
- ✅ 성능 최적화 완료
- ✅ 문서화 완료

## 🚀 시작 준비

### 즉시 시작 가능한 작업
1. **컴포넌트 인터페이스 설계** (Day 1)
2. **기본 디렉토리 구조 생성** (Day 1)
3. **JSON 스키마 초안 작성** (Day 1)

### 다음 단계
첫 번째 작업부터 시작하시겠습니까? 어떤 부분부터 구현해보고 싶으신지 알려주세요!
