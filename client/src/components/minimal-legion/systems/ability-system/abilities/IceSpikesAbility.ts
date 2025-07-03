import * as Phaser from 'phaser';
import { GameplayAbility } from '../core/GameplayAbility';
import { AbilityContext, AbilityActivationResult } from '../types/AbilityTypes';
import { GameplayEffect } from '../core/GameplayEffect';

export class IceSpikesAbility extends GameplayAbility {
  readonly id = 'ice_spikes';
  readonly name = 'Ice Spikes';
  readonly description = 'Creates a line of ice spikes that slows and damages enemies';
  readonly cooldown = 4000; // 4 seconds
  readonly manaCost = 20;
  readonly tags = ['damage', 'slow', 'ice', 'area'];

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
    const target = context.target || this.getDefaultTarget(context);

    if (!target) {
      return { success: false, failureReason: 'No target found' };
    }

    // Calculate direction and create spikes
    const angle = Phaser.Math.Angle.Between(owner.x, owner.y, target.x, target.y);
    const numSpikes = 5;
    const spikeDistance = 40;

    for (let i = 0; i < numSpikes; i++) {
      const distance = (i + 1) * spikeDistance;
      const spikeX = owner.x + Math.cos(angle) * distance;
      const spikeY = owner.y + Math.sin(angle) * distance;

      // Delayed spike creation
      scene.time.delayedCall(i * 150, () => {
        this.createIceSpike(scene, spikeX, spikeY);
        this.checkSpikeCollision(scene, spikeX, spikeY);
      });
    }

    return { success: true };
  }

  private getDefaultTarget(context: AbilityContext): any {
    const owner = context.owner;
    const scene = context.scene;
    const mainScene = scene as any;
    const enemies = mainScene.enemies?.children?.entries || [];
    
    // Find nearest enemy
    let nearest = null;
    let nearestDistance = Infinity;
    
    enemies.forEach((enemy: any) => {
      if (!enemy.active) return;
      const distance = Phaser.Math.Distance.Between(owner.x, owner.y, enemy.x, enemy.y);
      if (distance < nearestDistance) {
        nearestDistance = distance;
        nearest = enemy;
      }
    });

    return nearest;
  }

  private createIceSpike(scene: any, x: number, y: number): void {
    // Spike emergence effect
    const spike = scene.add.graphics();
    spike.fillStyle(0x87ceeb, 0.8);
    spike.fillRect(x - 8, y - 8, 16, 16);
    spike.setScale(0.1);

    // Spike animation
    scene.tweens.add({
      targets: spike,
      scaleX: 1,
      scaleY: 1.5,
      duration: 200,
      ease: 'Back.easeOut'
    });

    // Ground crack effect
    const crack = scene.add.graphics();
    crack.lineStyle(2, 0x87ceeb, 0.6);
    crack.beginPath();
    crack.moveTo(x - 15, y);
    crack.lineTo(x + 15, y);
    crack.strokePath();

    // Auto cleanup
    scene.time.delayedCall(3000, () => {
      scene.tweens.add({
        targets: [spike, crack],
        alpha: 0,
        duration: 500,
        onComplete: () => {
          spike.destroy();
          crack.destroy();
        }
      });
    });
  }

  private checkSpikeCollision(scene: any, spikeX: number, spikeY: number): void {
    const mainScene = scene as any;
    const enemies = mainScene.enemies?.children?.entries || [];
    
    enemies.forEach((enemy: any) => {
      if (!enemy.active) return;
      const distance = Phaser.Math.Distance.Between(spikeX, spikeY, enemy.x, enemy.y);
      
      if (distance <= 25) { // Spike hit radius
        // Apply damage
        const damageEffect = GameplayEffect.createInstantDamage(25);
        if (enemy.abilitySystem) {
          enemy.abilitySystem.applyGameplayEffect(damageEffect);
        } else if (enemy.takeDamage) {
          enemy.takeDamage(25);
        }

        // Apply slow effect
        const slowEffect = GameplayEffect.createAttributeBuff('moveSpeed', -30, 3000); // -30 speed for 3 seconds
        if (enemy.abilitySystem) {
          enemy.abilitySystem.applyGameplayEffect(slowEffect);
        }

        // Visual hit effect
        const hitEffect = scene.add.graphics();
        hitEffect.fillStyle(0x87ceeb, 0.5);
        hitEffect.fillCircle(enemy.x, enemy.y, 20);
        scene.tweens.add({
          targets: hitEffect,
          scale: 2,
          alpha: 0,
          duration: 300,
          onComplete: () => hitEffect.destroy()
        });
      }
    });
  }
}