import type { Entity } from '../minimal-legion/types/minimalLegion';

export class GameScene extends Phaser.Scene {
  private gameEngineService: any; // GameEngineService 인터페이스 유지
  private lastUpdateTime: number = 0;
  
  // Phaser Game Objects 맵핑
  private entitySprites: Map<string, Phaser.GameObjects.Sprite> = new Map();
  private playerSprite: Phaser.GameObjects.Sprite | null = null;
  private background: Phaser.GameObjects.Graphics | null = null;
  
  // React와의 통신을 위한 이벤트 에미터
  private gameStateUpdateCallback?: (gameState: any) => void;

  constructor() {
    super({ key: 'GameScene' });
  }

  init(data: { gameEngineService: any, onGameStateUpdate?: (gameState: any) => void }) {
    this.gameEngineService = data.gameEngineService;
    this.gameStateUpdateCallback = data.onGameStateUpdate;
  }

  preload() {
    // 기본 색상으로 간단한 스프라이트 생성
    this.createColorSprites();
  }

  create() {
    // 배경 생성
    this.createBackground();
    
    // 게임 엔진 초기화 (기존 구조 유지)
    if (this.gameEngineService) {
      this.gameEngineService.initialize();
    }

    // 키보드 입력 설정
    this.setupInput();

    // 게임 루프 시작
    this.time.addEvent({
      delay: 16, // 60 FPS
      callback: this.gameUpdate,
      callbackScope: this,
      loop: true
    });

    console.log('Phaser GameScene created and initialized');
  }

  private createColorSprites() {
    // 플레이어용 스프라이트 (파란색 원)
    this.add.graphics()
      .fillStyle(0x3B82F6)
      .fillCircle(10, 10, 10)
      .generateTexture('player', 20, 20);

    // 적 스프라이트 (빨간색 원)
    this.add.graphics()
      .fillStyle(0xDC2626)
      .fillCircle(9, 9, 9)
      .generateTexture('enemy', 18, 18);

    // 아군 스프라이트 (초록색 원)
    this.add.graphics()
      .fillStyle(0x10B981)
      .fillCircle(8, 8, 8)
      .generateTexture('ally', 16, 16);

    // 투사체 스프라이트 (작은 노란색 원)
    this.add.graphics()
      .fillStyle(0xFBBF24)
      .fillCircle(2, 2, 2)
      .generateTexture('projectile', 4, 4);
  }

  private createBackground() {
    this.background = this.add.graphics();
    this.background.fillStyle(0x1a1a1a); // 어두운 배경
    this.background.fillRect(0, 0, this.scale.width, this.scale.height);
    
    // 격자 그리기
    this.background.lineStyle(1, 0x333333, 0.5);
    for (let x = 0; x < this.scale.width; x += 50) {
      this.background.lineBetween(x, 0, x, this.scale.height);
    }
    for (let y = 0; y < this.scale.height; y += 50) {
      this.background.lineBetween(0, y, this.scale.width, y);
    }
  }

  private setupInput() {
    // WASD 키 설정
    const cursors = this.input.keyboard?.createCursorKeys();
    const wasd = this.input.keyboard?.addKeys('W,S,A,D') as any;

    // 키 입력을 게임 엔진에 전달
    this.input.keyboard?.on('keydown', (event: KeyboardEvent) => {
      if (!this.gameEngineService) return;

      const direction = { x: 0, y: 0 };
      
      // WASD 또는 화살표 키 처리
      if (event.code === 'KeyW' || event.code === 'ArrowUp') direction.y = -1;
      if (event.code === 'KeyS' || event.code === 'ArrowDown') direction.y = 1;
      if (event.code === 'KeyA' || event.code === 'ArrowLeft') direction.x = -1;
      if (event.code === 'KeyD' || event.code === 'ArrowRight') direction.x = 1;

      if (direction.x !== 0 || direction.y !== 0) {
        this.gameEngineService.movePlayer(direction);
      }
    });
  }

  private gameUpdate() {
    if (!this.gameEngineService) return;

    const now = Date.now();
    const deltaTime = (now - this.lastUpdateTime) / 1000;
    this.lastUpdateTime = now;

    // 게임 엔진 업데이트 (기존 로직 유지)
    this.gameEngineService.update(deltaTime);
    
    // 게임 상태 가져오기
    const gameState = this.gameEngineService.getGameState();
    
    // Phaser 스프라이트들 업데이트
    this.updateSprites(gameState);
    
    // React UI에 게임 상태 전달
    if (this.gameStateUpdateCallback) {
      this.gameStateUpdateCallback(gameState);
    }
  }

  private updateSprites(gameState: any) {
    // 기존 스프라이트들 정리
    this.clearUnusedSprites(gameState);

    // 플레이어 업데이트
    if (gameState.player) {
      this.updatePlayerSprite(gameState.player);
    }

    // 적들 업데이트
    gameState.enemies.forEach((enemy: Entity) => {
      this.updateEntitySprite(enemy, 'enemy');
    });

    // 아군들 업데이트
    gameState.allies.forEach((ally: Entity) => {
      this.updateEntitySprite(ally, 'ally');
    });

    // 투사체들 업데이트
    gameState.projectiles.forEach((projectile: Entity) => {
      this.updateEntitySprite(projectile, 'projectile');
    });
  }

  private updatePlayerSprite(player: Entity) {
    if (!this.playerSprite) {
      this.playerSprite = this.add.sprite(
        player.position.x + this.scale.width / 2,
        player.position.y + this.scale.height / 2,
        'player'
      );
    }

    // 플레이어 위치 업데이트 (화면 중앙 기준)
    this.playerSprite.setPosition(
      player.position.x + this.scale.width / 2,
      player.position.y + this.scale.height / 2
    );
  }

  private updateEntitySprite(entity: Entity, textureKey: string) {
    let sprite = this.entitySprites.get(entity.id);

    if (!sprite) {
      // 새 스프라이트 생성
      sprite = this.add.sprite(
        entity.position.x + this.scale.width / 2,
        entity.position.y + this.scale.height / 2,
        textureKey
      );
      this.entitySprites.set(entity.id, sprite);
    }

    // 위치 업데이트
    sprite.setPosition(
      entity.position.x + this.scale.width / 2,
      entity.position.y + this.scale.height / 2
    );

    // 체력바 표시 (적과 아군만)
    if (entity.type === 'enemy' || entity.type === 'ally') {
      this.updateHealthBar(sprite, entity);
    }
  }

  private updateHealthBar(sprite: Phaser.GameObjects.Sprite, entity: Entity) {
    // 간단한 체력바 구현
    const healthPercentage = entity.health / entity.maxHealth;
    
    // 체력바 배경
    if (!sprite.getData('healthBarBg')) {
      const healthBarBg = this.add.graphics();
      healthBarBg.fillStyle(0x000000);
      healthBarBg.fillRect(-10, -15, 20, 3);
      sprite.setData('healthBarBg', healthBarBg);
    }

    // 체력바
    if (!sprite.getData('healthBar')) {
      const healthBar = this.add.graphics();
      sprite.setData('healthBar', healthBar);
    }

    const healthBar = sprite.getData('healthBar');
    healthBar.clear();
    healthBar.fillStyle(healthPercentage > 0.5 ? 0x10B981 : healthPercentage > 0.25 ? 0xF59E0B : 0xDC2626);
    healthBar.fillRect(
      sprite.x - 10,
      sprite.y - 15,
      20 * healthPercentage,
      3
    );
  }

  private clearUnusedSprites(gameState: any) {
    // 존재하지 않는 엔티티의 스프라이트 제거
    const currentEntityIds = new Set([
      ...gameState.enemies.map((e: Entity) => e.id),
      ...gameState.allies.map((a: Entity) => a.id),
      ...gameState.projectiles.map((p: Entity) => p.id)
    ]);

    this.entitySprites.forEach((sprite, id) => {
      if (!currentEntityIds.has(id)) {
        // 체력바도 함께 제거
        const healthBar = sprite.getData('healthBar');
        const healthBarBg = sprite.getData('healthBarBg');
        if (healthBar) healthBar.destroy();
        if (healthBarBg) healthBarBg.destroy();
        
        sprite.destroy();
        this.entitySprites.delete(id);
      }
    });
  }

  // React에서 호출할 수 있는 메서드들
  public getGameEngineService() {
    return this.gameEngineService;
  }

  public setGameEngineService(service: any) {
    this.gameEngineService = service;
  }
}