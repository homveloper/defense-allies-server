// Ability Queue System for GAS v2
// Manages ability execution order, priorities, and timing

import { AbilityContext, EnhancedAbilityContext } from '../types/AbilityTypes';
import { EnhancedAbilitySystemComponent } from './AbilitySystemComponent';

export interface QueuedAbility {
  readonly id: string;
  readonly abilityId: string;
  readonly context: EnhancedAbilityContext;
  readonly priority: number;
  readonly queueTime: number;
  readonly executeAfter?: number; // timestamp when it can execute
  readonly delay?: number; // delay in ms from queue time
  readonly metadata?: Record<string, any>;
}

export interface QueueOptions {
  priority?: number; // Higher = executed first (default: 0)
  delay?: number; // Delay in ms before execution (default: 0)
  replace?: boolean; // Replace existing queued ability with same ID (default: false)
  interrupt?: string[]; // Ability IDs to interrupt when this is queued
  metadata?: Record<string, any>;
}

export enum QueueExecutionMode {
  AUTO = 'auto',     // Automatically execute when possible
  MANUAL = 'manual', // Manual execution only
  BATCH = 'batch'    // Execute all at once
}

export interface QueueStats {
  totalQueued: number;
  totalExecuted: number;
  totalCancelled: number;
  totalInterrupted: number;
  averageWaitTime: number;
  currentQueueSize: number;
}

export class AbilityQueue {
  private queue: QueuedAbility[] = [];
  private isExecuting: boolean = false;
  private executionMode: QueueExecutionMode = QueueExecutionMode.AUTO;
  private stats: QueueStats;
  private nextId: number = 1;

  constructor(
    private asc: EnhancedAbilitySystemComponent,
    mode: QueueExecutionMode = QueueExecutionMode.AUTO
  ) {
    this.executionMode = mode;
    this.stats = {
      totalQueued: 0,
      totalExecuted: 0,
      totalCancelled: 0,
      totalInterrupted: 0,
      averageWaitTime: 0,
      currentQueueSize: 0
    };
  }

  // === CORE QUEUE OPERATIONS ===

  enqueue(
    abilityId: string,
    context: Omit<EnhancedAbilityContext, 'metadata'>,
    options: QueueOptions = {}
  ): string {
    const now = Date.now();
    const queueId = `queue_${this.nextId++}`;
    
    const queuedAbility: QueuedAbility = {
      id: queueId,
      abilityId,
      context: {
        ...context,
        metadata: {
          ...context.metadata,
          ...options.metadata,
          queueId,
          queueTime: now
        }
      },
      priority: options.priority ?? 0,
      queueTime: now,
      executeAfter: options.delay ? now + options.delay : now,
      delay: options.delay,
      metadata: options.metadata
    };

    // Handle replacements
    if (options.replace) {
      this.removeByAbilityId(abilityId);
    }

    // Handle interrupts
    if (options.interrupt && options.interrupt.length > 0) {
      options.interrupt.forEach(interruptId => {
        this.interruptByAbilityId(interruptId);
      });
    }

    // Insert with priority sorting
    this.insertSorted(queuedAbility);
    
    // Update stats
    this.stats.totalQueued++;
    this.stats.currentQueueSize = this.queue.length;

    // Emit event
    this.asc.emit('ability-queued', {
      queueId,
      abilityId,
      priority: queuedAbility.priority,
      queueSize: this.queue.length,
      timestamp: now
    });

    // Auto execute if enabled
    if (this.executionMode === QueueExecutionMode.AUTO) {
      this.processQueue();
    }

    return queueId;
  }

  dequeue(): QueuedAbility | null {
    if (this.queue.length === 0) return null;
    
    const now = Date.now();
    
    // Find first ability that can execute
    const executableIndex = this.queue.findIndex(item => 
      item.executeAfter <= now
    );
    
    if (executableIndex === -1) return null;
    
    const queuedAbility = this.queue.splice(executableIndex, 1)[0];
    this.stats.currentQueueSize = this.queue.length;
    
    return queuedAbility;
  }

  peek(): QueuedAbility | null {
    if (this.queue.length === 0) return null;
    
    const now = Date.now();
    return this.queue.find(item => item.executeAfter <= now) || null;
  }

  // === QUEUE MANAGEMENT ===

  cancel(queueId: string): boolean {
    const index = this.queue.findIndex(item => item.id === queueId);
    if (index === -1) return false;

    const cancelled = this.queue.splice(index, 1)[0];
    this.stats.totalCancelled++;
    this.stats.currentQueueSize = this.queue.length;

    this.asc.emit('ability-queue-cancelled', {
      queueId,
      abilityId: cancelled.abilityId,
      reason: 'manual-cancel',
      timestamp: Date.now()
    });

    return true;
  }

  cancelAll(): number {
    const cancelledCount = this.queue.length;
    const cancelled = [...this.queue];
    
    this.queue = [];
    this.stats.totalCancelled += cancelledCount;
    this.stats.currentQueueSize = 0;

    cancelled.forEach(item => {
      this.asc.emit('ability-queue-cancelled', {
        queueId: item.id,
        abilityId: item.abilityId,
        reason: 'cancel-all',
        timestamp: Date.now()
      });
    });

    return cancelledCount;
  }

  interrupt(queueId: string): boolean {
    const index = this.queue.findIndex(item => item.id === queueId);
    if (index === -1) return false;

    const interrupted = this.queue.splice(index, 1)[0];
    this.stats.totalInterrupted++;
    this.stats.currentQueueSize = this.queue.length;

    this.asc.emit('ability-queue-interrupted', {
      queueId,
      abilityId: interrupted.abilityId,
      timestamp: Date.now()
    });

    return true;
  }

  private removeByAbilityId(abilityId: string): number {
    const removed = this.queue.filter(item => item.abilityId === abilityId);
    this.queue = this.queue.filter(item => item.abilityId !== abilityId);
    
    this.stats.currentQueueSize = this.queue.length;
    return removed.length;
  }

  private interruptByAbilityId(abilityId: string): number {
    const interrupted = this.queue.filter(item => item.abilityId === abilityId);
    this.queue = this.queue.filter(item => item.abilityId !== abilityId);
    
    this.stats.totalInterrupted += interrupted.length;
    this.stats.currentQueueSize = this.queue.length;

    interrupted.forEach(item => {
      this.asc.emit('ability-queue-interrupted', {
        queueId: item.id,
        abilityId: item.abilityId,
        timestamp: Date.now()
      });
    });

    return interrupted.length;
  }

  // === EXECUTION ===

  async execute(): Promise<number> {
    if (this.isExecuting) return 0;

    this.isExecuting = true;
    let executedCount = 0;

    try {
      while (this.queue.length > 0) {
        const queuedAbility = this.dequeue();
        if (!queuedAbility) break;

        const success = await this.executeQueuedAbility(queuedAbility);
        if (success) {
          executedCount++;
        }

        // Break if we're in auto mode (execute one at a time)
        if (this.executionMode === QueueExecutionMode.AUTO) {
          break;
        }
      }
    } finally {
      this.isExecuting = false;
    }

    return executedCount;
  }

  async processQueue(): Promise<void> {
    if (this.isExecuting || this.executionMode === QueueExecutionMode.MANUAL) {
      return;
    }

    await this.execute();
  }

  private async executeQueuedAbility(queuedAbility: QueuedAbility): Promise<boolean> {
    const waitTime = Date.now() - queuedAbility.queueTime;
    
    try {
      // Try to activate the ability
      const result = await this.asc.tryActivateAbility(
        queuedAbility.abilityId,
        queuedAbility.context
      );

      if (result.success) {
        this.stats.totalExecuted++;
        this.updateAverageWaitTime(waitTime);

        this.asc.emit('ability-queue-executed', {
          queueId: queuedAbility.id,
          abilityId: queuedAbility.abilityId,
          waitTime,
          timestamp: Date.now()
        });

        return true;
      } else {
        this.asc.emit('ability-queue-failed', {
          queueId: queuedAbility.id,
          abilityId: queuedAbility.abilityId,
          reason: result.failureReason || 'Unknown error',
          waitTime,
          timestamp: Date.now()
        });

        return false;
      }
    } catch (error) {
      console.error(`Error executing queued ability ${queuedAbility.abilityId}:`, error);
      
      this.asc.emit('ability-queue-failed', {
        queueId: queuedAbility.id,
        abilityId: queuedAbility.abilityId,
        reason: 'Execution error',
        waitTime,
        timestamp: Date.now()
      });

      return false;
    }
  }

  // === UTILITY METHODS ===

  private insertSorted(queuedAbility: QueuedAbility): void {
    // Sort by priority (higher first), then by queue time (earlier first)
    let insertIndex = 0;
    
    for (let i = 0; i < this.queue.length; i++) {
      const existing = this.queue[i];
      
      if (queuedAbility.priority > existing.priority) {
        insertIndex = i;
        break;
      } else if (queuedAbility.priority === existing.priority) {
        if (queuedAbility.queueTime < existing.queueTime) {
          insertIndex = i;
          break;
        }
      }
      
      insertIndex = i + 1;
    }
    
    this.queue.splice(insertIndex, 0, queuedAbility);
  }

  private updateAverageWaitTime(waitTime: number): void {
    const total = this.stats.averageWaitTime * (this.stats.totalExecuted - 1) + waitTime;
    this.stats.averageWaitTime = total / this.stats.totalExecuted;
  }

  // === GETTERS ===

  getMode(): QueueExecutionMode {
    return this.executionMode;
  }

  setMode(mode: QueueExecutionMode): void {
    this.executionMode = mode;
  }

  getSize(): number {
    return this.queue.length;
  }

  isEmpty(): boolean {
    return this.queue.length === 0;
  }

  getStats(): QueueStats {
    return { ...this.stats };
  }

  getQueuedAbilities(): QueuedAbility[] {
    return [...this.queue];
  }

  hasQueuedAbility(abilityId: string): boolean {
    return this.queue.some(item => item.abilityId === abilityId);
  }

  getQueuedAbility(queueId: string): QueuedAbility | null {
    return this.queue.find(item => item.id === queueId) || null;
  }

  // === CLEANUP ===

  clear(): void {
    this.cancelAll();
  }

  destroy(): void {
    this.clear();
    this.isExecuting = false;
  }
}