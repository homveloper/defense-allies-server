import Phaser from 'phaser';
import { Player } from '../entities/Player';
import { Enemy } from '../entities/Enemy';
import { Projectile } from '../entities/Projectile';
import { Ally } from '../entities/Ally';
import { RotatingOrb } from '../entities/RotatingOrb';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';

export class MainScene extends Phaser.Scene {
  private player!: Player;
  private enemies!: Phaser.Physics.Arcade.Group;
  private allies!: Phaser.Physics.Arcade.Group;
  private projectiles!: Phaser.Physics.Arcade.Group;
  private enemyProjectiles!: Phaser.Physics.Arcade.Group;
  private rotatingOrbs: RotatingOrb[] = [];
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private wasd!: {
    up: Phaser.Input.Keyboard.Key;
    down: Phaser.Input.Keyboard.Key;
    left: Phaser.Input.Keyboard.Key;
    right: Phaser.Input.Keyboard.Key;
  };
  private waveStartTime: number = 0;
  private enemySpawnTimer: number = 0;
  private hudText!: Phaser.GameObjects.Text;
  private debugInfo: Record<string, unknown> = {};
  private playerDamageTime: number = 0;
  private playerDamageCooldown: number = 1000; // 1초 쿨다운
  private combatCooldowns: Map<string, number> = new Map(); // 전투 쿨다운 관리
  private combatCooldownTime: number = 800; // 0.8초 쿨다운
  private gridGraphics!: Phaser.GameObjects.Graphics;

  constructor() {
    super({ key: 'MainScene' });
  }

  create() {
    console.log('MainScene create() called');
    
    // Reset game state
    const store = useMinimalLegionStore.getState();
    store.resetGame();
    
    // Create grid background
    this.createGridBackground();
    
    // Input setup
    this.cursors = this.input.keyboard!.createCursorKeys();
    this.wasd = {
      up: this.input.keyboard!.addKey('W'),
      down: this.input.keyboard!.addKey('S'),
      left: this.input.keyboard!.addKey('A'),
      right: this.input.keyboard!.addKey('D'),
    };

    // Groups
    this.enemies = this.physics.add.group({
      classType: Enemy,
      runChildUpdate: true,
    });

    this.allies = this.physics.add.group({
      classType: Ally,
      runChildUpdate: true,
    });

    this.projectiles = this.physics.add.group({
      classType: Projectile,
      runChildUpdate: true,
      maxSize: 100, // 최대 투사체 수 제한
    });

    this.enemyProjectiles = this.physics.add.group({
      classType: Projectile,
      runChildUpdate: true,
      maxSize: 200, // 적 투사체는 더 많이 허용
    });

    // Player
    console.log('Creating player...');
    this.player = new Player(this, 600, 400);
    console.log('Player created:', {
      active: this.player.active,
      visible: this.player.visible,
      position: { x: this.player.x, y: this.player.y },
      bodyExists: !!this.player.body
    });

    // Collisions
    this.physics.add.overlap(
      this.projectiles,
      this.enemies,
      this.handleProjectileEnemyCollision,
      undefined,
      this
    );

    this.physics.add.overlap(
      this.enemyProjectiles,
      this.player,
      this.handleEnemyProjectilePlayerCollision,
      undefined,
      this
    );

    this.physics.add.overlap(
      this.enemyProjectiles,
      this.allies,
      this.handleEnemyProjectileAllyCollision,
      undefined,
      this
    );
    
    // 캐릭터 간 충돌 처리 (겹치지 않게)
    this.physics.add.collider(this.enemies, this.enemies); // 적끼리 충돌
    this.physics.add.collider(this.allies, this.allies); // 아군끼리 충돌
    this.physics.add.collider(this.enemies, this.allies, this.handleEnemyAllyCollision, undefined, this); // 적과 아군 충돌
    this.physics.add.collider(this.player, this.enemies, this.handlePlayerEnemyCollision, undefined, this); // 플레이어와 적 충돌
    this.physics.add.collider(this.player, this.allies); // 플레이어와 아군 충돌
    
    // 회전 오브와 적들 충돌 처리는 update에서 수동으로 처리

    // HUD
    this.createHUD();

    // Start first wave
    this.startWave();
  }

  update(_time: number, delta: number) {
    // Check if scene is shutting down
    if (!this.scene.isActive() || this.scene.isPaused()) {
      return;
    }
    
    // Check if game is paused (for level up modal)
    const store = useMinimalLegionStore.getState();
    if (store.isPaused) {
      return;
    }
    
    // Debug: Check player state
    if (this.player) {
      this.debugInfo = {
        playerExists: !!this.player,
        playerActive: this.player.active,
        playerVisible: this.player.visible,
        playerX: this.player.x,
        playerY: this.player.y,
        playerBodyExists: !!this.player.body,
        playerBodyX: this.player.body?.x,
        playerBodyY: this.player.body?.y,
        inDisplayList: this.children.exists(this.player),
        sceneActive: this.scene.isActive(),
      };
      
      // Log warning if player is missing or has issues
      if (!this.player.active || !this.player.visible || !this.player.body) {
        console.warn('Player state issue:', this.debugInfo);
        
        // Try to recreate player if completely broken
        if (!this.player.scene || !this.children.exists(this.player)) {
          console.log('Recreating player...');
          this.recreatePlayer();
          return;
        }
      }
      
      // Check if player is out of bounds
      if (this.player.x < -100 || this.player.x > 1300 || this.player.y < -100 || this.player.y > 900) {
        console.warn('Player out of bounds:', { x: this.player.x, y: this.player.y });
        // Reset player position if too far out
        this.player.setPosition(600, 400);
      }
    } else {
      console.error('Player is null/undefined! Recreating...');
      this.recreatePlayer();
      return;
    }

    // Player movement
    const moveX =
      (this.cursors.right.isDown || this.wasd.right.isDown ? 1 : 0) -
      (this.cursors.left.isDown || this.wasd.left.isDown ? 1 : 0);
    const moveY =
      (this.cursors.down.isDown || this.wasd.down.isDown ? 1 : 0) -
      (this.cursors.up.isDown || this.wasd.up.isDown ? 1 : 0);

    if (this.player && this.player.active) {
      this.player.move(moveX, moveY);
      this.player.update(); // 플레이어 업데이트 호출
    }

    // Find nearest enemy for player
    const nearestEnemy = this.findNearestEnemy(this.player.x, this.player.y);
    if (nearestEnemy) {
      this.player.setTarget(nearestEnemy);
    }

    // Enemy spawning - 웨이브가 올라갈수록 더 빠르게 생성
    this.enemySpawnTimer += delta;
    const gameStore = useMinimalLegionStore.getState();
    const spawnRate = Math.max(800, 1800 - (gameStore.wave - 1) * 80); // 더 빠른 생성 (최소 0.8초)
    const maxEnemies = 20 + gameStore.wave * 3; // 더 많은 최대 적 수
    
    if (this.enemySpawnTimer > spawnRate && this.enemies.countActive() < maxEnemies) {
      this.spawnEnemy();
      this.enemySpawnTimer = 0;
    }
    
    // 회전 오브와 적 충돌 처리
    this.handleOrbEnemyCollisions();
    
    // 투사체 정리 (비활성화된 것들 제거)
    this.cleanupInactiveProjectiles();

    // Update HUD
    this.updateHUD();

    // Check wave completion
    const waveStore = useMinimalLegionStore.getState();
    if (waveStore.enemiesRemaining === 0 && this.enemies.countActive() === 0) {
      this.nextWave();
    }
  }

  private createGridBackground() {
    this.gridGraphics = this.add.graphics();
    this.gridGraphics.lineStyle(1, 0xe0e0e0, 0.3); // 연한 회색 그리드
    
    const gridSize = 50;
    const width = 1200;
    const height = 800;
    
    // 세로선 그리기
    for (let x = 0; x <= width; x += gridSize) {
      this.gridGraphics.moveTo(x, 0);
      this.gridGraphics.lineTo(x, height);
    }
    
    // 가로선 그리기
    for (let y = 0; y <= height; y += gridSize) {
      this.gridGraphics.moveTo(0, y);
      this.gridGraphics.lineTo(width, y);
    }
    
    this.gridGraphics.strokePath();
    
    // 그리드를 맨 뒤로 보내기
    this.gridGraphics.setDepth(-1);
  }

  private createHUD() {
    const style = {
      font: '16px Arial',
      fill: '#333333', // 어두운 회색으로 변경
    };

    this.hudText = this.add.text(10, 10, '', style);
    this.hudText.setScrollFactor(0);
  }

  private updateHUD() {
    const store = useMinimalLegionStore.getState();
    const hudInfo = [
      `Wave: ${store.wave}`,
      `Health: ${store.player.health}/${store.player.maxHealth}`,
      `Level: ${store.player.level}`,
      `EXP: ${store.player.experience}/${store.player.experienceToNext}`,
      `Allies: ${store.allies}/${store.maxAllies}`,
      `Score: ${store.score}`,
      `Enemies: ${this.enemies.countActive()}`,
    ];

    this.hudText.setText(hudInfo.join('\n'));
  }

  private startWave() {
    const store = useMinimalLegionStore.getState();
    // 웨이브별 적 수 증가: 기본 8마리에서 웨이브당 4마리씩 증가
    const enemyCount = 8 + (store.wave - 1) * 4;
    store.setEnemiesRemaining(enemyCount);
    this.waveStartTime = this.time.now;
    
    console.log(`Wave ${store.wave} started with ${enemyCount} enemies`);
  }

  private nextWave() {
    const store = useMinimalLegionStore.getState();
    
    // 웨이브 완료 보너스 경험치 및 점수
    const waveBonus = store.wave * 50;
    store.addExperience(waveBonus);
    store.updateScore(waveBonus);
    
    console.log(`Wave ${store.wave} completed! Bonus: ${waveBonus} XP & Score`);
    
    store.nextWave();
    this.time.delayedCall(3000, () => this.startWave());
  }

  private spawnEnemy() {
    const spawnStore = useMinimalLegionStore.getState();
    if (spawnStore.enemiesRemaining <= 0) return;

    const side = Phaser.Math.Between(0, 3);
    let x, y;

    switch (side) {
      case 0: // Top
        x = Phaser.Math.Between(0, 1200);
        y = -50;
        break;
      case 1: // Right
        x = 1250;
        y = Phaser.Math.Between(0, 800);
        break;
      case 2: // Bottom
        x = Phaser.Math.Between(0, 1200);
        y = 850;
        break;
      default: // Left
        x = -50;
        y = Phaser.Math.Between(0, 800);
    }

    const enemyStore = useMinimalLegionStore.getState();
    const enemy = new Enemy(this, x, y, enemyStore.wave);
    this.enemies.add(enemy);
    enemy.setTarget(this.player);

    spawnStore.setEnemiesRemaining(spawnStore.enemiesRemaining - 1);
  }

  private findNearestEnemy(x: number, y: number): Enemy | null {
    let nearest: Enemy | null = null;
    let nearestDistance = Infinity;

    this.enemies.children.entries.forEach((enemy) => {
      if (enemy.active) {
        const distance = Phaser.Math.Distance.Between(
          x,
          y,
          enemy.body!.position.x,
          enemy.body!.position.y
        );
        if (distance < nearestDistance) {
          nearestDistance = distance;
          nearest = enemy as Enemy;
        }
      }
    });

    return nearest;
  }

  private handleProjectileEnemyCollision(
    projectile: object,
    enemy: object
  ) {
    const proj = projectile as Projectile;
    const en = enemy as Enemy;

    if (!proj.active || !en.active) return;

    en.takeDamage(proj.damage);
    
    // 투사체 완전 제거
    this.projectiles.remove(proj, true, true);
    
    if (en.health <= 0) {
      this.convertEnemyToAlly(en);
    }
  }

  private handleEnemyProjectilePlayerCollision(
    projectile: object
  ) {
    const proj = projectile as Projectile;
    
    if (!proj.active) return;
    
    const damage = proj.damage || 0;
    
    if (damage > 0 && this.player && this.player.active) {
      this.player.takeDamage(damage);
    }
    
    // 투사체 완전 제거
    this.enemyProjectiles.remove(proj, true, true);
  }

  private handleEnemyProjectileAllyCollision(
    projectile: object,
    ally: object
  ) {
    const proj = projectile as Projectile;
    const al = ally as Ally;

    if (!proj.active || !al.active) return;

    al.takeDamage(proj.damage);
    
    // 투사체 완전 제거
    this.enemyProjectiles.remove(proj, true, true);

    if (al.health <= 0) {
      const store = useMinimalLegionStore.getState();
      store.removeAlly();
    }
  }

  private convertEnemyToAlly(enemy: Enemy) {
    const store = useMinimalLegionStore.getState();
    
    // Add experience
    store.addExperience(10);
    store.updateScore(10);

    // Check if we can add more allies
    if (store.allies < store.maxAllies) {
      // 적에서 아군으로 전환 시 스펙 낮춤 (70%)
      const ally = new Ally(this, enemy.x, enemy.y, {
        healthMultiplier: 0.7,
        attackMultiplier: 0.7,
        speedMultiplier: 0.8
      });
      ally.setPlayer(this.player);
      this.allies.add(ally);
      store.addAlly();
      
      console.log('Enemy converted to ally with reduced stats (70% health/attack, 80% speed)');
    }

    enemy.destroy();
  }

  fireProjectile(x: number, y: number, targetX: number, targetY: number, damage: number, isEnemy: boolean = false) {
    console.log(`Creating projectile at (${x}, ${y}) targeting (${targetX}, ${targetY})`);
    
    const projectile = new Projectile(this, x, y);
    
    // Add to appropriate group first
    if (isEnemy) {
      this.enemyProjectiles.add(projectile);
    } else {
      this.projectiles.add(projectile);
    }
    
    // Fire immediately - physics body should be initialized by now
    if (projectile.active && projectile.body) {
      projectile.fire(targetX, targetY, damage);
    } else {
      // If body isn't ready, use next frame
      this.time.delayedCall(16, () => { // One frame at 60fps
        if (projectile.active && projectile.body) {
          projectile.fire(targetX, targetY, damage);
        } else {
          console.error('Projectile body still not ready after one frame');
        }
      });
    }
  }

  // 플레이어와 적 충돌 처리
  private handlePlayerEnemyCollision(
    _player: object,
    enemy: object
  ) {
    const en = enemy as Enemy;
    const pl = this.player;
    const now = this.time.now;
    
    // 쿨다운 체크
    if (now - this.playerDamageTime < this.playerDamageCooldown) {
      return; // 아직 쿨다운 중이면 데미지 없음
    }
    
    if (en && pl && en.active && pl.active) {
      // 적이 플레이어에게 데미지
      const damage = 15; // 적의 기본 접촉 데미지
      pl.takeDamage(damage);
      this.playerDamageTime = now; // 쿨다운 시작
      
      console.log(`Player took ${damage} damage from enemy collision (cooldown started)`);
      
      // 적을 약간 뒤로 밀어내기
      const angle = Phaser.Math.Angle.Between(pl.x, pl.y, en.x, en.y);
      const pushForce = 150;
      
      if (en.body) {
        en.body.setVelocity(
          Math.cos(angle) * pushForce,
          Math.sin(angle) * pushForce
        );
      }
      
      // 플레이어도 약간 밀려나게 하기
      if (pl.body) {
        pl.body.setVelocity(
          Math.cos(angle + Math.PI) * 80,
          Math.sin(angle + Math.PI) * 80
        );
      }
    }
  }

  // 적과 아군 충돌 처리
  private handleEnemyAllyCollision(
    enemy: object,
    ally: object
  ) {
    const en = enemy as Enemy;
    const al = ally as Ally;
    const now = this.time.now;
    
    if (!en || !al || !en.active || !al.active) return;
    
    // 각 오브젝트에 대한 고유 ID 생성 (충돌 쿨다운용)
    const enemyId = `enemy_${en.x}_${en.y}_${Math.floor(now / 1000)}`;
    const allyId = `ally_${al.x}_${al.y}_${Math.floor(now / 1000)}`;
    const combatKey = `${enemyId}_vs_${allyId}`;
    
    // 쿨다운 체크
    const lastCombatTime = this.combatCooldowns.get(combatKey) || 0;
    if (now - lastCombatTime < this.combatCooldownTime) {
      return; // 아직 쿨다운 중
    }
    
    // 상호 데미지
    const enemyDamage = 8;
    const allyDamage = 10;
    
    en.takeDamage(allyDamage);
    al.takeDamage(enemyDamage);
    
    this.combatCooldowns.set(combatKey, now);
    
    console.log(`Combat: Enemy took ${allyDamage} damage, Ally took ${enemyDamage} damage`);
    
    // 적이 죽었는지 확인
    if (en.health <= 0) {
      this.convertEnemyToAlly(en);
    }
    
    // 아군이 죽었는지 확인
    if (al.health <= 0) {
      const store = useMinimalLegionStore.getState();
      store.removeAlly();
      al.destroy();
    }
    
    // 상호 밀어내기
    const angle = Phaser.Math.Angle.Between(en.x, en.y, al.x, al.y);
    const pushForce = 120;
    
    if (en.body) {
      en.body.setVelocity(
        Math.cos(angle + Math.PI) * pushForce,
        Math.sin(angle + Math.PI) * pushForce
      );
    }
    
    if (al.body) {
      al.body.setVelocity(
        Math.cos(angle) * pushForce,
        Math.sin(angle) * pushForce
      );
    }
  }

  dealMeleeDamage(target: Phaser.GameObjects.GameObject, damage: number) {
    if (!damage || isNaN(damage) || damage <= 0) {
      console.warn('Invalid melee damage:', damage);
      return;
    }
    
    if (target === this.player && this.player && this.player.active) {
      console.log('Player taking melee damage:', damage, 'at position:', this.player.x, this.player.y);
      this.player.takeDamage(damage);
      
      // Debug: Check player state after damage
      setTimeout(() => {
        if (!this.player || !this.player.active) {
          console.error('Player became inactive after taking damage!');
          console.log('Debug info:', this.debugInfo);
        }
      }, 10);
    } else if (this.allies.children.entries.includes(target)) {
      const ally = target as Ally;
      if (ally && ally.active) {
        ally.takeDamage(damage);
        if (ally.health <= 0) {
          const store = useMinimalLegionStore.getState();
          store.removeAlly();
        }
      }
    }
  }

  // Method to recreate player if it gets destroyed
  private recreatePlayer() {
    try {
      console.log('Attempting to recreate player...');
      
      // Remove old player if it exists
      if (this.player) {
        this.player.destroy();
      }
      
      // Create new player
      this.player = new Player(this, 600, 400);
      
      console.log('Player recreated successfully:', {
        active: this.player.active,
        visible: this.player.visible,
        position: { x: this.player.x, y: this.player.y },
        bodyExists: !!this.player.body
      });
      
    } catch (error) {
      console.error('Failed to recreate player:', error);
    }
  }
  
  // Method to get debug info for React component
  getDebugInfo() {
    return {
      player: this.player,
      enemies: this.enemies?.children?.entries || [],
      allies: this.allies?.children?.entries || [],
      rotatingOrbs: this.rotatingOrbs || [],
      debugInfo: this.debugInfo,
      sceneActive: this.scene?.isActive() || false,
    };
  }
  
  // 회전 오브 추가
  addRotatingOrb(orb: RotatingOrb) {
    this.rotatingOrbs.push(orb);
    console.log(`Rotating orb added. Total orbs: ${this.rotatingOrbs.length}`);
  }
  
  // 회전 오브와 적 충돌 처리
  private handleOrbEnemyCollisions() {
    this.rotatingOrbs.forEach(orb => {
      if (!orb || !orb.active) return;
      
      this.enemies.children.entries.forEach(enemy => {
        if (!enemy.active) return;
        
        const distance = Phaser.Math.Distance.Between(
          orb.x, orb.y,
          enemy.body!.position.x, enemy.body!.position.y
        );
        
        if (distance < 25) { // 충돌 거리
          orb.hitEnemy(enemy as { takeDamage?: (amount: number) => void });
        }
      });
    });
    
    // 비활성화된 오브들 제거
    this.rotatingOrbs = this.rotatingOrbs.filter(orb => orb && orb.active);
    
    // 오래된 쿨다운 정리 (5초 이상 된 것들 제거)
    const currentTime = this.time.now;
    const cleanupTime = 5000;
    for (const [key, time] of this.combatCooldowns.entries()) {
      if (currentTime - time > cleanupTime) {
        this.combatCooldowns.delete(key);
      }
    }
  }
  
  // 비활성화된 투사체 정리
  private cleanupInactiveProjectiles() {
    // 플레이어 투사체 정리
    this.projectiles.children.entries.forEach(projectile => {
      if (!projectile.active || !projectile.body) {
        this.projectiles.remove(projectile, true, true);
      }
    });
    
    // 적 투사체 정리
    this.enemyProjectiles.children.entries.forEach(projectile => {
      if (!projectile.active || !projectile.body) {
        this.enemyProjectiles.remove(projectile, true, true);
      }
    });
  }
  
  // Custom cleanup method
  cleanupScene() {
    console.log('MainScene cleanup called');
    this.rotatingOrbs.forEach(orb => {
      if (orb && orb.active) {
        orb.destroy();
      }
    });
    this.rotatingOrbs = [];
    
    // 모든 투사체 정리
    this.projectiles.clear(true, true);
    this.enemyProjectiles.clear(true, true);
  }
}