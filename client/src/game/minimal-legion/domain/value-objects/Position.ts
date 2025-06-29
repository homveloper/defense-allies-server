export class Position {
  constructor(
    public readonly x: number,
    public readonly y: number
  ) {}

  static create(x: number, y: number): Position {
    return new Position(x, y);
  }

  static zero(): Position {
    return new Position(0, 0);
  }

  distanceTo(other: Position): number {
    const dx = this.x - other.x;
    const dy = this.y - other.y;
    return Math.sqrt(dx * dx + dy * dy);
  }

  add(other: Position): Position {
    return new Position(this.x + other.x, this.y + other.y);
  }

  subtract(other: Position): Position {
    return new Position(this.x - other.x, this.y - other.y);
  }

  multiply(scalar: number): Position {
    return new Position(this.x * scalar, this.y * scalar);
  }

  normalize(): Position {
    const length = Math.sqrt(this.x * this.x + this.y * this.y);
    if (length === 0) return Position.zero();
    return new Position(this.x / length, this.y / length);
  }

  equals(other: Position): boolean {
    return this.x === other.x && this.y === other.y;
  }

  toString(): string {
    return `(${this.x}, ${this.y})`;
  }
}