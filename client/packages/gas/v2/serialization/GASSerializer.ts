// GAS Serialization System
// Handles serialization of abilities, effects, attributes, and complete ASC state

import { EnhancedAbilitySystemComponent, IGameplayAbility } from '../AbilitySystemComponent';
import { GameplayAttribute } from '../GameplayAttribute';
import { GameplayEffect } from '../GameplayEffect';
import { GameplayAbility } from '../GameplayAbility';
import { QueuedAbility, QueueStats } from '../AbilityQueue';
import { 
  getSerializationRegistry, 
  SerializationOptions, 
  SerializedData 
} from './SerializationCodec';
import { JsonCodec, JsonCodecFactory } from './JsonCodec';
import { 
  GameplayAttributeData, 
  GameplayEffectSpec, 
  ActiveGameplayEffect,
  AttributeModifier
} from '../../types/AbilityTypes';

// Initialize default JSON codec
const registry = getSerializationRegistry();
registry.register(JsonCodecFactory.createGASCodec());
registry.setDefault('json');

export interface SerializableAbility {
  id: string;
  name: string;
  description: string;
  cooldown: number;
  type: string; // Class name or type identifier
  data?: Record<string, any>; // Additional ability-specific data
}

export interface SerializableASCState {
  version: string;
  timestamp: number;
  owner: {
    id?: string;
    name?: string;
    type?: string;
  };
  
  attributes: Record<string, GameplayAttributeData>;
  abilities: SerializableAbility[];
  activeEffects: ActiveGameplayEffect[];
  tags: string[];
  cooldowns: Record<string, { endTime: number; duration: number }>;
  
  // Queue state
  queuedAbilities: QueuedAbility[];
  queueStats: QueueStats;
  queueMode: string;
  
  // Conditions and global state
  globalConditions: string[];
  
  metadata?: Record<string, any>;
}

export interface SerializationContext {
  includeOwner?: boolean;
  includeQueue?: boolean;
  includeStats?: boolean;
  includeMethods?: boolean; // Whether to serialize ability methods (dangerous)
  customData?: Record<string, any>;
}

export class GASSerializer {
  private static instance: GASSerializer;
  
  static getInstance(): GASSerializer {
    if (!GASSerializer.instance) {
      GASSerializer.instance = new GASSerializer();
    }
    return GASSerializer.instance;
  }
  
  // === INDIVIDUAL COMPONENT SERIALIZATION ===
  
  serializeAttribute(attribute: GameplayAttribute, options?: SerializationOptions): SerializedData {
    const data: GameplayAttributeData = {
      name: attribute.name,
      baseValue: attribute.baseValue,
      currentValue: attribute.currentValue,
      maxValue: attribute.maxValue,
      modifiers: attribute.getModifiers()
    };
    
    return registry.serialize(data, options);
  }
  
  deserializeAttribute(serialized: SerializedData): GameplayAttribute {
    const data = registry.deserialize<GameplayAttributeData>(serialized);
    const attribute = new GameplayAttribute(data.name, data.baseValue, data.maxValue);
    
    attribute.currentValue = data.currentValue;
    
    // Restore modifiers
    data.modifiers.forEach(modifier => {
      attribute.addModifier(modifier);
    });
    
    return attribute;
  }
  
  serializeEffect(effect: GameplayEffect, options?: SerializationOptions): SerializedData {
    return registry.serialize(effect.spec, options);
  }
  
  deserializeEffect(serialized: SerializedData): GameplayEffect {
    const spec = registry.deserialize<GameplayEffectSpec>(serialized);
    return new GameplayEffect(spec);
  }
  
  serializeAbility(
    ability: IGameplayAbility, 
    context?: SerializationContext,
    options?: SerializationOptions
  ): SerializedData {
    const data: SerializableAbility = {
      id: ability.id,
      name: ability.name,
      description: ability.description,
      cooldown: ability.cooldown,
      type: ability.constructor.name
    };
    
    // Add custom data if the ability supports it
    if ('serialize' in ability && typeof ability.serialize === 'function') {
      data.data = (ability as any).serialize();
    }
    
    return registry.serialize(data, options);
  }
  
  // Note: Deserializing abilities requires a registry of ability classes
  // This is intentionally limited for security reasons
  deserializeAbility(
    serialized: SerializedData, 
    abilityRegistry: Map<string, new() => IGameplayAbility>
  ): IGameplayAbility {
    const data = registry.deserialize<SerializableAbility>(serialized);
    
    const AbilityClass = abilityRegistry.get(data.type);
    if (!AbilityClass) {
      throw new Error(`Unknown ability type: ${data.type}`);
    }
    
    const ability = new AbilityClass();
    
    // Restore custom data if the ability supports it
    if (data.data && 'deserialize' in ability && typeof ability.deserialize === 'function') {
      (ability as any).deserialize(data.data);
    }
    
    return ability;
  }
  
  // === COMPLETE ASC SERIALIZATION ===
  
  serializeASC(
    asc: EnhancedAbilitySystemComponent,
    context: SerializationContext = {},
    options?: SerializationOptions
  ): SerializedData {
    const state: SerializableASCState = {
      version: '2.0.0',
      timestamp: Date.now(),
      owner: context.includeOwner ? {
        id: asc.getOwner()?.id,
        name: asc.getOwner()?.name,
        type: asc.getOwner()?.constructor?.name
      } : {},
      
      attributes: this.serializeAttributes(asc),
      abilities: this.serializeAbilities(asc, context),
      activeEffects: this.serializeActiveEffects(asc),
      tags: Array.from(asc.getOwner().abilitySystem?.tagSystem?.getAllTags?.() || []),
      cooldowns: this.serializeCooldowns(asc),
      
      queuedAbilities: context.includeQueue ? asc.getQueuedAbilities?.() || [] : [],
      queueStats: context.includeStats ? asc.getQueueStats?.() || {} as QueueStats : {} as QueueStats,
      queueMode: asc.getQueueMode?.()?.toString() || 'auto',
      
      globalConditions: [], // Would need access to private field
      
      metadata: {
        serializationContext: context,
        ...context.customData
      }
    };
    
    return registry.serialize(state, options);
  }
  
  private serializeAttributes(asc: EnhancedAbilitySystemComponent): Record<string, GameplayAttributeData> {
    const result: Record<string, GameplayAttributeData> = {};
    const debugInfo = asc.getDebugInfo();
    
    Object.entries(debugInfo.attributes).forEach(([name, attrInfo]) => {
      result[name] = {
        name,
        baseValue: attrInfo.baseValue || 0,
        currentValue: attrInfo.currentValue || 0,
        maxValue: attrInfo.maxValue,
        modifiers: attrInfo.modifiers || []
      };
    });
    
    return result;
  }
  
  private serializeAbilities(
    asc: EnhancedAbilitySystemComponent, 
    context: SerializationContext
  ): SerializableAbility[] {
    const abilities = asc.getAllAbilities();
    
    return abilities.map(ability => ({
      id: ability.id,
      name: ability.name,
      description: ability.description,
      cooldown: ability.cooldown,
      type: ability.constructor.name,
      data: context.includeMethods && 'serialize' in ability ? 
        (ability as any).serialize() : undefined
    }));
  }
  
  private serializeActiveEffects(asc: EnhancedAbilitySystemComponent): ActiveGameplayEffect[] {
    return asc.getAllActiveEffects();
  }
  
  private serializeCooldowns(asc: EnhancedAbilitySystemComponent): Record<string, { endTime: number; duration: number }> {
    const result: Record<string, { endTime: number; duration: number }> = {};
    const debugInfo = asc.getDebugInfo();
    
    Object.entries(debugInfo.cooldowns).forEach(([abilityId, cooldownInfo]) => {
      result[abilityId] = {
        endTime: Date.now() + cooldownInfo.remaining,
        duration: cooldownInfo.total
      };
    });
    
    return result;
  }
  
  // === UTILITY METHODS ===
  
  createSnapshot(
    asc: EnhancedAbilitySystemComponent,
    name?: string,
    options?: SerializationOptions
  ): SerializedData {
    const context: SerializationContext = {
      includeOwner: true,
      includeQueue: true,
      includeStats: true,
      includeMethods: false,
      customData: {
        snapshotName: name || `snapshot_${Date.now()}`,
        createdAt: new Date().toISOString()
      }
    };
    
    return this.serializeASC(asc, context, options);
  }
  
  createSaveState(
    asc: EnhancedAbilitySystemComponent,
    options?: SerializationOptions
  ): SerializedData {
    const context: SerializationContext = {
      includeOwner: false, // Don't save owner reference
      includeQueue: false, // Don't save queue (temporary state)
      includeStats: false, // Don't save stats (runtime data)
      includeMethods: false // Security: never save methods
    };
    
    return this.serializeASC(asc, context, options);
  }
  
  exportConfiguration(
    asc: EnhancedAbilitySystemComponent,
    options?: SerializationOptions
  ): SerializedData {
    const context: SerializationContext = {
      includeOwner: false,
      includeQueue: false,
      includeStats: false,
      includeMethods: false,
      customData: {
        exportType: 'configuration',
        exportedAt: new Date().toISOString()
      }
    };
    
    return this.serializeASC(asc, context, {
      ...options,
      codec: 'json' // Force JSON for configurations
    });
  }
  
  // === FORMAT CONVERSION ===
  
  convertFormat(
    serialized: SerializedData,
    targetCodec: string,
    options?: SerializationOptions
  ): SerializedData {
    // Deserialize with original codec
    const data = registry.deserialize(serialized);
    
    // Re-serialize with target codec
    return registry.serialize(data, {
      ...options,
      codec: targetCodec
    });
  }
  
  // === VALIDATION ===
  
  validateSerializedData(serialized: SerializedData): boolean {
    try {
      const data = registry.deserialize<SerializableASCState>(serialized);
      
      // Basic structure validation
      return !!(
        data.version &&
        data.timestamp &&
        data.attributes &&
        data.abilities &&
        Array.isArray(data.abilities) &&
        typeof data.attributes === 'object'
      );
    } catch {
      return false;
    }
  }
  
  getMetadata(serialized: SerializedData): Record<string, any> {
    return {
      codec: serialized.codec,
      version: serialized.version,
      timestamp: serialized.timestamp,
      compressed: serialized.compressed,
      ...serialized.metadata
    };
  }
}

// Export singleton instance
export const gasSerializer = GASSerializer.getInstance();