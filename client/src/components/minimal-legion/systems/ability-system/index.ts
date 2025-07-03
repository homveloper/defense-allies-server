// Legacy compatibility layer
// This file provides backward compatibility for existing imports
// New code should import directly from @/packages/gas

// Re-export core GAS components
export {
  AbilitySystemComponent,
  GameplayAttribute,
  GameplayTagSystem,
  GameplayEffect,
  GameplayAbility,
  GASUtils as AbilitySystemUtils
} from '@/packages/gas';

export type { IGameplayAbility } from '@/packages/gas';

// Re-export types
export type {
  AbilityContext,
  AbilityCost,
  CooldownInfo,
  AttributeModifier,
  GameplayAttributeData,
  GameplayEffectSpec,
  ActiveGameplayEffect,
  AbilityActivationResult,
  AbilitySystemEvents,
  AbilitySystemEventHandler
} from '@/packages/gas';

// Re-export game-specific abilities
export { BasicAttackAbility } from '../../abilities/BasicAttackAbility';
export { FireballAbility } from '../../abilities/FireballAbility';
export { HealAbility } from '../../abilities/HealAbility';

// Re-export Arena abilities (for backward compatibility)
export { LightningBoltAbility } from '../../../ability-arena/abilities/LightningBoltAbility';
export { IceSpikesAbility } from '../../../ability-arena/abilities/IceSpikesAbility';
export { TeleportAbility } from '../../../ability-arena/abilities/TeleportAbility';
export { ShieldBubbleAbility } from '../../../ability-arena/abilities/ShieldBubbleAbility';