# Defense Allies 타워 매트릭스 시스템

## 📋 문서 정보
- **작성일**: 2024년
- **버전**: v1.0
- **목적**: 종족 매트릭스 기반 타워 밸런스 매트릭스 설계 시스템
- **기반**: [18종족 매트릭스 최적화](18-race-matrix-optimization.md)

## 🎯 타워 매트릭스 설계 원칙

### 핵심 개념
1. **종족 기반 상속**: 각 타워는 소속 종족의 매트릭스 특성을 기본으로 함
2. **티어별 차별화**: Basic → Advanced → Cooperation 순으로 특화도 증가
3. **역할별 특성**: 공격/방어/유틸/시너지 중 특정 역할에 특화
4. **비용 대비 효율**: 건설 비용과 매트릭스 파워의 균형
5. **협력 메커니즘**: Cooperation 타워는 다종족 협력 시에만 건설 가능

### 매트릭스 상속 구조
```yaml
타워 매트릭스 = 종족 기본 매트릭스 × 티어 계수 × 역할 특화 × 비용 보정

where:
- 종족 기본 매트릭스: 해당 종족의 power_matrix
- 티어 계수: Basic(0.8), Advanced(1.2), Cooperation(1.5)
- 역할 특화: 특정 매트릭스 요소 강화/약화
- 비용 보정: 건설 비용에 따른 파워 조정
```

## 🏗️ 티어별 타워 설계

### Basic 타워 (기본형)
```yaml
설계 원칙:
  - 종족 매트릭스의 80% 파워
  - 균형잡힌 능력치 분배
  - 저렴한 비용으로 접근성 확보
  - 모든 상황에서 안정적 성능

계산 공식:
  basic_matrix = race_matrix × 0.8 × balance_modifier

balance_modifier: [[1.0, 1.0], [1.0, 1.0]]  # 균형 유지
```

### Advanced 타워 (고급형)
```yaml
설계 원칙:
  - 종족 매트릭스의 120% 파워
  - 특정 역할에 특화
  - 높은 비용, 높은 성능
  - 전략적 선택 필요

계산 공식:
  advanced_matrix = race_matrix × 1.2 × specialization_modifier

specialization_modifier 예시:
  - 공격 특화: [[1.5, 0.7], [1.2, 0.8]]
  - 방어 특화: [[0.7, 1.5], [0.8, 1.2]]
  - 유틸 특화: [[0.9, 1.1], [1.4, 1.0]]
```

### Cooperation 타워 (협력형)
```yaml
설계 원칙:
  - 종족 매트릭스의 150% 파워
  - 다종족 협력 시에만 건설 가능
  - 극도로 특화된 성능
  - 팀 시너지 극대화

계산 공식:
  cooperation_matrix = race_matrix × 1.5 × cooperation_modifier

cooperation_modifier 예시:
  - 시너지 극대화: [[1.0, 1.0], [1.8, 1.8]]
  - 극한 특화: [[2.0, 0.5], [1.5, 1.0]]
```

## 🎮 종족별 타워 매트릭스 설계

### 1. 휴먼 연합 (Human Alliance)
```yaml
종족 기본 매트릭스: [[1.0, 1.0], [1.0, 1.0]]

Basic 타워:
  knight_fortress:
    matrix: [[0.8, 0.8], [0.8, 0.8]]
    role: "균형형 방어"
    cost: {gold: 100, mana: 50}

  merchant_guild:
    matrix: [[0.6, 0.9], [0.9, 0.9]]
    role: "자원 생산"
    cost: {gold: 80, mana: 40}

  mage_tower:
    matrix: [[0.9, 0.7], [0.8, 0.8]]
    role: "마법 공격"
    cost: {gold: 120, mana: 80}

Advanced 타워:
  castle_walls:
    matrix: [[0.6, 1.8], [1.0, 1.2]]
    role: "극한 방어"
    cost: {gold: 300, mana: 150}

  cathedral:
    matrix: [[0.8, 1.0], [1.6, 1.4]]
    role: "팀 버프"
    cost: {gold: 250, mana: 200}

  royal_palace:
    matrix: [[1.4, 1.0], [1.2, 1.2]]
    role: "지휘 중심"
    cost: {gold: 400, mana: 200}

Cooperation 타워:
  alliance_fortress:
    matrix: [[1.2, 1.2], [1.8, 1.8]]
    role: "다종족 협력 거점"
    cost: {gold: 600, mana: 400}
    requirements: {cooperation_players: 2}

  peace_tower:
    matrix: [[0.8, 1.2], [2.2, 2.0]]
    role: "평화 협정 효과"
    cost: {gold: 500, mana: 500}
    requirements: {cooperation_players: 3}

  unity_command:
    matrix: [[1.5, 1.5], [2.0, 2.0]]
    role: "통합 지휘소"
    cost: {gold: 800, mana: 600}
    requirements: {cooperation_players: 4}
```

### 2. 드래곤 종족 (Dragon Clan)
```yaml
종족 기본 매트릭스: [[1.5, 0.5], [1.3, 0.7]]

Basic 타워:
  fire_spire:
    matrix: [[1.2, 0.4], [1.0, 0.6]]
    role: "화염 공격"
    cost: {gold: 150, mana: 100}

  dragon_nest:
    matrix: [[1.0, 0.6], [1.2, 0.4]]
    role: "드래곤 소환"
    cost: {gold: 200, mana: 150}

  treasure_vault:
    matrix: [[0.8, 0.8], [0.8, 0.8]]
    role: "자원 저장"
    cost: {gold: 100, mana: 50}

Advanced 타워:
  inferno_citadel:
    matrix: [[2.4, 0.3], [1.8, 0.6]]
    role: "극한 화력"
    cost: {gold: 500, mana: 400}

  ancient_lair:
    matrix: [[1.8, 0.6], [1.5, 1.2]]
    role: "고대 드래곤"
    cost: {gold: 600, mana: 500}

  molten_forge:
    matrix: [[1.2, 0.9], [1.8, 0.9]]
    role: "장비 강화"
    cost: {gold: 400, mana: 300}

Cooperation 타워:
  dragon_alliance:
    matrix: [[2.2, 0.8], [2.0, 1.0]]
    role: "드래곤 연합"
    cost: {gold: 800, mana: 600}
    requirements: {cooperation_players: 2}

  elemental_fusion:
    matrix: [[1.8, 1.2], [1.5, 1.5]]
    role: "원소 융합"
    cost: {gold: 700, mana: 700}
    requirements: {cooperation_players: 3}

  apocalypse_engine:
    matrix: [[3.0, 0.5], [2.5, 1.0]]
    role: "종말 병기"
    cost: {gold: 1200, mana: 1000}
    requirements: {cooperation_players: 4}
```

### 3. 엘프 왕국 (Elven Kingdom)
```yaml
종족 기본 매트릭스: [[1.3, 0.7], [1.2, 0.8]]

Basic 타워:
  archer_post:
    matrix: [[1.0, 0.6], [1.0, 0.6]]
    role: "원거리 공격"
    cost: {gold: 120, mana: 60}

  tree_sanctuary:
    matrix: [[0.8, 0.8], [1.2, 0.8]]
    role: "자연 치유"
    cost: {gold: 100, mana: 80}

  wind_shrine:
    matrix: [[1.2, 0.4], [0.8, 0.8]]
    role: "속도 버프"
    cost: {gold: 140, mana: 100}

Advanced 타워:
  moonwell_spire:
    matrix: [[1.0, 1.2], [1.8, 1.2]]
    role: "달빛 마법"
    cost: {gold: 350, mana: 300}

  ancient_grove:
    matrix: [[1.2, 0.9], [1.8, 1.5]]
    role: "고대 숲"
    cost: {gold: 400, mana: 350}

  starfall_tower:
    matrix: [[2.0, 0.6], [1.5, 0.9]]
    role: "별빛 공격"
    cost: {gold: 450, mana: 400}

Cooperation 타워:
  nature_alliance:
    matrix: [[1.5, 1.2], [2.4, 1.8]]
    role: "자연 연합"
    cost: {gold: 600, mana: 500}
    requirements: {cooperation_players: 2}

  world_tree:
    matrix: [[1.8, 1.5], [2.0, 2.4]]
    role: "세계수"
    cost: {gold: 800, mana: 800}
    requirements: {cooperation_players: 3}

  harmony_nexus:
    matrix: [[2.2, 1.8], [2.5, 2.0]]
    role: "조화의 중심"
    cost: {gold: 1000, mana: 900}
    requirements: {cooperation_players: 4}
```

## ⚖️ 타워 밸런싱 규칙

### 1. 비용 대비 효율성
```python
def calculate_cost_efficiency(tower_matrix: np.ndarray, cost: Dict[str, int]) -> float:
    """타워의 비용 대비 효율성 계산"""

    # 매트릭스 파워 계산
    matrix_power = np.linalg.norm(tower_matrix, 'fro')

    # 총 비용 계산 (금 + 마나*1.5)
    total_cost = cost['gold'] + cost['mana'] * 1.5

    # 효율성 = 파워 / 비용
    efficiency = matrix_power / total_cost

    return efficiency

# 모든 타워의 효율성이 비슷해야 함 (±20% 범위)
target_efficiency = 0.01  # 기준 효율성
```

### 2. 티어별 파워 제약
```python
def validate_tier_power(tower_matrix: np.ndarray, tier: str, race_matrix: np.ndarray) -> bool:
    """티어별 파워 제약 검증"""

    tower_power = np.linalg.norm(tower_matrix, 'fro')
    race_power = np.linalg.norm(race_matrix, 'fro')

    power_ratio = tower_power / race_power

    tier_constraints = {
        'basic': (0.7, 0.9),      # 70-90%
        'advanced': (1.1, 1.3),   # 110-130%
        'cooperation': (1.4, 1.6) # 140-160%
    }

    min_ratio, max_ratio = tier_constraints[tier]
    return min_ratio <= power_ratio <= max_ratio
```

### 3. 역할별 특화도
```python
def calculate_specialization_score(tower_matrix: np.ndarray) -> Dict[str, float]:
    """타워의 역할별 특화도 계산"""

    # 매트릭스 요소별 가중치
    offensive_score = tower_matrix[0, 0] * 0.6 + tower_matrix[1, 0] * 0.4
    defensive_score = tower_matrix[0, 1] * 0.6 + tower_matrix[1, 1] * 0.4
    individual_score = tower_matrix[0, 0] * 0.5 + tower_matrix[0, 1] * 0.5
    cooperation_score = tower_matrix[1, 0] * 0.5 + tower_matrix[1, 1] * 0.5

    return {
        'offensive': offensive_score,
        'defensive': defensive_score,
        'individual': individual_score,
        'cooperation': cooperation_score
    }
```

## 🔧 타워 매트릭스 생성 도구

### 자동 생성 시스템
```python
class TowerMatrixGenerator:
    """타워 매트릭스 자동 생성기"""

    def __init__(self):
        self.tier_multipliers = {
            'basic': 0.8,
            'advanced': 1.2,
            'cooperation': 1.5
        }

        self.role_modifiers = {
            'balanced': [[1.0, 1.0], [1.0, 1.0]],
            'offensive': [[1.5, 0.7], [1.2, 0.8]],
            'defensive': [[0.7, 1.5], [0.8, 1.2]],
            'utility': [[0.9, 1.1], [1.4, 1.0]],
            'synergy': [[1.0, 1.0], [1.6, 1.6]]
        }

    def generate_tower_matrix(self,
                            race_matrix: np.ndarray,
                            tier: str,
                            role: str,
                            cost: Dict[str, int]) -> np.ndarray:
        """타워 매트릭스 생성"""

        # 1. 기본 계산
        base_matrix = race_matrix * self.tier_multipliers[tier]

        # 2. 역할 특화 적용
        role_modifier = np.array(self.role_modifiers[role])
        specialized_matrix = base_matrix * role_modifier

        # 3. 비용 보정
        cost_factor = self.calculate_cost_factor(cost, tier)
        final_matrix = specialized_matrix * cost_factor

        # 4. 제약 조건 검증 및 조정
        final_matrix = self.apply_constraints(final_matrix, tier, race_matrix)

        return final_matrix

    def calculate_cost_factor(self, cost: Dict[str, int], tier: str) -> float:
        """비용에 따른 보정 계수"""
        total_cost = cost['gold'] + cost['mana'] * 1.5

        # 티어별 기준 비용
        base_costs = {
            'basic': 150,
            'advanced': 400,
            'cooperation': 700
        }

        base_cost = base_costs[tier]
        cost_ratio = total_cost / base_cost

        # 비용이 높을수록 파워 증가 (제한적)
        return min(1.0 + (cost_ratio - 1.0) * 0.3, 1.5)

    def apply_constraints(self, matrix: np.ndarray, tier: str, race_matrix: np.ndarray) -> np.ndarray:
        """제약 조건 적용"""

        # 파워 제약
        current_power = np.linalg.norm(matrix, 'fro')
        race_power = np.linalg.norm(race_matrix, 'fro')

        tier_constraints = {
            'basic': (0.7, 0.9),
            'advanced': (1.1, 1.3),
            'cooperation': (1.4, 1.6)
        }

        min_ratio, max_ratio = tier_constraints[tier]
        target_power = race_power * ((min_ratio + max_ratio) / 2)

        # 파워 조정
        if current_power != 0:
            adjustment_factor = target_power / current_power
            matrix = matrix * adjustment_factor

        return matrix

# 사용 예시
generator = TowerMatrixGenerator()

# 휴먼 기본 타워 생성
human_matrix = np.array([[1.0, 1.0], [1.0, 1.0]])
knight_fortress_matrix = generator.generate_tower_matrix(
    race_matrix=human_matrix,
    tier='basic',
    role='defensive',
    cost={'gold': 100, 'mana': 50}
)

print(f"Knight Fortress Matrix:\n{knight_fortress_matrix}")
```

## 📊 타워 매트릭스 검증

### 밸런스 검증 도구
```python
class TowerBalanceValidator:
    """타워 밸런스 검증기"""

    def validate_race_towers(self, race_id: str, towers: List[Dict]) -> Dict:
        """종족 내 타워들의 밸런스 검증"""

        results = {
            'efficiency_balance': self.check_efficiency_balance(towers),
            'tier_progression': self.check_tier_progression(towers),
            'role_coverage': self.check_role_coverage(towers),
            'cost_scaling': self.check_cost_scaling(towers)
        }

        overall_score = sum(results.values()) / len(results)
        results['overall_balance'] = overall_score

        return results

    def check_efficiency_balance(self, towers: List[Dict]) -> float:
        """효율성 균형 검사"""
        efficiencies = []

        for tower in towers:
            matrix = np.array(tower['matrix'])
            cost = tower['cost']
            efficiency = calculate_cost_efficiency(matrix, cost)
            efficiencies.append(efficiency)

        # 효율성 분산이 낮을수록 좋음
        variance = np.var(efficiencies)
        balance_score = 1.0 / (1.0 + variance * 1000)

        return balance_score

    def generate_balance_report(self, all_races_towers: Dict) -> str:
        """전체 밸런스 리포트 생성"""

        report = "# Tower Balance Report\n\n"

        for race_id, towers in all_races_towers.items():
            validation = self.validate_race_towers(race_id, towers)

            report += f"## {race_id}\n"
            report += f"- Overall Balance: {validation['overall_balance']:.2f}\n"
            report += f"- Efficiency Balance: {validation['efficiency_balance']:.2f}\n"
            report += f"- Tier Progression: {validation['tier_progression']:.2f}\n"
            report += f"- Role Coverage: {validation['role_coverage']:.2f}\n"
            report += f"- Cost Scaling: {validation['cost_scaling']:.2f}\n\n"

        return report
```

## 🏭 전체 타워 매트릭스 생성 시스템

### 18개 종족 × 9개 타워 = 162개 타워 자동 생성

```python
class CompleteTowerSystem:
    """전체 타워 시스템 생성기"""

    def __init__(self):
        self.generator = TowerMatrixGenerator()
        self.validator = TowerBalanceValidator()

        # 18개 종족 기본 매트릭스 (이전에 최적화된 값들)
        self.race_matrices = {
            'human_alliance': np.array([[1.0, 1.0], [1.0, 1.0]]),
            'elven_kingdom': np.array([[1.3, 0.7], [1.2, 0.8]]),
            'dwarven_clan': np.array([[0.7, 1.3], [0.8, 1.2]]),
            'orc_tribe': np.array([[1.4, 0.6], [1.1, 0.9]]),
            'undead_legion': np.array([[0.9, 1.1], [0.8, 1.2]]),
            'dragon_clan': np.array([[1.5, 0.5], [1.3, 0.7]]),
            'mechanical_empire': np.array([[1.1, 0.9], [1.0, 1.0]]),
            'angel_legion': np.array([[0.8, 1.2], [0.9, 1.1]]),
            'elemental_spirits': np.array([[1.0, 1.0], [0.9, 1.1]]),
            'ocean_empire': np.array([[0.6, 1.4], [1.0, 1.0]]),
            'plant_kingdom': np.array([[0.5, 1.5], [0.7, 1.3]]),
            'insect_swarm': np.array([[1.2, 0.8], [0.6, 1.4]]),
            'crystal_beings': np.array([[0.8, 1.2], [1.1, 0.9]]),
            'time_weavers': np.array([[1.6, 0.4], [0.5, 1.5]]),
            'shadow_clan': np.array([[1.3, 0.7], [0.4, 1.6]]),
            'cosmic_empire': np.array([[1.1, 0.9], [1.3, 0.7]]),
            'viral_collective': np.array([[0.9, 1.1], [1.4, 0.6]]),
            'harmony_tribe': np.array([[0.7, 1.3], [1.2, 0.8]])
        }

        # 각 종족별 타워 템플릿
        self.tower_templates = {
            'basic': [
                {'role': 'balanced', 'cost_range': (80, 120)},
                {'role': 'offensive', 'cost_range': (100, 140)},
                {'role': 'defensive', 'cost_range': (90, 130)}
            ],
            'advanced': [
                {'role': 'offensive', 'cost_range': (300, 450)},
                {'role': 'defensive', 'cost_range': (280, 420)},
                {'role': 'utility', 'cost_range': (320, 480)}
            ],
            'cooperation': [
                {'role': 'synergy', 'cost_range': (500, 700)},
                {'role': 'offensive', 'cost_range': (600, 800)},
                {'role': 'utility', 'cost_range': (550, 750)}
            ]
        }

    def generate_all_towers(self) -> Dict[str, List[Dict]]:
        """모든 종족의 모든 타워 생성"""

        all_towers = {}

        for race_id, race_matrix in self.race_matrices.items():
            race_towers = []

            for tier, templates in self.tower_templates.items():
                for i, template in enumerate(templates):
                    tower = self.generate_single_tower(
                        race_id, race_matrix, tier, i, template
                    )
                    race_towers.append(tower)

            all_towers[race_id] = race_towers

        return all_towers

    def generate_single_tower(self, race_id: str, race_matrix: np.ndarray,
                            tier: str, index: int, template: Dict) -> Dict:
        """단일 타워 생성"""

        # 비용 계산
        cost_min, cost_max = template['cost_range']
        base_cost = np.random.randint(cost_min, cost_max + 1)

        # 금/마나 비율 (종족별 특성 반영)
        mana_ratio = self.get_race_mana_ratio(race_id)
        gold_cost = int(base_cost * (1 - mana_ratio))
        mana_cost = int(base_cost * mana_ratio)

        cost = {'gold': gold_cost, 'mana': mana_cost}

        # 매트릭스 생성
        tower_matrix = self.generator.generate_tower_matrix(
            race_matrix, tier, template['role'], cost
        )

        # 타워 정보 구성
        tower_name = self.generate_tower_name(race_id, tier, index)
        tower_id = f"{race_id}_{tier}_{index + 1}"

        return {
            'id': tower_id,
            'name': tower_name,
            'race_id': race_id,
            'tier': tier,
            'role': template['role'],
            'matrix': tower_matrix.tolist(),
            'cost': cost,
            'power_rating': self.calculate_tower_power_rating(tower_matrix),
            'specialization': calculate_specialization_score(tower_matrix)
        }

    def get_race_mana_ratio(self, race_id: str) -> float:
        """종족별 마나 의존도"""
        mana_ratios = {
            'human_alliance': 0.4,      # 균형
            'elven_kingdom': 0.6,       # 마법 중심
            'dwarven_clan': 0.2,        # 물리 중심
            'orc_tribe': 0.3,           # 물리 중심
            'undead_legion': 0.5,       # 마법 중심
            'dragon_clan': 0.7,         # 높은 마나
            'mechanical_empire': 0.3,   # 기술 중심
            'angel_legion': 0.6,        # 신성 마법
            'elemental_spirits': 0.8,   # 순수 마법
            'ocean_empire': 0.5,        # 균형
            'plant_kingdom': 0.4,       # 자연 마법
            'insect_swarm': 0.2,        # 생물학적
            'crystal_beings': 0.9,      # 에너지 중심
            'time_weavers': 0.8,        # 시간 마법
            'shadow_clan': 0.7,         # 어둠 마법
            'cosmic_empire': 0.8,       # 우주 에너지
            'viral_collective': 0.3,    # 생물학적
            'harmony_tribe': 0.6        # 음파 마법
        }
        return mana_ratios.get(race_id, 0.5)

    def generate_tower_name(self, race_id: str, tier: str, index: int) -> str:
        """타워 이름 생성"""

        race_themes = {
            'human_alliance': ['Knight', 'Castle', 'Royal', 'Noble', 'Guard'],
            'elven_kingdom': ['Moon', 'Star', 'Wind', 'Tree', 'Leaf'],
            'dwarven_clan': ['Iron', 'Stone', 'Forge', 'Hammer', 'Anvil'],
            'orc_tribe': ['War', 'Blood', 'Rage', 'Skull', 'Bone'],
            'undead_legion': ['Death', 'Shadow', 'Bone', 'Soul', 'Grave'],
            'dragon_clan': ['Fire', 'Flame', 'Inferno', 'Molten', 'Ember'],
            'mechanical_empire': ['Steel', 'Gear', 'Engine', 'Circuit', 'Core'],
            'angel_legion': ['Divine', 'Holy', 'Sacred', 'Blessed', 'Pure'],
            'elemental_spirits': ['Primal', 'Essence', 'Spirit', 'Force', 'Energy'],
            'ocean_empire': ['Tide', 'Wave', 'Deep', 'Current', 'Pearl'],
            'plant_kingdom': ['Root', 'Bloom', 'Thorn', 'Seed', 'Grove'],
            'insect_swarm': ['Hive', 'Swarm', 'Colony', 'Nest', 'Queen'],
            'crystal_beings': ['Crystal', 'Prism', 'Shard', 'Gem', 'Facet'],
            'time_weavers': ['Temporal', 'Chrono', 'Eternal', 'Moment', 'Era'],
            'shadow_clan': ['Void', 'Eclipse', 'Phantom', 'Wraith', 'Shade'],
            'cosmic_empire': ['Stellar', 'Cosmic', 'Nebula', 'Galaxy', 'Void'],
            'viral_collective': ['Strain', 'Mutation', 'Evolution', 'Adaptation', 'Gene'],
            'harmony_tribe': ['Harmony', 'Melody', 'Rhythm', 'Echo', 'Resonance']
        }

        tier_suffixes = {
            'basic': ['Post', 'Tower', 'Outpost', 'Station', 'Base'],
            'advanced': ['Citadel', 'Fortress', 'Stronghold', 'Bastion', 'Keep'],
            'cooperation': ['Nexus', 'Alliance', 'Unity', 'Convergence', 'Synthesis']
        }

        theme = race_themes[race_id][index % len(race_themes[race_id])]
        suffix = tier_suffixes[tier][index % len(tier_suffixes[tier])]

        return f"{theme} {suffix}"

    def calculate_tower_power_rating(self, tower_matrix: np.ndarray) -> float:
        """타워 파워 레이팅 계산"""
        # 기본 파워 계산 (파워 레이팅 시스템 활용)
        base_power = calculate_base_power(tower_matrix)

        # 타워는 환경/시너지 보너스 없이 순수 파워만
        return base_power * 10  # 1000 기준으로 스케일링

# 전체 시스템 실행
def generate_complete_tower_database():
    """완전한 타워 데이터베이스 생성"""

    system = CompleteTowerSystem()

    print("🏗️ 162개 타워 매트릭스 생성 중...")
    all_towers = system.generate_all_towers()

    print("⚖️ 밸런스 검증 중...")
    validator = TowerBalanceValidator()

    total_towers = 0
    balance_scores = []

    for race_id, towers in all_towers.items():
        validation = validator.validate_race_towers(race_id, towers)
        balance_scores.append(validation['overall_balance'])
        total_towers += len(towers)

        print(f"✅ {race_id}: {len(towers)}개 타워, 밸런스 점수: {validation['overall_balance']:.2f}")

    overall_balance = sum(balance_scores) / len(balance_scores)

    print(f"\n🎯 전체 결과:")
    print(f"- 총 타워 수: {total_towers}개")
    print(f"- 평균 밸런스 점수: {overall_balance:.2f}")
    print(f"- 생성 완료!")

    return all_towers

# 실행 예시
if __name__ == "__main__":
    tower_database = generate_complete_tower_database()

    # 샘플 출력
    print("\n📋 샘플 타워 정보:")
    human_towers = tower_database['human_alliance']
    for tower in human_towers[:3]:  # 처음 3개만 출력
        print(f"- {tower['name']} ({tower['tier']})")
        print(f"  매트릭스: {tower['matrix']}")
        print(f"  비용: {tower['cost']}")
        print(f"  파워: {tower['power_rating']:.1f}")
        print()
```

## 📊 타워 데이터베이스 구조

### JSON 출력 형식
```json
{
  "human_alliance": [
    {
      "id": "human_alliance_basic_1",
      "name": "Knight Post",
      "race_id": "human_alliance",
      "tier": "basic",
      "role": "balanced",
      "matrix": [[0.8, 0.8], [0.8, 0.8]],
      "cost": {"gold": 60, "mana": 40},
      "power_rating": 160.0,
      "specialization": {
        "offensive": 0.8,
        "defensive": 0.8,
        "individual": 0.8,
        "cooperation": 0.8
      },
      "abilities": [
        {
          "id": "defensive_stance",
          "name": "방어 태세",
          "effect_matrix": [[0.7, 1.3], [0.8, 1.2]],
          "cooldown": 30,
          "duration": 15
        }
      ]
    }
  ]
}
```

### 타워 능력 시스템
```python
class TowerAbilitySystem:
    """타워 능력 시스템"""

    def __init__(self):
        self.ability_templates = {
            'offensive': [
                {'name': '집중 공격', 'matrix_mod': [[1.5, 0.8], [1.2, 0.9]]},
                {'name': '연속 공격', 'matrix_mod': [[1.3, 0.9], [1.1, 1.0]]},
                {'name': '범위 공격', 'matrix_mod': [[1.2, 1.0], [1.4, 1.1]]}
            ],
            'defensive': [
                {'name': '방어 태세', 'matrix_mod': [[0.7, 1.3], [0.8, 1.2]]},
                {'name': '보호막', 'matrix_mod': [[0.8, 1.4], [0.9, 1.3]]},
                {'name': '재생', 'matrix_mod': [[0.9, 1.2], [1.0, 1.1]]}
            ],
            'utility': [
                {'name': '자원 생산', 'matrix_mod': [[0.8, 0.8], [1.2, 1.2]]},
                {'name': '속도 증가', 'matrix_mod': [[1.1, 1.1], [1.2, 1.2]]},
                {'name': '시야 확장', 'matrix_mod': [[1.0, 1.0], [1.3, 1.3]]}
            ]
        }

    def generate_tower_abilities(self, tower: Dict) -> List[Dict]:
        """타워 능력 생성"""

        role = tower['role']
        tier = tower['tier']

        # 티어별 능력 개수
        ability_counts = {'basic': 1, 'advanced': 2, 'cooperation': 3}
        num_abilities = ability_counts[tier]

        abilities = []
        available_abilities = self.ability_templates.get(role, self.ability_templates['offensive'])

        for i in range(num_abilities):
            template = available_abilities[i % len(available_abilities)]

            ability = {
                'id': f"{template['name'].lower().replace(' ', '_')}",
                'name': template['name'],
                'effect_matrix': template['matrix_mod'],
                'cooldown': 30 + i * 15,  # 30, 45, 60초
                'duration': 10 + i * 5    # 10, 15, 20초
            }

            abilities.append(ability)

        return abilities

# 능력 포함 완전 생성
def generate_complete_towers_with_abilities():
    """능력 포함 완전 타워 생성"""

    system = CompleteTowerSystem()
    ability_system = TowerAbilitySystem()

    all_towers = system.generate_all_towers()

    # 각 타워에 능력 추가
    for race_id, towers in all_towers.items():
        for tower in towers:
            tower['abilities'] = ability_system.generate_tower_abilities(tower)

    return all_towers
```

## 🎯 타워 시스템 완성도

### 생성된 타워 통계
```yaml
총 타워 수: 162개 (18종족 × 9타워)

티어별 분포:
- Basic: 54개 (18종족 × 3타워)
- Advanced: 54개 (18종족 × 3타워)
- Cooperation: 54개 (18종족 × 3타워)

역할별 분포:
- Balanced: 18개
- Offensive: 36개
- Defensive: 36개
- Utility: 36개
- Synergy: 36개

예상 밸런스 점수: 0.85+ (85% 이상)
```

### 시스템 특징
1. **자동 생성**: 종족 매트릭스 기반 완전 자동화
2. **밸런스 보장**: 비용 대비 효율성 균등화
3. **역할 특화**: 명확한 역할별 차별화
4. **확장 가능**: 새로운 종족/타워 쉽게 추가
5. **검증 시스템**: 자동 밸런스 검증 및 리포트

**Defense Allies의 타워 시스템이 이제 완전히 체계화되었습니다!** 🏗️

---

**다음 단계**: 타워 업그레이드 시스템 및 동적 매트릭스 변화 구현
