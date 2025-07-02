import Phaser from 'phaser';
import { MainScene } from '../scenes/MainScene';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';
import { RotatingOrb } from './RotatingOrb';

export class Player extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  private sprite: Phaser.GameObjects.Graphics;
  private healthBar: Phaser.GameObjects.Graphics;
  private moveSpeed: number;
  private target: Phaser.GameObjects.GameObject | null = null;
  private lastFireTime: number = 0;
  private fireRate: number = 500; // milliseconds between shots
  private rotatingOrbs: RotatingOrb[] = [];
  private orbSpawnTime: number = 0;

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y);

    // Create player sprite (blue circle)
    this.sprite = scene.add.graphics();
    this.sprite.fillStyle(0x3498db);
    this.sprite.fillCircle(0, 0, 20);
    this.add(this.sprite);

    // Create health bar
    this.healthBar = scene.add.graphics();
    this.add(this.healthBar);
    this.updateHealthBar();

    // Add to scene first
    scene.add.existing(this);

    // Then enable physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    this.body.setCircle(20);

    // Get initial stats from store
    const store = useMinimalLegionStore.getState();
    this.moveSpeed = store.player.moveSpeed;
    
    // 게임 시작 3초 후 첫 번째 오브 생성
    this.orbSpawnTime = 3000; // 상대적 시간으로 설정
    
    console.log('Player created with rotating orb ability');
  }

  move(x: number, y: number) {
    if (!this.body) {
      console.warn('Player move called but body is missing!');
      return;
    }
    
    if (!this.active) {
      console.warn('Player move called but player is inactive!');
      return;
    }
    
    // Check if game is over - don't move if dead
    const store = useMinimalLegionStore.getState();
    if (store.isGameOver) {
      this.body.setVelocity(0, 0);
      return;
    }
    
    const length = Math.sqrt(x * x + y * y);
    if (length > 0) {
      this.body.setVelocity(
        (x / length) * this.moveSpeed * 60,
        (y / length) * this.moveSpeed * 60
      );
    } else {
      this.body.setVelocity(0, 0);
    }
    
    // Debug: Log position changes if player moves too far
    if (Math.abs(this.x) > 2000 || Math.abs(this.y) > 2000) {
      console.error('Player moved too far!', {
        position: { x: this.x, y: this.y },
        velocity: { x: this.body.velocity.x, y: this.body.velocity.y },
        moveInput: { x, y }
      });
    }
  }

  setTarget(target: Phaser.GameObjects.GameObject | null) {
    this.target = target;
  }

  takeDamage(amount: number) {
    if (!amount || isNaN(amount) || amount < 0) {
      console.warn('Invalid damage amount:', amount);
      return;
    }

    console.log('Player takeDamage called:', {
      amount,
      playerActive: this.active,
      playerVisible: this.visible,
      playerPosition: { x: this.x, y: this.y },
      bodyExists: !!this.body,
      sceneExists: !!this.scene
    });

    const store = useMinimalLegionStore.getState();
    const currentHealth = store.player.health || 0;
    const newHealth = Math.max(0, currentHealth - amount);
    
    console.log(`Player taking ${amount} damage. Health: ${currentHealth} -> ${newHealth}`);
    
    store.updatePlayer({ health: newHealth });
    this.updateHealthBar();

    if (newHealth <= 0) {
      console.log('Player defeated! Calling gameOver');
      this.handlePlayerDefeat();
    }
  }

  private handlePlayerDefeat() {
    console.log('Player defeat - current state:', {
      active: this.active,
      visible: this.visible,
      position: { x: this.x, y: this.y },
      bodyExists: !!this.body
    });
    
    const store = useMinimalLegionStore.getState();
    store.gameOver();
    
    // Keep player object active but stop movement
    this.body?.setVelocity(0, 0);
    
    // 모든 오브 제거
    this.rotatingOrbs.forEach(orb => {
      if (orb && orb.active) {
        orb.destroy();
      }
    });
    this.rotatingOrbs = [];
    
    // Visual feedback for death
    this.sprite.clear();
    this.sprite.fillStyle(0x666666);
    this.sprite.fillCircle(0, 0, 20);
  }

  update() {
    // 기본 공격
    if (this.target && this.target.active) {
      const store = useMinimalLegionStore.getState();
      const distance = Phaser.Math.Distance.Between(
        this.x,
        this.y,
        this.target.body!.position.x,
        this.target.body!.position.y
      );

      if (distance <= store.player.range) {
        const now = this.scene.time.now;
        if (now - this.lastFireTime > this.fireRate / store.player.attackSpeed) {
          this.fire();
          this.lastFireTime = now;
        }
      }
    }
    
    // 회전 오브 관리
    this.manageRotatingOrbs();
    
    // 오브들 업데이트
    this.rotatingOrbs.forEach(orb => {
      if (orb && orb.active) {
        orb.update();
      }
    });
    
    // 비활성화된 오브들 제거
    this.rotatingOrbs = this.rotatingOrbs.filter(orb => orb && orb.active);
  }
  
  private manageRotatingOrbs() {
    const now = this.scene.time.now;
    const store = useMinimalLegionStore.getState();
    
    // 게임 오버 시 오브 생성 중단
    if (store.isGameOver) return;
    
    // 첫 번째 오브 생성 (게임 시작 3초 후)
    if (this.rotatingOrbs.length === 0 && now >= this.orbSpawnTime) {
      this.createRotatingOrb(0);
      this.orbSpawnTime = Number.MAX_SAFE_INTEGER; // 첫 오브 생성 후 더 이상 생성 방지
      console.log('First rotating orb created');
    }
    
    // 레벨 5마다 추가 오브 생성 (최대 3개)
    const maxOrbs = Math.min(Math.floor(store.player.level / 5) + 1, 3);
    if (this.rotatingOrbs.length < maxOrbs) {
      const angle = (360 / maxOrbs) * this.rotatingOrbs.length;
      this.createRotatingOrb(angle);
      console.log(`Additional rotating orb created. Total: ${this.rotatingOrbs.length + 1}`);
    }
  }
  
  private createRotatingOrb(startAngle: number) {
    const orb = new RotatingOrb(this.scene, this, startAngle);
    this.rotatingOrbs.push(orb);
    
    // MainScene에 오브 등록 (충돌 처리를 위해)
    const mainScene = this.scene as MainScene;
    mainScene.addRotatingOrb(orb);
  }
  
  getRotatingOrbs(): RotatingOrb[] {
    return this.rotatingOrbs;
  }

  private fire() {
    if (!this.target || !this.target.active) return;

    const store = useMinimalLegionStore.getState();
    const mainScene = this.scene as MainScene;
    
    // Get target position - use body position for GameObject
    const targetX = this.target.body!.position.x;
    const targetY = this.target.body!.position.y;
    
    console.log(`Player firing projectile from (${this.x}, ${this.y}) to target (${targetX}, ${targetY})`);
    
    mainScene.fireProjectile(
      this.x,
      this.y,
      targetX,
      targetY,
      store.player.attackPower
    );
  }

  private updateHealthBar() {
    const store = useMinimalLegionStore.getState();
    const health = store.player.health || 0;
    const maxHealth = store.player.maxHealth || 100;
    
    this.healthBar.clear();

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-25, -35, 50, 6);

    // Health
    const healthPercent = Math.max(0, Math.min(1, health / maxHealth));
    const color = healthPercent > 0.5 ? 0x00ff00 : healthPercent > 0.25 ? 0xffff00 : 0xff0000;
    this.healthBar.fillStyle(color);
    this.healthBar.fillRect(-25, -35, 50 * healthPercent, 6);
  }
}