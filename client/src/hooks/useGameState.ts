import { useState, useEffect, useCallback } from 'react'

export interface Enemy {
  id: string
  position: [number, number, number]
  health: number
  maxHealth: number
  speed: number
  pathIndex: number
  value: number
  type: 'basic' | 'fast' | 'tank'
}

export interface Tower {
  id: string
  position: [number, number, number]
  type: 'basic' | 'splash' | 'slow' | 'laser'
  level: number
  damage: number
  range: number
  attackSpeed: number
  lastAttack: number
  cost: number
}

export interface GameState {
  wave: number
  health: number
  maxHealth: number
  gold: number
  score: number
  enemies: Enemy[]
  towers: Tower[]
  isWaveActive: boolean
  waveTimer: number
  gameStatus: 'playing' | 'paused' | 'game-over' | 'victory'
}

const INITIAL_STATE: GameState = {
  wave: 1,
  health: 100,
  maxHealth: 100,
  gold: 300,
  score: 0,
  enemies: [],
  towers: [],
  isWaveActive: false,
  waveTimer: 0,
  gameStatus: 'playing'
}

const WAVE_CONFIGS = [
  { enemies: 10, enemyTypes: ['basic'], spawnInterval: 1000 },
  { enemies: 15, enemyTypes: ['basic', 'fast'], spawnInterval: 800 },
  { enemies: 20, enemyTypes: ['basic', 'fast', 'tank'], spawnInterval: 600 },
  // Add more wave configurations...
]

export function useGameState() {
  const [gameState, setGameState] = useState<GameState>(INITIAL_STATE)
  const [spawnTimer, setSpawnTimer] = useState(0)
  const [enemiesSpawned, setEnemiesSpawned] = useState(0)

  // Start a new wave
  const startWave = useCallback(() => {
    console.log('Starting wave:', gameState.wave)
    setGameState(prev => ({
      ...prev,
      isWaveActive: true,
      waveTimer: 0
    }))
    setEnemiesSpawned(0)
    setSpawnTimer(0)
  }, [gameState.wave])

  // Spawn enemy
  const spawnEnemy = useCallback((type: Enemy['type']) => {
    console.log('Spawning enemy:', type)
    const enemyStats = {
      basic: { health: 50, speed: 2.5, value: 10 },
      fast: { health: 30, speed: 4, value: 15 },
      tank: { health: 150, speed: 1.5, value: 25 }
    }

    const stats = enemyStats[type]
    const newEnemy: Enemy = {
      id: `enemy-${Date.now()}-${Math.random()}`,
      position: [0, 0, 10], // 시작점 좌표 (0, 10) - Three.js 좌표계
      health: stats.health,
      maxHealth: stats.health,
      speed: stats.speed,
      pathIndex: 0,
      value: stats.value,
      type
    }

    setGameState(prev => ({
      ...prev,
      enemies: [...prev.enemies, newEnemy]
    }))
  }, [])

  // Place tower
  const placeTower = useCallback((position: [number, number, number], type: Tower['type']) => {
    const towerStats = {
      basic: { damage: 25, range: 2, attackSpeed: 1000, cost: 50 },
      splash: { damage: 40, range: 1.5, attackSpeed: 1500, cost: 100 },
      slow: { damage: 15, range: 2.5, attackSpeed: 800, cost: 75 },
      laser: { damage: 60, range: 3, attackSpeed: 2000, cost: 150 }
    }

    const stats = towerStats[type]
    
    setGameState(prev => {
      if (prev.gold < stats.cost) return prev

      const newTower: Tower = {
        id: `tower-${Date.now()}-${Math.random()}`,
        position,
        type,
        level: 1,
        damage: stats.damage,
        range: stats.range,
        attackSpeed: stats.attackSpeed,
        lastAttack: 0,
        cost: stats.cost
      }

      return {
        ...prev,
        towers: [...prev.towers, newTower],
        gold: prev.gold - stats.cost
      }
    })
  }, [])

  // Remove enemy
  const removeEnemy = useCallback((enemyId: string, reachedEnd = false) => {
    setGameState(prev => {
      const enemy = prev.enemies.find(e => e.id === enemyId)
      if (!enemy) return prev

      const newHealth = reachedEnd ? prev.health - 10 : prev.health
      const newGold = reachedEnd ? prev.gold : prev.gold + enemy.value
      const newScore = reachedEnd ? prev.score : prev.score + enemy.value * 10

      return {
        ...prev,
        enemies: prev.enemies.filter(e => e.id !== enemyId),
        health: Math.max(0, newHealth),
        gold: newGold,
        score: newScore,
        gameStatus: newHealth <= 0 ? 'game-over' : prev.gameStatus
      }
    })
  }, [])

  // Damage enemy
  const damageEnemy = useCallback((enemyId: string, damage: number) => {
    setGameState(prev => {
      const updatedEnemies = prev.enemies.map(enemy => {
        if (enemy.id === enemyId) {
          const newHealth = enemy.health - damage
          if (newHealth <= 0) {
            // Award gold and score for killing enemy
            return null
          }
          return { ...enemy, health: newHealth }
        }
        return enemy
      }).filter(Boolean) as Enemy[]

      const killedEnemy = prev.enemies.find(e => e.id === enemyId && e.health - damage <= 0)
      
      return {
        ...prev,
        enemies: updatedEnemies,
        gold: killedEnemy ? prev.gold + killedEnemy.value : prev.gold,
        score: killedEnemy ? prev.score + killedEnemy.value * 10 : prev.score
      }
    })
  }, [])

  // Update enemy position (throttled for smooth movement)
  const updateEnemyPosition = useCallback((enemyId: string, position: [number, number, number], pathIndex: number) => {
    setGameState(prev => ({
      ...prev,
      enemies: prev.enemies.map(enemy =>
        enemy.id === enemyId ? { ...enemy, position, pathIndex } : enemy
      )
    }))
  }, [])

  // Game loop for wave management
  useEffect(() => {
    console.log('Wave effect:', { isActive: gameState.isWaveActive, status: gameState.gameStatus, wave: gameState.wave })
    
    if (!gameState.isWaveActive || gameState.gameStatus !== 'playing') return

    const currentWave = WAVE_CONFIGS[Math.min(gameState.wave - 1, WAVE_CONFIGS.length - 1)]
    console.log('Current wave config:', currentWave, 'Enemies spawned:', enemiesSpawned)
    
    const interval = setInterval(() => {
      setSpawnTimer(prev => {
        const newTimer = prev + 100
        
        // Spawn enemies
        if (enemiesSpawned < currentWave.enemies && newTimer >= currentWave.spawnInterval) {
          const enemyType = currentWave.enemyTypes[
            Math.floor(Math.random() * currentWave.enemyTypes.length)
          ] as Enemy['type']
          
          console.log('Spawning enemy at timer:', newTimer, 'type:', enemyType)
          spawnEnemy(enemyType)
          setEnemiesSpawned(prev => prev + 1)
          return 0
        }
        
        return newTimer
      })
    }, 100)

    return () => clearInterval(interval)
  }, [gameState.isWaveActive, gameState.gameStatus, gameState.wave, enemiesSpawned, spawnEnemy])

  // Check wave completion separately
  useEffect(() => {
    if (!gameState.isWaveActive) return

    const currentWave = WAVE_CONFIGS[Math.min(gameState.wave - 1, WAVE_CONFIGS.length - 1)]
    
    if (enemiesSpawned >= currentWave.enemies && gameState.enemies.length === 0) {
      setGameState(prev => ({
        ...prev,
        isWaveActive: false,
        wave: prev.wave + 1,
        gold: prev.gold + 50 // Wave completion bonus
      }))
      setEnemiesSpawned(0) // Reset for next wave
    }
  }, [gameState.enemies.length, enemiesSpawned, gameState.isWaveActive, gameState.wave])

  return {
    gameState,
    startWave,
    placeTower,
    removeEnemy,
    damageEnemy,
    updateEnemyPosition,
    setGameState
  }
}