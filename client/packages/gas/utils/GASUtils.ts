import { AbilitySystemComponent } from '../core/AbilitySystemComponent';
import { GameplayEffect } from '../core/GameplayEffect';

/**
 * Utility functions for GAS operations
 * Pure game-agnostic utilities for the Gameplay Ability System
 */
export class GASUtils {
  /**
   * Creates a basic ability system component with common attributes
   */
  static createBasicASC(owner: any, config?: {
    health?: number;
    maxHealth?: number;
    mana?: number;
    maxMana?: number;
    stamina?: number;
    maxStamina?: number;
    attackPower?: number;
    defense?: number;
    moveSpeed?: number;
  }): AbilitySystemComponent {
    const asc = new AbilitySystemComponent(owner);
    
    // Add basic attributes with defaults
    const cfg = config || {};
    asc.addAttribute('health', cfg.health || 100, cfg.maxHealth || cfg.health || 100);
    asc.addAttribute('mana', cfg.mana || 50, cfg.maxMana || cfg.mana || 50);
    asc.addAttribute('stamina', cfg.stamina || 100, cfg.maxStamina || cfg.stamina || 100);
    asc.addAttribute('attackPower', cfg.attackPower || 25);
    asc.addAttribute('defense', cfg.defense || 5);
    asc.addAttribute('moveSpeed', cfg.moveSpeed || 100);
    
    return asc;
  }

  /**
   * Creates a player-focused ability system component
   * Alias for createBasicASC with player-friendly defaults
   */
  static createPlayerAbilitySystem(owner: any): AbilitySystemComponent {
    return GASUtils.createBasicASC(owner, {
      health: 100,
      mana: 50,
      attackPower: 25,
      defense: 5,
      moveSpeed: 100
    });
  }

  /**
   * Creates an enemy-focused ability system component
   */
  static createEnemyAbilitySystem(owner: any, config?: {
    health?: number;
    attackPower?: number;
    defense?: number;
    moveSpeed?: number;
  }): AbilitySystemComponent {
    const cfg = config || {};
    return GASUtils.createBasicASC(owner, {
      health: cfg.health || 50,
      attackPower: cfg.attackPower || 15,
      defense: cfg.defense || 2,
      moveSpeed: cfg.moveSpeed || 80
    });
  }

  /**
   * Creates common damage effects
   */
  static createDamageEffects() {
    return {
      // Instant damage
      smallDamage: GameplayEffect.createInstantDamage(15),
      mediumDamage: GameplayEffect.createInstantDamage(30),
      largeDamage: GameplayEffect.createInstantDamage(50),
      
      // Damage over time
      poison: GameplayEffect.createDamageOverTime(8, 5000, 1000), // 8 damage per second for 5 seconds
      burn: GameplayEffect.createDamageOverTime(12, 3000, 500),   // 12 damage per 0.5s for 3 seconds
      
      // Burst damage
      explosion: GameplayEffect.createInstantDamage(100),
      criticalHit: GameplayEffect.createInstantDamage(75),
    };
  }

  /**
   * Creates common healing effects
   */
  static createHealingEffects() {
    return {
      // Instant healing
      smallHeal: GameplayEffect.createInstantHeal(20),
      mediumHeal: GameplayEffect.createInstantHeal(40),
      largeHeal: GameplayEffect.createInstantHeal(60),
      fullHeal: GameplayEffect.createInstantHeal(9999),
      
      // Healing over time
      regeneration: GameplayEffect.createHealOverTime(5, 10000, 1000), // 5 heal per second for 10 seconds
      fastRegen: GameplayEffect.createHealOverTime(3, 5000, 250),      // 3 heal per 0.25s for 5 seconds
    };
  }

  /**
   * Creates common buff effects
   */
  static createBuffEffects() {
    return {
      // Attribute buffs
      attackBuff: GameplayEffect.createAttributeBuff('attackPower', 10, 10000),
      defenseBuff: GameplayEffect.createAttributeBuff('defense', 5, 8000),
      speedBuff: GameplayEffect.createAttributeBuff('moveSpeed', 20, 6000),
      healthBuff: GameplayEffect.createAttributeBuff('health', 25, 15000),
      
      // Powerful buffs
      berserker: GameplayEffect.createAttributeBuff('attackPower', 50, 8000),
      ironSkin: GameplayEffect.createAttributeBuff('defense', 20, 12000),
      swiftness: GameplayEffect.createAttributeBuff('moveSpeed', 100, 5000),
    };
  }

  /**
   * Creates common debuff effects
   */
  static createDebuffEffects() {
    return {
      // Attribute debuffs
      weakness: GameplayEffect.createAttributeBuff('attackPower', -15, 8000),
      vulnerability: GameplayEffect.createAttributeBuff('defense', -10, 6000),
      slowness: GameplayEffect.createAttributeBuff('moveSpeed', -30, 10000),
      
      // Status effects
      stun: GameplayEffect.createStun(2000),      // 2 second stun
      longStun: GameplayEffect.createStun(5000),  // 5 second stun
      freeze: GameplayEffect.createStun(3000),    // 3 second freeze (same as stun mechanically)
    };
  }

  /**
   * Creates common status effects with tags
   */
  static createStatusEffects() {
    return {
      // Immunity and protection
      invincible: new GameplayEffect({
        id: 'invincible',
        name: 'Invincible',
        duration: 5000,
        grantedTags: ['invincible', 'immune_damage']
      }),
      
      immunity: new GameplayEffect({
        id: 'immunity',
        name: 'Status Immunity',
        duration: 8000,
        grantedTags: ['immune_debuff', 'immune_stun']
      }),
      
      // Vision and detection
      stealth: new GameplayEffect({
        id: 'stealth',
        name: 'Stealth',
        duration: 10000,
        grantedTags: ['stealthed', 'hidden']
      }),
      
      truesight: new GameplayEffect({
        id: 'truesight',
        name: 'True Sight',
        duration: 15000,
        grantedTags: ['truesight', 'detect_stealth']
      }),
    };
  }

  /**
   * Utility function to find entities within range
   * Generic implementation that works with any entity array
   */
  static findEntitiesInRange(
    center: { x: number; y: number },
    entities: Array<{ x: number; y: number; active?: boolean }>,
    range: number,
    filterFn?: (entity: any) => boolean
  ): any[] {
    return entities.filter(entity => {
      // Skip inactive entities
      if (entity.active === false) return false;
      
      // Calculate distance
      const dx = entity.x - center.x;
      const dy = entity.y - center.y;
      const distance = Math.sqrt(dx * dx + dy * dy);
      
      // Check range
      if (distance > range) return false;
      
      // Apply custom filter if provided
      if (filterFn && !filterFn(entity)) return false;
      
      return true;
    });
  }

  /**
   * Find the nearest entity to a center point
   */
  static findNearestEntity(
    center: { x: number; y: number },
    entities: Array<{ x: number; y: number; active?: boolean }>,
    maxRange: number = Infinity,
    filterFn?: (entity: any) => boolean
  ): any | null {
    let nearest: any = null;
    let nearestDistance = Infinity;

    entities.forEach(entity => {
      // Skip inactive entities
      if (entity.active === false) return;
      
      // Apply custom filter if provided
      if (filterFn && !filterFn(entity)) return;
      
      // Calculate distance
      const dx = entity.x - center.x;
      const dy = entity.y - center.y;
      const distance = Math.sqrt(dx * dx + dy * dy);
      
      // Check if this is the nearest within range
      if (distance <= maxRange && distance < nearestDistance) {
        nearestDistance = distance;
        nearest = entity;
      }
    });

    return nearest;
  }

  /**
   * Calculate angle between two points
   */
  static calculateAngle(from: { x: number; y: number }, to: { x: number; y: number }): number {
    return Math.atan2(to.y - from.y, to.x - from.x);
  }

  /**
   * Calculate distance between two points
   */
  static calculateDistance(from: { x: number; y: number }, to: { x: number; y: number }): number {
    const dx = to.x - from.x;
    const dy = to.y - from.y;
    return Math.sqrt(dx * dx + dy * dy);
  }

  // Backward compatibility aliases
  static findNearestEnemy = GASUtils.findNearestEntity;
  static findEnemiesInRange = GASUtils.findEntitiesInRange;
  static logAbilitySystemState = GASUtils.logASCState;

  /**
   * Debug helper to log ability system state
   */
  static logASCState(asc: AbilitySystemComponent, label: string = 'ASC'): void {
    console.group(`${label} Debug Info`);
    console.log('Attributes:', asc.getDebugInfo().attributes);
    console.log('Abilities:', asc.getDebugInfo().abilities);
    console.log('Active Effects:', asc.getDebugInfo().activeEffects);
    console.log('Tags:', asc.getDebugInfo().tags);
    console.log('Cooldowns:', asc.getDebugInfo().cooldowns);
    console.groupEnd();
  }

  /**
   * Validate ability context has required properties
   */
  static validateAbilityContext(context: any, requiredProps: string[]): boolean {
    for (const prop of requiredProps) {
      if (context[prop] === undefined || context[prop] === null) {
        console.warn(`Missing required property '${prop}' in ability context`);
        return false;
      }
    }
    return true;
  }

  /**
   * Create a simple damage calculation with modifiers
   */
  static calculateDamage(baseDamage: number, attackPower: number = 0, multiplier: number = 1): number {
    return Math.floor((baseDamage + attackPower) * multiplier);
  }

  /**
   * Create a simple healing calculation with modifiers
   */
  static calculateHealing(baseHealing: number, healPower: number = 0, multiplier: number = 1): number {
    return Math.floor((baseHealing + healPower) * multiplier);
  }
}