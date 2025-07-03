# 🎮 GAS - Gameplay Ability System

A complete implementation of Unreal Engine's **Gameplay Ability System (GAS)** for JavaScript/TypeScript games with **versioned components** for seamless migration.

## 🚀 Version Overview

| Version | Status | Features | Use Case |
|---------|---------|----------|----------|
| **v1** | 🟢 Stable | Core GAS functionality | Production, existing projects |
| **v2** | 🟡 Enhanced | + Events + Conditions | New projects, advanced features |

## 📦 Installation & Usage

### 🔄 **Three Ways to Import**

#### 1. **Default (V1 - Backward Compatible)**
```typescript
// Uses stable v1 by default
import { AbilitySystemComponent, GameplayEffect } from '@/packages/gas';

const asc = new AbilitySystemComponent(player);
```

#### 2. **Version-Specific Import**
```typescript
// Explicitly choose version
import { v1, v2 } from '@/packages/gas';

// V1 - Stable
const ascV1 = new v1.AbilitySystemComponent(player);

// V2 - Enhanced
const ascV2 = new v2.AbilitySystemComponent(player);
```

#### 3. **Mixed Approach (Recommended for Migration)**
```typescript
// Use v1 for core, v2 for new features
import { AbilitySystemComponent } from '@/packages/gas'; // v1
import { ConditionManager, EnhancedEventSystem } from '@/packages/gas'; // v2 features

const asc = new AbilitySystemComponent(player); // v1
const conditionManager = new ConditionManager(); // v2
```

---

## 📋 **V1 - Stable (Production Ready)**

### ✅ **Features**
- ✅ Ability System Component (ASC)
- ✅ Gameplay Abilities with cooldowns
- ✅ Gameplay Effects (buffs/debuffs)
- ✅ Gameplay Attributes (health, mana, etc.)
- ✅ Gameplay Tags
- ✅ Basic event system
- ✅ Cost management

### 🎯 **Quick Start V1**
```typescript
import { v1 } from '@/packages/gas';

// Create ASC
const asc = new v1.AbilitySystemComponent(player);

// Add attributes
asc.addAttribute('health', 100, 100);
asc.addAttribute('mana', 50, 50);

// Create ability
class FireballAbility extends v1.GameplayAbility {
  readonly id = 'fireball';
  readonly name = 'Fireball';
  readonly cooldown = 3000;
  
  async activate(context) {
    const damage = v1.GameplayEffect.createInstantDamage(50);
    context.target.abilitySystem.applyGameplayEffect(damage);
    return true;
  }
}

// Grant and use
asc.grantAbility(new FireballAbility());
asc.tryActivateAbility('fireball', { owner: player, target: enemy });
```

---

## 🚀 **V2 - Enhanced (Advanced Features)**

### ✨ **Additional Features**
- 🎯 **Enhanced Event System** - Priority, filtering, history
- 🔍 **Condition System** - Complex ability requirements  
- 🔄 **Combo System** - Ability chaining and sequences
- 📊 **Advanced Debugging** - Detailed state inspection
- ⚡ **Better Performance** - Optimized event handling

### 🎯 **Quick Start V2**
```typescript
import { v2 } from '@/packages/gas';

// Create enhanced ASC
const asc = new v2.AbilitySystemComponent(player);

// Enhanced event handling with priority and filtering
asc.on('ability-activated', (data) => {
  console.log(`${data.abilityId} used at ${data.timestamp}`);
}, { 
  priority: v2.EventPriority.HIGH,
  filter: (data) => data.abilityId === 'fireball'
});

// Advanced conditions
const conditions = [
  v2.ConditionManager.createAttributeCondition({
    attribute: 'health',
    operator: '>',
    value: 50,
    percentage: true // 50% of max health
  }),
  v2.ConditionManager.createTagCondition({
    tags: ['stunned', 'silenced'],
    mode: 'none' // Must not have these tags
  })
];

// Enhanced ability activation
asc.tryActivateAbility('fireball', {
  owner: player,
  target: enemy,
  conditions: conditions,
  metadata: { source: 'player-input' }
});
```

---

## 🔄 **Migration Strategy**

### 📈 **Gradual Migration Path**

1. **Phase 1: Keep Existing (V1)**
   ```typescript
   // No changes needed - existing code works
   import { AbilitySystemComponent } from '@/packages/gas';
   ```

2. **Phase 2: Add V2 Features**
   ```typescript
   // Add enhanced features alongside existing
   import { AbilitySystemComponent } from '@/packages/gas'; // v1
   import { ConditionManager } from '@/packages/gas'; // v2
   ```

3. **Phase 3: Full V2 Migration**
   ```typescript
   // Switch to v2 when ready
   import { v2 } from '@/packages/gas';
   const asc = new v2.AbilitySystemComponent(player);
   ```

---

## 🎮 **Real-World Examples**

### 🏟️ **Ability Arena (V2 Enhanced)**
```typescript
import { v2, ConditionManager } from '@/packages/gas';

class LightningBoltAbility extends v2.GameplayAbility {
  readonly id = 'lightning_bolt';
  readonly name = 'Lightning Bolt';
  readonly cooldown = 3000;

  async activate(context) {
    // V2 enhanced context
    const { owner, target, metadata } = context;
    
    // Use enhanced event system
    context.owner.abilitySystem.emit('ability-charging', {
      abilityId: this.id,
      chargeTime: 1000
    });

    return true;
  }
}

// Advanced combo condition
const comboCondition = ConditionManager.createComboCondition({
  requiredSequence: ['fireball', 'ice_spikes'],
  maxInterval: 3000,
  mustBeExact: false
});
```

### ⚔️ **Minimal Legion (V1 Stable)**
```typescript
import { AbilitySystemComponent, GameplayEffect } from '@/packages/gas';

// Simple, reliable v1 usage
class TowerDefensePlayer {
  constructor() {
    this.abilitySystem = new AbilitySystemComponent(this);
    this.abilitySystem.addAttribute('health', 100);
    this.abilitySystem.addAttribute('mana', 50);
  }
  
  castFireball(target) {
    return this.abilitySystem.tryActivateAbility('fireball', {
      owner: this,
      target: target
    });
  }
}
```

---

## 🔍 **Feature Comparison**

| Feature | V1 | V2 | Notes |
|---------|----|----|-------|
| Basic ASC | ✅ | ✅ | Same API |
| Abilities | ✅ | ✅ | Same API |
| Effects | ✅ | ✅ | Same API |
| Events | Basic | Enhanced | V2 has priority, filtering |
| Conditions | Manual | System | V2 has built-in condition types |
| Debugging | Basic | Advanced | V2 has detailed state info |
| Performance | Good | Better | V2 optimized event handling |
| Bundle Size | Smaller | Larger | V2 includes more features |
| Ability Queue | ❌ | ✅ | V2 only: Priority-based queuing |
| Serialization | ❌ | ✅ | V2 only: Save/load, JSON export |

---

## 🛠️ **When to Use Which Version**

### 🟢 **Use V1 When:**
- ✅ Building production systems
- ✅ Need maximum stability  
- ✅ Want minimal bundle size
- ✅ Simple ability requirements
- ✅ Existing codebase migration

### 🚀 **Use V2 When:**
- ✅ Building new complex systems
- ✅ Need advanced event handling
- ✅ Complex ability conditions
- ✅ Combo/chain systems
- ✅ Advanced debugging needs
- ✅ Real-time multiplayer
- ✅ Save/load functionality needed
- ✅ Data persistence required

---

## 📚 **API Documentation**

### V1 Core Methods
```typescript
// ASC v1
asc.addAttribute(name, baseValue, maxValue?)
asc.grantAbility(ability)
asc.tryActivateAbility(id, context)
asc.applyGameplayEffect(effect)
asc.on(event, handler) // Basic events
```

### V2 Enhanced Methods
```typescript
// ASC v2 (includes all v1 methods plus:)
asc.addCondition(condition)
asc.addGlobalCondition(conditionId)
asc.on(event, handler, options) // Enhanced events
asc.emit(event, data) // Manual event emission
asc.getDebugInfo() // Detailed state

// Ability Queue System
asc.queueAbility(abilityId, context, options)
asc.cancelQueuedAbility(queueId)
asc.processAbilityQueue()

// Serialization System
gasSerializer.createSnapshot(asc)
gasSerializer.createSaveState(asc)
gasSerializer.exportConfiguration(asc)
```

---

## 🎯 **Version History**

- **v1.0.0** - Stable core GAS implementation
- **v2.0.0** - Enhanced events, conditions, ability queue, serialization

---

*Built with ❤️ for the Defense Allies project*