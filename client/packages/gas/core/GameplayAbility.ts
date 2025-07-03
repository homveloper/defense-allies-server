import { AbilityContext, AbilityCost } from '../types/AbilityTypes';
import { AbilitySystemComponent, IGameplayAbility } from './AbilitySystemComponent';

export abstract class GameplayAbility implements IGameplayAbility {
  abstract readonly id: string;
  abstract readonly name: string;
  abstract readonly description: string;
  
  // Ability configuration
  public readonly cooldown: number = 0; // milliseconds
  public readonly costs: AbilityCost[] = [];
  public readonly requiredTags: string[] = [];
  public readonly blockedByTags: string[] = [];
  public readonly range: number = -1; // -1 for unlimited range
  
  constructor(config?: Partial<GameplayAbility>) {
    if (config) {
      Object.assign(this, config);
    }
  }

  // Check if ability can be activated
  canActivate(context: AbilityContext): boolean {
    const { owner } = context;
    
    // Check if owner has ability system component
    if (!owner.abilitySystem || !(owner.abilitySystem instanceof AbilitySystemComponent)) {
      return false;
    }

    const asc = owner.abilitySystem as AbilitySystemComponent;

    // Check required tags
    if (this.requiredTags.length > 0 && !asc.hasAllTags(this.requiredTags)) {
      return false;
    }

    // Check blocked tags
    if (this.blockedByTags.length > 0 && asc.hasAnyTag(this.blockedByTags)) {
      return false;
    }

    // Check costs
    if (!this.canPayCosts(asc)) {
      return false;
    }

    // Check range if target is specified
    if (context.target && this.range > 0) {
      if (!this.isTargetInRange(context)) {
        return false;
      }
    }

    // Additional custom checks
    return this.canActivateCustom(context);
  }

  // Abstract method for ability-specific activation logic
  abstract activate(context: AbilityContext): Promise<boolean>;

  // Override this for custom activation checks
  protected canActivateCustom(context: AbilityContext): boolean {
    return true;
  }

  // Check if we can pay the costs
  protected canPayCosts(asc: AbilitySystemComponent): boolean {
    return this.costs.every(cost => {
      const currentValue = asc.getAttributeValue(cost.attribute);
      return currentValue >= cost.amount;
    });
  }

  // Pay the costs for this ability
  protected payCosts(asc: AbilitySystemComponent): boolean {
    if (!this.canPayCosts(asc)) {
      return false;
    }

    this.costs.forEach(cost => {
      asc.modifyAttribute(cost.attribute, -cost.amount);
    });

    return true;
  }

  // Check if target is in range
  protected isTargetInRange(context: AbilityContext): boolean {
    const { owner, target } = context;
    
    if (!target || this.range <= 0) {
      return true; // No range limit or no target
    }

    // Calculate distance
    const dx = target.x - owner.x;
    const dy = target.y - owner.y;
    const distance = Math.sqrt(dx * dx + dy * dy);

    return distance <= this.range;
  }

  // Get remaining cooldown for this ability
  getCooldownRemaining(asc: AbilitySystemComponent): number {
    return asc.getCooldownRemaining(this.id);
  }

  // Helper method to create simple projectile
  protected createProjectile(context: AbilityContext, config: {
    sprite?: string;
    speed?: number;
    onHit?: (target: any) => void;
    color?: number;
  }): void {
    const { owner, target, scene } = context;
    
    if (!target) return;

    const speed = config.speed || 300;
    const color = config.color || 0xffffff;

    // Create visual projectile
    let projectile: Phaser.GameObjects.GameObject;
    
    if (config.sprite) {
      projectile = scene.add.sprite(owner.x, owner.y, config.sprite);
    } else {
      // Create simple circle projectile
      const graphics = scene.add.graphics();
      graphics.fillStyle(color);
      graphics.fillCircle(0, 0, 4);
      graphics.x = owner.x;
      graphics.y = owner.y;
      projectile = graphics;
    }

    // Calculate movement
    const dx = target.x - owner.x;
    const dy = target.y - owner.y;
    const distance = Math.sqrt(dx * dx + dy * dy);
    const duration = (distance / speed) * 1000;

    // Animate projectile
    scene.tweens.add({
      targets: projectile,
      x: target.x,
      y: target.y,
      duration: duration,
      ease: 'Linear',
      onComplete: () => {
        // Hit effect
        if (config.onHit) {
          config.onHit(target);
        }
        
        // Destroy projectile
        projectile.destroy();
      }
    });
  }

  // Helper method to create area effect
  protected createAreaEffect(context: AbilityContext, config: {
    x: number;
    y: number;
    radius: number;
    color?: number;
    duration?: number;
    onAffected?: (target: any) => void;
  }): void {
    const { scene } = context;
    
    const color = config.color || 0xff0000;
    const duration = config.duration || 500;

    // Create visual effect
    const circle = scene.add.graphics();
    circle.fillStyle(color, 0.3);
    circle.fillCircle(config.x, config.y, config.radius);
    circle.lineStyle(3, color, 0.8);
    circle.strokeCircle(config.x, config.y, config.radius);

    // Animate effect
    scene.tweens.add({
      targets: circle,
      scale: { from: 0.1, to: 1 },
      alpha: { from: 1, to: 0 },
      duration: duration,
      ease: 'Quad.easeOut',
      onComplete: () => circle.destroy()
    });

    // Find affected targets (this would need to be implemented based on your game's entity system)
    if (config.onAffected) {
      // This is a placeholder - you'd implement proper target finding based on your game
      console.log(`Area effect at (${config.x}, ${config.y}) with radius ${config.radius}`);
    }
  }

  // Helper method to create visual effect on caster
  protected createCasterEffect(context: AbilityContext, config: {
    color?: number;
    scale?: number;
    duration?: number;
  }): void {
    const { owner, scene } = context;
    
    const color = config.color || 0x00ff00;
    const scale = config.scale || 1.5;
    const duration = config.duration || 300;

    // Create glow effect around caster
    const glow = scene.add.graphics();
    glow.lineStyle(4, color, 0.8);
    glow.strokeCircle(owner.x, owner.y, 30);

    scene.tweens.add({
      targets: glow,
      scale: { from: 0.5, to: scale },
      alpha: { from: 1, to: 0 },
      duration: duration,
      ease: 'Quad.easeOut',
      onComplete: () => glow.destroy()
    });
  }

  // Helper method for target validation
  protected isValidTarget(target: any, context: AbilityContext): boolean {
    if (!target || !target.active) {
      return false;
    }

    // Add more validation logic as needed
    // For example, check if target is an enemy, ally, etc.
    
    return true;
  }

  // Debug helper
  toString(): string {
    return `${this.name} (${this.id}): cooldown=${this.cooldown}ms, costs=[${this.costs.map(c => `${c.amount} ${c.attribute}`).join(', ')}]`;
  }
}