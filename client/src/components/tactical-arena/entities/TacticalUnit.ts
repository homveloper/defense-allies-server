// Tactical Unit - Core unit class for tactical combat
// Integrates with GAS v2 turn-based systems

import { v2 } from '../../../../packages/gas';
import { TurnPhase, TurnBasedAbilityContext } from '../../../../packages/gas/v2/turn-based/TurnBasedContext';

export interface TacticalUnitStats {
  maxHealth: number;
  health: number;
  armor: number;
  accuracy: number;
  movement: number;
  initiative: number;
  cover: number; // Cover bonus percentage
}

export interface Position {
  x: number;
  y: number;
}

export interface StatusEffect {
  id: string;
  name: string;
  duration: number;
  type: 'buff' | 'debuff' | 'neutral';
  description: string;
}

export interface UnitAction {
  id: string;
  name: string;
  description: string;
  phase: TurnPhase[];
  cost: Record<string, number>;
  range: number;
  canUse: boolean;
  cooldown?: number;
  requiresTarget?: boolean;
  targetType?: 'enemy' | 'ally' | 'tile' | 'self';
}

export type UnitFaction = 'player' | 'enemy' | 'neutral';

export class TacticalUnit {
  public readonly id: string;
  public readonly name: string;
  public readonly faction: UnitFaction;
  public position: Position;
  public stats: TacticalUnitStats;
  public statusEffects: StatusEffect[] = [];
  
  // GAS Integration
  public abilitySystem: v2.AbilitySystemComponent;
  
  // Visual properties
  public color: string;
  public size: number = 30;
  public isSelected: boolean = false;
  public isTargeted: boolean = false;
  
  // Tactical state
  public facing: 'north' | 'south' | 'east' | 'west' = 'north';
  public inCover: boolean = false;
  public hasMovedThisTurn: boolean = false;
  public overwatchActive: boolean = false;

  constructor(
    id: string,
    name: string,
    faction: UnitFaction,
    initialPosition: Position,
    stats: TacticalUnitStats
  ) {
    this.id = id;
    this.name = name;
    this.faction = faction;
    this.position = { ...initialPosition };
    this.stats = { ...stats };
    
    // Set faction colors
    this.color = {
      player: '#4A90E2',
      enemy: '#E24A4A',
      neutral: '#F5A623'
    }[faction];

    // Initialize GAS
    this.abilitySystem = new v2.AbilitySystemComponent(this);
    this.setupAbilities();
    this.setupResources();
  }

  // === GAS SETUP ===

  private setupResources(): void {
    // Add tactical resources
    this.abilitySystem.addAttribute('health', this.stats.health, this.stats.maxHealth);
    this.abilitySystem.addAttribute('armor', this.stats.armor);
    this.abilitySystem.addAttribute('accuracy', this.stats.accuracy);
    
    // Add turn-based resources using the resource manager
    const resourceManager = (this.abilitySystem as any).resourceManager;
    if (resourceManager) {
      resourceManager.addResource({
        id: 'action_points',
        name: 'Action Points',
        baseValue: 2,
        maxValue: 2,
        refreshPattern: v2.ResourceRefreshPattern.PER_TURN,
        refreshAmount: 2
      });

      resourceManager.addResource({
        id: 'movement_points',
        name: 'Movement Points',
        baseValue: this.stats.movement,
        maxValue: this.stats.movement,
        refreshPattern: v2.ResourceRefreshPattern.PER_TURN,
        refreshAmount: this.stats.movement
      });

      resourceManager.addResource({
        id: 'reaction_points',
        name: 'Reaction Points',
        baseValue: 1,
        maxValue: 1,
        refreshPattern: v2.ResourceRefreshPattern.PER_TURN,
        refreshAmount: 1
      });
    }

    // Add faction tags
    this.abilitySystem.addTag(this.faction);
    this.abilitySystem.addTag('unit');
    this.abilitySystem.addTag('alive');
  }

  private setupAbilities(): void {
    // Grant basic tactical abilities
    this.abilitySystem.grantAbility(new MoveAbility());
    this.abilitySystem.grantAbility(new AttackAbility());
    this.abilitySystem.grantAbility(new AimedShotAbility());
    this.abilitySystem.grantAbility(new OverwatchAbility());
    this.abilitySystem.grantAbility(new TakeCoverAbility());
    this.abilitySystem.grantAbility(new ReloadAbility());
  }

  // === UNIT ACTIONS ===

  getAvailableActions(currentPhase: TurnPhase): UnitAction[] {
    const actions: UnitAction[] = [];
    
    // Get all tactical abilities manually
    const tacticalAbilities = [
      this.abilitySystem.getAbility('move'),
      this.abilitySystem.getAbility('attack'),
      this.abilitySystem.getAbility('aimed_shot'),
      this.abilitySystem.getAbility('overwatch'),
      this.abilitySystem.getAbility('take_cover'),
      this.abilitySystem.getAbility('reload')
    ].filter(Boolean) as TacticalAbility[];

    for (const ability of tacticalAbilities) {
      const action = ability.getActionInfo(this, currentPhase);
      if (action) {
        actions.push(action);
      }
    }

    return actions;
  }

  async executeAction(actionId: string, target?: TacticalUnit | Position): Promise<boolean> {
    const ability = this.abilitySystem.getAbility(actionId);
    if (!ability) return false;

    const context: TurnBasedAbilityContext = {
      owner: this,
      target,
      scene: null as any, // Would be provided by the game engine
      currentTurn: 1, // Would be provided by turn manager
      currentRound: 1,
      activePlayer: this.id,
      phase: TurnPhase.MAIN_ACTION // Would be current phase
    };

    const result = await this.abilitySystem.tryActivateAbility(actionId, context);
    return result.success;
  }

  // === COMBAT METHODS ===

  takeDamage(amount: number, _source?: TacticalUnit): number {
    // Apply armor reduction
    const armorReduction = Math.min(amount * 0.1 * this.stats.armor, amount * 0.8);
    const finalDamage = Math.max(1, amount - armorReduction);
    
    this.stats.health = Math.max(0, this.stats.health - finalDamage);
    this.abilitySystem.setAttributeValue('health', this.stats.health);

    // Check if unit died
    if (this.stats.health <= 0) {
      this.onDeath();
    }

    return finalDamage;
  }

  heal(amount: number): number {
    const oldHealth = this.stats.health;
    this.stats.health = Math.min(this.stats.maxHealth, this.stats.health + amount);
    this.abilitySystem.setAttributeValue('health', this.stats.health);
    
    return this.stats.health - oldHealth;
  }

  // === POSITIONING ===

  moveTo(newPosition: Position): boolean {
    // Check if movement is valid (would check map bounds, obstacles, etc.)
    this.position = { ...newPosition };
    this.hasMovedThisTurn = true;
    return true;
  }

  getDistanceTo(target: Position): number {
    return Math.abs(this.position.x - target.x) + Math.abs(this.position.y - target.y);
  }

  isInRange(target: Position, range: number): boolean {
    return this.getDistanceTo(target) <= range;
  }

  // === STATUS EFFECTS ===

  addStatusEffect(effect: StatusEffect): void {
    // Remove existing effect of same type
    this.statusEffects = this.statusEffects.filter(e => e.id !== effect.id);
    this.statusEffects.push({ ...effect });
    
    // Apply effect tags
    this.abilitySystem.addTag(effect.id);
  }

  removeStatusEffect(effectId: string): boolean {
    const index = this.statusEffects.findIndex(e => e.id === effectId);
    if (index === -1) return false;

    this.statusEffects.splice(index, 1);
    this.abilitySystem.removeTag(effectId);
    return true;
  }

  processStatusEffects(): void {
    for (let i = this.statusEffects.length - 1; i >= 0; i--) {
      const effect = this.statusEffects[i];
      effect.duration--;
      
      if (effect.duration <= 0) {
        this.removeStatusEffect(effect.id);
      }
    }
  }

  // === TURN PROCESSING ===

  startTurn(): void {
    this.hasMovedThisTurn = false;
    this.overwatchActive = false;
    
    // Process GAS turn start
    const resourceManager = (this.abilitySystem as any).resourceManager;
    if (resourceManager) {
      resourceManager.processTurnStart();
    }
    
    // Process status effects
    this.processStatusEffects();
  }

  endTurn(): void {
    // Any end-of-turn processing
  }

  // === UTILITY METHODS ===

  isAlive(): boolean {
    return this.stats.health > 0;
  }

  isEnemy(other: TacticalUnit): boolean {
    return this.faction !== other.faction && other.faction !== 'neutral';
  }

  canSee(target: Position): boolean {
    // Simple line-of-sight check (would be more complex in real implementation)
    return this.getDistanceTo(target) <= 8; // Max sight range
  }

  getResourceSummary(): Record<string, { current: number; max: number }> {
    const resources: Record<string, { current: number; max: number }> = {};
    
    // Get from resource manager if available
    const resourceManager = (this.abilitySystem as any).resourceManager;
    if (resourceManager) {
      const summary = resourceManager.getResourceSummary();
      for (const [id, info] of Object.entries(summary)) {
        const resourceInfo = info as { current: number; max: number };
        resources[id] = {
          current: resourceInfo.current,
          max: resourceInfo.max
        };
      }
    }
    
    return resources;
  }

  // === EVENTS ===

  private onDeath(): void {
    this.abilitySystem.removeTag('alive');
    this.abilitySystem.addTag('dead');
    
    // Clear all status effects
    this.statusEffects = [];
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    return {
      id: this.id,
      name: this.name,
      faction: this.faction,
      position: this.position,
      stats: this.stats,
      statusEffects: this.statusEffects,
      facing: this.facing,
      inCover: this.inCover,
      hasMovedThisTurn: this.hasMovedThisTurn,
      overwatchActive: this.overwatchActive
    };
  }

  deserialize(data: Record<string, any>): void {
    this.position = data.position || this.position;
    this.stats = { ...this.stats, ...data.stats };
    this.statusEffects = data.statusEffects || [];
    this.facing = data.facing || this.facing;
    this.inCover = data.inCover || false;
    this.hasMovedThisTurn = data.hasMovedThisTurn || false;
    this.overwatchActive = data.overwatchActive || false;
  }
}

// === TACTICAL ABILITIES ===

abstract class TacticalAbility extends v2.GameplayAbility {
  abstract getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null;
}

class MoveAbility extends TacticalAbility {
  readonly id = 'move';
  readonly name = 'Move';
  readonly description = 'Move to a new position';
  readonly cooldown = 0;

  getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null {
    if (![TurnPhase.MOVEMENT, TurnPhase.MAIN_ACTION].includes(phase)) {
      return null;
    }

    const resources = unit.getResourceSummary();
    const canUse = (resources.movement_points?.current || 0) > 0;

    return {
      id: this.id,
      name: this.name,
      description: this.description,
      phase: [TurnPhase.MOVEMENT, TurnPhase.MAIN_ACTION],
      cost: { movement_points: 1 },
      range: 1,
      canUse,
      requiresTarget: true,
      targetType: 'tile'
    };
  }

  canActivate(_context: any): boolean {
    return true;
  }

  async activate(context: any): Promise<boolean> {
    const unit = context.owner as TacticalUnit;
    const target = context.target as Position;
    
    if (unit.isInRange(target, 1)) {
      return unit.moveTo(target);
    }
    
    return false;
  }

  getCooldownRemaining(): number {
    return 0;
  }
}

class AttackAbility extends TacticalAbility {
  readonly id = 'attack';
  readonly name = 'Attack';
  readonly description = 'Basic ranged attack';
  readonly cooldown = 0;

  getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null {
    if (![TurnPhase.MAIN_ACTION].includes(phase)) {
      return null;
    }

    const resources = unit.getResourceSummary();
    const canUse = (resources.action_points?.current || 0) >= 1;

    return {
      id: this.id,
      name: this.name,
      description: this.description,
      phase: [TurnPhase.MAIN_ACTION],
      cost: { action_points: 1 },
      range: 4,
      canUse,
      requiresTarget: true,
      targetType: 'enemy'
    };
  }

  canActivate(context: any): boolean {
    const unit = context.owner as TacticalUnit;
    const target = context.target as TacticalUnit;
    
    return unit.isEnemy(target) && unit.isInRange(target.position, 4);
  }

  async activate(context: any): Promise<boolean> {
    const attacker = context.owner as TacticalUnit;
    const target = context.target as TacticalUnit;
    
    // Calculate hit chance
    const baseAccuracy = attacker.stats.accuracy;
    const coverPenalty = target.inCover ? 25 : 0;
    const finalAccuracy = Math.max(5, baseAccuracy - coverPenalty);
    
    const hitRoll = Math.random() * 100;
    
    if (hitRoll <= finalAccuracy) {
      // Hit! Calculate damage
      const baseDamage = 25 + Math.random() * 15; // 25-40 damage
      const actualDamage = target.takeDamage(baseDamage, attacker);
      console.log(`${attacker.name} hits ${target.name} for ${actualDamage} damage!`);
      return true;
    } else {
      console.log(`${attacker.name} misses ${target.name}!`);
      return true; // Still consumes the action
    }
  }

  getCooldownRemaining(): number {
    return 0;
  }
}

class AimedShotAbility extends TacticalAbility {
  readonly id = 'aimed_shot';
  readonly name = 'Aimed Shot';
  readonly description = 'High accuracy shot that uses 2 action points';
  readonly cooldown = 0;

  getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null {
    if (![TurnPhase.MAIN_ACTION].includes(phase)) {
      return null;
    }

    const resources = unit.getResourceSummary();
    const canUse = (resources.action_points?.current || 0) >= 2;

    return {
      id: this.id,
      name: this.name,
      description: this.description,
      phase: [TurnPhase.MAIN_ACTION],
      cost: { action_points: 2 },
      range: 6,
      canUse,
      requiresTarget: true,
      targetType: 'enemy'
    };
  }

  canActivate(context: any): boolean {
    const unit = context.owner as TacticalUnit;
    const target = context.target as TacticalUnit;
    
    return unit.isEnemy(target) && unit.isInRange(target.position, 6);
  }

  async activate(context: any): Promise<boolean> {
    const attacker = context.owner as TacticalUnit;
    const target = context.target as TacticalUnit;
    
    // Higher accuracy than basic attack
    const baseAccuracy = attacker.stats.accuracy + 30;
    const coverPenalty = target.inCover ? 15 : 0; // Less cover penalty
    const finalAccuracy = Math.max(10, baseAccuracy - coverPenalty);
    
    const hitRoll = Math.random() * 100;
    
    if (hitRoll <= finalAccuracy) {
      const baseDamage = 35 + Math.random() * 20; // 35-55 damage
      const actualDamage = target.takeDamage(baseDamage, attacker);
      console.log(`${attacker.name} lands an aimed shot on ${target.name} for ${actualDamage} damage!`);
      return true;
    } else {
      console.log(`${attacker.name} misses their aimed shot on ${target.name}!`);
      return true;
    }
  }

  getCooldownRemaining(): number {
    return 0;
  }
}

class OverwatchAbility extends TacticalAbility {
  readonly id = 'overwatch';
  readonly name = 'Overwatch';
  readonly description = 'Prepare to shoot at moving enemies';
  readonly cooldown = 0;

  getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null {
    if (![TurnPhase.MAIN_ACTION].includes(phase)) {
      return null;
    }

    const resources = unit.getResourceSummary();
    const canUse = (resources.action_points?.current || 0) >= 1 && !unit.overwatchActive;

    return {
      id: this.id,
      name: this.name,
      description: this.description,
      phase: [TurnPhase.MAIN_ACTION],
      cost: { action_points: 1 },
      range: 0,
      canUse,
      requiresTarget: false,
      targetType: 'self'
    };
  }

  canActivate(context: any): boolean {
    const unit = context.owner as TacticalUnit;
    return !unit.overwatchActive;
  }

  async activate(context: any): Promise<boolean> {
    const unit = context.owner as TacticalUnit;
    unit.overwatchActive = true;
    console.log(`${unit.name} is now on overwatch!`);
    return true;
  }

  getCooldownRemaining(): number {
    return 0;
  }
}

class TakeCoverAbility extends TacticalAbility {
  readonly id = 'take_cover';
  readonly name = 'Take Cover';
  readonly description = 'Gain cover bonus until next turn';
  readonly cooldown = 0;

  getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null {
    if (![TurnPhase.BONUS_ACTION].includes(phase)) {
      return null;
    }

    const resources = unit.getResourceSummary();
    const canUse = (resources.action_points?.current || 0) >= 0 && !unit.inCover;

    return {
      id: this.id,
      name: this.name,
      description: this.description,
      phase: [TurnPhase.BONUS_ACTION],
      cost: {},
      range: 0,
      canUse,
      requiresTarget: false,
      targetType: 'self'
    };
  }

  canActivate(context: any): boolean {
    const unit = context.owner as TacticalUnit;
    return !unit.inCover;
  }

  async activate(context: any): Promise<boolean> {
    const unit = context.owner as TacticalUnit;
    unit.inCover = true;
    console.log(`${unit.name} takes cover!`);
    return true;
  }

  getCooldownRemaining(): number {
    return 0;
  }
}

class ReloadAbility extends TacticalAbility {
  readonly id = 'reload';
  readonly name = 'Reload';
  readonly description = 'Quick reload action';
  readonly cooldown = 0;

  getActionInfo(unit: TacticalUnit, phase: TurnPhase): UnitAction | null {
    if (![TurnPhase.BONUS_ACTION].includes(phase)) {
      return null;
    }

    return {
      id: this.id,
      name: this.name,
      description: this.description,
      phase: [TurnPhase.BONUS_ACTION],
      cost: {},
      range: 0,
      canUse: true,
      requiresTarget: false,
      targetType: 'self'
    };
  }

  canActivate(_context: any): boolean {
    return true;
  }

  async activate(context: any): Promise<boolean> {
    const unit = context.owner as TacticalUnit;
    console.log(`${unit.name} reloads!`);
    return true;
  }

  getCooldownRemaining(): number {
    return 0;
  }
}