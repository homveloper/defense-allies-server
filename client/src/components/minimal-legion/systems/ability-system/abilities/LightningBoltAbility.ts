import * as Phaser from 'phaser';
import { GameplayAbility } from '../core/GameplayAbility';
import { AbilityContext, AbilityActivationResult } from '../types/AbilityTypes';
import { GameplayEffect } from '../core/GameplayEffect';

export class LightningBoltAbility extends GameplayAbility {
  readonly id = 'lightning_bolt';
  readonly name = 'Lightning Bolt';
  readonly description = 'Strikes multiple enemies with lightning';
  readonly cooldown = 3000; // 3 seconds
  readonly manaCost = 15;
  readonly tags = ['damage', 'area', 'lightning'];

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

    // Find all enemies within range
    const mainScene = scene as any;
    const enemies = mainScene.enemies?.children?.entries || [];
    const targetsInRange = enemies.filter((enemy: any) => {
      if (!enemy.active) return false;
      const distance = Phaser.Math.Distance.Between(owner.x, owner.y, enemy.x, enemy.y);
      return distance <= 300; // Lightning range
    }).slice(0, 3); // Max 3 targets

    if (targetsInRange.length === 0) {
      return { success: false, failureReason: 'No targets in range' };
    }

    // Lightning visual effects
    targetsInRange.forEach((target: any, index: number) => {
      // Delayed strikes for dramatic effect
      scene.time.delayedCall(index * 100, () => {
        this.createLightningEffect(scene, owner.x, owner.y, target.x, target.y);
        
        // Apply damage
        const damageEffect = GameplayEffect.createInstantDamage(40);
        if (target.abilitySystem) {
          target.abilitySystem.applyGameplayEffect(damageEffect);
        } else if (target.takeDamage) {
          target.takeDamage(40);
        }
      });
    });

    return { success: true };
  }

  private createLightningEffect(scene: any, startX: number, startY: number, endX: number, endY: number): void {
    // Main lightning bolt
    const lightning = scene.add.graphics();
    lightning.lineStyle(3, 0x00ffff, 1);
    
    // Jagged lightning path
    const segments = 8;
    const points = [];
    for (let i = 0; i <= segments; i++) {
      const t = i / segments;
      const x = startX + (endX - startX) * t + (Math.random() - 0.5) * 20;
      const y = startY + (endY - startY) * t + (Math.random() - 0.5) * 20;
      points.push({ x, y });
    }

    // Draw lightning
    lightning.beginPath();
    lightning.moveTo(points[0].x, points[0].y);
    for (let i = 1; i < points.length; i++) {
      lightning.lineTo(points[i].x, points[i].y);
    }
    lightning.strokePath();

    // Lightning flash effect
    const flash = scene.add.graphics();
    flash.fillStyle(0xffffff, 0.8);
    flash.fillCircle(endX, endY, 30);

    // Animate and destroy
    scene.tweens.add({
      targets: [lightning, flash],
      alpha: 0,
      duration: 200,
      onComplete: () => {
        lightning.destroy();
        flash.destroy();
      }
    });
  }
}