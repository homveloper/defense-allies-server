import { GameplayEffectSpec, ActiveGameplayEffect, AttributeModifier } from '../types/AbilityTypes';

export class GameplayEffect {
  readonly spec: GameplayEffectSpec;

  constructor(spec: GameplayEffectSpec) {
    this.spec = { ...spec }; // Shallow copy to prevent external modification
  }

  // Create an active instance of this effect
  createActiveInstance(startTime: number = Date.now()): ActiveGameplayEffect {
    return {
      spec: this.spec,
      startTime,
      lastTickTime: startTime,
      stacks: 1,
      appliedModifiers: this.spec.attributeModifiers ? [...this.spec.attributeModifiers] : []
    };
  }

  // Check if effect should be removed (duration expired)
  shouldRemove(activeEffect: ActiveGameplayEffect, currentTime: number): boolean {
    if (this.spec.duration === -1) {
      return false; // Infinite duration
    }
    
    if (this.spec.duration === 0) {
      return true; // Instant effect, remove immediately after application
    }

    return (currentTime - activeEffect.startTime) >= this.spec.duration;
  }

  // Check if effect should tick (for periodic effects)
  shouldTick(activeEffect: ActiveGameplayEffect, currentTime: number): boolean {
    if (!this.spec.period || this.spec.period <= 0) {
      return false; // Not a periodic effect
    }

    const timeSinceLastTick = currentTime - (activeEffect.lastTickTime || activeEffect.startTime);
    return timeSinceLastTick >= this.spec.period;
  }

  // Apply stacking logic when the same effect is applied again
  applyStacking(existingEffect: ActiveGameplayEffect, currentTime: number): ActiveGameplayEffect {
    const policy = this.spec.stackingPolicy || 'none';
    const maxStacks = this.spec.maxStacks || 1;

    switch (policy) {
      case 'none':
        // Don't stack, return existing effect unchanged
        return existingEffect;

      case 'refresh':
        // Reset duration but don't increase stacks
        return {
          ...existingEffect,
          startTime: currentTime,
          lastTickTime: currentTime
        };

      case 'aggregate':
        // Add stacks up to maximum
        const newStacks = Math.min(existingEffect.stacks + 1, maxStacks);
        
        // Update attribute modifiers based on new stack count
        const updatedModifiers = this.calculateStackedModifiers(newStacks);
        
        return {
          ...existingEffect,
          stacks: newStacks,
          appliedModifiers: updatedModifiers,
          startTime: currentTime, // Also refresh duration
          lastTickTime: currentTime
        };

      default:
        return existingEffect;
    }
  }

  // Calculate modifiers adjusted for stack count
  private calculateStackedModifiers(stacks: number): AttributeModifier[] {
    if (!this.spec.attributeModifiers) {
      return [];
    }

    return this.spec.attributeModifiers.map(modifier => ({
      ...modifier,
      magnitude: modifier.magnitude * stacks
    }));
  }

  // Get remaining duration in milliseconds
  getRemainingDuration(activeEffect: ActiveGameplayEffect, currentTime: number): number {
    if (this.spec.duration === -1) {
      return -1; // Infinite
    }

    if (this.spec.duration === 0) {
      return 0; // Instant
    }

    const elapsed = currentTime - activeEffect.startTime;
    return Math.max(0, this.spec.duration - elapsed);
  }

  // Get progress as a percentage (0-1)
  getProgress(activeEffect: ActiveGameplayEffect, currentTime: number): number {
    if (this.spec.duration === -1 || this.spec.duration === 0) {
      return 1; // Infinite or instant effects are always "complete"
    }

    const elapsed = currentTime - activeEffect.startTime;
    return Math.min(1, elapsed / this.spec.duration);
  }

  // Static factory methods for common effect types
  static createInstantDamage(damage: number, damageType: string = 'physical'): GameplayEffect {
    return new GameplayEffect({
      id: `instant_damage_${Date.now()}`,
      name: `${damageType} Damage`,
      duration: 0, // Instant
      attributeModifiers: [
        {
          id: `damage_${Date.now()}`,
          attribute: 'health',
          operation: 'add',
          magnitude: -damage,
          source: 'instant_damage'
        }
      ]
    });
  }

  static createInstantHeal(healing: number): GameplayEffect {
    return new GameplayEffect({
      id: `instant_heal_${Date.now()}`,
      name: 'Healing',
      duration: 0, // Instant
      attributeModifiers: [
        {
          id: `heal_${Date.now()}`,
          attribute: 'health',
          operation: 'add',
          magnitude: healing,
          source: 'instant_heal'
        }
      ]
    });
  }

  static createAttributeBuff(
    attribute: string, 
    magnitude: number, 
    duration: number,
    operation: 'add' | 'multiply' = 'add'
  ): GameplayEffect {
    return new GameplayEffect({
      id: `buff_${attribute}_${Date.now()}`,
      name: `${attribute} Buff`,
      duration,
      stackingPolicy: 'aggregate',
      maxStacks: 5,
      attributeModifiers: [
        {
          id: `buff_${attribute}_${Date.now()}`,
          attribute,
          operation,
          magnitude,
          source: 'attribute_buff'
        }
      ],
      grantedTags: ['buffed', `buffed.${attribute}`]
    });
  }

  static createHealOverTime(healPerTick: number, duration: number, tickRate: number = 1000): GameplayEffect {
    return new GameplayEffect({
      id: `hot_${Date.now()}`,
      name: 'Heal Over Time',
      duration,
      period: tickRate,
      grantedTags: ['healing', 'regenerating'],
      onPeriodic: (target: any) => {
        // This will be handled by the AbilitySystemComponent
        // Just marking the behavior here
        console.log(`Healing ${target} for ${healPerTick}`);
      }
    });
  }

  static createDamageOverTime(damagePerTick: number, duration: number, tickRate: number = 1000): GameplayEffect {
    return new GameplayEffect({
      id: `dot_${Date.now()}`,
      name: 'Damage Over Time',
      duration,
      period: tickRate,
      grantedTags: ['damaged', 'burning'],
      onPeriodic: (target: any) => {
        // This will be handled by the AbilitySystemComponent
        console.log(`Damaging ${target} for ${damagePerTick}`);
      }
    });
  }

  static createStun(duration: number): GameplayEffect {
    return new GameplayEffect({
      id: `stun_${Date.now()}`,
      name: 'Stunned',
      duration,
      grantedTags: ['stunned', 'disabled', 'crowd_controlled'],
      attributeModifiers: [
        {
          id: `stun_speed_${Date.now()}`,
          attribute: 'moveSpeed',
          operation: 'multiply',
          magnitude: 0, // Can't move
          source: 'stun'
        }
      ]
    });
  }

  // Debug helper
  toString(): string {
    return `${this.spec.name} (${this.spec.id}): duration=${this.spec.duration}ms, modifiers=${this.spec.attributeModifiers?.length || 0}`;
  }
}