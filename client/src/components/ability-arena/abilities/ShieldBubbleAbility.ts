import * as Phaser from 'phaser';
import { GameplayAbility } from '@/packages/gas';
import { AbilityContext, AbilityActivationResult } from '@/packages/gas';
import { GameplayEffect } from '@/packages/gas';

export class ShieldBubbleAbility extends GameplayAbility {
  readonly id = 'shield_bubble';
  readonly name = 'Shield Bubble';
  readonly description = 'Creates a protective bubble that absorbs damage';
  readonly cooldown = 8000; // 8 seconds
  readonly manaCost = 30;
  readonly tags = ['defense', 'shield', 'protection'];

  constructor() {
    super();
  }

  async activate(context: AbilityContext): Promise<boolean> {
    const result = await this.executeAbility(context);
    return result.success;
  }

  protected async executeAbility(context: AbilityContext): Promise<AbilityActivationResult> {
    const owner = context.owner;
    const scene = context.scene;

    // Apply shield effect
    const shieldEffect = new GameplayEffect({
      id: `shield_bubble_${Date.now()}`,
      name: 'Shield Bubble',
      duration: 10000, // 10 seconds
      grantedTags: ['shielded', 'protected'],
      attributeModifiers: [
        {
          id: `shield_${Date.now()}`,
          attribute: 'defense',
          operation: 'add',
          magnitude: 50,
          source: 'shield_bubble'
        }
      ]
    });

    if (owner.abilitySystem) {
      owner.abilitySystem.applyGameplayEffect(shieldEffect);
    }

    // Create visual shield
    this.createShieldVisual(scene, owner);

    return { success: true };
  }

  private createShieldVisual(scene: any, owner: any): void {
    // Shield bubble visual
    const shield = scene.add.graphics();
    shield.lineStyle(3, 0x00aaff, 0.8);
    shield.fillStyle(0x00aaff, 0.1);
    shield.fillCircle(0, 0, 25);
    shield.strokeCircle(0, 0, 25);

    // Attach to player
    owner.add(shield);

    // Pulsing animation
    scene.tweens.add({
      targets: shield,
      scaleX: 1.1,
      scaleY: 1.1,
      alpha: 0.7,
      duration: 1000,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.easeInOut'
    });

    // Energy particles
    const createEnergyParticle = () => {
      if (!shield.active) return;
      
      const angle = Math.random() * Math.PI * 2;
      const radius = 20 + Math.random() * 10;
      const particle = scene.add.graphics();
      particle.fillStyle(0x00aaff, 0.8);
      particle.fillCircle(0, 0, 1);
      
      const startX = owner.x + Math.cos(angle) * radius;
      const startY = owner.y + Math.sin(angle) * radius;
      particle.setPosition(startX, startY);

      scene.tweens.add({
        targets: particle,
        x: owner.x,
        y: owner.y,
        alpha: 0,
        duration: 800,
        onComplete: () => particle.destroy()
      });
    };

    // Particle timer
    const particleTimer = scene.time.addEvent({
      delay: 200,
      callback: createEnergyParticle,
      repeat: 49 // 10 seconds worth
    });

    // Remove shield after duration
    scene.time.delayedCall(10000, () => {
      if (shield && shield.active) {
        scene.tweens.add({
          targets: shield,
          alpha: 0,
          scale: 0,
          duration: 500,
          onComplete: () => {
            shield.destroy();
            particleTimer.remove();
          }
        });
      }
    });
  }
}