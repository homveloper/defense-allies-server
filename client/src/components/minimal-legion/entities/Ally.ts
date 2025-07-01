import Phaser from 'phaser';
import { MainScene } from '../scenes/MainScene';
import { Player } from './Player';
import { Enemy } from './Enemy';

export class Ally extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  public health: number;
  private maxHealth: number;
  private sprite: Phaser.GameObjects.Graphics;
  private healthBar: Phaser.GameObjects.Graphics;
  private moveSpeed: number;
  private player: Player | null = null;
  private target: Enemy | null = null;
  private attackPower: number;
  private attackRange: number = 120;
  private lastFireTime: number = 0;
  private fireRate: number = 700;
  private followDistance: number = 80;

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y);

    // Ally stats
    this.maxHealth = 50;
    this.health = this.maxHealth;
    this.moveSpeed = 4;
    this.attackPower = 8;

    // Create ally sprite (light blue circle)
    this.sprite = scene.add.graphics();
    this.sprite.fillStyle(0x5dade2);
    this.sprite.fillCircle(0, 0, 12);
    this.add(this.sprite);

    // Create health bar
    this.healthBar = scene.add.graphics();
    this.add(this.healthBar);
    this.updateHealthBar();

    // Physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    this.body.setCircle(12);

    scene.add.existing(this);
  }

  setPlayer(player: Player) {
    this.player = player;
  }

  takeDamage(amount: number) {
    this.health -= amount;
    this.updateHealthBar();

    if (this.health <= 0) {
      this.destroy();
    }
  }

  update() {
    if (!this.player || !this.player.active) return;

    // Find nearest enemy
    const mainScene = this.scene as MainScene;
    const enemies = mainScene['enemies'].children.entries;
    
    let nearestEnemy: Enemy | null = null;
    let nearestDistance = Infinity;

    enemies.forEach((enemy) => {
      if (enemy.active) {
        const distance = Phaser.Math.Distance.Between(
          this.x,
          this.y,
          enemy.body!.position.x,
          enemy.body!.position.y
        );
        if (distance < nearestDistance && distance < this.attackRange * 2) {
          nearestDistance = distance;
          nearestEnemy = enemy as Enemy;
        }
      }
    });

    this.target = nearestEnemy;

    // Attack if target is in range
    if (this.target && nearestDistance <= this.attackRange) {
      this.body.setVelocity(0, 0);
      
      const now = this.scene.time.now;
      if (now - this.lastFireTime > this.fireRate) {
        this.fire();
        this.lastFireTime = now;
      }
    } else {
      // Follow player
      const playerDistance = Phaser.Math.Distance.Between(
        this.x,
        this.y,
        this.player.x,
        this.player.y
      );

      if (playerDistance > this.followDistance) {
        const angle = Phaser.Math.Angle.Between(
          this.x,
          this.y,
          this.player.x,
          this.player.y
        );

        this.body.setVelocity(
          Math.cos(angle) * this.moveSpeed * 60,
          Math.sin(angle) * this.moveSpeed * 60
        );
      } else {
        this.body.setVelocity(0, 0);
      }
    }
  }

  private fire() {
    if (!this.target || !this.target.active) return;

    const mainScene = this.scene as MainScene;
    mainScene.fireProjectile(
      this.x,
      this.y,
      this.target.x,
      this.target.y,
      this.attackPower
    );
  }

  private updateHealthBar() {
    this.healthBar.clear();

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-15, -20, 30, 3);

    // Health
    const healthPercent = this.health / this.maxHealth;
    this.healthBar.fillStyle(0x00ff00);
    this.healthBar.fillRect(-15, -20, 30 * healthPercent, 3);
  }
}