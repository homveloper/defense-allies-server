# Defense Allies N차원 매트릭스 밸런싱 시스템

## 📋 문서 정보
- **작성일**: 2024년
- **버전**: v1.0
- **목적**: Defense Allies 게임의 N차원 매트릭스 기반 밸런싱 시스템 설계
- **범위**: 종족, 타워, 환경 변수의 다차원 수치 모델링

## 🎯 시스템 개요

### 기존 문제점
기존의 단일 수치(100점) 밸런싱은 복잡한 게임 메커니즘을 정확히 표현할 수 없습니다:
- 공격력 150, 방어력 50인 타워와 공격력 75, 방어력 125인 타워가 동일하게 평가됨
- 상황별 강약점을 구분할 수 없음
- 협력 효과, 환경 상호작용을 정확히 모델링 불가

### 해결책: N×N 매트릭스 시스템
**다차원 매트릭스**를 도입하여 게임의 모든 복잡성을 수학적으로 정확히 표현합니다.

## 🔢 기본 매트릭스 구조

### 2×2 기본 매트릭스
```yaml
tower_power_matrix:
  base_matrix: [[1.0, 1.0], [1.0, 1.0]]  # 모든 타워의 기준 파워
  dimensions:
    row: [offensive_power, defensive_power]    # 행: 능력 유형
    col: [individual_mode, cooperation_mode]   # 열: 플레이 모드
```

### 실제 예시
```yaml
human_basic_tower:
  power_matrix: [[1.0, 0.8], [0.9, 1.1]]
  interpretation:
    individual_offensive: 1.0  # 개별 공격력
    individual_defensive: 0.8  # 개별 방어력  
    cooperation_offensive: 0.9 # 협력 공격력
    cooperation_defensive: 1.1 # 협력 방어력

dragon_basic_tower:
  power_matrix: [[1.5, 0.6], [1.2, 0.7]]
  interpretation:
    # 높은 공격력, 낮은 방어력
    # 협력 시에도 공격 우위, 방어 취약 유지
```

### 4×4 확장 매트릭스
```yaml
advanced_tower_matrix:
  dimensions: [offensive, defensive, utility, synergy]
  
elven_archer_tower:
  power_matrix:
    - [1.3, 0.7, 0.9, 1.0]  # offensive 기준
    - [0.6, 1.1, 0.8, 1.2]  # defensive 기준  
    - [0.8, 0.9, 1.4, 1.1]  # utility 기준
    - [1.1, 1.0, 1.2, 1.3]  # synergy 기준
  
  key_values:
    pure_offense: 1.3      # [0,0] 순수 공격력
    offense_defense_trade: 0.7  # [0,1] 공격 시 방어 취약
    utility_synergy: 1.1   # [2,3] 유틸리티가 시너지에 기여
```

## ⚖️ 매트릭스 연산 규칙

### 1. 기본 균형 법칙
```python
# 모든 타워의 매트릭스 프로베니우스 노름이 동일해야 함
frobenius_norm(tower_matrix) = constant_value

# 예: 2×2 매트릭스의 경우
human_tower: [[1.0, 0.8], [0.9, 1.1]] 
→ ||A||_F = √(1.0² + 0.8² + 0.9² + 1.1²) = 1.85

dragon_tower: [[1.5, 0.6], [1.2, 0.7]]
→ ||A||_F = √(1.5² + 0.6² + 1.2² + 0.7²) = 2.02
# 불균형! 조정 필요
```

### 2. 협력 효과 (매트릭스 곱셈)
```python
# 두 종족이 협력할 때
cooperation_result = race1_matrix × race2_matrix

# 예: 휴먼 + 엘프 협력
human_matrix = [[1.0, 0.8], [0.9, 1.1]]
elf_matrix = [[1.3, 0.7], [0.6, 1.2]]

result = [[1.0×1.3 + 0.8×0.6, 1.0×0.7 + 0.8×1.2],
          [0.9×1.3 + 1.1×0.6, 0.9×0.7 + 1.1×1.2]]
       = [[1.78, 1.66], [1.83, 1.95]]
```

### 3. 환경 적용 (아다마르 곱)
```python
# 환경 효과는 요소별 곱셈
final_matrix = base_matrix ⊙ environment_matrix

# 예: 숲에서의 엘프
elf_base = [[1.3, 0.7], [0.6, 1.2]]
forest_modifier = [[1.4, 1.2], [1.3, 1.5]]
forest_elf = [[1.3×1.4, 0.7×1.2], [0.6×1.3, 1.2×1.5]]
           = [[1.82, 0.84], [0.78, 1.8]]
```

## 🌍 환경 변수 매트릭스

### 시간대 효과 매트릭스
```yaml
time_modifiers:
  day:
    angel_legion: [[1.2, 1.0], [1.1, 1.3]]
    undead_legion: [[0.8, 1.0], [0.9, 0.7]]
    
  night:
    angel_legion: [[0.9, 1.0], [0.8, 1.0]]
    undead_legion: [[1.4, 1.2], [1.3, 1.5]]
```

### 날씨 효과 매트릭스
```yaml
weather_modifiers:
  clear: [[1.0, 1.0], [1.0, 1.0]]  # 기준값
  rain: [[0.9, 1.1], [0.8, 1.2]]   # 방어 유리
  storm: [[0.7, 1.3], [0.6, 1.4]]  # 극단적 방어 유리
  snow: [[0.8, 1.2], [0.9, 1.1]]   # 약간 방어 유리
```

### 지형 효과 매트릭스
```yaml
terrain_modifiers:
  plain: [[1.0, 1.0], [1.0, 1.0]]     # 기준값
  forest: [[1.4, 1.2], [1.3, 1.5]]    # 엘프 유리
  mountain: [[0.8, 1.4], [1.2, 0.9]]  # 드워프 유리  
  desert: [[0.6, 0.8], [0.7, 0.5]]    # 기계 불리
```

## 🔄 종족 상호작용 매트릭스

### 8×8 핵심 종족 상호작용
```yaml
race_interaction_matrix:
  # 행: 자신, 열: 협력 상대
  #        Hum  Elf  Dwa  Orc  Und  Dra  Mec  Ang
  Human:   [1.0, 1.2, 1.1, 0.9, 0.8, 1.0, 1.1, 1.3]
  Elven:   [1.2, 1.0, 0.9, 0.7, 0.6, 1.1, 0.8, 1.4]
  Dwarven: [1.1, 0.9, 1.0, 1.3, 0.7, 1.2, 1.4, 0.9]
  Orc:     [0.9, 0.7, 1.3, 1.0, 1.1, 1.0, 0.6, 0.5]
  Undead:  [0.8, 0.6, 0.7, 1.1, 1.0, 1.2, 0.9, 0.3]
  Dragon:  [1.0, 1.1, 1.2, 1.0, 1.2, 1.0, 0.8, 0.7]
  Mech:    [1.1, 0.8, 1.4, 0.6, 0.9, 0.8, 1.0, 1.0]
  Angel:   [1.3, 1.4, 0.9, 0.5, 0.3, 0.7, 1.0, 1.0]
```

### 해석
- `Human-Elven: 1.2` → 휴먼이 엘프와 협력 시 20% 보너스
- `Undead-Angel: 0.3` → 언데드가 천사와 협력 시 70% 페널티
- 대각선은 항상 1.0 (자기 자신과의 상호작용)

## 📊 밸런스 메트릭

### 수학적 측정 지표
```python
balance_metrics = {
    'frobenius_norm': norm(matrix, 'fro'),     # 총 파워 측정
    'determinant': det(matrix),                # 파워 집중도
    'trace': trace(matrix),                    # 대각합 (핵심 능력)
    'eigenvalues': eig(matrix)[0],             # 주요 특성값
    'condition_number': cond(matrix),          # 수치적 안정성
    'spectral_radius': max(abs(eig(matrix)[0])) # 최대 고유값
}
```

### 균형 조건
```yaml
balance_constraints:
  frobenius_norm: 
    target: 2.0
    tolerance: ±0.1
    
  determinant:
    range: [0.5, 1.5]
    
  eigenvalues:
    real_part_range: [0.3, 1.7]
    complex_part_max: 0.2
    
  condition_number:
    max_value: 10.0  # 수치적 안정성 보장
```

## 🎯 동적 밸런싱 알고리즘

### 실시간 모니터링
```python
def monitor_game_balance(game_state):
    player_matrices = [calculate_effective_matrix(player) 
                      for player in game_state.players]
    
    balance_score = calculate_balance_score(player_matrices)
    
    if balance_score < BALANCE_THRESHOLD:
        trigger_balancing_event(game_state)

def calculate_balance_score(matrices):
    # 모든 플레이어 매트릭스의 프로베니우스 노름 분산
    norms = [norm(matrix, 'fro') for matrix in matrices]
    return 1.0 / (1.0 + np.var(norms))
```

### 적응형 환경 생성
```python
def generate_balancing_environment(player_races):
    current_power = calculate_team_power_matrix(player_races)
    target_power = get_ideal_balance_matrix()
    
    # 현재 상태를 목표 상태로 이끄는 환경 찾기
    optimal_env = optimize_environment(current_power, target_power)
    return optimal_env

def optimize_environment(current, target):
    def objective(env_params):
        env_matrix = create_environment_matrix(env_params)
        result = current ⊙ env_matrix
        return frobenius_norm(result - target)
    
    return minimize(objective, initial_guess)
```

## 🔮 확장 변수 매트릭스

### 우주 이벤트 매트릭스
```yaml
cosmic_events:
  meteor_shower:
    duration: 180  # 초
    effect_matrices:
      dragon_clan: [[1.4, 1.0], [1.2, 1.1]]
      cosmic_empire: [[1.6, 1.2], [1.4, 1.3]]
      global_damage: [[0.9, 0.9], [0.9, 0.9]]
      
  solar_eclipse:
    duration: 300
    effect_matrices:
      undead_legion: [[1.5, 1.3], [1.4, 1.6]]
      shadow_clan: [[1.8, 1.5], [1.7, 1.9]]
      angel_legion: [[0.7, 0.8], [0.6, 0.9]]
```

### 복합 효과 계산
```python
def apply_combo_effect(base_matrix, events):
    result = base_matrix.copy()
    
    for event in events:
        event_matrix = get_event_matrix(event)
        result = result ⊙ event_matrix
    
    # 복합 효과 보너스/페널티
    if len(events) >= 2:
        combo_bonus = calculate_combo_matrix(events)
        result = result ⊙ combo_bonus
    
    return result
```

## 📈 시각화 및 분석

### 히트맵 표현
```yaml
visualization:
  heatmap:
    x_axis: [offensive, defensive, utility, synergy]
    y_axis: [individual, team, environment, special]
    color_scale: [0.0, 2.0]
    colormap: "RdYlBu_r"
    
  radar_chart:
    axes: [attack, defense, support, control, mobility]
    normalization: "frobenius_norm"
    overlay_races: true
```

### 주성분 분석 (PCA)
```python
def analyze_race_diversity(race_matrices):
    # 매트릭스를 벡터로 변환
    vectors = [matrix.flatten() for matrix in race_matrices]
    
    # PCA 적용
    pca = PCA(n_components=2)
    reduced = pca.fit_transform(vectors)
    
    # 다양성 점수 계산
    diversity_score = calculate_spread(reduced)
    return diversity_score, reduced
```

---

**다음 단계**: 18개 종족 매트릭스 수치 최적화 및 실제 구현
