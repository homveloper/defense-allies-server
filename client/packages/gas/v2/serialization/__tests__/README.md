# GAS v2 Serialization Tests

Comprehensive unit test suite for the GAS v2 serialization system.

## 🧪 Test Coverage

### JsonCodec Tests (`JsonCodec.test.ts`)
- ✅ Basic serialization/deserialization
- ✅ Configuration options (prettify, compression, custom replacer/reviver)
- ✅ Error handling (circular references, invalid JSON)
- ✅ Validation functionality
- ✅ Compression support
- ✅ Metadata handling
- ✅ Factory methods
- ✅ GAS-specific codec features (Date, Map, Set, Functions)
- ✅ Edge cases and performance scenarios
- ✅ Unicode character support

### SerializationCodec Tests (`SerializationCodec.test.ts`)
- ✅ Base codec functionality
- ✅ Registry management (register, unregister, list, clear)
- ✅ Default codec handling
- ✅ Serialization operations with multiple codecs
- ✅ Deserialization with version checking
- ✅ Singleton registry pattern
- ✅ Integration scenarios

### GASSerializer Tests (`GASSerializer.test.ts`)
- ✅ Singleton pattern
- ✅ Individual component serialization (attributes, effects, abilities)
- ✅ Complete ASC serialization with different contexts
- ✅ Save state vs snapshot vs configuration export
- ✅ Ability registry and deserialization
- ✅ Error handling and validation
- ✅ Complex integration scenarios
- ✅ Data integrity across multiple cycles

## 🚀 Running Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch

# Run with verbose output
npm run test:verbose

# Debug tests
npm run test:debug
```

## 📊 Coverage Targets

- **Branches**: 80%
- **Functions**: 80%
- **Lines**: 80%
- **Statements**: 80%

## 🔧 Test Configuration

### Jest Configuration
- **Environment**: Node.js
- **Preset**: ts-jest
- **Timeout**: 10 seconds
- **Setup**: Custom matchers and utilities

### Custom Matchers
- `toBeSerializable()` - Checks if object can be JSON serialized
- `toBeValidSerializedData()` - Validates serialized data structure

### Test Utilities
- `createMockAbility()` - Creates mock ability for testing
- `createMockGameplayEffect()` - Creates mock effect for testing

## 📁 Test Structure

```
__tests__/
├── JsonCodec.test.ts           # JSON codec unit tests
├── SerializationCodec.test.ts  # Core serialization infrastructure tests
├── GASSerializer.test.ts       # GAS-specific serialization tests
├── jest.config.js             # Jest configuration
├── jest.setup.js              # Test setup and utilities
├── package.json               # Test dependencies and scripts
└── README.md                  # This file
```

## 🎯 Test Scenarios

### Basic Functionality
- Serialize/deserialize simple and complex objects
- Handle null, undefined, arrays, nested objects
- Configuration options (prettify, compression, custom functions)

### Error Handling
- Circular reference detection
- Invalid JSON handling
- Missing codec errors
- Version mismatch warnings

### GAS-Specific Features
- Ability serialization with custom data
- Attribute serialization with modifiers
- Effect serialization with complex configurations
- Complete ASC state management

### Integration Scenarios
- Multiple serialize-deserialize cycles
- Cross-codec format conversion
- Complex game state preservation
- Performance with large datasets

## 📝 Writing New Tests

When adding new tests:

1. **Follow naming conventions**: `describe('FeatureName')` and `test('should do something')`
2. **Use beforeEach/afterEach**: Clean setup and teardown
3. **Test edge cases**: null, undefined, empty objects, large datasets
4. **Mock external dependencies**: Use Jest mocks for complex dependencies
5. **Assert thoroughly**: Check both positive and negative cases
6. **Document complex scenarios**: Add comments for non-obvious test logic

### Example Test Structure

```typescript
describe('NewFeature', () => {
  let testSubject: NewFeature;

  beforeEach(() => {
    testSubject = new NewFeature();
  });

  describe('Basic Functionality', () => {
    test('should handle normal case', () => {
      const result = testSubject.doSomething('input');
      expect(result).toBe('expected');
    });

    test('should handle edge case', () => {
      expect(() => testSubject.doSomething(null)).toThrow();
    });
  });

  describe('Error Handling', () => {
    test('should throw meaningful error for invalid input', () => {
      expect(() => testSubject.doSomething('invalid'))
        .toThrow('Expected error message');
    });
  });
});
```

## 🐛 Debugging Tests

### Common Issues

1. **Circular References**: Check for objects referencing themselves
2. **Async Operations**: Ensure proper await/return in async tests
3. **Mock Interference**: Clear mocks between tests
4. **Type Issues**: Verify TypeScript types in test assertions

### Debug Commands

```bash
# Run specific test file
npm test JsonCodec.test.ts

# Run specific test case
npm test -- --testNamePattern="should serialize"

# Debug with Chrome DevTools
npm run test:debug
```

## 🔄 Continuous Integration

Tests are designed to run in CI environments with:
- Deterministic results (no random values)
- Proper cleanup (no side effects between tests)
- Clear error messages
- Reasonable timeouts

## 📚 Related Documentation

- [GAS v2 Serialization System](../README.md)
- [JSON Codec Documentation](../JsonCodec.ts)
- [Serialization Registry](../SerializationCodec.ts)
- [GAS Serializer](../GASSerializer.ts)