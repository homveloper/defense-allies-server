import { Entity, Vector2D } from '../types/minimalLegion';

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
    
    // 엔티티 유효성 검사
    if (!player || player.health <= 0) return;
    if (!enemies || enemies.length === 0) return;
    
    // Update attack cooldowns
    for (const [id, cooldown] of this.attackCooldowns.entries()) {
      if (cooldown > 0) {
        this.attackCooldowns.set(id, Math.max(0, cooldown - deltaTime));
      }
    }
    
    // 상태 유효성 검사
    enemies = enemies.filter(e => e && e.health > 0);
    allies = allies.filter(a => a && a.health > 0);
    
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
    const attackCooldown = 0.2; // 더 빠른 공격 (0.3 -> 0.2초)
    const range = 400; // 더 넓은 공격 범위 (250 -> 400px)
    
    if ((this.attackCooldowns.get(player.id) || 0) > 0) {
      return;
    }
    
    const target = this.findClosestTarget(player.position, enemies, range);
    if (!target) {
      return;
    }
    
    this.fireProjectile(player, target, spawnProjectile);
    this.attackCooldowns.set(player.id, attackCooldown);
  }
  
  private handleAllyAttack(
    ally: Entity,
    enemies: Entity[],
    spawnProjectile: (projectile: Entity) => void
  ) {
    const attackCooldown = 0.4; // 더 빠른 공격 (0.6 -> 0.4초)
    const range = 150; // 더 넓은 범위 (120 -> 150)
    
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
    const range = 250; // 플레이어와 같은 범위로 조정
    
    if ((this.attackCooldowns.get(enemy.id) || 0) > 0) return;
    
    const targets = [player, ...allies];
    const target = this.findClosestTarget(enemy.position, targets, range);
    if (!target) {
      return;
    }
    
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
      id: `proj_${shooter.type}_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
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