import { Entity, EnemyType } from '@/types/minimalLegion';

export const enemyTypes: EnemyType[] = [
  {
    id: 'normal',
    name: '일반병',
    health: 30,
    damage: 5,
    speed: 3,
    size: 18,
    color: '#EF4444',
    experience: 10
  },
  {
    id: 'fast',
    name: '질주병',
    health: 20,
    damage: 3,
    speed: 8,
    size: 16,
    color: '#F59E0B',
    experience: 15
  },
  {
    id: 'tank',
    name: '탱커',
    health: 100,
    damage: 10,
    speed: 2,
    size: 24,
    color: '#991B1B',
    experience: 25
  },
  {
    id: 'ranged',
    name: '궁수',
    health: 25,
    damage: 8,
    speed: 3,
    size: 18,
    color: '#DC2626',
    experience: 20,
    special: ['ranged']
  }
];

export class EnemySpawnSystem {
  private spawnTimer = 0;
  private spawnInterval = 2; // seconds
  private enemiesSpawnedInWave = 0;
  private maxEnemiesPerWave = 10;
  
  constructor() {}
  
  update(deltaTime: number, wave: number, spawnEnemy: (enemy: Entity) => void) {
    this.spawnTimer += deltaTime;
    
    // Adjust spawn rate and enemy count based on wave
    this.maxEnemiesPerWave = 10 + (wave - 1) * 5;
    this.spawnInterval = Math.max(0.5, 2 - (wave - 1) * 0.1);
    
    if (this.spawnTimer >= this.spawnInterval && this.enemiesSpawnedInWave < this.maxEnemiesPerWave) {
      this.spawnTimer = 0;
      this.spawnEnemy(wave, spawnEnemy);
    }
  }
  
  private spawnEnemy(wave: number, spawnEnemy: (enemy: Entity) => void) {
    // Select enemy type based on wave
    let availableTypes = enemyTypes.slice(0, Math.min(enemyTypes.length, Math.ceil(wave / 2)));
    const enemyType = availableTypes[Math.floor(Math.random() * availableTypes.length)];
    
    // Random spawn position on edges
    const side = Math.floor(Math.random() * 4);
    let x, y;
    
    switch (side) {
      case 0: // top
        x = Math.random() * 1200;
        y = -50;
        break;
      case 1: // right
        x = 1250;
        y = Math.random() * 800;
        break;
      case 2: // bottom
        x = Math.random() * 1200;
        y = 850;
        break;
      case 3: // left
        x = -50;
        y = Math.random() * 800;
        break;
      default:
        x = 600;
        y = 400;
    }
    
    // Scale enemy stats based on wave
    const statMultiplier = 1 + (wave - 1) * 0.1;
    
    const enemy: Entity = {
      id: `enemy-${Date.now()}-${Math.random()}`,
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
    
    spawnEnemy(enemy);
    this.enemiesSpawnedInWave++;
  }
  
  startNewWave() {
    this.enemiesSpawnedInWave = 0;
    this.spawnTimer = 0;
  }
  
  isWaveComplete(): boolean {
    return this.enemiesSpawnedInWave >= this.maxEnemiesPerWave;
  }
}