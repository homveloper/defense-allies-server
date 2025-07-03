// Turn-Based Timing System for GAS v2
// Handles turn-based cooldowns, durations, and timing mechanics

export interface TurnBasedTimingConfig {
  cooldownTurns: number;          // Turns to wait before next use
  usesPerTurn: number;           // Max uses per turn
  usesPerRound: number;          // Max uses per round
  chargeTimeInTurns?: number;    // Turns to charge before activation
  channelTimeInTurns?: number;   // Turns required to maintain
  globalCooldown?: boolean;      // Affects global cooldown
}

export interface CooldownInfo {
  turnsRemaining: number;
  totalTurns: number;
  usesThisTurn: number;
  maxUsesThisTurn: number;
  usesThisRound: number;
  maxUsesThisRound: number;
  isCharging: boolean;
  chargeProgress: number;
  isChanneling: boolean;
  channelProgress: number;
}

export class TurnBasedCooldown {
  public readonly abilityId: string;
  public readonly config: TurnBasedTimingConfig;
  
  private turnsRemaining: number = 0;
  private usesThisTurn: number = 0;
  private usesThisRound: number = 0;
  private lastUsedTurn: number = -1;
  private lastUsedRound: number = -1;
  
  // Charging and channeling state
  private isCharging: boolean = false;
  private chargingStartTurn: number = -1;
  private isChanneling: boolean = false;
  private channelingStartTurn: number = -1;
  private channelInterrupted: boolean = false;

  constructor(abilityId: string, config: TurnBasedTimingConfig) {
    this.abilityId = abilityId;
    this.config = { ...config };
  }

  // === AVAILABILITY CHECKS ===

  isAvailable(currentTurn: number, currentRound: number): boolean {
    return this.canUse(currentTurn, currentRound) && 
           !this.isOnCooldown() &&
           !this.isCharging;
  }

  canUse(currentTurn: number, currentRound: number): boolean {
    // Check turn-based usage limits
    if (this.usesThisTurn >= this.config.usesPerTurn) {
      return false;
    }

    if (this.usesThisRound >= this.config.usesPerRound) {
      return false;
    }

    // Check if still channeling
    if (this.isChanneling && !this.channelInterrupted) {
      return false;
    }

    return true;
  }

  isOnCooldown(): boolean {
    return this.turnsRemaining > 0;
  }

  isBeingCharged(): boolean {
    return this.isCharging;
  }

  isBeingChanneled(): boolean {
    return this.isChanneling && !this.channelInterrupted;
  }

  // === USAGE TRACKING ===

  use(currentTurn: number, currentRound: number): boolean {
    if (!this.canUse(currentTurn, currentRound)) {
      return false;
    }

    // Start cooldown
    this.turnsRemaining = this.config.cooldownTurns;
    this.lastUsedTurn = currentTurn;
    this.lastUsedRound = currentRound;

    // Track usage
    if (this.lastUsedTurn !== currentTurn) {
      this.usesThisTurn = 0;
    }
    if (this.lastUsedRound !== currentRound) {
      this.usesThisRound = 0;
    }

    this.usesThisTurn++;
    this.usesThisRound++;

    // Start channeling if required
    if (this.config.channelTimeInTurns && this.config.channelTimeInTurns > 0) {
      this.startChanneling(currentTurn);
    }

    return true;
  }

  // === CHARGING SYSTEM ===

  startCharging(currentTurn: number): boolean {
    if (this.isCharging || this.isChanneling) {
      return false;
    }

    this.isCharging = true;
    this.chargingStartTurn = currentTurn;
    return true;
  }

  cancelCharging(): void {
    this.isCharging = false;
    this.chargingStartTurn = -1;
  }

  isChargeComplete(currentTurn: number): boolean {
    if (!this.isCharging || !this.config.chargeTimeInTurns) {
      return false;
    }

    const chargeTurns = currentTurn - this.chargingStartTurn;
    return chargeTurns >= this.config.chargeTimeInTurns;
  }

  completeCharging(): boolean {
    if (!this.isCharging) {
      return false;
    }

    this.isCharging = false;
    this.chargingStartTurn = -1;
    return true;
  }

  // === CHANNELING SYSTEM ===

  startChanneling(currentTurn: number): boolean {
    if (this.isChanneling) {
      return false;
    }

    this.isChanneling = true;
    this.channelingStartTurn = currentTurn;
    this.channelInterrupted = false;
    return true;
  }

  interruptChanneling(): void {
    this.channelInterrupted = true;
  }

  isChannelComplete(currentTurn: number): boolean {
    if (!this.isChanneling || !this.config.channelTimeInTurns) {
      return false;
    }

    const channelTurns = currentTurn - this.channelingStartTurn;
    return channelTurns >= this.config.channelTimeInTurns;
  }

  completeChanneling(): boolean {
    if (!this.isChanneling) {
      return false;
    }

    this.isChanneling = false;
    this.channelingStartTurn = -1;
    this.channelInterrupted = false;
    return true;
  }

  // === TURN PROCESSING ===

  processTurnStart(currentTurn: number, currentRound: number): void {
    // Reduce cooldown
    if (this.turnsRemaining > 0) {
      this.turnsRemaining--;
    }

    // Reset turn usage if new turn
    if (this.lastUsedTurn !== currentTurn) {
      this.usesThisTurn = 0;
    }

    // Reset round usage if new round
    if (this.lastUsedRound !== currentRound) {
      this.usesThisRound = 0;
    }

    // Process charging
    if (this.isCharging && this.isChargeComplete(currentTurn)) {
      // Charge complete, ready to activate
    }

    // Process channeling
    if (this.isChanneling && this.isChannelComplete(currentTurn)) {
      this.completeChanneling();
    }
  }

  // === GETTERS ===

  getCooldownInfo(currentTurn: number): CooldownInfo {
    return {
      turnsRemaining: this.turnsRemaining,
      totalTurns: this.config.cooldownTurns,
      usesThisTurn: this.usesThisTurn,
      maxUsesThisTurn: this.config.usesPerTurn,
      usesThisRound: this.usesThisRound,
      maxUsesThisRound: this.config.usesPerRound,
      isCharging: this.isCharging,
      chargeProgress: this.getChargeProgress(currentTurn),
      isChanneling: this.isChanneling,
      channelProgress: this.getChannelProgress(currentTurn)
    };
  }

  private getChargeProgress(currentTurn: number): number {
    if (!this.isCharging || !this.config.chargeTimeInTurns) {
      return 0;
    }

    const elapsed = currentTurn - this.chargingStartTurn;
    return Math.min(1, elapsed / this.config.chargeTimeInTurns);
  }

  private getChannelProgress(currentTurn: number): number {
    if (!this.isChanneling || !this.config.channelTimeInTurns) {
      return 0;
    }

    const elapsed = currentTurn - this.channelingStartTurn;
    return Math.min(1, elapsed / this.config.channelTimeInTurns);
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    return {
      abilityId: this.abilityId,
      turnsRemaining: this.turnsRemaining,
      usesThisTurn: this.usesThisTurn,
      usesThisRound: this.usesThisRound,
      lastUsedTurn: this.lastUsedTurn,
      lastUsedRound: this.lastUsedRound,
      isCharging: this.isCharging,
      chargingStartTurn: this.chargingStartTurn,
      isChanneling: this.isChanneling,
      channelingStartTurn: this.channelingStartTurn,
      channelInterrupted: this.channelInterrupted
    };
  }

  deserialize(data: Record<string, any>): void {
    this.turnsRemaining = data.turnsRemaining ?? 0;
    this.usesThisTurn = data.usesThisTurn ?? 0;
    this.usesThisRound = data.usesThisRound ?? 0;
    this.lastUsedTurn = data.lastUsedTurn ?? -1;
    this.lastUsedRound = data.lastUsedRound ?? -1;
    this.isCharging = data.isCharging ?? false;
    this.chargingStartTurn = data.chargingStartTurn ?? -1;
    this.isChanneling = data.isChanneling ?? false;
    this.channelingStartTurn = data.channelingStartTurn ?? -1;
    this.channelInterrupted = data.channelInterrupted ?? false;
  }
}

export class TurnBasedTimingManager {
  private cooldowns: Map<string, TurnBasedCooldown> = new Map();
  private currentTurn: number = 0;
  private currentRound: number = 0;
  private globalCooldownActive: boolean = false;
  private globalCooldownTurns: number = 0;

  // === COOLDOWN MANAGEMENT ===

  registerAbility(abilityId: string, config: TurnBasedTimingConfig): void {
    const cooldown = new TurnBasedCooldown(abilityId, config);
    this.cooldowns.set(abilityId, cooldown);
  }

  unregisterAbility(abilityId: string): boolean {
    return this.cooldowns.delete(abilityId);
  }

  isAbilityAvailable(abilityId: string): boolean {
    const cooldown = this.cooldowns.get(abilityId);
    if (!cooldown) return false;

    return cooldown.isAvailable(this.currentTurn, this.currentRound) &&
           !this.isGlobalCooldownActive();
  }

  useAbility(abilityId: string): boolean {
    const cooldown = this.cooldowns.get(abilityId);
    if (!cooldown) return false;

    if (!this.isAbilityAvailable(abilityId)) {
      return false;
    }

    const success = cooldown.use(this.currentTurn, this.currentRound);
    
    if (success && cooldown.config.globalCooldown) {
      this.activateGlobalCooldown(1); // 1 turn global cooldown
    }

    return success;
  }

  // === CHARGING SYSTEM ===

  startCharging(abilityId: string): boolean {
    const cooldown = this.cooldowns.get(abilityId);
    if (!cooldown) return false;

    return cooldown.startCharging(this.currentTurn);
  }

  cancelCharging(abilityId: string): void {
    const cooldown = this.cooldowns.get(abilityId);
    if (cooldown) {
      cooldown.cancelCharging();
    }
  }

  isChargeComplete(abilityId: string): boolean {
    const cooldown = this.cooldowns.get(abilityId);
    if (!cooldown) return false;

    return cooldown.isChargeComplete(this.currentTurn);
  }

  completeCharging(abilityId: string): boolean {
    const cooldown = this.cooldowns.get(abilityId);
    if (!cooldown) return false;

    return cooldown.completeCharging();
  }

  // === CHANNELING SYSTEM ===

  interruptChanneling(abilityId: string): void {
    const cooldown = this.cooldowns.get(abilityId);
    if (cooldown) {
      cooldown.interruptChanneling();
    }
  }

  isChannelComplete(abilityId: string): boolean {
    const cooldown = this.cooldowns.get(abilityId);
    if (!cooldown) return false;

    return cooldown.isChannelComplete(this.currentTurn);
  }

  // === GLOBAL COOLDOWN ===

  activateGlobalCooldown(turns: number): void {
    this.globalCooldownActive = true;
    this.globalCooldownTurns = turns;
  }

  isGlobalCooldownActive(): boolean {
    return this.globalCooldownActive && this.globalCooldownTurns > 0;
  }

  getGlobalCooldownRemaining(): number {
    return this.globalCooldownActive ? this.globalCooldownTurns : 0;
  }

  // === TURN PROCESSING ===

  processTurnStart(turn?: number, round?: number): void {
    if (turn !== undefined) {
      this.currentTurn = turn;
    } else {
      this.currentTurn++;
    }

    if (round !== undefined) {
      this.currentRound = round;
    }

    // Process global cooldown
    if (this.globalCooldownActive && this.globalCooldownTurns > 0) {
      this.globalCooldownTurns--;
      if (this.globalCooldownTurns <= 0) {
        this.globalCooldownActive = false;
      }
    }

    // Process all ability cooldowns
    for (const cooldown of this.cooldowns.values()) {
      cooldown.processTurnStart(this.currentTurn, this.currentRound);
    }
  }

  processRoundStart(round?: number): void {
    if (round !== undefined) {
      this.currentRound = round;
    } else {
      this.currentRound++;
    }
  }

  // === GETTERS ===

  getCooldownInfo(abilityId: string): CooldownInfo | null {
    const cooldown = this.cooldowns.get(abilityId);
    return cooldown ? cooldown.getCooldownInfo(this.currentTurn) : null;
  }

  getAllCooldowns(): Map<string, CooldownInfo> {
    const result = new Map<string, CooldownInfo>();
    
    for (const [abilityId, cooldown] of this.cooldowns.entries()) {
      result.set(abilityId, cooldown.getCooldownInfo(this.currentTurn));
    }
    
    return result;
  }

  getCurrentTurn(): number {
    return this.currentTurn;
  }

  getCurrentRound(): number {
    return this.currentRound;
  }

  // === UTILITY METHODS ===

  getTimingSummary(): Record<string, any> {
    const summary: Record<string, any> = {
      currentTurn: this.currentTurn,
      currentRound: this.currentRound,
      globalCooldown: {
        active: this.globalCooldownActive,
        remaining: this.globalCooldownTurns
      },
      abilities: {}
    };

    for (const [abilityId, cooldown] of this.cooldowns.entries()) {
      summary.abilities[abilityId] = cooldown.getCooldownInfo(this.currentTurn);
    }

    return summary;
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    const cooldownData: Record<string, any> = {};
    
    for (const [id, cooldown] of this.cooldowns.entries()) {
      cooldownData[id] = cooldown.serialize();
    }

    return {
      currentTurn: this.currentTurn,
      currentRound: this.currentRound,
      globalCooldownActive: this.globalCooldownActive,
      globalCooldownTurns: this.globalCooldownTurns,
      cooldowns: cooldownData
    };
  }

  deserialize(data: Record<string, any>): void {
    this.currentTurn = data.currentTurn ?? 0;
    this.currentRound = data.currentRound ?? 0;
    this.globalCooldownActive = data.globalCooldownActive ?? false;
    this.globalCooldownTurns = data.globalCooldownTurns ?? 0;

    if (data.cooldowns) {
      for (const [id, cooldownData] of Object.entries(data.cooldowns)) {
        const cooldown = this.cooldowns.get(id);
        if (cooldown) {
          cooldown.deserialize(cooldownData as Record<string, any>);
        }
      }
    }
  }

  // === CLEANUP ===

  clear(): void {
    this.cooldowns.clear();
    this.currentTurn = 0;
    this.currentRound = 0;
    this.globalCooldownActive = false;
    this.globalCooldownTurns = 0;
  }
}