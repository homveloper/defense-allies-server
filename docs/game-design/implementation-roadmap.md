# Defense Allies 구현 로드맵

## 📋 기획 레벨 TODO 리스트

### Phase 1: 핵심 시스템 설계 (2-3주)

#### 🎭 종족 시스템 상세 설계
- [ ] **종족별 밸런스 시트 작성**
  - 8개 종족의 수치적 밸런스 검증
  - 종족 간 상성 관계 매트릭스 작성
  - 종족별 고유 타워 스펙 정의
  - 종족 시너지 조합 효과 계산

- [ ] **종족별 타워 트리 설계**
  - 각 종족당 10-15개 고유 타워 설계
  - 타워 업그레이드 경로 정의
  - 종족 특화 능력 상세 기획
  - 협력 타워 건설 조건 설정

- [ ] **종족 선택 UI/UX 플로우**
  - 종족 선택 화면 와이어프레임
  - 종족 정보 표시 방식 설계
  - 팀 구성 추천 시스템 기획
  - 종족별 튜토리얼 시나리오

#### 🌍 환경 변수 시스템 설계
- [ ] **환경 요소 조합 매트릭스**
  - 시간대 × 날씨 × 지형 조합표 (5×6×5 = 150가지)
  - 각 조합별 종족 영향도 계산
  - 밸런스 브레이킹 조합 식별 및 조정
  - 플레이어 선호도 예측 모델

- [ ] **동적 환경 변화 시스템**
  - 게임 중 환경 변화 타이밍 설계
  - 환경 변화 예고 시스템 기획
  - 플레이어 적응 시간 계산
  - 환경 변화에 따른 전략 변경 가이드

- [ ] **트레이드오프 메커니즘**
  - 각 환경에서의 최적/최악 종족 조합
  - 불리한 환경 극복 방법 설계
  - 환경 적응 보상 시스템
  - 창의적 해결책 인센티브 구조

#### 📊 난이도 시스템 설계
- [ ] **1-100 난이도 레벨 상세 기획**
  - 각 레벨별 적 스펙 공식 정의
  - 환경 복잡도 증가 곡선 설계
  - 협력 요구 수준 단계별 정의
  - 보상 체계 및 언락 컨텐츠 매핑

- [ ] **적응형 난이도 알고리즘**
  - 팀 성과 측정 지표 정의
  - 실시간 난이도 조절 공식
  - 학습 곡선 최적화 방법
  - 좌절감 방지 메커니즘

### Phase 2: 게임플레이 메커니즘 설계 (3-4주)

#### 🤝 협력 시스템 심화 설계
- [ ] **종족 간 협력 메커니즘**
  - 종족별 협력 특화 능력 설계
  - 다종족 연합 보너스 계산
  - 협력 실패 페널티 시스템
  - 협력 성공 피드백 시스템

- [ ] **실시간 의사결정 시스템**
  - 긴급 상황 대응 프로토콜
  - 팀 투표 시스템 설계
  - 리더십 역할 순환 메커니즘
  - 갈등 해결 시스템

#### 🎮 게임플레이 루프 최적화
- [ ] **매치 구성 알고리즘**
  - 종족 조합 밸런싱 로직
  - 실력 매칭 시스템
  - 환경 선호도 고려 매칭
  - 재매칭 방지 시스템

- [ ] **진행도 및 성장 시스템**
  - 개인/팀 경험치 시스템
  - 종족 숙련도 시스템
  - 환경 적응도 시스템
  - 협력 스킬 발전 트리

### Phase 3: 사용자 경험 설계 (2-3주)

#### 📱 UI/UX 상세 설계
- [ ] **종족별 UI 테마**
  - 8개 종족별 고유 UI 스타일
  - 종족 특성 반영 인터페이스
  - 접근성 고려 디자인
  - 다국어 지원 설계

- [ ] **환경 정보 표시 시스템**
  - 실시간 환경 상태 표시
  - 종족별 영향도 시각화
  - 환경 변화 예고 시스템
  - 적응 가이드 표시

#### 🎓 온보딩 및 학습 시스템
- [ ] **단계별 튜토리얼**
  - 종족별 맞춤 튜토리얼
  - 환경 적응 훈련 모드
  - 협력 스킬 연습 시스템
  - AI 파트너 훈련 모드

---

## 💻 개발 단계 TODO 리스트

### 📊 게임 정적 데이터 포맷 정의

#### 선택된 기술 스택
- **포맷**: JSON Schema + JSON
- **스키마 도구**: https://json-schema.app/ (시각화)
- **검증 도구**: jsonschema 라이브러리 (Python/Node.js)
- **에디팅**: VS Code + JSON Schema 확장

#### 데이터 구조 설계 원칙
- **관계 표현**: `$ref`를 이용한 참조 관계
- **검증**: JSON Schema로 타입 및 제약 조건 정의
- **확장성**: 추후 데이터 추가를 고려한 유연한 구조
- **성능**: 클라이언트 메모리 제약을 고려한 효율적 구조

#### 정적 데이터 카테고리
1. **타워 데이터**: 스탯, 비용, 레벨링, 능력, 시너지
2. **종족 데이터**: 특성, 보너스, 고유 타워 목록
3. **환경 데이터**: 시간대/날씨/지형 효과
4. **난이도 데이터**: 레벨별 스케일링 공식
5. **로컬라이제이션**: 다국어 문자열 키

### Phase 1: 기반 인프라 구축 (4-6주)

#### 🏗️ 서버 아키텍처 구현

**GuardianApp (인증 전용)**
- [ ] **사용자 인증 시스템**
  ```go
  // 우선순위: 높음
  type AuthService struct {
      redis    *redis.Client
      jwtKey   []byte
  }

  func (s *AuthService) AuthenticateUser(credentials *UserCredentials) (*AuthToken, error)
  func (s *AuthService) ValidateToken(token string) (*UserClaims, error)
  func (s *AuthService) RefreshToken(refreshToken string) (*AuthToken, error)
  ```

**TimeSquareApp (게임 로직)**
- [ ] **종족 시스템 데이터 구조**
  ```go
  // 우선순위: 높음
  type Race struct {
      ID           string
      Name         string
      Strengths    map[string]float64
      Weaknesses   map[string]float64
      UniqueTowers []TowerType
      Abilities    []RaceAbility
  }
  ```

- [ ] **환경 시스템 엔진**
  ```go
  // 우선순위: 높음
  type Environment struct {
      TimeOfDay     TimeType
      Weather       WeatherType
      Terrain       TerrainType
      MagicLevel    MagicType
      SpecialEvents []EventType
      Effects       map[string]map[string]float64
  }
  ```

- [ ] **난이도 시스템 구현**
  ```go
  // 우선순위: 중간
  type DifficultyLevel struct {
      Level                int
      EnemyScaling        EnemyScalingConfig
      EnvironmentalEffects EnvironmentConfig
      CooperationRequirements CoopConfig
      Rewards             RewardConfig
  }
  ```

#### 📊 Redis 데이터 모델링

**GuardianApp 데이터 스키마**
- [ ] **인증 데이터 스키마**
  - `auth:user:{userId}` - 사용자 기본 정보
  - `auth:session:{sessionId}` - 로그인 세션 정보
  - `auth:token:{tokenId}` - JWT 토큰 관리
  - `auth:refresh:{userId}` - 리프레시 토큰

**TimeSquareApp 데이터 스키마**
- [ ] **종족 데이터 스키마**
  - `race:config:{raceId}` - 종족 기본 설정
  - `race:towers:{raceId}` - 종족별 타워 정보
  - `race:abilities:{raceId}` - 종족 특수 능력
  - `race:synergy:combinations` - 종족 조합 효과

- [ ] **환경 데이터 스키마**
  - `environment:templates` - 환경 템플릿 목록
  - `environment:effects:{envId}` - 환경별 효과
  - `game:environment:{gameId}` - 게임별 환경 상태
  - `environment:history:{gameId}` - 환경 변화 이력

- [ ] **난이도 데이터 스키마**
  - `difficulty:levels` - 난이도 레벨 설정
  - `difficulty:scaling:{level}` - 레벨별 스케일링
  - `player:progress:{playerId}` - 플레이어 진행도
  - `team:difficulty:{gameId}` - 팀 난이도 상태

**CommandApp 데이터 스키마**
- [ ] **관리자 데이터 스키마**
  - `admin:stats:{date}` - 일별 통계 데이터
  - `admin:balance:{raceId}` - 종족별 밸런스 데이터
  - `admin:reports:{reportId}` - 분석 리포트
  - `admin:config:global` - 전역 설정

### Phase 2: 핵심 게임 로직 구현 (6-8주)

#### 🎭 종족 시스템 구현
- [ ] **종족 선택 및 매칭**
  ```go
  // 우선순위: 높음
  func (s *MatchService) CreateBalancedMatch(players []Player) (*Match, error)
  func (s *RaceService) ValidateTeamComposition(races []RaceType) bool
  func (s *RaceService) CalculateTeamSynergy(races []RaceType) float64
  ```

- [ ] **종족별 타워 시스템**
  ```go
  // 우선순위: 높음
  func (s *TowerService) BuildRaceTower(raceId, towerType string) (*Tower, error)
  func (s *TowerService) CalculateRaceModifiers(tower *Tower) *TowerStats
  func (s *TowerService) CheckBuildPermissions(playerId, towerType string) bool
  ```

- [ ] **종족 능력 시스템**
  ```go
  // 우선순위: 중간
  func (s *AbilityService) ActivateRaceAbility(playerId, abilityId string) error
  func (s *AbilityService) CalculateAbilityCooldown(raceId, abilityId string) time.Duration
  func (s *AbilityService) CheckAbilityRequirements(playerId, abilityId string) bool
  ```

#### 🌍 환경 시스템 구현
- [ ] **환경 생성 및 관리**
  ```go
  // 우선순위: 높음
  func (s *EnvironmentService) GenerateRandomEnvironment() *Environment
  func (s *EnvironmentService) ApplyEnvironmentalEffects(gameId string) error
  func (s *EnvironmentService) UpdateEnvironment(gameId string, changes *EnvironmentChange) error
  ```

- [ ] **환경 효과 계산**
  ```go
  // 우선순위: 높음
  func (s *EffectService) CalculateRaceEnvironmentModifiers(raceId string, env *Environment) map[string]float64
  func (s *EffectService) ApplyEnvironmentToTower(tower *Tower, env *Environment) *TowerStats
  func (s *EffectService) GetEnvironmentPenalties(raceId string, env *Environment) []Penalty
  ```

#### 📊 난이도 시스템 구현
- [ ] **적응형 난이도 엔진**
  ```go
  // 우선순위: 중간
  func (s *DifficultyService) CalculateTeamPerformance(gameId string) *PerformanceMetrics
  func (s *DifficultyService) AdjustDifficulty(gameId string, performance *PerformanceMetrics) error
  func (s *DifficultyService) GenerateWaveForLevel(level int, teamComposition []RaceType) *Wave
  ```

### Phase 3: 실시간 시스템 구현 (4-5주)

#### 🔄 실시간 이벤트 처리
- [ ] **종족별 이벤트 스트리밍**
  ```go
  // 우선순위: 높음
  func (s *EventService) StreamRaceEvents(gameId, playerId string) <-chan *RaceEvent
  func (s *EventService) BroadcastEnvironmentChange(gameId string, change *EnvironmentChange) error
  func (s *EventService) NotifyDifficultyAdjustment(gameId string, newLevel int) error
  ```

- [ ] **협력 액션 처리**
  ```go
  // 우선순위: 높음
  func (s *CoopService) ProcessRaceSpecificCooperation(action *CooperationAction) error
  func (s *CoopService) ValidateInterRaceCooperation(fromRace, toRace string, action *CoopAction) bool
  func (s *CoopService) CalculateCooperationBonus(races []RaceType, action *CoopAction) float64
  ```

#### 📈 성과 측정 및 분석
- [ ] **종족별 성과 추적**
  ```go
  // 우선순위: 중간
  func (s *StatsService) TrackRacePerformance(gameId, playerId string, metrics *RaceMetrics) error
  func (s *StatsService) AnalyzeEnvironmentAdaptation(gameId string) *AdaptationReport
  func (s *StatsService) GenerateDifficultyProgressReport(playerId string) *ProgressReport
  ```

### Phase 4: 사용자 인터페이스 구현 (3-4주)

#### 🎨 종족별 UI 구현
- [ ] **종족 선택 인터페이스**
  - 종족 정보 표시 컴포넌트
  - 종족 간 비교 도구
  - 팀 구성 시뮬레이터
  - 종족별 미리보기 시스템

- [ ] **환경 정보 표시**
  - 실시간 환경 상태 위젯
  - 종족별 영향도 표시
  - 환경 변화 타임라인
  - 적응 가이드 팝업

#### 📱 게임 내 UI 구현
- [ ] **종족 특화 HUD**
  - 종족별 능력 버튼
  - 종족 시너지 표시기
  - 환경 효과 인디케이터
  - 협력 기회 알림

### Phase 5: 테스트 및 밸런싱 (지속적)

#### 🧪 시스템 테스트
- [ ] **종족 밸런스 테스트**
  - 종족별 승률 분석
  - 종족 조합 효과 검증
  - 환경별 종족 성능 테스트
  - 플레이어 선호도 조사

- [ ] **환경 시스템 테스트**
  - 환경 조합 밸런스 검증
  - 트레이드오프 효과 측정
  - 환경 변화 타이밍 최적화
  - 플레이어 적응 시간 분석

- [ ] **난이도 시스템 테스트**
  - 학습 곡선 검증
  - 적응형 난이도 효과 측정
  - 플레이어 만족도 조사
  - 이탈률 분석

---

## 🎯 우선순위 및 마일스톤

### 🚀 MVP (Minimum Viable Product) - 8주
1. 기본 종족 시스템 (4개 종족)
2. 단순 환경 시스템 (3가지 조합)
3. 기본 난이도 시스템 (1-20레벨)
4. 핵심 협력 메커니즘

### 🌟 Beta Version - 16주
1. 전체 종족 시스템 (8개 종족)
2. 완전한 환경 시스템 (모든 조합)
3. 전체 난이도 시스템 (1-100레벨)
4. 고급 협력 및 소셜 기능

### 🏆 Release Version - 24주
1. 밸런싱 완료
2. 성능 최적화
3. 사용자 경험 개선
4. 확장 컨텐츠 준비
