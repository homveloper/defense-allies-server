import { Position } from '../value-objects/Position';
import { Health } from '../value-objects/Health';
import { Velocity } from '../value-objects/Velocity';

export abstract class GameEntity {
  constructor(
    protected readonly _id: string,
    protected _position: Position,
    protected _health: Health,
    protected _velocity: Velocity,
    protected readonly _damage: number,
    protected readonly _speed: number,
    protected readonly _size: number,
    protected readonly _color: string
  ) {}

  // Getters
  get id(): string {
    return this._id;
  }

  get position(): Position {
    return this._position;
  }

  get health(): Health {
    return this._health;
  }

  get velocity(): Velocity {
    return this._velocity;
  }

  get damage(): number {
    return this._damage;
  }

  get speed(): number {
    return this._speed;
  }

  get size(): number {
    return this._size;
  }

  get color(): string {
    return this._color;
  }

  get isAlive(): boolean {
    return this._health.isAlive;
  }

  get isDead(): boolean {
    return this._health.isDead;
  }

  // Actions
  moveTo(newPosition: Position): void {
    this._position = newPosition;
  }

  setVelocity(newVelocity: Velocity): void {
    this._velocity = newVelocity;
  }

  takeDamage(damage: number): void {
    this._health = this._health.takeDamage(damage);
  }

  heal(amount: number): void {
    this._health = this._health.heal(amount);
  }

  update(deltaTime: number): void {
    // 기본 이동 로직
    const deltaPosition = this._velocity.toPosition().multiply(deltaTime);
    this._position = this._position.add(deltaPosition);
  }

  // 충돌 감지
  isCollidingWith(other: GameEntity): boolean {
    const distance = this._position.distanceTo(other.position);
    const collisionDistance = (this._size + other.size) / 2;
    return distance <= collisionDistance;
  }

  // 범위 내 확인
  isInRangeOf(other: GameEntity, range: number): boolean {
    return this._position.distanceTo(other.position) <= range;
  }

  abstract getType(): string;
}