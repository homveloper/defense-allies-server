import { Entity, EnemyType } from '../types/minimalLegion';

export const enemyTypes: EnemyType[] = [
  {
    id: 'normal',
    name: '일반병',
    health: 150,
    damage: 5,
    speed: 3,
    size: 18,
    color: '#EF4444',
    experience: 10
  },
  {
    id: 'fast',
    name: '질주병',
    health: 100,
    damage: 3,
    speed: 8,
    size: 16,
    color: '#F59E0B',
    experience: 15
  },
  {
    id: 'tank',
    name: '탱커',
    health: 300,
    damage: 10,
    speed: 2,
    size: 24,
    color: '#991B1B',
    experience: 25
  },
  {
    id: 'ranged',
    name: '궁수',
    health: 120,
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
  private spawnInterval = 1; // seconds - 더 빠르게 스폰
  private enemiesSpawnedInWave = 0;
  private maxEnemiesPerWave = 20; // 더 많은 적
  
  constructor() {
  }
  
  update(deltaTime: number, wave: number, playerPosition: { x: number; y: number }, spawnEnemy: (enemy: Entity) => void) {
    this.spawnTimer += deltaTime;
    
    // Adjust spawn rate and enemy count based on wave
    this.maxEnemiesPerWave = 20 + (wave - 1) * 5;
    this.spawnInterval = Math.max(0.3, 1 - (wave - 1) * 0.05);
    
    if (this.spawnTimer >= this.spawnInterval && this.enemiesSpawnedInWave < this.maxEnemiesPerWave) {
      this.spawnTimer = 0;
      this.spawnEnemy(wave, playerPosition, spawnEnemy);
    }
  }
  
  private spawnEnemy(wave: number, playerPosition: { x: number; y: number }, spawnEnemy: (enemy: Entity) => void) {
    // Select enemy type based on wave
    let availableTypes = enemyTypes.slice(0, Math.min(enemyTypes.length, Math.ceil(wave / 2)));
    if (availableTypes.length === 0) availableTypes = [enemyTypes[0]]; // 최소 1개 보장
    const enemyType = availableTypes[Math.floor(Math.random() * availableTypes.length)];
    
    // Random spawn position around player (화면 밖)
    const spawnDistance = 450; // 플레이어로부터 거리 (700 -> 450px)
    const angle = Math.random() * Math.PI * 2; // 랜덤 각도
    
    const x = playerPosition.x + Math.cos(angle) * spawnDistance;
    const y = playerPosition.y + Math.sin(angle) * spawnDistance;
    
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