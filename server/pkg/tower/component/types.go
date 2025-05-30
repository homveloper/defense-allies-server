// Package component types defines all the enumerated types and constants
// used throughout the component system.
package component

// ComponentType represents the type of a component
type ComponentType string

// Core component types for the LEGO-style tower system
const (
	// Targeting Components
	ComponentTypeSingleTarget ComponentType = "single_target"
	ComponentTypeMultiTarget  ComponentType = "multi_target"
	ComponentTypeAreaTarget   ComponentType = "area_target"
	ComponentTypeChainTarget  ComponentType = "chain_target"
	ComponentTypeClosest      ComponentType = "closest_target"
	ComponentTypeWeakest      ComponentType = "weakest_target"
	ComponentTypeStrongest    ComponentType = "strongest_target"

	// Damage Components
	ComponentTypeBasicDamage    ComponentType = "basic_damage"
	ComponentTypeFireDamage     ComponentType = "fire_damage"
	ComponentTypeIceDamage      ComponentType = "ice_damage"
	ComponentTypeElectricDamage ComponentType = "electric_damage"
	ComponentTypePoisonDamage   ComponentType = "poison_damage"
	ComponentTypePhysicalDamage ComponentType = "physical_damage"
	ComponentTypeMagicalDamage  ComponentType = "magical_damage"
	ComponentTypePercentDamage  ComponentType = "percent_damage"
	ComponentTypeTrueDamage     ComponentType = "true_damage"

	// Effect Components
	ComponentTypeBurnEffect   ComponentType = "burn_effect"
	ComponentTypeFreezeEffect ComponentType = "freeze_effect"
	ComponentTypeSlowEffect   ComponentType = "slow_effect"
	ComponentTypeStunEffect   ComponentType = "stun_effect"
	ComponentTypePoisonEffect ComponentType = "poison_effect"
	ComponentTypeBuffEffect   ComponentType = "buff_effect"
	ComponentTypeDebuffEffect ComponentType = "debuff_effect"
	ComponentTypeHealEffect   ComponentType = "heal_effect"
	ComponentTypeShieldEffect ComponentType = "shield_effect"

	// Range Components
	ComponentTypeRangeCheck   ComponentType = "range_check"
	ComponentTypeAreaOfEffect ComponentType = "area_of_effect"
	ComponentTypeConeArea     ComponentType = "cone_area"
	ComponentTypeLineArea     ComponentType = "line_area"
	ComponentTypeCircleArea   ComponentType = "circle_area"

	// Projectile Components
	ComponentTypeProjectileLaunch    ComponentType = "projectile_launch"
	ComponentTypeHomingProjectile    ComponentType = "homing_projectile"
	ComponentTypePiercingProjectile  ComponentType = "piercing_projectile"
	ComponentTypeBouncingProjectile  ComponentType = "bouncing_projectile"
	ComponentTypeExplosiveProjectile ComponentType = "explosive_projectile"

	// Conditional Components
	ComponentTypeConditionCheck ComponentType = "condition_check"
	ComponentTypeEnemyTypeCheck ComponentType = "enemy_type_check"
	ComponentTypeHealthCheck    ComponentType = "health_check"
	ComponentTypeDistanceCheck  ComponentType = "distance_check"
	ComponentTypeTimeCheck      ComponentType = "time_check"
	ComponentTypeRandomChance   ComponentType = "random_chance"

	// Utility Components
	ComponentTypeCooldown       ComponentType = "cooldown"
	ComponentTypeResourceCost   ComponentType = "resource_cost"
	ComponentTypeAnimation      ComponentType = "animation"
	ComponentTypeSound          ComponentType = "sound"
	ComponentTypeParticleEffect ComponentType = "particle_effect"
	ComponentTypeStatModifier   ComponentType = "stat_modifier"

	// Advanced Components
	ComponentTypeSynergy          ComponentType = "synergy"
	ComponentTypeChainReaction    ComponentType = "chain_reaction"
	ComponentTypeComboTrigger     ComponentType = "combo_trigger"
	ComponentTypeMatrixModifier   ComponentType = "matrix_modifier"
	ComponentTypeEnvironmentCheck ComponentType = "environment_check"
)

// ComponentCategory groups related component types
type ComponentCategory string

const (
	CategoryTargeting   ComponentCategory = "targeting"
	CategoryDamage      ComponentCategory = "damage"
	CategoryEffect      ComponentCategory = "effect"
	CategoryRange       ComponentCategory = "range"
	CategoryProjectile  ComponentCategory = "projectile"
	CategoryConditional ComponentCategory = "conditional"
	CategoryUtility     ComponentCategory = "utility"
	CategoryAdvanced    ComponentCategory = "advanced"
	CategoryEnvironment ComponentCategory = "environment"
	CategorySynergy     ComponentCategory = "synergy"
)

// DataType represents the type of data that flows between components
type DataType string

const (
	// Basic data types
	DataTypeString  DataType = "string"
	DataTypeInt     DataType = "int"
	DataTypeFloat   DataType = "float"
	DataTypeBool    DataType = "bool"
	DataTypeVector2 DataType = "vector2"
	DataTypeMatrix  DataType = "matrix"

	// Game entity types
	DataTypeTarget  DataType = "target"
	DataTypeTargets DataType = "targets"
	DataTypeEnemy   DataType = "enemy"
	DataTypeEnemies DataType = "enemies"
	DataTypeTower   DataType = "tower"
	DataTypeTowers  DataType = "towers"
	DataTypePlayer  DataType = "player"
	DataTypePlayers DataType = "players"

	// Game data types
	DataTypeDamage   DataType = "damage"
	DataTypeEffect   DataType = "effect"
	DataTypeEffects  DataType = "effects"
	DataTypeEvent    DataType = "event"
	DataTypeEvents   DataType = "events"
	DataTypePosition DataType = "position"
	DataTypeArea     DataType = "area"
	DataTypeRange    DataType = "range"

	// Complex types
	DataTypeAny    DataType = "any"
	DataTypeArray  DataType = "array"
	DataTypeObject DataType = "object"
	DataTypeMap    DataType = "map"
)

// EffectType represents different types of effects that can be applied
type EffectType string

const (
	// Damage effects
	EffectTypeDamage         EffectType = "damage"
	EffectTypeDamageOverTime EffectType = "damage_over_time"
	EffectTypeInstantDamage  EffectType = "instant_damage"

	// Status effects
	EffectTypeBurn      EffectType = "burn"
	EffectTypeFreeze    EffectType = "freeze"
	EffectTypeSlow      EffectType = "slow"
	EffectTypeStun      EffectType = "stun"
	EffectTypePoison    EffectType = "poison"
	EffectTypeBleed     EffectType = "bleed"
	EffectTypeConfusion EffectType = "confusion"
	EffectTypeFear      EffectType = "fear"
	EffectTypeCharm     EffectType = "charm"
	EffectTypeSilence   EffectType = "silence"
	EffectTypeRoot      EffectType = "root"
	EffectTypeKnockback EffectType = "knockback"

	// Buff effects
	EffectTypeAttackBuff   EffectType = "attack_buff"
	EffectTypeDefenseBuff  EffectType = "defense_buff"
	EffectTypeSpeedBuff    EffectType = "speed_buff"
	EffectTypeAccuracyBuff EffectType = "accuracy_buff"
	EffectTypeRangeBuff    EffectType = "range_buff"
	EffectTypeCooldownBuff EffectType = "cooldown_buff"

	// Debuff effects
	EffectTypeAttackDebuff   EffectType = "attack_debuff"
	EffectTypeDefenseDebuff  EffectType = "defense_debuff"
	EffectTypeSpeedDebuff    EffectType = "speed_debuff"
	EffectTypeAccuracyDebuff EffectType = "accuracy_debuff"
	EffectTypeRangeDebuff    EffectType = "range_debuff"
	EffectTypeCooldownDebuff EffectType = "cooldown_debuff"

	// Healing effects
	EffectTypeHeal         EffectType = "heal"
	EffectTypeHealOverTime EffectType = "heal_over_time"
	EffectTypeShield       EffectType = "shield"
	EffectTypeBarrier      EffectType = "barrier"
	EffectTypeRegeneration EffectType = "regeneration"

	// Special effects
	EffectTypeInvisibility    EffectType = "invisibility"
	EffectTypeInvulnerability EffectType = "invulnerability"
	EffectTypeReflection      EffectType = "reflection"
	EffectTypeAbsorption      EffectType = "absorption"
	EffectTypeConversion      EffectType = "conversion"
	EffectTypeTeleport        EffectType = "teleport"
	EffectTypeTransform       EffectType = "transform"
	EffectTypeClone           EffectType = "clone"
	EffectTypeSummon          EffectType = "summon"
	EffectTypeBanish          EffectType = "banish"
)

// EventType represents different types of game events
type EventType string

const (
	// Combat events
	EventTypeAttack      EventType = "attack"
	EventTypeHit         EventType = "hit"
	EventTypeMiss        EventType = "miss"
	EventTypeCritical    EventType = "critical"
	EventTypeDamageDealt EventType = "damage_dealt"
	EventTypeDamageTaken EventType = "damage_taken"
	EventTypeKill        EventType = "kill"
	EventTypeDeath       EventType = "death"

	// Tower events
	EventTypeTowerPlaced     EventType = "tower_placed"
	EventTypeTowerUpgraded   EventType = "tower_upgraded"
	EventTypeTowerSold       EventType = "tower_sold"
	EventTypeTowerDestroyed  EventType = "tower_destroyed"
	EventTypeAbilityUsed     EventType = "ability_used"
	EventTypeAbilityCooldown EventType = "ability_cooldown"

	// Enemy events
	EventTypeEnemySpawned    EventType = "enemy_spawned"
	EventTypeEnemyReachedEnd EventType = "enemy_reached_end"
	EventTypeEnemyKilled     EventType = "enemy_killed"
	EventTypeWaveStarted     EventType = "wave_started"
	EventTypeWaveCompleted   EventType = "wave_completed"
	EventTypeBossSpawned     EventType = "boss_spawned"

	// Effect events
	EventTypeEffectApplied   EventType = "effect_applied"
	EventTypeEffectExpired   EventType = "effect_expired"
	EventTypeEffectStacked   EventType = "effect_stacked"
	EventTypeEffectDispelled EventType = "effect_dispelled"

	// Environment events
	EventTypeEnvironmentChange EventType = "environment_change"
	EventTypeWeatherChange     EventType = "weather_change"
	EventTypeTimeChange        EventType = "time_change"
	EventTypeTerrainChange     EventType = "terrain_change"

	// Synergy events
	EventTypeSynergyActivated EventType = "synergy_activated"
	EventTypeSynergyLost      EventType = "synergy_lost"
	EventTypeComboTriggered   EventType = "combo_triggered"
	EventTypeChainReaction    EventType = "chain_reaction"

	// System events
	EventTypeGameStarted     EventType = "game_started"
	EventTypeGameEnded       EventType = "game_ended"
	EventTypePlayerJoined    EventType = "player_joined"
	EventTypePlayerLeft      EventType = "player_left"
	EventTypeResourceChanged EventType = "resource_changed"
)

// Priority levels for component execution and event handling
type Priority int

const (
	PriorityLowest   Priority = 0
	PriorityLow      Priority = 25
	PriorityNormal   Priority = 50
	PriorityHigh     Priority = 75
	PriorityHighest  Priority = 100
	PriorityCritical Priority = 200
)

// ConnectionType represents how components can be connected
type ConnectionType string

const (
	ConnectionTypeSequential  ConnectionType = "sequential"  // Output of A becomes input of B
	ConnectionTypeParallel    ConnectionType = "parallel"    // A and B execute in parallel
	ConnectionTypeBranch      ConnectionType = "branch"      // A's output goes to multiple components
	ConnectionTypeMerge       ConnectionType = "merge"       // Multiple outputs merge into one input
	ConnectionTypeConditional ConnectionType = "conditional" // Connection depends on condition
)

// ConnectionPriority represents the priority of a connection
type ConnectionPriority string

const (
	ConnectionPriorityLowest  ConnectionPriority = "lowest"
	ConnectionPriorityLow     ConnectionPriority = "low"
	ConnectionPriorityNormal  ConnectionPriority = "normal"
	ConnectionPriorityHigh    ConnectionPriority = "high"
	ConnectionPriorityHighest ConnectionPriority = "highest"
)

// ExecutionMode represents how a component should be executed
type ExecutionMode string

const (
	ExecutionModeImmediate ExecutionMode = "immediate" // Execute immediately when triggered
	ExecutionModeDeferred  ExecutionMode = "deferred"  // Execute at end of frame
	ExecutionModeQueued    ExecutionMode = "queued"    // Add to execution queue
	ExecutionModeAsync     ExecutionMode = "async"     // Execute asynchronously
)

// ComponentState represents the current state of a component
type ComponentState string

const (
	ComponentStateIdle      ComponentState = "idle"
	ComponentStateExecuting ComponentState = "executing"
	ComponentStateWaiting   ComponentState = "waiting"
	ComponentStateError     ComponentState = "error"
	ComponentStateDisabled  ComponentState = "disabled"
	ComponentStateDestroyed ComponentState = "destroyed"
)
