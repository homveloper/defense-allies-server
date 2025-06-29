import { Entity, Vector2D } from '@/types/minimalLegion';

export class MovementSystem {
  constructor() {}
  
  update(
    deltaTime: number,
    player: Entity,
    allies: Entity[],
    enemies: Entity[],
    projectiles: Entity[]
  ): {
    updatedAllies: Entity[];
    updatedEnemies: Entity[];
    updatedProjectiles: Entity[];
  } {
    // Update ally positions (follow player)
    const updatedAllies = allies.map(ally => this.updateAllyMovement(ally, player, allies, deltaTime));
    
    // Update enemy positions (move towards player)
    const updatedEnemies = enemies.map(enemy => this.updateEnemyMovement(enemy, player, deltaTime));
    
    // Update projectile positions
    const updatedProjectiles = projectiles
      .map(projectile => this.updateProjectileMovement(projectile, deltaTime))
      .filter(projectile => this.isProjectileInBounds(projectile));
    
    return {
      updatedAllies,
      updatedEnemies,
      updatedProjectiles
    };
  }
  
  private updateAllyMovement(
    ally: Entity,
    player: Entity,
    allAllies: Entity[],
    deltaTime: number
  ): Entity {
    const desiredDistance = 40; // Distance to maintain from player
    const separationDistance = 30; // Distance to maintain from other allies
    
    // Calculate desired position around player
    const angle = Math.atan2(
      ally.position.y - player.position.y,
      ally.position.x - player.position.x
    );
    
    const desiredPosition = {
      x: player.position.x + Math.cos(angle) * desiredDistance,
      y: player.position.y + Math.sin(angle) * desiredDistance
    };
    
    // Calculate movement vector towards desired position
    let moveVector = {
      x: desiredPosition.x - ally.position.x,
      y: desiredPosition.y - ally.position.y
    };
    
    // Add separation from other allies
    for (const otherAlly of allAllies) {
      if (otherAlly.id === ally.id) continue;
      
      const distance = this.getDistance(ally.position, otherAlly.position);
      if (distance < separationDistance && distance > 0) {
        const separationForce = {
          x: (ally.position.x - otherAlly.position.x) / distance,
          y: (ally.position.y - otherAlly.position.y) / distance
        };
        moveVector.x += separationForce.x * 20;
        moveVector.y += separationForce.y * 20;
      }
    }
    
    // Normalize and apply speed
    const normalized = this.normalize(moveVector);
    const speed = ally.speed * 60; // Convert to pixels per second
    
    return {
      ...ally,
      position: {
        x: ally.position.x + normalized.x * speed * deltaTime,
        y: ally.position.y + normalized.y * speed * deltaTime
      }
    };
  }
  
  private updateEnemyMovement(
    enemy: Entity,
    player: Entity,
    deltaTime: number
  ): Entity {
    // Simple movement towards player
    const direction = this.normalize({
      x: player.position.x - enemy.position.x,
      y: player.position.y - enemy.position.y
    });
    
    const speed = enemy.speed * 60; // Convert to pixels per second
    
    return {
      ...enemy,
      position: {
        x: enemy.position.x + direction.x * speed * deltaTime,
        y: enemy.position.y + direction.y * speed * deltaTime
      }
    };
  }
  
  private updateProjectileMovement(
    projectile: Entity,
    deltaTime: number
  ): Entity {
    return {
      ...projectile,
      position: {
        x: projectile.position.x + projectile.velocity.x * deltaTime,
        y: projectile.position.y + projectile.velocity.y * deltaTime
      }
    };
  }
  
  private isProjectileInBounds(projectile: Entity): boolean {
    const margin = 100;
    return (
      projectile.position.x >= -margin &&
      projectile.position.x <= 1200 + margin &&
      projectile.position.y >= -margin &&
      projectile.position.y <= 800 + margin
    );
  }
  
  private getDistance(pos1: Vector2D, pos2: Vector2D): number {
    const dx = pos1.x - pos2.x;
    const dy = pos1.y - pos2.y;
    return Math.sqrt(dx * dx + dy * dy);
  }
  
  private normalize(vector: Vector2D): Vector2D {
    const length = Math.sqrt(vector.x * vector.x + vector.y * vector.y);
    if (length === 0) return { x: 0, y: 0 };
    return { x: vector.x / length, y: vector.y / length };
  }
}