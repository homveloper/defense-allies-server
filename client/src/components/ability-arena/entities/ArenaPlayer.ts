import * as Phaser from 'phaser';
import { 
  AbilitySystemComponent, 
  AbilitySystemUtils,
  BasicAttackAbility,
  FireballAbility,
  HealAbility,
  LightningBoltAbility,
  IceSpikesAbility,
  TeleportAbility,
  ShieldBubbleAbility,
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

  // Random ability system
  private currentRandomAbility: any = null;
  private availableAbilities = [
    FireballAbility,
    LightningBoltAbility,
    IceSpikesAbility,
    TeleportAbility,
    ShieldBubbleAbility,
    HealAbility
  ];

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y);

    // Initialize ability system
    this.abilitySystem = AbilitySystemUtils.createPlayerAbilitySystem(this);
    
    // Grant additional abilities for testing
    this.abilitySystem.grantAbility(new FireballAbility());
    
    // Set initial random ability
    this.swapRandomAbility();
    
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
    
    // Setup keyboard controls
    this.setupKeyboardControls();
    
    console.log('Arena Player created with ability system');
    AbilitySystemUtils.logAbilitySystemState(this.abilitySystem, 'Arena Player');
    
    // Register globally for debug panel access
    if (typeof window !== 'undefined') {
      (window as any).currentArenaPlayer = this;
    }
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

  // Setup keyboard controls for random ability system
  private setupKeyboardControls(): void {
    // F key for random ability swap
    const fKey = this.scene.input.keyboard?.addKey('F');
    if (fKey) {
      fKey.on('down', () => {
        this.swapRandomAbility();
      });
    }

    // Left click for ability activation
    this.scene.input.on('pointerdown', (pointer: any) => {
      if (pointer.leftButtonDown()) {
        this.activateRandomAbility(pointer.worldX, pointer.worldY);
      }
    });
  }

  // Swap to a random ability
  private swapRandomAbility(): void {
    // Remove current random ability if exists
    if (this.currentRandomAbility) {
      const abilityId = this.currentRandomAbility.id;
      this.abilitySystem.removeAbility(abilityId);
    }

    // Select random ability
    const randomIndex = Math.floor(Math.random() * this.availableAbilities.length);
    const AbilityClass = this.availableAbilities[randomIndex];
    this.currentRandomAbility = new AbilityClass();

    // Grant new ability
    this.abilitySystem.grantAbility(this.currentRandomAbility);

    // Visual feedback
    this.createAbilitySwapEffect();

    // Update UI or log
    console.log(`🎯 Random ability swapped to: ${this.currentRandomAbility.name}`);
    
    // Update store
    const store = useAbilityArenaStore.getState();
    store.updateCurrentRandomAbility(this.currentRandomAbility.name);
    
    // Show ability name in game
    this.showAbilityName(this.currentRandomAbility.name);
  }

  // Activate the current random ability
  private activateRandomAbility(targetX: number, targetY: number): void {
    if (!this.currentRandomAbility) {
      console.log('No random ability available');
      return;
    }

    // Create target context
    const target = { x: targetX, y: targetY };
    
    // Try to activate the ability
    this.abilitySystem.tryActivateAbility(this.currentRandomAbility.id, {
      target,
      scene: this.scene
    }).then(result => {
      if (result.success) {
        console.log(`✨ ${this.currentRandomAbility.name} activated!`);
      } else {
        console.log(`❌ ${this.currentRandomAbility.name} failed: ${result.failureReason || 'Unknown reason'}`);
      }
    });
  }

  // Visual effect for ability swap
  private createAbilitySwapEffect(): void {
    // Colorful rings around player
    const colors = [0xff00ff, 0x00ffff, 0xffff00, 0xff6600];
    
    colors.forEach((color, index) => {
      const ring = this.scene.add.graphics();
      ring.lineStyle(3, color, 0.8);
      ring.strokeCircle(this.x, this.y, 20 + index * 10);
      
      this.scene.tweens.add({
        targets: ring,
        scaleX: 2,
        scaleY: 2,
        alpha: 0,
        duration: 800 + index * 100,
        ease: 'Power2',
        onComplete: () => ring.destroy()
      });
    });

    // Sparkle particles
    for (let i = 0; i < 12; i++) {
      const angle = (i / 12) * Math.PI * 2;
      const particle = this.scene.add.graphics();
      particle.fillStyle(0xffffff, 0.9);
      
      // Draw a simple diamond/star shape
      particle.beginPath();
      particle.moveTo(this.x, this.y - 4);
      particle.lineTo(this.x + 2, this.y);
      particle.lineTo(this.x, this.y + 4);
      particle.lineTo(this.x - 2, this.y);
      particle.closePath();
      particle.fillPath();
      
      const targetX = this.x + Math.cos(angle) * 40;
      const targetY = this.y + Math.sin(angle) * 40;
      
      this.scene.tweens.add({
        targets: particle,
        x: targetX,
        y: targetY,
        scaleX: 0.1,
        scaleY: 0.1,
        alpha: 0,
        duration: 600,
        onComplete: () => particle.destroy()
      });
    }
  }

  // Show ability name on screen
  private showAbilityName(abilityName: string): void {
    const text = this.scene.add.text(this.x, this.y - 40, abilityName, {
      fontSize: '16px',
      color: '#ffff00',
      stroke: '#000000',
      strokeThickness: 2,
      fontStyle: 'bold'
    });
    text.setOrigin(0.5);

    // Animate text
    this.scene.tweens.add({
      targets: text,
      y: text.y - 30,
      alpha: 0,
      scale: 1.2,
      duration: 2000,
      onComplete: () => text.destroy()
    });
  }

  // Public method to get current random ability info
  public getCurrentRandomAbility(): string {
    return this.currentRandomAbility ? this.currentRandomAbility.name : 'None';
  }
}