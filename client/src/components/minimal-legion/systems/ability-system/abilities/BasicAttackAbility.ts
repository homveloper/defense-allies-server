import { GameplayAbility } from '../core/GameplayAbility';
import { GameplayEffect } from '../core/GameplayEffect';
import { AbilityContext } from '../types/AbilityTypes';

export class BasicAttackAbility extends GameplayAbility {
  readonly id = 'basic_attack';
  readonly name = 'Basic Attack';
  readonly description = 'A simple melee or ranged attack';
  readonly cooldown = 800; // 0.8 seconds
  readonly range = 100; // Attack range

  protected canActivateCustom(context: AbilityContext): boolean {
    // Need a target to attack
    return !!context.target && this.isValidTarget(context.target, context);
  }

  async activate(context: AbilityContext): Promise<boolean> {
    const { owner, target } = context;
    
    if (!target) return false;

    const asc = owner.abilitySystem;
    if (!asc) return false;

    // Get attack power from attributes
    const attackPower = asc.getAttributeFinalValue('attackPower') || 25;
    
    // Calculate damage (could add random variation, critical hits, etc.)
    const baseDamage = attackPower;
    const damage = Math.floor(baseDamage * (0.8 + Math.random() * 0.4)); // ±20% variation

    // Create damage effect
    const damageEffect = GameplayEffect.createInstantDamage(damage, 'physical');
    
    // Apply damage to target
    if (target.abilitySystem) {
      target.abilitySystem.applyGameplayEffect(damageEffect);
    } else {
      // Fallback for entities without ability system
      if (typeof target.takeDamage === 'function') {
        target.takeDamage(damage);
      }
    }

    // Create visual effects
    this.createAttackEffect(context, damage);

    console.log(`${owner.constructor.name} attacks ${target.constructor.name} for ${damage} damage`);
    
    return true;
  }

  private createAttackEffect(context: AbilityContext, damage: number): void {
    const { owner, target, scene } = context;
    
    // Create projectile or melee effect based on range
    const distance = Phaser.Math.Distance.Between(owner.x, owner.y, target.x, target.y);
    
    if (distance > 50) {
      // Ranged attack - create projectile
      this.createProjectile(context, {
        color: 0xffff00,
        speed: 400,
        onHit: (hitTarget) => {
          this.createHitEffect(scene, hitTarget.x, hitTarget.y, damage);
        }
      });
    } else {
      // Melee attack - immediate effect
      this.createHitEffect(scene, target.x, target.y, damage);
      this.createCasterEffect(context, {
        color: 0xff6600,
        scale: 1.2,
        duration: 200
      });
    }
  }

  private createHitEffect(scene: Phaser.Scene, x: number, y: number, damage: number): void {
    // Impact effect
    const impact = scene.add.graphics();
    impact.fillStyle(0xff0000, 0.8);
    impact.fillCircle(x, y, 15);
    
    scene.tweens.add({
      targets: impact,
      scale: { from: 0.1, to: 2 },
      alpha: { from: 1, to: 0 },
      duration: 300,
      ease: 'Quad.easeOut',
      onComplete: () => impact.destroy()
    });

    // Damage number
    const damageText = scene.add.text(x, y - 20, damage.toString(), {
      font: 'bold 16px Arial',
      color: '#ff0000',
      stroke: '#ffffff',
      strokeThickness: 2
    });
    damageText.setOrigin(0.5);

    scene.tweens.add({
      targets: damageText,
      y: damageText.y - 30,
      alpha: 0,
      duration: 1000,
      ease: 'Quad.easeOut',
      onComplete: () => damageText.destroy()
    });
  }
}