import { GameEntity } from './GameEntity';
import { Position } from '../value-objects/Position';
import { Health } from '../value-objects/Health';
import { Velocity } from '../value-objects/Velocity';

export class Projectile extends GameEntity {
  private readonly _ownerId: string;
  private readonly _ownerType: 'player' | 'enemy' | 'ally';
  private readonly _maxLifetime: number = 3.0; // 3초 후 자동 소멸
  private _lifetime: number = 0;
  
  // 타겟 예측을 위한 원래 타겟 정보
  public originalTarget?: GameEntity;

  constructor(
    id: string,
    position: Position,
    velocity: Velocity,
    damage: number,
    ownerId: string,
    ownerType: 'player' | 'enemy' | 'ally'
  ) {
    const color = ownerType === 'enemy' ? '#DC2626' : '#3B82F6';
    
    super(
      id,
      position,
      Health.full(1), // 투사체는 체력 1
      velocity,
      damage,
      velocity.magnitude, // 속도 = 투사체 속력
      4, // 작은 크기
      color
    );

    this._ownerId = ownerId;
    this._ownerType = ownerType;
  }

  static create(
    ownerId: string,
    ownerType: 'player' | 'enemy' | 'ally',
    startPosition: Position,
    targetPosition: Position,
    damage: number,
    speed: number = 400
  ): Projectile {
    const direction = targetPosition.subtract(startPosition);
    const velocity = Velocity.fromDirection(direction, speed);
    const id = `proj_${ownerType}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    
    return new Projectile(id, startPosition, velocity, damage, ownerId, ownerType);
  }

  getType(): string {
    return 'projectile';
  }

  get ownerId(): string {
    return this._ownerId;
  }

  get ownerType(): 'player' | 'enemy' | 'ally' {
    return this._ownerType;
  }

  get lifetime(): number {
    return this._lifetime;
  }

  get maxLifetime(): number {
    return this._maxLifetime;
  }

  get isExpired(): boolean {
    return this._lifetime >= this._maxLifetime;
  }

  // 투사체 업데이트
  update(deltaTime: number): void {
    super.update(deltaTime); // 기본 이동
    this._lifetime += deltaTime;
  }

  // 화면 경계 확인
  isOutOfBounds(screenWidth: number, screenHeight: number, margin: number = 100): boolean {
    return (
      this._position.x < -margin ||
      this._position.x > screenWidth + margin ||
      this._position.y < -margin ||
      this._position.y > screenHeight + margin
    );
  }

  // 충돌 처리
  onHit(target: GameEntity): { shouldDestroy: boolean; damageDealt: number } {
    // 투사체는 충돌 시 파괴됨
    this.takeDamage(1);
    
    return {
      shouldDestroy: true,
      damageDealt: this._damage
    };
  }

  // 대상과 충돌 가능한지 확인
  canHit(target: GameEntity): boolean {
    // 자신의 소유자와는 충돌하지 않음
    if (target.id === this._ownerId) return false;

    // 같은 팀끼리는 충돌하지 않음
    if (this._ownerType === 'player' && target.getType() === 'ally') return false;
    if (this._ownerType === 'ally' && target.getType() === 'player') return false;
    if (this._ownerType === 'enemy' && target.getType() === 'enemy') return false;

    return target.isAlive;
  }
}