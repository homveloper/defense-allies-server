// Core Ability System Types

export interface AbilityContext {
  owner: any; // Entity that owns the ability
  target?: any; // Target entity (optional)
  scene: Phaser.Scene; // Current game scene
  payload?: any; // Additional data
}

export interface AbilityCost {
  attribute: string; // 'mana', 'health', etc.
  amount: number;
}

export interface CooldownInfo {
  remaining: number; // Time remaining in milliseconds
  total: number; // Total cooldown duration
}

export interface AttributeModifier {
  id: string;
  attribute: string;
  operation: 'add' | 'multiply' | 'override';
  magnitude: number;
  source: string; // Source of the modifier (ability id, effect id, etc.)
}

export interface GameplayAttributeData {
  name: string;
  baseValue: number;
  currentValue: number;
  maxValue?: number;
  modifiers: AttributeModifier[];
}

export interface GameplayEffectSpec {
  id: string;
  name: string;
  duration: number; // -1 for infinite, 0 for instant
  period?: number; // For periodic effects
  stackingPolicy?: 'none' | 'aggregate' | 'refresh';
  maxStacks?: number;
  
  // Modifiers
  attributeModifiers?: AttributeModifier[];
  grantedTags?: string[];
  removedTags?: string[];
  
  // Callbacks
  onApplied?: (target: any) => void;
  onRemoved?: (target: any) => void;
  onPeriodic?: (target: any) => void;
}

export interface ActiveGameplayEffect {
  spec: GameplayEffectSpec;
  startTime: number;
  lastTickTime?: number;
  stacks: number;
  appliedModifiers: AttributeModifier[];
}

export interface AbilityActivationResult {
  success: boolean;
  failureReason?: string;
  cooldownRemaining?: number;
}

// Enhanced Event System
export interface AbilitySystemEvents {
  // Ability Events
  'ability-activated': { abilityId: string; context: AbilityContext; timestamp: number };
  'ability-failed': { abilityId: string; reason: string; context: AbilityContext; timestamp: number };
  'ability-cooldown-started': { abilityId: string; duration: number; timestamp: number };
  'ability-cooldown-ended': { abilityId: string; timestamp: number };
  'ability-interrupted': { abilityId: string; reason: string; context: AbilityContext; timestamp: number };
  'ability-blocked': { abilityId: string; reason: string; blockedBy: string[]; timestamp: number };
  
  // Effect Events
  'effect-applied': { effectId: string; target: any; source?: any; timestamp: number };
  'effect-removed': { effectId: string; target: any; reason: 'expired' | 'dispelled' | 'replaced'; timestamp: number };
  'effect-expired': { effectId: string; target: any; duration: number; timestamp: number };
  'effect-refreshed': { effectId: string; target: any; newDuration: number; stacks: number; timestamp: number };
  'effect-stacked': { effectId: string; target: any; currentStacks: number; maxStacks: number; timestamp: number };
  
  // Attribute Events
  'attribute-changed': { attribute: string; oldValue: number; newValue: number; change: number; source?: string; timestamp: number };
  'attribute-minimum-reached': { attribute: string; value: number; minimum: number; timestamp: number };
  'attribute-maximum-reached': { attribute: string; value: number; maximum: number; timestamp: number };
  'attribute-depleted': { attribute: string; previousValue: number; timestamp: number };
  'attribute-restored': { attribute: string; newValue: number; restoredAmount: number; timestamp: number };
  
  // Tag Events
  'tag-added': { tag: string; source?: string; timestamp: number };
  'tag-removed': { tag: string; reason: 'effect-expired' | 'manual' | 'replaced'; timestamp: number };
  'tag-blocked': { tag: string; blockedBy: string[]; timestamp: number };
  
  // Cost Events
  'cost-paid': { abilityId: string; costs: AbilityCost[]; timestamp: number };
  'cost-insufficient': { abilityId: string; requiredCost: AbilityCost; availableAmount: number; timestamp: number };
  
  // Timing Events
  'cooldown-tick': { abilityId: string; remaining: number; total: number; timestamp: number };
  'effect-tick': { effectId: string; tickNumber: number; totalTicks: number; timestamp: number };
  
  // Queue Events
  'ability-queued': { queueId: string; abilityId: string; priority: number; queueSize: number; timestamp: number };
  'ability-queue-executed': { queueId: string; abilityId: string; waitTime: number; timestamp: number };
  'ability-queue-cancelled': { queueId: string; abilityId: string; reason: string; timestamp: number };
  'ability-queue-interrupted': { queueId: string; abilityId: string; timestamp: number };
  'ability-queue-failed': { queueId: string; abilityId: string; reason: string; waitTime: number; timestamp: number };
  
  // System Events
  'asc-initialized': { owner: any; timestamp: number };
  'asc-destroyed': { owner: any; timestamp: number };
  'debug-info-requested': { requestId: string; timestamp: number };
}

export type AbilitySystemEventHandler<T extends keyof AbilitySystemEvents> = (
  data: AbilitySystemEvents[T]
) => void;

// Enhanced Event Priority System
export enum EventPriority {
  HIGHEST = 0,
  HIGH = 1,
  NORMAL = 2,
  LOW = 3,
  LOWEST = 4
}

export interface EventListener<T extends keyof AbilitySystemEvents> {
  handler: AbilitySystemEventHandler<T>;
  priority: EventPriority;
  once?: boolean; // Remove after first trigger
  filter?: (data: AbilitySystemEvents[T]) => boolean; // Conditional execution
}

// Condition System Types
export interface AbilityCondition {
  id: string;
  name: string;
  description?: string;
  check(context: AbilityContext): boolean | Promise<boolean>;
  onFailure?: (context: AbilityContext, reason: string) => void;
}

export interface AttributeConditionConfig {
  attribute: string;
  operator: '>' | '<' | '>=' | '<=' | '===' | '!==';
  value: number;
  percentage?: boolean; // Check against percentage of max value
}

export interface TagConditionConfig {
  tags: string[];
  mode: 'all' | 'any' | 'none'; // all: has all tags, any: has any tag, none: has no tags
}

export interface CooldownConditionConfig {
  abilityId: string;
  state: 'ready' | 'on-cooldown';
}

export interface TimeConditionConfig {
  timeWindow: {
    start: number; // milliseconds since game start
    end: number;
  };
  gameTime?: boolean; // Use game time vs real time
}

export interface ComboConditionConfig {
  requiredSequence: string[]; // Array of ability IDs in order
  maxInterval: number; // Maximum time between abilities in sequence
  mustBeExact?: boolean; // Sequence must be exact or can have other abilities in between
}

export interface ResourceConditionConfig {
  attribute: string;
  amount: number;
  operation: 'cost' | 'minimum' | 'exact'; // cost: can pay amount, minimum: has at least amount, exact: has exactly amount
}

// Condition Result
export interface ConditionResult {
  passed: boolean;
  reason?: string;
  data?: any; // Additional context about why condition failed/passed
}

// Advanced Context with Conditions
export interface EnhancedAbilityContext extends AbilityContext {
  conditions?: AbilityCondition[];
  skipConditions?: string[]; // Condition IDs to skip
  forceActivation?: boolean; // Bypass all conditions
  metadata?: Record<string, any>; // Additional metadata
}