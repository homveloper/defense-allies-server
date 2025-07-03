// Turn-Based Context Enhancement for GAS v2
// Extends the existing AbilityContext with turn-based information

import { EnhancedAbilityContext, AbilityCondition } from '../../types/AbilityTypes';
import { ResourceCost, ResourceUsage } from './TurnBasedResource';
import { TurnOrderEntry } from './InitiativeSystem';

export enum TurnPhase {
  START = 'start',                    // Turn start phase
  MOVEMENT = 'movement',              // Movement phase
  MAIN_ACTION = 'main_action',        // Main action phase
  BONUS_ACTION = 'bonus_action',      // Bonus action phase
  REACTION = 'reaction',              // Reaction phase
  END = 'end'                         // Turn end phase
}

export interface TurnAction {
  actionId: string;
  abilityId: string;
  entityId: string;
  turn: number;
  phase: TurnPhase;
  timestamp: number;
  costs: ResourceUsage[];
  success: boolean;
  metadata?: Record<string, any>;
}

export interface TurnBasedAbilityContext extends EnhancedAbilityContext {
  // Turn Information
  currentTurn: number;
  currentRound: number;
  activePlayer: string;
  phase: TurnPhase;
  
  // Resource Information
  resourceCosts?: ResourceCost[];
  availableResources?: Record<string, number>;
  resourceUsage?: ResourceUsage[];
  
  // Turn Order Information
  turnOrder?: TurnOrderEntry[];
  playerPosition?: number;
  nextPlayer?: string;
  previousPlayer?: string;
  
  // Action History
  turnHistory?: TurnAction[];
  roundHistory?: TurnAction[];
  recentActions?: TurnAction[];
  
  // Turn Constraints
  actionsThisTurn?: number;
  maxActionsThisTurn?: number;
  movementUsed?: number;
  maxMovementThisTurn?: number;
  
  // Phase Constraints
  allowedInPhase?: boolean;
  phaseRestrictions?: string[];
  
  // Timing Information
  chargingTurns?: number;
  channelingTurns?: number;
  cooldownTurns?: number;
  
  // Turn-specific conditions
  turnConditions?: AbilityCondition[];
  phaseConditions?: AbilityCondition[];
  
  // Environmental factors
  environmentEffects?: string[];
  positionModifiers?: Record<string, number>;
  
  // Turn planning
  isPlanned?: boolean;
  plannedTurn?: number;
  plannedPhase?: TurnPhase;
}

export class TurnContextBuilder {
  private context: Partial<TurnBasedAbilityContext> = {};

  constructor(baseContext?: EnhancedAbilityContext) {
    if (baseContext) {
      this.context = { ...baseContext };
    }
  }

  // === BASIC CONTEXT ===

  setOwner(owner: any): this {
    this.context.owner = owner;
    return this;
  }

  setTarget(target: any): this {
    this.context.target = target;
    return this;
  }

  setScene(scene: any): this {
    this.context.scene = scene;
    return this;
  }

  // === TURN INFORMATION ===

  setTurnInfo(turn: number, round: number, activePlayer: string, phase: TurnPhase): this {
    this.context.currentTurn = turn;
    this.context.currentRound = round;
    this.context.activePlayer = activePlayer;
    this.context.phase = phase;
    return this;
  }

  setPhase(phase: TurnPhase): this {
    this.context.phase = phase;
    return this;
  }

  // === RESOURCE INFORMATION ===

  setResourceCosts(costs: ResourceCost[]): this {
    this.context.resourceCosts = costs;
    return this;
  }

  setAvailableResources(resources: Record<string, number>): this {
    this.context.availableResources = resources;
    return this;
  }

  setResourceUsage(usage: ResourceUsage[]): this {
    this.context.resourceUsage = usage;
    return this;
  }

  // === TURN ORDER ===

  setTurnOrder(turnOrder: TurnOrderEntry[], playerPosition: number): this {
    this.context.turnOrder = turnOrder;
    this.context.playerPosition = playerPosition;
    
    // Set next and previous players
    if (turnOrder.length > 0) {
      const nextPos = (playerPosition + 1) % turnOrder.length;
      const prevPos = playerPosition === 0 ? turnOrder.length - 1 : playerPosition - 1;
      
      this.context.nextPlayer = turnOrder[nextPos]?.entityId;
      this.context.previousPlayer = turnOrder[prevPos]?.entityId;
    }
    
    return this;
  }

  // === ACTION HISTORY ===

  setTurnHistory(history: TurnAction[]): this {
    this.context.turnHistory = history;
    return this;
  }

  setRoundHistory(history: TurnAction[]): this {
    this.context.roundHistory = history;
    return this;
  }

  setRecentActions(actions: TurnAction[]): this {
    this.context.recentActions = actions;
    return this;
  }

  // === TURN CONSTRAINTS ===

  setActionConstraints(used: number, max: number): this {
    this.context.actionsThisTurn = used;
    this.context.maxActionsThisTurn = max;
    return this;
  }

  setMovementConstraints(used: number, max: number): this {
    this.context.movementUsed = used;
    this.context.maxMovementThisTurn = max;
    return this;
  }

  // === PHASE CONSTRAINTS ===

  setPhaseAllowed(allowed: boolean): this {
    this.context.allowedInPhase = allowed;
    return this;
  }

  setPhaseRestrictions(restrictions: string[]): this {
    this.context.phaseRestrictions = restrictions;
    return this;
  }

  // === TIMING INFORMATION ===

  setTimingInfo(charging?: number, channeling?: number, cooldown?: number): this {
    this.context.chargingTurns = charging;
    this.context.channelingTurns = channeling;
    this.context.cooldownTurns = cooldown;
    return this;
  }

  // === CONDITIONS ===

  setTurnConditions(conditions: AbilityCondition[]): this {
    this.context.turnConditions = conditions;
    return this;
  }

  setPhaseConditions(conditions: AbilityCondition[]): this {
    this.context.phaseConditions = conditions;
    return this;
  }

  addConditions(conditions: AbilityCondition[]): this {
    if (!this.context.conditions) {
      this.context.conditions = [];
    }
    this.context.conditions.push(...conditions);
    return this;
  }

  // === ENVIRONMENT ===

  setEnvironmentEffects(effects: string[]): this {
    this.context.environmentEffects = effects;
    return this;
  }

  setPositionModifiers(modifiers: Record<string, number>): this {
    this.context.positionModifiers = modifiers;
    return this;
  }

  // === PLANNING ===

  setPlanningInfo(isPlanned: boolean, plannedTurn?: number, plannedPhase?: TurnPhase): this {
    this.context.isPlanned = isPlanned;
    this.context.plannedTurn = plannedTurn;
    this.context.plannedPhase = plannedPhase;
    return this;
  }

  // === METADATA ===

  setMetadata(metadata: Record<string, any>): this {
    this.context.metadata = { ...this.context.metadata, ...metadata };
    return this;
  }

  addMetadata(key: string, value: any): this {
    if (!this.context.metadata) {
      this.context.metadata = {};
    }
    this.context.metadata[key] = value;
    return this;
  }

  // === BUILD ===

  build(): TurnBasedAbilityContext {
    // Validate required fields
    if (!this.context.owner) {
      throw new Error('Owner is required for TurnBasedAbilityContext');
    }

    if (!this.context.scene) {
      throw new Error('Scene is required for TurnBasedAbilityContext');
    }

    if (this.context.currentTurn === undefined) {
      this.context.currentTurn = 1;
    }

    if (this.context.currentRound === undefined) {
      this.context.currentRound = 1;
    }

    if (!this.context.phase) {
      this.context.phase = TurnPhase.MAIN_ACTION;
    }

    return this.context as TurnBasedAbilityContext;
  }

  // === UTILITY METHODS ===

  clone(): TurnContextBuilder {
    const newBuilder = new TurnContextBuilder();
    newBuilder.context = { ...this.context };
    return newBuilder;
  }

  reset(): this {
    this.context = {};
    return this;
  }
}

export class TurnContextValidator {
  static validateContext(context: TurnBasedAbilityContext): { valid: boolean; errors: string[] } {
    const errors: string[] = [];

    // Required fields
    if (!context.owner) {
      errors.push('Owner is required');
    }

    if (!context.scene) {
      errors.push('Scene is required');
    }

    if (context.currentTurn === undefined || context.currentTurn < 1) {
      errors.push('Current turn must be >= 1');
    }

    if (context.currentRound === undefined || context.currentRound < 1) {
      errors.push('Current round must be >= 1');
    }

    // Phase validation
    if (!Object.values(TurnPhase).includes(context.phase)) {
      errors.push(`Invalid phase: ${context.phase}`);
    }

    // Resource validation
    if (context.resourceCosts) {
      for (const cost of context.resourceCosts) {
        if (cost.amount < 0) {
          errors.push(`Resource cost cannot be negative: ${cost.resourceId}`);
        }
      }
    }

    // Action constraints
    if (context.actionsThisTurn !== undefined && context.maxActionsThisTurn !== undefined) {
      if (context.actionsThisTurn > context.maxActionsThisTurn) {
        errors.push('Actions this turn exceeds maximum allowed');
      }
    }

    // Movement constraints
    if (context.movementUsed !== undefined && context.maxMovementThisTurn !== undefined) {
      if (context.movementUsed > context.maxMovementThisTurn) {
        errors.push('Movement used this turn exceeds maximum allowed');
      }
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  static validatePhaseTransition(fromPhase: TurnPhase, toPhase: TurnPhase): boolean {
    const phaseOrder = [
      TurnPhase.START,
      TurnPhase.MOVEMENT,
      TurnPhase.MAIN_ACTION,
      TurnPhase.BONUS_ACTION,
      TurnPhase.REACTION,
      TurnPhase.END
    ];

    const fromIndex = phaseOrder.indexOf(fromPhase);
    const toIndex = phaseOrder.indexOf(toPhase);

    // Can only move forward in phase order or stay in same phase
    return toIndex >= fromIndex;
  }

  static canUseAbilityInPhase(abilityPhases: TurnPhase[], currentPhase: TurnPhase): boolean {
    return abilityPhases.includes(currentPhase);
  }
}

export class TurnContextUtils {
  static extractTurnInfo(context: TurnBasedAbilityContext): {
    turn: number;
    round: number;
    phase: TurnPhase;
    activePlayer: string;
  } {
    return {
      turn: context.currentTurn,
      round: context.currentRound,
      phase: context.phase,
      activePlayer: context.activePlayer
    };
  }

  static extractResourceInfo(context: TurnBasedAbilityContext): {
    costs: ResourceCost[];
    available: Record<string, number>;
    usage: ResourceUsage[];
  } {
    return {
      costs: context.resourceCosts || [],
      available: context.availableResources || {},
      usage: context.resourceUsage || []
    };
  }

  static extractActionHistory(context: TurnBasedAbilityContext): {
    turn: TurnAction[];
    round: TurnAction[];
    recent: TurnAction[];
  } {
    return {
      turn: context.turnHistory || [],
      round: context.roundHistory || [],
      recent: context.recentActions || []
    };
  }

  static isPlayerTurn(context: TurnBasedAbilityContext, playerId: string): boolean {
    return context.activePlayer === playerId;
  }

  static isPhaseAllowed(context: TurnBasedAbilityContext, requiredPhases: TurnPhase[]): boolean {
    return requiredPhases.includes(context.phase);
  }

  static hasResourcesForCost(context: TurnBasedAbilityContext, costs: ResourceCost[]): boolean {
    if (!context.availableResources) return false;

    for (const cost of costs) {
      if (cost.required) {
        const available = context.availableResources[cost.resourceId] || 0;
        if (available < cost.amount) {
          return false;
        }
      }
    }

    return true;
  }

  static getRemainingActions(context: TurnBasedAbilityContext): number {
    const used = context.actionsThisTurn || 0;
    const max = context.maxActionsThisTurn || 1;
    return Math.max(0, max - used);
  }

  static getRemainingMovement(context: TurnBasedAbilityContext): number {
    const used = context.movementUsed || 0;
    const max = context.maxMovementThisTurn || 0;
    return Math.max(0, max - used);
  }

  static createActionRecord(
    context: TurnBasedAbilityContext,
    abilityId: string,
    success: boolean,
    costs: ResourceUsage[] = []
  ): TurnAction {
    return {
      actionId: `action_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
      abilityId,
      entityId: context.activePlayer,
      turn: context.currentTurn,
      phase: context.phase,
      timestamp: Date.now(),
      costs,
      success,
      metadata: context.metadata
    };
  }
}