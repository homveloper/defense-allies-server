// Example: Using the Ability Queue System in GAS v2
// Demonstrates priority-based ability queueing, delays, and execution modes

import { v2 } from '../index';
import { QueueExecutionMode } from '../v2/AbilityQueue';

// Example abilities for demonstration
class QuickStrikeAbility extends v2.GameplayAbility {
  readonly id = 'quick_strike';
  readonly name = 'Quick Strike';
  readonly description = 'Fast attack with low damage';
  readonly cooldown = 1000; // 1 second

  async activate(context: any): Promise<boolean> {
    console.log(`${context.owner.name} performs Quick Strike!`);
    // Quick animation, immediate damage
    return true;
  }
}

class PowerSlamAbility extends v2.GameplayAbility {
  readonly id = 'power_slam';
  readonly name = 'Power Slam';
  readonly description = 'Slow but powerful attack';
  readonly cooldown = 3000; // 3 seconds

  async activate(context: any): Promise<boolean> {
    console.log(`${context.owner.name} charges up Power Slam!`);
    // Longer animation, heavy damage
    return true;
  }
}

class HealAbility extends v2.GameplayAbility {
  readonly id = 'heal';
  readonly name = 'Heal';
  readonly description = 'Restore health over time';
  readonly cooldown = 5000; // 5 seconds

  async activate(context: any): Promise<boolean> {
    console.log(`${context.owner.name} casts Heal!`);
    return true;
  }
}

// Example player/character
class ExamplePlayer {
  constructor(public name: string) {
    this.abilitySystem = new v2.AbilitySystemComponent(this);
    this.setupAbilities();
    this.setupEventHandlers();
  }

  abilitySystem: v2.AbilitySystemComponent;

  private setupAbilities(): void {
    // Add attributes
    this.abilitySystem.addAttribute('health', 100, 100);
    this.abilitySystem.addAttribute('mana', 50, 50);

    // Grant abilities
    this.abilitySystem.grantAbility(new QuickStrikeAbility());
    this.abilitySystem.grantAbility(new PowerSlamAbility());
    this.abilitySystem.grantAbility(new HealAbility());
  }

  private setupEventHandlers(): void {
    // Queue events
    this.abilitySystem.on('ability-queued', (data) => {
      console.log(`📋 Queued ${data.abilityId} (priority: ${data.priority}, queue size: ${data.queueSize})`);
    });

    this.abilitySystem.on('ability-queue-executed', (data) => {
      console.log(`✅ Executed ${data.abilityId} after ${data.waitTime}ms wait`);
    });

    this.abilitySystem.on('ability-queue-cancelled', (data) => {
      console.log(`❌ Cancelled ${data.abilityId}: ${data.reason}`);
    });

    this.abilitySystem.on('ability-queue-interrupted', (data) => {
      console.log(`⚡ Interrupted ${data.abilityId}`);
    });
  }

  // Convenience methods for demonstration
  queueQuickStrike(priority: number = 0, delay: number = 0): string {
    return this.abilitySystem.queueAbility('quick_strike', {
      owner: this,
      scene: null as any
    }, { priority, delay });
  }

  queuePowerSlam(priority: number = 1, delay: number = 0): string {
    return this.abilitySystem.queueAbility('power_slam', {
      owner: this,
      scene: null as any
    }, { priority, delay });
  }

  queueHeal(priority: number = 2, delay: number = 0): string {
    return this.abilitySystem.queueAbility('heal', {
      owner: this,
      scene: null as any
    }, { priority, delay });
  }
}

// === EXAMPLE SCENARIOS ===

async function demonstrateBasicQueueing(): Promise<void> {
  console.log('\n=== Basic Queue Example ===');
  
  const player = new ExamplePlayer('Hero');
  
  // Queue abilities in different order
  console.log('Queueing abilities...');
  player.queueQuickStrike(0); // Low priority
  player.queueHeal(2);        // High priority
  player.queuePowerSlam(1);   // Medium priority
  
  console.log(`Queue size: ${player.abilitySystem.getQueueSize()}`);
  
  // Execute queue - should execute in priority order: Heal (2) -> Power Slam (1) -> Quick Strike (0)
  console.log('\nProcessing queue...');
  await player.abilitySystem.processAbilityQueue();
}

async function demonstrateDelayedExecution(): Promise<void> {
  console.log('\n=== Delayed Execution Example ===');
  
  const player = new ExamplePlayer('Mage');
  
  // Queue abilities with delays
  console.log('Queueing abilities with delays...');
  player.queueQuickStrike(0, 0);     // Execute immediately
  player.queuePowerSlam(0, 2000);    // Execute after 2 seconds
  player.queueHeal(0, 1000);         // Execute after 1 second
  
  console.log('Processing queue with delays...');
  await player.abilitySystem.processAbilityQueue();
  
  // Wait for delayed abilities
  await new Promise(resolve => setTimeout(resolve, 3000));
}

async function demonstrateInterruption(): Promise<void> {
  console.log('\n=== Interruption Example ===');
  
  const player = new ExamplePlayer('Warrior');
  
  // Queue multiple abilities
  console.log('Queueing abilities...');
  const quickId = player.queueQuickStrike(0);
  const powerSlamId = player.queuePowerSlam(1);
  const healId = player.queueHeal(2);
  
  console.log(`Queued 3 abilities, queue size: ${player.abilitySystem.getQueueSize()}`);
  
  // Queue a high-priority ability that interrupts Power Slam
  console.log('Queueing emergency heal with interrupt...');
  player.abilitySystem.queueAbility('heal', {
    owner: player,
    scene: null as any
  }, {
    priority: 3,
    interrupt: ['power_slam'], // This will interrupt any queued Power Slam
    replace: true // Replace existing heal
  });
  
  console.log(`After interrupt, queue size: ${player.abilitySystem.getQueueSize()}`);
  
  await player.abilitySystem.processAbilityQueue();
}

async function demonstrateExecutionModes(): Promise<void> {
  console.log('\n=== Execution Modes Example ===');
  
  const player = new ExamplePlayer('Paladin');
  
  // Manual mode - abilities don't auto-execute
  console.log('Setting to MANUAL mode...');
  player.abilitySystem.setQueueMode(QueueExecutionMode.MANUAL);
  
  player.queueQuickStrike(0);
  player.queuePowerSlam(1);
  player.queueHeal(2);
  
  console.log(`Queue size: ${player.abilitySystem.getQueueSize()}`);
  console.log('Abilities queued but not executing (manual mode)');
  
  // Manually execute one
  console.log('Manually executing one ability...');
  await player.abilitySystem.processAbilityQueue();
  console.log(`Queue size after manual execution: ${player.abilitySystem.getQueueSize()}`);
  
  // Switch to BATCH mode - execute all at once
  console.log('Switching to BATCH mode and executing all...');
  player.abilitySystem.setQueueMode(QueueExecutionMode.BATCH);
  await player.abilitySystem.processAbilityQueue();
  console.log(`Queue size after batch execution: ${player.abilitySystem.getQueueSize()}`);
}

async function demonstrateStats(): Promise<void> {
  console.log('\n=== Queue Statistics Example ===');
  
  const player = new ExamplePlayer('Scout');
  
  // Queue and execute multiple abilities
  for (let i = 0; i < 5; i++) {
    player.queueQuickStrike(Math.floor(Math.random() * 3));
    await new Promise(resolve => setTimeout(resolve, 100)); // Small delay between queues
  }
  
  await player.abilitySystem.processAbilityQueue();
  
  const stats = player.abilitySystem.getQueueStats();
  console.log('Queue Statistics:');
  console.log(`- Total Queued: ${stats.totalQueued}`);
  console.log(`- Total Executed: ${stats.totalExecuted}`);
  console.log(`- Total Cancelled: ${stats.totalCancelled}`);
  console.log(`- Average Wait Time: ${stats.averageWaitTime.toFixed(2)}ms`);
  console.log(`- Current Queue Size: ${stats.currentQueueSize}`);
}

// === RUN EXAMPLES ===

async function runAllExamples(): Promise<void> {
  console.log('🎮 GAS v2 - Ability Queue System Examples\n');
  
  try {
    await demonstrateBasicQueueing();
    await demonstrateDelayedExecution();
    await demonstrateInterruption();
    await demonstrateExecutionModes();
    await demonstrateStats();
    
    console.log('\n✅ All examples completed successfully!');
  } catch (error) {
    console.error('❌ Error running examples:', error);
  }
}

// Export for use in other files
export {
  ExamplePlayer,
  QuickStrikeAbility,
  PowerSlamAbility,
  HealAbility,
  runAllExamples
};

// Run if this file is executed directly
if (require.main === module) {
  runAllExamples();
}