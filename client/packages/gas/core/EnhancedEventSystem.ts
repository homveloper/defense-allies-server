import { 
  AbilitySystemEvents, 
  AbilitySystemEventHandler, 
  EventListener,
  EventPriority 
} from '../types/AbilityTypes';

/**
 * Enhanced Event System with priority, filtering, and advanced features
 */
export class EnhancedEventSystem {
  private listeners: Map<keyof AbilitySystemEvents, EventListener<any>[]> = new Map();
  private eventHistory: Array<{
    event: keyof AbilitySystemEvents;
    data: any;
    timestamp: number;
  }> = [];
  private maxHistorySize: number = 1000;
  private paused: boolean = false;

  /**
   * Add event listener with priority and optional filtering
   */
  on<T extends keyof AbilitySystemEvents>(
    event: T,
    handler: AbilitySystemEventHandler<T>,
    options: {
      priority?: EventPriority;
      once?: boolean;
      filter?: (data: AbilitySystemEvents[T]) => boolean;
    } = {}
  ): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }

    const listener: EventListener<T> = {
      handler,
      priority: options.priority || EventPriority.NORMAL,
      once: options.once || false,
      filter: options.filter
    };

    const eventListeners = this.listeners.get(event)!;
    eventListeners.push(listener);

    // Sort by priority (lower number = higher priority)
    eventListeners.sort((a, b) => a.priority - b.priority);
  }

  /**
   * Add one-time event listener
   */
  once<T extends keyof AbilitySystemEvents>(
    event: T,
    handler: AbilitySystemEventHandler<T>,
    options: {
      priority?: EventPriority;
      filter?: (data: AbilitySystemEvents[T]) => boolean;
    } = {}
  ): void {
    this.on(event, handler, { ...options, once: true });
  }

  /**
   * Remove specific event listener
   */
  off<T extends keyof AbilitySystemEvents>(
    event: T,
    handler: AbilitySystemEventHandler<T>
  ): void {
    const eventListeners = this.listeners.get(event);
    if (!eventListeners) return;

    const index = eventListeners.findIndex(listener => listener.handler === handler);
    if (index !== -1) {
      eventListeners.splice(index, 1);
    }
  }

  /**
   * Remove all listeners for an event
   */
  removeAllListeners<T extends keyof AbilitySystemEvents>(event?: T): void {
    if (event) {
      this.listeners.delete(event);
    } else {
      this.listeners.clear();
    }
  }

  /**
   * Emit event to all listeners
   */
  emit<T extends keyof AbilitySystemEvents>(
    event: T,
    data: AbilitySystemEvents[T]
  ): void {
    if (this.paused) return;

    // Add timestamp if not present
    const eventData = {
      ...data,
      timestamp: data.timestamp || Date.now()
    } as AbilitySystemEvents[T];

    // Store in history
    this.addToHistory(event, eventData);

    const eventListeners = this.listeners.get(event);
    if (!eventListeners) return;

    // Process listeners in priority order
    const listenersToRemove: number[] = [];

    for (let i = 0; i < eventListeners.length; i++) {
      const listener = eventListeners[i];

      try {
        // Check filter condition
        if (listener.filter && !listener.filter(eventData)) {
          continue;
        }

        // Execute handler
        listener.handler(eventData);

        // Mark for removal if it's a one-time listener
        if (listener.once) {
          listenersToRemove.push(i);
        }
      } catch (error) {
        console.error(`Error in event handler for '${event}':`, error);
      }
    }

    // Remove one-time listeners (in reverse order to maintain indices)
    listenersToRemove.reverse().forEach(index => {
      eventListeners.splice(index, 1);
    });
  }

  /**
   * Add event to history
   */
  private addToHistory<T extends keyof AbilitySystemEvents>(
    event: T,
    data: AbilitySystemEvents[T]
  ): void {
    this.eventHistory.push({
      event,
      data,
      timestamp: Date.now()
    });

    // Limit history size
    if (this.eventHistory.length > this.maxHistorySize) {
      this.eventHistory.shift();
    }
  }

  /**
   * Get event history
   */
  getEventHistory(
    options: {
      eventType?: keyof AbilitySystemEvents;
      since?: number; // timestamp
      limit?: number;
    } = {}
  ): Array<{ event: keyof AbilitySystemEvents; data: any; timestamp: number }> {
    let history = this.eventHistory;

    // Filter by event type
    if (options.eventType) {
      history = history.filter(entry => entry.event === options.eventType);
    }

    // Filter by timestamp
    if (options.since) {
      history = history.filter(entry => entry.timestamp >= options.since!);
    }

    // Limit results
    if (options.limit) {
      history = history.slice(-options.limit);
    }

    return history;
  }

  /**
   * Clear event history
   */
  clearHistory(): void {
    this.eventHistory = [];
  }

  /**
   * Pause/resume event processing
   */
  setPaused(paused: boolean): void {
    this.paused = paused;
  }

  /**
   * Check if event system is paused
   */
  isPaused(): boolean {
    return this.paused;
  }

  /**
   * Get listener count for an event
   */
  getListenerCount<T extends keyof AbilitySystemEvents>(event: T): number {
    return this.listeners.get(event)?.length || 0;
  }

  /**
   * Get all registered events
   */
  getRegisteredEvents(): Array<keyof AbilitySystemEvents> {
    return Array.from(this.listeners.keys());
  }

  /**
   * Create event filter functions for common patterns
   */
  static createFilters() {
    return {
      // Filter by source/owner
      byOwner: (owner: any) => 
        (data: any) => data.owner === owner || data.target === owner,

      // Filter by ability ID
      byAbility: (abilityId: string) => 
        (data: any) => data.abilityId === abilityId,

      // Filter by effect ID
      byEffect: (effectId: string) => 
        (data: any) => data.effectId === effectId,

      // Filter by attribute
      byAttribute: (attribute: string) => 
        (data: any) => data.attribute === attribute,

      // Filter by tag
      byTag: (tag: string) => 
        (data: any) => data.tag === tag,

      // Filter by time range
      byTimeRange: (start: number, end: number) => 
        (data: any) => data.timestamp >= start && data.timestamp <= end,

      // Filter by value threshold
      byValueThreshold: (property: string, threshold: number, operator: '>' | '<' | '>=' | '<=') => 
        (data: any) => {
          const value = data[property];
          if (typeof value !== 'number') return false;
          switch (operator) {
            case '>': return value > threshold;
            case '<': return value < threshold;
            case '>=': return value >= threshold;
            case '<=': return value <= threshold;
            default: return false;
          }
        },

      // Combine multiple filters with AND logic
      and: (...filters: Array<(data: any) => boolean>) => 
        (data: any) => filters.every(filter => filter(data)),

      // Combine multiple filters with OR logic
      or: (...filters: Array<(data: any) => boolean>) => 
        (data: any) => filters.some(filter => filter(data)),

      // Negate a filter
      not: (filter: (data: any) => boolean) => 
        (data: any) => !filter(data)
    };
  }

  /**
   * Debugging and monitoring
   */
  getDebugInfo(): {
    totalListeners: number;
    eventTypes: Array<{ event: keyof AbilitySystemEvents; listenerCount: number }>;
    historySize: number;
    isPaused: boolean;
    recentEvents: Array<{ event: keyof AbilitySystemEvents; timestamp: number }>;
  } {
    return {
      totalListeners: Array.from(this.listeners.values())
        .reduce((total, listeners) => total + listeners.length, 0),
      eventTypes: Array.from(this.listeners.entries())
        .map(([event, listeners]) => ({ event, listenerCount: listeners.length })),
      historySize: this.eventHistory.length,
      isPaused: this.paused,
      recentEvents: this.eventHistory.slice(-10).map(entry => ({
        event: entry.event,
        timestamp: entry.timestamp
      }))
    };
  }
}