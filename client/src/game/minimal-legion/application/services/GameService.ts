import { Player } from '../../domain/entities/Player';
import { Enemy } from '../../domain/entities/Enemy';
import { Ally } from '../../domain/entities/Ally';
import { Projectile } from '../../domain/entities/Projectile';
import { Position } from '../../domain/value-objects/Position';

import { CombatService } from './CombatService';
import { MovementService } from './MovementService';
import { EnemySpawnService, WaveConfig } from './EnemySpawnService';
import { PhysicsService } from './PhysicsService';

export interface GameState {
  player: Player;
  allies: Ally[];
  enemies: Enemy[];
  projectiles: Projectile[];
  camera: { x: number; y: number };
  wave: number;
  score: number;
  isGameRunning: boolean;
  isPaused: boolean;
}

export interface ConversionConfig {
  conversionRate: number; // 0.0 ~ 1.0
  enabled: boolean;
}

export class GameService {
  private gameState: GameState;
  private lastUpdateTime: number = 0;
  
  // 서비스들
  private combatService: CombatService;
  private movementService: MovementService;
  private enemySpawnService: EnemySpawnService;
  private physicsService: PhysicsService;
  
  // 설정 - 군단 시스템
  private conversionConfig: ConversionConfig = {
    conversionRate: 0.8, // 80% 확률로 증가 (군단 시스템)
    enabled: true
  };

  constructor() {
    this.combatService = new CombatService();
    this.movementService = new MovementService();
    this.enemySpawnService = new EnemySpawnService();
    this.physicsService = new PhysicsService();
    
    this.gameState = this.createInitialGameState();
    this.setupEventListeners();
  }

  private createInitialGameState(): GameState {
    return {
      player: Player.create(),
      allies: [],
      enemies: [],
      projectiles: [],
      camera: { x: -600, y: -400 },
      wave: 1,
      score: 0,
      isGameRunning: false,
      isPaused: false
    };
  }

  private setupEventListeners(): void {
    // 투사체 생성 이벤트
    this.combatService.onProjectileCreated(({ projectile }) => {
      this.gameState.projectiles.push(projectile);
      this.physicsService.addEntity(projectile, 'projectile');
    });

    // 엔티티 사망 이벤트
    this.combatService.onEntityDeath(({ entity, killer }) => {
      this.handleEntityDeath(entity, killer);
    });

    // 적 스폰 이벤트
    this.enemySpawnService.onEnemySpawned(({ enemy, wave }) => {
      this.gameState.enemies.push(enemy);
      this.physicsService.addEntity(enemy, 'enemy');
    });
  }

  // 게임 초기화
  initialize(): void {
    this.gameState = this.createInitialGameState();
    this.gameState.isGameRunning = true;
    this.lastUpdateTime = Date.now();
    
    // 플레이어를 물리 시스템에 추가
    this.physicsService.addEntity(this.gameState.player, 'player');
    
    console.log('OOP Game Service initialized with physics');
  }

  // 게임 업데이트 (메인 루프)
  update(deltaTime: number): void {
    if (!this.gameState.isGameRunning || this.gameState.isPaused) return;

    const currentTime = Date.now() / 1000; // seconds

    // 서비스들 업데이트
    this.updateMovement(deltaTime);
    this.updatePhysics(deltaTime);
    this.updateCombat(deltaTime, currentTime);
    this.updateEnemySpawn(deltaTime);
    
    // 죽은 엔티티들 정리
    this.cleanupDeadEntities();
    
    // 웨이브 진행 확인
    this.checkWaveProgression();

    this.lastUpdateTime = Date.now();
  }

  private updateMovement(deltaTime: number): void {
    this.movementService.update(
      deltaTime,
      this.gameState.player,
      this.gameState.allies,
      this.gameState.enemies,
      this.gameState.projectiles
    );
  }

  private updateCombat(deltaTime: number, currentTime: number): void {
    this.combatService.update(
      deltaTime,
      currentTime,
      this.gameState.player,
      this.gameState.allies,
      this.gameState.enemies,
      this.gameState.projectiles
    );
  }

  private updatePhysics(deltaTime: number): void {
    this.physicsService.update(deltaTime);
  }

  private updateEnemySpawn(deltaTime: number): void {
    const waveConfig = EnemySpawnService.createWaveConfig(this.gameState.wave);
    
    this.enemySpawnService.update(
      deltaTime,
      this.gameState.wave,
      this.gameState.player.position,
      waveConfig
    );
  }

  private handleEntityDeath(entity: any, killer?: any): void {
    if (entity.getType() === 'enemy') {
      const enemy = entity as Enemy;
      const reward = enemy.calculateReward();
      
      this.gameState.score += reward.score;
      
      // 군단 시스템: 적 → 아군 변환 시도
      if (this.conversionConfig.enabled && this.shouldConvertEnemy()) {
        const ally = Ally.createFromEnemy(enemy);
        ally.applyConversionBonus();
        this.gameState.allies.push(ally);
        this.physicsService.addEntity(ally, 'ally');
        
        console.log(`🛡️ LEGION SYSTEM: ${enemy.enemyType} enemy recruited as ally! Army size: ${this.gameState.allies.length}`);
      }
      
      // 죽은 적을 물리 시스템에서 제거
      this.physicsService.removeEntity(enemy.id);
    } else if (entity.getType() === 'player') {
      this.gameState.isGameRunning = false;
      console.log('Game Over: Player died');
    }
  }

  private shouldConvertEnemy(): boolean {
    return Math.random() < this.conversionConfig.conversionRate;
  }

  private cleanupDeadEntities(): void {
    // 죽은 엔티티들을 물리 시스템에서 제거
    const deadEnemies = this.gameState.enemies.filter(e => !e.isAlive);
    const deadAllies = this.gameState.allies.filter(a => !a.isAlive);
    const deadProjectiles = this.gameState.projectiles.filter(p => !p.isAlive);
    
    deadEnemies.forEach(enemy => this.physicsService.removeEntity(enemy.id));
    deadAllies.forEach(ally => this.physicsService.removeEntity(ally.id));
    deadProjectiles.forEach(projectile => this.physicsService.removeEntity(projectile.id));
    
    // 게임 상태에서 제거
    this.gameState.enemies = this.gameState.enemies.filter(e => e.isAlive);
    this.gameState.allies = this.gameState.allies.filter(a => a.isAlive);
    this.gameState.projectiles = this.gameState.projectiles.filter(p => p.isAlive);
  }

  private checkWaveProgression(): void {
    const waveConfig = EnemySpawnService.createWaveConfig(this.gameState.wave);
    
    if (this.enemySpawnService.isWaveComplete(waveConfig) && 
        this.gameState.enemies.length === 0) {
      this.startNextWave();
    }
  }

  private startNextWave(): void {
    this.gameState.wave++;
    this.gameState.score += 50; // 웨이브 완료 보너스
    this.enemySpawnService.startNewWave();
    
    console.log(`Wave ${this.gameState.wave} started`);
  }

  // 플레이어 입력 처리
  movePlayer(direction: { x: number; y: number }): void {
    this.movementService.handlePlayerInput(this.gameState.player, direction);
    
    // 플레이어 위치를 물리 시스템에 동기화
    this.physicsService.updateEntityPosition(this.gameState.player.id, this.gameState.player.position);
  }

  // 게임 상태 접근자
  getGameState(): GameState {
    return {
      ...this.gameState,
      // 불변성을 위한 복사본 반환
      player: this.gameState.player,
      allies: [...this.gameState.allies],
      enemies: [...this.gameState.enemies],
      projectiles: [...this.gameState.projectiles]
    };
  }

  // 게임 제어
  pauseGame(): void {
    this.gameState.isPaused = true;
  }

  resumeGame(): void {
    this.gameState.isPaused = false;
  }

  stopGame(): void {
    this.gameState.isGameRunning = false;
  }

  resetGame(): void {
    this.initialize();
  }

  // 설정 변경
  setConversionRate(rate: number): void {
    this.conversionConfig.conversionRate = Math.max(0, Math.min(1, rate));
  }

  toggleConversion(): void {
    this.conversionConfig.enabled = !this.conversionConfig.enabled;
  }

  // 디버그 및 치트
  addTestEnemy(): void {
    const spawnPos = this.movementService.calculateEnemySpawnPosition(
      this.gameState.player.position
    );
    const enemy = Enemy.createNormal(`test-${Date.now()}`, spawnPos);
    this.gameState.enemies.push(enemy);
  }

  healPlayer(amount: number = 50): void {
    this.gameState.player.heal(amount);
  }

  // 통계 정보
  getGameStats(): {
    playTime: number;
    enemiesKilled: number;
    allyCount: number;
    currentWave: number;
    score: number;
  } {
    return {
      playTime: (Date.now() - this.lastUpdateTime) / 1000,
      enemiesKilled: Math.floor(this.gameState.score / 100), // 추정치
      allyCount: this.gameState.allies.length,
      currentWave: this.gameState.wave,
      score: this.gameState.score
    };
  }
}