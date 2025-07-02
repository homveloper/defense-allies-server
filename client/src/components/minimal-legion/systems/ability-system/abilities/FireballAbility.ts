import { GameplayAbility } from '../core/GameplayAbility';
import { GameplayEffect } from '../core/GameplayEffect';
import { AbilityContext } from '../types/AbilityTypes';

export class FireballAbility extends GameplayAbility {
  readonly id = 'fireball';
  readonly name = 'Fireball';
  readonly description = 'Launches a fireball that deals fire damage to enemies';
  readonly cooldown = 3000; // 3 seconds
  readonly range = 200; // Spell range
  readonly costs = [
    { attribute: 'mana', amount: 15 }
  ];

  protected canActivateCustom(context: AbilityContext): boolean {
    return !!context.target && this.isValidTarget(context.target, context);
  }

  async activate(context: AbilityContext): Promise<boolean> {
    const { owner, target } = context;
    
    if (!target) return false;

    const asc = owner.abilitySystem;
    if (!asc) return false;

    // Pay costs
    if (!this.payCosts(asc)) {
      return false;
    }

    // Calculate damage
    const spellPower = asc.getAttributeFinalValue('spellPower') || asc.getAttributeFinalValue('attackPower') || 25;
    const baseDamage = Math.floor(spellPower * 1.4); // 140% of spell power
    const damage = Math.floor(baseDamage * (0.9 + Math.random() * 0.2)); // ±10% variation

    // Create fireball projectile with delay for impact
    await this.createFireballProjectile(context, damage);

    console.log(`${owner.constructor.name} casts Fireball for ${damage} fire damage`);
    
    return true;
  }

  private async createFireballProjectile(context: AbilityContext, damage: number): Promise<void> {
    const { owner, target, scene } = context;
    
    return new Promise((resolve) => {
      // Create fireball visual
      const fireball = scene.add.graphics();
      fireball.fillStyle(0xff4500, 1);
      fireball.fillCircle(0, 0, 8);
      fireball.lineStyle(2, 0xffa500, 0.8);
      fireball.strokeCircle(0, 0, 8);
      fireball.x = owner.x;
      fireball.y = owner.y;

      // Add flame trail effect
      const particles: Phaser.GameObjects.Graphics[] = [];
      const createTrail = () => {
        if (fireball.active) {
          const particle = scene.add.graphics();
          particle.fillStyle(0xff6600, 0.6);
          particle.fillCircle(fireball.x, fireball.y, 4);
          particles.push(particle);

          // Fade out particle
          scene.tweens.add({
            targets: particle,
            alpha: 0,
            scale: 0.1,
            duration: 500,
            onComplete: () => {
              particle.destroy();
              const index = particles.indexOf(particle);
              if (index > -1) particles.splice(index, 1);
            }
          });
        }
      };

      // Create trail particles periodically
      const trailTimer = scene.time.addEvent({
        delay: 50,
        callback: createTrail,
        repeat: -1
      });

      // Calculate travel time and animate
      const distance = Phaser.Math.Distance.Between(owner.x, owner.y, target.x, target.y);
      const speed = 250; // pixels per second
      const duration = (distance / speed) * 1000;

      scene.tweens.add({
        targets: fireball,
        x: target.x,
        y: target.y,
        duration: duration,
        ease: 'Quad.easeOut',
        onComplete: () => {
          // Stop trail
          trailTimer.destroy();
          
          // Create explosion effect
          this.createExplosion(scene, target.x, target.y, damage);
          
          // Apply damage
          this.applyFireballDamage(target, damage);
          
          // Clean up
          fireball.destroy();
          particles.forEach(p => p.destroy());
          
          resolve();
        }
      });

      // Add slight rotation to fireball
      scene.tweens.add({
        targets: fireball,
        rotation: Math.PI * 4,
        duration: duration,
        ease: 'Linear'
      });
    });
  }

  private createExplosion(scene: Phaser.Scene, x: number, y: number, damage: number): void {
    // Main explosion
    const explosion = scene.add.graphics();
    explosion.fillStyle(0xff4500, 0.8);
    explosion.fillCircle(x, y, 20);
    explosion.lineStyle(3, 0xffa500, 1);
    explosion.strokeCircle(x, y, 20);

    scene.tweens.add({
      targets: explosion,
      scale: { from: 0.1, to: 3 },
      alpha: { from: 1, to: 0 },
      duration: 400,
      ease: 'Quad.easeOut',
      onComplete: () => explosion.destroy()
    });

    // Secondary explosion rings
    for (let i = 0; i < 3; i++) {
      scene.time.delayedCall(i * 100, () => {
        const ring = scene.add.graphics();
        ring.lineStyle(2, 0xff6600, 0.6);
        ring.strokeCircle(x, y, 10 + i * 15);

        scene.tweens.add({
          targets: ring,
          scale: { from: 0.5, to: 2 },
          alpha: { from: 0.8, to: 0 },
          duration: 600,
          ease: 'Quad.easeOut',
          onComplete: () => ring.destroy()
        });
      });
    }

    // Damage numbers
    const damageText = scene.add.text(x, y - 30, damage.toString(), {
      font: 'bold 18px Arial',
      color: '#ff4500',
      stroke: '#000000',
      strokeThickness: 3
    });
    damageText.setOrigin(0.5);

    scene.tweens.add({
      targets: damageText,
      y: damageText.y - 40,
      scale: { from: 1, to: 1.5 },
      alpha: { from: 1, to: 0 },
      duration: 1200,
      ease: 'Back.easeOut',
      onComplete: () => damageText.destroy()
    });

    // Screen shake for dramatic effect
    if (scene.cameras && scene.cameras.main) {
      scene.cameras.main.shake(200, 0.01);
    }
  }

  private applyFireballDamage(target: any, damage: number): void {
    if (target.abilitySystem) {
      // Apply fire damage
      const fireEffect = GameplayEffect.createInstantDamage(damage, 'fire');
      target.abilitySystem.applyGameplayEffect(fireEffect);

      // Small chance to apply burning effect
      if (Math.random() < 0.3) { // 30% chance
        const burnEffect = GameplayEffect.createDamageOverTime(5, 3000, 1000); // 5 damage per second for 3 seconds
        target.abilitySystem.applyGameplayEffect(burnEffect);
      }
    } else {
      // Fallback for entities without ability system
      if (typeof target.takeDamage === 'function') {
        target.takeDamage(damage);
      }
    }
  }
}