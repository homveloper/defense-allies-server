// Phase-Based Execution System for GAS v2
// Manages turn phases and phase-specific ability restrictions

import { TurnPhase } from './TurnBasedContext';

export interface PhaseConfig {
  id: TurnPhase;
  name: string;
  description: string;
  duration: number;              // Duration in milliseconds for timed phases
  isOptional: boolean;           // Can this phase be skipped?
  autoAdvance: boolean;          // Automatically advance after duration?
  allowedAbilityTypes: string[]; // Types of abilities allowed in this phase
  allowedActions: string[];      // Specific actions allowed
  restrictedActions: string[];   // Actions specifically forbidden
  maxActions?: number;           // Max actions allowed in this phase
  resourceLimits?: Record<string, number>; // Resource usage limits
}

export interface PhaseTransition {
  from: TurnPhase;
  to: TurnPhase;
  condition?: () => boolean;     // Custom condition for transition
  onTransition?: () => void;     // Callback when transition occurs
  autoTrigger?: boolean;         // Automatically trigger when condition met
}

export interface PhaseExecutionContext {
  currentPhase: TurnPhase;
  previousPhase?: TurnPhase;
  nextPhase?: TurnPhase;
  phaseStartTime: number;
  phaseDuration: number;
  actionsInPhase: number;
  maxActionsInPhase: number;
  resourcesUsedInPhase: Record<string, number>;
  canAdvance: boolean;
  isTimedOut: boolean;
}

export class PhaseDefinition {
  public readonly config: PhaseConfig;
  private startTime: number = 0;
  private actionsExecuted: number = 0;
  private resourcesUsed: Record<string, number> = {};

  constructor(config: PhaseConfig) {
    this.config = config;
  }

  // === PHASE LIFECYCLE ===

  start(): void {
    this.startTime = Date.now();
    this.actionsExecuted = 0;
    this.resourcesUsed = {};
  }

  end(): void {
    // Phase cleanup logic
  }

  reset(): void {
    this.startTime = 0;
    this.actionsExecuted = 0;
    this.resourcesUsed = {};
  }

  // === PHASE CHECKS ===

  canExecuteAction(actionType: string, abilityType?: string): boolean {
    // Check action limits
    if (this.config.maxActions && this.actionsExecuted >= this.config.maxActions) {
      return false;
    }

    // Check allowed actions
    if (this.config.allowedActions.length > 0 && !this.config.allowedActions.includes(actionType)) {
      return false;
    }

    // Check restricted actions
    if (this.config.restrictedActions.includes(actionType)) {
      return false;
    }

    // Check ability type restrictions
    if (abilityType && this.config.allowedAbilityTypes.length > 0) {
      if (!this.config.allowedAbilityTypes.includes(abilityType)) {
        return false;
      }
    }

    return true;
  }

  canUseResource(resourceId: string, amount: number): boolean {
    if (!this.config.resourceLimits) {
      return true;
    }

    const limit = this.config.resourceLimits[resourceId];
    if (limit === undefined) {
      return true;
    }

    const used = this.resourcesUsed[resourceId] || 0;
    return (used + amount) <= limit;
  }

  isComplete(): boolean {
    // Check if phase duration expired
    if (this.config.duration > 0) {
      const elapsed = Date.now() - this.startTime;
      if (elapsed >= this.config.duration) {
        return true;
      }
    }

    // Check if max actions reached
    if (this.config.maxActions && this.actionsExecuted >= this.config.maxActions) {
      return true;
    }

    return false;
  }

  canAdvance(): boolean {
    // Always can advance if optional
    if (this.config.isOptional) {
      return true;
    }

    // Can advance if auto-advance and complete
    if (this.config.autoAdvance && this.isComplete()) {
      return true;
    }

    // Manual advancement otherwise
    return true;
  }

  // === ACTION TRACKING ===

  executeAction(actionType: string, resourceCosts: Record<string, number> = {}): boolean {
    if (!this.canExecuteAction(actionType)) {
      return false;
    }

    // Check resource limits
    for (const [resourceId, amount] of Object.entries(resourceCosts)) {
      if (!this.canUseResource(resourceId, amount)) {
        return false;
      }
    }

    // Track action and resource usage
    this.actionsExecuted++;
    
    for (const [resourceId, amount] of Object.entries(resourceCosts)) {
      this.resourcesUsed[resourceId] = (this.resourcesUsed[resourceId] || 0) + amount;
    }

    return true;
  }

  // === GETTERS ===

  getExecutionContext(): PhaseExecutionContext {
    const elapsed = Date.now() - this.startTime;
    
    return {
      currentPhase: this.config.id,
      phaseStartTime: this.startTime,
      phaseDuration: elapsed,
      actionsInPhase: this.actionsExecuted,
      maxActionsInPhase: this.config.maxActions || Infinity,
      resourcesUsedInPhase: { ...this.resourcesUsed },
      canAdvance: this.canAdvance(),
      isTimedOut: this.config.duration > 0 && elapsed >= this.config.duration
    };
  }

  getRemainingTime(): number {
    if (this.config.duration <= 0) {
      return Infinity;
    }

    const elapsed = Date.now() - this.startTime;
    return Math.max(0, this.config.duration - elapsed);
  }

  getRemainingActions(): number {
    if (!this.config.maxActions) {
      return Infinity;
    }

    return Math.max(0, this.config.maxActions - this.actionsExecuted);
  }
}

export class PhaseManager {
  private phases: Map<TurnPhase, PhaseDefinition> = new Map();
  private transitions: PhaseTransition[] = [];
  private currentPhase: TurnPhase = TurnPhase.START;
  private phaseHistory: TurnPhase[] = [];
  private isActive: boolean = false;

  constructor() {
    this.setupDefaultPhases();
    this.setupDefaultTransitions();
  }

  // === SETUP ===

  private setupDefaultPhases(): void {
    const defaultPhases: PhaseConfig[] = [
      {
        id: TurnPhase.START,
        name: 'Turn Start',
        description: 'Beginning of turn effects and setup',
        duration: 0,
        isOptional: false,
        autoAdvance: true,
        allowedAbilityTypes: ['passive', 'trigger'],
        allowedActions: ['trigger_effect', 'process_start'],
        restrictedActions: ['attack', 'move', 'cast']
      },
      {
        id: TurnPhase.MOVEMENT,
        name: 'Movement',
        description: 'Character movement phase',
        duration: 30000, // 30 seconds
        isOptional: true,
        autoAdvance: false,
        allowedAbilityTypes: ['movement', 'utility'],
        allowedActions: ['move', 'teleport', 'dash'],
        restrictedActions: ['attack', 'cast_spell'],
        maxActions: 1,
        resourceLimits: { 'movement_points': 3 }
      },
      {
        id: TurnPhase.MAIN_ACTION,
        name: 'Main Action',
        description: 'Primary action phase',
        duration: 60000, // 1 minute
        isOptional: false,
        autoAdvance: false,
        allowedAbilityTypes: ['attack', 'spell', 'ability', 'item'],
        allowedActions: ['attack', 'cast_spell', 'use_ability', 'use_item'],
        restrictedActions: [],
        maxActions: 1,
        resourceLimits: { 'action_points': 2 }
      },
      {
        id: TurnPhase.BONUS_ACTION,
        name: 'Bonus Action',
        description: 'Quick secondary actions',
        duration: 30000, // 30 seconds
        isOptional: true,
        autoAdvance: false,
        allowedAbilityTypes: ['quick', 'utility', 'consumable'],
        allowedActions: ['quick_attack', 'use_consumable', 'quick_spell'],
        restrictedActions: ['full_attack', 'complex_spell'],
        maxActions: 1,
        resourceLimits: { 'bonus_points': 1 }
      },
      {
        id: TurnPhase.REACTION,
        name: 'Reaction',
        description: 'Reactive abilities and responses',
        duration: 0,
        isOptional: true,
        autoAdvance: true,
        allowedAbilityTypes: ['reaction', 'counter'],
        allowedActions: ['counter_attack', 'defensive_ability'],
        restrictedActions: ['attack', 'move'],
        maxActions: 1
      },
      {
        id: TurnPhase.END,
        name: 'Turn End',
        description: 'End of turn effects and cleanup',
        duration: 0,
        isOptional: false,
        autoAdvance: true,
        allowedAbilityTypes: ['passive', 'trigger'],
        allowedActions: ['trigger_effect', 'process_end'],
        restrictedActions: ['attack', 'move', 'cast']
      }
    ];

    for (const phaseConfig of defaultPhases) {
      this.phases.set(phaseConfig.id, new PhaseDefinition(phaseConfig));
    }
  }

  private setupDefaultTransitions(): void {
    const phaseOrder = [
      TurnPhase.START,
      TurnPhase.MOVEMENT,
      TurnPhase.MAIN_ACTION,
      TurnPhase.BONUS_ACTION,
      TurnPhase.REACTION,
      TurnPhase.END
    ];

    // Create sequential transitions
    for (let i = 0; i < phaseOrder.length - 1; i++) {
      this.transitions.push({
        from: phaseOrder[i],
        to: phaseOrder[i + 1],
        autoTrigger: true
      });
    }

    // Allow skipping optional phases
    this.transitions.push({
      from: TurnPhase.START,
      to: TurnPhase.MAIN_ACTION,
      condition: () => {
        const movementPhase = this.phases.get(TurnPhase.MOVEMENT);
        return movementPhase?.config.isOptional || false;
      }
    });

    this.transitions.push({
      from: TurnPhase.MAIN_ACTION,
      to: TurnPhase.END,
      condition: () => {
        const bonusPhase = this.phases.get(TurnPhase.BONUS_ACTION);
        return bonusPhase?.config.isOptional || false;
      }
    });
  }

  // === PHASE MANAGEMENT ===

  addPhase(config: PhaseConfig): void {
    this.phases.set(config.id, new PhaseDefinition(config));
  }

  removePhase(phaseId: TurnPhase): boolean {
    return this.phases.delete(phaseId);
  }

  addTransition(transition: PhaseTransition): void {
    this.transitions.push(transition);
  }

  // === PHASE EXECUTION ===

  startTurn(initialPhase: TurnPhase = TurnPhase.START): boolean {
    if (this.isActive) {
      return false;
    }

    this.currentPhase = initialPhase;
    this.phaseHistory = [initialPhase];
    this.isActive = true;

    const phase = this.phases.get(initialPhase);
    if (phase) {
      phase.start();
    }

    return true;
  }

  endTurn(): void {
    const currentPhaseObj = this.phases.get(this.currentPhase);
    if (currentPhaseObj) {
      currentPhaseObj.end();
    }

    this.isActive = false;
    this.phaseHistory = [];
  }

  advancePhase(): TurnPhase | null {
    const currentPhaseObj = this.phases.get(this.currentPhase);
    if (!currentPhaseObj || !currentPhaseObj.canAdvance()) {
      return null;
    }

    // Find valid transition
    const validTransition = this.transitions.find(t => 
      t.from === this.currentPhase && 
      (!t.condition || t.condition())
    );

    if (!validTransition) {
      return null;
    }

    // End current phase
    currentPhaseObj.end();

    // Start new phase
    const nextPhase = validTransition.to;
    const nextPhaseObj = this.phases.get(nextPhase);
    
    if (nextPhaseObj) {
      nextPhaseObj.start();
      this.currentPhase = nextPhase;
      this.phaseHistory.push(nextPhase);

      // Execute transition callback
      if (validTransition.onTransition) {
        validTransition.onTransition();
      }

      return nextPhase;
    }

    return null;
  }

  forcePhase(phaseId: TurnPhase): boolean {
    const phase = this.phases.get(phaseId);
    if (!phase) {
      return false;
    }

    // End current phase
    const currentPhaseObj = this.phases.get(this.currentPhase);
    if (currentPhaseObj) {
      currentPhaseObj.end();
    }

    // Start new phase
    this.currentPhase = phaseId;
    this.phaseHistory.push(phaseId);
    phase.start();

    return true;
  }

  // === ABILITY CHECKS ===

  canExecuteAbility(abilityType: string, actionType: string): boolean {
    if (!this.isActive) {
      return false;
    }

    const phase = this.phases.get(this.currentPhase);
    if (!phase) {
      return false;
    }

    return phase.canExecuteAction(actionType, abilityType);
  }

  executeAbilityInPhase(
    abilityType: string, 
    actionType: string, 
    resourceCosts: Record<string, number> = {}
  ): boolean {
    if (!this.canExecuteAbility(abilityType, actionType)) {
      return false;
    }

    const phase = this.phases.get(this.currentPhase);
    if (!phase) {
      return false;
    }

    return phase.executeAction(actionType, resourceCosts);
  }

  // === GETTERS ===

  getCurrentPhase(): TurnPhase {
    return this.currentPhase;
  }

  getCurrentPhaseDefinition(): PhaseDefinition | null {
    return this.phases.get(this.currentPhase) || null;
  }

  getPhaseHistory(): TurnPhase[] {
    return [...this.phaseHistory];
  }

  getNextPhase(): TurnPhase | null {
    const transition = this.transitions.find(t => 
      t.from === this.currentPhase && 
      (!t.condition || t.condition())
    );
    
    return transition ? transition.to : null;
  }

  getAllPhases(): PhaseDefinition[] {
    return Array.from(this.phases.values());
  }

  isPhaseActive(): boolean {
    return this.isActive;
  }

  // === AUTO PROGRESSION ===

  update(): TurnPhase | null {
    if (!this.isActive) {
      return null;
    }

    const phase = this.phases.get(this.currentPhase);
    if (!phase) {
      return null;
    }

    // Check for auto-advance
    if (phase.config.autoAdvance && phase.isComplete()) {
      return this.advancePhase();
    }

    return null;
  }

  // === UTILITY METHODS ===

  getPhaseProgress(): number {
    const phase = this.phases.get(this.currentPhase);
    if (!phase || phase.config.duration <= 0) {
      return 0;
    }

    const remaining = phase.getRemainingTime();
    const total = phase.config.duration;
    
    return Math.max(0, Math.min(1, (total - remaining) / total));
  }

  getPhaseSummary(): Record<string, any> {
    const phase = this.phases.get(this.currentPhase);
    
    return {
      isActive: this.isActive,
      currentPhase: this.currentPhase,
      phaseHistory: this.phaseHistory,
      nextPhase: this.getNextPhase(),
      phaseProgress: this.getPhaseProgress(),
      executionContext: phase?.getExecutionContext(),
      remainingTime: phase?.getRemainingTime(),
      remainingActions: phase?.getRemainingActions()
    };
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    return {
      currentPhase: this.currentPhase,
      phaseHistory: this.phaseHistory,
      isActive: this.isActive,
      phases: Object.fromEntries(
        Array.from(this.phases.entries()).map(([id, phase]) => [
          id,
          phase.getExecutionContext()
        ])
      )
    };
  }

  deserialize(data: Record<string, any>): void {
    this.currentPhase = data.currentPhase || TurnPhase.START;
    this.phaseHistory = data.phaseHistory || [];
    this.isActive = data.isActive || false;
    
    // Note: Phase execution context would need to be restored
    // This is a simplified implementation
  }

  // === CLEANUP ===

  reset(): void {
    this.endTurn();
    this.currentPhase = TurnPhase.START;
    this.phaseHistory = [];
    
    // Reset all phases
    for (const phase of this.phases.values()) {
      phase.reset();
    }
  }

  clear(): void {
    this.reset();
    this.phases.clear();
    this.transitions = [];
  }
}