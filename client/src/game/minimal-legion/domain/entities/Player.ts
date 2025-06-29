import { GameEntity } from './GameEntity';
import { Position } from '../value-objects/Position';
import { Health } from '../value-objects/Health';
import { Velocity } from '../value-objects/Velocity';

export class Player extends GameEntity {
  private _lastAttackTime: number = 0;
  private readonly _attackCooldown: number = 0.2; // 초
  private readonly _attackRange: number = 400; // 픽셀

  constructor(
    id: string = 'player',
    position: Position = Position.zero(),
    health: Health = Health.full(100),
    damage: number = 15,
    speed: number = 5
  ) {
    super(
      id,
      position,
      health,
      Velocity.zero(),
      damage,
      speed,
      20, // size
      '#3B82F6' // blue color
    );
  }

  static create(): Player {
    return new Player();
  }

  getType(): string {
    return 'player';
  }

  get attackRange(): number {
    return this._attackRange;
  }

  get attackCooldown(): number {
    return this._attackCooldown;
  }

  // 플레이어 이동 (WASD 입력)
  moveInDirection(direction: Position): void {
    const normalizedDirection = direction.normalize();
    const velocity = Velocity.fromDirection(normalizedDirection, this._speed * 60); // pixels per second
    this.setVelocity(velocity);
  }

  stopMoving(): void {
    this.setVelocity(Velocity.zero());
  }

  // 공격 가능 여부 확인
  canAttack(currentTime: number): boolean {
    return currentTime - this._lastAttackTime >= this._attackCooldown;
  }

  // 공격 실행
  attack(currentTime: number): void {
    this._lastAttackTime = currentTime;
  }

  // 이전 타겟 추적을 위한 변수
  private _lastTarget: GameEntity | null = null;
  private _targetSwitchCooldown: number = 0.5; // 0.5초간 같은 타겟 유지
  private _lastTargetSwitchTime: number = 0;

  // 가장 가까운 적 찾기 (개선된 버전)
  findNearestEnemy(enemies: GameEntity[]): GameEntity | null {
    const currentTime = Date.now() / 1000;
    const aliveEnemies = enemies.filter(enemy => enemy.isAlive);
    
    if (aliveEnemies.length === 0) {
      this._lastTarget = null;
      return null;
    }

    // 이전 타겟이 여전히 유효한지 확인
    if (this._lastTarget && 
        this._lastTarget.isAlive && 
        aliveEnemies.includes(this._lastTarget) &&
        currentTime - this._lastTargetSwitchTime < this._targetSwitchCooldown) {
      
      const distanceToLastTarget = this._position.distanceTo(this._lastTarget.position);
      if (distanceToLastTarget <= this._attackRange) {
        return this._lastTarget; // 기존 타겟 유지
      }
    }

    // 새로운 최적 타겟 찾기
    const bestTarget = this.findOptimalTarget(aliveEnemies);
    
    if (bestTarget && bestTarget !== this._lastTarget) {
      this._lastTarget = bestTarget;
      this._lastTargetSwitchTime = currentTime;
    }

    return bestTarget;
  }

  // 최적 타겟 선택 알고리즘
  private findOptimalTarget(enemies: GameEntity[]): GameEntity | null {
    if (enemies.length === 0) return null;

    let bestTarget: GameEntity | null = null;
    let bestScore = -1;

    for (const enemy of enemies) {
      const distance = this._position.distanceTo(enemy.position);
      
      // 사정거리 밖의 적은 제외
      if (distance > this._attackRange) continue;

      // 타겟 점수 계산 (낮을수록 우선순위 높음)
      const score = this.calculateTargetScore(enemy, distance);
      
      if (score > bestScore) {
        bestScore = score;
        bestTarget = enemy;
      }
    }

    return bestTarget;
  }

  // 타겟 점수 계산 (거리, 체력, 위협도 종합)
  private calculateTargetScore(enemy: GameEntity, distance: number): number {
    // 기본 점수 (거리가 가까울수록 높은 점수)
    const distanceScore = (this._attackRange - distance) / this._attackRange;
    
    // 체력 점수 (체력이 낮을수록 높은 점수 - 마무리 우선)
    const healthPercent = enemy.health.currentValue / enemy.health.maximumValue;
    const healthScore = 1 - healthPercent;
    
    // 속도 점수 (느린 적 우선 - 명중률 향상)
    const speedScore = Math.max(0, 1 - (enemy.speed / 10));
    
    // 종합 점수 (가중평균)
    return (distanceScore * 0.4) + (healthScore * 0.4) + (speedScore * 0.2);
  }

  // 타겟 예측 위치 계산 (이동 중인 적 대상)
  predictTargetPosition(target: GameEntity, projectileSpeed: number = 300): Position {
    if (!target.velocity || target.velocity.magnitude === 0) {
      return target.position; // 정지 중이면 현재 위치 반환
    }

    // 투사체가 타겟에 도달하는 시간 계산
    const distance = this._position.distanceTo(target.position);
    const timeToTarget = distance / projectileSpeed;
    
    // 예상 위치 계산
    const predictedX = target.position.x + (target.velocity.x * timeToTarget);
    const predictedY = target.position.y + (target.velocity.y * timeToTarget);
    
    return new Position(predictedX, predictedY);
  }

  // 플레이어 상태 리셋
  reset(): void {
    this._position = Position.zero();
    this._health = Health.full(100);
    this._velocity = Velocity.zero();
    this._lastAttackTime = 0;
    this._lastTarget = null;
    this._lastTargetSwitchTime = 0;
  }

  // 플레이어 업그레이드 (추후 확장용)
  upgrade(type: 'damage' | 'speed' | 'health', amount: number): void {
    switch (type) {
      case 'health':
        this._health = this._health.heal(amount);
        break;
      // damage, speed는 readonly이므로 별도 처리 필요
    }
  }
}