import { 
  AbilityCondition, 
  AbilityContext, 
  ConditionResult,
  AttributeConditionConfig,
  TagConditionConfig,
  CooldownConditionConfig,
  TimeConditionConfig,
  ComboConditionConfig,
  ResourceConditionConfig
} from '../types/AbilityTypes';

/**
 * Base condition class that all specific conditions extend
 */
export abstract class BaseCondition implements AbilityCondition {
  public readonly id: string;
  public readonly name: string;
  public readonly description?: string;

  constructor(id: string, name: string, description?: string) {
    this.id = id;
    this.name = name;
    this.description = description;
  }

  abstract check(context: AbilityContext): boolean | Promise<boolean>;

  onFailure?(context: AbilityContext, reason: string): void {
    console.warn(`Condition '${this.name}' failed: ${reason}`);
  }
}

/**
 * Attribute-based condition (health > 50, mana >= 30, etc.)
 */
export class AttributeCondition extends BaseCondition {
  private config: AttributeConditionConfig;

  constructor(config: AttributeConditionConfig) {
    super(
      `attr_${config.attribute}_${config.operator}_${config.value}`,
      `${config.attribute} ${config.operator} ${config.value}${config.percentage ? '%' : ''}`,
      `Requires ${config.attribute} to be ${config.operator} ${config.value}${config.percentage ? '% of maximum' : ''}`
    );
    this.config = config;
  }

  check(context: AbilityContext): boolean {
    const owner = context.owner;
    if (!owner?.abilitySystem) return false;

    const attribute = owner.abilitySystem.getAttribute(this.config.attribute);
    if (!attribute) return false;

    let checkValue = this.config.value;
    let currentValue = attribute.currentValue;

    // Convert to percentage if needed
    if (this.config.percentage && attribute.maxValue) {
      checkValue = (this.config.value / 100) * attribute.maxValue;
    }

    switch (this.config.operator) {
      case '>': return currentValue > checkValue;
      case '<': return currentValue < checkValue;
      case '>=': return currentValue >= checkValue;
      case '<=': return currentValue <= checkValue;
      case '===': return currentValue === checkValue;
      case '!==': return currentValue !== checkValue;
      default: return false;
    }
  }
}

/**
 * Tag-based condition (has specific tags or doesn't have tags)
 */
export class TagCondition extends BaseCondition {
  private config: TagConditionConfig;

  constructor(config: TagConditionConfig) {
    const modeDesc = config.mode === 'all' ? 'all of' : 
                     config.mode === 'any' ? 'any of' : 'none of';
    super(
      `tag_${config.mode}_${config.tags.join('_')}`,
      `Has ${modeDesc}: ${config.tags.join(', ')}`,
      `Requires having ${modeDesc} the tags: ${config.tags.join(', ')}`
    );
    this.config = config;
  }

  check(context: AbilityContext): boolean {
    const owner = context.owner;
    if (!owner?.abilitySystem) return false;

    const hasTags = this.config.tags.map(tag => owner.abilitySystem.hasTag(tag));

    switch (this.config.mode) {
      case 'all': return hasTags.every(has => has);
      case 'any': return hasTags.some(has => has);
      case 'none': return !hasTags.some(has => has);
      default: return false;
    }
  }
}

/**
 * Cooldown-based condition (ability ready or on cooldown)
 */
export class CooldownCondition extends BaseCondition {
  private config: CooldownConditionConfig;

  constructor(config: CooldownConditionConfig) {
    super(
      `cooldown_${config.abilityId}_${config.state}`,
      `${config.abilityId} is ${config.state}`,
      `Requires ability '${config.abilityId}' to be ${config.state}`
    );
    this.config = config;
  }

  check(context: AbilityContext): boolean {
    const owner = context.owner;
    if (!owner?.abilitySystem) return false;

    const cooldownRemaining = owner.abilitySystem.getCooldownRemaining(this.config.abilityId);
    const isOnCooldown = cooldownRemaining > 0;

    return this.config.state === 'on-cooldown' ? isOnCooldown : !isOnCooldown;
  }
}

/**
 * Time-based condition (only usable during certain time windows)
 */
export class TimeCondition extends BaseCondition {
  private config: TimeConditionConfig;

  constructor(config: TimeConditionConfig) {
    super(
      `time_${config.timeWindow.start}_${config.timeWindow.end}`,
      `Time window: ${config.timeWindow.start}-${config.timeWindow.end}ms`,
      `Only usable between ${config.timeWindow.start}ms and ${config.timeWindow.end}ms`
    );
    this.config = config;
  }

  check(context: AbilityContext): boolean {
    const currentTime = this.config.gameTime 
      ? (context.scene as any)?.time?.now || Date.now()
      : Date.now();

    return currentTime >= this.config.timeWindow.start && 
           currentTime <= this.config.timeWindow.end;
  }
}

/**
 * Combo-based condition (requires specific ability sequence)
 */
export class ComboCondition extends BaseCondition {
  private config: ComboConditionConfig;
  private static recentAbilities: Map<any, Array<{id: string, timestamp: number}>> = new Map();

  constructor(config: ComboConditionConfig) {
    super(
      `combo_${config.requiredSequence.join('_')}`,
      `Combo: ${config.requiredSequence.join(' -> ')}`,
      `Requires the ability sequence: ${config.requiredSequence.join(' -> ')}`
    );
    this.config = config;
  }

  check(context: AbilityContext): boolean {
    const owner = context.owner;
    if (!owner) return false;

    const recentAbilities = ComboCondition.recentAbilities.get(owner) || [];
    const now = Date.now();

    // Clean old abilities outside the time window
    const validAbilities = recentAbilities.filter(
      ability => now - ability.timestamp <= this.config.maxInterval * this.config.requiredSequence.length
    );

    if (validAbilities.length < this.config.requiredSequence.length - 1) {
      return false; // Not enough recent abilities
    }

    // Check if the sequence matches
    const sequence = validAbilities.slice(-(this.config.requiredSequence.length - 1));
    
    if (this.config.mustBeExact) {
      // Exact sequence match
      return sequence.every((ability, index) => 
        ability.id === this.config.requiredSequence[index]
      );
    } else {
      // Subsequence match (can have other abilities in between)
      let sequenceIndex = 0;
      for (const ability of sequence) {
        if (ability.id === this.config.requiredSequence[sequenceIndex]) {
          sequenceIndex++;
        }
      }
      return sequenceIndex === this.config.requiredSequence.length - 1;
    }
  }

  static recordAbilityUse(owner: any, abilityId: string): void {
    if (!ComboCondition.recentAbilities.has(owner)) {
      ComboCondition.recentAbilities.set(owner, []);
    }
    
    const abilities = ComboCondition.recentAbilities.get(owner)!;
    abilities.push({ id: abilityId, timestamp: Date.now() });

    // Keep only recent abilities (last 10 to prevent memory leaks)
    if (abilities.length > 10) {
      abilities.shift();
    }
  }
}

/**
 * Resource-based condition (can pay cost, has minimum amount, etc.)
 */
export class ResourceCondition extends BaseCondition {
  private config: ResourceConditionConfig;

  constructor(config: ResourceConditionConfig) {
    const opDesc = config.operation === 'cost' ? 'can pay' :
                   config.operation === 'minimum' ? 'has at least' : 'has exactly';
    super(
      `resource_${config.attribute}_${config.operation}_${config.amount}`,
      `${opDesc} ${config.amount} ${config.attribute}`,
      `Requires ${opDesc} ${config.amount} ${config.attribute}`
    );
    this.config = config;
  }

  check(context: AbilityContext): boolean {
    const owner = context.owner;
    if (!owner?.abilitySystem) return false;

    const currentValue = owner.abilitySystem.getAttributeValue(this.config.attribute);
    
    switch (this.config.operation) {
      case 'cost': return currentValue >= this.config.amount;
      case 'minimum': return currentValue >= this.config.amount;
      case 'exact': return currentValue === this.config.amount;
      default: return false;
    }
  }
}

/**
 * Condition Manager - handles multiple conditions
 */
export class ConditionManager {
  private conditions: Map<string, AbilityCondition> = new Map();

  addCondition(condition: AbilityCondition): void {
    this.conditions.set(condition.id, condition);
  }

  removeCondition(conditionId: string): void {
    this.conditions.delete(conditionId);
  }

  getCondition(conditionId: string): AbilityCondition | undefined {
    return this.conditions.get(conditionId);
  }

  async checkConditions(
    conditionIds: string[], 
    context: AbilityContext,
    skipConditions: string[] = []
  ): Promise<ConditionResult> {
    const relevantConditions = conditionIds
      .filter(id => !skipConditions.includes(id))
      .map(id => this.conditions.get(id))
      .filter(condition => condition !== undefined) as AbilityCondition[];

    for (const condition of relevantConditions) {
      try {
        const passed = await condition.check(context);
        if (!passed) {
          condition.onFailure?.(context, `Condition '${condition.name}' not met`);
          return {
            passed: false,
            reason: `Condition failed: ${condition.name}`,
            data: { failedCondition: condition.id }
          };
        }
      } catch (error) {
        console.error(`Error checking condition '${condition.name}':`, error);
        return {
          passed: false,
          reason: `Condition error: ${condition.name}`,
          data: { error: error instanceof Error ? error.message : 'Unknown error' }
        };
      }
    }

    return { passed: true };
  }

  // Factory methods for common conditions
  static createAttributeCondition(config: AttributeConditionConfig): AttributeCondition {
    return new AttributeCondition(config);
  }

  static createTagCondition(config: TagConditionConfig): TagCondition {
    return new TagCondition(config);
  }

  static createCooldownCondition(config: CooldownConditionConfig): CooldownCondition {
    return new CooldownCondition(config);
  }

  static createTimeCondition(config: TimeConditionConfig): TimeCondition {
    return new TimeCondition(config);
  }

  static createComboCondition(config: ComboConditionConfig): ComboCondition {
    return new ComboCondition(config);
  }

  static createResourceCondition(config: ResourceConditionConfig): ResourceCondition {
    return new ResourceCondition(config);
  }
}