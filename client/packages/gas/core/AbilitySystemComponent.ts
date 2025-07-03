import { GameplayAttribute } from './GameplayAttribute';
import { GameplayTagSystem } from './GameplayTagSystem';
import { GameplayEffect } from './GameplayEffect';
import { 
  AbilityContext, 
  ActiveGameplayEffect, 
  AbilityActivationResult,
  AbilitySystemEvents,
  AbilitySystemEventHandler
} from '../types/AbilityTypes';

export interface IGameplayAbility {
  readonly id: string;
  readonly name: string;
  readonly description: string;
  readonly cooldown: number;
  
  canActivate(context: AbilityContext): boolean;
  activate(context: AbilityContext): Promise<boolean>;
  getCooldownRemaining(asc: AbilitySystemComponent): number;
}

export class AbilitySystemComponent {
  private owner: any;
  private attributes: Map<string, GameplayAttribute> = new Map();
  private abilities: Map<string, IGameplayAbility> = new Map();
  private activeEffects: Map<string, ActiveGameplayEffect> = new Map();
  private cooldowns: Map<string, number> = new Map(); // abilityId -> endTime
  private tagSystem: GameplayTagSystem = new GameplayTagSystem();
  
  // Event system
  private eventHandlers: Map<keyof AbilitySystemEvents, AbilitySystemEventHandler<any>[]> = new Map();

  constructor(owner: any) {
    this.owner = owner;
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
    return attribute ? attribute.finalValue : 0;
  }

  setAttributeValue(name: string, value: number): boolean {
    const attribute = this.attributes.get(name);
    if (attribute) {
      const oldValue = attribute.currentValue;
      attribute.currentValue = value;
      
      if (oldValue !== attribute.currentValue) {
        this.emitEvent('attribute-changed', {
          attribute: name,
          oldValue,
          newValue: attribute.currentValue
        });
      }
      
      return true;
    }
    return false;
  }

  modifyAttribute(name: string, amount: number): boolean {
    const current = this.getAttributeValue(name);
    return this.setAttributeValue(name, current + amount);
  }

  // === ABILITY MANAGEMENT ===

  grantAbility(ability: IGameplayAbility): void {
    this.abilities.set(ability.id, ability);
  }

  removeAbility(abilityId: string): boolean {
    return this.abilities.delete(abilityId);
  }

  hasAbility(abilityId: string): boolean {
    return this.abilities.has(abilityId);
  }

  getAbility(abilityId: string): IGameplayAbility | undefined {
    return this.abilities.get(abilityId);
  }

  getAllAbilities(): IGameplayAbility[] {
    return Array.from(this.abilities.values());
  }

  // === ABILITY ACTIVATION ===

  async tryActivateAbility(abilityId: string, payload?: any): Promise<AbilityActivationResult> {
    const ability = this.abilities.get(abilityId);
    
    if (!ability) {
      return { success: false, failureReason: 'Ability not found' };
    }

    // Check cooldown
    const cooldownRemaining = ability.getCooldownRemaining(this);
    if (cooldownRemaining > 0) {
      return { 
        success: false, 
        failureReason: 'On cooldown',
        cooldownRemaining 
      };
    }

    // Create context
    const context: AbilityContext = {
      owner: this.owner,
      target: payload?.target,
      scene: payload?.scene || this.owner.scene,
      payload
    };

    // Check if can activate
    if (!ability.canActivate(context)) {
      return { success: false, failureReason: 'Cannot activate' };
    }

    try {
      // Activate ability
      const success = await ability.activate(context);
      
      if (success) {
        // Start cooldown
        this.startCooldown(abilityId, ability.cooldown);
        
        // Emit event
        this.emitEvent('ability-activated', { abilityId, context });
        
        return { success: true };
      } else {
        this.emitEvent('ability-failed', { abilityId, reason: 'Activation failed' });
        return { success: false, failureReason: 'Activation failed' };
      }
    } catch (error) {
      console.error(`Error activating ability ${abilityId}:`, error);
      this.emitEvent('ability-failed', { abilityId, reason: 'Execution error' });
      return { success: false, failureReason: 'Execution error' };
    }
  }

  // === COOLDOWN MANAGEMENT ===

  private startCooldown(abilityId: string, duration: number): void {
    const endTime = Date.now() + duration;
    this.cooldowns.set(abilityId, endTime);
  }

  getCooldownRemaining(abilityId: string): number {
    const endTime = this.cooldowns.get(abilityId);
    if (!endTime) return 0;
    
    const remaining = endTime - Date.now();
    return Math.max(0, remaining);
  }

  isOnCooldown(abilityId: string): boolean {
    return this.getCooldownRemaining(abilityId) > 0;
  }

  // === EFFECT MANAGEMENT ===

  applyGameplayEffect(effect: GameplayEffect): boolean {
    const activeInstance = effect.createActiveInstance();
    const effectId = effect.spec.id;

    // Handle stacking
    const existingEffect = this.activeEffects.get(effectId);
    if (existingEffect) {
      const stackedEffect = effect.applyStacking(existingEffect, Date.now());
      this.activeEffects.set(effectId, stackedEffect);
      
      // Update attribute modifiers
      this.updateEffectModifiers(effectId, stackedEffect);
    } else {
      // New effect
      this.activeEffects.set(effectId, activeInstance);
      
      // Apply attribute modifiers
      this.applyEffectModifiers(activeInstance);
      
      // Apply tags
      if (effect.spec.grantedTags) {
        effect.spec.grantedTags.forEach(tag => this.addTag(tag));
      }
      
      if (effect.spec.removedTags) {
        effect.spec.removedTags.forEach(tag => this.removeTag(tag));
      }
      
      // Call onApplied callback
      if (effect.spec.onApplied) {
        effect.spec.onApplied(this.owner);
      }
    }

    this.emitEvent('effect-applied', { effectId, target: this.owner });
    return true;
  }

  removeGameplayEffect(effectId: string): boolean {
    const activeEffect = this.activeEffects.get(effectId);
    if (!activeEffect) return false;

    // Remove attribute modifiers
    this.removeEffectModifiers(activeEffect);
    
    // Remove tags
    if (activeEffect.spec.grantedTags) {
      activeEffect.spec.grantedTags.forEach(tag => this.removeTag(tag));
    }
    
    if (activeEffect.spec.removedTags) {
      activeEffect.spec.removedTags.forEach(tag => this.addTag(tag));
    }
    
    // Call onRemoved callback
    if (activeEffect.spec.onRemoved) {
      activeEffect.spec.onRemoved(this.owner);
    }

    this.activeEffects.delete(effectId);
    this.emitEvent('effect-removed', { effectId, target: this.owner });
    
    return true;
  }

  hasGameplayEffect(effectId: string): boolean {
    return this.activeEffects.has(effectId);
  }

  getActiveEffect(effectId: string): ActiveGameplayEffect | undefined {
    return this.activeEffects.get(effectId);
  }

  getAllActiveEffects(): ActiveGameplayEffect[] {
    return Array.from(this.activeEffects.values());
  }

  // === EFFECT HELPER METHODS ===

  private applyEffectModifiers(activeEffect: ActiveGameplayEffect): void {
    if (!activeEffect.appliedModifiers) return;

    activeEffect.appliedModifiers.forEach(modifier => {
      const attribute = this.attributes.get(modifier.attribute);
      if (attribute) {
        attribute.addModifier(modifier);
      }
    });
  }

  private removeEffectModifiers(activeEffect: ActiveGameplayEffect): void {
    if (!activeEffect.appliedModifiers) return;

    activeEffect.appliedModifiers.forEach(modifier => {
      const attribute = this.attributes.get(modifier.attribute);
      if (attribute) {
        attribute.removeModifier(modifier.id);
      }
    });
  }

  private updateEffectModifiers(effectId: string, activeEffect: ActiveGameplayEffect): void {
    // Remove old modifiers
    this.attributes.forEach(attribute => {
      attribute.removeModifiersFromSource(effectId);
    });

    // Apply new modifiers
    this.applyEffectModifiers(activeEffect);
  }

  // === TAG MANAGEMENT ===

  addTag(tag: string): void {
    this.tagSystem.addTag(tag);
    this.emitEvent('tag-added', { tag });
  }

  removeTag(tag: string): void {
    if (this.tagSystem.removeTag(tag)) {
      this.emitEvent('tag-removed', { tag });
    }
  }

  hasTag(tag: string): boolean {
    return this.tagSystem.hasTag(tag);
  }

  hasAnyTag(tags: string[]): boolean {
    return this.tagSystem.hasAnyTag(tags);
  }

  hasAllTags(tags: string[]): boolean {
    return this.tagSystem.hasAllTags(tags);
  }

  matchesTagPattern(pattern: string): boolean {
    return this.tagSystem.matchesPattern(pattern);
  }

  // === UPDATE LOOP ===

  update(deltaTime: number): void {
    const currentTime = Date.now();
    
    // Update cooldowns (cleanup expired ones)
    for (const [abilityId, endTime] of this.cooldowns.entries()) {
      if (currentTime >= endTime) {
        this.cooldowns.delete(abilityId);
      }
    }

    // Update effects
    const effectsToRemove: string[] = [];
    
    for (const [effectId, activeEffect] of this.activeEffects.entries()) {
      const effect = new GameplayEffect(activeEffect.spec);
      
      // Check if effect should be removed
      if (effect.shouldRemove(activeEffect, currentTime)) {
        effectsToRemove.push(effectId);
        continue;
      }
      
      // Check if effect should tick
      if (effect.shouldTick(activeEffect, currentTime)) {
        // Update last tick time
        activeEffect.lastTickTime = currentTime;
        
        // Execute periodic effect
        if (activeEffect.spec.onPeriodic) {
          activeEffect.spec.onPeriodic(this.owner);
        }
      }
    }

    // Remove expired effects
    effectsToRemove.forEach(effectId => {
      this.removeGameplayEffect(effectId);
    });
  }

  // === EVENT SYSTEM ===

  on<T extends keyof AbilitySystemEvents>(
    event: T, 
    handler: AbilitySystemEventHandler<T>
  ): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  off<T extends keyof AbilitySystemEvents>(
    event: T, 
    handler: AbilitySystemEventHandler<T>
  ): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index >= 0) {
        handlers.splice(index, 1);
      }
    }
  }

  private emitEvent<T extends keyof AbilitySystemEvents>(
    event: T, 
    data: AbilitySystemEvents[T]
  ): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  // === UTILITY METHODS ===

  getOwner(): any {
    return this.owner;
  }

  // Debug helper
  getDebugInfo(): any {
    return {
      attributes: Array.from(this.attributes.entries()).map(([name, attr]) => ({
        name,
        current: attr.currentValue,
        final: attr.finalValue,
        modifiers: attr.modifiers.length
      })),
      abilities: Array.from(this.abilities.keys()),
      activeEffects: Array.from(this.activeEffects.keys()),
      tags: this.tagSystem.getAllTags(),
      cooldowns: Array.from(this.cooldowns.entries()).map(([id, endTime]) => ({
        abilityId: id,
        remaining: Math.max(0, endTime - Date.now())
      }))
    };
  }

  // Cleanup method
  destroy(): void {
    // Remove all effects
    Array.from(this.activeEffects.keys()).forEach(effectId => {
      this.removeGameplayEffect(effectId);
    });
    
    // Clear all data
    this.attributes.clear();
    this.abilities.clear();
    this.activeEffects.clear();
    this.cooldowns.clear();
    this.tagSystem.clear();
    this.eventHandlers.clear();
  }
}