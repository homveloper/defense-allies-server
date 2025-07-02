# Minimal Legion - Ability System Usage Examples

## 🎯 Usage Scenarios

이 문서는 설계된 Ability System의 구체적인 사용 예제들을 보여줍니다.

## 🔥 Example 1: Fireball Ability

### Ability Definition
```typescript
class FireballAbility extends GameplayAbility {
  id = 'fireball';
  name = 'Fireball';
  description = 'Launches a fireball that deals damage to enemies';
  
  costs = [
    { attribute: 'mana', amount: 15 }
  ];
  cooldown = 3000; // 3 seconds
  
  requiredTags = [];
  blockedByTags = ['silenced', 'stunned'];
  
  async activate(context: AbilityContext): Promise<boolean> {
    const { owner, target, scene } = context;
    
    // Create projectile
    const projectile = scene.add.sprite(owner.x, owner.y, 'fireball');
    
    // Apply movement effect
    scene.tweens.add({
      targets: projectile,
      x: target.x,
      y: target.y,
      duration: 500,
      onComplete: () => {
        // Create damage effect
        const damageEffect = new DamageEffect({
          id: 'fireball_damage',
          damage: 35,
          damageType: 'fire'
        });
        
        // Apply to target
        target.abilitySystem.applyGameplayEffect(damageEffect);
        
        // Visual explosion
        this.createExplosionEffect(scene, target.x, target.y);
        
        projectile.destroy();
      }
    });
    
    return true;
  }
  
  private createExplosionEffect(scene: Phaser.Scene, x: number, y: number) {
    const explosion = scene.add.graphics();
    explosion.fillStyle(0xff4500, 0.8);
    explosion.fillCircle(x, y, 30);
    
    scene.tweens.add({
      targets: explosion,
      scale: { from: 0.1, to: 2 },
      alpha: { from: 1, to: 0 },
      duration: 300,
      onComplete: () => explosion.destroy()
    });
  }
}
```

### Usage
```typescript
// Player learns fireball
player.abilitySystem.grantAbility(new FireballAbility());

// Player tries to cast fireball
const target = nearestEnemy;
const success = player.abilitySystem.tryActivateAbility('fireball', { target });

if (success) {
  console.log('Fireball cast successfully!');
} else {
  console.log('Cannot cast fireball - insufficient mana or on cooldown');
}
```

## 🛡️ Example 2: Shield Buff Effect

### Effect Definition
```typescript
class ShieldBuffEffect extends GameplayEffect {
  constructor(magnitude: number, duration: number) {
    super({
      id: 'shield_buff',
      name: 'Magic Shield',
      duration: duration,
      
      attributeModifiers: [
        {
          id: 'shield_defense',
          attribute: 'defense',
          operation: 'add',
          magnitude: magnitude
        }
      ],
      
      grantedTags: ['buffed', 'shielded'],
      
      onApplied: (target) => {
        // Visual shield effect
        this.createShieldVisual(target);
      },
      
      onRemoved: (target) => {
        // Remove visual shield
        this.removeShieldVisual(target);
      }
    });
  }
  
  private createShieldVisual(target: any) {
    const shield = target.scene.add.graphics();
    shield.lineStyle(3, 0x3498db, 0.6);
    shield.strokeCircle(0, 0, 25);
    target.add(shield);
    target.shieldVisual = shield;
    
    // Pulse animation
    target.scene.tweens.add({
      targets: shield,
      alpha: { from: 0.6, to: 0.3 },
      duration: 1000,
      yoyo: true,
      repeat: -1
    });
  }
  
  private removeShieldVisual(target: any) {
    if (target.shieldVisual) {
      target.shieldVisual.destroy();
      target.shieldVisual = null;
    }
  }
}
```

### Shield Ability
```typescript
class ShieldAbility extends GameplayAbility {
  id = 'shield';
  name = 'Magic Shield';
  description = 'Creates a protective barrier that increases defense';
  
  costs = [
    { attribute: 'mana', amount: 20 }
  ];
  cooldown = 8000; // 8 seconds
  
  async activate(context: AbilityContext): Promise<boolean> {
    const { owner } = context;
    
    // Create shield effect
    const shieldEffect = new ShieldBuffEffect(15, 5000); // +15 defense for 5 seconds
    
    // Apply to self
    owner.abilitySystem.applyGameplayEffect(shieldEffect);
    
    return true;
  }
}
```

## ⚡ Example 3: Lightning Chain Ability

### Complex Multi-Target Ability
```typescript
class LightningChainAbility extends GameplayAbility {
  id = 'lightning_chain';
  name = 'Chain Lightning';
  description = 'Lightning that jumps between enemies';
  
  costs = [
    { attribute: 'mana', amount: 30 }
  ];
  cooldown = 6000;
  
  async activate(context: AbilityContext): Promise<boolean> {
    const { owner, target, scene } = context;
    
    // Find chain targets
    const targets = this.findChainTargets(owner, target, 3); // Max 3 jumps
    
    // Execute chain
    await this.executeChain(scene, targets);
    
    return true;
  }
  
  private findChainTargets(owner: any, initialTarget: any, maxJumps: number): any[] {
    const targets = [initialTarget];
    let currentTarget = initialTarget;
    
    for (let i = 0; i < maxJumps - 1; i++) {
      const nextTarget = this.findNearestUntargetedEnemy(currentTarget, targets, 100);
      if (!nextTarget) break;
      
      targets.push(nextTarget);
      currentTarget = nextTarget;
    }
    
    return targets;
  }
  
  private async executeChain(scene: Phaser.Scene, targets: any[]): Promise<void> {
    for (let i = 0; i < targets.length; i++) {
      const target = targets[i];
      const prevTarget = i > 0 ? targets[i - 1] : null;
      
      // Create lightning visual
      if (prevTarget) {
        this.createLightningBolt(scene, prevTarget, target);
      }
      
      // Calculate damage (decreases with each jump)
      const baseDamage = 40;
      const damage = baseDamage * Math.pow(0.75, i); // 25% reduction per jump
      
      // Apply damage
      const damageEffect = new DamageEffect({
        id: `lightning_damage_${i}`,
        damage: damage,
        damageType: 'lightning'
      });
      
      target.abilitySystem.applyGameplayEffect(damageEffect);
      
      // Delay between jumps
      if (i < targets.length - 1) {
        await this.delay(200);
      }
    }
  }
  
  private createLightningBolt(scene: Phaser.Scene, from: any, to: any) {
    const lightning = scene.add.graphics();
    lightning.lineStyle(3, 0xffff00, 1);
    
    // Create jagged lightning path
    const steps = 5;
    const dx = (to.x - from.x) / steps;
    const dy = (to.y - from.y) / steps;
    
    lightning.moveTo(from.x, from.y);
    
    for (let i = 1; i <= steps; i++) {
      const x = from.x + dx * i + (Math.random() - 0.5) * 20;
      const y = from.y + dy * i + (Math.random() - 0.5) * 20;
      lightning.lineTo(x, y);
    }
    
    lightning.strokePath();
    
    // Fade out lightning
    scene.tweens.add({
      targets: lightning,
      alpha: 0,
      duration: 300,
      onComplete: () => lightning.destroy()
    });
  }
  
  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

## 🩹 Example 4: Heal Over Time Effect

### Periodic Healing Effect
```typescript
class HealOverTimeEffect extends GameplayEffect {
  constructor(healPerTick: number, duration: number, tickRate: number = 1000) {
    super({
      id: 'heal_over_time',
      name: 'Regeneration',
      duration: duration,
      period: tickRate,
      
      grantedTags: ['healing', 'regenerating'],
      
      onPeriodic: (target) => {
        // Heal the target
        const healthAttr = target.abilitySystem.getAttribute('health');
        if (healthAttr) {
          const newValue = Math.min(
            healthAttr.maxValue || healthAttr.currentValue,
            healthAttr.currentValue + healPerTick
          );
          target.abilitySystem.setAttributeValue('health', newValue);
          
          // Visual healing effect
          this.showHealingNumbers(target, healPerTick);
        }
      }
    });
  }
  
  private showHealingNumbers(target: any, amount: number) {
    const healText = target.scene.add.text(target.x, target.y - 20, `+${amount}`, {
      font: '12px Arial',
      color: '#00ff00'
    });
    healText.setOrigin(0.5);
    
    target.scene.tweens.add({
      targets: healText,
      y: healText.y - 30,
      alpha: 0,
      duration: 1000,
      onComplete: () => healText.destroy()
    });
  }
}
```

## 🎮 Example 5: Player Integration

### Complete Player Setup
```typescript
export class EnhancedPlayer extends Player {
  public abilitySystem: AbilitySystemComponent;
  
  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, x, y);
    
    // Initialize ability system
    this.abilitySystem = new AbilitySystemComponent(this);
    
    // Setup attributes
    this.setupAttributes();
    
    // Grant starting abilities
    this.setupStartingAbilities();
    
    // Setup input handling
    this.setupAbilityInput();
  }
  
  private setupAttributes() {
    this.abilitySystem.addAttribute('health', 100, 100);
    this.abilitySystem.addAttribute('mana', 60, 60);
    this.abilitySystem.addAttribute('attackPower', 25);
    this.abilitySystem.addAttribute('defense', 10);
    this.abilitySystem.addAttribute('moveSpeed', 100);
  }
  
  private setupStartingAbilities() {
    this.abilitySystem.grantAbility(new FireballAbility());
    this.abilitySystem.grantAbility(new ShieldAbility());
    this.abilitySystem.grantAbility(new HealAbility());
  }
  
  private setupAbilityInput() {
    // Q key for fireball
    this.scene.input.keyboard!.on('keydown-Q', () => {
      const target = this.findNearestEnemy();
      if (target) {
        this.abilitySystem.tryActivateAbility('fireball', { target });
      }
    });
    
    // E key for shield
    this.scene.input.keyboard!.on('keydown-E', () => {
      this.abilitySystem.tryActivateAbility('shield');
    });
    
    // R key for heal
    this.scene.input.keyboard!.on('keydown-R', () => {
      this.abilitySystem.tryActivateAbility('heal');
    });
  }
  
  update() {
    super.update();
    
    // Update ability system
    this.abilitySystem.update(this.scene.game.loop.delta);
  }
}
```

## 🏷️ Example 6: Gameplay Tags Usage

### Tag-Based Conditional Logic
```typescript
// Stun effect that prevents ability use
class StunEffect extends GameplayEffect {
  constructor(duration: number) {
    super({
      id: 'stun',
      name: 'Stunned',
      duration: duration,
      grantedTags: ['stunned', 'disabled', 'crowd_controlled'],
      
      onApplied: (target) => {
        // Visual stun effect
        const stunIcon = target.scene.add.image(target.x, target.y - 30, 'stun_icon');
        target.stunIcon = stunIcon;
      },
      
      onRemoved: (target) => {
        if (target.stunIcon) {
          target.stunIcon.destroy();
          target.stunIcon = null;
        }
      }
    });
  }
}

// Ability that checks for stun
class MovementAbility extends GameplayAbility {
  canActivate(context: AbilityContext): boolean {
    const { owner } = context;
    
    // Cannot move while stunned
    if (owner.abilitySystem.hasTag('stunned')) {
      return false;
    }
    
    return super.canActivate(context);
  }
}
```

## 📊 Example 7: Attribute Calculations

### Complex Attribute Modifiers
```typescript
// Example: Character with multiple defense modifiers
const player = new EnhancedPlayer(scene, 100, 100);

// Base defense: 10
console.log(player.abilitySystem.getAttributeValue('defense')); // 10

// Apply shield buff (+15 defense)
const shieldEffect = new ShieldBuffEffect(15, 5000);
player.abilitySystem.applyGameplayEffect(shieldEffect);
console.log(player.abilitySystem.getAttributeValue('defense')); // 25

// Apply armor item (+20% defense multiplier)
const armorEffect = new AttributeModifierEffect('defense', 'multiply', 1.2);
player.abilitySystem.applyGameplayEffect(armorEffect);
console.log(player.abilitySystem.getAttributeValue('defense')); // 30 (25 * 1.2)

// Apply curse (-5 defense)
const curseEffect = new AttributeModifierEffect('defense', 'add', -5);
player.abilitySystem.applyGameplayEffect(curseEffect);
console.log(player.abilitySystem.getAttributeValue('defense')); // 24 ((10 + 15 - 5) * 1.2)
```

이 예제들은 실제 구현할 때 참고할 수 있는 구체적인 사용 패턴들을 보여줍니다. 각 예제는 단계적으로 구현하여 테스트할 수 있습니다.