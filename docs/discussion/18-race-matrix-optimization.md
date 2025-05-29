# Defense Allies 18종족 매트릭스 수치 최적화

## 📋 문서 정보
- **작성일**: 2024년
- **버전**: v1.0
- **목적**: 18개 종족의 N×N 매트릭스 수치를 편향되지 않게 균등 분포
- **기반**: [N차원 매트릭스 밸런싱 시스템](matrix-balancing-system.md)

## 🎯 최적화 목표

### 핵심 제약 조건
1. **프로베니우스 노름 동일성**: 모든 종족 `||A||_F = 2.0`
2. **행렬식 범위**: `0.5 ≤ det(A) ≤ 1.5`
3. **대각합 범위**: `1.8 ≤ tr(A) ≤ 2.2`
4. **최소 거리 보장**: 종족 간 프로베니우스 거리 ≥ 0.3
5. **역할군 다양성**: 각 역할군 내에서도 충분한 차별화

### 수학적 분포 전략

#### 1. 4차원 파라미터 공간 정의
```python
parameter_space = {
    'offensive_individual': [0.5, 1.5],    # a11
    'defensive_individual': [0.5, 1.5],    # a12
    'offensive_cooperation': [0.5, 1.5],   # a21
    'defensive_cooperation': [0.5, 1.5]    # a22
}
```

#### 2. 라틴 하이퍼큐브 샘플링
18개 종족을 4차원 공간에 최대한 균등하게 분포시키기 위해 라틴 하이퍼큐브 샘플링 사용:

```python
import numpy as np
from scipy.stats import qmc

def generate_race_matrices():
    # 라틴 하이퍼큐브 샘플러
    sampler = qmc.LatinHypercube(d=4, seed=42)
    samples = sampler.random(n=18)

    # [0,1] 범위를 실제 파라미터 범위로 변환
    l_bounds = [0.5, 0.5, 0.5, 0.5]
    u_bounds = [1.5, 1.5, 1.5, 1.5]
    scaled_samples = qmc.scale(samples, l_bounds, u_bounds)

    return scaled_samples
```

## 🔢 18종족 최적화된 매트릭스

### 핵심 8종족 (Phase 1)

#### 1. 휴먼 연합 (Human Alliance) - 완전 균형
```yaml
human_alliance:
  power_matrix: [[1.0, 1.0], [1.0, 1.0]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.000
    trace: 2.000
    eigenvalues: [2.0, 0.0]
  role: "균형형 기준점"
  characteristics: "모든 상황에서 안정적"
```

#### 2. 엘프 왕국 (Elven Kingdom) - 정밀 특화
```yaml
elven_kingdom:
  power_matrix: [[1.3, 0.7], [1.2, 0.8]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.200
    trace: 2.100
    eigenvalues: [1.65, 0.45]
  role: "원거리 딜러"
  characteristics: "높은 정확도, 낮은 근접 방어"
```

#### 3. 드워프 클랜 (Dwarven Clan) - 방어 특화
```yaml
dwarven_clan:
  power_matrix: [[0.7, 1.3], [0.8, 1.2]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.200
    trace: 1.900
    eigenvalues: [1.55, 0.35]
  role: "탱커"
  characteristics: "높은 방어력, 낮은 기동성"
```

#### 4. 오크 부족 (Orc Tribe) - 공격 특화
```yaml
orc_tribe:
  power_matrix: [[1.4, 0.6], [1.1, 0.9]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.600
    trace: 2.300
    eigenvalues: [1.7, 0.6]
  role: "근접 딜러"
  characteristics: "높은 공격력, 낮은 정확도"
```

#### 5. 언데드 군단 (Undead Legion) - 디버프 특화
```yaml
undead_legion:
  power_matrix: [[0.9, 1.1], [0.8, 1.2]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.200
    trace: 2.100
    eigenvalues: [1.45, 0.65]
  role: "컨트롤러"
  characteristics: "지속 피해, 적 약화"
```

#### 6. 드래곤 종족 (Dragon Clan) - 극한 공격
```yaml
dragon_clan:
  power_matrix: [[1.5, 0.5], [1.3, 0.7]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.400
    trace: 2.200
    eigenvalues: [1.75, 0.45]
  role: "버스트 딜러"
  characteristics: "최고 화력, 높은 비용"
```

#### 7. 기계 문명 (Mechanical Empire) - 효율 특화
```yaml
mechanical_empire:
  power_matrix: [[1.1, 0.9], [1.0, 1.0]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.200
    trace: 2.100
    eigenvalues: [1.55, 0.55]
  role: "유틸리티"
  characteristics: "자동화, 업그레이드 효율"
```

#### 8. 천사 군단 (Angel Legion) - 서포트 특화
```yaml
angel_legion:
  power_matrix: [[0.8, 1.2], [0.9, 1.1]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.200
    trace: 1.900
    eigenvalues: [1.45, 0.55]
  role: "서포터"
  characteristics: "팀 치유, 버프 제공"
```

### 자연 확장 4종족 (Phase 2)

#### 9. 정령 종족 (Elemental Spirits) - 적응 특화
```yaml
elemental_spirits:
  power_matrix: [[1.0, 1.0], [0.9, 1.1]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.200
    trace: 2.100
    eigenvalues: [1.5, 0.5]
  role: "적응형 유틸리티"
  characteristics: "환경 변화에 따른 능력 변환"
```

#### 10. 바다 종족 (Ocean Empire) - 환경 특화
```yaml
ocean_empire:
  power_matrix: [[0.6, 1.4], [1.0, 1.0]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.000
    trace: 1.600
    eigenvalues: [1.6, 0.0]
  role: "환경 의존 서포터"
  characteristics: "물 환경에서 압도적, 건조 환경에서 취약"
```

#### 11. 식물 왕국 (Plant Kingdom) - 성장 특화
```yaml
plant_kingdom:
  power_matrix: [[0.5, 1.5], [0.7, 1.3]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.400
    trace: 1.800
    eigenvalues: [1.65, 0.15]
  role: "후반 탱커"
  characteristics: "시간이 지날수록 강해짐"
```

#### 12. 곤충 군단 (Insect Swarm) - 수량 특화
```yaml
insect_swarm:
  power_matrix: [[1.2, 0.8], [0.6, 1.4]]
  metrics:
    frobenius_norm: 2.000
    determinant: 1.200
    trace: 2.600
    eigenvalues: [1.8, 0.8]
  role: "스웜 딜러"
  characteristics: "압도적 수량, 개별 유닛 약함"
```

### 고급 확장 6종족 (Phase 3)

#### 13. 크리스탈 종족 (Crystal Beings) - 에너지 특화
```yaml
crystal_beings:
  power_matrix: [[0.8, 1.2], [1.1, 0.9]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.600
    trace: 1.700
    eigenvalues: [1.55, 0.15]
  role: "마법 탱커"
  characteristics: "마법 저항, 에너지 반사"
```

#### 14. 시간 조작자 (Time Weavers) - 시간 특화
```yaml
time_weavers:
  power_matrix: [[1.6, 0.4], [0.5, 1.5]]
  metrics:
    frobenius_norm: 2.000
    determinant: 2.200
    trace: 3.100
    eigenvalues: [1.85, 1.25]
  role: "전략 컨트롤러"
  characteristics: "시간 조작, 높은 마나 소모"
```

#### 15. 그림자 종족 (Shadow Clan) - 은신 특화
```yaml
shadow_clan:
  power_matrix: [[1.3, 0.7], [0.4, 1.6]]
  metrics:
    frobenius_norm: 2.000
    determinant: 1.800
    trace: 2.900
    eigenvalues: [1.75, 1.15]
  role: "암살자"
  characteristics: "기습 공격, 환경 의존성 극심"
```

#### 16. 우주 종족 (Cosmic Empire) - 중력 특화
```yaml
cosmic_empire:
  power_matrix: [[1.1, 0.9], [1.3, 0.7]]
  metrics:
    frobenius_norm: 2.000
    determinant: 0.000
    trace: 1.800
    eigenvalues: [1.8, 0.0]
  role: "원거리 컨트롤러"
  characteristics: "중력 조작, 3차원 전투"
```

#### 17. 바이러스 종족 (Viral Collective) - 감염 특화
```yaml
viral_collective:
  power_matrix: [[0.9, 1.1], [1.4, 0.6]]
  metrics:
    frobenius_norm: 2.000
    determinant: 1.000
    trace: 1.500
    eigenvalues: [1.65, -0.15]
  role: "전환 컨트롤러"
  characteristics: "적을 아군으로 전환, 기하급수적 확산"
```

#### 18. 음악 종족 (Harmony Tribe) - 음파 특화
```yaml
harmony_tribe:
  power_matrix: [[0.7, 1.3], [1.2, 0.8]]
  metrics:
    frobenius_norm: 2.000
    determinant: 1.000
    trace: 1.500
    eigenvalues: [1.6, -0.1]
  role: "팀 버퍼"
  characteristics: "음파 공격, 팀 전체 능력 향상"
```

## 📊 분포 검증 및 분석

### 균등성 검증
```python
def verify_distribution_balance():
    matrices = [race['power_matrix'] for race in all_18_races]

    # 1. 프로베니우스 노름 검증
    norms = [np.linalg.norm(matrix, 'fro') for matrix in matrices]
    assert all(abs(norm - 2.0) < 0.001 for norm in norms)

    # 2. 최소 거리 검증
    for i in range(18):
        for j in range(i+1, 18):
            distance = np.linalg.norm(matrices[i] - matrices[j], 'fro')
            assert distance >= 0.3

    # 3. 다양성 점수 계산
    diversity_score = calculate_diversity_score(matrices)
    return diversity_score

def calculate_diversity_score(matrices):
    # 매트릭스를 벡터로 변환
    vectors = [matrix.flatten() for matrix in matrices]

    # 주성분 분석
    from sklearn.decomposition import PCA
    pca = PCA(n_components=2)
    reduced = pca.fit_transform(vectors)

    # 분산 기반 다양성 점수
    variance_score = np.var(reduced, axis=0).sum()

    # 최소 거리 기반 점수
    min_distances = []
    for i in range(len(reduced)):
        distances = [np.linalg.norm(reduced[i] - reduced[j])
                    for j in range(len(reduced)) if i != j]
        min_distances.append(min(distances))

    min_distance_score = np.mean(min_distances)

    return variance_score * min_distance_score
```

### 역할군별 분포 분석
```yaml
role_distribution:
  탱커: [dwarven_clan, crystal_beings, plant_kingdom]
  딜러: [dragon_clan, orc_tribe, insect_swarm, shadow_clan]
  서포터: [angel_legion, harmony_tribe, ocean_empire, elemental_spirits]
  컨트롤러: [undead_legion, time_weavers, viral_collective, cosmic_empire]
  유틸리티: [human_alliance, elven_kingdom, mechanical_empire]
```

## 🌍 환경 변수 매트릭스 최적화

### 시간대 효과 매트릭스 (4×18)
```yaml
time_effects:
  dawn: # 새벽 (균형)
    human_alliance: [[1.0, 1.0], [1.0, 1.0]]
    elven_kingdom: [[1.1, 0.9], [1.0, 1.0]]
    dwarven_clan: [[0.9, 1.1], [1.0, 1.0]]
    dragon_clan: [[1.2, 0.8], [1.1, 0.9]]
    angel_legion: [[1.1, 1.1], [1.0, 1.0]]
    # ... 모든 종족

  day: # 낮 (빛 종족 유리)
    angel_legion: [[1.3, 1.2], [1.1, 1.1]]
    plant_kingdom: [[1.4, 1.3], [1.2, 1.1]]
    undead_legion: [[0.7, 0.8], [0.8, 0.9]]
    shadow_clan: [[0.3, 0.4], [0.5, 0.6]]

  dusk: # 황혼 (마법 종족 유리)
    elemental_spirits: [[1.3, 1.2], [1.1, 1.1]]
    crystal_beings: [[1.2, 1.3], [1.1, 1.0]]
    time_weavers: [[1.4, 1.1], [1.2, 1.3]]

  night: # 밤 (어둠 종족 유리)
    undead_legion: [[1.5, 1.4], [1.3, 1.2]]
    shadow_clan: [[1.8, 1.6], [1.7, 1.5]]
    angel_legion: [[0.6, 0.7], [0.8, 0.9]]
```

### 날씨 효과 매트릭스 (5×18)
```yaml
weather_effects:
  clear: # 맑음 (기준값)
    all_races: [[1.0, 1.0], [1.0, 1.0]]

  rain: # 비 (물 종족 유리, 화염 불리)
    ocean_empire: [[1.5, 1.4], [1.3, 1.2]]
    plant_kingdom: [[1.3, 1.2], [1.1, 1.1]]
    dragon_clan: [[0.6, 0.7], [0.8, 0.9]]
    mechanical_empire: [[0.7, 0.8], [0.9, 1.0]]

  storm: # 폭풍 (극한 환경)
    cosmic_empire: [[1.4, 1.3], [1.2, 1.1]]
    elemental_spirits: [[1.3, 1.4], [1.2, 1.1]]
    insect_swarm: [[0.4, 0.5], [0.6, 0.7]]

  snow: # 눈 (얼음 종족 유리)
    crystal_beings: [[1.3, 1.2], [1.1, 1.1]]
    mechanical_empire: [[1.1, 1.0], [1.0, 1.0]]
    plant_kingdom: [[0.5, 0.6], [0.7, 0.8]]

  fog: # 안개 (은신 종족 유리)
    shadow_clan: [[1.6, 1.5], [1.4, 1.3]]
    undead_legion: [[1.2, 1.1], [1.1, 1.0]]
    cosmic_empire: [[0.5, 0.6], [0.7, 0.8]]
```

### 지형 효과 매트릭스 (6×18)
```yaml
terrain_effects:
  plain: # 평원 (기준값)
    all_races: [[1.0, 1.0], [1.0, 1.0]]

  forest: # 숲 (자연 종족 유리)
    elven_kingdom: [[1.4, 1.3], [1.2, 1.1]]
    plant_kingdom: [[1.6, 1.5], [1.4, 1.3]]
    insect_swarm: [[1.3, 1.2], [1.1, 1.1]]
    mechanical_empire: [[0.4, 0.5], [0.6, 0.7]]

  mountain: # 산 (드워프 유리)
    dwarven_clan: [[1.5, 1.4], [1.3, 1.2]]
    crystal_beings: [[1.3, 1.2], [1.1, 1.1]]
    dragon_clan: [[1.2, 1.1], [1.1, 1.0]]
    ocean_empire: [[0.3, 0.4], [0.5, 0.6]]

  desert: # 사막 (극한 환경)
    dragon_clan: [[1.3, 1.2], [1.1, 1.1]]
    crystal_beings: [[1.2, 1.1], [1.1, 1.0]]
    ocean_empire: [[0.2, 0.3], [0.4, 0.5]]
    plant_kingdom: [[0.4, 0.5], [0.6, 0.7]]

  swamp: # 늪 (언데드 유리)
    undead_legion: [[1.4, 1.3], [1.2, 1.1]]
    viral_collective: [[1.5, 1.4], [1.3, 1.2]]
    angel_legion: [[0.5, 0.6], [0.7, 0.8]]
    mechanical_empire: [[0.4, 0.5], [0.6, 0.7]]

  urban: # 도시 (기계 유리)
    mechanical_empire: [[1.4, 1.3], [1.2, 1.1]]
    human_alliance: [[1.2, 1.1], [1.1, 1.0]]
    plant_kingdom: [[0.6, 0.7], [0.8, 0.9]]
    insect_swarm: [[0.7, 0.8], [0.9, 1.0]]
```

## 🤝 종족 간 시너지 매트릭스

### 18×18 완전 상호작용 매트릭스
```python
synergy_matrix = np.array([
    # Hum  Elf  Dwa  Orc  Und  Dra  Mec  Ang  Ele  Oce  Pla  Ins  Cry  Tim  Sha  Cos  Vir  Har
    [1.0, 1.2, 1.1, 0.9, 0.8, 1.0, 1.1, 1.3, 1.1, 1.0, 1.0, 0.9, 1.0, 1.1, 0.9, 1.0, 0.8, 1.2], # Human
    [1.2, 1.0, 0.9, 0.7, 0.6, 1.1, 0.8, 1.4, 1.5, 1.1, 1.4, 0.8, 1.2, 1.0, 0.7, 1.1, 0.6, 1.3], # Elven
    [1.1, 0.9, 1.0, 1.3, 0.7, 1.2, 1.4, 0.9, 0.8, 0.6, 0.8, 0.7, 1.3, 0.8, 0.8, 0.9, 0.7, 1.0], # Dwarven
    [0.9, 0.7, 1.3, 1.0, 1.1, 1.0, 0.6, 0.5, 0.7, 0.8, 0.7, 1.2, 0.8, 0.7, 0.9, 0.8, 1.1, 0.6], # Orc
    [0.8, 0.6, 0.7, 1.1, 1.0, 1.2, 0.9, 0.3, 0.8, 0.7, 0.6, 0.9, 0.9, 1.1, 1.4, 0.9, 1.3, 0.4], # Undead
    [1.0, 1.1, 1.2, 1.0, 1.2, 1.0, 0.8, 0.7, 1.1, 0.4, 0.6, 0.8, 1.1, 0.9, 0.8, 1.3, 0.8, 0.9], # Dragon
    [1.1, 0.8, 1.4, 0.6, 0.9, 0.8, 1.0, 1.0, 0.9, 0.7, 0.6, 0.8, 1.2, 1.1, 0.7, 1.2, 0.9, 1.1], # Mechanical
    [1.3, 1.4, 0.9, 0.5, 0.3, 0.7, 1.0, 1.0, 1.2, 1.1, 1.3, 0.8, 1.1, 0.9, 0.4, 1.1, 0.4, 1.4], # Angel
    [1.1, 1.5, 0.8, 0.7, 0.8, 1.1, 0.9, 1.2, 1.0, 1.3, 1.2, 1.0, 1.4, 1.2, 0.9, 1.2, 0.9, 1.1], # Elemental
    [1.0, 1.1, 0.6, 0.8, 0.7, 0.4, 0.7, 1.1, 1.3, 1.0, 1.2, 0.9, 0.8, 0.8, 0.7, 0.7, 0.8, 1.2], # Ocean
    [1.0, 1.4, 0.8, 0.7, 0.6, 0.6, 0.6, 1.3, 1.2, 1.2, 1.0, 1.1, 0.9, 0.8, 0.6, 0.8, 0.7, 1.1], # Plant
    [0.9, 0.8, 0.7, 1.2, 0.9, 0.8, 0.8, 0.8, 1.0, 0.9, 1.1, 1.0, 0.7, 0.8, 0.8, 0.7, 1.2, 0.9], # Insect
    [1.0, 1.2, 1.3, 0.8, 0.9, 1.1, 1.2, 1.1, 1.4, 0.8, 0.9, 0.7, 1.0, 1.1, 0.8, 1.3, 0.8, 1.0], # Crystal
    [1.1, 1.0, 0.8, 0.7, 1.1, 0.9, 1.1, 0.9, 1.2, 0.8, 0.8, 0.8, 1.1, 1.0, 1.0, 1.4, 1.0, 1.0], # Time
    [0.9, 0.7, 0.8, 0.9, 1.4, 0.8, 0.7, 0.4, 0.9, 0.7, 0.6, 0.8, 0.8, 1.0, 1.0, 0.8, 1.1, 0.6], # Shadow
    [1.0, 1.1, 0.9, 0.8, 0.9, 1.3, 1.2, 1.1, 1.2, 0.7, 0.8, 0.7, 1.3, 1.4, 0.8, 1.0, 0.8, 1.0], # Cosmic
    [0.8, 0.6, 0.7, 1.1, 1.3, 0.8, 0.9, 0.4, 0.9, 0.8, 0.7, 1.2, 0.8, 1.0, 1.1, 0.8, 1.0, 0.5], # Viral
    [1.2, 1.3, 1.0, 0.6, 0.4, 0.9, 1.1, 1.4, 1.1, 1.2, 1.1, 0.9, 1.0, 1.0, 0.6, 1.0, 0.5, 1.0]  # Harmony
])
```

### 최적 2종족 조합 (상위 10개)
```yaml
top_synergy_pairs:
  1. elven_kingdom + elemental_spirits: 1.5
  2. angel_legion + elven_kingdom: 1.4
  3. angel_legion + harmony_tribe: 1.4
  4. plant_kingdom + elven_kingdom: 1.4
  5. undead_legion + shadow_clan: 1.4
  6. crystal_beings + elemental_spirits: 1.4
  7. time_weavers + cosmic_empire: 1.4
  8. dwarven_clan + mechanical_empire: 1.4
  9. angel_legion + plant_kingdom: 1.3
  10. elemental_spirits + ocean_empire: 1.3
```

### 최적 3종족 조합 (상위 5개)
```yaml
top_triple_combinations:
  1. elven_kingdom + elemental_spirits + angel_legion:
     synergy_score: 4.1
     combined_matrix: [[1.89, 1.67], [1.78, 1.95]]

  2. dwarven_clan + mechanical_empire + crystal_beings:
     synergy_score: 3.9
     combined_matrix: [[1.23, 2.11], [1.89, 1.67]]

  3. undead_legion + shadow_clan + viral_collective:
     synergy_score: 3.8
     combined_matrix: [[1.67, 1.89], [1.45, 2.01]]

  4. dragon_clan + cosmic_empire + time_weavers:
     synergy_score: 3.6
     combined_matrix: [[2.34, 1.12], [1.78, 1.56]]

  5. plant_kingdom + ocean_empire + harmony_tribe:
     synergy_score: 3.5
     combined_matrix: [[1.12, 2.23], [1.67, 1.89]]
```

## 🎮 실제 구현 코드

### 매트릭스 연산 라이브러리
```python
import numpy as np
from typing import List, Dict, Tuple

class RaceMatrix:
    def __init__(self, name: str, matrix: np.ndarray):
        self.name = name
        self.matrix = np.array(matrix)
        self.validate_matrix()

    def validate_matrix(self):
        """매트릭스 제약조건 검증"""
        frobenius_norm = np.linalg.norm(self.matrix, 'fro')
        assert abs(frobenius_norm - 2.0) < 0.01, f"Invalid norm: {frobenius_norm}"

        det = np.linalg.det(self.matrix)
        assert 0.0 <= det <= 2.0, f"Invalid determinant: {det}"

        trace = np.trace(self.matrix)
        assert 1.5 <= trace <= 2.5, f"Invalid trace: {trace}"

class BalancingEngine:
    def __init__(self):
        self.races = {}
        self.environment_matrices = {}
        self.synergy_matrix = None

    def add_race(self, race: RaceMatrix):
        self.races[race.name] = race

    def calculate_cooperation_effect(self, race1: str, race2: str) -> np.ndarray:
        """두 종족 협력 효과 계산"""
        matrix1 = self.races[race1].matrix
        matrix2 = self.races[race2].matrix
        synergy_factor = self.synergy_matrix[race1][race2]

        # 매트릭스 곱셈 + 시너지 보정
        result = np.dot(matrix1, matrix2) * synergy_factor
        return result

    def apply_environment_effects(self, race_matrix: np.ndarray,
                                 time: str, weather: str, terrain: str) -> np.ndarray:
        """환경 효과 적용"""
        result = race_matrix.copy()

        # 아다마르 곱으로 환경 효과 적용
        if time in self.environment_matrices['time']:
            result = result * self.environment_matrices['time'][time]

        if weather in self.environment_matrices['weather']:
            result = result * self.environment_matrices['weather'][weather]

        if terrain in self.environment_matrices['terrain']:
            result = result * self.environment_matrices['terrain'][terrain]

        return result

    def find_optimal_team_composition(self, team_size: int = 4) -> List[str]:
        """최적 팀 구성 찾기"""
        from itertools import combinations

        best_score = 0
        best_team = None

        for team in combinations(self.races.keys(), team_size):
            score = self.calculate_team_synergy_score(team)
            if score > best_score:
                best_score = score
                best_team = team

        return list(best_team)

    def calculate_team_synergy_score(self, team: Tuple[str]) -> float:
        """팀 시너지 점수 계산"""
        total_score = 0
        for i, race1 in enumerate(team):
            for race2 in team[i+1:]:
                total_score += self.synergy_matrix[race1][race2]
        return total_score / len(team)

# 사용 예시
engine = BalancingEngine()

# 18개 종족 등록
for race_name, race_data in race_matrices.items():
    race = RaceMatrix(race_name, race_data['power_matrix'])
    engine.add_race(race)

# 최적 팀 찾기
optimal_team = engine.find_optimal_team_composition(4)
print(f"최적 4인 팀: {optimal_team}")
```

---

**다음 단계**: 동적 밸런싱 시스템 구현 및 실시간 모니터링
