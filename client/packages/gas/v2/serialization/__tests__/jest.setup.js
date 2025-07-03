// Jest setup file for GAS serialization tests

// Mock console methods to reduce noise during tests
global.console = {
  ...console,
  // Suppress console.warn for expected warnings (like version mismatches)
  warn: jest.fn(),
  // Keep error and log for debugging
  error: console.error,
  log: console.log,
};

// Global test utilities
global.createMockAbility = function(id, name, cooldown = 1000) {
  return {
    id,
    name,
    description: `Mock ${name}`,
    cooldown,
    canActivate: jest.fn(() => true),
    activate: jest.fn(() => Promise.resolve(true)),
    getCooldownRemaining: jest.fn(() => 0)
  };
};

global.createMockGameplayEffect = function(id, name, duration = 5000) {
  return {
    spec: {
      id,
      name,
      duration,
      stackingPolicy: 'none',
      attributeModifiers: [],
      grantedTags: [],
      removedTags: []
    },
    createActiveInstance: jest.fn(() => ({
      spec: this.spec,
      startTime: Date.now(),
      stacks: 1,
      appliedModifiers: []
    }))
  };
};

// Clean up after each test
afterEach(() => {
  // Clear all mocks
  jest.clearAllMocks();
  
  // Reset console mock call history
  console.warn.mockClear();
});

// Global error handler for unhandled promise rejections
process.on('unhandledRejection', (reason, promise) => {
  console.error('Unhandled Rejection at:', promise, 'reason:', reason);
});

// Extend Jest matchers if needed
expect.extend({
  toBeSerializable(received) {
    try {
      JSON.stringify(received);
      JSON.parse(JSON.stringify(received));
      return {
        message: () => `Expected ${received} not to be serializable`,
        pass: true,
      };
    } catch (error) {
      return {
        message: () => `Expected ${received} to be serializable, but got error: ${error.message}`,
        pass: false,
      };
    }
  },
  
  toBeValidSerializedData(received) {
    const requiredFields = ['codec', 'version', 'timestamp', 'compressed', 'data'];
    const hasAllFields = requiredFields.every(field => field in received);
    
    if (!hasAllFields) {
      return {
        message: () => `Expected object to have all required fields: ${requiredFields.join(', ')}`,
        pass: false,
      };
    }
    
    return {
      message: () => `Expected object not to be valid serialized data`,
      pass: true,
    };
  }
});