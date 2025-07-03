// Unit Tests for Serialization Codec Registry and Base Classes
// Tests the core serialization infrastructure

import { 
  SerializationCodecRegistry, 
  BaseSerializationCodec, 
  getSerializationRegistry,
  SerializationCodec,
  SerializedData 
} from '../SerializationCodec';

// Mock codec for testing
class MockCodec extends BaseSerializationCodec {
  readonly name = 'mock';
  readonly version = '1.0.0';
  readonly mimeType = 'application/mock';
  readonly supportsCompression = false;

  serialize<T>(data: T): string {
    return `MOCK:${JSON.stringify(data)}`;
  }

  deserialize<T>(serialized: string): T {
    if (!serialized.startsWith('MOCK:')) {
      throw new Error('Invalid mock format');
    }
    return JSON.parse(serialized.substring(5));
  }

  validate(data: any): boolean {
    return super.validate?.(data) ?? false;
  }
}

class CompressibleMockCodec extends BaseSerializationCodec {
  readonly name = 'compressible-mock';
  readonly version = '2.0.0';
  readonly mimeType = 'application/compressible-mock';
  readonly supportsCompression = true;

  serialize<T>(data: T): string {
    return `COMPRESSED_MOCK:${JSON.stringify(data)}`;
  }

  deserialize<T>(serialized: string): T {
    if (!serialized.startsWith('COMPRESSED_MOCK:')) {
      throw new Error('Invalid compressible mock format');
    }
    return JSON.parse(serialized.substring(17));
  }
}

describe('BaseSerializationCodec', () => {
  let codec: MockCodec;

  beforeEach(() => {
    codec = new MockCodec();
  });

  test('should have correct properties', () => {
    expect(codec.name).toBe('mock');
    expect(codec.version).toBe('1.0.0');
    expect(codec.mimeType).toBe('application/mock');
    expect(codec.supportsCompression).toBe(false);
  });

  test('should serialize data correctly', () => {
    const data = { test: 'value' };
    const result = codec.serialize(data);
    expect(result).toBe('MOCK:{"test":"value"}');
  });

  test('should deserialize data correctly', () => {
    const serialized = 'MOCK:{"test":"value"}';
    const result = codec.deserialize(serialized);
    expect(result).toEqual({ test: 'value' });
  });

  test('should validate non-null data', () => {
    expect(codec.validate({ test: 'value' })).toBe(true);
    expect(codec.validate('string')).toBe(true);
    expect(codec.validate(42)).toBe(true);
    expect(codec.validate([])).toBe(true);
  });

  test('should reject null and undefined during validation', () => {
    expect(codec.validate(null)).toBe(false);
    expect(codec.validate(undefined)).toBe(false);
  });

  test('should provide metadata', () => {
    const metadata = codec.getMetadata?.();
    expect(metadata).toEqual({
      name: 'mock',
      version: '1.0.0',
      mimeType: 'application/mock',
      supportsCompression: false
    });
  });
});

describe('SerializationCodecRegistry', () => {
  let registry: SerializationCodecRegistry;

  beforeEach(() => {
    registry = new SerializationCodecRegistry();
  });

  describe('Codec Management', () => {
    test('should register a codec', () => {
      const codec = new MockCodec();
      registry.register(codec);
      
      expect(registry.get('mock')).toBe(codec);
    });

    test('should register codec with case-insensitive name', () => {
      const codec = new MockCodec();
      registry.register(codec);
      
      expect(registry.get('MOCK')).toBe(codec);
      expect(registry.get('Mock')).toBe(codec);
      expect(registry.get('mock')).toBe(codec);
    });

    test('should unregister a codec', () => {
      const codec = new MockCodec();
      registry.register(codec);
      
      expect(registry.unregister('mock')).toBe(true);
      expect(registry.get('mock')).toBeUndefined();
    });

    test('should return false when unregistering non-existent codec', () => {
      expect(registry.unregister('nonexistent')).toBe(false);
    });

    test('should list all registered codecs', () => {
      const codec1 = new MockCodec();
      const codec2 = new CompressibleMockCodec();
      
      registry.register(codec1);
      registry.register(codec2);
      
      const list = registry.list();
      expect(list).toContain('mock');
      expect(list).toContain('compressible-mock');
      expect(list).toHaveLength(2);
    });

    test('should clear all codecs', () => {
      registry.register(new MockCodec());
      registry.register(new CompressibleMockCodec());
      
      expect(registry.list()).toHaveLength(2);
      
      registry.clear();
      expect(registry.list()).toHaveLength(0);
    });
  });

  describe('Default Codec Management', () => {
    test('should set and get default codec', () => {
      const codec = new MockCodec();
      registry.register(codec);
      
      expect(registry.setDefault('mock')).toBe(true);
      expect(registry.getDefault()).toBe(codec);
    });

    test('should return false when setting non-existent codec as default', () => {
      expect(registry.setDefault('nonexistent')).toBe(false);
    });

    test('should return undefined when no default is set', () => {
      expect(registry.getDefault()).toBeUndefined();
    });
  });

  describe('Serialization Operations', () => {
    beforeEach(() => {
      registry.register(new MockCodec());
      registry.setDefault('mock');
    });

    test('should serialize with default codec', () => {
      const data = { test: 'value' };
      const result = registry.serialize(data);
      
      expect(result.codec).toBe('mock');
      expect(result.version).toBe('1.0.0');
      expect(result.data).toBe('MOCK:{"test":"value"}');
      expect(result.compressed).toBe(false);
      expect(typeof result.timestamp).toBe('number');
    });

    test('should serialize with specified codec', () => {
      registry.register(new CompressibleMockCodec());
      
      const data = { test: 'value' };
      const result = registry.serialize(data, { codec: 'compressible-mock' });
      
      expect(result.codec).toBe('compressible-mock');
      expect(result.version).toBe('2.0.0');
    });

    test('should throw error for unknown codec', () => {
      const data = { test: 'value' };
      
      expect(() => {
        registry.serialize(data, { codec: 'unknown' });
      }).toThrow("Serialization codec 'unknown' not found");
    });

    test('should include metadata in serialized data', () => {
      const data = { test: 'value' };
      const metadata = { custom: 'metadata' };
      const result = registry.serialize(data, { metadata });
      
      expect(result.metadata).toMatchObject(metadata);
      expect(result.metadata).toMatchObject({
        name: 'mock',
        version: '1.0.0',
        mimeType: 'application/mock',
        supportsCompression: false,
        custom: 'metadata'
      });
    });

    test('should handle compression option', () => {
      registry.register(new CompressibleMockCodec());
      
      const data = { test: 'value' };
      const result = registry.serialize(data, { 
        codec: 'compressible-mock',
        compression: true 
      });
      
      expect(result.compressed).toBe(true);
    });

    test('should validate data when validation is requested', () => {
      const validData = { test: 'value' };
      const invalidData = null;
      
      expect(() => {
        registry.serialize(validData, { validation: true });
      }).not.toThrow();
      
      expect(() => {
        registry.serialize(invalidData, { validation: true });
      }).toThrow('Data validation failed');
    });
  });

  describe('Deserialization Operations', () => {
    beforeEach(() => {
      registry.register(new MockCodec());
    });

    test('should deserialize with correct codec', () => {
      const serializedData: SerializedData = {
        codec: 'mock',
        version: '1.0.0',
        timestamp: Date.now(),
        compressed: false,
        data: 'MOCK:{"test":"value"}'
      };
      
      const result = registry.deserialize(serializedData);
      expect(result).toEqual({ test: 'value' });
    });

    test('should throw error for unknown codec during deserialization', () => {
      const serializedData: SerializedData = {
        codec: 'unknown',
        version: '1.0.0',
        timestamp: Date.now(),
        compressed: false,
        data: 'some data'
      };
      
      expect(() => {
        registry.deserialize(serializedData);
      }).toThrow("Serialization codec 'unknown' not found");
    });

    test('should warn about version mismatch', () => {
      const consoleSpy = jest.spyOn(console, 'warn').mockImplementation();
      
      const serializedData: SerializedData = {
        codec: 'mock',
        version: '2.0.0', // Different from codec version 1.0.0
        timestamp: Date.now(),
        compressed: false,
        data: 'MOCK:{"test":"value"}'
      };
      
      registry.deserialize(serializedData);
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'Codec version mismatch: expected 1.0.0, got 2.0.0'
      );
      
      consoleSpy.mockRestore();
    });
  });
});

describe('getSerializationRegistry', () => {
  test('should return singleton instance', () => {
    const registry1 = getSerializationRegistry();
    const registry2 = getSerializationRegistry();
    
    expect(registry1).toBe(registry2);
    expect(registry1).toBeInstanceOf(SerializationCodecRegistry);
  });

  test('should maintain state across calls', () => {
    const registry1 = getSerializationRegistry();
    registry1.register(new MockCodec());
    
    const registry2 = getSerializationRegistry();
    expect(registry2.get('mock')).toBeInstanceOf(MockCodec);
  });
});

describe('Integration Tests', () => {
  let registry: SerializationCodecRegistry;

  beforeEach(() => {
    registry = getSerializationRegistry();
    registry.clear(); // Clean slate for each test
    registry.register(new MockCodec());
    registry.register(new CompressibleMockCodec());
    registry.setDefault('mock');
  });

  test('should handle complete serialize-deserialize cycle', () => {
    const originalData = {
      string: 'test',
      number: 42,
      boolean: true,
      array: [1, 2, 3],
      object: { nested: 'value' },
      null: null
    };
    
    const serialized = registry.serialize(originalData);
    const deserialized = registry.deserialize(serialized);
    
    expect(deserialized).toEqual(originalData);
  });

  test('should handle multiple codecs in same registry', () => {
    const data = { message: 'hello world' };
    
    const mockSerialized = registry.serialize(data, { codec: 'mock' });
    const compressibleSerialized = registry.serialize(data, { codec: 'compressible-mock' });
    
    expect(mockSerialized.codec).toBe('mock');
    expect(compressibleSerialized.codec).toBe('compressible-mock');
    
    const mockDeserialized = registry.deserialize(mockSerialized);
    expect(mockDeserialized).toEqual(data);
    
    // Note: compressible mock codec has different format, test separately
    const compressibleData = 'COMPRESSED_MOCK:' + JSON.stringify(data);
    const compressibleCodec = registry.get('compressible-mock');
    const compressibleDeserialized = compressibleCodec?.deserialize(compressibleData);
    expect(compressibleDeserialized).toEqual(data);
  });

  test('should preserve all metadata through serialize-deserialize cycle', () => {
    const data = { test: 'data' };
    const options = {
      metadata: { source: 'test', version: '1.0' },
      validation: true
    };
    
    const serialized = registry.serialize(data, options);
    
    expect(serialized.metadata).toMatchObject({
      source: 'test',
      version: '1.0',
      name: 'mock'
    });
    
    const deserialized = registry.deserialize(serialized);
    expect(deserialized).toEqual(data);
  });
});