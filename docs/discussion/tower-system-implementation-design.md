# Defense Allies 타워 시스템 구현 설계 문서

## 📋 문서 정보
- **작성일**: 2024년
- **버전**: v1.0
- **목적**: Defense Allies 타워 시스템 최적 구현 방안
- **범위**: 레고 스타일 모듈형 타워 시스템

## 🎯 설계 철학 및 핵심 원칙

### 설계 목표
1. **레고 스타일 모듈성**: 타워 능력을 블록처럼 자유롭게 조합
2. **데이터 주도 개발**: 모든 능력을 JSON/YAML로 정의 가능
3. **빠른 프로토타이핑**: 5분 내 새로운 타워 생성 가능
4. **서버 권위적 처리**: 모든 전투 계산을 서버에서 수행
5. **매트릭스 기반 밸런싱**: N×N 매트릭스로 복잡한 밸런스 관리

### 핵심 아키텍처 패턴
- **Data-Driven Component System**: 기본 구조
- **Atomic Component Pattern**: 레고 블록 구현
- **Matrix-Driven Calculation**: 밸런싱 시스템
- **Event-Driven Architecture**: 상호작용 처리
- **Rule-Based Behavior**: 조건부 로직

## 🏗️ 시스템 아키텍처

### 계층 구조
```
┌─────────────────────────────────────┐
│ UI Layer (Visual Tower Builder)     │ ← 레고 스타일 에디터
├─────────────────────────────────────┤
│ Data Layer (JSON/YAML Definitions)  │ ← 타워 정의 데이터
├─────────────────────────────────────┤
│ Logic Layer (Component System)      │ ← 컴포넌트 조합 엔진
├─────────────────────────────────────┤
│ Calculation Layer (Matrix Engine)   │ ← 매트릭스 기반 계산
├─────────────────────────────────────┤
│ Execution Layer (Server Combat)     │ ← 서버 전투 처리
└─────────────────────────────────────┘
```

### 핵심 컴포넌트

#### A. Atomic Component System
```go
type AtomicComponent interface {
    GetType() ComponentType
    Execute(context *ExecutionContext) ComponentResult
    GetInputs() []ComponentInput
    GetOutputs() []ComponentOutput
    CanConnectTo(other AtomicComponent) bool
}

// 기본 컴포넌트 타입들
- TargetingComponent    // 타겟 선택
- DamageComponent      // 피해 계산
- EffectComponent      // 상태 효과
- RangeComponent       // 사거리 처리
- CooldownComponent    // 쿨다운 관리
- ConditionalComponent // 조건부 로직
```

#### B. Component Assembly Engine
```go
type ComponentAssembly struct {
    Components  []AtomicComponent
    Connections []ComponentConnection
    Metadata    AssemblyMetadata
}

type AssemblyEngine struct {
    ComponentRegistry *ComponentRegistry
    ConnectionValidator *ConnectionValidator
    ExecutionPlanner *ExecutionPlanner
}
```

#### C. Matrix Calculation Engine
```go
type MatrixEngine struct {
    BaseMatrices    map[string]Matrix
    ModifierEngine  *ModifierEngine
    BalanceEngine   *BalanceEngine
}
```

## 🧱 레고 스타일 컴포넌트 설계

### 컴포넌트 카테고리

#### 1. 타겟팅 블록
```yaml
targeting_blocks:
  single_target:
    inputs: []
    outputs: [target]
    config: [range, priority]
  
  multi_target:
    inputs: []
    outputs: [targets]
    config: [range, max_targets, priority]
  
  area_target:
    inputs: []
    outputs: [area_targets]
    config: [radius, center_type]
```

#### 2. 공격 블록
```yaml
attack_blocks:
  direct_damage:
    inputs: [targets]
    outputs: [damage_events]
    config: [base_damage, damage_type]
  
  projectile_attack:
    inputs: [targets]
    outputs: [projectiles]
    config: [projectile_speed, homing]
  
  chain_attack:
    inputs: [initial_target]
    outputs: [chain_events]
    config: [chain_count, damage_decay, chain_range]
```

#### 3. 효과 블록
```yaml
effect_blocks:
  status_effect:
    inputs: [targets, trigger_data]
    outputs: [status_applications]
    config: [effect_type, duration, intensity]
  
  modifier_effect:
    inputs: [targets, base_values]
    outputs: [modified_values]
    config: [modifier_type, multiplier, duration]
```

### 블록 연결 시스템
```json
{
  "tower_assembly": {
    "blocks": [
      {
        "id": "targeting_1",
        "type": "multi_target",
        "config": {"range": 8, "max_targets": 3}
      },
      {
        "id": "damage_1", 
        "type": "fire_damage",
        "config": {"base_damage": 100, "element": "fire"}
      },
      {
        "id": "effect_1",
        "type": "burn_effect", 
        "config": {"duration": 3, "dps": 20}
      }
    ],
    "connections": [
      {"from": "targeting_1.targets", "to": "damage_1.targets"},
      {"from": "damage_1.affected_targets", "to": "effect_1.targets"}
    ]
  }
}
```

## 📊 매트릭스 기반 밸런싱

### 기본 매트릭스 구조
```yaml
# 2x2 기본 매트릭스 (확장 가능)
power_matrix: [[offensive, defensive], [individual, cooperation]]

# 종족별 기본 매트릭스
dragon_base: [[1.5, 0.8], [1.2, 0.9]]  # 공격 특화
elven_base:  [[1.2, 1.0], [0.9, 1.3]]  # 협력 특화
dwarf_base:  [[1.0, 1.4], [1.1, 1.0]]  # 방어 특화
```

### 매트릭스 연산 시스템
```go
type MatrixOperation interface {
    Apply(base Matrix, modifier Matrix) Matrix
}

// 연산 타입들
- HadamardProduct  // 환경 효과 (원소별 곱셈)
- MatrixMultiply   // 시너지 효과 (행렬 곱셈)
- MatrixAdd        // 버프 효과 (행렬 덧셈)
- ScalarMultiply   // 스케일링 (스칼라 곱셈)
```

## 🚀 빠른 프로토타이핑 시스템

### 프리셋 시스템
```json
{
  "presets": {
    "basic_archer": {
      "description": "기본 원거리 공격",
      "assembly": {
        "blocks": ["single_targeting", "projectile_attack", "accuracy_check"],
        "auto_connect": true
      }
    },
    "flame_tower": {
      "description": "화염 공격 타워",
      "assembly": {
        "blocks": ["area_targeting", "fire_damage", "burn_effect"],
        "auto_connect": true
      }
    }
  }
}
```

### 템플릿 시스템
```yaml
templates:
  elemental_damage:
    variables: [element_type, base_damage, special_effect]
    structure:
      - type: "targeting_block"
      - type: "{{element_type}}_damage_block"
        config: {damage: "{{base_damage}}"}
      - type: "{{special_effect}}_block"
```

## 🎮 서버 기반 전투 시스템

### 전투 처리 파이프라인
```
1. 타겟팅 → 2. 공격 계산 → 3. 효과 적용 → 4. 상태 동기화
```

### 배치 처리 최적화
```go
type CombatBatchProcessor struct {
    TowerBatch    []TowerAttack
    EffectBatch   []EffectApplication
    SynergyBatch  []SynergyCalculation
}
```

## 🔧 구현 기술 스택

### 백엔드 (Go)
- **컴포넌트 시스템**: 인터페이스 기반 모듈 설계
- **매트릭스 연산**: gonum/mat 라이브러리
- **JSON 처리**: encoding/json 표준 라이브러리
- **동시성**: goroutine + channel 패턴

### 데이터 저장 (Redis)
- **타워 정의**: JSON 문서로 저장
- **게임 상태**: 해시맵으로 실시간 상태 관리
- **프리셋/템플릿**: 키-값 저장

### 클라이언트 동기화
- **SSE**: 실시간 상태 업데이트
- **델타 압축**: 변경사항만 전송
- **예측 렌더링**: 클라이언트 측 보간

## 📈 성능 최적화 전략

### 메모리 최적화
- **컴포넌트 풀링**: sync.Pool 활용
- **객체 재사용**: 가비지 컬렉션 최소화
- **캐시 친화적 데이터**: 구조체 정렬

### 계산 최적화
- **공간 분할**: 그리드 기반 근접 검색
- **배치 처리**: SIMD 활용 가능한 구조
- **지연 계산**: 필요시에만 매트릭스 연산

### 네트워크 최적화
- **적응적 업데이트**: 네트워크 상태별 빈도 조절
- **압축**: 델타 + 딕셔너리 압축
- **배치 전송**: 여러 업데이트 묶어서 전송

## 🎯 확장성 고려사항

### 수평 확장
- **상태 분할**: 게임 룸별 서버 분산
- **로드 밸런싱**: 타워 수에 따른 부하 분산
- **캐시 계층**: Redis Cluster 활용

### 기능 확장
- **새 컴포넌트**: 플러그인 방식으로 추가
- **새 연산**: 매트릭스 연산 인터페이스 확장
- **새 이벤트**: 이벤트 버스 확장

---

**다음 단계**: 구현 TODO 리스트 작성 및 단계별 개발 진행
