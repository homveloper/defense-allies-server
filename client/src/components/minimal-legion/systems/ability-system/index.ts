// Core Components
export { AbilitySystemComponent } from './core/AbilitySystemComponent';
export type { IGameplayAbility } from './core/AbilitySystemComponent';
export { GameplayAttribute } from './core/GameplayAttribute';
export { GameplayTagSystem } from './core/GameplayTagSystem';
export { GameplayEffect } from './core/GameplayEffect';
export { GameplayAbility } from './core/GameplayAbility';

// Types
export type {
  AbilityContext,
  AbilityCost,
  CooldownInfo,
  AttributeModifier,
  GameplayAttributeData,
  GameplayEffectSpec,
  ActiveGameplayEffect,
  AbilityActivationResult,
  AbilitySystemEvents,
  AbilitySystemEventHandler
} from './types/AbilityTypes';

// Basic Abilities
export { BasicAttackAbility } from './abilities/BasicAttackAbility';
export { FireballAbility } from './abilities/FireballAbility';
export { HealAbility } from './abilities/HealAbility';

// Import required dependencies for utilities
import { AbilitySystemComponent } from './core/AbilitySystemComponent';
import { GameplayEffect } from './core/GameplayEffect';
import { BasicAttackAbility } from './abilities/BasicAttackAbility';
import { HealAbility } from './abilities/HealAbility';
import * as Phaser from 'phaser';

// Utility functions for quick setup
export class AbilitySystemUtils {
  /**
   * Creates a basic player ability system with common attributes
   */
  static createPlayerAbilitySystem(owner: any): AbilitySystemComponent {
    const asc = new AbilitySystemComponent(owner);
    
    // Add common player attributes
    asc.addAttribute('health', 100, 100);
    asc.addAttribute('mana', 50, 50);
    asc.addAttribute('attackPower', 25);
    asc.addAttribute('spellPower', 20);
    asc.addAttribute('healPower', 15);
    asc.addAttribute('defense', 5);
    asc.addAttribute('moveSpeed', 100);
    
    // Grant basic abilities
    asc.grantAbility(new BasicAttackAbility());
    asc.grantAbility(new HealAbility());
    
    return asc;
  }

  /**
   * Creates a basic enemy ability system with common attributes
   */
  static createEnemyAbilitySystem(owner: any, config?: {
    health?: number;
    attackPower?: number;
    defense?: number;
    moveSpeed?: number;
  }): AbilitySystemComponent {
    const asc = new AbilitySystemComponent(owner);
    
    // Add enemy attributes with optional config
    asc.addAttribute('health', config?.health || 50, config?.health || 50);
    asc.addAttribute('attackPower', config?.attackPower || 15);
    asc.addAttribute('defense', config?.defense || 2);
    asc.addAttribute('moveSpeed', config?.moveSpeed || 80);
    
    // Grant basic attack
    asc.grantAbility(new BasicAttackAbility());
    
    return asc;
  }

  /**
   * Creates common buff effects
   */
  static createCommonBuffs() {
    return {
      attackBuff: GameplayEffect.createAttributeBuff('attackPower', 10, 10000), // +10 attack for 10 seconds
      defenseBuff: GameplayEffect.createAttributeBuff('defense', 5, 8000), // +5 defense for 8 seconds
      speedBuff: GameplayEffect.createAttributeBuff('moveSpeed', 20, 6000), // +20 speed for 6 seconds
      healthBuff: GameplayEffect.createAttributeBuff('health', 25, 15000), // +25 health for 15 seconds
    };
  }

  /**
   * Creates common damage/healing effects
   */
  static createCommonEffects() {
    return {
      smallDamage: GameplayEffect.createInstantDamage(15),
      mediumDamage: GameplayEffect.createInstantDamage(30),
      largeDamage: GameplayEffect.createInstantDamage(50),
      smallHeal: GameplayEffect.createInstantHeal(20),
      mediumHeal: GameplayEffect.createInstantHeal(40),
      largeHeal: GameplayEffect.createInstantHeal(60),
      poison: GameplayEffect.createDamageOverTime(8, 5000, 1000), // 8 damage per second for 5 seconds
      regeneration: GameplayEffect.createHealOverTime(5, 10000, 1000), // 5 heal per second for 10 seconds
      stun: GameplayEffect.createStun(2000), // 2 second stun
    };
  }

  /**
   * Helper to find nearest enemy for targeting
   */
  static findNearestEnemy(owner: any, enemies: any[], maxRange: number = 200): any | null {
    let nearest: any = null;
    let nearestDistance = Infinity;

    enemies.forEach(enemy => {
      if (enemy.active && enemy !== owner) {
        const distance = Phaser.Math.Distance.Between(
          owner.x, owner.y,
          enemy.x, enemy.y
        );
        
        if (distance <= maxRange && distance < nearestDistance) {
          nearestDistance = distance;
          nearest = enemy;
        }
      }
    });

    return nearest;
  }

  /**
   * Helper to find all enemies in range
   */
  static findEnemiesInRange(owner: any, enemies: any[], range: number): any[] {
    return enemies.filter(enemy => {
      if (!enemy.active || enemy === owner) return false;
      
      const distance = Phaser.Math.Distance.Between(
        owner.x, owner.y,
        enemy.x, enemy.y
      );
      
      return distance <= range;
    });
  }

  /**
   * Debug helper to log ability system state
   */
  static logAbilitySystemState(asc: AbilitySystemComponent, label: string = 'ASC'): void {
    console.group(`${label} Debug Info`);
    console.log('Attributes:', asc.getDebugInfo().attributes);
    console.log('Abilities:', asc.getDebugInfo().abilities);
    console.log('Active Effects:', asc.getDebugInfo().activeEffects);
    console.log('Tags:', asc.getDebugInfo().tags);
    console.log('Cooldowns:', asc.getDebugInfo().cooldowns);
    console.groupEnd();
  }
}