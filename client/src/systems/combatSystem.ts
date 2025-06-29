import { Entity, Vector2D } from '@/types/minimalLegion';

export class CombatSystem {
  private attackCooldowns = new Map<string, number>();
  
  constructor() {}
  
  update(
    deltaTime: number,
    player: Entity,
    allies: Entity[],
    enemies: Entity[],
    spawnProjectile: (projectile: Entity) => void
  ) {
    // Update attack cooldowns
    for (const [id, cooldown] of this.attackCooldowns.entries()) {
      if (cooldown > 0) {
        this.attackCooldowns.set(id, cooldown - deltaTime);
      }
    }
    
    // Player attack
    this.handlePlayerAttack(player, enemies, spawnProjectile);
    
    // Ally attacks
    allies.forEach(ally => {
      this.handleAllyAttack(ally, enemies, spawnProjectile);
    });
    
    // Enemy attacks
    enemies.forEach(enemy => {
      this.handleEnemyAttack(enemy, player, allies, spawnProjectile);
    });
  }
  
  private handlePlayerAttack(
    player: Entity,
    enemies: Entity[],
    spawnProjectile: (projectile: Entity) => void
  ) {
    const attackCooldown = 0.5; // 2 attacks per second
    const range = 150;
    
    if ((this.attackCooldowns.get(player.id) || 0) > 0) return;
    
    // Find closest enemy
    const target = this.findClosestTarget(player.position, enemies, range);
    if (!target) return;
    
    // Fire projectile
    this.fireProjectile(player, target, spawnProjectile);
    this.attackCooldowns.set(player.id, attackCooldown);
  }
  
  private handleAllyAttack(
    ally: Entity,
    enemies: Entity[],
    spawnProjectile: (projectile: Entity) => void
  ) {
    const attackCooldown = 0.6; // Slightly slower than player
    const range = 120;
    
    if ((this.attackCooldowns.get(ally.id) || 0) > 0) return;
    
    const target = this.findClosestTarget(ally.position, enemies, range);
    if (!target) return;
    
    this.fireProjectile(ally, target, spawnProjectile);
    this.attackCooldowns.set(ally.id, attackCooldown);
  }
  
  private handleEnemyAttack(
    enemy: Entity,
    player: Entity,
    allies: Entity[],
    spawnProjectile: (projectile: Entity) => void
  ) {
    const attackCooldown = 1.0;
    const range = 100;
    
    if ((this.attackCooldowns.get(enemy.id) || 0) > 0) return;
    
    // Target player first, then allies
    const targets = [player, ...allies];
    const target = this.findClosestTarget(enemy.position, targets, range);
    if (!target) return;
    
    this.fireProjectile(enemy, target, spawnProjectile);
    this.attackCooldowns.set(enemy.id, attackCooldown);
  }
  
  private fireProjectile(
    shooter: Entity,
    target: Entity,
    spawnProjectile: (projectile: Entity) => void
  ) {
    const direction = this.normalize({
      x: target.position.x - shooter.position.x,
      y: target.position.y - shooter.position.y
    });
    
    const projectileSpeed = 400; // pixels per second
    
    const projectile: Entity = {
      id: `projectile-${Date.now()}-${Math.random()}`,
      position: { ...shooter.position },
      velocity: {
        x: direction.x * projectileSpeed,
        y: direction.y * projectileSpeed
      },
      health: 1,
      maxHealth: 1,
      damage: shooter.damage,
      speed: projectileSpeed,
      size: 4,
      color: shooter.type === 'enemy' ? '#DC2626' : '#3B82F6',
      type: 'projectile',
      owner: shooter.id
    };
    
    spawnProjectile(projectile);
  }
  
  private findClosestTarget(
    position: Vector2D,
    targets: Entity[],
    maxRange: number
  ): Entity | null {
    let closest: Entity | null = null;
    let closestDistance = maxRange;
    
    for (const target of targets) {
      const distance = this.getDistance(position, target.position);
      if (distance < closestDistance) {
        closest = target;
        closestDistance = distance;
      }
    }
    
    return closest;
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