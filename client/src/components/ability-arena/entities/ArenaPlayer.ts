import * as Phaser from 'phaser';
import { 
  AbilitySystemComponent, 
  AbilitySystemUtils,
  BasicAttackAbility,
  FireballAbility,
  HealAbility,
  GameplayEffect
} from '@/components/minimal-legion/systems/ability-system';
import { ArenaMainScene } from '../scenes/ArenaMainScene';
import { useAbilityArenaStore } from '@/store/abilityArenaStore';

export class ArenaPlayer extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  public abilitySystem: AbilitySystemComponent;
  public health: number = 100;
  
  // Visual components
  private sprite!: Phaser.GameObjects.Graphics;
  private healthBar!: Phaser.GameObjects.Graphics;
  
  // Movement
  private moveSpeed: number = 200;
  private moveX: number = 0;
  private moveY: number = 0;
  
  // Dash ability
  private dashSpeed: number = 500;
  private dashDuration: number = 200;
  private dashCooldown: number = 3000;
  private lastDashTime: number = 0;
  private isDashing: boolean = false;

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y);

    // Initialize ability system
    this.abilitySystem = AbilitySystemUtils.createPlayerAbilitySystem(this);
    
    // Grant additional abilities for testing
    this.abilitySystem.grantAbility(new FireballAbility());
    
    // Create visual representation
    this.createSprite();
    this.createHealthBar();

    // Setup physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    this.body.setCircle(15);

    scene.add.existing(this);

    // Listen for ability system events
    this.setupAbilitySystemEvents();
    
    console.log('Arena Player created with ability system');
    AbilitySystemUtils.logAbilitySystemState(this.abilitySystem, 'Arena Player');
  }

  private createSprite(): void {
    this.sprite = this.scene.add.graphics();
    
    // Main body (blue circle)
    this.sprite.fillStyle(0x3498db);
    this.sprite.fillCircle(0, 0, 15);
    
    // Direction indicator (small triangle)
    this.sprite.fillStyle(0xffffff);
    this.sprite.fillTriangle(10, 0, 5, -5, 5, 5);
    
    this.add(this.sprite);
  }

  private createHealthBar(): void {
    this.healthBar = this.scene.add.graphics();
    this.add(this.healthBar);
    this.updateHealthBar();
  }

  private setupAbilitySystemEvents(): void {
    this.abilitySystem.on('ability-activated', (data) => {
      const store = useAbilityArenaStore.getState();
      store.incrementAbilitiesUsed();
      console.log(`Player used ability: ${data.abilityId}`);
    });

    this.abilitySystem.on('attribute-changed', (data) => {
      if (data.attribute === 'health') {
        this.health = data.newValue;
        this.updateHealthBar();
        
        // Update store
        const store = useAbilityArenaStore.getState();
        store.updatePlayerHealth(this.health);
        
        // Check for death
        if (this.health <= 0) {
          this.onDeath();
        }
      } else if (data.attribute === 'mana') {
        const store = useAbilityArenaStore.getState();
        store.updatePlayerMana(data.newValue);
      }
    });
  }

  public setMovement(moveX: number, moveY: number): void {
    this.moveX = moveX;
    this.moveY = moveY;
  }

  public update(): void {
    // Update ability system
    this.abilitySystem.update(this.scene.game.loop.delta);

    // Handle movement
    if (!this.isDashing) {
      this.handleNormalMovement();
    }

    // Update sprite rotation based on movement
    if (this.moveX !== 0 || this.moveY !== 0) {
      const angle = Math.atan2(this.moveY, this.moveX);
      this.sprite.rotation = angle;
    }

    // Update health and mana display
    this.updateResourceBars();
  }

  private handleNormalMovement(): void {
    const speed = this.abilitySystem.getAttributeFinalValue('moveSpeed') || this.moveSpeed;
    
    // Normalize diagonal movement
    let velocityX = this.moveX * speed;
    let velocityY = this.moveY * speed;
    
    if (this.moveX !== 0 && this.moveY !== 0) {
      velocityX *= 0.707; // 1/√2
      velocityY *= 0.707;
    }

    this.body.setVelocity(velocityX, velocityY);
  }

  public useDash(): void {
    const now = this.scene.time.now;
    
    if (now - this.lastDashTime < this.dashCooldown || this.isDashing) {
      return; // On cooldown or already dashing
    }

    if (this.moveX === 0 && this.moveY === 0) {
      return; // No movement direction
    }

    this.lastDashTime = now;
    this.isDashing = true;

    // Calculate dash direction
    let dashX = this.moveX;
    let dashY = this.moveY;
    
    // Normalize diagonal dashes
    if (dashX !== 0 && dashY !== 0) {
      dashX *= 0.707;
      dashY *= 0.707;
    }

    // Apply dash velocity
    this.body.setVelocity(dashX * this.dashSpeed, dashY * this.dashSpeed);

    // Visual effect
    this.createDashEffect();

    // End dash after duration
    this.scene.time.delayedCall(this.dashDuration, () => {
      this.isDashing = false;
      this.body.setVelocity(0, 0);
    });

    console.log('Player dashed!');
  }

  private createDashEffect(): void {
    // Trail effect
    const trail = this.scene.add.graphics();
    trail.fillStyle(0x3498db, 0.3);
    trail.fillCircle(this.x, this.y, 20);
    
    this.scene.tweens.add({
      targets: trail,
      scale: 2,
      alpha: 0,
      duration: 300,
      onComplete: () => trail.destroy()
    });

    // Screen effect
    if (this.scene.cameras && this.scene.cameras.main) {
      this.scene.cameras.main.flash(100, 100, 150, 255);
    }
  }

  public useAbility(slot: 'Q' | 'E' | 'R'): void {
    const target = this.findNearestEnemy();
    
    switch (slot) {
      case 'Q':
        // Fireball
        if (target) {
          this.abilitySystem.tryActivateAbility('fireball', { 
            target, 
            scene: this.scene 
          });
        }
        break;
        
      case 'E':
        // Heal
        this.abilitySystem.tryActivateAbility('heal', { 
          scene: this.scene 
        });
        break;
        
      case 'R':
        // Ultimate (could be a powerful ability)
        console.log('Ultimate ability not implemented yet');
        break;
    }
  }

  public handleLeftClick(targetX: number, targetY: number): void {
    // Basic attack towards clicked position
    this.abilitySystem.tryActivateAbility('basic_attack', {
      target: { x: targetX, y: targetY, body: { position: { x: targetX, y: targetY } } },
      scene: this.scene
    });
  }

  public handleRightClick(targetX: number, targetY: number): void {
    // Cast fireball towards clicked position
    const target = this.findNearestEnemy() || { 
      x: targetX, 
      y: targetY, 
      body: { position: { x: targetX, y: targetY } } 
    };
    
    this.abilitySystem.tryActivateAbility('fireball', { 
      target, 
      scene: this.scene 
    });
  }

  private findNearestEnemy(): any {
    const mainScene = this.scene as ArenaMainScene;
    const enemies = (mainScene as any).enemies?.children?.entries || [];
    
    return AbilitySystemUtils.findNearestEnemy(this, enemies, 300);
  }

  public takeDamage(amount: number): void {
    // Damage is handled through the ability system
    const damageEffect = GameplayEffect.createInstantDamage(amount);
    this.abilitySystem.applyGameplayEffect(damageEffect);

    // Visual feedback
    this.createDamageEffect();
  }

  private createDamageEffect(): void {
    // Red flash effect overlay
    const flashEffect = this.scene.add.graphics();
    flashEffect.fillStyle(0xff0000, 0.5);
    flashEffect.fillCircle(this.x, this.y, 20); // Player size
    
    this.scene.time.delayedCall(100, () => {
      if (flashEffect) {
        flashEffect.destroy();
      }
    });

    // Screen shake
    if (this.scene.cameras && this.scene.cameras.main) {
      this.scene.cameras.main.shake(100, 0.01);
    }
  }

  private updateHealthBar(): void {
    this.healthBar.clear();

    const barWidth = 30;
    const barHeight = 4;
    const barY = -25;

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-barWidth/2, barY, barWidth, barHeight);

    // Health
    const maxHealth = this.abilitySystem.getAttribute('health')?.maxValue || 100;
    const healthPercent = this.health / maxHealth;
    const healthColor = healthPercent > 0.6 ? 0x00ff00 : 
                       healthPercent > 0.3 ? 0xffff00 : 0xff0000;
    
    this.healthBar.fillStyle(healthColor);
    this.healthBar.fillRect(-barWidth/2, barY, barWidth * healthPercent, barHeight);
  }

  private updateResourceBars(): void {
    // This could be expanded to show mana/stamina bars
    this.updateHealthBar();
  }

  private onDeath(): void {
    console.log('Player died!');
    
    // Death effect
    const explosion = this.scene.add.graphics();
    explosion.fillStyle(0xff0000, 0.5);
    explosion.fillCircle(this.x, this.y, 50);
    
    this.scene.tweens.add({
      targets: explosion,
      scale: 3,
      alpha: 0,
      duration: 500,
      onComplete: () => explosion.destroy()
    });

    // Disable player
    this.setActive(false);
    this.setVisible(false);
    this.body.enable = false;
  }

  // Getter for compatibility
  get maxHealth(): number {
    return this.abilitySystem.getAttribute('health')?.maxValue || 100;
  }

  get mana(): number {
    return this.abilitySystem.getAttributeValue('mana');
  }

  get maxMana(): number {
    return this.abilitySystem.getAttribute('mana')?.maxValue || 50;
  }

  get stamina(): number {
    return this.abilitySystem.getAttributeValue('stamina');
  }

  get maxStamina(): number {
    return this.abilitySystem.getAttribute('stamina')?.maxValue || 100;
  }

  // Debug method
  public logState(): void {
    AbilitySystemUtils.logAbilitySystemState(this.abilitySystem, 'Arena Player');
  }
}