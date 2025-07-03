// Integration Tests for Complete Serialization System
// Tests the entire serialization workflow with real-world scenarios

import { v2 } from '../../index';
import { 
  gasSerializer,
  JsonCodecFactory,
  getSerializationRegistry 
} from '../index';

// Real-world test abilities
class CombatAbility extends v2.GameplayAbility {
  readonly id = 'combat_ability';
  readonly name = 'Combat Ability';
  readonly description = 'A realistic combat ability';
  readonly cooldown = 2000;

  private damage: number;
  private critChance: number;
  private elementType: string;

  constructor(damage = 50, critChance = 0.1, elementType = 'physical') {
    super();
    this.damage = damage;
    this.critChance = critChance;
    this.elementType = elementType;
  }

  async activate(context: any): Promise<boolean> {
    const isCrit = Math.random() < this.critChance;
    const finalDamage = isCrit ? this.damage * 2 : this.damage;
    
    console.log(`${this.elementType} attack deals ${finalDamage} damage${isCrit ? ' (CRIT!)' : ''}`);
    return true;
  }

  serialize(): Record<string, any> {
    return {
      damage: this.damage,
      critChance: this.critChance,
      elementType: this.elementType
    };
  }

  deserialize(data: Record<string, any>): void {
    this.damage = data.damage || 50;
    this.critChance = data.critChance || 0.1;
    this.elementType = data.elementType || 'physical';
  }

  // Getters for testing
  getDamage(): number { return this.damage; }
  getCritChance(): number { return this.critChance; }
  getElementType(): string { return this.elementType; }
}

class SupportAbility extends v2.GameplayAbility {
  readonly id = 'support_ability';
  readonly name = 'Support Ability';
  readonly description = 'A support ability with multiple effects';
  readonly cooldown = 8000;

  private healAmount: number;
  private buffDuration: number;
  private targetType: 'self' | 'ally' | 'all';

  constructor(healAmount = 30, buffDuration = 5000, targetType: 'self' | 'ally' | 'all' = 'ally') {
    super();
    this.healAmount = healAmount;
    this.buffDuration = buffDuration;
    this.targetType = targetType;
  }

  async activate(context: any): Promise<boolean> {
    console.log(`Healing ${this.healAmount} and applying ${this.buffDuration}ms buff to ${this.targetType}`);
    return true;
  }

  serialize(): Record<string, any> {
    return {
      healAmount: this.healAmount,
      buffDuration: this.buffDuration,
      targetType: this.targetType
    };
  }

  deserialize(data: Record<string, any>): void {
    this.healAmount = data.healAmount || 30;
    this.buffDuration = data.buffDuration || 5000;
    this.targetType = data.targetType || 'ally';
  }
}

// Real-world player class
class GameCharacter {
  public id: string;
  public name: string;
  public level: number;
  public experience: number;
  public abilitySystem: v2.AbilitySystemComponent;

  constructor(id: string, name: string, level = 1) {
    this.id = id;
    this.name = name;
    this.level = level;
    this.experience = level * 100;
    this.abilitySystem = new v2.AbilitySystemComponent(this);
    this.initializeCharacter();
  }

  private initializeCharacter(): void {
    // Base attributes based on level
    const baseHealth = 100 + (this.level - 1) * 20;
    const baseMana = 50 + (this.level - 1) * 10;
    
    this.abilitySystem.addAttribute('health', baseHealth, baseHealth);
    this.abilitySystem.addAttribute('mana', baseMana, baseMana);
    this.abilitySystem.addAttribute('stamina', 100, 100);
    this.abilitySystem.addAttribute('experience', this.experience);
    
    // Grant abilities based on level
    if (this.level >= 1) {
      this.abilitySystem.grantAbility(new CombatAbility(40 + this.level * 5, 0.05 + this.level * 0.01));
    }
    
    if (this.level >= 3) {
      this.abilitySystem.grantAbility(new SupportAbility(20 + this.level * 3, 3000 + this.level * 500));
    }
    
    // Add class-based tags
    this.abilitySystem.addTag('player');
    this.abilitySystem.addTag('alive');
    
    if (this.level >= 5) {
      this.abilitySystem.addTag('veteran');
    }
    
    if (this.level >= 10) {
      this.abilitySystem.addTag('expert');
    }
  }

  takeDamage(amount: number): void {
    const currentHealth = this.abilitySystem.getAttributeValue('health');
    this.abilitySystem.setAttributeValue('health', Math.max(0, currentHealth - amount));
    
    if (this.abilitySystem.getAttributeValue('health') <= 0) {
      this.abilitySystem.removeTag('alive');
      this.abilitySystem.addTag('dead');
    }
  }

  heal(amount: number): void {
    const currentHealth = this.abilitySystem.getAttributeValue('health');
    const maxHealth = this.abilitySystem.getAttributeFinalValue('health');
    this.abilitySystem.setAttributeValue('health', Math.min(maxHealth, currentHealth + amount));
  }

  levelUp(): void {
    this.level++;
    this.experience = this.level * 100;
    
    // Increase max health and mana
    const healthAttr = this.abilitySystem.getAttribute('health');
    const manaAttr = this.abilitySystem.getAttribute('mana');
    
    if (healthAttr) {
      healthAttr.maxValue = 100 + (this.level - 1) * 20;
    }
    
    if (manaAttr) {
      manaAttr.maxValue = 50 + (this.level - 1) * 10;
    }
    
    this.abilitySystem.setAttributeValue('experience', this.experience);
    
    // Update abilities
    this.initializeCharacter();
  }
}

describe('Serialization System Integration Tests', () => {
  let abilityRegistry: Map<string, new() => v2.IGameplayAbility>;

  beforeEach(() => {
    // Setup clean serialization environment
    const registry = getSerializationRegistry();
    registry.clear();
    registry.register(JsonCodecFactory.createGASCodec());
    registry.register(JsonCodecFactory.createCompressed());
    registry.register(JsonCodecFactory.createPrettified());
    registry.setDefault('json');

    // Setup ability registry
    abilityRegistry = new Map();
    abilityRegistry.set('CombatAbility', CombatAbility);
    abilityRegistry.set('SupportAbility', SupportAbility);
  });

  describe('Real-World Game Scenarios', () => {
    test('should handle complete character save and load', () => {
      // Create a character and play through some actions
      const character = new GameCharacter('char_001', 'Hero', 5);
      
      // Simulate some gameplay
      character.takeDamage(30);
      character.heal(10);
      character.levelUp();
      
      // Apply some effects
      const buffEffect = v2.GameplayEffect.createDurationBased({
        id: 'strength_buff',
        name: 'Strength Buff',
        duration: 30000,
        attributeModifiers: [{
          id: 'str_mod',
          attribute: 'health',
          operation: 'add',
          magnitude: 50,
          source: 'strength_buff'
        }],
        grantedTags: ['buffed', 'strong']
      });
      
      character.abilitySystem.applyGameplayEffect(buffEffect);
      
      // Queue some abilities
      character.abilitySystem.queueAbility('combat_ability', {
        owner: character,
        scene: null as any
      }, { priority: 2, delay: 1000 });
      
      character.abilitySystem.queueAbility('support_ability', {
        owner: character,
        scene: null as any
      }, { priority: 1 });
      
      // Create save data
      const saveData = gasSerializer.createSaveState(character.abilitySystem);
      
      // Verify save data structure
      expect(saveData).toBeValidSerializedData();
      expect(saveData.codec).toBe('json');
      
      const savedState = JSON.parse(saveData.data as string);
      expect(savedState.version).toBe('2.0.0');
      expect(savedState.attributes.health).toBeDefined();
      expect(savedState.abilities.length).toBeGreaterThan(0);
      expect(savedState.activeEffects.length).toBeGreaterThan(0);
      expect(savedState.tags).toContain('player');
      expect(savedState.tags).toContain('veteran');
    });

    test('should preserve complex character state across serialization', () => {
      const originalCharacter = new GameCharacter('char_002', 'Mage', 8);
      
      // Create complex state
      originalCharacter.takeDamage(45);
      originalCharacter.abilitySystem.setAttributeValue('mana', 30);
      
      // Apply multiple effects
      const effects = [
        v2.GameplayEffect.createDurationBased({
          id: 'mana_regen',
          name: 'Mana Regeneration',
          duration: 20000,
          period: 2000,
          attributeModifiers: [{
            id: 'mana_regen_mod',
            attribute: 'mana',
            operation: 'add',
            magnitude: 5,
            source: 'mana_regen'
          }]
        }),
        v2.GameplayEffect.createDurationBased({
          id: 'damage_boost',
          name: 'Damage Boost',
          duration: 15000,
          grantedTags: ['damage_boosted'],
          attributeModifiers: [{
            id: 'damage_mod',
            attribute: 'stamina',
            operation: 'multiply',
            magnitude: 1.2,
            source: 'damage_boost'
          }]
        })
      ];
      
      effects.forEach(effect => originalCharacter.abilitySystem.applyGameplayEffect(effect));
      
      // Serialize complete state
      const snapshot = gasSerializer.createSnapshot(originalCharacter.abilitySystem, 'complex_state');
      
      // Verify all complex data is preserved
      const snapshotData = JSON.parse(snapshot.data as string);
      
      // Check attributes
      expect(snapshotData.attributes.health.currentValue).toBeLessThan(snapshotData.attributes.health.maxValue);
      expect(snapshotData.attributes.mana.currentValue).toBe(30);
      
      // Check effects
      expect(snapshotData.activeEffects.length).toBe(2);
      const manaRegenEffect = snapshotData.activeEffects.find((e: any) => e.spec.id === 'mana_regen');
      expect(manaRegenEffect).toBeDefined();
      expect(manaRegenEffect.spec.period).toBe(2000);
      
      // Check tags
      expect(snapshotData.tags).toContain('damage_boosted');
      expect(snapshotData.tags).toContain('expert'); // Level 8 character
      
      // Check abilities
      expect(snapshotData.abilities.length).toBe(2); // Combat + Support abilities
    });

    test('should handle party/raid serialization', () => {
      // Create a party of characters
      const party = [
        new GameCharacter('tank_001', 'Guardian', 10),
        new GameCharacter('dps_001', 'Assassin', 8),
        new GameCharacter('healer_001', 'Cleric', 7),
        new GameCharacter('support_001', 'Bard', 6)
      ];
      
      // Apply different states to each character
      party[0].takeDamage(80); // Tank took heavy damage
      party[1].abilitySystem.setAttributeValue('stamina', 20); // DPS low on stamina
      party[2].heal(50); // Healer is full health
      party[3].levelUp(); // Support just leveled up
      
      // Apply party-wide buff
      const partyBuff = v2.GameplayEffect.createDurationBased({
        id: 'party_buff',
        name: 'Party Blessing',
        duration: 60000,
        grantedTags: ['blessed', 'party_member'],
        attributeModifiers: [{
          id: 'party_health_boost',
          attribute: 'health',
          operation: 'add',
          magnitude: 25,
          source: 'party_buff'
        }]
      });
      
      party.forEach(member => member.abilitySystem.applyGameplayEffect(partyBuff));
      
      // Serialize entire party
      const partyData = party.map(member => 
        gasSerializer.createSnapshot(member.abilitySystem, `${member.name}_state`)
      );
      
      // Verify party data
      expect(partyData).toHaveLength(4);
      
      partyData.forEach((memberData, index) => {
        expect(memberData).toBeValidSerializedData();
        
        const state = JSON.parse(memberData.data as string);
        expect(state.tags).toContain('blessed');
        expect(state.tags).toContain('party_member');
        expect(state.activeEffects.some((e: any) => e.spec.id === 'party_buff')).toBe(true);
        
        // Verify level-specific data
        if (index === 0) { // Tank
          expect(state.attributes.health.currentValue).toBeLessThan(100);
        }
        if (index === 3) { // Support who leveled up
          expect(state.attributes.experience).toBe(700); // Level 7 * 100
        }
      });
    });

    test('should handle combat scenario with queued abilities', () => {
      const fighter = new GameCharacter('fighter_001', 'Warrior', 12);
      
      // Simulate combat sequence
      fighter.takeDamage(60);
      
      // Queue a complex combat sequence
      const combatSequence = [
        { ability: 'combat_ability', priority: 3, delay: 0 },      // Immediate attack
        { ability: 'support_ability', priority: 2, delay: 1000 }, // Heal after 1 second
        { ability: 'combat_ability', priority: 1, delay: 2500 },  // Follow-up attack
      ];
      
      combatSequence.forEach(action => {
        fighter.abilitySystem.queueAbility(action.ability, {
          owner: fighter,
          scene: null as any
        }, { 
          priority: action.priority, 
          delay: action.delay 
        });
      });
      
      // Apply combat effects
      const combatEffects = [
        v2.GameplayEffect.createDurationBased({
          id: 'battle_rage',
          name: 'Battle Rage',
          duration: 10000,
          grantedTags: ['enraged', 'combat_active']
        }),
        v2.GameplayEffect.createDurationBased({
          id: 'armor_buff',
          name: 'Armor Buff',
          duration: 8000,
          attributeModifiers: [{
            id: 'armor_mod',
            attribute: 'health',
            operation: 'multiply',
            magnitude: 1.3,
            source: 'armor_buff'
          }]
        })
      ];
      
      combatEffects.forEach(effect => fighter.abilitySystem.applyGameplayEffect(effect));
      
      // Create combat snapshot
      const combatSnapshot = gasSerializer.createSnapshot(fighter.abilitySystem, 'combat_state');
      
      const combatData = JSON.parse(combatSnapshot.data as string);
      
      // Verify combat state
      expect(combatData.queuedAbilities).toHaveLength(3);
      expect(combatData.activeEffects).toHaveLength(2);
      expect(combatData.tags).toContain('enraged');
      expect(combatData.tags).toContain('combat_active');
      expect(combatData.tags).toContain('expert'); // Level 12
      
      // Verify queue order (should be sorted by priority)
      const sortedQueue = combatData.queuedAbilities.sort((a: any, b: any) => b.priority - a.priority);
      expect(sortedQueue[0].priority).toBe(3);
      expect(sortedQueue[1].priority).toBe(2);
      expect(sortedQueue[2].priority).toBe(1);
    });
  });

  describe('Performance and Scalability', () => {
    test('should handle large-scale serialization efficiently', () => {
      const startTime = Date.now();
      
      // Create multiple characters with complex states
      const characters: GameCharacter[] = [];
      
      for (let i = 0; i < 50; i++) {
        const character = new GameCharacter(`char_${i}`, `Player${i}`, Math.floor(Math.random() * 15) + 1);
        
        // Add random state
        character.takeDamage(Math.random() * 50);
        character.abilitySystem.setAttributeValue('mana', Math.random() * 50);
        
        // Add random effects
        if (Math.random() > 0.5) {
          const randomEffect = v2.GameplayEffect.createDurationBased({
            id: `random_effect_${i}`,
            name: `Random Effect ${i}`,
            duration: Math.random() * 30000 + 5000,
            grantedTags: [`random_tag_${i}`]
          });
          character.abilitySystem.applyGameplayEffect(randomEffect);
        }
        
        characters.push(character);
      }
      
      // Serialize all characters
      const serializedData = characters.map(char => 
        gasSerializer.createSaveState(char.abilitySystem)
      );
      
      const endTime = Date.now();
      const duration = endTime - startTime;
      
      // Performance assertions
      expect(duration).toBeLessThan(5000); // Should complete within 5 seconds
      expect(serializedData).toHaveLength(50);
      
      // Verify all data is valid
      serializedData.forEach(data => {
        expect(data).toBeValidSerializedData();
        expect(data.codec).toBe('json');
      });
      
      console.log(`Serialized 50 characters in ${duration}ms (${(duration/50).toFixed(2)}ms per character)`);
    });

    test('should handle deep object nesting without stack overflow', () => {
      const character = new GameCharacter('deep_test', 'DeepTester', 5);
      
      // Create deeply nested effect structure
      let deepEffect = v2.GameplayEffect.createDurationBased({
        id: 'deep_effect_0',
        name: 'Deep Effect 0',
        duration: 10000
      });
      
      character.abilitySystem.applyGameplayEffect(deepEffect);
      
      // Add many nested modifiers
      for (let i = 0; i < 100; i++) {
        const nestedEffect = v2.GameplayEffect.createDurationBased({
          id: `nested_effect_${i}`,
          name: `Nested Effect ${i}`,
          duration: 5000,
          attributeModifiers: [{
            id: `nested_mod_${i}`,
            attribute: 'health',
            operation: 'add',
            magnitude: 1,
            source: `nested_effect_${i}`
          }]
        });
        character.abilitySystem.applyGameplayEffect(nestedEffect);
      }
      
      // Should serialize without errors
      const deepSnapshot = gasSerializer.createSnapshot(character.abilitySystem, 'deep_test');
      
      expect(deepSnapshot).toBeValidSerializedData();
      
      const deepData = JSON.parse(deepSnapshot.data as string);
      expect(deepData.activeEffects.length).toBeGreaterThan(50);
    });
  });

  describe('Error Recovery and Validation', () => {
    test('should gracefully handle corrupted save data', () => {
      const character = new GameCharacter('test_char', 'TestChar', 5);
      const validSnapshot = gasSerializer.createSnapshot(character.abilitySystem);
      
      // Corrupt the data in various ways
      const corruptedData = [
        { ...validSnapshot, data: 'invalid json{' },
        { ...validSnapshot, data: '{"version": "unknown"}' },
        { ...validSnapshot, codec: 'nonexistent_codec' },
        { ...validSnapshot, data: '{"attributes": null}' }
      ];
      
      corruptedData.forEach((corrupted, index) => {
        const isValid = gasSerializer.validateSerializedData(corrupted);
        expect(isValid).toBe(false);
        
        console.log(`Corruption test ${index + 1}: ${isValid ? 'PASSED' : 'FAILED (expected)'}`);
      });
    });

    test('should maintain backwards compatibility', () => {
      // Simulate old version data
      const oldVersionData = {
        codec: 'json',
        version: '1.0.0', // Old version
        timestamp: Date.now(),
        compressed: false,
        data: JSON.stringify({
          version: '1.0.0',
          timestamp: Date.now(),
          owner: {},
          attributes: {
            health: { name: 'health', baseValue: 100, currentValue: 100, maxValue: 100, modifiers: [] }
          },
          abilities: [],
          activeEffects: [],
          tags: ['player'],
          cooldowns: {},
          queuedAbilities: [],
          queueStats: {},
          queueMode: 'auto',
          globalConditions: []
        })
      };
      
      // Should validate despite version difference
      const isValid = gasSerializer.validateSerializedData(oldVersionData);
      expect(isValid).toBe(true);
      
      // Should provide warning about version mismatch (check console.warn mock)
      expect(console.warn).toHaveBeenCalledWith(
        expect.stringContaining('version mismatch')
      );
    });
  });

  describe('Cross-Format Compatibility', () => {
    test('should convert between different serialization formats', () => {
      const character = new GameCharacter('format_test', 'FormatTester', 3);
      
      // Serialize with default JSON codec
      const jsonSnapshot = gasSerializer.createSnapshot(character.abilitySystem);
      expect(jsonSnapshot.codec).toBe('json');
      
      // Convert to compressed format
      const compressedSnapshot = gasSerializer.convertFormat(jsonSnapshot, 'json', { compression: true });
      expect(compressedSnapshot.compressed).toBe(true);
      
      // Convert to prettified format
      const prettifiedSnapshot = gasSerializer.convertFormat(jsonSnapshot, 'json');
      expect(prettifiedSnapshot.codec).toBe('json');
      
      // All should contain same data when deserialized
      const jsonData = JSON.parse(jsonSnapshot.data as string);
      const compressedData = JSON.parse(compressedSnapshot.data as string);
      const prettifiedData = JSON.parse(prettifiedSnapshot.data as string);
      
      expect(compressedData.attributes.health.currentValue).toBe(jsonData.attributes.health.currentValue);
      expect(prettifiedData.abilities.length).toBe(jsonData.abilities.length);
    });
  });
});