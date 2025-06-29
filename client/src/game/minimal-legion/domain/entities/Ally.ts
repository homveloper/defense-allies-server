import { GameEntity } from './GameEntity';
import { Position } from '../value-objects/Position';
import { Health } from '../value-objects/Health';
import { Velocity } from '../value-objects/Velocity';
import { Enemy } from './Enemy';

export class Ally extends GameEntity {
  private readonly _attackRange: number = 150;
  private readonly _attackCooldown: number = 0.4;
  private _lastAttackTime: number = 0;
  private readonly _followDistance: number = 40;
  private readonly _separationDistance: number = 30;

  // 타겟팅 시스템 변수
  private _lastTarget: GameEntity | null = null;
  private _targetSwitchCooldown: number = 0.3; // 아군은 더 빠르게 타겟 전환
  private _lastTargetSwitchTime: number = 0;

  // 군단 시스템: 원본 적의 데이터 (외형 유지용)
  private _originalEnemyData?: {
    size: number;
    color: string;
    enemyType: string;
  };

  constructor(
    id: string,
    position: Position,
    health: Health = Health.full(80),
    damage: number = 12,
    speed: number = 4
  ) {
    super(
      id,
      position,
      health,
      Velocity.zero(),
      damage,
      speed,
      16, // size
      '#10B981' // green color
    );
  }

  static createFromEnemy(enemy: Enemy): Ally {
    const allyId = `ally_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    // 군단 시스템: 적의 60-70% 스펙으로 아군 생성
    const statReduction = 0.65; // 65% 스펙 유지
    
    const allyHealth = Health.full(Math.floor(enemy.health.maximumValue * statReduction));
    const allyDamage = Math.floor(enemy.damage * statReduction);
    const allySpeed = Math.max(2, Math.floor(enemy.speed * statReduction));

    // 적과 동일한 크기와 색상 유지 (군단의 정체성)
    const ally = new Ally(
      allyId, 
      enemy.position, 
      allyHealth, 
      allyDamage, 
      allySpeed
    );

    // 원본 적의 정보를 아군에 저장 (외형 유지용)
    ally.setOriginalEnemyData({
      size: enemy.size,
      color: enemy.color,
      enemyType: enemy.enemyType
    });

    return ally;
  }

  static create(id: string, position: Position): Ally {
    return new Ally(id, position);
  }

  getType(): string {
    return 'ally';
  }

  get attackRange(): number {
    return this._attackRange;
  }

  get attackCooldown(): number {
    return this._attackCooldown;
  }

  get followDistance(): number {
    return this._followDistance;
  }

  get separationDistance(): number {
    return this._separationDistance;
  }

  // 군단 시스템 관련 메서드들
  setOriginalEnemyData(data: { size: number; color: string; enemyType: string }): void {
    this._originalEnemyData = data;
  }

  get originalEnemyType(): string | undefined {
    return this._originalEnemyData?.enemyType;
  }

  // 외형 관련 오버라이드 (적의 외형 유지)
  get size(): number {
    return this._originalEnemyData?.size || this._size;
  }

  get color(): string {
    // 아군은 약간 다른 색조로 구분 (원본 색상을 녹색 계열로 변환)
    if (this._originalEnemyData?.color) {
      return this.convertToAllyColor(this._originalEnemyData.color);
    }
    return this._color;
  }

  private convertToAllyColor(enemyColor: string): string {
    // 적의 색상을 아군 색상으로 변환 (녹색 계열)
    const colorMap: Record<string, string> = {
      '#EF4444': '#10B981', // 빨간색 -> 에메랄드
      '#F59E0B': '#059669', // 주황색 -> 짙은 에메랄드  
      '#991B1B': '#047857', // 짙은 빨강 -> 짙은 녹색
      '#DC2626': '#065F46'  // 빨강 -> 매우 짙은 녹색
    };
    
    return colorMap[enemyColor] || '#10B981'; // 기본값은 에메랄드
  }

  // 플레이어 따라다니기
  followPlayer(playerPosition: Position, allAllies: Ally[]): void {
    // 플레이어 주변의 원하는 위치 계산
    const angle = Math.atan2(
      this._position.y - playerPosition.y,
      this._position.x - playerPosition.x
    );

    const desiredPosition = new Position(
      playerPosition.x + Math.cos(angle) * this._followDistance,
      playerPosition.y + Math.sin(angle) * this._followDistance
    );

    // 다른 아군들과의 분리 계산
    let separationForce = Position.zero();
    for (const otherAlly of allAllies) {
      if (otherAlly.id === this._id) continue;

      const distance = this._position.distanceTo(otherAlly.position);
      if (distance < this._separationDistance && distance > 0) {
        const separationDirection = this._position.subtract(otherAlly.position).normalize();
        separationForce = separationForce.add(separationDirection.multiply(20));
      }
    }

    // 최종 이동 방향 계산
    const moveDirection = desiredPosition.subtract(this._position).add(separationForce);
    const velocity = Velocity.fromDirection(moveDirection, this._speed * 60);
    this.setVelocity(velocity);
  }

  // 가장 가까운 적 찾기 (개선된 버전)
  findNearestEnemy(enemies: GameEntity[]): GameEntity | null {
    const currentTime = Date.now() / 1000;
    const aliveEnemies = enemies.filter(enemy => enemy.isAlive);
    
    if (aliveEnemies.length === 0) {
      this._lastTarget = null;
      return null;
    }

    // 이전 타겟이 여전히 유효한지 확인 (아군은 더 빠른 타겟 전환)
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

  // 아군 전용 최적 타겟 선택 (플레이어와 약간 다른 우선순위)
  private findOptimalTarget(enemies: GameEntity[]): GameEntity | null {
    if (enemies.length === 0) return null;

    let bestTarget: GameEntity | null = null;
    let bestScore = -1;

    for (const enemy of enemies) {
      const distance = this._position.distanceTo(enemy.position);
      
      // 사정거리 밖의 적은 제외
      if (distance > this._attackRange) continue;

      // 아군 전용 타겟 점수 계산
      const score = this.calculateTargetScore(enemy, distance);
      
      if (score > bestScore) {
        bestScore = score;
        bestTarget = enemy;
      }
    }

    return bestTarget;
  }

  // 아군 전용 타겟 점수 계산 (거리 중시, 빠른 반응)
  private calculateTargetScore(enemy: GameEntity, distance: number): number {
    // 거리 점수 (가까운 적 우선 - 아군은 근접 전투 선호)
    const distanceScore = (this._attackRange - distance) / this._attackRange;
    
    // 체력 점수 (체력이 낮은 적 우선 - 마무리 담당)
    const healthPercent = enemy.health.currentValue / enemy.health.maximumValue;
    const healthScore = 1 - healthPercent;
    
    // 위협도 점수 (높은 공격력의 적 우선)
    const threatScore = Math.min(1, enemy.damage / 20);
    
    // 아군 전용 가중평균 (거리를 더 중시)
    return (distanceScore * 0.5) + (healthScore * 0.3) + (threatScore * 0.2);
  }

  // 공격 가능 여부
  canAttack(currentTime: number): boolean {
    return currentTime - this._lastAttackTime >= this._attackCooldown;
  }

  // 공격 실행
  attack(currentTime: number): void {
    this._lastAttackTime = currentTime;
  }

  // 아군 강화 (변환 시 보너스) - 적 타입별 차별화
  applyConversionBonus(): void {
    if (!this._originalEnemyData) {
      // 일반 아군인 경우 기본 보너스
      this._health = this._health.heal(20);
      return;
    }

    // 원본 적의 타입에 따른 특별 보너스
    switch (this._originalEnemyData.enemyType) {
      case 'normal':
        // 일반 적: 체력 +15, 균형 잡힌 아군
        this._health = this._health.heal(15);
        break;
        
      case 'fast':
        // 빠른 적: 체력 +10, 속도 보정 (너무 느려지지 않도록)
        this._health = this._health.heal(10);
        console.log(`Fast enemy converted to agile ally: ${this.id}`);
        break;
        
      case 'tank':
        // 탱크 적: 체력 +30, 방어 중심 아군
        this._health = this._health.heal(30);
        console.log(`Tank enemy converted to defensive ally: ${this.id}`);
        break;
        
      case 'ranged':
        // 원거리 적: 체력 +12, 원거리 공격 유지
        this._health = this._health.heal(12);
        console.log(`Ranged enemy converted to support ally: ${this.id}`);
        break;
        
      default:
        this._health = this._health.heal(15);
    }
    
    console.log(`Enemy ${this._originalEnemyData.enemyType} converted to ally with ${this._health.currentValue}/${this._health.maximumValue} HP`);
  }

  // 경험치 기반 레벨업 (추후 확장)
  levelUp(): void {
    // 체력 증가, 공격력 증가 등
    this._health = this._health.heal(10);
  }
}