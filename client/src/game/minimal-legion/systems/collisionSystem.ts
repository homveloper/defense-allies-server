import { Entity, Vector2D } from '../types/minimalLegion';

export interface CollisionResult {
  projectileHits: { projectileId: string; targetId: string; damage: number }[];
  enemyConversions: string[];
  playerDamage: number;
  allyDamage: { allyId: string; damage: number }[];
}

export class CollisionSystem {
  constructor() {}
  
  checkCollisions(
    player: Entity,
    allies: Entity[],
    enemies: Entity[],
    projectiles: Entity[]
  ): CollisionResult {
    const result: CollisionResult = {
      projectileHits: [],
      enemyConversions: [],
      playerDamage: 0,
      allyDamage: []
    };
    
    // Check projectile vs entity collisions
    for (const projectile of projectiles) {
      const targets = projectile.owner?.startsWith('enemy') 
        ? [player, ...allies] 
        : enemies;
      
      for (const target of targets) {
        if (this.checkCircleCollision(projectile, target)) {
          result.projectileHits.push({
            projectileId: projectile.id,
            targetId: target.id,
            damage: projectile.damage
          });
          break; // Projectile can only hit one target
        }
      }
    }
    
    // Check enemy vs player/ally collisions
    for (const enemy of enemies) {
      // Check player collision
      if (this.checkCircleCollision(enemy, player)) {
        result.playerDamage += enemy.damage * 0.016; // Damage per frame at 60fps
      }
      
      // Check ally collisions
      for (const ally of allies) {
        if (this.checkCircleCollision(enemy, ally)) {
          result.allyDamage.push({
            allyId: ally.id,
            damage: enemy.damage * 0.016
          });
        }
      }
    }
    
    return result;
  }
  
  private checkCircleCollision(entity1: Entity, entity2: Entity): boolean {
    const distance = this.getDistance(entity1.position, entity2.position);
    const minDistance = (entity1.size + entity2.size) / 2;
    return distance < minDistance;
  }
  
  private getDistance(pos1: Vector2D, pos2: Vector2D): number {
    const dx = pos1.x - pos2.x;
    const dy = pos1.y - pos2.y;
    return Math.sqrt(dx * dx + dy * dy);
  }
}