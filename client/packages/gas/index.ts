// GAS (Gameplay Ability System) Core Package
// A complete implementation of Unreal Engine's GAS for JavaScript/TypeScript

// ===== VERSION-BASED EXPORTS =====
// Import specific version as needed for your project

// V1 - Stable, production-ready (DEFAULT)
export * as v1 from './v1';

// V2 - Enhanced with event system and conditions
export * as v2 from './v2';

// ===== DEFAULT EXPORTS (V1 for backward compatibility) =====
export { 
  AbilitySystemComponent,
  IGameplayAbility,
  GameplayAttribute,
  GameplayTagSystem,
  GameplayEffect,
  GameplayAbility
} from './v1';

// ===== OPTIONAL V2 FEATURES =====
// These are opt-in and don't break existing code
export { EnhancedEventSystem } from './v2/EnhancedEventSystem';
export { 
  ConditionManager,
  BaseCondition,
  AttributeCondition,
  TagCondition,
  CooldownCondition,
  TimeCondition,
  ComboCondition,
  ResourceCondition
} from './v2/ConditionSystem';
export { 
  AbilityQueue, 
  QueueOptions, 
  QueueExecutionMode, 
  QueuedAbility, 
  QueueStats 
} from './v2/AbilityQueue';

// Serialization System (V2 only)
export { 
  gasSerializer,
  GASSerializer,
  JsonCodec,
  JsonCodecFactory,
  getSerializationRegistry
} from './v2/serialization';

// Types
export type {
  AbilityContext,
  EnhancedAbilityContext,
  AbilityCost,
  CooldownInfo,
  AttributeModifier,
  GameplayAttributeData,
  GameplayEffectSpec,
  ActiveGameplayEffect,
  AbilityActivationResult,
  AbilitySystemEvents,
  AbilitySystemEventHandler,
  EventPriority,
  EventListener,
  AbilityCondition,
  AttributeConditionConfig,
  TagConditionConfig,
  CooldownConditionConfig,
  TimeConditionConfig,
  ComboConditionConfig,
  ResourceConditionConfig,
  ConditionResult
} from './types/AbilityTypes';

// Utilities
export { GASUtils } from './utils/GASUtils';

// Version
export const GAS_VERSION = '1.0.0';

// Package Info
export const GAS_INFO = {
  name: 'Gameplay Ability System',
  version: GAS_VERSION,
  description: 'A complete GAS implementation for JavaScript/TypeScript games',
  author: 'Defense Allies Team',
  license: 'MIT'
};