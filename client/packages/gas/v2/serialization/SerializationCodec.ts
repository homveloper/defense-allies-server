// Serialization Codec Interface for GAS v2
// Supports multiple serialization formats (JSON, BSON, MessagePack, etc.)

export interface SerializationCodec {
  readonly name: string;
  readonly version: string;
  readonly mimeType: string;
  
  serialize<T>(data: T): string | Uint8Array | Buffer;
  deserialize<T>(serialized: string | Uint8Array | Buffer): T;
  
  // Optional validation
  validate?(data: any): boolean;
  
  // Optional compression
  supportsCompression?: boolean;
  
  // Optional metadata
  getMetadata?(): Record<string, any>;
}

export interface SerializationOptions {
  codec?: string; // Codec name to use
  compression?: boolean; // Enable compression if supported
  validation?: boolean; // Enable validation if supported
  metadata?: Record<string, any>; // Additional metadata
}

export interface SerializedData {
  codec: string;
  version: string;
  timestamp: number;
  compressed: boolean;
  metadata?: Record<string, any>;
  data: string | Uint8Array | Buffer;
}

export abstract class BaseSerializationCodec implements SerializationCodec {
  abstract readonly name: string;
  abstract readonly version: string;
  abstract readonly mimeType: string;
  
  abstract serialize<T>(data: T): string | Uint8Array | Buffer;
  abstract deserialize<T>(serialized: string | Uint8Array | Buffer): T;
  
  validate?(data: any): boolean {
    // Default validation - check if data is not null/undefined
    return data !== null && data !== undefined;
  }
  
  supportsCompression?: boolean = false;
  
  getMetadata?(): Record<string, any> {
    return {
      name: this.name,
      version: this.version,
      mimeType: this.mimeType,
      supportsCompression: this.supportsCompression
    };
  }
}

// Registry for managing multiple codecs
export class SerializationCodecRegistry {
  private static instance: SerializationCodecRegistry;
  private codecs: Map<string, SerializationCodec> = new Map();
  private defaultCodec: string = 'json';
  
  static getInstance(): SerializationCodecRegistry {
    if (!SerializationCodecRegistry.instance) {
      SerializationCodecRegistry.instance = new SerializationCodecRegistry();
    }
    return SerializationCodecRegistry.instance;
  }
  
  register(codec: SerializationCodec): void {
    this.codecs.set(codec.name.toLowerCase(), codec);
  }
  
  unregister(codecName: string): boolean {
    return this.codecs.delete(codecName.toLowerCase());
  }
  
  get(codecName: string): SerializationCodec | undefined {
    return this.codecs.get(codecName.toLowerCase());
  }
  
  getDefault(): SerializationCodec | undefined {
    return this.codecs.get(this.defaultCodec);
  }
  
  setDefault(codecName: string): boolean {
    if (this.codecs.has(codecName.toLowerCase())) {
      this.defaultCodec = codecName.toLowerCase();
      return true;
    }
    return false;
  }
  
  list(): string[] {
    return Array.from(this.codecs.keys());
  }
  
  clear(): void {
    this.codecs.clear();
  }
  
  serialize<T>(
    data: T, 
    options: SerializationOptions = {}
  ): SerializedData {
    const codecName = options.codec || this.defaultCodec;
    const codec = this.get(codecName);
    
    if (!codec) {
      throw new Error(`Serialization codec '${codecName}' not found`);
    }
    
    // Validate if requested
    if (options.validation && codec.validate && !codec.validate(data)) {
      throw new Error(`Data validation failed for codec '${codecName}'`);
    }
    
    const serializedData = codec.serialize(data);
    
    return {
      codec: codec.name,
      version: codec.version,
      timestamp: Date.now(),
      compressed: options.compression && codec.supportsCompression || false,
      metadata: {
        ...codec.getMetadata?.(),
        ...options.metadata
      },
      data: serializedData
    };
  }
  
  deserialize<T>(serializedData: SerializedData): T {
    const codec = this.get(serializedData.codec);
    
    if (!codec) {
      throw new Error(`Serialization codec '${serializedData.codec}' not found`);
    }
    
    // Version compatibility check
    if (codec.version !== serializedData.version) {
      console.warn(`Codec version mismatch: expected ${codec.version}, got ${serializedData.version}`);
    }
    
    return codec.deserialize<T>(serializedData.data);
  }
}

// Helper function to get the registry instance
export function getSerializationRegistry(): SerializationCodecRegistry {
  return SerializationCodecRegistry.getInstance();
}