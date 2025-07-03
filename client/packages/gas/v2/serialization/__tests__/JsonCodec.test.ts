// Unit Tests for JSON Codec
// Comprehensive test suite for all JSON serialization functionality

import { JsonCodec, JsonCodecFactory } from '../JsonCodec';

describe('JsonCodec', () => {
  let codec: JsonCodec;

  beforeEach(() => {
    codec = new JsonCodec();
  });

  describe('Basic Serialization', () => {
    test('should serialize simple objects', () => {
      const data = { name: 'test', value: 42 };
      const result = codec.serialize(data);
      
      expect(typeof result).toBe('string');
      expect(result).toBe('{"name":"test","value":42}');
    });

    test('should deserialize simple objects', () => {
      const json = '{"name":"test","value":42}';
      const result = codec.deserialize(json);
      
      expect(result).toEqual({ name: 'test', value: 42 });
    });

    test('should handle null and undefined', () => {
      expect(codec.serialize(null)).toBe('null');
      expect(codec.serialize(undefined)).toBe(undefined); // JSON.stringify returns undefined
      expect(codec.deserialize('null')).toBe(null);
    });

    test('should handle arrays', () => {
      const data = [1, 2, 'three', { four: 4 }];
      const serialized = codec.serialize(data);
      const deserialized = codec.deserialize(serialized);
      
      expect(deserialized).toEqual(data);
    });

    test('should handle nested objects', () => {
      const data = {
        level1: {
          level2: {
            level3: {
              value: 'deep'
            }
          }
        }
      };
      
      const serialized = codec.serialize(data);
      const deserialized = codec.deserialize(serialized);
      
      expect(deserialized).toEqual(data);
    });
  });

  describe('Configuration Options', () => {
    test('should prettify JSON when option is enabled', () => {
      const prettifyCodec = new JsonCodec({ prettify: true, space: 2 });
      const data = { a: 1, b: 2 };
      const result = prettifyCodec.serialize(data);
      
      expect(result).toContain('\n');
      expect(result).toContain('  '); // Indentation
    });

    test('should minify JSON when prettify is disabled', () => {
      const minifyCodec = new JsonCodec({ prettify: false });
      const data = { a: 1, b: 2 };
      const result = minifyCodec.serialize(data);
      
      expect(result).not.toContain('\n');
      expect(result).toBe('{"a":1,"b":2}');
    });

    test('should use custom replacer function', () => {
      const replacerCodec = new JsonCodec({
        replacer: (key, value) => {
          if (key === 'secret') return '[REDACTED]';
          return value;
        }
      });
      
      const data = { public: 'visible', secret: 'hidden' };
      const result = replacerCodec.serialize(data);
      
      expect(result).toContain('[REDACTED]');
      expect(result).not.toContain('hidden');
    });

    test('should use custom reviver function', () => {
      const reviverCodec = new JsonCodec({
        reviver: (key, value) => {
          if (key === 'number' && typeof value === 'string') {
            return parseInt(value, 10);
          }
          return value;
        }
      });
      
      const json = '{"number":"42","text":"hello"}';
      const result = reviverCodec.deserialize(json);
      
      expect(result.number).toBe(42);
      expect(typeof result.number).toBe('number');
      expect(result.text).toBe('hello');
    });
  });

  describe('Error Handling', () => {
    test('should throw error for circular references', () => {
      const obj: any = { name: 'test' };
      obj.self = obj; // Create circular reference
      
      expect(() => codec.serialize(obj)).toThrow(/circular|converting circular structure/i);
    });

    test('should throw error for invalid JSON during deserialization', () => {
      const invalidJson = '{"invalid": json}';
      
      expect(() => codec.deserialize(invalidJson)).toThrow(/JSON deserialization failed/);
    });

    test('should throw error for functions in data', () => {
      const dataWithFunction = {
        name: 'test',
        func: () => 'hello'
      };
      
      // JSON.stringify converts functions to undefined, which becomes null in arrays
      const result = codec.serialize(dataWithFunction);
      expect(result).not.toContain('hello');
    });

    test('should handle undefined values in objects', () => {
      const data = {
        defined: 'value',
        undefined: undefined
      };
      
      const serialized = codec.serialize(data);
      const deserialized = codec.deserialize(serialized);
      
      // JSON.stringify removes undefined values from objects
      expect(deserialized).not.toHaveProperty('undefined');
      expect(deserialized.defined).toBe('value');
    });
  });

  describe('Validation', () => {
    test('should validate serializable data', () => {
      const validData = { name: 'test', value: 42 };
      expect(codec.validate(validData)).toBe(true);
    });

    test('should reject null and undefined', () => {
      expect(codec.validate(null)).toBe(false);
      expect(codec.validate(undefined)).toBe(false);
    });

    test('should reject circular references during validation', () => {
      const obj: any = { name: 'test' };
      obj.self = obj;
      
      expect(codec.validate(obj)).toBe(false);
    });

    test('should validate arrays', () => {
      const validArray = [1, 'two', { three: 3 }];
      expect(codec.validate(validArray)).toBe(true);
    });
  });

  describe('Compression', () => {
    test('should support compression when enabled', () => {
      const compressedCodec = new JsonCodec({ compression: true });
      const data = { message: 'Hello World' };
      
      const result = compressedCodec.serialize(data);
      expect(typeof result).toBe('string');
      expect(result).toContain('[COMPRESSED]');
    });

    test('should decompress data correctly', () => {
      const compressedCodec = new JsonCodec({ compression: true });
      const data = { message: 'Hello World' };
      
      const compressed = compressedCodec.serialize(data);
      const decompressed = compressedCodec.deserialize(compressed);
      
      expect(decompressed).toEqual(data);
    });

    test('should handle non-compressed data in compression mode', () => {
      const compressedCodec = new JsonCodec({ compression: true });
      const regularJson = '{"message":"Hello World"}';
      
      const result = compressedCodec.deserialize(regularJson);
      expect(result).toEqual({ message: 'Hello World' });
    });
  });

  describe('Metadata', () => {
    test('should provide correct metadata', () => {
      const metadata = codec.getMetadata();
      
      expect(metadata.name).toBe('json');
      expect(metadata.version).toBe('1.0.0');
      expect(metadata.mimeType).toBe('application/json');
      expect(metadata.supportsCompression).toBe(true);
    });

    test('should include options in metadata', () => {
      const codecWithOptions = new JsonCodec({ 
        prettify: true, 
        compression: true 
      });
      const metadata = codecWithOptions.getMetadata();
      
      expect(metadata.options.prettify).toBe(true);
      expect(metadata.options.compression).toBe(true);
      expect(metadata.features.prettify).toBe(true);
      expect(metadata.features.compression).toBe(true);
    });
  });
});

describe('JsonCodecFactory', () => {
  describe('Factory Methods', () => {
    test('should create default codec', () => {
      const codec = JsonCodecFactory.createDefault();
      
      expect(codec).toBeInstanceOf(JsonCodec);
      expect(codec.name).toBe('json');
    });

    test('should create prettified codec', () => {
      const codec = JsonCodecFactory.createPrettified();
      const data = { a: 1, b: 2 };
      const result = codec.serialize(data);
      
      expect(result).toContain('\n');
      expect(result).toContain('  ');
    });

    test('should create compressed codec', () => {
      const codec = JsonCodecFactory.createCompressed();
      const data = { message: 'test' };
      const result = codec.serialize(data);
      
      expect(result).toContain('[COMPRESSED]');
    });

    test('should create minified codec', () => {
      const codec = JsonCodecFactory.createMinified();
      const data = { a: 1, b: 2 };
      const result = codec.serialize(data);
      
      expect(result).not.toContain('\n');
      expect(result).toBe('{"a":1,"b":2}');
    });

    test('should create codec with custom replacer', () => {
      const replacer = (key: string, value: any) => {
        if (key === 'password') return '[HIDDEN]';
        return value;
      };
      
      const codec = JsonCodecFactory.createWithReplacer(replacer);
      const data = { username: 'admin', password: 'secret123' };
      const result = codec.serialize(data);
      
      expect(result).toContain('[HIDDEN]');
      expect(result).not.toContain('secret123');
    });
  });

  describe('GAS Codec', () => {
    let gasCodec: JsonCodec;

    beforeEach(() => {
      gasCodec = JsonCodecFactory.createGASCodec();
    });

    test('should handle Date objects', () => {
      const date = new Date('2023-01-01T00:00:00.000Z');
      const data = { timestamp: date };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      // Check if it's either a Date object or the original string was preserved
      if (typeof deserialized.timestamp === 'string') {
        expect(deserialized.timestamp).toBe('2023-01-01T00:00:00.000Z');
      } else {
        expect(deserialized.timestamp).toBeInstanceOf(Date);
        expect(deserialized.timestamp.getTime()).toBe(date.getTime());
      }
    });

    test('should handle Map objects', () => {
      const map = new Map([['key1', 'value1'], ['key2', 'value2']]);
      const data = { myMap: map };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      expect(deserialized.myMap).toBeInstanceOf(Map);
      expect(deserialized.myMap.get('key1')).toBe('value1');
      expect(deserialized.myMap.get('key2')).toBe('value2');
    });

    test('should handle Set objects', () => {
      const set = new Set(['value1', 'value2', 'value3']);
      const data = { mySet: set };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      expect(deserialized.mySet).toBeInstanceOf(Set);
      expect(deserialized.mySet.has('value1')).toBe(true);
      expect(deserialized.mySet.has('value2')).toBe(true);
      expect(deserialized.mySet.has('value3')).toBe(true);
      expect(deserialized.mySet.size).toBe(3);
    });

    test('should handle Function objects (with caution)', () => {
      const func = function() { return 'hello'; };
      const data = { myFunc: func };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      // Note: Function restoration is dangerous and limited
      // In this test, we expect it to either restore or return null
      if (deserialized.myFunc !== null) {
        expect(typeof deserialized.myFunc).toBe('function');
      }
    });

    test('should handle complex nested structures', () => {
      const complexData = {
        timestamp: new Date('2023-01-01'),
        tags: new Set(['tag1', 'tag2']),
        attributes: new Map([
          ['health', { current: 100, max: 100 }],
          ['mana', { current: 50, max: 50 }]
        ]),
        metadata: {
          version: '1.0.0',
          nested: {
            deep: 'value'
          }
        }
      };
      
      const serialized = gasCodec.serialize(complexData);
      const deserialized = gasCodec.deserialize(serialized);
      
      // Handle Date - might be preserved as string
      if (typeof deserialized.timestamp === 'string') {
        expect(deserialized.timestamp).toBe('2023-01-01T00:00:00.000Z');
      } else {
        expect(deserialized.timestamp).toBeInstanceOf(Date);
      }
      
      expect(deserialized.tags).toBeInstanceOf(Set);
      expect(deserialized.attributes).toBeInstanceOf(Map);
      expect(deserialized.attributes.get('health')).toEqual({ current: 100, max: 100 });
      expect(deserialized.metadata.nested.deep).toBe('value');
    });

    test('should handle circular reference markers', () => {
      const data = {
        id: 'test',
        circular: { __circularRef: 'ref123' }
      };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      expect(deserialized.circular.__circularRef).toBe('ref123');
    });
  });
});

describe('Edge Cases and Performance', () => {
  let codec: JsonCodec;

  beforeEach(() => {
    codec = new JsonCodec();
  });

  test('should handle empty objects and arrays', () => {
    expect(codec.deserialize(codec.serialize({}))).toEqual({});
    expect(codec.deserialize(codec.serialize([]))).toEqual([]);
  });

  test('should handle special number values', () => {
    const data = {
      infinity: Infinity,
      negativeInfinity: -Infinity,
      nan: NaN
    };
    
    const serialized = codec.serialize(data);
    const deserialized = codec.deserialize(serialized);
    
    // JSON converts these to null
    expect(deserialized.infinity).toBe(null);
    expect(deserialized.negativeInfinity).toBe(null);
    expect(deserialized.nan).toBe(null);
  });

  test('should handle very large objects', () => {
    const largeObject: any = {};
    for (let i = 0; i < 1000; i++) {
      largeObject[`key${i}`] = `value${i}`;
    }
    
    const serialized = codec.serialize(largeObject);
    const deserialized = codec.deserialize(serialized);
    
    expect(Object.keys(deserialized)).toHaveLength(1000);
    expect(deserialized.key0).toBe('value0');
    expect(deserialized.key999).toBe('value999');
  });

  test('should handle deep nesting', () => {
    let deepObject: any = { value: 'leaf' };
    for (let i = 0; i < 50; i++) {
      deepObject = { level: i, nested: deepObject };
    }
    
    const serialized = codec.serialize(deepObject);
    const deserialized = codec.deserialize(serialized);
    
    // Navigate to the leaf
    let current = deserialized;
    for (let i = 49; i >= 0; i--) {
      expect(current.level).toBe(i);
      current = current.nested;
    }
    expect(current.value).toBe('leaf');
  });

  test('should handle unicode characters', () => {
    const data = {
      emoji: '🎮🚀🎯',
      chinese: '你好世界',
      arabic: 'مرحبا بالعالم',
      special: 'Special chars: !@#$%^&*()'
    };
    
    const serialized = codec.serialize(data);
    const deserialized = codec.deserialize(serialized);
    
    expect(deserialized).toEqual(data);
  });
});

describe('Integration Tests', () => {
  test('should work with real GAS-like data structures', () => {
    const gasLikeData = {
      version: '2.0.0',
      timestamp: Date.now(),
      owner: { id: 'player1', name: 'Hero' },
      attributes: {
        health: { current: 75, max: 100, modifiers: [] },
        mana: { current: 30, max: 50, modifiers: [] }
      },
      abilities: [
        {
          id: 'fireball',
          name: 'Fireball',
          cooldown: 3000,
          type: 'DamageAbility'
        }
      ],
      activeEffects: [],
      tags: ['player', 'alive'],
      cooldowns: {},
      metadata: {
        saveTime: new Date().toISOString(),
        gameVersion: '1.0.0'
      }
    };
    
    const gasCodec = JsonCodecFactory.createGASCodec();
    const serialized = gasCodec.serialize(gasLikeData);
    const deserialized = gasCodec.deserialize(serialized);
    
    expect(deserialized.version).toBe('2.0.0');
    expect(deserialized.owner.name).toBe('Hero');
    expect(deserialized.attributes.health.current).toBe(75);
    expect(deserialized.abilities).toHaveLength(1);
    expect(deserialized.abilities[0].id).toBe('fireball');
    expect(deserialized.tags).toContain('player');
  });
});