import * as Phaser from 'phaser';
import { ArenaPlayer } from '../entities/ArenaPlayer';
import { ArenaEnemy } from '../entities/ArenaEnemy';
import { PowerUp } from '../entities/PowerUp';
import { useAbilityArenaStore } from '@/store/abilityArenaStore';

export class ArenaMainScene extends Phaser.Scene {
  private player!: ArenaPlayer;
  private enemies!: Phaser.Physics.Arcade.Group;
  private powerUps!: Phaser.Physics.Arcade.Group;
  private projectiles!: Phaser.Physics.Arcade.Group;
  private enemyProjectiles!: Phaser.Physics.Arcade.Group;
  
  // Input
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private wasd!: {
    up: Phaser.Input.Keyboard.Key;
    down: Phaser.Input.Keyboard.Key;
    left: Phaser.Input.Keyboard.Key;
    right: Phaser.Input.Keyboard.Key;
  };
  private spaceKey!: Phaser.Input.Keyboard.Key;
  private qKey!: Phaser.Input.Keyboard.Key;
  private eKey!: Phaser.Input.Keyboard.Key;
  private rKey!: Phaser.Input.Keyboard.Key;

  // Game state
  private gameStartTime: number = 0;
  private lastWaveTime: number = 0;
  private waveInterval: number = 30000; // 30 seconds between waves
  private enemySpawnTimer: number = 0;
  private enemySpawnRate: number = 2000; // 2 seconds
  private powerUpSpawnTimer: number = 0;
  private powerUpSpawnRate: number = 15000; // 15 seconds

  // Arena boundaries
  private readonly ARENA_WIDTH = 1000;
  private readonly ARENA_HEIGHT = 600;
  private readonly ARENA_X = 100;
  private readonly ARENA_Y = 100;

  constructor() {
    super({ key: 'ArenaMainScene' });
  }

  create() {
    console.log('ArenaMainScene created');
    
    // Register globally for dev panel access
    if (typeof window !== 'undefined') {
      (window as any).currentArenaScene = this;
      (window as any).phaserGame = this.game;
    }
    
    // Set world bounds to arena size
    this.physics.world.setBounds(
      this.ARENA_X, 
      this.ARENA_Y, 
      this.ARENA_WIDTH, 
      this.ARENA_HEIGHT
    );

    // Create arena background
    this.createArenaBackground();

    // Initialize input
    this.setupInput();

    // Create groups
    this.createGroups();

    // Create player
    this.createPlayer();

    // Setup collisions
    this.setupCollisions();

    // Start the game
    this.startGame();

    // Setup UI updates
    this.setupUIUpdates();
  }

  private createArenaBackground(): void {
    // Arena floor
    const floor = this.add.graphics();
    floor.fillStyle(0x1a1a1a);
    floor.fillRect(this.ARENA_X, this.ARENA_Y, this.ARENA_WIDTH, this.ARENA_HEIGHT);

    // Arena walls
    const walls = this.add.graphics();
    walls.lineStyle(4, 0x444444);
    walls.strokeRect(this.ARENA_X, this.ARENA_Y, this.ARENA_WIDTH, this.ARENA_HEIGHT);

    // Grid pattern
    const grid = this.add.graphics();
    grid.lineStyle(1, 0x333333, 0.3);
    
    const gridSize = 50;
    for (let x = this.ARENA_X; x <= this.ARENA_X + this.ARENA_WIDTH; x += gridSize) {
      grid.moveTo(x, this.ARENA_Y);
      grid.lineTo(x, this.ARENA_Y + this.ARENA_HEIGHT);
    }
    for (let y = this.ARENA_Y; y <= this.ARENA_Y + this.ARENA_HEIGHT; y += gridSize) {
      grid.moveTo(this.ARENA_X, y);
      grid.lineTo(this.ARENA_X + this.ARENA_WIDTH, y);
    }
    grid.strokePath();

    // Center marker
    const centerX = this.ARENA_X + this.ARENA_WIDTH / 2;
    const centerY = this.ARENA_Y + this.ARENA_HEIGHT / 2;
    const center = this.add.graphics();
    center.lineStyle(2, 0x666666);
    center.strokeCircle(centerX, centerY, 20);
    center.moveTo(centerX - 10, centerY);
    center.lineTo(centerX + 10, centerY);
    center.moveTo(centerX, centerY - 10);
    center.lineTo(centerX, centerY + 10);
    center.strokePath();
  }

  private setupInput(): void {
    this.cursors = this.input.keyboard!.createCursorKeys();
    this.wasd = {
      up: this.input.keyboard!.addKey('W'),
      down: this.input.keyboard!.addKey('S'),
      left: this.input.keyboard!.addKey('A'),
      right: this.input.keyboard!.addKey('D'),
    };
    this.spaceKey = this.input.keyboard!.addKey('SPACE');
    this.qKey = this.input.keyboard!.addKey('Q');
    this.eKey = this.input.keyboard!.addKey('E');
    this.rKey = this.input.keyboard!.addKey('R');

    // Mouse input for targeting
    this.input.on('pointerdown', (pointer: Phaser.Input.Pointer) => {
      if (pointer.leftButtonDown() && this.player && this.player.active) {
        this.player.handleLeftClick(pointer.worldX, pointer.worldY);
      }
      if (pointer.rightButtonDown() && this.player && this.player.active) {
        this.player.handleRightClick(pointer.worldX, pointer.worldY);
      }
    });
  }

  private createGroups(): void {
    this.enemies = this.physics.add.group({
      classType: ArenaEnemy,
      runChildUpdate: true,
      maxSize: 50
    });

    this.powerUps = this.physics.add.group({
      classType: PowerUp,
      runChildUpdate: true,
      maxSize: 10
    });

    this.projectiles = this.physics.add.group({
      runChildUpdate: true,
      maxSize: 100
    });

    this.enemyProjectiles = this.physics.add.group({
      runChildUpdate: true,
      maxSize: 200
    });
  }

  private createPlayer(): void {
    const centerX = this.ARENA_X + this.ARENA_WIDTH / 2;
    const centerY = this.ARENA_Y + this.ARENA_HEIGHT / 2;
    
    this.player = new ArenaPlayer(this, centerX, centerY);
  }

  private setupCollisions(): void {
    // Player projectiles vs enemies
    this.physics.add.overlap(
      this.projectiles,
      this.enemies,
      this.handleProjectileEnemyCollision,
      undefined,
      this
    );

    // Enemy projectiles vs player
    this.physics.add.overlap(
      this.enemyProjectiles,
      this.player,
      this.handleEnemyProjectilePlayerCollision,
      undefined,
      this
    );

    // Player vs power-ups
    this.physics.add.overlap(
      this.player,
      this.powerUps,
      this.handlePlayerPowerUpCollision,
      undefined,
      this
    );

    // Player vs enemies (damage collision)
    this.physics.add.overlap(
      this.player,
      this.enemies,
      this.handlePlayerEnemyCollision,
      undefined,
      this
    );
  }

  private startGame(): void {
    this.gameStartTime = this.time.now;
    this.lastWaveTime = this.time.now;
    
    const store = useAbilityArenaStore.getState();
    store.startGame();
    
    console.log('Arena game started!');
  }

  private setupUIUpdates(): void {
    // Update time alive every second
    this.time.addEvent({
      delay: 1000,
      callback: () => {
        const timeAlive = Math.floor((this.time.now - this.gameStartTime) / 1000);
        const store = useAbilityArenaStore.getState();
        store.updateTimeAlive(timeAlive);
      },
      loop: true
    });
  }

  update(_time: number, delta: number): void {
    const store = useAbilityArenaStore.getState();
    
    // Don't update if game is paused or over
    if (store.isPaused || store.isGameOver) {
      return;
    }

    // Update player
    if (this.player && this.player.active) {
      this.updatePlayerInput();
      this.player.update();
      
      // Check if player died
      if (this.player.health <= 0) {
        this.gameOver();
        return;
      }
    }

    // Update timers
    this.enemySpawnTimer += delta;
    this.powerUpSpawnTimer += delta;

    // Spawn enemies (check dev settings)
    let spawnEnabled = true;
    let spawnRateMultiplier = 1.0;
    if (typeof window !== 'undefined') {
      const devSettings = (window as any).devSettings;
      if (devSettings) {
        spawnEnabled = devSettings.enemySpawnEnabled !== false;
        spawnRateMultiplier = devSettings.enemySpawnRate || 1.0;
      }
    }
    
    if (spawnEnabled && this.enemySpawnTimer >= (this.enemySpawnRate / spawnRateMultiplier)) {
      this.spawnEnemy();
      this.enemySpawnTimer = 0;
    }

    // Spawn power-ups
    if (this.powerUpSpawnTimer >= this.powerUpSpawnRate) {
      this.spawnPowerUp();
      this.powerUpSpawnTimer = 0;
    }

    // Check for wave progression
    if (this.time.now - this.lastWaveTime >= this.waveInterval) {
      this.nextWave();
    }

    // Update difficulty based on time
    this.updateDifficulty();
  }

  private updatePlayerInput(): void {
    if (!this.player || !this.player.active) return;

    // Movement
    const moveX = 
      (this.cursors.right.isDown || this.wasd.right.isDown ? 1 : 0) -
      (this.cursors.left.isDown || this.wasd.left.isDown ? 1 : 0);
    const moveY = 
      (this.cursors.down.isDown || this.wasd.down.isDown ? 1 : 0) -
      (this.cursors.up.isDown || this.wasd.up.isDown ? 1 : 0);

    this.player.setMovement(moveX, moveY);

    // Abilities
    if (Phaser.Input.Keyboard.JustDown(this.spaceKey)) {
      this.player.useDash();
    }
    if (Phaser.Input.Keyboard.JustDown(this.qKey)) {
      this.player.useAbility('Q');
    }
    if (Phaser.Input.Keyboard.JustDown(this.eKey)) {
      this.player.useAbility('E');
    }
    if (Phaser.Input.Keyboard.JustDown(this.rKey)) {
      this.player.useAbility('R');
    }
  }

  private spawnEnemy(): void {
    // Check dev settings for max enemy count
    let maxEnemies = 20;
    if (typeof window !== 'undefined') {
      const devSettings = (window as any).devSettings;
      if (devSettings?.maxEnemyCount) {
        maxEnemies = devSettings.maxEnemyCount;
      }
    }
    
    if (this.enemies.countActive() >= maxEnemies) return; // Limit active enemies

    // Random spawn position at arena edges
    const side = Math.floor(Math.random() * 4);
    let x: number, y: number;

    switch (side) {
      case 0: // Top
        x = this.ARENA_X + Math.random() * this.ARENA_WIDTH;
        y = this.ARENA_Y;
        break;
      case 1: // Right
        x = this.ARENA_X + this.ARENA_WIDTH;
        y = this.ARENA_Y + Math.random() * this.ARENA_HEIGHT;
        break;
      case 2: // Bottom
        x = this.ARENA_X + Math.random() * this.ARENA_WIDTH;
        y = this.ARENA_Y + this.ARENA_HEIGHT;
        break;
      case 3: // Left
        x = this.ARENA_X;
        y = this.ARENA_Y + Math.random() * this.ARENA_HEIGHT;
        break;
      default:
        x = this.ARENA_X + this.ARENA_WIDTH / 2;
        y = this.ARENA_Y + this.ARENA_HEIGHT / 2;
    }

    const enemy = new ArenaEnemy(this, x, y, this.getEnemyType());
    this.enemies.add(enemy);
    enemy.setTarget(this.player);
  }

  // Dev Panel utility methods
  public spawnWave(): void {
    const waveSize = Math.min(10, 50 - this.enemies.countActive());
    for (let i = 0; i < waveSize; i++) {
      this.time.delayedCall(i * 100, () => {
        this.spawnEnemy();
      });
    }
  }

  public killAllEnemies(): void {
    this.enemies.children.entries.forEach((enemy: any) => {
      if (enemy.active && enemy.takeDamage) {
        enemy.takeDamage(9999);
      }
    });
  }

  public setSpawnEnabled(enabled: boolean): void {
    // This is handled by dev settings in update loop
    console.log(`Enemy spawning ${enabled ? 'enabled' : 'disabled'}`);
  }

  public setSpawnRate(rate: number): void {
    // This is handled by dev settings in update loop
    console.log(`Enemy spawn rate set to ${rate}x`);
  }

  public setMaxEnemyCount(count: number): void {
    // This is handled by dev settings in update loop
    console.log(`Max enemy count set to ${count}`);
  }

  private getEnemyType(): string {
    const store = useAbilityArenaStore.getState();
    const wave = store.stats.waveNumber;
    
    // Simple enemy type selection based on wave
    if (wave < 3) return 'grunt';
    if (wave < 6) return Math.random() < 0.7 ? 'grunt' : 'archer';
    if (wave < 10) {
      const rand = Math.random();
      if (rand < 0.4) return 'grunt';
      if (rand < 0.7) return 'archer';
      return 'mage';
    }
    
    // Higher waves - more variety
    const types = ['grunt', 'archer', 'mage', 'tank'];
    return types[Math.floor(Math.random() * types.length)];
  }

  private spawnPowerUp(): void {
    if (this.powerUps.countActive() >= 3) return; // Limit active power-ups

    const x = this.ARENA_X + 50 + Math.random() * (this.ARENA_WIDTH - 100);
    const y = this.ARENA_Y + 50 + Math.random() * (this.ARENA_HEIGHT - 100);

    const types = ['health', 'mana', 'damage', 'speed', 'shield'];
    const type = types[Math.floor(Math.random() * types.length)];

    const powerUp = new PowerUp(this, x, y, type);
    this.powerUps.add(powerUp);
  }

  private nextWave(): void {
    const store = useAbilityArenaStore.getState();
    store.incrementWave();
    store.addScore(100 * store.stats.waveNumber);
    
    this.lastWaveTime = this.time.now;
    
    // Increase spawn rate
    this.enemySpawnRate = Math.max(500, this.enemySpawnRate - 100);
    
    console.log(`Wave ${store.stats.waveNumber} started!`);
  }

  private updateDifficulty(): void {
    // Increase difficulty over time
    const timeAlive = (this.time.now - this.gameStartTime) / 1000;
    
    // Every 30 seconds, slightly increase spawn rate
    if (timeAlive > 0 && timeAlive % 30 < 1) {
      this.enemySpawnRate = Math.max(300, this.enemySpawnRate - 50);
    }
  }

  private handleProjectileEnemyCollision(projectile: any, enemy: any): void {
    if (!projectile.active || !enemy.active) return;

    const damage = projectile.damage || 20;
    enemy.takeDamage(damage);

    const store = useAbilityArenaStore.getState();
    store.addDamageDealt(damage);

    projectile.destroy();

    if (enemy.health <= 0) {
      this.enemyKilled(enemy);
    }
  }

  private handleEnemyProjectilePlayerCollision(projectile: any): void {
    if (!projectile.active || !this.player.active) return;

    const damage = projectile.damage || 10;
    this.player.takeDamage(damage);

    const store = useAbilityArenaStore.getState();
    store.addDamageTaken(damage);
    store.updatePlayerHealth(this.player.health);

    projectile.destroy();
  }

  private handlePlayerPowerUpCollision(_player: any, powerUp: any): void {
    if (!powerUp.active) return;

    powerUp.collect(this.player);
    
    const store = useAbilityArenaStore.getState();
    store.incrementPowerUps();
    store.addScore(50);
  }

  private handlePlayerEnemyCollision(_player: any, enemy: any): void {
    if (!enemy.active || !this.player.active) return;

    // Only damage if enough time has passed (avoid rapid damage)
    const now = this.time.now;
    if (!enemy.lastPlayerDamage || now - enemy.lastPlayerDamage > 1000) {
      const damage = enemy.contactDamage || 15;
      this.player.takeDamage(damage);

      const store = useAbilityArenaStore.getState();
      store.addDamageTaken(damage);
      store.updatePlayerHealth(this.player.health);

      enemy.lastPlayerDamage = now;
    }
  }

  private enemyKilled(enemy: any): void {
    const store = useAbilityArenaStore.getState();
    
    // Award experience and score
    const expReward = enemy.expReward || 25;
    const scoreReward = enemy.scoreReward || 10;
    
    store.addExperience(expReward);
    store.incrementKills();
    store.addScore(scoreReward);

    // Chance to drop power-up
    if (Math.random() < 0.1) { // 10% chance
      this.spawnPowerUp();
    }

    enemy.destroy();
  }

  private gameOver(): void {
    const store = useAbilityArenaStore.getState();
    store.endGame();
    
    console.log('Game Over!');
    
    // Stop all movement and effects
    this.physics.pause();
  }

  // Public methods for external access
  public getPlayer(): ArenaPlayer {
    return this.player;
  }

  public createProjectile(x: number, y: number, targetX: number, targetY: number, damage: number, isEnemy: boolean = false): void {
    // Create simple circle projectile texture
    const graphics = this.add.graphics();
    graphics.fillStyle(isEnemy ? 0xff4444 : 0x44ff44);
    graphics.fillCircle(0, 0, 3);
    
    const textureKey = 'projectile_' + Date.now();
    graphics.generateTexture(textureKey, 6, 6);
    graphics.destroy();
    
    const projectile = this.physics.add.sprite(x, y, textureKey);

    (projectile as any).damage = damage;

    // Calculate velocity
    const angle = Phaser.Math.Angle.Between(x, y, targetX, targetY);
    const speed = 400;
    const velocityX = Math.cos(angle) * speed;
    const velocityY = Math.sin(angle) * speed;

    projectile.setVelocity(velocityX, velocityY);

    // Add to appropriate group
    if (isEnemy) {
      this.enemyProjectiles.add(projectile);
    } else {
      this.projectiles.add(projectile);
    }

    // Destroy after 3 seconds
    this.time.delayedCall(3000, () => {
      if (projectile.active) {
        projectile.destroy();
      }
    });
  }
}