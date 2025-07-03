// Turn-Based Resource System for GAS v2
// Manages action points, turn-based resources, and resource refresh patterns

export enum ResourceRefreshPattern {
  PER_TURN = 'per-turn',           // Refresh every turn
  PER_ROUND = 'per-round',         // Refresh every round (all players)
  MANUAL = 'manual',               // Manual refresh only
  COOLDOWN_TURNS = 'cooldown-turns' // Refresh after X turns
}

export interface TurnBasedResourceConfig {
  id: string;
  name: string;
  baseValue: number;
  maxValue: number;
  refreshPattern: ResourceRefreshPattern;
  refreshAmount: number;           // How much to restore
  maxPerTurn?: number;            // Max usage per turn
  carryOverRatio?: number;        // Ratio carried to next turn (0-1)
  cooldownTurns?: number;         // For COOLDOWN_TURNS pattern
}

export interface ResourceCost {
  resourceId: string;
  amount: number;
  required: boolean;              // If false, will consume if available but not block
}

export interface ResourceUsage {
  resourceId: string;
  used: number;
  remaining: number;
  maxThisTurn: number;
}

export class TurnBasedResource {
  public readonly id: string;
  public readonly name: string;
  public baseValue: number;
  public maxValue: number;
  public currentValue: number;
  public readonly refreshPattern: ResourceRefreshPattern;
  public readonly refreshAmount: number;
  public readonly maxPerTurn: number;
  public readonly carryOverRatio: number;
  public readonly cooldownTurns: number;

  private usedThisTurn: number = 0;
  private lastRefreshTurn: number = 0;
  private cooldownRemaining: number = 0;

  constructor(config: TurnBasedResourceConfig) {
    this.id = config.id;
    this.name = config.name;
    this.baseValue = config.baseValue;
    this.maxValue = config.maxValue;
    this.currentValue = config.baseValue;
    this.refreshPattern = config.refreshPattern;
    this.refreshAmount = config.refreshAmount;
    this.maxPerTurn = config.maxPerTurn ?? config.baseValue;
    this.carryOverRatio = config.carryOverRatio ?? 0;
    this.cooldownTurns = config.cooldownTurns ?? 0;
  }

  // === RESOURCE MANAGEMENT ===

  canAfford(amount: number): boolean {
    return this.currentValue >= amount && 
           this.getRemainingThisTurn() >= amount;
  }

  consume(amount: number): boolean {
    if (!this.canAfford(amount)) {
      return false;
    }

    this.currentValue -= amount;
    this.usedThisTurn += amount;
    return true;
  }

  restore(amount: number): void {
    this.currentValue = Math.min(this.maxValue, this.currentValue + amount);
  }

  getRemainingThisTurn(): number {
    return Math.min(
      this.currentValue,
      this.maxPerTurn - this.usedThisTurn
    );
  }

  getUsedThisTurn(): number {
    return this.usedThisTurn;
  }

  // === TURN PROCESSING ===

  processTurnStart(currentTurn: number): void {
    // Carry over unused resources if configured
    if (this.carryOverRatio > 0) {
      const unused = this.maxPerTurn - this.usedThisTurn;
      const carryOver = Math.floor(unused * this.carryOverRatio);
      this.restore(carryOver);
    }

    // Reset turn usage
    this.usedThisTurn = 0;

    // Handle refresh patterns
    this.processRefresh(currentTurn);
  }

  processRoundStart(currentRound: number): void {
    if (this.refreshPattern === ResourceRefreshPattern.PER_ROUND) {
      this.refresh();
    }
  }

  private processRefresh(currentTurn: number): void {
    switch (this.refreshPattern) {
      case ResourceRefreshPattern.PER_TURN:
        this.refresh();
        break;

      case ResourceRefreshPattern.COOLDOWN_TURNS:
        if (this.cooldownRemaining > 0) {
          this.cooldownRemaining--;
          if (this.cooldownRemaining === 0) {
            this.refresh();
          }
        }
        break;

      case ResourceRefreshPattern.MANUAL:
        // No automatic refresh
        break;
    }

    this.lastRefreshTurn = currentTurn;
  }

  private refresh(): void {
    this.currentValue = Math.min(
      this.maxValue,
      this.currentValue + this.refreshAmount
    );
    this.cooldownRemaining = this.cooldownTurns;
  }

  // === MANUAL OPERATIONS ===

  manualRefresh(): void {
    this.refresh();
  }

  setCooldown(turns: number): void {
    this.cooldownRemaining = Math.max(0, turns);
  }

  // === GETTERS ===

  getPercentage(): number {
    return this.maxValue > 0 ? (this.currentValue / this.maxValue) * 100 : 0;
  }

  isOnCooldown(): boolean {
    return this.cooldownRemaining > 0;
  }

  getCooldownRemaining(): number {
    return this.cooldownRemaining;
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    return {
      id: this.id,
      currentValue: this.currentValue,
      usedThisTurn: this.usedThisTurn,
      lastRefreshTurn: this.lastRefreshTurn,
      cooldownRemaining: this.cooldownRemaining
    };
  }

  deserialize(data: Record<string, any>): void {
    this.currentValue = data.currentValue ?? this.baseValue;
    this.usedThisTurn = data.usedThisTurn ?? 0;
    this.lastRefreshTurn = data.lastRefreshTurn ?? 0;
    this.cooldownRemaining = data.cooldownRemaining ?? 0;
  }
}

export class TurnBasedResourceManager {
  private resources: Map<string, TurnBasedResource> = new Map();
  private currentTurn: number = 0;
  private currentRound: number = 0;

  // === RESOURCE REGISTRATION ===

  addResource(config: TurnBasedResourceConfig): TurnBasedResource {
    const resource = new TurnBasedResource(config);
    this.resources.set(config.id, resource);
    return resource;
  }

  removeResource(resourceId: string): boolean {
    return this.resources.delete(resourceId);
  }

  getResource(resourceId: string): TurnBasedResource | undefined {
    return this.resources.get(resourceId);
  }

  getAllResources(): TurnBasedResource[] {
    return Array.from(this.resources.values());
  }

  // === COST MANAGEMENT ===

  canAfford(costs: ResourceCost[]): boolean {
    for (const cost of costs) {
      if (!cost.required) continue;

      const resource = this.resources.get(cost.resourceId);
      if (!resource || !resource.canAfford(cost.amount)) {
        return false;
      }
    }
    return true;
  }

  consumeResources(costs: ResourceCost[]): ResourceUsage[] {
    // First check if we can afford all required costs
    if (!this.canAfford(costs)) {
      return [];
    }

    const usage: ResourceUsage[] = [];
    
    for (const cost of costs) {
      const resource = this.resources.get(cost.resourceId);
      if (!resource) continue;

      const beforeValue = resource.currentValue;
      const consumed = resource.consume(cost.amount);
      
      if (consumed || !cost.required) {
        const actualUsed = consumed ? cost.amount : 0;
        usage.push({
          resourceId: cost.resourceId,
          used: actualUsed,
          remaining: resource.currentValue,
          maxThisTurn: resource.getRemainingThisTurn()
        });
      }
    }

    return usage;
  }

  // === TURN PROCESSING ===

  processTurnStart(turn?: number): void {
    if (turn !== undefined) {
      this.currentTurn = turn;
    } else {
      this.currentTurn++;
    }

    for (const resource of this.resources.values()) {
      resource.processTurnStart(this.currentTurn);
    }
  }

  processRoundStart(round?: number): void {
    if (round !== undefined) {
      this.currentRound = round;
    } else {
      this.currentRound++;
    }

    for (const resource of this.resources.values()) {
      resource.processRoundStart(this.currentRound);
    }
  }

  // === CONVENIENCE METHODS ===

  // Predefined resource types for common use cases
  static createActionPoints(maxActions: number = 2): TurnBasedResourceConfig {
    return {
      id: 'action_points',
      name: 'Action Points',
      baseValue: maxActions,
      maxValue: maxActions,
      refreshPattern: ResourceRefreshPattern.PER_TURN,
      refreshAmount: maxActions,
      maxPerTurn: maxActions
    };
  }

  static createMovementPoints(maxMovement: number = 3): TurnBasedResourceConfig {
    return {
      id: 'movement_points',
      name: 'Movement Points',
      baseValue: maxMovement,
      maxValue: maxMovement,
      refreshPattern: ResourceRefreshPattern.PER_TURN,
      refreshAmount: maxMovement,
      maxPerTurn: maxMovement
    };
  }

  static createMana(maxMana: number = 10, regenPerTurn: number = 2): TurnBasedResourceConfig {
    return {
      id: 'mana',
      name: 'Mana',
      baseValue: maxMana,
      maxValue: maxMana,
      refreshPattern: ResourceRefreshPattern.PER_TURN,
      refreshAmount: regenPerTurn
    };
  }

  static createStamina(maxStamina: number = 5): TurnBasedResourceConfig {
    return {
      id: 'stamina',
      name: 'Stamina',
      baseValue: maxStamina,
      maxValue: maxStamina,
      refreshPattern: ResourceRefreshPattern.PER_ROUND,
      refreshAmount: maxStamina,
      carryOverRatio: 0.5 // Carry over 50% of unused stamina
    };
  }

  // === UTILITY METHODS ===

  getCurrentTurn(): number {
    return this.currentTurn;
  }

  getCurrentRound(): number {
    return this.currentRound;
  }

  getResourceSummary(): Record<string, any> {
    const summary: Record<string, any> = {};
    
    for (const [id, resource] of this.resources.entries()) {
      summary[id] = {
        current: resource.currentValue,
        max: resource.maxValue,
        usedThisTurn: resource.getUsedThisTurn(),
        remainingThisTurn: resource.getRemainingThisTurn(),
        percentage: resource.getPercentage(),
        onCooldown: resource.isOnCooldown(),
        cooldownRemaining: resource.getCooldownRemaining()
      };
    }
    
    return summary;
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    const resourceData: Record<string, any> = {};
    
    for (const [id, resource] of this.resources.entries()) {
      resourceData[id] = resource.serialize();
    }

    return {
      currentTurn: this.currentTurn,
      currentRound: this.currentRound,
      resources: resourceData
    };
  }

  deserialize(data: Record<string, any>): void {
    this.currentTurn = data.currentTurn ?? 0;
    this.currentRound = data.currentRound ?? 0;

    if (data.resources) {
      for (const [id, resourceData] of Object.entries(data.resources)) {
        const resource = this.resources.get(id);
        if (resource) {
          resource.deserialize(resourceData as Record<string, any>);
        }
      }
    }
  }

  // === CLEANUP ===

  clear(): void {
    this.resources.clear();
    this.currentTurn = 0;
    this.currentRound = 0;
  }
}