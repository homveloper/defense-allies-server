// Tactical Arena Engine - Core game engine for tactical combat
// Integrates GAS v2 turn-based systems with tactical combat mechanics

import { v2 } from '../../../../packages/gas';
import { TurnPhase, TurnBasedAbilityContext } from '../../../../packages/gas/v2/turn-based/TurnBasedContext';
import { TacticalUnit, TacticalUnitStats, Position, UnitFaction } from '../entities/TacticalUnit';
import { TacticalMap, MapConfig, TerrainType } from '../entities/TacticalMap';

export interface EngineConfig {
  width: number;
  height: number;
  mapSize: { width: number; height: number };
  playerUnits: number;
  enemyUnits: number;
  tileSize?: number;
}

export interface GameState {
  currentTurn: number;
  currentRound: number;
  activeUnit: TacticalUnit | null;
  currentPhase: TurnPhase;
  turnOrder: TacticalUnit[];
  selectedUnit: TacticalUnit | null;
  gameResult: 'player_victory' | 'enemy_victory' | 'draw' | null;
  isGameOver: boolean;
}

type GameEvent = 
  | 'game-state-changed'
  | 'ui-state-changed'
  | 'unit-selected'
  | 'action-completed'
  | 'turn-started'
  | 'turn-ended'
  | 'phase-changed'
  | 'game-ended';

export class TacticalArenaEngine {
  private canvas: HTMLCanvasElement;
  private ctx: CanvasRenderingContext2D;
  private config: EngineConfig;
  
  // Game Systems
  private map: TacticalMap;
  private phaseManager: v2.PhaseManager;
  private turnOrderManager: v2.TurnOrderManager;
  private initiativeCalculator: v2.InitiativeCalculator;
  
  // Game State
  private gameState: GameState;
  private eventListeners: Map<GameEvent, Array<(data: any) => void>> = new Map();
  
  // UI State
  private selectedTile: Position | null = null;
  private hoveredTile: Position | null = null;
  private highlightedTiles: Position[] = [];
  
  // Animation
  private animationId: number | null = null;
  private lastUpdateTime: number = 0;

  constructor(canvas: HTMLCanvasElement, config: EngineConfig) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d')!;
    this.config = config;
    
    // Initialize map
    const mapConfig: MapConfig = {
      width: config.mapSize.width,
      height: config.mapSize.height,
      tileSize: config.tileSize || 64,
      defaultTerrain: TerrainType.FLOOR
    };
    this.map = new TacticalMap(mapConfig);
    
    // Initialize turn systems
    this.phaseManager = new v2.PhaseManager();
    this.turnOrderManager = new v2.TurnOrderManager();
    this.initiativeCalculator = new v2.InitiativeCalculator(
      v2.InitiativeRollType.ONCE_PER_ROUND,
      false // Use fixed initiative for tactical combat
    );
    
    // Initialize game state
    this.gameState = {
      currentTurn: 1,
      currentRound: 1,
      activeUnit: null,
      currentPhase: TurnPhase.START,
      turnOrder: [],
      selectedUnit: null,
      gameResult: null,
      isGameOver: false
    };
    
    // Set up event listeners
    this.setupEventListeners();
  }

  // === INITIALIZATION ===

  initialize(): void {
    this.createUnits();
    this.setupTurnOrder();
    this.startGame();
    this.startRenderLoop();
  }

  private createUnits(): void {
    // Create player units
    const playerStartPositions = [
      { x: 0, y: 0 }, { x: 0, y: 1 }, { x: 0, y: 2 }, { x: 0, y: 3 }
    ];
    
    for (let i = 0; i < this.config.playerUnits; i++) {
      const stats: TacticalUnitStats = {
        maxHealth: 100,
        health: 100,
        armor: 2,
        accuracy: 75,
        movement: 3,
        initiative: 15 + Math.random() * 10,
        cover: 0
      };
      
      const unit = new TacticalUnit(
        `player_${i}`,
        `Player ${i + 1}`,
        'player',
        playerStartPositions[i] || { x: 0, y: i },
        stats
      );
      
      this.map.addUnit(unit, unit.position);
    }
    
    // Create enemy units
    const enemyStartPositions = [
      { x: this.config.mapSize.width - 1, y: 0 },
      { x: this.config.mapSize.width - 1, y: 1 },
      { x: this.config.mapSize.width - 1, y: 2 },
      { x: this.config.mapSize.width - 1, y: 3 }
    ];
    
    for (let i = 0; i < this.config.enemyUnits; i++) {
      const stats: TacticalUnitStats = {
        maxHealth: 80,
        health: 80,
        armor: 1,
        accuracy: 65,
        movement: 3,
        initiative: 10 + Math.random() * 10,
        cover: 0
      };
      
      const unit = new TacticalUnit(
        `enemy_${i}`,
        `Enemy ${i + 1}`,
        'enemy',
        enemyStartPositions[i] || { x: this.config.mapSize.width - 1, y: i },
        stats
      );
      
      this.map.addUnit(unit, unit.position);
    }
  }

  private setupTurnOrder(): void {
    const allUnits = this.map.getAllUnits();
    
    // Calculate initiative for all units
    const turnEntries = allUnits.map(unit => ({
      entityId: unit.id,
      initiative: unit.stats.initiative,
      delay: 0
    }));
    
    // Set turn order
    this.turnOrderManager.setTurnOrder(turnEntries);
    this.gameState.turnOrder = allUnits.sort((a, b) => b.stats.initiative - a.stats.initiative);
  }

  private startGame(): void {
    this.phaseManager.startTurn(TurnPhase.START);
    this.gameState.currentPhase = TurnPhase.START;
    this.gameState.activeUnit = this.gameState.turnOrder[0] || null;
    
    if (this.gameState.activeUnit) {
      this.gameState.activeUnit.startTurn();
    }
    
    this.emit('game-state-changed', this.gameState);
    this.emit('turn-started', { unit: this.gameState.activeUnit });
  }

  // === EVENT SYSTEM ===

  on(event: GameEvent, callback: (data: any) => void): void {
    if (!this.eventListeners.has(event)) {
      this.eventListeners.set(event, []);
    }
    this.eventListeners.get(event)!.push(callback);
  }

  private emit(event: GameEvent, data: any): void {
    const listeners = this.eventListeners.get(event);
    if (listeners) {
      listeners.forEach(callback => callback(data));
    }
  }

  // === GAME ACTIONS ===

  async executeAction(unitId: string, actionId: string, target?: TacticalUnit | Position): Promise<boolean> {
    const unit = this.map.getUnit(unitId);
    if (!unit || unit !== this.gameState.activeUnit) {
      return false;
    }

    const context: TurnBasedAbilityContext = {
      owner: unit,
      target,
      scene: this.map,
      currentTurn: this.gameState.currentTurn,
      currentRound: this.gameState.currentRound,
      activePlayer: unit.id,
      phase: this.gameState.currentPhase
    };

    const success = await unit.executeAction(actionId, target);
    
    this.emit('action-completed', { unitId, actionId, success });
    
    if (success) {
      this.checkGameEnd();
      this.emit('game-state-changed', this.gameState);
    }
    
    return success;
  }

  selectUnit(unitId: string): boolean {
    const unit = this.map.getUnit(unitId);
    if (!unit) return false;
    
    this.gameState.selectedUnit = unit;
    this.selectedTile = unit.position;
    
    // Highlight valid moves for active unit
    if (unit === this.gameState.activeUnit) {
      const resources = unit.getResourceSummary();
      const movementPoints = resources.movement_points?.current || 0;
      const validMoves = this.map.getValidMoves(unit.position, movementPoints);
      this.highlightedTiles = validMoves;
      this.map.highlightTiles(validMoves);
    }
    
    this.emit('unit-selected', unit);
    return true;
  }

  advancePhase(): boolean {
    const nextPhase = this.phaseManager.advancePhase();
    if (nextPhase) {
      this.gameState.currentPhase = nextPhase;
      this.emit('phase-changed', { phase: nextPhase });
      this.emit('game-state-changed', this.gameState);
      return true;
    }
    return false;
  }

  endCurrentTurn(): void {
    if (this.gameState.activeUnit) {
      this.gameState.activeUnit.endTurn();
    }
    
    this.phaseManager.endTurn();
    
    // Move to next unit
    const currentIndex = this.gameState.turnOrder.findIndex(u => u === this.gameState.activeUnit);
    const nextIndex = (currentIndex + 1) % this.gameState.turnOrder.length;
    
    if (nextIndex === 0) {
      // New round
      this.gameState.currentRound++;
      this.processNewRound();
    }
    
    this.gameState.currentTurn++;
    this.gameState.activeUnit = this.gameState.turnOrder[nextIndex];
    this.gameState.currentPhase = TurnPhase.START;
    
    if (this.gameState.activeUnit) {
      this.gameState.activeUnit.startTurn();
      this.phaseManager.startTurn(TurnPhase.START);
    }
    
    // Auto-select active unit if it's a player unit
    if (this.gameState.activeUnit?.faction === 'player') {
      this.selectUnit(this.gameState.activeUnit.id);
    }
    
    this.emit('turn-ended', { previousUnit: this.gameState.turnOrder[currentIndex] });
    this.emit('turn-started', { unit: this.gameState.activeUnit });
    this.emit('game-state-changed', this.gameState);
  }

  private processNewRound(): void {
    // Process all units for new round
    for (const unit of this.gameState.turnOrder) {
      // Reset turn-specific states
      unit.hasMovedThisTurn = false;
      unit.overwatchActive = false;
      
      // Process status effects
      unit.processStatusEffects();
    }
  }

  restart(): void {
    this.gameState.gameResult = null;
    this.gameState.isGameOver = false;
    this.gameState.currentTurn = 1;
    this.gameState.currentRound = 1;
    
    // Reset all units
    for (const unit of this.map.getAllUnits()) {
      unit.stats.health = unit.stats.maxHealth;
      unit.startTurn();
    }
    
    this.setupTurnOrder();
    this.startGame();
  }

  // === GAME LOGIC ===

  private checkGameEnd(): void {
    const playerUnits = this.map.getAllUnits().filter(u => u.faction === 'player' && u.isAlive());
    const enemyUnits = this.map.getAllUnits().filter(u => u.faction === 'enemy' && u.isAlive());
    
    if (playerUnits.length === 0) {
      this.gameState.gameResult = 'enemy_victory';
      this.gameState.isGameOver = true;
      this.emit('game-ended', 'enemy_victory');
    } else if (enemyUnits.length === 0) {
      this.gameState.gameResult = 'player_victory';
      this.gameState.isGameOver = true;
      this.emit('game-ended', 'player_victory');
    }
    
    // Remove dead units from turn order
    this.gameState.turnOrder = this.gameState.turnOrder.filter(u => u.isAlive());
  }

  // === INPUT HANDLING ===

  private setupEventListeners(): void {
    this.canvas.addEventListener('click', this.handleClick.bind(this));
    this.canvas.addEventListener('mousemove', this.handleMouseMove.bind(this));
  }

  private handleClick(event: MouseEvent): void {
    const rect = this.canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    
    const tilePos = this.screenToTile(x, y);
    if (!this.map.isValidPosition(tilePos)) return;
    
    const unit = this.map.getUnitAt(tilePos);
    
    if (unit) {
      // Select unit
      this.selectUnit(unit.id);
    } else if (this.gameState.selectedUnit && this.gameState.selectedUnit === this.gameState.activeUnit) {
      // Try to move active unit
      const resources = this.gameState.selectedUnit.getResourceSummary();
      const movementPoints = resources.movement_points?.current || 0;
      
      if (movementPoints > 0) {
        const cost = this.map.getMovementCost(this.gameState.selectedUnit.position, tilePos);
        if (cost <= movementPoints && this.map.isPassable(tilePos)) {
          this.executeAction(this.gameState.selectedUnit.id, 'move', tilePos);
        }
      }
    }
  }

  private handleMouseMove(event: MouseEvent): void {
    const rect = this.canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    
    const tilePos = this.screenToTile(x, y);
    if (this.map.isValidPosition(tilePos)) {
      this.hoveredTile = tilePos;
    } else {
      this.hoveredTile = null;
    }
  }

  // === COORDINATE CONVERSION ===

  private screenToTile(screenX: number, screenY: number): Position {
    const tileSize = this.map.config.tileSize;
    return {
      x: Math.floor(screenX / tileSize),
      y: Math.floor(screenY / tileSize)
    };
  }

  private tileToScreen(tileX: number, tileY: number): Position {
    const tileSize = this.map.config.tileSize;
    return {
      x: tileX * tileSize,
      y: tileY * tileSize
    };
  }

  // === RENDERING ===

  private startRenderLoop(): void {
    const render = (timestamp: number) => {
      this.update(timestamp);
      this.render();
      this.animationId = requestAnimationFrame(render);
    };
    this.animationId = requestAnimationFrame(render);
  }

  private update(timestamp: number): void {
    const deltaTime = timestamp - this.lastUpdateTime;
    this.lastUpdateTime = timestamp;
    
    // Update phase manager for auto-advancement
    this.phaseManager.update();
  }

  private render(): void {
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    
    this.renderMap();
    this.renderUnits();
    this.renderUI();
  }

  private renderMap(): void {
    const tileSize = this.map.config.tileSize;
    
    for (let y = 0; y < this.map.config.height; y++) {
      for (let x = 0; x < this.map.config.width; x++) {
        const tile = this.map.getTile({ x, y });
        if (!tile) continue;
        
        const screenPos = this.tileToScreen(x, y);
        
        // Draw tile background
        this.ctx.fillStyle = this.getTileColor(tile);
        this.ctx.fillRect(screenPos.x, screenPos.y, tileSize, tileSize);
        
        // Draw tile border
        this.ctx.strokeStyle = '#444';
        this.ctx.strokeRect(screenPos.x, screenPos.y, tileSize, tileSize);
        
        // Highlight if selected or hovered
        if (this.selectedTile && x === this.selectedTile.x && y === this.selectedTile.y) {
          this.ctx.fillStyle = 'rgba(255, 255, 0, 0.3)';
          this.ctx.fillRect(screenPos.x, screenPos.y, tileSize, tileSize);
        }
        
        if (this.hoveredTile && x === this.hoveredTile.x && y === this.hoveredTile.y) {
          this.ctx.fillStyle = 'rgba(255, 255, 255, 0.2)';
          this.ctx.fillRect(screenPos.x, screenPos.y, tileSize, tileSize);
        }
        
        if (tile.isHighlighted) {
          this.ctx.fillStyle = 'rgba(0, 255, 0, 0.3)';
          this.ctx.fillRect(screenPos.x, screenPos.y, tileSize, tileSize);
        }
      }
    }
  }

  private getTileColor(tile: any): string {
    switch (tile.type) {
      case TerrainType.FLOOR: return '#2a2a2a';
      case TerrainType.COVER: return '#4a4a4a';
      case TerrainType.DIFFICULT: return '#6a4a2a';
      case TerrainType.WALL: return '#1a1a1a';
      default: return '#2a2a2a';
    }
  }

  private renderUnits(): void {
    const tileSize = this.map.config.tileSize;
    
    for (const unit of this.map.getAllUnits()) {
      if (!unit.isAlive()) continue;
      
      const screenPos = this.tileToScreen(unit.position.x, unit.position.y);
      const centerX = screenPos.x + tileSize / 2;
      const centerY = screenPos.y + tileSize / 2;
      const radius = unit.size / 2;
      
      // Draw unit circle
      this.ctx.beginPath();
      this.ctx.arc(centerX, centerY, radius, 0, 2 * Math.PI);
      this.ctx.fillStyle = unit.color;
      this.ctx.fill();
      
      // Draw selection ring
      if (unit === this.gameState.selectedUnit) {
        this.ctx.beginPath();
        this.ctx.arc(centerX, centerY, radius + 3, 0, 2 * Math.PI);
        this.ctx.strokeStyle = '#fff';
        this.ctx.lineWidth = 2;
        this.ctx.stroke();
      }
      
      // Draw active unit indicator
      if (unit === this.gameState.activeUnit) {
        this.ctx.beginPath();
        this.ctx.arc(centerX, centerY, radius + 6, 0, 2 * Math.PI);
        this.ctx.strokeStyle = '#ffff00';
        this.ctx.lineWidth = 3;
        this.ctx.stroke();
      }
      
      // Draw health bar
      const healthBarWidth = tileSize - 10;
      const healthBarHeight = 4;
      const healthBarY = screenPos.y + tileSize - 8;
      const healthPercent = unit.stats.health / unit.stats.maxHealth;
      
      this.ctx.fillStyle = '#333';
      this.ctx.fillRect(screenPos.x + 5, healthBarY, healthBarWidth, healthBarHeight);
      
      this.ctx.fillStyle = healthPercent > 0.5 ? '#0f0' : healthPercent > 0.25 ? '#ff0' : '#f00';
      this.ctx.fillRect(screenPos.x + 5, healthBarY, healthBarWidth * healthPercent, healthBarHeight);
    }
  }

  private renderUI(): void {
    // Render any additional UI elements directly on canvas if needed
    // Most UI is handled by React components
  }

  // === CLEANUP ===

  destroy(): void {
    if (this.animationId) {
      cancelAnimationFrame(this.animationId);
    }
    
    this.eventListeners.clear();
  }

  // === GETTERS ===

  getGameState(): GameState {
    return { ...this.gameState };
  }

  getMap(): TacticalMap {
    return this.map;
  }
}