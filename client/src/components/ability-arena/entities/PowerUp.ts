import * as Phaser from 'phaser';
import { GameplayEffect } from '@/components/minimal-legion/systems/ability-system';

export class PowerUp extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  
  private sprite!: Phaser.GameObjects.Graphics;
  private powerUpType: string;
  private lifetime: number = 30000; // 30 seconds
  private creationTime: number;

  constructor(scene: Phaser.Scene, x: number, y: number, type: string) {
    super(scene, x, y);

    this.powerUpType = type;
    this.creationTime = scene.time.now;

    // Create visual representation
    this.createSprite();

    // Setup physics
    scene.physics.world.enable(this);
    this.body.setCircle(15);

    scene.add.existing(this);

    // Add floating animation
    this.createFloatingAnimation();
    
    console.log(`PowerUp created: ${type} at (${x}, ${y})`);
  }

  private createSprite(): void {
    this.sprite = this.scene.add.graphics();
    
    const color = this.getTypeColor();
    const size = 12;
    
    // Outer glow
    this.sprite.fillStyle(color, 0.3);
    this.sprite.fillCircle(0, 0, size + 5);
    
    // Main shape
    this.sprite.fillStyle(color, 1);
    
    switch (this.powerUpType) {
      case 'health':
        // Cross shape
        this.sprite.fillRect(-2, -8, 4, 16);
        this.sprite.fillRect(-8, -2, 16, 4);
        break;
        
      case 'mana':
        // Diamond
        this.sprite.fillTriangle(0, -size, size, 0, 0, size);
        this.sprite.fillTriangle(0, -size, -size, 0, 0, size);
        break;
        
      case 'damage':
        // Star burst
        for (let i = 0; i < 8; i++) {
          const angle = (i * Math.PI) / 4;
          const x1 = Math.cos(angle) * 4;
          const y1 = Math.sin(angle) * 4;
          const x2 = Math.cos(angle) * size;
          const y2 = Math.sin(angle) * size;
          this.sprite.fillTriangle(0, 0, x1, y1, x2, y2);
        }
        break;
        
      case 'speed':
        // Arrow pointing up
        this.sprite.fillTriangle(0, -size, -6, 4, 6, 4);
        this.sprite.fillRect(-2, 4, 4, 8);
        break;
        
      case 'shield':
        // Shield shape
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(size * 0.7, -size * 0.3);
        this.sprite.lineTo(size * 0.7, size * 0.5);
        this.sprite.lineTo(0, size);
        this.sprite.lineTo(-size * 0.7, size * 0.5);
        this.sprite.lineTo(-size * 0.7, -size * 0.3);
        this.sprite.closePath();
        this.sprite.fill();
        break;
        
      default:
        this.sprite.fillCircle(0, 0, size);
    }
    
    this.add(this.sprite);
  }

  private getTypeColor(): number {
    switch (this.powerUpType) {
      case 'health': return 0x27ae60; // Green
      case 'mana': return 0x3498db;   // Blue
      case 'damage': return 0xe74c3c; // Red
      case 'speed': return 0xf1c40f;  // Yellow
      case 'shield': return 0x9b59b6; // Purple
      default: return 0xffffff;       // White
    }
  }

  private createFloatingAnimation(): void {
    // Vertical floating
    this.scene.tweens.add({
      targets: this,
      y: this.y - 10,
      duration: 2000,
      ease: 'Sine.easeInOut',
      yoyo: true,
      repeat: -1
    });

    // Gentle rotation
    this.scene.tweens.add({
      targets: this.sprite,
      rotation: Math.PI * 2,
      duration: 4000,
      ease: 'Linear',
      repeat: -1
    });

    // Pulsing glow effect
    this.scene.tweens.add({
      targets: this.sprite,
      alpha: 0.7,
      duration: 1500,
      ease: 'Sine.easeInOut',
      yoyo: true,
      repeat: -1
    });
  }

  public update(): void {
    // Check lifetime
    if (this.scene.time.now - this.creationTime > this.lifetime) {
      this.despawn();
    }
  }

  public collect(player: any): void {
    // Apply power-up effect
    this.applyEffect(player);
    
    // Visual collection effect
    this.createCollectionEffect();
    
    // Remove power-up
    this.destroy();
  }

  private applyEffect(player: any): void {
    if (!player.abilitySystem) {
      console.warn('Player does not have ability system');
      return;
    }

    const asc = player.abilitySystem;

    switch (this.powerUpType) {
      case 'health':
        // Restore 50 health
        const healEffect = GameplayEffect.createInstantHeal(50);
        asc.applyGameplayEffect(healEffect);
        console.log('Health power-up collected: +50 health');
        break;
        
      case 'mana':
        // Restore 30 mana
        const manaAmount = 30;
        const currentMana = asc.getAttributeValue('mana');
        const maxMana = asc.getAttribute('mana')?.maxValue || 50;
        asc.setAttributeValue('mana', Math.min(maxMana, currentMana + manaAmount));
        console.log('Mana power-up collected: +30 mana');
        break;
        
      case 'damage':
        // Temporary damage boost
        const damageBuffEffect = GameplayEffect.createAttributeBuff('attackPower', 15, 20000); // +15 attack for 20 seconds
        asc.applyGameplayEffect(damageBuffEffect);
        console.log('Damage power-up collected: +15 attack power for 20 seconds');
        break;
        
      case 'speed':
        // Temporary speed boost
        const speedBuffEffect = GameplayEffect.createAttributeBuff('moveSpeed', 50, 15000); // +50 speed for 15 seconds
        asc.applyGameplayEffect(speedBuffEffect);
        console.log('Speed power-up collected: +50 move speed for 15 seconds');
        break;
        
      case 'shield':
        // Temporary shield effect
        const shieldEffect = new GameplayEffect({
          id: `shield_${Date.now()}`,
          name: 'Shield',
          duration: 30000, // 30 seconds
          grantedTags: ['shielded'],
          onApplied: (target: any) => {
            console.log('Shield activated - absorbs next 3 hits');
            // This would need custom logic to absorb hits
          }
        });
        asc.applyGameplayEffect(shieldEffect);
        console.log('Shield power-up collected: absorbs next 3 hits');
        break;
    }
  }

  private createCollectionEffect(): void {
    const color = this.getTypeColor();
    
    // Burst effect
    for (let i = 0; i < 8; i++) {
      const angle = (i * Math.PI) / 4;
      const distance = 30 + Math.random() * 20;
      
      const particle = this.scene.add.graphics();
      particle.fillStyle(color, 0.8);
      particle.fillCircle(this.x, this.y, 3);
      
      const targetX = this.x + Math.cos(angle) * distance;
      const targetY = this.y + Math.sin(angle) * distance;
      
      this.scene.tweens.add({
        targets: particle,
        x: targetX,
        y: targetY,
        scale: 0.1,
        alpha: 0,
        duration: 400,
        ease: 'Quad.easeOut',
        onComplete: () => particle.destroy()
      });
    }

    // Central flash
    const flash = this.scene.add.graphics();
    flash.fillStyle(0xffffff, 0.9);
    flash.fillCircle(this.x, this.y, 20);
    
    this.scene.tweens.add({
      targets: flash,
      scale: 3,
      alpha: 0,
      duration: 300,
      ease: 'Quad.easeOut',
      onComplete: () => flash.destroy()
    });

    // Power-up text
    const text = this.scene.add.text(this.x, this.y - 30, this.getTypeName(), {
      font: 'bold 12px Arial',
      color: `#${color.toString(16).padStart(6, '0')}`,
      stroke: '#000000',
      strokeThickness: 2
    });
    text.setOrigin(0.5);

    this.scene.tweens.add({
      targets: text,
      y: text.y - 20,
      alpha: 0,
      duration: 1000,
      ease: 'Quad.easeOut',
      onComplete: () => text.destroy()
    });
  }

  private getTypeName(): string {
    switch (this.powerUpType) {
      case 'health': return 'Health +50';
      case 'mana': return 'Mana +30';
      case 'damage': return 'Damage +15';
      case 'speed': return 'Speed +50';
      case 'shield': return 'Shield';
      default: return 'Power Up';
    }
  }

  private despawn(): void {
    // Despawn animation
    this.scene.tweens.add({
      targets: this,
      scale: 0,
      alpha: 0,
      duration: 500,
      ease: 'Back.easeIn',
      onComplete: () => this.destroy()
    });
  }
}