import Phaser from 'phaser';
import { MainScene } from '../scenes/MainScene';

export class Enemy extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  public health: number;
  private maxHealth: number;
  private sprite: Phaser.GameObjects.Graphics;
  private healthBar: Phaser.GameObjects.Graphics;
  private moveSpeed: number;
  private target: Phaser.GameObjects.GameObject | null = null;
  private attackPower: number;
  private attackRange: number = 100;
  private meleeRange: number = 25;
  private lastFireTime: number = 0;
  private lastMeleeTime: number = 0;
  private fireRate: number = 1500;
  private meleeRate: number = 1200;

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y);

    // Enemy stats
    this.maxHealth = 30;
    this.health = this.maxHealth;
    this.moveSpeed = 3;
    this.attackPower = 5;

    // Create enemy sprite (red circle)
    this.sprite = scene.add.graphics();
    this.sprite.fillStyle(0xe74c3c);
    this.sprite.fillCircle(0, 0, 15);
    this.add(this.sprite);

    // Create health bar
    this.healthBar = scene.add.graphics();
    this.add(this.healthBar);
    this.updateHealthBar();

    // Physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    this.body.setCircle(15);

    scene.add.existing(this);
  }

  setTarget(target: Phaser.GameObjects.GameObject) {
    this.target = target;
  }

  takeDamage(amount: number) {
    this.health -= amount;
    this.updateHealthBar();

    // Flash effect
    this.sprite.clear();
    this.sprite.fillStyle(0xffffff);
    this.sprite.fillCircle(0, 0, 15);

    this.scene.time.delayedCall(100, () => {
      this.sprite.clear();
      this.sprite.fillStyle(0xe74c3c);
      this.sprite.fillCircle(0, 0, 15);
    });
  }

  update() {
    if (!this.target || !this.target.active) return;

    const distance = Phaser.Math.Distance.Between(
      this.x,
      this.y,
      this.target.body!.position.x,
      this.target.body!.position.y
    );

    const now = this.scene.time.now;

    // Always move towards target unless very close for melee
    if (distance > this.meleeRange) {
      const angle = Phaser.Math.Angle.Between(
        this.x,
        this.y,
        this.target.body!.position.x,
        this.target.body!.position.y
      );

      this.body.setVelocity(
        Math.cos(angle) * this.moveSpeed * 60,
        Math.sin(angle) * this.moveSpeed * 60
      );
    } else {
      // Very close - slow down but keep moving slightly
      const angle = Phaser.Math.Angle.Between(
        this.x,
        this.y,
        this.target.body!.position.x,
        this.target.body!.position.y
      );

      this.body.setVelocity(
        Math.cos(angle) * this.moveSpeed * 20,
        Math.sin(angle) * this.moveSpeed * 20
      );
    }

    // Attack based on distance
    if (distance <= this.meleeRange && now - this.lastMeleeTime > this.meleeRate) {
      this.meleeAttack();
      this.lastMeleeTime = now;
    } else if (distance <= this.attackRange && distance > this.meleeRange && now - this.lastFireTime > this.fireRate) {
      this.fire();
      this.lastFireTime = now;
    }
  }

  private meleeAttack() {
    if (!this.target || !this.target.active || !this.attackPower) return;

    const mainScene = this.scene as MainScene;
    const damage = this.attackPower || 5; // Default damage if undefined
    
    console.log(`Enemy dealing ${damage} melee damage`);
    mainScene.dealMeleeDamage(this.target, damage);

    // Visual feedback for melee attack
    if (this.sprite && this.active) {
      this.sprite.clear();
      this.sprite.fillStyle(0xffffff);
      this.sprite.fillCircle(0, 0, 18);

      this.scene.time.delayedCall(100, () => {
        if (this.sprite && this.active) {
          this.sprite.clear();
          this.sprite.fillStyle(0xe74c3c);
          this.sprite.fillCircle(0, 0, 15);
        }
      });
    }
  }

  private fire() {
    if (!this.target || !this.target.active) return;

    const mainScene = this.scene as MainScene;
    mainScene.fireProjectile(
      this.x,
      this.y,
      this.target.body!.position.x,
      this.target.body!.position.y,
      this.attackPower,
      true // isEnemy
    );
  }

  private updateHealthBar() {
    this.healthBar.clear();

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-20, -25, 40, 4);

    // Health
    const healthPercent = this.health / this.maxHealth;
    this.healthBar.fillStyle(0xff0000);
    this.healthBar.fillRect(-20, -25, 40 * healthPercent, 4);
  }
}