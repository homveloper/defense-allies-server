// GAS v2 Turn-Based System
// Complete turn-based gaming framework for GAS

// Core Systems
export { 
  TurnBasedResource, 
  TurnBasedResourceManager, 
  ResourceRefreshPattern,
  type TurnBasedResourceConfig,
  type ResourceCost,
  type ResourceUsage
} from './TurnBasedResource';

export { 
  TurnBasedCooldown, 
  TurnBasedTimingManager,
  type TurnBasedTimingConfig,
  type CooldownInfo
} from './TurnBasedTiming';

export { 
  InitiativeCalculator, 
  TurnOrderManager,
  InitiativeRollType,
  type InitiativeConfig,
  type TurnOrderEntry,
  type DelayedAction
} from './InitiativeSystem';

export { 
  PhaseManager, 
  PhaseDefinition,
  type PhaseConfig,
  type PhaseTransition,
  type PhaseExecutionContext
} from './PhaseSystem';

// Context and Utilities
export { 
  TurnContextBuilder, 
  TurnContextValidator, 
  TurnContextUtils,
  TurnPhase,
  type TurnBasedAbilityContext,
  type TurnAction
} from './TurnBasedContext';

// Version info
export const TURN_BASED_VERSION = '2.0.0';
export const TURN_BASED_VERSION_NAME = 'Enhanced Turn-Based';

// Utility function to create a complete turn-based setup
export function createTurnBasedSetup(config?: {
  useRandomInitiative?: boolean;
  initiativeRollType?: InitiativeRollType;
  enablePhaseSystem?: boolean;
  defaultResourceConfigs?: TurnBasedResourceConfig[];
}) {
  const setup = {
    resourceManager: new TurnBasedResourceManager(),
    timingManager: new TurnBasedTimingManager(),
    initiativeCalculator: new InitiativeCalculator(
      config?.initiativeRollType || InitiativeRollType.ONCE_PER_COMBAT,
      config?.useRandomInitiative ?? true
    ),
    turnOrderManager: new TurnOrderManager(),
    phaseManager: config?.enablePhaseSystem !== false ? new PhaseManager() : null
  };

  // Add default resources if provided
  if (config?.defaultResourceConfigs) {
    for (const resourceConfig of config.defaultResourceConfigs) {
      setup.resourceManager.addResource(resourceConfig);
    }
  } else {
    // Add common default resources
    setup.resourceManager.addResource(TurnBasedResourceManager.createActionPoints(2));
    setup.resourceManager.addResource(TurnBasedResourceManager.createMovementPoints(3));
    setup.resourceManager.addResource(TurnBasedResourceManager.createMana(10, 2));
  }

  return setup;
}

// Predefined configurations for common game types
export const TURN_BASED_PRESETS = {
  // Classic turn-based RPG
  CLASSIC_RPG: {
    useRandomInitiative: true,
    initiativeRollType: InitiativeRollType.ONCE_PER_COMBAT,
    enablePhaseSystem: true,
    defaultResourceConfigs: [
      TurnBasedResourceManager.createActionPoints(1),
      TurnBasedResourceManager.createMovementPoints(6),
      TurnBasedResourceManager.createMana(20, 3)
    ]
  },

  // Tactical combat system
  TACTICAL_COMBAT: {
    useRandomInitiative: false,
    initiativeRollType: InitiativeRollType.ONCE_PER_ROUND,
    enablePhaseSystem: true,
    defaultResourceConfigs: [
      TurnBasedResourceManager.createActionPoints(2),
      TurnBasedResourceManager.createMovementPoints(4),
      {
        id: 'reaction_points',
        name: 'Reaction Points',
        baseValue: 1,
        maxValue: 1,
        refreshPattern: ResourceRefreshPattern.PER_TURN,
        refreshAmount: 1
      }
    ]
  },

  // Card/board game style
  CARD_GAME: {
    useRandomInitiative: false,
    initiativeRollType: InitiativeRollType.ONCE_PER_COMBAT,
    enablePhaseSystem: false,
    defaultResourceConfigs: [
      {
        id: 'energy',
        name: 'Energy',
        baseValue: 3,
        maxValue: 10,
        refreshPattern: ResourceRefreshPattern.PER_TURN,
        refreshAmount: 1,
        carryOverRatio: 1.0 // Keep all unused energy
      }
    ]
  },

  // Real-time with turns (like MMO combat)
  REALTIME_TURNS: {
    useRandomInitiative: true,
    initiativeRollType: InitiativeRollType.DYNAMIC,
    enablePhaseSystem: false,
    defaultResourceConfigs: [
      {
        id: 'action_time',
        name: 'Action Time',
        baseValue: 100,
        maxValue: 100,
        refreshPattern: ResourceRefreshPattern.COOLDOWN_TURNS,
        refreshAmount: 100,
        cooldownTurns: 3
      }
    ]
  }
};