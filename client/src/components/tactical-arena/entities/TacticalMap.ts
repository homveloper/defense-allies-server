// Tactical Map - Grid-based tactical combat map
// Handles terrain, line-of-sight, movement validation, and tactical positioning

import { Position, TacticalUnit } from './TacticalUnit';

export interface Tile {
  x: number;
  y: number;
  type: TerrainType;
  elevation: number;
  cover: CoverType;
  occupied: boolean;
  occupant?: TacticalUnit;
  isVisible: boolean;
  isHighlighted: boolean;
  movementCost: number;
}

export enum TerrainType {
  FLOOR = 'floor',
  WALL = 'wall',
  COVER = 'cover',
  WATER = 'water',
  DIFFICULT = 'difficult'
}

export enum CoverType {
  NONE = 'none',
  PARTIAL = 'partial',  // 25% cover bonus
  FULL = 'full'         // 50% cover bonus
}

export interface MapConfig {
  width: number;
  height: number;
  tileSize: number;
  defaultTerrain: TerrainType;
}

export class TacticalMap {
  public readonly config: MapConfig;
  private tiles: Tile[][];
  private units: Map<string, TacticalUnit> = new Map();
  
  constructor(config: MapConfig) {
    this.config = config;
    this.tiles = this.generateTiles();
    this.generateRandomTerrain();
  }

  // === MAP GENERATION ===

  private generateTiles(): Tile[][] {
    const tiles: Tile[][] = [];
    
    for (let y = 0; y < this.config.height; y++) {
      tiles[y] = [];
      for (let x = 0; x < this.config.width; x++) {
        tiles[y][x] = {
          x,
          y,
          type: this.config.defaultTerrain,
          elevation: 0,
          cover: CoverType.NONE,
          occupied: false,
          isVisible: true,
          isHighlighted: false,
          movementCost: 1
        };
      }
    }
    
    return tiles;
  }

  private generateRandomTerrain(): void {
    // Add some cover positions strategically
    const coverPositions = [
      { x: 2, y: 1 }, { x: 5, y: 1 },
      { x: 1, y: 3 }, { x: 6, y: 3 },
      { x: 3, y: 4 }, { x: 4, y: 4 }
    ];

    for (const pos of coverPositions) {
      if (this.isValidPosition(pos)) {
        const tile = this.tiles[pos.y][pos.x];
        tile.type = TerrainType.COVER;
        tile.cover = Math.random() > 0.5 ? CoverType.FULL : CoverType.PARTIAL;
        tile.movementCost = 1;
      }
    }

    // Add some difficult terrain
    const difficultPositions = [
      { x: 0, y: 2 }, { x: 7, y: 2 },
      { x: 2, y: 5 }, { x: 5, y: 5 }
    ];

    for (const pos of difficultPositions) {
      if (this.isValidPosition(pos)) {
        const tile = this.tiles[pos.y][pos.x];
        tile.type = TerrainType.DIFFICULT;
        tile.movementCost = 2;
      }
    }
  }

  // === UNIT MANAGEMENT ===

  addUnit(unit: TacticalUnit, position: Position): boolean {
    if (!this.isValidPosition(position) || this.isOccupied(position)) {
      return false;
    }

    this.units.set(unit.id, unit);
    unit.position = { ...position };
    
    const tile = this.tiles[position.y][position.x];
    tile.occupied = true;
    tile.occupant = unit;
    
    return true;
  }

  removeUnit(unitId: string): boolean {
    const unit = this.units.get(unitId);
    if (!unit) return false;

    const tile = this.tiles[unit.position.y][unit.position.x];
    tile.occupied = false;
    tile.occupant = undefined;
    
    this.units.delete(unitId);
    return true;
  }

  moveUnit(unitId: string, newPosition: Position): boolean {
    const unit = this.units.get(unitId);
    if (!unit) return false;

    if (!this.isValidPosition(newPosition) || this.isOccupied(newPosition)) {
      return false;
    }

    // Clear old position
    const oldTile = this.tiles[unit.position.y][unit.position.x];
    oldTile.occupied = false;
    oldTile.occupant = undefined;

    // Set new position
    unit.position = { ...newPosition };
    const newTile = this.tiles[newPosition.y][newPosition.x];
    newTile.occupied = true;
    newTile.occupant = unit;

    return true;
  }

  // === PATHFINDING & MOVEMENT ===

  getValidMoves(from: Position, movementPoints: number): Position[] {
    const validMoves: Position[] = [];
    const visited = new Set<string>();
    const queue: { pos: Position; cost: number }[] = [{ pos: from, cost: 0 }];

    while (queue.length > 0) {
      const { pos, cost } = queue.shift()!;
      const key = `${pos.x},${pos.y}`;
      
      if (visited.has(key)) continue;
      visited.add(key);

      if (cost > 0 && cost <= movementPoints) {
        validMoves.push({ ...pos });
      }

      // Check adjacent tiles
      const directions = [
        { x: 0, y: -1 }, { x: 1, y: 0 }, { x: 0, y: 1 }, { x: -1, y: 0 }
      ];

      for (const dir of directions) {
        const nextPos = { x: pos.x + dir.x, y: pos.y + dir.y };
        const nextKey = `${nextPos.x},${nextPos.y}`;
        
        if (visited.has(nextKey) || !this.isValidPosition(nextPos)) continue;
        if (this.isOccupied(nextPos)) continue;

        const tile = this.tiles[nextPos.y][nextPos.x];
        const nextCost = cost + tile.movementCost;
        
        if (nextCost <= movementPoints) {
          queue.push({ pos: nextPos, cost: nextCost });
        }
      }
    }

    return validMoves;
  }

  getMovementCost(from: Position, to: Position): number {
    // Simple Manhattan distance with terrain costs
    if (!this.isValidPosition(from) || !this.isValidPosition(to)) {
      return Infinity;
    }

    const path = this.findPath(from, to);
    if (!path || path.length === 0) return Infinity;

    let totalCost = 0;
    for (let i = 1; i < path.length; i++) {
      const tile = this.tiles[path[i].y][path[i].x];
      totalCost += tile.movementCost;
    }

    return totalCost;
  }

  findPath(from: Position, to: Position): Position[] | null {
    // Simple A* pathfinding
    const openSet = [{ pos: from, gScore: 0, fScore: this.heuristic(from, to) }];
    const closedSet = new Set<string>();
    const cameFrom = new Map<string, Position>();
    const gScore = new Map<string, number>();
    
    gScore.set(`${from.x},${from.y}`, 0);

    while (openSet.length > 0) {
      // Get node with lowest f score
      openSet.sort((a, b) => a.fScore - b.fScore);
      const current = openSet.shift()!;
      const currentKey = `${current.pos.x},${current.pos.y}`;

      if (current.pos.x === to.x && current.pos.y === to.y) {
        // Reconstruct path
        const path: Position[] = [];
        let currentPos = current.pos;
        
        while (currentPos) {
          path.unshift(currentPos);
          const key = `${currentPos.x},${currentPos.y}`;
          currentPos = cameFrom.get(key)!;
        }
        
        return path;
      }

      closedSet.add(currentKey);

      // Check neighbors
      const directions = [
        { x: 0, y: -1 }, { x: 1, y: 0 }, { x: 0, y: 1 }, { x: -1, y: 0 }
      ];

      for (const dir of directions) {
        const neighbor = { x: current.pos.x + dir.x, y: current.pos.y + dir.y };
        const neighborKey = `${neighbor.x},${neighbor.y}`;
        
        if (closedSet.has(neighborKey) || !this.isValidPosition(neighbor)) continue;
        if (this.isOccupied(neighbor) && !(neighbor.x === to.x && neighbor.y === to.y)) continue;

        const tile = this.tiles[neighbor.y][neighbor.x];
        const tentativeGScore = (gScore.get(currentKey) || 0) + tile.movementCost;
        
        const existingGScore = gScore.get(neighborKey) || Infinity;
        if (tentativeGScore < existingGScore) {
          cameFrom.set(neighborKey, current.pos);
          gScore.set(neighborKey, tentativeGScore);
          
          const fScore = tentativeGScore + this.heuristic(neighbor, to);
          
          if (!openSet.find(n => n.pos.x === neighbor.x && n.pos.y === neighbor.y)) {
            openSet.push({ pos: neighbor, gScore: tentativeGScore, fScore });
          }
        }
      }
    }

    return null; // No path found
  }

  private heuristic(a: Position, b: Position): number {
    return Math.abs(a.x - b.x) + Math.abs(a.y - b.y);
  }

  // === LINE OF SIGHT ===

  hasLineOfSight(from: Position, to: Position): boolean {
    if (!this.isValidPosition(from) || !this.isValidPosition(to)) {
      return false;
    }

    // Bresenham's line algorithm
    const dx = Math.abs(to.x - from.x);
    const dy = Math.abs(to.y - from.y);
    const sx = from.x < to.x ? 1 : -1;
    const sy = from.y < to.y ? 1 : -1;
    let err = dx - dy;

    let x = from.x;
    let y = from.y;

    while (true) {
      // Check if current position blocks line of sight
      if (x !== from.x || y !== from.y) { // Skip starting position
        if (!this.isValidPosition({ x, y })) return false;
        
        const tile = this.tiles[y][x];
        if (tile.type === TerrainType.WALL) return false;
      }

      // Reached target
      if (x === to.x && y === to.y) break;

      const e2 = 2 * err;
      if (e2 > -dy) {
        err -= dy;
        x += sx;
      }
      if (e2 < dx) {
        err += dx;
        y += sy;
      }
    }

    return true;
  }

  getVisiblePositions(from: Position, range: number): Position[] {
    const visible: Position[] = [];
    
    for (let y = Math.max(0, from.y - range); y <= Math.min(this.config.height - 1, from.y + range); y++) {
      for (let x = Math.max(0, from.x - range); x <= Math.min(this.config.width - 1, from.x + range); x++) {
        const pos = { x, y };
        const distance = this.heuristic(from, pos);
        
        if (distance <= range && this.hasLineOfSight(from, pos)) {
          visible.push(pos);
        }
      }
    }
    
    return visible;
  }

  // === COVER SYSTEM ===

  getCoverBonus(position: Position): number {
    if (!this.isValidPosition(position)) return 0;
    
    const tile = this.tiles[position.y][position.x];
    
    switch (tile.cover) {
      case CoverType.PARTIAL: return 25;
      case CoverType.FULL: return 50;
      default: return 0;
    }
  }

  isInCover(position: Position): boolean {
    if (!this.isValidPosition(position)) return false;
    
    const tile = this.tiles[position.y][position.x];
    return tile.cover !== CoverType.NONE;
  }

  // === VALIDATION ===

  isValidPosition(position: Position): boolean {
    return position.x >= 0 && position.x < this.config.width &&
           position.y >= 0 && position.y < this.config.height;
  }

  isOccupied(position: Position): boolean {
    if (!this.isValidPosition(position)) return false;
    return this.tiles[position.y][position.x].occupied;
  }

  isPassable(position: Position): boolean {
    if (!this.isValidPosition(position)) return false;
    const tile = this.tiles[position.y][position.x];
    return tile.type !== TerrainType.WALL && !tile.occupied;
  }

  // === GETTERS ===

  getTile(position: Position): Tile | null {
    if (!this.isValidPosition(position)) return null;
    return this.tiles[position.y][position.x];
  }

  getUnit(unitId: string): TacticalUnit | undefined {
    return this.units.get(unitId);
  }

  getUnitAt(position: Position): TacticalUnit | null {
    if (!this.isValidPosition(position)) return null;
    const tile = this.tiles[position.y][position.x];
    return tile.occupant || null;
  }

  getAllUnits(): TacticalUnit[] {
    return Array.from(this.units.values());
  }

  getUnitsInRange(center: Position, range: number): TacticalUnit[] {
    const unitsInRange: TacticalUnit[] = [];
    
    for (const unit of this.units.values()) {
      const distance = this.heuristic(center, unit.position);
      if (distance <= range) {
        unitsInRange.push(unit);
      }
    }
    
    return unitsInRange;
  }

  // === UTILITY ===

  clearHighlights(): void {
    for (let y = 0; y < this.config.height; y++) {
      for (let x = 0; x < this.config.width; x++) {
        this.tiles[y][x].isHighlighted = false;
      }
    }
  }

  highlightTiles(positions: Position[]): void {
    this.clearHighlights();
    for (const pos of positions) {
      if (this.isValidPosition(pos)) {
        this.tiles[pos.y][pos.x].isHighlighted = true;
      }
    }
  }

  // === SERIALIZATION ===

  serialize(): Record<string, any> {
    return {
      config: this.config,
      tiles: this.tiles.map(row => 
        row.map(tile => ({
          ...tile,
          occupant: tile.occupant ? tile.occupant.id : undefined
        }))
      ),
      units: Object.fromEntries(
        Array.from(this.units.entries()).map(([id, unit]) => [id, unit.serialize()])
      )
    };
  }

  deserialize(_data: Record<string, any>): void {
    // This would restore map state from saved data
    // Implementation would depend on specific needs
  }
}