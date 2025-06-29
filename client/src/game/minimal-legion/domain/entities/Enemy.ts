import { GameEntity } from './GameEntity';
import { Position } from '../value-objects/Position';
import { Health } from '../value-objects/Health';
import { Velocity } from '../value-objects/Velocity';

export type EnemyType = 'normal' | 'fast' | 'tank' | 'ranged';

export interface EnemyConfig {
  health: number;
  damage: number;
  speed: number;
  size: number;
  color: string;
  experience: number;
  attackRange?: number;
  attackCooldown?: number;
}

export class Enemy extends GameEntity {
  private readonly _enemyType: EnemyType;
  private readonly _experience: number;
  private readonly _attackRange: number;
  private readonly _attackCooldown: number;
  private _lastAttackTime: number = 0;
  private _target: Position | null = null;
  
  // 충돌 회피 관련 변수
  private readonly _separationRadius: number;
  private readonly _separationWeight: number = 1.5;

  constructor(
    id: string,
    position: Position,
    enemyType: EnemyType,
    config: EnemyConfig
  ) {
    super(
      id,
      position,
      Health.full(config.health),
      Velocity.zero(),
      config.damage,
      config.speed,
      config.size,
      config.color
    );

    this._enemyType = enemyType;
    this._experience = config.experience;
    this._attackRange = config.attackRange || 250;
    this._attackCooldown = config.attackCooldown || 1.0;
  }

  static createNormal(id: string, position: Position): Enemy {
    return new Enemy(id, position, 'normal', {
      health: 80, // 150 -> 80으로 낮춤
      damage: 5,
      speed: 3,
      size: 18,
      color: '#EF4444',
      experience: 10
    });
  }

  static createFast(id: string, position: Position): Enemy {
    return new Enemy(id, position, 'fast', {
      health: 60, // 100 -> 60으로 낮춤
      damage: 3,
      speed: 8,
      size: 16,
      color: '#F59E0B',
      experience: 15
    });
  }

  static createTank(id: string, position: Position): Enemy {
    return new Enemy(id, position, 'tank', {
      health: 150, // 300 -> 150으로 낮춤
      damage: 10,
      speed: 2,
      size: 24,
      color: '#991B1B',
      experience: 25
    });
  }

  static createRanged(id: string, position: Position): Enemy {
    return new Enemy(id, position, 'ranged', {
      health: 70, // 120 -> 70으로 낮춤
      damage: 8,
      speed: 3,
      size: 18,
      color: '#DC2626',
      experience: 20,
      attackRange: 300
    });
  }

  getType(): string {
    return 'enemy';
  }

  get enemyType(): EnemyType {
    return this._enemyType;
  }

  get experience(): number {
    return this._experience;
  }

  get attackRange(): number {
    return this._attackRange;
  }

  get attackCooldown(): number {
    return this._attackCooldown;
  }

  // 타겟 설정 (플레이어 위치)
  setTarget(target: Position): void {
    this._target = target;
  }

  // 플레이어를 향해 이동
  moveTowardsTarget(): void {
    if (!this._target) return;

    const direction = this._target.subtract(this._position);
    const velocity = Velocity.fromDirection(direction, this._speed * 60);
    this.setVelocity(velocity);
  }

  // 공격 가능 여부
  canAttack(currentTime: number, target: GameEntity): boolean {
    if (currentTime - this._lastAttackTime < this._attackCooldown) return false;
    return this.isInRangeOf(target, this._attackRange);
  }

  // 공격 실행
  attack(currentTime: number): void {
    this._lastAttackTime = currentTime;
  }

  // 웨이브에 따른 스탯 조정
  scaleForWave(wave: number): Enemy {
    const multiplier = 1 + (wave - 1) * 0.1;
    const scaledHealth = Math.floor(this._health.maximumValue * multiplier);
    const scaledDamage = Math.floor(this._damage * multiplier);

    // 새로운 Enemy 인스턴스 생성 (불변성)
    const scaledConfig: EnemyConfig = {
      health: scaledHealth,
      damage: scaledDamage,
      speed: this._speed,
      size: this._size,
      color: this._color,
      experience: this._experience,
      attackRange: this._attackRange,
      attackCooldown: this._attackCooldown
    };

    return new Enemy(this._id, this._position, this._enemyType, scaledConfig);
  }

  // 적 제거시 보상 계산
  calculateReward(): { experience: number; score: number } {
    return {
      experience: this._experience,
      score: this._experience * 10
    };
  }
}