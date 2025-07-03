// JSON Serialization Codec for GAS v2
// High-performance JSON codec with optional compression and validation

import { BaseSerializationCodec } from './SerializationCodec';

export interface JsonCodecOptions {
  prettify?: boolean; // Pretty print JSON
  space?: number; // Number of spaces for indentation
  replacer?: (key: string, value: any) => any; // JSON.stringify replacer
  reviver?: (key: string, value: any) => any; // JSON.parse reviver
  compression?: boolean; // Enable LZ compression
}

export class JsonCodec extends BaseSerializationCodec {
  readonly name = 'json';
  readonly version = '1.0.0';
  readonly mimeType = 'application/json';
  readonly supportsCompression = true;
  
  private options: JsonCodecOptions;
  
  constructor(options: JsonCodecOptions = {}) {
    super();
    this.options = {
      prettify: false,
      space: 2,
      compression: false,
      ...options
    };
  }
  
  serialize<T>(data: T): string {
    try {
      const jsonString = JSON.stringify(
        data, 
        this.options.replacer, 
        this.options.prettify ? this.options.space : undefined
      );
      
      if (this.options.compression) {
        return this.compress(jsonString);
      }
      
      return jsonString;
    } catch (error) {
      throw new Error(`JSON serialization failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  deserialize<T>(serialized: string): T {
    try {
      let jsonString = serialized;
      
      // Check if data is compressed
      if (this.isCompressed(serialized)) {
        jsonString = this.decompress(serialized);
      }
      
      return JSON.parse(jsonString, this.options.reviver);
    } catch (error) {
      throw new Error(`JSON deserialization failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }
  
  validate(data: any): boolean {
    if (!super.validate?.(data)) {
      return false;
    }
    
    try {
      // Try to serialize and deserialize to validate JSON compatibility
      const serialized = JSON.stringify(data);
      JSON.parse(serialized);
      return true;
    } catch {
      return false;
    }
  }
  
  // Simple LZ-style compression (placeholder - in real app you'd use a proper library)
  private compress(data: string): string {
    if (!this.options.compression) return data;
    
    // Simple run-length encoding for demo
    // In production, use libraries like lz-string, pako, etc.
    return `[COMPRESSED]${data}`;
  }
  
  private decompress(data: string): string {
    if (!this.isCompressed(data)) return data;
    
    // Remove compression marker
    return data.replace('[COMPRESSED]', '');
  }
  
  private isCompressed(data: string): boolean {
    return data.startsWith('[COMPRESSED]');
  }
  
  getMetadata(): Record<string, any> {
    return {
      ...super.getMetadata?.(),
      options: this.options,
      features: {
        prettify: this.options.prettify,
        compression: this.options.compression,
        customReplacer: !!this.options.replacer,
        customReviver: !!this.options.reviver
      }
    };
  }
}

// Factory function for creating JSON codec with common configurations
export class JsonCodecFactory {
  static createDefault(): JsonCodec {
    return new JsonCodec();
  }
  
  static createPrettified(): JsonCodec {
    return new JsonCodec({ 
      prettify: true, 
      space: 2 
    });
  }
  
  static createCompressed(): JsonCodec {
    return new JsonCodec({ 
      compression: true 
    });
  }
  
  static createMinified(): JsonCodec {
    return new JsonCodec({ 
      prettify: false 
    });
  }
  
  static createWithReplacer(
    replacer: (key: string, value: any) => any,
    reviver?: (key: string, value: any) => any
  ): JsonCodec {
    return new JsonCodec({ 
      replacer, 
      reviver 
    });
  }
  
  // Special codec for GAS data - handles functions, dates, etc.
  static createGASCodec(): JsonCodec {
    return new JsonCodec({
      replacer: (_key: string, value: any) => {
        // Handle special types commonly found in GAS
        if (value instanceof Date) {
          return { __type: 'Date', value: value.toISOString() };
        }
        
        if (typeof value === 'function') {
          return { __type: 'Function', value: value.toString() };
        }
        
        if (value instanceof Map) {
          return { 
            __type: 'Map', 
            value: Array.from(value.entries()) 
          };
        }
        
        if (value instanceof Set) {
          return { 
            __type: 'Set', 
            value: Array.from(value) 
          };
        }
        
        // Handle circular references
        if (typeof value === 'object' && value !== null) {
          if (value.__circularRef) {
            return { __type: 'CircularRef', id: value.__circularRef };
          }
        }
        
        return value;
      },
      
      reviver: (_key: string, value: any) => {
        if (typeof value === 'object' && value !== null && value.__type) {
          switch (value.__type) {
            case 'Date':
              return new Date(value.value);
              
            case 'Function':
              // Note: eval is dangerous in production, use with caution
              try {
                return new Function('return ' + value.value)();
              } catch {
                return null; // Return null if function can't be restored
              }
              
            case 'Map':
              return new Map(value.value);
              
            case 'Set':
              return new Set(value.value);
              
            case 'CircularRef':
              // Handle circular references (would need additional context)
              return { __circularRef: value.id };
              
            default:
              return value;
          }
        }
        
        return value;
      }
    });
  }
}