import Phaser from 'phaser';
import { MainScene } from '../scenes/MainScene';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';

export class Player extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  private sprite: Phaser.GameObjects.Graphics;
  private healthBar: Phaser.GameObjects.Graphics;
  private moveSpeed: number;
  private target: Phaser.GameObjects.GameObject | null = null;
  private lastFireTime: number = 0;
  private fireRate: number = 500; // milliseconds between shots

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

    // Physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    this.body.setCircle(20);

    // Get initial stats from store
    const store = useMinimalLegionStore.getState();
    this.moveSpeed = store.player.moveSpeed;

    scene.add.existing(this);
  }

  move(x: number, y: number) {
    const length = Math.sqrt(x * x + y * y);
    if (length > 0) {
      this.body.setVelocity(
        (x / length) * this.moveSpeed * 60,
        (y / length) * this.moveSpeed * 60
      );
    } else {
      this.body.setVelocity(0, 0);
    }
  }

  setTarget(target: Phaser.GameObjects.GameObject | null) {
    this.target = target;
  }

  takeDamage(amount: number) {
    const store = useMinimalLegionStore.getState();
    const newHealth = Math.max(0, store.player.health - amount);
    store.updatePlayer({ health: newHealth });
    this.updateHealthBar();

    if (newHealth <= 0) {
      store.gameOver();
    }
  }

  update() {
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
  }

  private fire() {
    if (!this.target || !this.target.active) return;

    const store = useMinimalLegionStore.getState();
    const mainScene = this.scene as MainScene;
    mainScene.fireProjectile(
      this.x,
      this.y,
      this.target.body!.position.x,
      this.target.body!.position.y,
      store.player.attackPower
    );
  }

  private updateHealthBar() {
    const store = useMinimalLegionStore.getState();
    this.healthBar.clear();

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-25, -35, 50, 6);

    // Health
    const healthPercent = store.player.health / store.player.maxHealth;
    this.healthBar.fillStyle(0x00ff00);
    this.healthBar.fillRect(-25, -35, 50 * healthPercent, 6);
  }
}