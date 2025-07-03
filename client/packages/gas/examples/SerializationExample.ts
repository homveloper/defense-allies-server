// Example: GAS v2 Serialization System
// Demonstrates JSON serialization, multiple codecs, and real-world usage scenarios

import { v2 } from '../index';
import { 
  gasSerializer, 
  JsonCodecFactory, 
  getSerializationRegistry,
  SerializationContext 
} from '../v2/serialization';

// Example serializable ability
class SerializableFireballAbility extends v2.GameplayAbility {
  readonly id = 'serializable_fireball';
  readonly name = 'Serializable Fireball';
  readonly description = 'A fireball that can be saved and restored';
  readonly cooldown = 3000;
  
  // Custom ability data
  private damage: number = 50;
  private radius: number = 100;
  private element: string = 'fire';
  
  constructor(damage?: number, radius?: number, element?: string) {
    super();
    if (damage !== undefined) this.damage = damage;
    if (radius !== undefined) this.radius = radius;
    if (element !== undefined) this.element = element;
  }
  
  async activate(context: any): Promise<boolean> {
    console.log(`🔥 Fireball deals ${this.damage} ${this.element} damage in ${this.radius}px radius`);
    return true;
  }
  
  // Custom serialization for ability-specific data
  serialize(): Record<string, any> {
    return {
      damage: this.damage,
      radius: this.radius,
      element: this.element
    };
  }
  
  deserialize(data: Record<string, any>): void {
    this.damage = data.damage || 50;
    this.radius = data.radius || 100;
    this.element = data.element || 'fire';
  }
  
  // Getters for external access
  getDamage(): number { return this.damage; }
  getRadius(): number { return this.radius; }
  getElement(): string { return this.element; }
}

class SerializableHealAbility extends v2.GameplayAbility {
  readonly id = 'serializable_heal';
  readonly name = 'Serializable Heal';
  readonly description = 'A heal ability with custom properties';
  readonly cooldown = 5000;
  
  private healAmount: number = 30;
  private healOverTime: boolean = false;
  private duration: number = 0;
  
  constructor(healAmount?: number, healOverTime?: boolean, duration?: number) {
    super();
    if (healAmount !== undefined) this.healAmount = healAmount;
    if (healOverTime !== undefined) this.healOverTime = healOverTime;
    if (duration !== undefined) this.duration = duration;
  }
  
  async activate(context: any): Promise<boolean> {
    if (this.healOverTime) {
      console.log(`💚 Healing ${this.healAmount} over ${this.duration}ms`);
    } else {
      console.log(`💚 Instant heal: ${this.healAmount}`);
    }
    return true;
  }
  
  serialize(): Record<string, any> {
    return {
      healAmount: this.healAmount,
      healOverTime: this.healOverTime,
      duration: this.duration
    };
  }
  
  deserialize(data: Record<string, any>): void {
    this.healAmount = data.healAmount || 30;
    this.healOverTime = data.healOverTime || false;
    this.duration = data.duration || 0;
  }
}

// Create an ability registry for deserialization
const abilityRegistry = new Map<string, new() => v2.IGameplayAbility>();
abilityRegistry.set('SerializableFireballAbility', SerializableFireballAbility);
abilityRegistry.set('SerializableHealAbility', SerializableHealAbility);

// Example player class
class SerializablePlayer {
  constructor(public name: string) {
    this.abilitySystem = new v2.AbilitySystemComponent(this);
    this.setupAbilities();
  }
  
  abilitySystem: v2.AbilitySystemComponent;
  
  private setupAbilities(): void {
    // Add attributes
    this.abilitySystem.addAttribute('health', 100, 100);
    this.abilitySystem.addAttribute('mana', 50, 50);
    this.abilitySystem.addAttribute('experience', 0);
    
    // Grant abilities with custom properties
    this.abilitySystem.grantAbility(new SerializableFireballAbility(75, 150, 'ice')); // Ice fireball
    this.abilitySystem.grantAbility(new SerializableHealAbility(50, true, 5000)); // HoT heal
    
    // Add some tags
    this.abilitySystem.addTag('player');
    this.abilitySystem.addTag('mage');
    
    // Apply some effects
    const buffEffect = v2.GameplayEffect.createDurationBased({
      id: 'strength_buff',
      name: 'Strength Buff',
      duration: 10000,
      attributeModifiers: [{
        id: 'str_buff_mod',
        attribute: 'mana',
        operation: 'add',
        magnitude: 20,
        source: 'strength_buff'
      }]
    });
    
    this.abilitySystem.applyGameplayEffect(buffEffect);
  }
}

// === EXAMPLE SCENARIOS ===

async function demonstrateBasicSerialization(): Promise<void> {
  console.log('\n=== Basic Serialization Example ===');
  
  const player = new SerializablePlayer('Wizard');
  
  // Create a snapshot
  console.log('Creating snapshot...');
  const snapshot = gasSerializer.createSnapshot(player.abilitySystem, 'demo_snapshot');
  
  console.log('Snapshot metadata:', gasSerializer.getMetadata(snapshot));
  console.log('Snapshot size:', JSON.stringify(snapshot).length, 'bytes');
  
  // Validate the snapshot
  const isValid = gasSerializer.validateSerializedData(snapshot);
  console.log('Snapshot valid:', isValid);
}

async function demonstrateAttributeSerialization(): Promise<void> {
  console.log('\n=== Attribute Serialization Example ===');
  
  const player = new SerializablePlayer('Paladin');
  
  // Modify some attributes
  player.abilitySystem.setAttributeValue('health', 75);
  player.abilitySystem.setAttributeValue('experience', 1250);
  
  // Serialize individual attribute
  const healthAttr = player.abilitySystem.getAttribute('health');
  if (healthAttr) {
    const serializedHealth = gasSerializer.serializeAttribute(healthAttr);
    console.log('Serialized health attribute:', serializedHealth.data);
    
    // Deserialize it back
    const restoredHealth = gasSerializer.deserializeAttribute(serializedHealth);
    console.log('Restored health:', {
      name: restoredHealth.name,
      current: restoredHealth.currentValue,
      max: restoredHealth.maxValue
    });
  }
}

async function demonstrateAbilitySerialization(): Promise<void> {
  console.log('\n=== Ability Serialization Example ===');
  
  const customFireball = new SerializableFireballAbility(100, 200, 'lightning');
  
  // Serialize the ability
  const serializedAbility = gasSerializer.serializeAbility(customFireball, { includeMethods: true });
  console.log('Serialized ability data:', serializedAbility.data);
  
  // Deserialize it back
  try {
    const restoredAbility = gasSerializer.deserializeAbility(serializedAbility, abilityRegistry) as SerializableFireballAbility;
    console.log('Restored ability:', {
      id: restoredAbility.id,
      name: restoredAbility.name,
      damage: restoredAbility.getDamage(),
      radius: restoredAbility.getRadius(),
      element: restoredAbility.getElement()
    });
  } catch (error) {
    console.error('Ability deserialization failed:', error);
  }
}

async function demonstrateCompleteSaveState(): Promise<void> {
  console.log('\n=== Complete Save State Example ===');
  
  const player = new SerializablePlayer('Adventurer');
  
  // Make some changes to the game state
  player.abilitySystem.setAttributeValue('health', 65);
  player.abilitySystem.setAttributeValue('mana', 30);
  player.abilitySystem.addTag('battle_hardened');
  
  // Queue some abilities
  player.abilitySystem.queueAbility('serializable_fireball', {
    owner: player,
    scene: null as any
  }, { priority: 1, delay: 1000 });
  
  player.abilitySystem.queueAbility('serializable_heal', {
    owner: player,
    scene: null as any
  }, { priority: 2 });
  
  // Create save state (excludes temporary data like queue)
  const saveState = gasSerializer.createSaveState(player.abilitySystem);
  console.log('Save state created, size:', JSON.stringify(saveState).length, 'bytes');
  
  // Create full snapshot (includes everything)
  const fullSnapshot = gasSerializer.createSnapshot(player.abilitySystem, 'complete_state');
  console.log('Full snapshot size:', JSON.stringify(fullSnapshot).length, 'bytes');
  
  console.log('Queue included in snapshot:', 
    JSON.stringify(fullSnapshot.data).includes('queuedAbilities'));
  console.log('Queue included in save state:', 
    JSON.stringify(saveState.data).includes('queuedAbilities'));
}

async function demonstrateMultipleCodecs(): Promise<void> {
  console.log('\n=== Multiple Codecs Example ===');
  
  const player = new SerializablePlayer('Codec Tester');
  const registry = getSerializationRegistry();
  
  // Get available codecs
  console.log('Available codecs:', registry.list());
  
  // Serialize with different codecs
  const testData = { message: 'Hello GAS!', timestamp: Date.now() };
  
  const jsonSerialized = registry.serialize(testData, { codec: 'json' });
  console.log('JSON serialized:', {
    codec: jsonSerialized.codec,
    size: JSON.stringify(jsonSerialized.data).length
  });
  
  // Convert between formats
  try {
    const converted = gasSerializer.convertFormat(jsonSerialized, 'json', { compression: true });
    console.log('Converted format:', {
      originalCodec: jsonSerialized.codec,
      newCodec: converted.codec,
      compressed: converted.compressed
    });
  } catch (error) {
    console.log('Format conversion demo (would work with multiple codecs)');
  }
}

async function demonstrateConfigurationExport(): Promise<void> {
  console.log('\n=== Configuration Export Example ===');
  
  const player = new SerializablePlayer('Config Master');
  
  // Export as configuration (clean, no runtime data)
  const config = gasSerializer.exportConfiguration(player.abilitySystem);
  
  console.log('Configuration exported:');
  console.log('- Codec:', config.codec);
  console.log('- Timestamp:', config.timestamp);
  console.log('- Size:', JSON.stringify(config.data).length, 'bytes');
  console.log('- Contains abilities:', JSON.stringify(config.data).includes('abilities'));
  console.log('- Contains queue data:', JSON.stringify(config.data).includes('queuedAbilities'));
}

// === RUN ALL EXAMPLES ===

async function runAllSerializationExamples(): Promise<void> {
  console.log('🗃️ GAS v2 - Serialization System Examples\n');
  
  try {
    await demonstrateBasicSerialization();
    await demonstrateAttributeSerialization();
    await demonstrateAbilitySerialization();
    await demonstrateCompleteSaveState();
    await demonstrateMultipleCodecs();
    await demonstrateConfigurationExport();
    
    console.log('\n✅ All serialization examples completed successfully!');
  } catch (error) {
    console.error('❌ Error running serialization examples:', error);
  }
}

// Export for use in other files
export {
  SerializableFireballAbility,
  SerializableHealAbility,
  SerializablePlayer,
  abilityRegistry,
  runAllSerializationExamples
};

// Run if this file is executed directly
if (require.main === module) {
  runAllSerializationExamples();
}