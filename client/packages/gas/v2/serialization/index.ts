// GAS v2 Serialization System
// Complete serialization solution with multiple codec support

// Core interfaces and base classes
export { 
  SerializationCodec, 
  SerializationOptions, 
  SerializedData, 
  BaseSerializationCodec,
  SerializationCodecRegistry,
  getSerializationRegistry
} from './SerializationCodec';

// JSON Codec implementation
export { 
  JsonCodec, 
  JsonCodecOptions, 
  JsonCodecFactory 
} from './JsonCodec';

// GAS-specific serialization
export { 
  GASSerializer,
  SerializableAbility,
  SerializableASCState,
  SerializationContext,
  gasSerializer
} from './GASSerializer';

// Utility functions
export function initializeDefaultCodecs(): void {
  const registry = getSerializationRegistry();
  
  // Register JSON codecs
  registry.register(JsonCodecFactory.createDefault());
  registry.register(JsonCodecFactory.createGASCodec());
  registry.register(JsonCodecFactory.createCompressed());
  
  // Set GAS codec as default
  registry.setDefault('json');
}

// Quick serialization functions
export function quickSerialize<T>(data: T, codecName?: string): SerializedData {
  return getSerializationRegistry().serialize(data, { codec: codecName });
}

export function quickDeserialize<T>(serialized: SerializedData): T {
  return getSerializationRegistry().deserialize<T>(serialized);
}

// Only auto-initialize if not in test environment
if (typeof process === 'undefined' || process.env.NODE_ENV !== 'test') {
  initializeDefaultCodecs();
}