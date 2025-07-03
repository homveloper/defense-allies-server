import { GameplayAbility, GameplayEffect, AbilityContext } from '@/packages/gas';

export class HealAbility extends GameplayAbility {
  readonly id = 'heal';
  readonly name = 'Heal';
  readonly description = 'Restores health to self or ally';
  readonly cooldown = 5000; // 5 seconds
  readonly range = 120; // Healing range
  readonly costs = [
    { attribute: 'mana', amount: 20 }
  ];

  protected canActivateCustom(context: AbilityContext): boolean {
    const { target, owner } = context;
    
    // Can heal self or allies
    const validTarget = target || owner;
    
    // Check if target can be healed (has health attribute and isn't at full health)
    if (validTarget?.abilitySystem) {
      const health = validTarget.abilitySystem.getAttribute('health');
      if (health && health.currentValue < (health.maxValue || health.baseValue)) {
        return true;
      }
    }
    
    return false;
  }

  async activate(context: AbilityContext): Promise<boolean> {
    const { owner, target } = context;
    
    const asc = owner.abilitySystem;
    if (!asc) return false;

    // Pay costs
    if (!this.payCosts(asc)) {
      return false;
    }

    // Determine heal target (self if no target specified)
    const healTarget = target || owner;
    
    // Calculate healing amount
    const healPower = asc.getAttributeFinalValue('healPower') || asc.getAttributeFinalValue('spellPower') || 30;
    const baseHealing = Math.floor(healPower * 1.2); // 120% of heal power
    const healing = Math.floor(baseHealing * (0.9 + Math.random() * 0.2)); // ±10% variation

    // Create healing effect
    const healEffect = GameplayEffect.createInstantHeal(healing);
    
    // Apply healing
    if (healTarget.abilitySystem) {
      healTarget.abilitySystem.applyGameplayEffect(healEffect);
    } else {
      // Fallback for entities without ability system
      if (typeof healTarget.heal === 'function') {
        healTarget.heal(healing);
      }
    }

    // Create visual effects
    this.createHealingEffect(context, healTarget, healing);

    const targetName = healTarget === owner ? 'self' : healTarget.constructor.name;
    console.log(`${owner.constructor.name} heals ${targetName} for ${healing} health`);
    
    return true;
  }

  private createHealingEffect(context: AbilityContext, healTarget: any, healing: number): void {
    const { scene } = context;
    
    // Casting effect on caster
    this.createCasterEffect(context, {
      color: 0x00ff00,
      scale: 1.3,
      duration: 400
    });

    // Healing beam/connection if healing another target
    if (healTarget !== context.owner) {
      this.createHealingBeam(scene, context.owner, healTarget);
    }

    // Healing effect on target
    this.createTargetHealingEffect(scene, healTarget, healing);
  }

  private createHealingBeam(scene: Phaser.Scene, caster: any, target: any): void {
    // Create healing beam
    const beam = scene.add.graphics();
    beam.lineStyle(4, 0x00ff00, 0.8);
    beam.moveTo(caster.x, caster.y);
    beam.lineTo(target.x, target.y);
    beam.strokePath();

    // Animate beam
    scene.tweens.add({
      targets: beam,
      alpha: { from: 1, to: 0 },
      duration: 600,
      ease: 'Quad.easeOut',
      onComplete: () => beam.destroy()
    });

    // Create particles along the beam
    const steps = 10;
    for (let i = 0; i <= steps; i++) {
      const t = i / steps;
      const x = caster.x + (target.x - caster.x) * t;
      const y = caster.y + (target.y - caster.y) * t;

      scene.time.delayedCall(i * 50, () => {
        const particle = scene.add.graphics();
        particle.fillStyle(0x00ff88, 0.8);
        particle.fillCircle(x, y, 3);

        scene.tweens.add({
          targets: particle,
          scale: { from: 1, to: 2 },
          alpha: { from: 0.8, to: 0 },
          duration: 400,
          ease: 'Quad.easeOut',
          onComplete: () => particle.destroy()
        });
      });
    }
  }

  private createTargetHealingEffect(scene: Phaser.Scene, target: any, healing: number): void {
    // Main healing aura
    const aura = scene.add.graphics();
    aura.fillStyle(0x00ff00, 0.3);
    aura.fillCircle(target.x, target.y, 40);
    aura.lineStyle(3, 0x00ff88, 0.8);
    aura.strokeCircle(target.x, target.y, 40);

    scene.tweens.add({
      targets: aura,
      scale: { from: 0.1, to: 1.5 },
      alpha: { from: 0.8, to: 0 },
      duration: 800,
      ease: 'Quad.easeOut',
      onComplete: () => aura.destroy()
    });

    // Healing sparkles
    for (let i = 0; i < 8; i++) {
      scene.time.delayedCall(i * 100, () => {
        const angle = (Math.PI * 2 / 8) * i;
        const radius = 20 + Math.random() * 20;
        const sparkleX = target.x + Math.cos(angle) * radius;
        const sparkleY = target.y + Math.sin(angle) * radius;

        const sparkle = scene.add.graphics();
        sparkle.fillStyle(0x88ff88, 1);
        sparkle.fillCircle(sparkleX, sparkleY, 2);

        // Animate sparkle moving towards target
        scene.tweens.add({
          targets: sparkle,
          x: target.x,
          y: target.y - 10,
          scale: { from: 1, to: 0.1 },
          alpha: { from: 1, to: 0 },
          duration: 600,
          ease: 'Quad.easeIn',
          onComplete: () => sparkle.destroy()
        });
      });
    }

    // Healing numbers
    const healText = scene.add.text(target.x, target.y - 25, `+${healing}`, {
      font: 'bold 16px Arial',
      color: '#00ff00',
      stroke: '#ffffff',
      strokeThickness: 2
    });
    healText.setOrigin(0.5);

    scene.tweens.add({
      targets: healText,
      y: healText.y - 35,
      scale: { from: 1, to: 1.2 },
      alpha: { from: 1, to: 0 },
      duration: 1200,
      ease: 'Back.easeOut',
      onComplete: () => healText.destroy()
    });

    // Pulse effect on target for visual feedback
    if (target.sprite || target.setTint) {
      const originalTint = target.tint || 0xffffff;
      
      // Green tint for healing
      if (target.setTint) {
        target.setTint(0x88ff88);
        
        scene.time.delayedCall(200, () => {
          if (target.active && target.setTint) {
            target.setTint(originalTint);
          }
        });
      }
    }
  }
}