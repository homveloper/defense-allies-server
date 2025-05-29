# Defense Allies 동적 밸런싱 시스템 구현

## 📋 문서 정보
- **작성일**: 2024년
- **버전**: v1.0
- **목적**: 실시간 게임 밸런스 모니터링 및 자동 조정 시스템 구현
- **기반**: [18종족 매트릭스 최적화](18-race-matrix-optimization.md)

## 🎯 동적 밸런싱 목표

### 핵심 원칙
1. **실시간 모니터링**: 게임 중 지속적인 밸런스 상태 추적
2. **자동 조정**: 불균형 감지 시 즉시 환경 변수로 보정
3. **플레이어 경험 보존**: 밸런싱이 게임 재미를 해치지 않도록
4. **예측적 조정**: 미래 불균형 상황 예측 및 사전 대응

## 🔍 실시간 모니터링 시스템

### 밸런스 메트릭 정의
```python
import numpy as np
from dataclasses import dataclass
from typing import Dict, List, Tuple
import time

@dataclass
class BalanceMetrics:
    """게임 밸런스 측정 지표"""
    frobenius_variance: float      # 팀 간 파워 분산
    win_rate_deviation: float      # 승률 편차
    resource_efficiency: float     # 자원 효율성 차이
    cooperation_index: float       # 협력 활용도
    adaptation_speed: float        # 환경 적응 속도
    
    def overall_balance_score(self) -> float:
        """종합 밸런스 점수 (0~1, 1이 완벽한 균형)"""
        weights = [0.3, 0.25, 0.2, 0.15, 0.1]
        metrics = [
            1.0 / (1.0 + self.frobenius_variance),
            1.0 / (1.0 + self.win_rate_deviation),
            1.0 / (1.0 + abs(self.resource_efficiency - 1.0)),
            self.cooperation_index,
            self.adaptation_speed
        ]
        return sum(w * m for w, m in zip(weights, metrics))

class GameStateMonitor:
    """게임 상태 실시간 모니터링"""
    
    def __init__(self, update_interval: float = 30.0):
        self.update_interval = update_interval
        self.balance_history: List[BalanceMetrics] = []
        self.player_matrices: Dict[str, np.ndarray] = {}
        self.environment_state = {
            'time': 'day',
            'weather': 'clear', 
            'terrain': 'plain'
        }
        
    def update_player_state(self, player_id: str, race: str, 
                           towers: List[Dict], resources: Dict):
        """플레이어 상태 업데이트"""
        # 현재 플레이어의 효과적인 파워 매트릭스 계산
        base_matrix = self.get_race_matrix(race)
        tower_bonus = self.calculate_tower_bonus(towers)
        env_modifier = self.get_environment_modifier(race)
        
        effective_matrix = base_matrix * tower_bonus * env_modifier
        self.player_matrices[player_id] = effective_matrix
        
    def calculate_current_balance(self) -> BalanceMetrics:
        """현재 밸런스 상태 계산"""
        if len(self.player_matrices) < 2:
            return BalanceMetrics(0, 0, 1, 1, 1)
            
        matrices = list(self.player_matrices.values())
        
        # 1. 프로베니우스 노름 분산
        norms = [np.linalg.norm(matrix, 'fro') for matrix in matrices]
        frobenius_var = np.var(norms)
        
        # 2. 승률 편차 (시뮬레이션 기반)
        win_rates = self.simulate_win_rates(matrices)
        win_rate_dev = np.std(win_rates)
        
        # 3. 자원 효율성
        resource_eff = self.calculate_resource_efficiency()
        
        # 4. 협력 지수
        coop_index = self.calculate_cooperation_index()
        
        # 5. 적응 속도
        adapt_speed = self.calculate_adaptation_speed()
        
        return BalanceMetrics(
            frobenius_variance=frobenius_var,
            win_rate_deviation=win_rate_dev,
            resource_efficiency=resource_eff,
            cooperation_index=coop_index,
            adaptation_speed=adapt_speed
        )
    
    def simulate_win_rates(self, matrices: List[np.ndarray]) -> List[float]:
        """매트릭스 기반 승률 시뮬레이션"""
        win_rates = []
        
        for i, matrix_i in enumerate(matrices):
            wins = 0
            total_matches = 0
            
            for j, matrix_j in enumerate(matrices):
                if i != j:
                    # 매트릭스 대결 시뮬레이션
                    power_i = np.linalg.norm(matrix_i, 'fro')
                    power_j = np.linalg.norm(matrix_j, 'fro')
                    
                    # 시그모이드 함수로 승률 계산
                    win_prob = 1 / (1 + np.exp(-(power_i - power_j)))
                    wins += win_prob
                    total_matches += 1
            
            win_rates.append(wins / total_matches if total_matches > 0 else 0.5)
        
        return win_rates

class AdaptiveEnvironmentGenerator:
    """적응형 환경 생성기"""
    
    def __init__(self):
        self.environment_effects = self.load_environment_matrices()
        self.balance_threshold = 0.7  # 밸런스 점수 임계값
        
    def generate_balancing_environment(self, 
                                     current_matrices: Dict[str, np.ndarray],
                                     target_balance: float = 0.85) -> Dict[str, str]:
        """밸런싱을 위한 최적 환경 생성"""
        
        best_env = None
        best_score = 0
        
        # 모든 환경 조합 테스트
        for time in ['dawn', 'day', 'dusk', 'night']:
            for weather in ['clear', 'rain', 'storm', 'snow', 'fog']:
                for terrain in ['plain', 'forest', 'mountain', 'desert', 'swamp', 'urban']:
                    
                    # 이 환경에서의 밸런스 점수 계산
                    env_score = self.evaluate_environment_balance(
                        current_matrices, time, weather, terrain
                    )
                    
                    if env_score > best_score:
                        best_score = env_score
                        best_env = {
                            'time': time,
                            'weather': weather, 
                            'terrain': terrain
                        }
        
        return best_env
    
    def evaluate_environment_balance(self, 
                                   matrices: Dict[str, np.ndarray],
                                   time: str, weather: str, terrain: str) -> float:
        """특정 환경에서의 밸런스 점수 평가"""
        
        # 환경 효과 적용
        modified_matrices = {}
        for player_id, matrix in matrices.items():
            race = self.get_player_race(player_id)
            env_modifier = self.get_environment_modifier(race, time, weather, terrain)
            modified_matrices[player_id] = matrix * env_modifier
        
        # 수정된 매트릭스들의 밸런스 점수 계산
        norms = [np.linalg.norm(matrix, 'fro') for matrix in modified_matrices.values()]
        variance = np.var(norms)
        
        # 분산이 낮을수록 좋은 점수
        return 1.0 / (1.0 + variance)

class DynamicBalancer:
    """동적 밸런싱 메인 엔진"""
    
    def __init__(self):
        self.monitor = GameStateMonitor()
        self.env_generator = AdaptiveEnvironmentGenerator()
        self.balance_history: List[Tuple[float, BalanceMetrics]] = []
        self.intervention_cooldown = 120.0  # 2분 쿨다운
        self.last_intervention = 0
        
    def run_balancing_cycle(self):
        """밸런싱 사이클 실행"""
        current_time = time.time()
        
        # 현재 밸런스 상태 측정
        current_balance = self.monitor.calculate_current_balance()
        balance_score = current_balance.overall_balance_score()
        
        # 기록 저장
        self.balance_history.append((current_time, current_balance))
        
        # 밸런스 점수가 임계값 이하이고 쿨다운이 끝났다면
        if (balance_score < 0.7 and 
            current_time - self.last_intervention > self.intervention_cooldown):
            
            self.trigger_balancing_intervention(current_balance)
            self.last_intervention = current_time
    
    def trigger_balancing_intervention(self, current_balance: BalanceMetrics):
        """밸런싱 개입 실행"""
        
        # 1. 환경 변경을 통한 밸런싱
        if current_balance.frobenius_variance > 0.5:
            new_env = self.env_generator.generate_balancing_environment(
                self.monitor.player_matrices
            )
            self.apply_environment_change(new_env)
        
        # 2. 임시 버프/디버프 적용
        if current_balance.win_rate_deviation > 0.3:
            self.apply_temporary_adjustments()
        
        # 3. 특수 이벤트 발생
        if current_balance.cooperation_index < 0.4:
            self.trigger_cooperation_event()
    
    def apply_environment_change(self, new_environment: Dict[str, str]):
        """환경 변경 적용"""
        print(f"🌍 환경 변경: {new_environment}")
        
        # 게임 서버에 환경 변경 명령 전송
        self.send_environment_update(new_environment)
        
        # 플레이어들에게 알림
        self.notify_players_environment_change(new_environment)
    
    def apply_temporary_adjustments(self):
        """임시 조정 적용"""
        
        # 가장 약한 팀에게 임시 버프
        weakest_players = self.identify_weakest_players()
        
        for player_id in weakest_players:
            buff_matrix = np.array([[1.1, 1.05], [1.05, 1.1]])
            self.apply_temporary_buff(player_id, buff_matrix, duration=180)
        
        print(f"⚡ 임시 조정: {len(weakest_players)}명에게 버프 적용")
    
    def trigger_cooperation_event(self):
        """협력 촉진 이벤트 발생"""
        
        cooperation_events = [
            "diplomatic_summit",    # 외교 정상회담
            "trade_festival",      # 무역 축제
            "alliance_bonus",      # 동맹 보너스
            "shared_victory"       # 공동 승리 조건
        ]
        
        selected_event = np.random.choice(cooperation_events)
        self.activate_special_event(selected_event)
        
        print(f"🤝 협력 이벤트 발생: {selected_event}")

class PredictiveBalancer:
    """예측적 밸런싱 시스템"""
    
    def __init__(self):
        self.ml_model = self.load_balance_prediction_model()
        self.prediction_horizon = 300  # 5분 후 예측
        
    def predict_future_balance(self, 
                             current_state: Dict,
                             time_horizon: int = 300) -> BalanceMetrics:
        """미래 밸런스 상태 예측"""
        
        # 현재 상태를 특성 벡터로 변환
        features = self.extract_features(current_state)
        
        # ML 모델로 미래 상태 예측
        predicted_metrics = self.ml_model.predict(features, time_horizon)
        
        return BalanceMetrics(*predicted_metrics)
    
    def recommend_preemptive_actions(self, 
                                   predicted_balance: BalanceMetrics) -> List[str]:
        """예측된 불균형에 대한 사전 대응 추천"""
        
        recommendations = []
        
        if predicted_balance.frobenius_variance > 0.6:
            recommendations.append("환경 변화 준비")
        
        if predicted_balance.win_rate_deviation > 0.4:
            recommendations.append("약한 팀 지원 이벤트 준비")
        
        if predicted_balance.cooperation_index < 0.3:
            recommendations.append("협력 촉진 이벤트 예약")
        
        return recommendations

# 메인 실행 루프
class BalancingOrchestrator:
    """밸런싱 시스템 총괄 관리"""
    
    def __init__(self):
        self.dynamic_balancer = DynamicBalancer()
        self.predictive_balancer = PredictiveBalancer()
        self.running = False
        
    async def start_balancing_system(self):
        """밸런싱 시스템 시작"""
        self.running = True
        
        while self.running:
            try:
                # 현재 밸런스 체크 및 조정
                self.dynamic_balancer.run_balancing_cycle()
                
                # 미래 상태 예측 및 사전 대응
                current_state = self.get_current_game_state()
                predicted_balance = self.predictive_balancer.predict_future_balance(
                    current_state
                )
                
                # 예측 기반 권장사항 생성
                recommendations = self.predictive_balancer.recommend_preemptive_actions(
                    predicted_balance
                )
                
                if recommendations:
                    print(f"🔮 예측 권장사항: {recommendations}")
                
                # 30초 대기
                await asyncio.sleep(30)
                
            except Exception as e:
                print(f"❌ 밸런싱 오류: {e}")
                await asyncio.sleep(60)  # 오류 시 1분 대기
    
    def stop_balancing_system(self):
        """밸런싱 시스템 중지"""
        self.running = False
        print("🛑 동적 밸런싱 시스템 중지")

# 사용 예시
if __name__ == "__main__":
    import asyncio
    
    orchestrator = BalancingOrchestrator()
    
    try:
        asyncio.run(orchestrator.start_balancing_system())
    except KeyboardInterrupt:
        orchestrator.stop_balancing_system()
```

---

**다음 단계**: JSON Schema 기반 데이터 구조 구현 및 Redis 연동
