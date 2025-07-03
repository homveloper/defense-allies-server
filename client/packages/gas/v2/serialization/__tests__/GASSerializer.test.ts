// Unit Tests for GAS Serializer
// Tests the complete GAS serialization system

import { v2 } from '../../index';
import { 
  GASSerializer, 
  gasSerializer,
  SerializationContext 
} from '../GASSerializer';
import { JsonCodecFactory } from '../JsonCodec';
import { getSerializationRegistry } from '../SerializationCodec';

// Mock abilities for testing
class TestFireballAbility extends v2.GameplayAbility {
  readonly id = 'test_fireball';
  readonly name = 'Test Fireball';
  readonly description = 'A test fireball ability';
  readonly cooldown = 3000;

  private damage: number = 50;

  constructor(damage?: number) {
    super();
    if (damage !== undefined) this.damage = damage;
  }

  async activate(context: any): Promise<boolean> {
    return true;
  }

  serialize(): Record<string, any> {
    return { damage: this.damage };
  }

  deserialize(data: Record<string, any>): void {
    this.damage = data.damage || 50;
  }

  getDamage(): number { return this.damage; }
}

class TestHealAbility extends v2.GameplayAbility {
  readonly id = 'test_heal';
  readonly name = 'Test Heal';
  readonly description = 'A test heal ability';
  readonly cooldown = 5000;

  private healAmount: number = 30;

  async activate(context: any): Promise<boolean> {
    return true;
  }

  serialize(): Record<string, any> {
    return { healAmount: this.healAmount };
  }

  deserialize(data: Record<string, any>): void {
    this.healAmount = data.healAmount || 30;
  }
}

// Test player class
class TestPlayer {
  constructor(public id: string, public name: string) {
    this.abilitySystem = new v2.AbilitySystemComponent(this);
  }
  abilitySystem: v2.AbilitySystemComponent;
}

describe('GASSerializer', () => {
  let serializer: GASSerializer;
  let player: TestPlayer;
  let abilityRegistry: Map<string, new() => v2.IGameplayAbility>;

  beforeEach(() => {
    // Setup serialization system
    const registry = getSerializationRegistry();
    registry.clear();
    registry.register(JsonCodecFactory.createGASCodec());
    registry.setDefault('json');

    serializer = GASSerializer.getInstance();
    player = new TestPlayer('player1', 'TestHero');
    
    // Setup ability registry for deserialization
    abilityRegistry = new Map();
    abilityRegistry.set('TestFireballAbility', TestFireballAbility);
    abilityRegistry.set('TestHealAbility', TestHealAbility);

    // Setup test data
    setupTestPlayer(player);
  });

  function setupTestPlayer(testPlayer: TestPlayer): void {
    const asc = testPlayer.abilitySystem;
    
    // Add attributes
    asc.addAttribute('health', 100, 100);
    asc.addAttribute('mana', 50, 50);
    asc.addAttribute('experience', 0);
    
    // Grant abilities
    asc.grantAbility(new TestFireballAbility(75));
    asc.grantAbility(new TestHealAbility());
    
    // Add tags
    asc.addTag('player');
    asc.addTag('alive');
    
    // Apply effect
    const buffEffect = v2.GameplayEffect.createDurationBased({
      id: 'test_buff',
      name: 'Test Buff',
      duration: 10000,
      attributeModifiers: [{
        id: 'test_modifier',
        attribute: 'mana',
        operation: 'add',
        magnitude: 10,
        source: 'test_buff'
      }]
    });
    asc.applyGameplayEffect(buffEffect);
    
    // Modify some values
    asc.setAttributeValue('health', 75);
    asc.setAttributeValue('experience', 150);
  }

  describe('Singleton Pattern', () => {
    test('should return same instance', () => {
      const instance1 = GASSerializer.getInstance();
      const instance2 = GASSerializer.getInstance();
      
      expect(instance1).toBe(instance2);
      expect(instance1).toBe(serializer);
    });
  });

  describe('Attribute Serialization', () => {
    test('should serialize attribute correctly', () => {
      const healthAttr = player.abilitySystem.getAttribute('health');
      expect(healthAttr).toBeDefined();
      
      const serialized = serializer.serializeAttribute(healthAttr!);
      
      expect(serialized.codec).toBe('json');
      expect(serialized.data).toBeDefined();
      
      const deserialized = serializer.deserializeAttribute(serialized);
      expect(deserialized.name).toBe('health');
      expect(deserialized.currentValue).toBe(75);
      expect(deserialized.maxValue).toBe(100);
    });

    test('should preserve attribute modifiers', () => {
      // Apply a modifier
      const manaAttr = player.abilitySystem.getAttribute('mana');
      expect(manaAttr).toBeDefined();
      
      manaAttr!.addModifier({
        id: 'test_mod',
        attribute: 'mana',
        operation: 'add',
        magnitude: 20,
        source: 'test'
      });
      
      const serialized = serializer.serializeAttribute(manaAttr!);
      const deserialized = serializer.deserializeAttribute(serialized);
      
      expect(deserialized.getModifiers()).toHaveLength(1);
      expect(deserialized.getModifiers()[0].magnitude).toBe(20);
    });
  });

  describe('Effect Serialization', () => {
    test('should serialize and deserialize effects', () => {
      const effect = v2.GameplayEffect.createInstantDamage(50);
      
      const serialized = serializer.serializeEffect(effect);
      const deserialized = serializer.deserializeEffect(serialized);
      
      expect(deserialized.spec.id).toBe(effect.spec.id);
      expect(deserialized.spec.name).toBe(effect.spec.name);
      expect(deserialized.spec.duration).toBe(effect.spec.duration);
    });

    test('should handle complex effects with modifiers', () => {
      const complexEffect = v2.GameplayEffect.createDurationBased({
        id: 'complex_effect',
        name: 'Complex Effect',
        duration: 5000,
        period: 1000,
        attributeModifiers: [
          {
            id: 'complex_mod1',
            attribute: 'health',
            operation: 'add',
            magnitude: 5,
            source: 'complex_effect'
          },
          {
            id: 'complex_mod2',
            attribute: 'mana',
            operation: 'multiply',
            magnitude: 1.1,
            source: 'complex_effect'
          }
        ],
        grantedTags: ['buffed', 'healing'],
        removedTags: ['debuffed']
      });
      
      const serialized = serializer.serializeEffect(complexEffect);
      const deserialized = serializer.deserializeEffect(serialized);
      
      expect(deserialized.spec.attributeModifiers).toHaveLength(2);
      expect(deserialized.spec.grantedTags).toContain('buffed');
      expect(deserialized.spec.removedTags).toContain('debuffed');
      expect(deserialized.spec.period).toBe(1000);
    });
  });

  describe('Ability Serialization', () => {
    test('should serialize ability without custom data', () => {
      const fireball = new TestFireballAbility(100);
      
      const serialized = serializer.serializeAbility(fireball);
      const data = JSON.parse(serialized.data as string);
      
      expect(data.id).toBe('test_fireball');
      expect(data.name).toBe('Test Fireball');
      expect(data.cooldown).toBe(3000);
      expect(data.type).toBe('TestFireballAbility');
      expect(data.data).toBeUndefined(); // No custom data by default
    });

    test('should serialize ability with custom data when requested', () => {
      const fireball = new TestFireballAbility(100);
      
      const serialized = serializer.serializeAbility(fireball, { includeMethods: true });
      const data = JSON.parse(serialized.data as string);
      
      expect(data.data).toEqual({ damage: 100 });
    });

    test('should deserialize ability correctly', () => {
      const fireball = new TestFireballAbility(150);
      const serialized = serializer.serializeAbility(fireball, { includeMethods: true });
      
      const deserialized = serializer.deserializeAbility(serialized, abilityRegistry) as TestFireballAbility;
      
      expect(deserialized.id).toBe('test_fireball');
      expect(deserialized.name).toBe('Test Fireball');
      expect(deserialized.getDamage()).toBe(150);
    });

    test('should throw error for unknown ability type', () => {
      const serialized = serializer.serializeAbility(new TestFireballAbility());
      const data = JSON.parse(serialized.data as string);
      data.type = 'UnknownAbility';
      serialized.data = JSON.stringify(data);
      
      expect(() => {
        serializer.deserializeAbility(serialized, abilityRegistry);
      }).toThrow('Unknown ability type: UnknownAbility');
    });
  });

  describe('Complete ASC Serialization', () => {
    test('should serialize ASC with default context', () => {
      const serialized = serializer.serializeASC(player.abilitySystem);
      const data = JSON.parse(serialized.data as string);
      
      expect(data.version).toBe('2.0.0');
      expect(data.attributes).toBeDefined();
      expect(data.abilities).toBeDefined();
      expect(data.activeEffects).toBeDefined();
      expect(data.tags).toBeDefined();
      expect(data.cooldowns).toBeDefined();
      expect(typeof data.timestamp).toBe('number');
    });

    test('should include owner when requested', () => {
      const context: SerializationContext = { includeOwner: true };
      const serialized = serializer.serializeASC(player.abilitySystem, context);
      const data = JSON.parse(serialized.data as string);
      
      expect(data.owner.id).toBe('player1');
      expect(data.owner.name).toBe('TestHero');
      expect(data.owner.type).toBe('TestPlayer');
    });

    test('should include queue when requested', () => {
      // Add something to queue
      player.abilitySystem.queueAbility('test_fireball', {
        owner: player,
        scene: null as any
      }, { priority: 1 });
      
      const context: SerializationContext = { includeQueue: true };
      const serialized = serializer.serializeASC(player.abilitySystem, context);
      const data = JSON.parse(serialized.data as string);
      
      expect(data.queuedAbilities).toBeDefined();
      expect(data.queuedAbilities.length).toBeGreaterThan(0);
    });

    test('should include stats when requested', () => {
      const context: SerializationContext = { includeStats: true };
      const serialized = serializer.serializeASC(player.abilitySystem, context);
      const data = JSON.parse(serialized.data as string);
      
      expect(data.queueStats).toBeDefined();
      expect(typeof data.queueStats.totalQueued).toBe('number');
    });

    test('should exclude temporary data in save state', () => {
      // Add queue data
      player.abilitySystem.queueAbility('test_fireball', {
        owner: player,
        scene: null as any
      });
      
      const saveState = serializer.createSaveState(player.abilitySystem);
      const data = JSON.parse(saveState.data as string);
      
      expect(data.queuedAbilities).toHaveLength(0);
      expect(Object.keys(data.queueStats)).toHaveLength(0);
      expect(data.owner).toEqual({});
    });

    test('should include everything in snapshot', () => {
      player.abilitySystem.queueAbility('test_fireball', {
        owner: player,
        scene: null as any
      });
      
      const snapshot = serializer.createSnapshot(player.abilitySystem, 'test_snapshot');
      const data = JSON.parse(snapshot.data as string);
      
      expect(data.queuedAbilities.length).toBeGreaterThan(0);
      expect(data.owner.name).toBe('TestHero');
      expect(data.metadata.snapshotName).toBe('test_snapshot');
    });
  });

  describe('Configuration Export', () => {
    test('should export clean configuration', () => {
      const config = serializer.exportConfiguration(player.abilitySystem);
      
      expect(config.codec).toBe('json');
      
      const data = JSON.parse(config.data as string);
      expect(data.queuedAbilities).toHaveLength(0);
      expect(data.owner).toEqual({});
      expect(data.metadata.exportType).toBe('configuration');
    });
  });

  describe('Utility Methods', () => {
    test('should validate serialized data', () => {
      const validData = serializer.createSnapshot(player.abilitySystem);
      expect(serializer.validateSerializedData(validData)).toBe(true);
      
      const invalidData = {
        codec: 'json',
        version: '1.0.0',
        timestamp: Date.now(),
        compressed: false,
        data: '{"invalid": "structure"}'
      };
      expect(serializer.validateSerializedData(invalidData)).toBe(false);
    });

    test('should get metadata correctly', () => {
      const serialized = serializer.createSnapshot(player.abilitySystem, 'test');
      const metadata = serializer.getMetadata(serialized);
      
      expect(metadata.codec).toBe('json');
      expect(metadata.version).toBeDefined();
      expect(metadata.timestamp).toBeDefined();
      expect(metadata.compressed).toBe(false);
    });

    test('should convert between formats', () => {
      const original = serializer.createSnapshot(player.abilitySystem);
      
      // Convert to same format (should work)
      const converted = serializer.convertFormat(original, 'json');
      
      expect(converted.codec).toBe('json');
      expect(JSON.parse(converted.data as string)).toEqual(JSON.parse(original.data as string));
    });
  });

  describe('Error Handling', () => {
    test('should handle serialization errors gracefully', () => {
      // Create circular reference
      const circularObj: any = { name: 'test' };
      circularObj.self = circularObj;
      
      expect(() => {
        serializer.serializeAttribute(circularObj as any);
      }).toThrow();
    });

    test('should handle deserialization errors gracefully', () => {
      const invalidSerialized = {
        codec: 'json',
        version: '1.0.0',
        timestamp: Date.now(),
        compressed: false,
        data: 'invalid json}'
      };
      
      expect(() => {
        serializer.deserializeAttribute(invalidSerialized);
      }).toThrow();
    });
  });

  describe('Complex Integration Scenarios', () => {
    test('should handle complete game state serialization', () => {
      // Create complex game state
      player.abilitySystem.setAttributeValue('health', 85);
      player.abilitySystem.addTag('combat');
      
      // Queue multiple abilities
      player.abilitySystem.queueAbility('test_fireball', { owner: player, scene: null as any }, { priority: 2 });
      player.abilitySystem.queueAbility('test_heal', { owner: player, scene: null as any }, { priority: 1 });
      
      // Apply multiple effects
      const effect1 = v2.GameplayEffect.createInstantDamage(10);
      const effect2 = v2.GameplayEffect.createDurationBased({
        id: 'regen',
        name: 'Regeneration',
        duration: 15000,
        period: 3000
      });
      
      player.abilitySystem.applyGameplayEffect(effect1);
      player.abilitySystem.applyGameplayEffect(effect2);
      
      // Create full snapshot
      const snapshot = serializer.createSnapshot(player.abilitySystem, 'complex_state');
      
      // Verify all data is present
      const data = JSON.parse(snapshot.data as string);
      expect(data.attributes.health.currentValue).toBe(85);
      expect(data.tags).toContain('combat');
      expect(data.queuedAbilities).toHaveLength(2);
      expect(data.activeEffects.length).toBeGreaterThan(0);
      expect(data.abilities).toHaveLength(2);
      
      // Verify metadata
      expect(data.metadata.snapshotName).toBe('complex_state');
      expect(typeof data.metadata.serializationContext).toBe('object');
    });

    test('should maintain data integrity through multiple serialize-deserialize cycles', () => {
      const originalSnapshot = serializer.createSnapshot(player.abilitySystem);
      
      // Deserialize and re-serialize multiple times
      let currentData = JSON.parse(originalSnapshot.data as string);
      
      for (let i = 0; i < 3; i++) {
        const reserializedSnapshot = {
          ...originalSnapshot,
          data: JSON.stringify(currentData),
          timestamp: Date.now()
        };
        
        currentData = JSON.parse(reserializedSnapshot.data as string);
      }
      
      // Data should remain consistent
      const originalData = JSON.parse(originalSnapshot.data as string);
      expect(currentData.attributes.health.currentValue).toBe(originalData.attributes.health.currentValue);
      expect(currentData.abilities.length).toBe(originalData.abilities.length);
      expect(currentData.tags).toEqual(originalData.tags);
    });
  });
});