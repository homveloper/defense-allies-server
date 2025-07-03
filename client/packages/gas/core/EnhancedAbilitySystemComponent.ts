import { GameplayAttribute } from './GameplayAttribute';
import { GameplayTagSystem } from './GameplayTagSystem';
import { GameplayEffect } from './GameplayEffect';
import { EnhancedEventSystem } from './EnhancedEventSystem';
import { ConditionManager, ComboCondition } from './ConditionSystem';
import { 
  AbilityContext, 
  EnhancedAbilityContext,
  ActiveGameplayEffect, 
  AbilityActivationResult,
  AbilitySystemEvents,
  AbilitySystemEventHandler,
  EventPriority,
  AbilityCondition
} from '../types/AbilityTypes';

export interface IGameplayAbility {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly cooldown: number;
  
  canActivate(context: AbilityContext): boolean;
  activate(context: AbilityContext): Promise<boolean>;
  getCooldownRemaining(asc: EnhancedAbilitySystemComponent): number;
}

/**
 * Enhanced Ability System Component with advanced event system and conditions
 */
export class EnhancedAbilitySystemComponent {
  private owner: any;
  private attributes: Map<string, GameplayAttribute> = new Map();
  private abilities: Map<string, IGameplayAbility> = new Map();
  private activeEffects: Map<string, ActiveGameplayEffect> = new Map();
  private cooldowns: Map<string, number> = new Map(); // abilityId -> endTime
  private tagSystem: GameplayTagSystem = new GameplayTagSystem();
  
  // Enhanced systems
  private eventSystem: EnhancedEventSystem = new EnhancedEventSystem();
  private conditionManager: ConditionManager = new ConditionManager();
  private globalConditions: string[] = []; // Conditions that apply to all abilities

  constructor(owner: any) {
    this.owner = owner;
    this.setupDefaultEventHandlers();
    this.emit('asc-initialized', { owner });
  }

  // === EVENT SYSTEM ===

  on<T extends keyof AbilitySystemEvents>(
    event: T,
    handler: AbilitySystemEventHandler<T>,
    options: {
      priority?: EventPriority;
      once?: boolean;
      filter?: (data: AbilitySystemEvents[T]) => boolean;
    } = {}
  ): void {
    this.eventSystem.on(event, handler, options);
  }

  once<T extends keyof AbilitySystemEvents>(
    event: T,
    handler: AbilitySystemEventHandler<T>,
    options: {
      priority?: EventPriority;
      filter?: (data: AbilitySystemEvents[T]) => boolean;
    } = {}
  ): void {
    this.eventSystem.once(event, handler, options);
  }

  off<T extends keyof AbilitySystemEvents>(
    event: T,
    handler: AbilitySystemEventHandler<T>
  ): void {
    this.eventSystem.off(event, handler);
  }

  emit<T extends keyof AbilitySystemEvents>(
    event: T,
    data: Omit<AbilitySystemEvents[T], 'timestamp'>
  ): void {
    this.eventSystem.emit(event, { ...data, timestamp: Date.now() } as AbilitySystemEvents[T]);
  }

  // === CONDITION SYSTEM ===

  addCondition(condition: AbilityCondition): void {
    this.conditionManager.addCondition(condition);
  }

  removeCondition(conditionId: string): void {
    this.conditionManager.removeCondition(conditionId);
  }

  addGlobalCondition(conditionId: string): void {
    if (!this.globalConditions.includes(conditionId)) {
      this.globalConditions.push(conditionId);
    }
  }

  removeGlobalCondition(conditionId: string): void {
    const index = this.globalConditions.indexOf(conditionId);
    if (index !== -1) {
      this.globalConditions.splice(index, 1);
    }
  }

  // === ATTRIBUTE MANAGEMENT ===
  
  addAttribute(name: string, baseValue: number, maxValue?: number): GameplayAttribute {
    const attribute = new GameplayAttribute(name, baseValue, maxValue);
    this.attributes.set(name, attribute);
    return attribute;
  }

  getAttribute(name: string): GameplayAttribute | undefined {
    return this.attributes.get(name);
  }

  getAttributeValue(name: string): number {
    const attribute = this.attributes.get(name);
    return attribute ? attribute.currentValue : 0;
  }

  getAttributeFinalValue(name: string): number {
    const attribute = this.attributes.get(name);
    return attribute ? attribute.getFinalValue() : 0;
  }

  setAttributeValue(name: string, value: number, source?: string): void {
    const attribute = this.attributes.get(name);
    if (!attribute) return;

    const oldValue = attribute.currentValue;
    attribute.setCurrentValue(value);
    const newValue = attribute.currentValue;
    const change = newValue - oldValue;

    this.emit('attribute-changed', { 
      attribute: name, 
      oldValue, 
      newValue, 
      change,
      source 
    });

    // Check for special attribute states
    if (newValue <= 0 && oldValue > 0) {
      this.emit('attribute-depleted', { attribute: name, previousValue: oldValue });
    }

    if (attribute.maxValue) {
      if (newValue >= attribute.maxValue && oldValue < attribute.maxValue) {
        this.emit('attribute-maximum-reached', { attribute: name, value: newValue, maximum: attribute.maxValue });
      }
      if (newValue <= 0) {
        this.emit('attribute-minimum-reached', { attribute: name, value: newValue, minimum: 0 });
      }
    }

    if (newValue > oldValue && oldValue <= 0) {
      this.emit('attribute-restored', { attribute: name, newValue, restoredAmount: change });
    }
  }

  // === ABILITY MANAGEMENT ===

  grantAbility(ability: IGameplayAbility): void {
    this.abilities.set(ability.id, ability);
  }

  removeAbility(abilityId: string): void {
    this.abilities.delete(abilityId);
    this.cooldowns.delete(abilityId);
  }

  hasAbility(abilityId: string): boolean {
    return this.abilities.has(abilityId);
  }

  getAbility(abilityId: string): IGameplayAbility | undefined {
    return this.abilities.get(abilityId);
  }

  async tryActivateAbility(
    abilityId: string, 
    context: AbilityContext | EnhancedAbilityContext
  ): Promise<AbilityActivationResult> {
    const ability = this.abilities.get(abilityId);
    if (!ability) {
      const result = { success: false, failureReason: 'Ability not found' };
      this.emit('ability-failed', { abilityId, reason: result.failureReason!, context });
      return result;
    }

    // Enhanced context with conditions
    const enhancedContext = context as EnhancedAbilityContext;
    
    // Skip all checks if force activation is enabled
    if (!enhancedContext.forceActivation) {
      // Check global conditions
      const globalConditionResult = await this.conditionManager.checkConditions(
        this.globalConditions, 
        context,
        enhancedContext.skipConditions
      );
      
      if (!globalConditionResult.passed) {
        const result = { success: false, failureReason: globalConditionResult.reason };
        this.emit('ability-blocked', { 
          abilityId, 
          reason: result.failureReason!, 
          blockedBy: this.globalConditions
        });
        return result;
      }

      // Check ability-specific conditions
      if (enhancedContext.conditions) {
        const conditionIds = enhancedContext.conditions.map(c => c.id);
        const conditionResult = await this.conditionManager.checkConditions(
          conditionIds,
          context,
          enhancedContext.skipConditions
        );

        if (!conditionResult.passed) {
          const result = { success: false, failureReason: conditionResult.reason };
          this.emit('ability-blocked', { 
            abilityId, 
            reason: result.failureReason!, 
            blockedBy: conditionIds
          });
          return result;
        }
      }

      // Standard ability checks
      if (!ability.canActivate(context)) {
        const cooldownRemaining = this.getCooldownRemaining(abilityId);
        const result = { 
          success: false, 
          failureReason: cooldownRemaining > 0 ? 'On cooldown' : 'Cannot activate',
          cooldownRemaining 
        };
        this.emit('ability-failed', { abilityId, reason: result.failureReason!, context });
        return result;
      }
    }

    try {
      // Record ability use for combo tracking
      ComboCondition.recordAbilityUse(this.owner, abilityId);

      // Activate ability
      const success = await ability.activate(context);
      
      if (success) {
        // Start cooldown
        if (ability.cooldown > 0) {
          const endTime = Date.now() + ability.cooldown;
          this.cooldowns.set(abilityId, endTime);
          this.emit('ability-cooldown-started', { abilityId, duration: ability.cooldown });
        }

        this.emit('ability-activated', { abilityId, context });
        return { success: true };
      } else {
        const result = { success: false, failureReason: 'Ability execution failed' };
        this.emit('ability-failed', { abilityId, reason: result.failureReason!, context });
        return result;
      }
    } catch (error) {
      const reason = error instanceof Error ? error.message : 'Unknown error';
      const result = { success: false, failureReason: `Ability error: ${reason}` };
      this.emit('ability-failed', { abilityId, reason: result.failureReason!, context });
      return result;
    }
  }

  getCooldownRemaining(abilityId: string): number {
    const endTime = this.cooldowns.get(abilityId);
    if (!endTime) return 0;
    
    const remaining = Math.max(0, endTime - Date.now());
    if (remaining === 0) {
      this.cooldowns.delete(abilityId);
      this.emit('ability-cooldown-ended', { abilityId });
    }
    
    return remaining;
  }

  // === EFFECT MANAGEMENT ===

  applyGameplayEffect(effect: GameplayEffect, source?: any): void {
    const activeEffect: ActiveGameplayEffect = {
      spec: effect,
      startTime: Date.now(),
      stacks: 1,
      appliedModifiers: []
    };

    // Handle stacking
    const existingEffect = this.activeEffects.get(effect.id);
    if (existingEffect) {
      switch (effect.stackingPolicy) {
        case 'aggregate':
          activeEffect.stacks = Math.min(
            existingEffect.stacks + 1,
            effect.maxStacks || Infinity
          );
          this.emit('effect-stacked', {
            effectId: effect.id,
            target: this.owner,
            currentStacks: activeEffect.stacks,
            maxStacks: effect.maxStacks || Infinity
          });
          break;
        case 'refresh':
          activeEffect.stacks = existingEffect.stacks;
          this.emit('effect-refreshed', {
            effectId: effect.id,
            target: this.owner,
            newDuration: effect.duration,
            stacks: activeEffect.stacks
          });
          break;
        case 'none':
        default:
          return; // Don't apply if stacking is not allowed
      }
      
      // Remove old effect first
      this.removeGameplayEffect(effect.id, 'replaced');
    }

    this.activeEffects.set(effect.id, activeEffect);

    // Apply modifiers and tags
    this.applyEffectModifiers(activeEffect);

    this.emit('effect-applied', { effectId: effect.id, target: this.owner, source });

    // Handle instant effects
    if (effect.duration === 0) {
      effect.onApplied?.(this.owner);
      this.removeGameplayEffect(effect.id, 'expired');
      return;
    }

    // Handle timed effects
    if (effect.duration > 0) {
      setTimeout(() => {
        if (this.activeEffects.has(effect.id)) {
          this.emit('effect-expired', { 
            effectId: effect.id, 
            target: this.owner, 
            duration: effect.duration 
          });
          this.removeGameplayEffect(effect.id, 'expired');
        }
      }, effect.duration);
    }

    effect.onApplied?.(this.owner);
  }

  private applyEffectModifiers(activeEffect: ActiveGameplayEffect): void {
    const effect = activeEffect.spec;
    
    // Apply attribute modifiers
    if (effect.attributeModifiers) {
      for (const modifier of effect.attributeModifiers) {
        const attribute = this.attributes.get(modifier.attribute);
        if (attribute) {
          const scaledModifier = {
            ...modifier,
            magnitude: modifier.magnitude * activeEffect.stacks
          };
          attribute.addModifier(scaledModifier);
          activeEffect.appliedModifiers.push(scaledModifier);
        }
      }
    }

    // Apply tags
    if (effect.grantedTags) {
      for (const tag of effect.grantedTags) {
        this.tagSystem.addTag(tag);
        this.emit('tag-added', { tag, source: effect.id });
      }
    }

    if (effect.removedTags) {
      for (const tag of effect.removedTags) {
        this.tagSystem.removeTag(tag);
        this.emit('tag-removed', { tag, reason: 'manual' });
      }
    }
  }

  removeGameplayEffect(effectId: string, reason: 'expired' | 'dispelled' | 'replaced'): void {
    const activeEffect = this.activeEffects.get(effectId);
    if (!activeEffect) return;

    // Remove attribute modifiers
    for (const modifier of activeEffect.appliedModifiers) {
      const attribute = this.attributes.get(modifier.attribute);
      if (attribute) {
        attribute.removeModifier(modifier.id);
      }
    }

    // Remove tags
    const effect = activeEffect.spec;
    if (effect.grantedTags) {
      for (const tag of effect.grantedTags) {
        this.tagSystem.removeTag(tag);
        this.emit('tag-removed', { tag, reason: 'effect-expired' });
      }
    }

    this.activeEffects.delete(effectId);
    this.emit('effect-removed', { effectId, target: this.owner, reason });

    effect.onRemoved?.(this.owner);
  }

  // === TAG MANAGEMENT ===

  hasTag(tag: string): boolean {
    return this.tagSystem.hasTag(tag);
  }

  hasAllTags(tags: string[]): boolean {
    return this.tagSystem.hasAllTags(tags);
  }

  hasAnyTag(tags: string[]): boolean {
    return this.tagSystem.hasAnyTag(tags);
  }

  addTag(tag: string, source?: string): void {
    this.tagSystem.addTag(tag);
    this.emit('tag-added', { tag, source });
  }

  removeTag(tag: string): void {
    this.tagSystem.removeTag(tag);
    this.emit('tag-removed', { tag, reason: 'manual' });
  }

  // === UPDATE LOOP ===

  update(deltaTime: number): void {
    // Update cooldowns
    for (const [abilityId, endTime] of this.cooldowns.entries()) {
      const remaining = endTime - Date.now();
      if (remaining <= 0) {
        this.cooldowns.delete(abilityId);
        this.emit('ability-cooldown-ended', { abilityId });
      } else {
        // Emit cooldown tick for debugging/UI
        const ability = this.abilities.get(abilityId);
        if (ability) {
          this.emit('cooldown-tick', { 
            abilityId, 
            remaining, 
            total: ability.cooldown 
          });
        }
      }
    }

    // Update periodic effects
    for (const [effectId, activeEffect] of this.activeEffects.entries()) {
      const effect = activeEffect.spec;
      if (effect.period && effect.onPeriodic) {
        const now = Date.now();
        const timeSinceStart = now - activeEffect.startTime;
        const timeSinceLastTick = activeEffect.lastTickTime ? now - activeEffect.lastTickTime : timeSinceStart;

        if (timeSinceLastTick >= effect.period) {
          effect.onPeriodic(this.owner);
          activeEffect.lastTickTime = now;
          
          const tickNumber = Math.floor(timeSinceStart / effect.period);
          const totalTicks = effect.duration > 0 ? Math.floor(effect.duration / effect.period) : -1;
          
          this.emit('effect-tick', { effectId, tickNumber, totalTicks });
        }
      }
    }
  }

  // === DEBUGGING ===

  getDebugInfo(): {
    attributes: Record<string, any>;
    abilities: Array<{ id: string; name: string; cooldown: number }>;
    activeEffects: Array<{ id: string; name: string; stacks: number; timeRemaining: number }>;
    tags: string[];
    cooldowns: Record<string, { remaining: number; total: number; percentage: number }>;
    eventSystem: any;
  } {
    this.emit('debug-info-requested', { requestId: `debug_${Date.now()}` });

    return {
      attributes: Object.fromEntries(
        Array.from(this.attributes.entries()).map(([name, attr]) => [
          name,
          {
            baseValue: attr.baseValue,
            currentValue: attr.currentValue,
            finalValue: attr.getFinalValue(),
            maxValue: attr.maxValue,
            modifiers: attr.getModifiers()
          }
        ])
      ),
      abilities: Array.from(this.abilities.values()).map(ability => ({
        id: ability.id,
        name: ability.name,
        cooldown: this.getCooldownRemaining(ability.id)
      })),
      activeEffects: Array.from(this.activeEffects.values()).map(activeEffect => ({
        id: activeEffect.spec.id,
        name: activeEffect.spec.name,
        stacks: activeEffect.stacks,
        timeRemaining: activeEffect.spec.duration > 0 
          ? Math.max(0, (activeEffect.startTime + activeEffect.spec.duration) - Date.now())
          : -1
      })),
      tags: Array.from(this.tagSystem.getAllTags()),
      cooldowns: Object.fromEntries(
        Array.from(this.cooldowns.entries()).map(([abilityId, endTime]) => {
          const remaining = Math.max(0, endTime - Date.now());
          const ability = this.abilities.get(abilityId);
          const total = ability?.cooldown || 0;
          return [
            abilityId,
            {
              remaining,
              total,
              percentage: total > 0 ? (remaining / total) * 100 : 0
            }
          ];
        })
      ),
      eventSystem: this.eventSystem.getDebugInfo()
    };
  }

  // === CLEANUP ===

  destroy(): void {
    this.emit('asc-destroyed', { owner: this.owner });
    this.eventSystem.removeAllListeners();
    this.abilities.clear();
    this.activeEffects.clear();
    this.cooldowns.clear();
    this.attributes.clear();
  }

  private setupDefaultEventHandlers(): void {
    // Example: Auto-cleanup expired effects
    this.on('effect-expired', (data) => {
      console.log(`Effect '${data.effectId}' expired after ${data.duration}ms`);
    });

    // Example: Log critical attribute changes
    this.on('attribute-depleted', (data) => {
      console.warn(`Attribute '${data.attribute}' depleted! Previous value: ${data.previousValue}`);
    });
  }
}