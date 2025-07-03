import { AttributeModifier, GameplayAttributeData } from '../types/AbilityTypes';

export class GameplayAttribute {
  private data: GameplayAttributeData;
  private cachedFinalValue: number | null = null;
  private isDirty: boolean = true;

  constructor(name: string, baseValue: number, maxValue?: number) {
    this.data = {
      name,
      baseValue,
      currentValue: baseValue,
      maxValue,
      modifiers: []
    };
  }

  get name(): string {
    return this.data.name;
  }

  get baseValue(): number {
    return this.data.baseValue;
  }

  set baseValue(value: number) {
    this.data.baseValue = value;
    this.markDirty();
  }

  get currentValue(): number {
    return this.data.currentValue;
  }

  set currentValue(value: number) {
    const clampedValue = this.clampValue(value);
    if (this.data.currentValue !== clampedValue) {
      this.data.currentValue = clampedValue;
      this.markDirty();
    }
  }

  get maxValue(): number | undefined {
    return this.data.maxValue;
  }

  set maxValue(value: number | undefined) {
    this.data.maxValue = value;
    // Re-clamp current value if max changed
    this.currentValue = this.data.currentValue;
  }

  get finalValue(): number {
    if (this.isDirty || this.cachedFinalValue === null) {
      this.cachedFinalValue = this.calculateFinalValue();
      this.isDirty = false;
    }
    return this.cachedFinalValue;
  }

  get modifiers(): readonly AttributeModifier[] {
    return this.data.modifiers;
  }

  // Add a modifier to this attribute
  addModifier(modifier: AttributeModifier): void {
    // Check if modifier with same ID already exists
    const existingIndex = this.data.modifiers.findIndex(m => m.id === modifier.id);
    
    if (existingIndex >= 0) {
      // Replace existing modifier
      this.data.modifiers[existingIndex] = modifier;
    } else {
      // Add new modifier
      this.data.modifiers.push(modifier);
    }
    
    this.markDirty();
  }

  // Remove a modifier by ID
  removeModifier(modifierId: string): boolean {
    const initialLength = this.data.modifiers.length;
    this.data.modifiers = this.data.modifiers.filter(m => m.id !== modifierId);
    
    if (this.data.modifiers.length !== initialLength) {
      this.markDirty();
      return true;
    }
    
    return false;
  }

  // Remove all modifiers from a specific source
  removeModifiersFromSource(source: string): number {
    const initialLength = this.data.modifiers.length;
    this.data.modifiers = this.data.modifiers.filter(m => m.source !== source);
    
    const removedCount = initialLength - this.data.modifiers.length;
    if (removedCount > 0) {
      this.markDirty();
    }
    
    return removedCount;
  }

  // Check if attribute has a modifier from specific source
  hasModifierFromSource(source: string): boolean {
    return this.data.modifiers.some(m => m.source === source);
  }

  // Calculate the final value considering all modifiers
  private calculateFinalValue(): number {
    let finalValue = this.data.baseValue;

    // Apply additive modifiers first
    const additiveModifiers = this.data.modifiers.filter(m => m.operation === 'add');
    for (const modifier of additiveModifiers) {
      finalValue += modifier.magnitude;
    }

    // Apply multiplicative modifiers
    const multiplicativeModifiers = this.data.modifiers.filter(m => m.operation === 'multiply');
    for (const modifier of multiplicativeModifiers) {
      finalValue *= modifier.magnitude;
    }

    // Apply override modifiers (last one wins)
    const overrideModifiers = this.data.modifiers.filter(m => m.operation === 'override');
    if (overrideModifiers.length > 0) {
      finalValue = overrideModifiers[overrideModifiers.length - 1].magnitude;
    }

    return this.clampValue(finalValue);
  }

  // Clamp value within bounds if maxValue is set
  private clampValue(value: number): number {
    if (this.data.maxValue !== undefined) {
      return Math.min(Math.max(0, value), this.data.maxValue);
    }
    return Math.max(0, value); // At least ensure non-negative
  }

  private markDirty(): void {
    this.isDirty = true;
  }

  // Get a snapshot of the current attribute state
  getSnapshot(): GameplayAttributeData {
    return {
      name: this.data.name,
      baseValue: this.data.baseValue,
      currentValue: this.data.currentValue,
      maxValue: this.data.maxValue,
      modifiers: [...this.data.modifiers] // Shallow copy
    };
  }

  // Restore from snapshot
  restoreFromSnapshot(snapshot: GameplayAttributeData): void {
    this.data = {
      ...snapshot,
      modifiers: [...snapshot.modifiers] // Shallow copy
    };
    this.markDirty();
  }

  // Debug helper
  toString(): string {
    return `${this.name}: ${this.currentValue}/${this.maxValue || '∞'} (base: ${this.baseValue}, final: ${this.finalValue}, modifiers: ${this.data.modifiers.length})`;
  }
}