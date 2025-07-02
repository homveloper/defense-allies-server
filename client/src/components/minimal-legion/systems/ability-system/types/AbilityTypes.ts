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

// Event types for ability system
export interface AbilitySystemEvents {
  'ability-activated': { abilityId: string; context: AbilityContext };
  'ability-failed': { abilityId: string; reason: string };
  'effect-applied': { effectId: string; target: any };
  'effect-removed': { effectId: string; target: any };
  'attribute-changed': { attribute: string; oldValue: number; newValue: number };
  'tag-added': { tag: string };
  'tag-removed': { tag: string };
}

export type AbilitySystemEventHandler<T extends keyof AbilitySystemEvents> = (
  data: AbilitySystemEvents[T]
) => void;