import { Player } from '../../domain/entities/Player';
import { Enemy } from '../../domain/entities/Enemy';
import { Ally } from '../../domain/entities/Ally';
import { Projectile } from '../../domain/entities/Projectile';
import { GameEntity } from '../../domain/entities/GameEntity';
import { Position } from '../../domain/value-objects/Position';

export interface ProjectileCreatedEvent {
  projectile: Projectile;
}

export interface EntityDeathEvent {
  entity: GameEntity;
  killer?: GameEntity;
}

export class CombatService {
  private eventListeners: {
    projectileCreated: Array<(event: ProjectileCreatedEvent) => void>;
    entityDeath: Array<(event: EntityDeathEvent) => void>;
  } = {
    projectileCreated: [],
    entityDeath: []
  };

  // 이벤트 리스너 등록
  onProjectileCreated(callback: (event: ProjectileCreatedEvent) => void): void {
    this.eventListeners.projectileCreated.push(callback);
  }

  onEntityDeath(callback: (event: EntityDeathEvent) => void): void {
    this.eventListeners.entityDeath.push(callback);
  }

  // 전투 업데이트
  update(
    deltaTime: number,
    currentTime: number,
    player: Player,
    allies: Ally[],
    enemies: Enemy[],
    projectiles: Projectile[]
  ): void {
    // 플레이어 공격 처리
    this.handlePlayerCombat(currentTime, player, enemies);

    // 아군 공격 처리
    this.handleAllyCombat(currentTime, allies, enemies);

    // 적 공격 처리
    this.handleEnemyCombat(currentTime, enemies, player, allies);

    // 투사체 충돌 처리
    this.handleProjectileCollisions(projectiles, player, allies, enemies);
  }

  private handlePlayerCombat(currentTime: number, player: Player, enemies: Enemy[]): void {
    if (!player.isAlive || !player.canAttack(currentTime)) return;

    const target = player.findNearestEnemy(enemies);
    if (!target) return;

    // 타겟 예측 위치로 투사체 생성
    const predictedPosition = player.predictTargetPosition(target);
    this.createSmartProjectile(player, target, predictedPosition, currentTime);
    player.attack(currentTime);
  }

  private handleAllyCombat(currentTime: number, allies: Ally[], enemies: Enemy[]): void {
    for (const ally of allies) {
      if (!ally.isAlive || !ally.canAttack(currentTime)) continue;

      const target = ally.findNearestEnemy(enemies);
      if (!target) continue;

      // 아군도 기본적인 타겟 예측 사용 (현재 위치 기준)
      this.createProjectile(ally, target, currentTime);
      ally.attack(currentTime);
    }
  }

  private handleEnemyCombat(
    currentTime: number,
    enemies: Enemy[],
    player: Player,
    allies: Ally[]
  ): void {
    const targets = [player, ...allies].filter(entity => entity.isAlive);

    for (const enemy of enemies) {
      if (!enemy.isAlive) continue;

      // 가장 가까운 대상 찾기
      let nearestTarget: GameEntity | null = null;
      let nearestDistance = enemy.attackRange;

      for (const target of targets) {
        const distance = enemy.position.distanceTo(target.position);
        if (distance < nearestDistance) {
          nearestTarget = target;
          nearestDistance = distance;
        }
      }

      if (nearestTarget && enemy.canAttack(currentTime, nearestTarget)) {
        this.createProjectile(enemy, nearestTarget, currentTime);
        enemy.attack(currentTime);
      }
    }
  }

  private createProjectile(shooter: GameEntity, target: GameEntity, currentTime: number): void {
    const projectile = Projectile.create(
      shooter.id,
      shooter.getType() as 'player' | 'enemy' | 'ally',
      shooter.position,
      target.position,
      shooter.damage
    );

    // 이벤트 발생
    this.eventListeners.projectileCreated.forEach(callback => {
      callback({ projectile });
    });
  }

  // 예측 위치로 발사하는 스마트 투사체
  private createSmartProjectile(
    shooter: GameEntity, 
    target: GameEntity, 
    predictedPosition: Position, 
    currentTime: number
  ): void {
    const projectile = Projectile.create(
      shooter.id,
      shooter.getType() as 'player' | 'enemy' | 'ally',
      shooter.position,
      predictedPosition, // 예측 위치로 발사
      shooter.damage
    );

    // 원래 타겟 정보도 저장 (충돌 검사용)
    projectile.originalTarget = target;

    // 이벤트 발생
    this.eventListeners.projectileCreated.forEach(callback => {
      callback({ projectile });
    });
  }

  private handleProjectileCollisions(
    projectiles: Projectile[],
    player: Player,
    allies: Ally[],
    enemies: Enemy[]
  ): void {
    const allTargets = [player, ...allies, ...enemies];

    for (const projectile of projectiles) {
      if (!projectile.isAlive) continue;

      for (const target of allTargets) {
        if (!target.isAlive || !projectile.canHit(target)) continue;

        if (projectile.isCollidingWith(target)) {
          const hitResult = projectile.onHit(target);
          target.takeDamage(hitResult.damageDealt);

          // 대상이 죽었는지 확인
          if (target.isDead) {
            this.eventListeners.entityDeath.forEach(callback => {
              callback({ entity: target, killer: projectile });
            });
          }

          break; // 하나의 투사체는 하나의 대상만 명중
        }
      }
    }
  }

  // 특수 공격 (폭발, 관통 등) - 추후 확장용
  createSpecialProjectile(
    shooter: GameEntity,
    targetPosition: Position,
    type: 'explosive' | 'piercing' | 'homing'
  ): void {
    // TODO: 특수 투사체 구현
  }

  // 범위 공격 처리
  handleAreaOfEffectAttack(
    center: Position,
    radius: number,
    damage: number,
    targets: GameEntity[]
  ): GameEntity[] {
    const hitTargets: GameEntity[] = [];

    for (const target of targets) {
      if (!target.isAlive) continue;

      const distance = center.distanceTo(target.position);
      if (distance <= radius) {
        target.takeDamage(damage);
        hitTargets.push(target);

        if (target.isDead) {
          this.eventListeners.entityDeath.forEach(callback => {
            callback({ entity: target });
          });
        }
      }
    }

    return hitTargets;
  }
}