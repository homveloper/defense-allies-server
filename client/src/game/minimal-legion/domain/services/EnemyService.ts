import { Entity, EnemyType } from '../../types/minimalLegion';
import { IEnemyRepository } from '../../data/repositories/EnemyRepository';
import { enemyTypes } from '../../systems/enemySpawnSystem';

export interface IEnemyService {
  spawnEnemy(wave: number, playerPosition: { x: number; y: number }): Entity;
  addEnemy(enemy: Entity): void;
  getVisibleEnemies(camera: { x: number; y: number }, screenWidth: number, screenHeight: number): Entity[];
  updateEnemyPosition(enemyId: string, newPosition: { x: number; y: number }): void;
  updateEnemyHealth(enemyId: string, newHealth: number): void;
  removeEnemy(enemyId: string): void;
  getEnemyCount(): number;
  getEnemiesInRange(center: { x: number; y: number }, radius: number): Entity[];
}

export class EnemyService implements IEnemyService {
  constructor(private enemyRepository: IEnemyRepository) {}

  spawnEnemy(wave: number, playerPosition: { x: number; y: number }): Entity {
    // 웨이브에 따른 적 타입 선택
    const availableTypes = enemyTypes.slice(0, Math.min(enemyTypes.length, Math.ceil(wave / 2)));
    if (availableTypes.length === 0) {
      throw new Error('No enemy types available');
    }
    
    const enemyType = availableTypes[Math.floor(Math.random() * availableTypes.length)];
    
    // 플레이어 주변 원형으로 스폰 위치 계산
    const spawnDistance = 700;
    const angle = Math.random() * Math.PI * 2;
    const x = playerPosition.x + Math.cos(angle) * spawnDistance;
    const y = playerPosition.y + Math.sin(angle) * spawnDistance;
    
    // 웨이브에 따른 스탯 증가
    const statMultiplier = 1 + (wave - 1) * 0.1;
    
    const enemy: Entity = {
      id: `enemy-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      position: { x, y },
      velocity: { x: 0, y: 0 },
      health: Math.floor(enemyType.health * statMultiplier),
      maxHealth: Math.floor(enemyType.health * statMultiplier),
      damage: Math.floor(enemyType.damage * statMultiplier),
      speed: enemyType.speed,
      size: enemyType.size,
      color: enemyType.color,
      type: 'enemy'
    };
    
    this.enemyRepository.add(enemy);
    console.log(`EnemyService: Spawned ${enemyType.name} at (${x.toFixed(0)}, ${y.toFixed(0)})`);
    
    return enemy;
  }

  addEnemy(enemy: Entity): void {
    this.enemyRepository.add(enemy);
  }

  getVisibleEnemies(camera: { x: number; y: number }, screenWidth: number, screenHeight: number): Entity[] {
    const margin = 100; // 화면 밖 여유분
    return this.enemyRepository.getAll().filter(enemy => {
      const screenX = enemy.position.x - camera.x;
      const screenY = enemy.position.y - camera.y;
      
      return screenX >= -margin && 
             screenX <= screenWidth + margin && 
             screenY >= -margin && 
             screenY <= screenHeight + margin;
    });
  }

  updateEnemyPosition(enemyId: string, newPosition: { x: number; y: number }): void {
    this.enemyRepository.update(enemyId, { position: newPosition });
  }

  updateEnemyHealth(enemyId: string, newHealth: number): void {
    const enemy = this.enemyRepository.getById(enemyId);
    if (enemy) {
      if (newHealth <= 0) {
        this.removeEnemy(enemyId);
      } else {
        this.enemyRepository.update(enemyId, { health: newHealth });
      }
    }
  }

  removeEnemy(enemyId: string): void {
    this.enemyRepository.remove(enemyId);
  }

  getEnemyCount(): number {
    return this.enemyRepository.count();
  }

  getEnemiesInRange(center: { x: number; y: number }, radius: number): Entity[] {
    return this.enemyRepository.getInRange(center, radius);
  }
}