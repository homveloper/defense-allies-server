import { Enemy, EnemyType, EnemyConfig } from './Enemy'
import { Tower, TowerType, TowerConfig } from './Tower'
import { Position, Waypoint } from './MovementHandler'
import { MapConfig } from './MapConfig'

export interface GameWorldConfig {
  mapConfig: MapConfig
  enemySpawnPoint: Position
}

export class GameWorld {
  private enemies: Map<string, Enemy> = new Map()
  private towers: Map<string, Tower> = new Map()
  private config: GameWorldConfig
  private lastUpdateTime: number = 0

  // 게임 상태
  private score: number = 0
  private gold: number = 300
  private health: number = 100
  private wave: number = 1

  constructor(config: GameWorldConfig) {
    this.config = config
  }

  // 매 프레임 업데이트
  update(): void {
    const currentTime = Date.now()
    const deltaTime = currentTime - this.lastUpdateTime || 16.67
    this.lastUpdateTime = currentTime

    // 적 업데이트
    this.updateEnemies(deltaTime)
    
    // 타워 업데이트
    this.updateTowers(deltaTime)
    
    // 죽은 적 제거 및 보상 처리
    this.cleanupDeadEnemies()
    
    // 경로 완주한 적 처리
    this.handleEnemiesReachedEnd()
  }

  // 적들 업데이트
  private updateEnemies(deltaTime: number): void {
    this.enemies.forEach(enemy => {
      enemy.update(deltaTime)
    })
  }

  // 타워들 업데이트
  private updateTowers(deltaTime: number): void {
    const enemyList = Array.from(this.enemies.values())
    this.towers.forEach(tower => {
      tower.update(deltaTime, enemyList)
    })
  }

  // 죽은 적 제거 및 보상
  private cleanupDeadEnemies(): void {
    const deadEnemies: Enemy[] = []
    
    this.enemies.forEach(enemy => {
      if (!enemy.isAliveStatus()) {
        deadEnemies.push(enemy)
        this.enemies.delete(enemy.id)
      }
    })

    // 보상 지급
    deadEnemies.forEach(enemy => {
      this.gold += enemy.value
      this.score += enemy.value * 10
    })
  }

  // 경로 완주한 적 처리
  private handleEnemiesReachedEnd(): void {
    const reachedEnemies: Enemy[] = []
    
    this.enemies.forEach(enemy => {
      if (enemy.hasReachedEnd()) {
        reachedEnemies.push(enemy)
        this.enemies.delete(enemy.id)
      }
    })

    // 체력 감소
    reachedEnemies.forEach(enemy => {
      this.health -= 10 // 적 1마리당 체력 10 감소
    })

    this.health = Math.max(0, this.health)
  }

  // 적 생성
  createEnemy(id: string, type: EnemyType): Enemy | null {
    const config = this.getEnemyConfig(type)
    if (!config) return null

    const enemy = new Enemy(
      id,
      config,
      this.config.enemySpawnPoint,
      this.config.mapConfig.enemyPath
    )

    this.enemies.set(id, enemy)
    return enemy
  }

  // 타워 생성
  createTower(id: string, type: TowerType, position: Position): Tower | null {
    const config = this.getTowerConfig(type)
    if (!config) return null

    // 골드 체크
    if (this.gold < config.cost) {
      return null
    }

    // 위치 유효성 체크 (경로나 다른 타워와 겹치는지)
    if (!this.isValidTowerPosition(position)) {
      return null
    }

    const tower = new Tower(id, config, position)
    this.towers.set(id, tower)
    this.gold -= config.cost

    return tower
  }

  // 타워 위치 유효성 검사
  private isValidTowerPosition(position: Position): boolean {
    const mapConfig = this.config.mapConfig
    
    // 맵 경계 확인
    if (position.x < 0 || position.x >= mapConfig.size.width || 
        position.z < 0 || position.z >= mapConfig.size.height) {
      return false
    }
    
    // 블록된 셀 확인
    const cellKey = `${Math.floor(position.x)}-${Math.floor(position.z)}`
    if (mapConfig.blockedCells?.has(cellKey)) {
      return false
    }
    
    // 다른 타워와 겹치는지 확인
    for (const tower of this.towers.values()) {
      const towerPos = tower.getPosition()
      const distance = Math.sqrt(
        Math.pow(position.x - towerPos.x, 2) +
        Math.pow(position.z - towerPos.z, 2)
      )
      if (distance < 0.1) {
        return false
      }
    }

    // 경로와 겹치는지 확인
    for (const waypoint of mapConfig.enemyPath) {
      const distance = Math.sqrt(
        Math.pow(position.x - waypoint.x, 2) +
        Math.pow(position.z - waypoint.y, 2)
      )
      if (distance < 0.5) {
        return false
      }
    }

    return true
  }

  // 적 제거
  removeEnemy(id: string): boolean {
    return this.enemies.delete(id)
  }

  // 타워 제거
  removeTower(id: string): boolean {
    const tower = this.towers.get(id)
    if (tower) {
      this.gold += tower.getSellPrice()
      return this.towers.delete(id)
    }
    return false
  }

  // 타워 업그레이드
  upgradeTower(id: string): boolean {
    const tower = this.towers.get(id)
    if (!tower) return false

    const cost = tower.getUpgradeCost()
    if (this.gold < cost) return false

    tower.levelUp()
    this.gold -= cost
    return true
  }

  // 게터 메서드들
  getEnemies(): Enemy[] {
    return Array.from(this.enemies.values())
  }

  getTowers(): Tower[] {
    return Array.from(this.towers.values())
  }

  getEnemy(id: string): Enemy | undefined {
    return this.enemies.get(id)
  }

  getTower(id: string): Tower | undefined {
    return this.towers.get(id)
  }

  getScore(): number {
    return this.score
  }

  getGold(): number {
    return this.gold
  }

  getHealth(): number {
    return this.health
  }

  getWave(): number {
    return this.wave
  }

  // 웨이브 증가
  nextWave(): void {
    this.wave++
  }

  // 골드 추가 (치트용)
  addGold(amount: number): void {
    this.gold += amount
  }
  
  // 골드 차감
  subtractGold(amount: number): boolean {
    if (this.gold >= amount) {
      this.gold -= amount
      return true
    }
    return false
  }

  // 적 설정 정보
  private getEnemyConfig(type: EnemyType): EnemyConfig | null {
    const configs: Record<EnemyType, EnemyConfig> = {
      basic: {
        type: 'basic',
        health: 50,
        speed: 2.5,
        value: 10,
        color: '#f59e0b',
        size: { radius: 0.2, height: 0.3 }
      },
      fast: {
        type: 'fast',
        health: 30,
        speed: 4,
        value: 15,
        color: '#10b981',
        size: { radius: 0.15, height: 0.2 }
      },
      tank: {
        type: 'tank',
        health: 150,
        speed: 1.5,
        value: 25,
        color: '#ef4444',
        size: { radius: 0.3, height: 0.4 }
      }
    }
    return configs[type] || null
  }

  // 타워 설정 정보
  private getTowerConfig(type: TowerType): TowerConfig | null {
    const configs: Record<TowerType, TowerConfig> = {
      basic: {
        type: 'basic',
        damage: 25,
        range: 2,
        attackSpeed: 1000,
        cost: 50,
        color: '#2563eb',
        size: { radius: 0.3, height: 0.6 }
      },
      splash: {
        type: 'splash',
        damage: 40,
        range: 1.5,
        attackSpeed: 1500,
        cost: 100,
        color: '#8b5cf6',
        size: { radius: 0.3, height: 0.8 }
      },
      slow: {
        type: 'slow',
        damage: 15,
        range: 2.5,
        attackSpeed: 800,
        cost: 75,
        color: '#f59e0b',
        size: { radius: 0.3, height: 0.5 }
      },
      laser: {
        type: 'laser',
        damage: 60,
        range: 3,
        attackSpeed: 2000,
        cost: 150,
        color: '#22c55e',
        size: { radius: 0.3, height: 1.0 }
      }
    }
    return configs[type] || null
  }

  // 게임 상태 확인
  isGameOver(): boolean {
    return this.health <= 0
  }

  // 웨이브 클리어 확인
  isWaveCleared(): boolean {
    return this.enemies.size === 0
  }

  // 디버그 정보
  getDebugInfo(): any {
    return {
      enemies: this.enemies.size,
      towers: this.towers.size,
      score: this.score,
      gold: this.gold,
      health: this.health,
      wave: this.wave
    }
  }
}