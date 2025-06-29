import { Entity } from '../../types/minimalLegion';
import { IEnemyService } from './EnemyService';
import { IGameStateRepository } from '../../data/repositories/GameStateRepository';
import { EnemySpawnSystem } from '../../systems/enemySpawnSystem';
import { CombatSystem } from '../../systems/combatSystem';
import { MovementSystem } from '../../systems/movementSystem';
import { CollisionSystem } from '../../systems/collisionSystem';

export interface IGameEngineService {
  initialize(): void;
  update(deltaTime: number): void;
  getGameState(): {
    player: Entity | null;
    allies: Entity[];
    enemies: Entity[];
    projectiles: Entity[];
    camera: { x: number; y: number };
    wave: number;
    score: number;
  };
  movePlayer(direction: { x: number; y: number }): void;
}

export class GameEngineService implements IGameEngineService {
  private enemySpawnSystem: EnemySpawnSystem;
  private combatSystem: CombatSystem;
  private movementSystem: MovementSystem;
  private collisionSystem: CollisionSystem;

  constructor(
    private enemyService: IEnemyService,
    private gameStateRepository: IGameStateRepository
  ) {
    this.enemySpawnSystem = new EnemySpawnSystem();
    this.combatSystem = new CombatSystem();
    this.movementSystem = new MovementSystem();
    this.collisionSystem = new CollisionSystem();
  }

  initialize(): void {
    // 초기 플레이어 생성
    const initialPlayer: Entity = {
      id: 'player',
      position: { x: 0, y: 0 },
      velocity: { x: 0, y: 0 },
      health: 100,
      maxHealth: 100,
      damage: 15,
      speed: 5,
      size: 20,
      color: '#3B82F6',
      type: 'player'
    };

    this.gameStateRepository.setPlayer(initialPlayer);
    this.gameStateRepository.setCamera({ x: -600, y: -400 });
    this.gameStateRepository.setWave(1);
    this.gameStateRepository.setScore(0);
    this.gameStateRepository.setProjectiles([]); // 빈 투사체 배열로 시작
    this.enemySpawnSystem.startNewWave();
    
    // 테스트용 적 생성 (플레이어 근처)
    const testEnemy = this.enemyService.spawnEnemy(1, { x: 200, y: 0 }); // 플레이어로부터 200px 떨어진 곳
  }

  update(deltaTime: number): void {
    const player = this.gameStateRepository.getPlayer();
    if (!player) return;

    const wave = this.gameStateRepository.getWave();

    // 적 스폰 업데이트
    this.enemySpawnSystem.update(deltaTime, wave, player.position, (enemy) => {
      this.enemyService.spawnEnemy(wave, player.position);
    });

    // 게임 상태 가져오기
    const allies = this.gameStateRepository.getAllies();
    const projectiles = this.gameStateRepository.getProjectiles();
    
    // 모든 적을 가져와서 전투 시스템에 전달 (거리 제한 없음)
    const allEnemies = this.enemyService.getEnemiesInRange(player.position, 2000); // 넓은 범위
    

    // 전투 시스템 업데이트
    this.combatSystem.update(
      deltaTime,
      player,
      allies,
      allEnemies,
      (projectile) => {
        try {
          const currentProjectiles = this.gameStateRepository.getProjectiles() || [];
          const newProjectiles = [...currentProjectiles, projectile];
          this.gameStateRepository.setProjectiles(newProjectiles);
        } catch (error) {
          console.error('Error spawning projectile:', error);
        }
      }
    );

    // 이동 시스템 업데이트 (전투 시스템 후 최신 투사체 상태 사용)
    const currentProjectiles = this.gameStateRepository.getProjectiles();
    
    const { updatedAllies, updatedEnemies, updatedProjectiles } = this.movementSystem.update(
      deltaTime,
      player,
      allies,
      allEnemies,
      currentProjectiles
    );

    // 플레이어 위치 업데이트
    const newPlayerPosition = {
      x: player.position.x + player.velocity.x * player.speed * 60 * deltaTime,
      y: player.position.y + player.velocity.y * player.speed * 60 * deltaTime
    };

    const updatedPlayer = { ...player, position: newPlayerPosition };

    // 카메라 업데이트 (플레이어 중심)
    const newCamera = {
      x: newPlayerPosition.x - 600,
      y: newPlayerPosition.y - 400
    };

    // 상태 저장
    this.gameStateRepository.setPlayer(updatedPlayer);
    this.gameStateRepository.setAllies(updatedAllies);
    this.gameStateRepository.setProjectiles(updatedProjectiles);
    this.gameStateRepository.setCamera(newCamera);

    // 적 위치 업데이트
    updatedEnemies.forEach(enemy => {
      this.enemyService.updateEnemyPosition(enemy.id, enemy.position);
    });

    // 충돌 검사
    const collisionResult = this.collisionSystem.checkCollisions(
      updatedPlayer,
      updatedAllies,
      updatedEnemies,
      updatedProjectiles
    );

    // 충돌 결과 처리
    this.handleCollisionResults(collisionResult);
  }

  private handleCollisionResults(collisionResult: any): void {
    if (!collisionResult || !collisionResult.projectileHits) return;
    
    // 투사체 명중 처리
    collisionResult.projectileHits.forEach((hit: any) => {
      // 투사체 제거
      const projectiles = this.gameStateRepository.getProjectiles();
      const filteredProjectiles = projectiles.filter(p => p.id !== hit.projectileId);
      this.gameStateRepository.setProjectiles(filteredProjectiles);

      // 적 데미지 처리
      const enemy = this.enemyService.getEnemiesInRange({ x: 0, y: 0 }, Infinity)
        .find(e => e.id === hit.targetId);
      
      if (enemy) {
        const newHealth = enemy.health - hit.damage;
        this.enemyService.updateEnemyHealth(enemy.id, newHealth);
      }
    });
  }

  getGameState() {
    const camera = this.gameStateRepository.getCamera();
    const visibleEnemies = this.enemyService.getVisibleEnemies(camera, 1200, 800);
    const projectiles = this.gameStateRepository.getProjectiles();
    

    return {
      player: this.gameStateRepository.getPlayer(),
      allies: this.gameStateRepository.getAllies(),
      enemies: visibleEnemies, // 렌더링용으로는 보이는 적만
      projectiles,
      camera,
      wave: this.gameStateRepository.getWave(),
      score: this.gameStateRepository.getScore()
    };
  }

  movePlayer(direction: { x: number; y: number }): void {
    const player = this.gameStateRepository.getPlayer();
    if (player) {
      this.gameStateRepository.setPlayer({
        ...player,
        velocity: direction
      });
    }
  }
}