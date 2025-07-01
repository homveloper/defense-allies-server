// Common game-related types for tower defense and other games

export interface Enemy {
  id: string
  position: [number, number, number]
  health: number
  maxHealth: number
  speed: number
  type: string
  pathIndex?: number
  value?: number
}

export interface Tower {
  id: string
  position: [number, number, number]
  type: string
  damage: number
  range: number
  attackSpeed: number
  lastAttack?: number
}

export interface GameState {
  enemies: Enemy[]
  towers: Tower[]
  health?: number
  score?: number
  wave?: number
  isRunning?: boolean
}

export interface GameStateHook {
  gameState: GameState
  placeTower: (position: [number, number, number], type: string) => void
  damageEnemy?: (id: string, damage: number) => void
  startWave?: () => void
  pauseGame?: () => void
  resetGame?: () => void
}

export interface Colors {
  [key: string]: string | { [key: string]: string }
}

export interface GameEngineService {
  update: (deltaTime: number) => void
  getGameState: () => GameState
  initialize: () => void
  destroy: () => void
  [key: string]: unknown
}