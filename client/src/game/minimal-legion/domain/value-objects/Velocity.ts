import { Position } from './Position';

export class Velocity {
  constructor(
    public readonly x: number,
    public readonly y: number
  ) {}

  static create(x: number, y: number): Velocity {
    return new Velocity(x, y);
  }

  static zero(): Velocity {
    return new Velocity(0, 0);
  }

  static fromDirection(direction: Position, speed: number): Velocity {
    const normalized = direction.normalize();
    return new Velocity(normalized.x * speed, normalized.y * speed);
  }

  get magnitude(): number {
    return Math.sqrt(this.x * this.x + this.y * this.y);
  }

  toPosition(): Position {
    return new Position(this.x, this.y);
  }

  multiply(scalar: number): Velocity {
    return new Velocity(this.x * scalar, this.y * scalar);
  }

  add(other: Velocity): Velocity {
    return new Velocity(this.x + other.x, this.y + other.y);
  }

  equals(other: Velocity): boolean {
    return this.x === other.x && this.y === other.y;
  }

  toString(): string {
    return `Velocity(${this.x}, ${this.y})`;
  }
}