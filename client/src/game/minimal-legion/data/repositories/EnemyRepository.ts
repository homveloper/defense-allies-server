import { Entity } from '../../types/minimalLegion';

export interface IEnemyRepository {
  add(enemy: Entity): void;
  remove(id: string): void;
  getById(id: string): Entity | null;
  getAll(): Entity[];
  getInRange(center: { x: number; y: number }, radius: number): Entity[];
  update(id: string, updates: Partial<Entity>): void;
  clear(): void;
  count(): number;
}

export class EnemyRepository implements IEnemyRepository {
  private enemies = new Map<string, Entity>();

  add(enemy: Entity): void {
    this.enemies.set(enemy.id, { ...enemy });
    console.log(`Enemy added to repository: ${enemy.id} at (${enemy.position.x}, ${enemy.position.y})`);
  }

  remove(id: string): void {
    const deleted = this.enemies.delete(id);
    if (deleted) {
      console.log(`Enemy removed from repository: ${id}`);
    }
  }

  getById(id: string): Entity | null {
    return this.enemies.get(id) || null;
  }

  getAll(): Entity[] {
    return Array.from(this.enemies.values());
  }

  getInRange(center: { x: number; y: number }, radius: number): Entity[] {
    return this.getAll().filter(enemy => {
      const distance = Math.sqrt(
        Math.pow(enemy.position.x - center.x, 2) + 
        Math.pow(enemy.position.y - center.y, 2)
      );
      return distance <= radius;
    });
  }

  update(id: string, updates: Partial<Entity>): void {
    const enemy = this.enemies.get(id);
    if (enemy) {
      this.enemies.set(id, { ...enemy, ...updates });
    }
  }

  clear(): void {
    this.enemies.clear();
    console.log('Enemy repository cleared');
  }

  count(): number {
    return this.enemies.size;
  }
}