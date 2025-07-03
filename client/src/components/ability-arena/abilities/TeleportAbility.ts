import * as Phaser from 'phaser';
import { GameplayAbility } from '@/packages/gas';
import { AbilityContext, AbilityActivationResult } from '@/packages/gas';

export class TeleportAbility extends GameplayAbility {
  readonly id = 'teleport';
  readonly name = 'Teleport';
  readonly description = 'Instantly teleports to target location';
  readonly cooldown = 5000; // 5 seconds
  readonly manaCost = 25;
  readonly tags = ['movement', 'teleport', 'escape'];

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
    
    // Get mouse position or use default
    let targetX, targetY;
    
    if (context.target && context.target.x !== undefined) {
      targetX = context.target.x;
      targetY = context.target.y;
    } else {
      // Use scene center as fallback
      targetX = scene.cameras.main.centerX;
      targetY = scene.cameras.main.centerY;
    }

    // Limit teleport range
    const maxRange = 200;
    const distance = Phaser.Math.Distance.Between(owner.x, owner.y, targetX, targetY);
    
    if (distance > maxRange) {
      const angle = Phaser.Math.Angle.Between(owner.x, owner.y, targetX, targetY);
      targetX = owner.x + Math.cos(angle) * maxRange;
      targetY = owner.y + Math.sin(angle) * maxRange;
    }

    // Create teleport effects
    this.createTeleportEffect(scene, owner.x, owner.y, false); // Start position
    
    // Teleport the player
    owner.x = targetX;
    owner.y = targetY;
    
    this.createTeleportEffect(scene, targetX, targetY, true); // End position

    return { success: true };
  }

  private createTeleportEffect(scene: any, x: number, y: number, isArrival: boolean): void {
    // Portal effect
    const portal = scene.add.graphics();
    portal.fillStyle(isArrival ? 0x00ff00 : 0xff00ff, 0.7);
    
    // Draw portal rings
    for (let i = 0; i < 3; i++) {
      const radius = 10 + i * 8;
      portal.lineStyle(3, isArrival ? 0x00ff00 : 0xff00ff, 0.8 - i * 0.2);
      portal.strokeCircle(x, y, radius);
    }

    // Particle burst
    for (let i = 0; i < 12; i++) {
      const angle = (i / 12) * Math.PI * 2;
      const particle = scene.add.graphics();
      particle.fillStyle(isArrival ? 0x00ff00 : 0xff00ff, 0.8);
      particle.fillCircle(x, y, 2);

      const targetX = x + Math.cos(angle) * 30;
      const targetY = y + Math.sin(angle) * 30;

      scene.tweens.add({
        targets: particle,
        x: targetX,
        y: targetY,
        alpha: 0,
        scale: 0.1,
        duration: 400,
        onComplete: () => particle.destroy()
      });
    }

    // Portal animation
    scene.tweens.add({
      targets: portal,
      scaleX: isArrival ? 0.1 : 2,
      scaleY: isArrival ? 0.1 : 2,
      alpha: 0,
      duration: 400,
      ease: isArrival ? 'Back.easeIn' : 'Back.easeOut',
      onComplete: () => portal.destroy()
    });
  }
}