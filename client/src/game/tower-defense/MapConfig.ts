import { Waypoint } from './MovementHandler'

export interface MapSize {
  width: number
  height: number
}

export interface MapBounds {
  minX: number
  maxX: number
  minY: number
  maxY: number
}

export type MapShape = 'rectangle' | 'square' | 'L-shape' | 'U-shape' | 'cross'

export interface MapConfig {
  size: MapSize
  shape: MapShape
  enemyPath: Waypoint[]
  spawnPoint: Waypoint
  exitPoint: Waypoint
  blockedCells?: Set<string> // 배치 불가능한 셀들
}

export class MapGenerator {
  
  // 기본 10x10 맵 생성
  static createSquareMap(size: number = 10): MapConfig {
    const enemyPath = this.generateSimplePath(size, size)
    
    return {
      size: { width: size, height: size },
      shape: 'square',
      enemyPath,
      spawnPoint: enemyPath[0],
      exitPoint: enemyPath[enemyPath.length - 1],
      blockedCells: new Set()
    }
  }
  
  // 직사각형 맵 생성
  static createRectangleMap(width: number, height: number): MapConfig {
    const enemyPath = this.generateSimplePath(width, height)
    
    return {
      size: { width, height },
      shape: 'rectangle',
      enemyPath,
      spawnPoint: enemyPath[0],
      exitPoint: enemyPath[enemyPath.length - 1],
      blockedCells: new Set()
    }
  }
  
  // L자 형태 맵 생성
  static createLShapeMap(size: number = 10): MapConfig {
    const blockedCells = new Set<string>()
    
    // L자 형태의 블록된 영역 (우상단 절반)
    const halfSize = Math.floor(size / 2)
    for (let x = halfSize; x < size; x++) {
      for (let y = 0; y < halfSize; y++) {
        blockedCells.add(`${x}-${y}`)
      }
    }
    
    const enemyPath = this.generateLShapePath(size, blockedCells)
    
    return {
      size: { width: size, height: size },
      shape: 'L-shape',
      enemyPath,
      spawnPoint: enemyPath[0],
      exitPoint: enemyPath[enemyPath.length - 1],
      blockedCells
    }
  }
  
  // U자 형태 맵 생성
  static createUShapeMap(width: number = 12, height: number = 8): MapConfig {
    const blockedCells = new Set<string>()
    
    // U자 형태의 중앙 블록
    const centerWidth = Math.floor(width / 3)
    const centerStartX = Math.floor(width / 3)
    const centerHeight = Math.floor(height * 0.6)
    
    for (let x = centerStartX; x < centerStartX + centerWidth; x++) {
      for (let y = 0; y < centerHeight; y++) {
        blockedCells.add(`${x}-${y}`)
      }
    }
    
    const enemyPath = this.generateUShapePath(width, height, blockedCells)
    
    return {
      size: { width, height },
      shape: 'U-shape',
      enemyPath,
      spawnPoint: enemyPath[0],
      exitPoint: enemyPath[enemyPath.length - 1],
      blockedCells
    }
  }
  
  // 십자 형태 맵 생성
  static createCrossMap(size: number = 11): MapConfig {
    const blockedCells = new Set<string>()
    
    // 십자 형태 - 모서리 4개 블록
    const cornerSize = Math.floor(size / 3)
    
    // 좌상단
    for (let x = 0; x < cornerSize; x++) {
      for (let y = 0; y < cornerSize; y++) {
        blockedCells.add(`${x}-${y}`)
      }
    }
    
    // 우상단
    for (let x = size - cornerSize; x < size; x++) {
      for (let y = 0; y < cornerSize; y++) {
        blockedCells.add(`${x}-${y}`)
      }
    }
    
    // 좌하단
    for (let x = 0; x < cornerSize; x++) {
      for (let y = size - cornerSize; y < size; y++) {
        blockedCells.add(`${x}-${y}`)
      }
    }
    
    // 우하단
    for (let x = size - cornerSize; x < size; x++) {
      for (let y = size - cornerSize; y < size; y++) {
        blockedCells.add(`${x}-${y}`)
      }
    }
    
    const enemyPath = this.generateCrossPath(size, blockedCells)
    
    return {
      size: { width: size, height: size },
      shape: 'cross',
      enemyPath,
      spawnPoint: enemyPath[0],
      exitPoint: enemyPath[enemyPath.length - 1],
      blockedCells
    }
  }
  
  // 간단한 S자 경로 생성 (직사각형용)
  private static generateSimplePath(width: number, height: number): Waypoint[] {
    const centerY = Math.floor(height / 2)
    const quarterY = Math.floor(height / 4)
    const threeQuarterY = Math.floor(height * 3 / 4)
    const quarterX = Math.floor(width / 4)
    const threeQuarterX = Math.floor(width * 3 / 4)
    
    return [
      { x: 0, y: centerY },
      { x: quarterX, y: centerY },
      { x: quarterX, y: quarterY },
      { x: threeQuarterX, y: quarterY },
      { x: threeQuarterX, y: threeQuarterY },
      { x: width - 1, y: threeQuarterY }
    ]
  }
  
  // L자 형태 경로 생성
  private static generateLShapePath(size: number, blockedCells: Set<string>): Waypoint[] {
    const halfSize = Math.floor(size / 2)
    
    return [
      { x: 0, y: halfSize },
      { x: halfSize - 2, y: halfSize },
      { x: halfSize - 2, y: size - 2 },
      { x: size - 2, y: size - 2 },
      { x: size - 1, y: size - 2 }
    ]
  }
  
  // U자 형태 경로 생성
  private static generateUShapePath(width: number, height: number, blockedCells: Set<string>): Waypoint[] {
    const centerY = Math.floor(height * 0.7)
    
    return [
      { x: 0, y: centerY },
      { x: Math.floor(width / 4), y: centerY },
      { x: Math.floor(width / 4), y: height - 1 },
      { x: Math.floor(width * 3 / 4), y: height - 1 },
      { x: Math.floor(width * 3 / 4), y: centerY },
      { x: width - 1, y: centerY }
    ]
  }
  
  // 십자 형태 경로 생성
  private static generateCrossPath(size: number, blockedCells: Set<string>): Waypoint[] {
    const center = Math.floor(size / 2)
    const quarter = Math.floor(size / 4)
    const threeQuarter = Math.floor(size * 3 / 4)
    
    return [
      { x: center, y: 0 },
      { x: center, y: quarter },
      { x: quarter, y: quarter },
      { x: quarter, y: threeQuarter },
      { x: center, y: threeQuarter },
      { x: center, y: size - 1 }
    ]
  }
  
  // 경로 셀들을 Set으로 변환
  static getPathCells(path: Waypoint[]): Set<string> {
    const cells = new Set<string>()
    
    for (let i = 0; i < path.length - 1; i++) {
      const start = path[i]
      const end = path[i + 1]
      
      if (start.x === end.x) {
        // 수직 라인
        const minY = Math.min(start.y, end.y)
        const maxY = Math.max(start.y, end.y)
        for (let y = minY; y <= maxY; y++) {
          cells.add(`${start.x}-${y}`)
        }
      } else if (start.y === end.y) {
        // 수평 라인
        const minX = Math.min(start.x, end.x)
        const maxX = Math.max(start.x, end.x)
        for (let x = minX; x <= maxX; x++) {
          cells.add(`${x}-${start.y}`)
        }
      } else {
        // 대각선 (필요한 경우)
        cells.add(`${start.x}-${start.y}`)
        cells.add(`${end.x}-${end.y}`)
      }
    }
    
    return cells
  }
  
  // 맵 경계 확인
  static isValidPosition(x: number, y: number, mapConfig: MapConfig): boolean {
    if (x < 0 || x >= mapConfig.size.width || y < 0 || y >= mapConfig.size.height) {
      return false
    }
    
    // 블록된 셀 확인
    if (mapConfig.blockedCells?.has(`${x}-${y}`)) {
      return false
    }
    
    return true
  }
  
  // 미리 정의된 맵들
  static getPredefinedMaps(): { [key: string]: MapConfig } {
    return {
      'small-square': this.createSquareMap(8),
      'medium-square': this.createSquareMap(10),
      'large-square': this.createSquareMap(12),
      'small-rect': this.createRectangleMap(12, 8),
      'wide-rect': this.createRectangleMap(15, 8),
      'l-shape': this.createLShapeMap(10),
      'u-shape': this.createUShapeMap(12, 8),
      'cross': this.createCrossMap(11)
    }
  }
}