// Example of how to use the Ability System
// This file demonstrates basic usage patterns

import { 
  AbilitySystemComponent, 
  GameplayEffect, 
  BasicAttackAbility, 
  FireballAbility, 
  HealAbility,
  AbilitySystemUtils
} from '../index';

/**
 * Example Player class with Ability System integration
 */
export class ExamplePlayer {
  public abilitySystem: AbilitySystemComponent;
  public x: number = 0;
  public y: number = 0;
  public scene: any;
  public active: boolean = true;

  constructor(scene: any, x: number, y: number) {
    this.scene = scene;
    this.x = x;
    this.y = y;
    
    // Initialize ability system using utility
    this.abilitySystem = AbilitySystemUtils.createPlayerAbilitySystem(this);
    
    // Add fireball ability
    this.abilitySystem.grantAbility(new FireballAbility());
    
    // Set up event listeners
    this.setupEventListeners();
  }

  private setupEventListeners(): void {
    // Listen for ability activations
    this.abilitySystem.on('ability-activated', (data) => {
      console.log(`Player activated ${data.abilityId}`);
    });

    // Listen for attribute changes
    this.abilitySystem.on('attribute-changed', (data) => {
      console.log(`${data.attribute} changed from ${data.oldValue} to ${data.newValue}`);
      
      // Handle death
      if (data.attribute === 'health' && data.newValue <= 0) {
        this.onDeath();
      }
    });

    // Listen for effects
    this.abilitySystem.on('effect-applied', (data) => {
      console.log(`Effect ${data.effectId} applied to player`);
    });
  }

  // Update method to be called each frame
  update(deltaTime: number): void {
    this.abilitySystem.update(deltaTime);
  }

  // Try to attack nearest enemy
  attackNearestEnemy(enemies: any[]): boolean {
    const target = AbilitySystemUtils.findNearestEnemy(this, enemies, 150);
    if (target) {
      const result = this.abilitySystem.tryActivateAbility('basic_attack', { 
        target, 
        scene: this.scene 
      });
      return result.success;
    }
    return false;
  }

  // Cast fireball at target
  castFireball(target: any): boolean {
    const result = this.abilitySystem.tryActivateAbility('fireball', { 
      target, 
      scene: this.scene 
    });
    return result.success;
  }

  // Heal self
  heal(): boolean {
    const result = this.abilitySystem.tryActivateAbility('heal', { 
      scene: this.scene 
    });
    return result.success;
  }

  // Apply damage (for compatibility with existing system)
  takeDamage(amount: number): void {
    const damageEffect = GameplayEffect.createInstantDamage(amount);
    this.abilitySystem.applyGameplayEffect(damageEffect);
  }

  // Get current health
  get health(): number {
    return this.abilitySystem.getAttributeValue('health');
  }

  // Get current mana
  get mana(): number {
    return this.abilitySystem.getAttributeValue('mana');
  }

  private onDeath(): void {
    console.log('Player died!');
    this.active = false;
    // Add death logic here
  }

  // Debug helper
  logState(): void {
    AbilitySystemUtils.logAbilitySystemState(this.abilitySystem, 'Player');
  }
}

/**
 * Example Enemy class with Ability System
 */
export class ExampleEnemy {
  public abilitySystem: AbilitySystemComponent;
  public x: number = 0;
  public y: number = 0;
  public scene: any;
  public active: boolean = true;

  constructor(scene: any, x: number, y: number, config?: {
    health?: number;
    attackPower?: number;
  }) {
    this.scene = scene;
    this.x = x;
    this.y = y;
    
    // Initialize ability system
    this.abilitySystem = AbilitySystemUtils.createEnemyAbilitySystem(this, config);
    
    // Set up death listener
    this.abilitySystem.on('attribute-changed', (data) => {
      if (data.attribute === 'health' && data.newValue <= 0) {
        this.onDeath();
      }
    });
  }

  update(deltaTime: number): void {
    this.abilitySystem.update(deltaTime);
  }

  // Simple AI: attack player if in range
  tryAttackPlayer(player: ExamplePlayer): boolean {
    const distance = Math.sqrt(
      Math.pow(this.x - player.x, 2) + Math.pow(this.y - player.y, 2)
    );
    
    if (distance <= 100) { // Attack range
      const result = this.abilitySystem.tryActivateAbility('basic_attack', { 
        target: player, 
        scene: this.scene 
      });
      return result.success;
    }
    
    return false;
  }

  takeDamage(amount: number): void {
    const damageEffect = GameplayEffect.createInstantDamage(amount);
    this.abilitySystem.applyGameplayEffect(damageEffect);
  }

  get health(): number {
    return this.abilitySystem.getAttributeValue('health');
  }

  private onDeath(): void {
    console.log('Enemy died!');
    this.active = false;
    // Add death logic, drops, experience, etc.
  }
}

/**
 * Example game loop demonstrating ability system usage
 */
export class AbilitySystemExample {
  private player: ExamplePlayer;
  private enemies: ExampleEnemy[] = [];
  private scene: any;

  constructor(scene: any) {
    this.scene = scene;
    this.player = new ExamplePlayer(scene, 100, 100);
    
    // Create some enemies
    this.enemies = [
      new ExampleEnemy(scene, 200, 150, { health: 40, attackPower: 12 }),
      new ExampleEnemy(scene, 300, 120, { health: 30, attackPower: 8 }),
      new ExampleEnemy(scene, 250, 200, { health: 60, attackPower: 15 })
    ];
  }

  // Main update loop
  update(deltaTime: number): void {
    // Update player
    this.player.update(deltaTime);
    
    // Update enemies
    this.enemies.forEach(enemy => {
      if (enemy.active) {
        enemy.update(deltaTime);
        // Simple AI
        enemy.tryAttackPlayer(this.player);
      }
    });
    
    // Remove dead enemies
    this.enemies = this.enemies.filter(enemy => enemy.active);
  }

  // Example input handling
  handleInput(key: string): void {
    switch (key) {
      case 'SPACE':
        // Attack nearest enemy
        this.player.attackNearestEnemy(this.enemies);
        break;
        
      case 'Q':
        // Cast fireball on nearest enemy
        const target = AbilitySystemUtils.findNearestEnemy(this.player, this.enemies);
        if (target) {
          this.player.castFireball(target);
        }
        break;
        
      case 'E':
        // Heal
        this.player.heal();
        break;
        
      case 'T':
        // Apply test buff
        const buffEffect = GameplayEffect.createAttributeBuff('attackPower', 15, 5000);
        this.player.abilitySystem.applyGameplayEffect(buffEffect);
        break;
        
      case 'R':
        // Apply test debuff to nearest enemy
        const nearestEnemy = AbilitySystemUtils.findNearestEnemy(this.player, this.enemies);
        if (nearestEnemy) {
          const poisonEffect = GameplayEffect.createDamageOverTime(5, 3000, 1000);
          nearestEnemy.abilitySystem.applyGameplayEffect(poisonEffect);
        }
        break;
        
      case 'L':
        // Log player state
        this.player.logState();
        break;
    }
  }

  // Get game state for debugging
  getGameState(): any {
    return {
      player: {
        health: this.player.health,
        mana: this.player.mana,
        position: { x: this.player.x, y: this.player.y },
        abilities: this.player.abilitySystem.getAllAbilities().map(a => a.name),
        activeEffects: this.player.abilitySystem.getAllActiveEffects().length
      },
      enemies: this.enemies.map(enemy => ({
        health: enemy.health,
        position: { x: enemy.x, y: enemy.y },
        activeEffects: enemy.abilitySystem.getAllActiveEffects().length
      }))
    };
  }
}

// Usage example:
/*
// In your Phaser scene:
const abilityExample = new AbilitySystemExample(this);

// In your update loop:
abilityExample.update(deltaTime);

// In your input handling:
this.input.keyboard.on('keydown-SPACE', () => abilityExample.handleInput('SPACE'));
this.input.keyboard.on('keydown-Q', () => abilityExample.handleInput('Q'));
this.input.keyboard.on('keydown-E', () => abilityExample.handleInput('E'));
*/