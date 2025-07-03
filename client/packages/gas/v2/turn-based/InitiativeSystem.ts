// Initiative and Turn Order System for GAS v2
// Manages turn order, initiative calculation, and turn sequence

import { GameplayAttribute } from '../GameplayAttribute';
import { AttributeModifier } from '../../types/AbilityTypes';

export interface InitiativeConfig {
  baseInitiative: number;
  attribute?: string;              // Attribute to use for initiative (e.g., 'dexterity', 'speed')
  randomBonus?: number;           // Random bonus range (0 to randomBonus)
  tieBreaker?: number;            // Manual tie breaker value
  modifiers?: AttributeModifier[]; // Initiative modifiers
}

export interface TurnOrderEntry {
  entityId: string;
  initiative: number;
  baseInitiative: number;
  calculatedInitiative: number;
  position: number;               // Position in turn order (0-based)
  isDelayed: boolean;            // Has delayed their turn
  delayedToPosition?: number;    // Position they delayed to
}

export interface DelayedAction {
  entityId: string;
  originalPosition: number;
  delayedToPosition: number;
  delayedToTurn?: number;        // If delaying to next turn
}

export enum InitiativeRollType {
  ONCE_PER_COMBAT = 'once_per_combat',    // Roll once, use for entire combat
  ONCE_PER_ROUND = 'once_per_round',      // Re-roll each round
  DYNAMIC = 'dynamic'                      // Can change during combat
}

export class InitiativeCalculator {
  private rollType: InitiativeRollType;
  private useRandomness: boolean;

  constructor(
    rollType: InitiativeRollType = InitiativeRollType.ONCE_PER_COMBAT,
    useRandomness: boolean = true
  ) {
    this.rollType = rollType;
    this.useRandomness = useRandomness;
  }

  calculateInitiative(
    entityId: string,
    config: InitiativeConfig,
    attributeValue?: number
  ): number {
    let initiative = config.baseInitiative;

    // Add attribute bonus if specified
    if (config.attribute && attributeValue !== undefined) {
      initiative += attributeValue;
    }

    // Apply modifiers
    if (config.modifiers) {
      for (const modifier of config.modifiers) {
        switch (modifier.operation) {
          case 'add':
            initiative += modifier.magnitude;
            break;
          case 'multiply':
            initiative *= modifier.magnitude;
            break;
          case 'override':
            initiative = modifier.magnitude;
            break;
        }
      }
    }

    // Add random bonus
    if (this.useRandomness && config.randomBonus && config.randomBonus > 0) {
      initiative += Math.random() * config.randomBonus;
    }

    return Math.max(0, initiative);
  }

  shouldRerollInitiative(currentRound: number): boolean {
    return this.rollType === InitiativeRollType.ONCE_PER_ROUND && currentRound > 1;
  }

  canChangeInitiative(): boolean {
    return this.rollType === InitiativeRollType.DYNAMIC;
  }
}

export class TurnOrderManager {
  private turnOrder: TurnOrderEntry[] = [];
  private currentPosition: number = 0;
  private currentRound: number = 1;
  private currentTurn: number = 1;
  private delayedActions: DelayedAction[] = [];
  private initiativeCalculator: InitiativeCalculator;

  constructor(calculator?: InitiativeCalculator) {
    this.initiativeCalculator = calculator || new InitiativeCalculator();
  }

  // === INITIALIZATION ===

  addEntity(
    entityId: string, 
    config: InitiativeConfig, 
    attributeValue?: number
  ): TurnOrderEntry {
    const initiative = this.initiativeCalculator.calculateInitiative(
      entityId, 
      config, 
      attributeValue
    );

    const entry: TurnOrderEntry = {
      entityId,
      initiative,
      baseInitiative: config.baseInitiative,
      calculatedInitiative: initiative,
      position: -1, // Will be set when order is determined
      isDelayed: false
    };

    this.turnOrder.push(entry);
    this.sortTurnOrder();
    
    return entry;
  }

  removeEntity(entityId: string): boolean {
    const index = this.turnOrder.findIndex(entry => entry.entityId === entityId);
    if (index === -1) return false;

    // If removing current entity, adjust position
    const removedEntry = this.turnOrder[index];
    if (removedEntry.position <= this.currentPosition) {
      this.currentPosition = Math.max(0, this.currentPosition - 1);
    }

    this.turnOrder.splice(index, 1);
    this.updatePositions();
    
    return true;
  }

  private sortTurnOrder(): void {
    this.turnOrder.sort((a, b) => {
      // Higher initiative goes first
      if (b.initiative !== a.initiative) {
        return b.initiative - a.initiative;
      }
      
      // Tie breaker
      const aConfig = this.getEntityConfig(a.entityId);
      const bConfig = this.getEntityConfig(b.entityId);
      
      if (aConfig?.tieBreaker !== undefined && bConfig?.tieBreaker !== undefined) {
        return bConfig.tieBreaker - aConfig.tieBreaker;
      }
      
      // Fall back to entity ID for consistent ordering
      return a.entityId.localeCompare(b.entityId);
    });

    this.updatePositions();
  }

  private updatePositions(): void {
    this.turnOrder.forEach((entry, index) => {
      entry.position = index;
    });
  }

  private getEntityConfig(entityId: string): InitiativeConfig | undefined {
    // This would need to be provided by the implementation
    // For now, return undefined
    return undefined;
  }

  // === TURN MANAGEMENT ===

  getCurrentActiveEntity(): string | null {
    if (this.turnOrder.length === 0) return null;
    
    const activeEntry = this.turnOrder[this.currentPosition];
    return activeEntry ? activeEntry.entityId : null;
  }

  getCurrentTurnInfo(): { entityId: string | null; position: number; round: number; turn: number } {
    return {
      entityId: this.getCurrentActiveEntity(),
      position: this.currentPosition,
      round: this.currentRound,
      turn: this.currentTurn
    };
  }

  nextTurn(): { entityId: string | null; newRound: boolean } {
    if (this.turnOrder.length === 0) {
      return { entityId: null, newRound: false };
    }

    this.currentPosition++;
    this.currentTurn++;
    
    let newRound = false;
    
    // Check if round ended
    if (this.currentPosition >= this.turnOrder.length) {
      this.currentPosition = 0;
      this.currentRound++;
      newRound = true;
      
      // Process delayed actions
      this.processDelayedActions();
      
      // Re-roll initiative if needed
      if (this.initiativeCalculator.shouldRerollInitiative(this.currentRound)) {
        this.rerollInitiative();
      }
    }

    return {
      entityId: this.getCurrentActiveEntity(),
      newRound
    };
  }

  // === DELAY ACTIONS ===

  delayTurn(entityId: string, positions: number): boolean {
    const currentEntry = this.turnOrder.find(
      entry => entry.entityId === entityId && entry.position === this.currentPosition
    );
    
    if (!currentEntry) return false;

    const targetPosition = Math.min(
      this.turnOrder.length - 1,
      this.currentPosition + positions
    );

    // Create delayed action
    const delayedAction: DelayedAction = {
      entityId,
      originalPosition: this.currentPosition,
      delayedToPosition: targetPosition
    };

    // If delaying past end of round, move to next round
    if (targetPosition >= this.turnOrder.length) {
      delayedAction.delayedToTurn = this.currentTurn + (this.turnOrder.length - this.currentPosition);
      delayedAction.delayedToPosition = targetPosition - this.turnOrder.length;
    }

    this.delayedActions.push(delayedAction);
    currentEntry.isDelayed = true;
    currentEntry.delayedToPosition = targetPosition;

    // Move to next entity immediately
    return true;
  }

  private processDelayedActions(): void {
    // Sort delayed actions by their target position
    this.delayedActions.sort((a, b) => a.delayedToPosition - b.delayedToPosition);

    // Insert delayed actions back into turn order
    for (const delayedAction of this.delayedActions) {
      const entry = this.turnOrder.find(e => e.entityId === delayedAction.entityId);
      if (entry) {
        entry.isDelayed = false;
        entry.delayedToPosition = undefined;
        
        // Move entry to new position
        const currentIndex = this.turnOrder.indexOf(entry);
        this.turnOrder.splice(currentIndex, 1);
        this.turnOrder.splice(delayedAction.delayedToPosition, 0, entry);
      }
    }

    this.delayedActions = [];
    this.updatePositions();
  }

  // === INITIATIVE CHANGES ===

  rerollInitiative(): void {
    // This would need entity configs to recalculate
    // For now, just re-sort existing order
    this.sortTurnOrder();
  }

  changeInitiative(entityId: string, newInitiative: number): boolean {
    if (!this.initiativeCalculator.canChangeInitiative()) {
      return false;
    }

    const entry = this.turnOrder.find(e => e.entityId === entityId);
    if (!entry) return false;

    entry.initiative = newInitiative;
    entry.calculatedInitiative = newInitiative;
    
    this.sortTurnOrder();
    return true;
  }

  // === GETTERS ===

  getTurnOrder(): TurnOrderEntry[] {
    return [...this.turnOrder]; // Return copy
  }

  getEntityPosition(entityId: string): number {
    const entry = this.turnOrder.find(e => e.entityId === entityId);
    return entry ? entry.position : -1;
  }

  getEntitiesInRange(startPosition: number, endPosition: number): string[] {
    return this.turnOrder
      .slice(startPosition, endPosition + 1)
      .map(entry => entry.entityId);
  }

  getNextEntity(entityId: string): string | null {
    const currentPos = this.getEntityPosition(entityId);
    if (currentPos === -1) return null;

    const nextPos = (currentPos + 1) % this.turnOrder.length;
    return this.turnOrder[nextPos]?.entityId || null;
  }

  getPreviousEntity(entityId: string): string | null {
    const currentPos = this.getEntityPosition(entityId);
    if (currentPos === -1) return null;

    const prevPos = currentPos === 0 ? this.turnOrder.length - 1 : currentPos - 1;
    return this.turnOrder[prevPos]?.entityId || null;
  }

  // === ROUND MANAGEMENT ===

  getCurrentRound(): number {
    return this.currentRound;
  }

  getCurrentTurn(): number {
    return this.currentTurn;
  }

  isFirstTurnOfRound(): boolean {
    return this.currentPosition === 0;
  }

  isLastTurnOfRound(): boolean {
    return this.currentPosition === this.turnOrder.length - 1;
  }

  // === UTILITY METHODS ===

  getInitiativeSummary(): Record<string, any> {
    return {
      currentRound: this.currentRound,
      currentTurn: this.currentTurn,
      currentPosition: this.currentPosition,
      activeEntity: this.getCurrentActiveEntity(),
      turnOrder: this.turnOrder.map(entry => ({
        entityId: entry.entityId,
        initiative: entry.initiative,
        position: entry.position,
        isDelayed: entry.isDelayed,
        isActive: entry.position === this.currentPosition
      })),
      delayedActions: this.delayedActions.length,
      nextEntity: this.turnOrder[(this.currentPosition + 1) % this.turnOrder.length]?.entityId
    };
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    return {
      turnOrder: this.turnOrder,
      currentPosition: this.currentPosition,
      currentRound: this.currentRound,
      currentTurn: this.currentTurn,
      delayedActions: this.delayedActions
    };
  }

  deserialize(data: Record<string, any>): void {
    this.turnOrder = data.turnOrder || [];
    this.currentPosition = data.currentPosition || 0;
    this.currentRound = data.currentRound || 1;
    this.currentTurn = data.currentTurn || 1;
    this.delayedActions = data.delayedActions || [];
  }

  // === CLEANUP ===

  reset(): void {
    this.turnOrder = [];
    this.currentPosition = 0;
    this.currentRound = 1;
    this.currentTurn = 1;
    this.delayedActions = [];
  }

  clear(): void {
    this.reset();
  }
}