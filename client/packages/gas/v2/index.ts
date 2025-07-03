// GAS v2 - Enhanced Version with Advanced Features
// Includes event system, condition system, and enhanced capabilities

// Core Components (Enhanced)
export { EnhancedAbilitySystemComponent as AbilitySystemComponent } from './AbilitySystemComponent';
export type { IGameplayAbility } from './AbilitySystemComponent';
export { GameplayAttribute } from './GameplayAttribute';
export { GameplayTagSystem } from './GameplayTagSystem';
export { GameplayEffect } from './GameplayEffect';
export { GameplayAbility } from './GameplayAbility';

// V2 Exclusive Features
export { EnhancedEventSystem } from './EnhancedEventSystem';
export { 
  ConditionManager,
  BaseCondition,
  AttributeCondition,
  TagCondition,
  CooldownCondition,
  TimeCondition,
  ComboCondition,
  ResourceCondition
} from './ConditionSystem';
export { 
  AbilityQueue, 
  QueueOptions, 
  QueueExecutionMode, 
  QueuedAbility, 
  QueueStats 
} from './AbilityQueue';

// Serialization System
export * from './serialization';

// Turn-Based System
export * from './turn-based';

// Version info
export const VERSION = '2.0.0';
export const VERSION_NAME = 'Enhanced';