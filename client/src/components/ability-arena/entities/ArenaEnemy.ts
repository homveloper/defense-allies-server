import * as Phaser from 'phaser';
import { 
  AbilitySystemComponent, 
  AbilitySystemUtils,
  GameplayEffect
} from '@/components/minimal-legion/systems/ability-system';

export class ArenaEnemy extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  public abilitySystem: AbilitySystemComponent;
  public health: number = 50;
  public contactDamage: number = 15;
  public expReward: number = 25;
  public scoreReward: number = 10;
  public lastPlayerDamage: number = 0;

  // Visual components
  private sprite!: Phaser.GameObjects.Graphics;
  private healthBar!: Phaser.GameObjects.Graphics;
  
  // AI
  private target: any = null;
  private moveSpeed: number = 100;
  private attackRange: number = 100;
  private lastAttackTime: number = 0;
  private attackCooldown: number = 2000;

  // Enemy type
  private enemyType: string;

  constructor(scene: Phaser.Scene, x: number, y: number, enemyType: string = 'grunt') {
    super(scene, x, y);

    this.enemyType = enemyType;

    // Initialize ability system
    this.abilitySystem = AbilitySystemUtils.createEnemyAbilitySystem(this, {
      health: this.getTypeHealth(),
      attackPower: this.getTypeAttack(),
      defense: this.getTypeDefense(),
      moveSpeed: this.getTypeMoveSpeed()
    });

    // Setup type-specific properties
    this.setupEnemyType();

    // Create visual representation
    this.createSprite();
    this.createHealthBar();

    // Setup physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    this.body.setCircle(this.getTypeSize());

    scene.add.existing(this);

    // Listen for ability system events
    this.setupAbilitySystemEvents();
    
    console.log(`Arena Enemy created: ${enemyType}`);
  }

  private getTypeHealth(): number {
    switch (this.enemyType) {
      case 'grunt': return 40;
      case 'archer': return 30;
      case 'mage': return 25;
      case 'tank': return 80;
      default: return 50;
    }
  }

  private getTypeAttack(): number {
    switch (this.enemyType) {
      case 'grunt': return 12;
      case 'archer': return 15;
      case 'mage': return 20;
      case 'tank': return 25;
      default: return 15;
    }
  }

  private getTypeDefense(): number {
    switch (this.enemyType) {
      case 'grunt': return 2;
      case 'archer': return 1;
      case 'mage': return 1;
      case 'tank': return 8;
      default: return 2;
    }
  }

  private getTypeMoveSpeed(): number {
    switch (this.enemyType) {
      case 'grunt': return 80;
      case 'archer': return 90;
      case 'mage': return 70;
      case 'tank': return 50;
      default: return 80;
    }
  }

  private getTypeSize(): number {
    switch (this.enemyType) {
      case 'grunt': return 12;
      case 'archer': return 10;
      case 'mage': return 11;
      case 'tank': return 18;
      default: return 12;
    }
  }

  private getTypeColor(): number {
    switch (this.enemyType) {
      case 'grunt': return 0xe74c3c; // Red
      case 'archer': return 0xf39c12; // Orange
      case 'mage': return 0x9b59b6; // Purple
      case 'tank': return 0x7f8c8d; // Gray
      default: return 0xe74c3c;
    }
  }

  private setupEnemyType(): void {
    this.health = this.abilitySystem.getAttributeValue('health');
    this.moveSpeed = this.abilitySystem.getAttributeValue('moveSpeed');
    
    // Type-specific setup
    switch (this.enemyType) {
      case 'archer':
        this.attackRange = 200;
        this.attackCooldown = 1500;
        this.expReward = 30;
        this.scoreReward = 15;
        break;
      case 'mage':
        this.attackRange = 180;
        this.attackCooldown = 2500;
        this.expReward = 35;
        this.scoreReward = 20;
        break;
      case 'tank':
        this.attackRange = 80;
        this.attackCooldown = 3000;
        this.contactDamage = 25;
        this.expReward = 50;
        this.scoreReward = 30;
        break;
    }
  }

  private createSprite(): void {
    this.sprite = this.scene.add.graphics();
    
    const size = this.getTypeSize();
    const color = this.getTypeColor();
    
    // Main body
    this.sprite.fillStyle(color);
    
    switch (this.enemyType) {
      case 'grunt':
        this.sprite.fillCircle(0, 0, size);
        break;
      case 'archer':
        // Triangle
        this.sprite.fillTriangle(0, -size, -size, size, size, size);
        break;
      case 'mage':
        // Star
        this.drawStar(size);
        break;
      case 'tank':
        // Square
        this.sprite.fillRect(-size, -size, size * 2, size * 2);
        break;
      default:
        this.sprite.fillCircle(0, 0, size);
    }
    
    this.add(this.sprite);
  }

  private drawStar(size: number): void {
    const spikes = 5;
    const outerRadius = size;
    const innerRadius = size * 0.5;
    
    this.sprite.beginPath();
    
    for (let i = 0; i < spikes * 2; i++) {
      const radius = i % 2 === 0 ? outerRadius : innerRadius;
      const angle = (i * Math.PI) / spikes;
      const x = Math.cos(angle) * radius;
      const y = Math.sin(angle) * radius;
      
      if (i === 0) {
        this.sprite.moveTo(x, y);
      } else {
        this.sprite.lineTo(x, y);
      }
    }
    
    this.sprite.closePath();
    this.sprite.fill();
  }

  private createHealthBar(): void {
    this.healthBar = this.scene.add.graphics();
    this.add(this.healthBar);
    this.updateHealthBar();
  }

  private setupAbilitySystemEvents(): void {
    this.abilitySystem.on('attribute-changed', (data) => {
      if (data.attribute === 'health') {
        this.health = data.newValue;
        this.updateHealthBar();
        
        if (this.health <= 0) {
          this.onDeath();
        }
      }
    });
  }

  public setTarget(target: any): void {
    this.target = target;
  }

  public update(): void {
    if (!this.target || !this.target.active) return;

    // Update ability system
    this.abilitySystem.update(this.scene.game.loop.delta);

    // AI behavior
    this.updateAI();
  }

  private updateAI(): void {
    const distance = Phaser.Math.Distance.Between(
      this.x, this.y,
      this.target.x, this.target.y
    );

    // Move towards target if not in attack range
    if (distance > this.attackRange) {
      this.moveTowardsTarget();
    } else {
      // In range - try to attack
      this.tryAttack();
    }
  }

  private moveTowardsTarget(): void {
    const angle = Phaser.Math.Angle.Between(
      this.x, this.y,
      this.target.x, this.target.y
    );

    let speed = this.moveSpeed;
    
    // Apply dev settings speed multiplier
    if (typeof window !== 'undefined') {
      const devSettings = (window as any).devSettings;
      if (devSettings?.enemySpeedMultiplier) {
        speed *= devSettings.enemySpeedMultiplier;
      }
    }

    const velocityX = Math.cos(angle) * speed;
    const velocityY = Math.sin(angle) * speed;

    this.body.setVelocity(velocityX, velocityY);

    // Rotate sprite to face movement direction
    this.sprite.rotation = angle;
  }

  private tryAttack(): void {
    const now = this.scene.time.now;
    
    if (now - this.lastAttackTime < this.attackCooldown) {
      return; // Still on cooldown
    }

    this.lastAttackTime = now;

    // Different attack patterns based on type
    switch (this.enemyType) {
      case 'archer':
        this.rangedAttack();
        break;
      case 'mage':
        this.magicAttack();
        break;
      default:
        this.meleeAttack();
        break;
    }
  }

  private meleeAttack(): void {
    // Basic melee attack using ability system
    this.abilitySystem.tryActivateAbility('basic_attack', {
      target: this.target,
      scene: this.scene
    });
  }

  private rangedAttack(): void {
    // Create projectile towards target
    const mainScene = this.scene as any;
    if (mainScene.createProjectile) {
      mainScene.createProjectile(
        this.x, this.y,
        this.target.x, this.target.y,
        this.abilitySystem.getAttributeFinalValue('attackPower'),
        true // isEnemy
      );
    }

    // Visual effect
    this.createAttackEffect(0xf39c12);
  }

  private magicAttack(): void {
    // Magical projectile with different visuals
    const mainScene = this.scene as any;
    if (mainScene.createProjectile) {
      mainScene.createProjectile(
        this.x, this.y,
        this.target.x, this.target.y,
        this.abilitySystem.getAttributeFinalValue('attackPower') * 1.2,
        true // isEnemy
      );
    }

    // Magical effect
    this.createMagicEffect();
  }

  private createAttackEffect(color: number): void {
    const effect = this.scene.add.graphics();
    effect.fillStyle(color, 0.7);
    effect.fillCircle(this.x, this.y, 20);
    
    this.scene.tweens.add({
      targets: effect,
      scale: 2,
      alpha: 0,
      duration: 300,
      onComplete: () => effect.destroy()
    });
  }

  private createMagicEffect(): void {
    // Multiple magic particles
    for (let i = 0; i < 5; i++) {
      this.scene.time.delayedCall(i * 50, () => {
        const particle = this.scene.add.graphics();
        particle.fillStyle(0x9b59b6, 0.8);
        particle.fillCircle(
          this.x + (Math.random() - 0.5) * 30,
          this.y + (Math.random() - 0.5) * 30,
          3
        );
        
        this.scene.tweens.add({
          targets: particle,
          scale: 2,
          alpha: 0,
          duration: 500,
          onComplete: () => particle.destroy()
        });
      });
    }
  }

  public takeDamage(amount: number): void {
    // Check dev settings for invincibility
    if (typeof window !== 'undefined') {
      const devSettings = (window as any).devSettings;
      if (devSettings?.enemyInvincible) {
        console.log('Enemy damage blocked by dev settings');
        return;
      }
    }
    
    // Check ability system tags for invincibility
    if (this.abilitySystem.hasTag('invincible')) {
      console.log('Enemy damage blocked by invincible tag');
      return;
    }
    
    // Apply damage multiplier from dev settings
    let finalAmount = amount;
    if (typeof window !== 'undefined') {
      const devSettings = (window as any).devSettings;
      if (devSettings?.enemyDamageMultiplier) {
        finalAmount *= devSettings.enemyDamageMultiplier;
      }
    }
    
    // Damage is handled through the ability system
    const damageEffect = GameplayEffect.createInstantDamage(finalAmount);
    this.abilitySystem.applyGameplayEffect(damageEffect);

    // Visual feedback
    this.createDamageEffect();
  }

  private createDamageEffect(): void {
    // Red flash effect overlay
    const flashEffect = this.scene.add.graphics();
    flashEffect.fillStyle(0xff0000, 0.5);
    
    const size = this.getTypeSize();
    switch (this.enemyType) {
      case 'grunt':
        flashEffect.fillCircle(this.x, this.y, size);
        break;
      case 'archer':
        flashEffect.fillTriangle(this.x, this.y - size, this.x - size, this.y + size, this.x + size, this.y + size);
        break;
      case 'mage':
        // Simple circle for star shape
        flashEffect.fillCircle(this.x, this.y, size);
        break;
      case 'tank':
        flashEffect.fillRect(this.x - size, this.y - size, size * 2, size * 2);
        break;
      default:
        flashEffect.fillCircle(this.x, this.y, size);
    }
    
    this.scene.time.delayedCall(100, () => {
      if (flashEffect) {
        flashEffect.destroy();
      }
    });
  }

  private updateHealthBar(): void {
    this.healthBar.clear();

    const barWidth = 20;
    const barHeight = 3;
    const barY = -this.getTypeSize() - 8;

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-barWidth/2, barY, barWidth, barHeight);

    // Health
    const maxHealth = this.abilitySystem.getAttribute('health')?.maxValue || 50;
    const healthPercent = this.health / maxHealth;
    const healthColor = healthPercent > 0.6 ? 0x00ff00 : 
                       healthPercent > 0.3 ? 0xffff00 : 0xff0000;
    
    this.healthBar.fillStyle(healthColor);
    this.healthBar.fillRect(-barWidth/2, barY, barWidth * healthPercent, barHeight);
  }

  private onDeath(): void {
    console.log(`${this.enemyType} enemy died!`);
    
    // Death effect
    const explosion = this.scene.add.graphics();
    explosion.fillStyle(this.getTypeColor(), 0.7);
    explosion.fillCircle(this.x, this.y, this.getTypeSize() * 2);
    
    this.scene.tweens.add({
      targets: explosion,
      scale: 3,
      alpha: 0,
      duration: 400,
      onComplete: () => explosion.destroy()
    });

    // Mark for removal
    this.setActive(false);
    this.setVisible(false);
    if (this.body) {
      this.body.enable = false;
    }
  }
}