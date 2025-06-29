import { Player } from '../../domain/entities/Player';
import { Enemy } from '../../domain/entities/Enemy';
import { Ally } from '../../domain/entities/Ally';
import { Projectile } from '../../domain/entities/Projectile';
import { Position } from '../../domain/value-objects/Position';

export interface MovementBounds {
  width: number;
  height: number;
  margin: number;
}

export class MovementService {
  private readonly defaultBounds: MovementBounds = {
    width: 1200,
    height: 800,
    margin: 100
  };

  // 모든 엔티티 이동 업데이트
  update(
    deltaTime: number,
    player: Player,
    allies: Ally[],
    enemies: Enemy[],
    projectiles: Projectile[],
    bounds: MovementBounds = this.defaultBounds
  ): void {
    // 플레이어 이동
    this.updatePlayerMovement(deltaTime, player, bounds);

    // 아군 이동 (플레이어 따라다니기)
    this.updateAllyMovement(deltaTime, allies, player);

    // 적 이동 (플레이어 추적)
    this.updateEnemyMovement(deltaTime, enemies, player);

    // 투사체 이동
    this.updateProjectileMovement(deltaTime, projectiles, bounds);
  }

  private updatePlayerMovement(deltaTime: number, player: Player, bounds: MovementBounds): void {
    if (!player.isAlive) return;

    // 기본 이동 업데이트
    player.update(deltaTime);

    // 경계 내에 플레이어 유지
    this.keepPlayerInBounds(player, bounds);
  }

  private updateAllyMovement(deltaTime: number, allies: Ally[], player: Player): void {
    for (const ally of allies) {
      if (!ally.isAlive) continue;

      // 플레이어 따라다니기
      ally.followPlayer(player.position, allies);
      ally.update(deltaTime);
    }
  }

  private updateEnemyMovement(deltaTime: number, enemies: Enemy[], player: Player): void {
    if (!player.isAlive) return;

    for (const enemy of enemies) {
      if (!enemy.isAlive) continue;

      // 다른 적들과의 충돌 회피를 고려한 이동
      const otherEnemies = enemies.filter(e => e.id !== enemy.id && e.isAlive);
      this.moveEnemyWithCollisionAvoidance(enemy, player.position, otherEnemies);
      enemy.update(deltaTime);
    }
  }

  // 충돌 회피를 포함한 적 이동
  private moveEnemyWithCollisionAvoidance(
    enemy: Enemy, 
    targetPosition: Position, 
    otherEnemies: Enemy[]
  ): void {
    // 기본 이동 방향 계산
    let moveDirection = targetPosition.subtract(enemy.position);
    
    // 충돌 회피 검사
    const avoidanceRadius = enemy.size * 1.5; // 적 크기의 1.5배
    let avoidanceForce = Position.zero();
    
    for (const otherEnemy of otherEnemies) {
      const distance = enemy.position.distanceTo(otherEnemy.position);
      const minDistance = avoidanceRadius;
      
      if (distance < minDistance && distance > 0) {
        // 회피 방향 계산 (서로 밀어냄)
        const pushDirection = enemy.position.subtract(otherEnemy.position);
        const pushStrength = (minDistance - distance) / minDistance;
        const pushForce = pushDirection.normalize().multiply(pushStrength * 100);
        
        avoidanceForce = avoidanceForce.add(pushForce);
        
        // 디버그 로그
        console.log(`Enemy ${enemy.id} avoiding ${otherEnemy.id}, distance: ${distance.toFixed(1)}, strength: ${pushStrength.toFixed(2)}`);
      }
    }
    
    // 최종 이동 방향 = 타겟 방향 + 회피 힘
    const finalDirection = moveDirection.normalize().multiply(enemy.speed * 60)
      .add(avoidanceForce);
    
    if (finalDirection.x !== 0 || finalDirection.y !== 0) {
      const velocity = Velocity.create(finalDirection.x, finalDirection.y);
      enemy.setVelocity(velocity);
    } else {
      // 기본 이동
      enemy.setTarget(targetPosition);
      enemy.moveTowardsTarget();
    }
  }

  private updateProjectileMovement(
    deltaTime: number,
    projectiles: Projectile[],
    bounds: MovementBounds
  ): void {
    for (const projectile of projectiles) {
      if (!projectile.isAlive) continue;

      projectile.update(deltaTime);

      // 화면 밖으로 나간 투사체나 수명이 다한 투사체 제거 마킹
      if (projectile.isOutOfBounds(bounds.width, bounds.height, bounds.margin) || 
          projectile.isExpired) {
        projectile.takeDamage(1); // 투사체 파괴
      }
    }
  }

  private keepPlayerInBounds(player: Player, bounds: MovementBounds): void {
    const position = player.position;
    const size = player.size / 2;

    let newX = position.x;
    let newY = position.y;

    // X축 경계 확인
    if (newX - size < -bounds.width / 2) {
      newX = -bounds.width / 2 + size;
    } else if (newX + size > bounds.width / 2) {
      newX = bounds.width / 2 - size;
    }

    // Y축 경계 확인
    if (newY - size < -bounds.height / 2) {
      newY = -bounds.height / 2 + size;
    } else if (newY + size > bounds.height / 2) {
      newY = bounds.height / 2 - size;
    }

    // 위치가 변경된 경우에만 업데이트
    if (newX !== position.x || newY !== position.y) {
      player.moveTo(new Position(newX, newY));
    }
  }

  // 플레이어 입력 처리
  handlePlayerInput(player: Player, inputDirection: { x: number; y: number }): void {
    if (!player.isAlive) return;

    if (inputDirection.x === 0 && inputDirection.y === 0) {
      player.stopMoving();
    } else {
      const direction = new Position(inputDirection.x, inputDirection.y);
      player.moveInDirection(direction);
    }
  }

  // 적 스폰 위치 계산
  calculateEnemySpawnPosition(
    playerPosition: Position,
    spawnDistance: number = 450
  ): Position {
    const angle = Math.random() * Math.PI * 2;
    const x = playerPosition.x + Math.cos(angle) * spawnDistance;
    const y = playerPosition.y + Math.sin(angle) * spawnDistance;
    
    return new Position(x, y);
  }

  // 경로 찾기 (간단한 A* 알고리즘, 추후 확장용)
  findPath(start: Position, end: Position, obstacles: Position[]): Position[] {
    // TODO: 장애물 회피 경로 찾기 구현
    return [start, end];
  }

  // 충돌 회피 (Boids 알고리즘 기반)
  calculateSeparationForce(
    entity: Position,
    neighbors: Position[],
    separationRadius: number
  ): Position {
    let separationForce = Position.zero();
    let count = 0;

    for (const neighbor of neighbors) {
      const distance = entity.distanceTo(neighbor);
      if (distance > 0 && distance < separationRadius) {
        // 거리가 가까울수록 더 강한 회피력 적용
        const strength = (separationRadius - distance) / separationRadius;
        const diff = entity.subtract(neighbor).normalize().multiply(strength);
        separationForce = separationForce.add(diff);
        count++;
      }
    }

    if (count > 0) {
      separationForce = separationForce.multiply(1 / count);
    }

    return separationForce;
  }

  // 그룹 이동 (아군들이 함께 움직이도록)
  calculateCohesionForce(entity: Position, neighbors: Position[]): Position {
    if (neighbors.length === 0) return Position.zero();

    let center = Position.zero();
    for (const neighbor of neighbors) {
      center = center.add(neighbor);
    }
    center = center.multiply(1 / neighbors.length);

    return center.subtract(entity).normalize();
  }
}