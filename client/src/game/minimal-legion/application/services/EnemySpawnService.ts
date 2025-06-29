import { Enemy, EnemyType } from '../../domain/entities/Enemy';
import { Position } from '../../domain/value-objects/Position';

export interface WaveConfig {
  enemyCount: number;
  enemyTypes: EnemyType[];
  spawnInterval: number; // seconds
  difficultyMultiplier: number;
}

export interface EnemySpawnedEvent {
  enemy: Enemy;
  wave: number;
}

export class EnemySpawnService {
  private spawnTimer: number = 0;
  private enemiesSpawnedInWave: number = 0;
  private readonly spawnDistance: number = 450;
  
  private eventListeners: Array<(event: EnemySpawnedEvent) => void> = [];

  // 이벤트 리스너 등록
  onEnemySpawned(callback: (event: EnemySpawnedEvent) => void): void {
    this.eventListeners.push(callback);
  }

  // 스폰 시스템 업데이트
  update(
    deltaTime: number,
    currentWave: number,
    playerPosition: Position,
    waveConfig: WaveConfig
  ): void {
    this.spawnTimer += deltaTime;

    if (this.shouldSpawnEnemy(waveConfig)) {
      this.spawnEnemy(currentWave, playerPosition, waveConfig);
      this.spawnTimer = 0;
    }
  }

  private shouldSpawnEnemy(waveConfig: WaveConfig): boolean {
    return (
      this.enemiesSpawnedInWave < waveConfig.enemyCount &&
      this.spawnTimer >= waveConfig.spawnInterval
    );
  }

  private spawnEnemy(wave: number, playerPosition: Position, waveConfig: WaveConfig): void {
    // 랜덤 적 타입 선택
    const enemyType = this.selectEnemyType(wave, waveConfig.enemyTypes);
    
    // 스폰 위치 계산
    const spawnPosition = this.calculateSpawnPosition(playerPosition);
    
    // 적 생성
    const enemy = this.createEnemy(enemyType, spawnPosition);
    
    // 웨이브에 따른 스탯 조정
    const scaledEnemy = enemy.scaleForWave(wave);

    // 이벤트 발생
    this.eventListeners.forEach(callback => {
      callback({ enemy: scaledEnemy, wave });
    });

    this.enemiesSpawnedInWave++;
  }

  private selectEnemyType(wave: number, availableTypes: EnemyType[]): EnemyType {
    // 웨이브에 따라 사용 가능한 적 타입 제한
    const maxTypeIndex = Math.min(availableTypes.length, Math.ceil(wave / 2));
    const waveAvailableTypes = availableTypes.slice(0, maxTypeIndex);
    
    if (waveAvailableTypes.length === 0) {
      return 'normal'; // 기본값
    }

    return waveAvailableTypes[Math.floor(Math.random() * waveAvailableTypes.length)];
  }

  private calculateSpawnPosition(playerPosition: Position): Position {
    const angle = Math.random() * Math.PI * 2;
    const x = playerPosition.x + Math.cos(angle) * this.spawnDistance;
    const y = playerPosition.y + Math.sin(angle) * this.spawnDistance;
    
    return new Position(x, y);
  }

  private createEnemy(type: EnemyType, position: Position): Enemy {
    const id = `enemy-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

    switch (type) {
      case 'normal':
        return Enemy.createNormal(id, position);
      case 'fast':
        return Enemy.createFast(id, position);
      case 'tank':
        return Enemy.createTank(id, position);
      case 'ranged':
        return Enemy.createRanged(id, position);
      default:
        return Enemy.createNormal(id, position);
    }
  }

  // 웨이브 설정 생성
  static createWaveConfig(wave: number): WaveConfig {
    const baseEnemyCount = 20;
    const baseSpawnInterval = 1.0;
    
    return {
      enemyCount: baseEnemyCount + (wave - 1) * 5,
      enemyTypes: this.getAvailableEnemyTypes(wave),
      spawnInterval: Math.max(0.3, baseSpawnInterval - (wave - 1) * 0.05),
      difficultyMultiplier: 1 + (wave - 1) * 0.1
    };
  }

  private static getAvailableEnemyTypes(wave: number): EnemyType[] {
    const allTypes: EnemyType[] = ['normal', 'fast', 'tank', 'ranged'];
    
    // 웨이브에 따라 점진적으로 새로운 적 타입 해금
    if (wave <= 2) return ['normal'];
    if (wave <= 4) return ['normal', 'fast'];
    if (wave <= 7) return ['normal', 'fast', 'tank'];
    return allTypes;
  }

  // 새 웨이브 시작
  startNewWave(): void {
    this.enemiesSpawnedInWave = 0;
    this.spawnTimer = 0;
  }

  // 현재 웨이브 완료 여부
  isWaveComplete(waveConfig: WaveConfig): boolean {
    return this.enemiesSpawnedInWave >= waveConfig.enemyCount;
  }

  // 스폰 통계
  getSpawnStats(): { spawned: number; timer: number } {
    return {
      spawned: this.enemiesSpawnedInWave,
      timer: this.spawnTimer
    };
  }

  // 긴급 상황 처리 (보스 스폰 등)
  spawnBoss(playerPosition: Position, wave: number): Enemy {
    const bossPosition = this.calculateSpawnPosition(playerPosition);
    const bossId = `boss-${Date.now()}`;
    
    // 보스는 탱커 기반으로 생성하되 더 강하게
    const boss = Enemy.createTank(bossId, bossPosition);
    return boss.scaleForWave(wave * 2); // 2배 강화
  }
}