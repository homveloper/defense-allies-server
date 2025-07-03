// Simple Tests for JSON Codec without complex dependencies
// This tests the core serialization functionality without GAS dependencies

import { JsonCodec, JsonCodecFactory } from '../JsonCodec';
import { 
  SerializationCodecRegistry, 
  getSerializationRegistry 
} from '../SerializationCodec';

describe('Simple Serialization Tests', () => {
  describe('JsonCodec Basic Functionality', () => {
    let codec: JsonCodec;

    beforeEach(() => {
      codec = new JsonCodec();
    });

    test('should serialize and deserialize simple data', () => {
      const data = { name: 'test', value: 42, active: true };
      const serialized = codec.serialize(data);
      const deserialized = codec.deserialize(serialized);
      
      expect(deserialized).toEqual(data);
    });

    test('should handle arrays', () => {
      const data = [1, 'two', { three: 3 }, [4, 5]];
      const serialized = codec.serialize(data);
      const deserialized = codec.deserialize(serialized);
      
      expect(deserialized).toEqual(data);
    });

    test('should validate data correctly', () => {
      expect(codec.validate({ test: 'data' })).toBe(true);
      expect(codec.validate(null)).toBe(false);
      expect(codec.validate(undefined)).toBe(false);
    });

    test('should provide correct metadata', () => {
      const metadata = codec.getMetadata();
      expect(metadata.name).toBe('json');
      expect(metadata.version).toBe('1.0.0');
      expect(metadata.mimeType).toBe('application/json');
    });
  });

  describe('JsonCodecFactory', () => {
    test('should create different codec configurations', () => {
      const defaultCodec = JsonCodecFactory.createDefault();
      const prettifiedCodec = JsonCodecFactory.createPrettified();
      const compressedCodec = JsonCodecFactory.createCompressed();
      
      expect(defaultCodec.name).toBe('json');
      expect(prettifiedCodec.name).toBe('json');
      expect(compressedCodec.name).toBe('json');
      
      // Test prettified output
      const data = { a: 1, b: 2 };
      const prettified = prettifiedCodec.serialize(data);
      expect(prettified).toContain('\n'); // Should have newlines
      
      const minified = defaultCodec.serialize(data);
      expect(minified).not.toContain('\n'); // Should not have newlines
    });

    test('should handle custom replacer/reviver', () => {
      const codec = JsonCodecFactory.createWithReplacer(
        (key, value) => key === 'secret' ? '[HIDDEN]' : value,
        (key, value) => key === 'number' && typeof value === 'string' ? parseInt(value) : value
      );
      
      const data = { public: 'visible', secret: 'hidden', number: '42' };
      const serialized = codec.serialize(data);
      
      expect(serialized).toContain('[HIDDEN]');
      expect(serialized).not.toContain('hidden');
      
      const withNumberString = '{"public":"visible","secret":"[HIDDEN]","number":"42"}';
      const deserialized = codec.deserialize(withNumberString);
      
      expect(deserialized.number).toBe(42);
      expect(typeof deserialized.number).toBe('number');
    });
  });

  describe('Registry Functionality', () => {
    let registry: SerializationCodecRegistry;

    beforeEach(() => {
      registry = new SerializationCodecRegistry();
    });

    test('should register and use codecs', () => {
      const codec = JsonCodecFactory.createDefault();
      registry.register(codec);
      registry.setDefault('json');
      
      const data = { test: 'registry' };
      const serialized = registry.serialize(data);
      
      expect(serialized.codec).toBe('json');
      expect(serialized.data).toContain('registry');
      
      const deserialized = registry.deserialize(serialized);
      expect(deserialized).toEqual(data);
    });

    test('should handle singleton registry', () => {
      const registry1 = getSerializationRegistry();
      const registry2 = getSerializationRegistry();
      
      expect(registry1).toBe(registry2);
    });
  });

  describe('Error Handling', () => {
    let codec: JsonCodec;

    beforeEach(() => {
      codec = new JsonCodec();
    });

    test('should handle circular references', () => {
      const obj: any = { name: 'test' };
      obj.self = obj;
      
      expect(() => codec.serialize(obj)).toThrow();
    });

    test('should handle invalid JSON', () => {
      expect(() => codec.deserialize('invalid json{')).toThrow();
    });

    test('should handle special values', () => {
      const data = {
        infinity: Infinity,
        nan: NaN,
        undef: undefined
      };
      
      const serialized = codec.serialize(data);
      const deserialized = codec.deserialize(serialized);
      
      // JSON converts these to null or removes them
      expect(deserialized.infinity).toBe(null);
      expect(deserialized.nan).toBe(null);
      expect(deserialized).not.toHaveProperty('undef');
    });
  });

  describe('Complex Data Types', () => {
    let gasCodec: JsonCodec;

    beforeEach(() => {
      gasCodec = JsonCodecFactory.createGASCodec();
    });

    test('should handle Map objects', () => {
      const map = new Map([['key1', 'value1'], ['key2', 'value2']]);
      const data = { myMap: map };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      expect(deserialized.myMap).toBeInstanceOf(Map);
      expect(deserialized.myMap.get('key1')).toBe('value1');
      expect(deserialized.myMap.size).toBe(2);
    });

    test('should handle Set objects', () => {
      const set = new Set(['a', 'b', 'c']);
      const data = { mySet: set };
      
      const serialized = gasCodec.serialize(data);
      const deserialized = gasCodec.deserialize(serialized);
      
      expect(deserialized.mySet).toBeInstanceOf(Set);
      expect(deserialized.mySet.has('a')).toBe(true);
      expect(deserialized.mySet.size).toBe(3);
    });

    test('should handle Date objects', () => {
      const date = new Date('2023-01-01T00:00:00.000Z');
      const data = { timestamp: date };
      
      const serialized = gasCodec.serialize(data);
      console.log('Serialized:', serialized);
      const deserialized = gasCodec.deserialize(serialized);
      console.log('Deserialized:', deserialized);
      
      // GAS codec should either restore Date objects or preserve ISO string
      if (deserialized.timestamp instanceof Date) {
        expect(deserialized.timestamp.getTime()).toBe(date.getTime());
      } else {
        // If it's a string, it should be the ISO string
        expect(deserialized.timestamp).toBe(date.toISOString());
      }
    });
  });
});