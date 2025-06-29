export class Health {
  constructor(
    private current: number,
    private readonly maximum: number
  ) {
    if (current < 0) throw new Error('Current health cannot be negative');
    if (maximum <= 0) throw new Error('Maximum health must be positive');
    if (current > maximum) throw new Error('Current health cannot exceed maximum');
  }

  static create(current: number, maximum: number): Health {
    return new Health(current, maximum);
  }

  static full(maximum: number): Health {
    return new Health(maximum, maximum);
  }

  get currentValue(): number {
    return this.current;
  }

  get maximumValue(): number {
    return this.maximum;
  }

  get percentage(): number {
    return this.current / this.maximum;
  }

  get isAlive(): boolean {
    return this.current > 0;
  }

  get isDead(): boolean {
    return this.current <= 0;
  }

  takeDamage(damage: number): Health {
    if (damage < 0) throw new Error('Damage cannot be negative');
    const newCurrent = Math.max(0, this.current - damage);
    return new Health(newCurrent, this.maximum);
  }

  heal(amount: number): Health {
    if (amount < 0) throw new Error('Heal amount cannot be negative');
    const newCurrent = Math.min(this.maximum, this.current + amount);
    return new Health(newCurrent, this.maximum);
  }

  equals(other: Health): boolean {
    return this.current === other.current && this.maximum === other.maximum;
  }

  toString(): string {
    return `${this.current}/${this.maximum} (${(this.percentage * 100).toFixed(1)}%)`;
  }
}