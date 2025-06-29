import { Component } from './ECS'

// 위치 컴포넌트
export interface PositionComponent extends Component {
  type: 'position'
  x: number
  y: number
  z: number
}

// 이동 컴포넌트 (경로 기반)
export interface MovementComponent extends Component {
  type: 'movement'
  speed: number
  path: Array<{ x: number, y: number }>
  currentPathIndex: number
  isMoving: boolean
}

// 렌더링 컴포넌트
export interface RenderComponent extends Component {
  type: 'render'
  meshType: 'enemy' | 'tower'
  color: string
  size: { radius: number, height: number }
  rotation: number
}

// 체력 컴포넌트
export interface HealthComponent extends Component {
  type: 'health'
  current: number
  max: number
}

// 적 타입 컴포넌트
export interface EnemyTypeComponent extends Component {
  type: 'enemyType'
  enemyType: 'basic' | 'fast' | 'tank'
  value: number
}

// 타워 타입 컴포넌트
export interface TowerTypeComponent extends Component {
  type: 'towerType'
  towerType: 'basic' | 'splash' | 'slow' | 'laser'
  damage: number
  range: number
  attackSpeed: number
  lastAttack: number
}

// 컴포넌트 팩토리 함수들
export const createPositionComponent = (x: number, y: number, z: number): PositionComponent => ({
  type: 'position',
  x, y, z
})

export const createMovementComponent = (
  speed: number, 
  path: Array<{ x: number, y: number }>
): MovementComponent => ({
  type: 'movement',
  speed,
  path,
  currentPathIndex: 0,
  isMoving: true
})

export const createRenderComponent = (
  meshType: 'enemy' | 'tower',
  color: string,
  size: { radius: number, height: number }
): RenderComponent => ({
  type: 'render',
  meshType,
  color,
  size,
  rotation: 0
})

export const createHealthComponent = (max: number): HealthComponent => ({
  type: 'health',
  current: max,
  max
})

export const createEnemyTypeComponent = (
  enemyType: 'basic' | 'fast' | 'tank',
  value: number
): EnemyTypeComponent => ({
  type: 'enemyType',
  enemyType,
  value
})

export const createTowerTypeComponent = (
  towerType: 'basic' | 'splash' | 'slow' | 'laser',
  damage: number,
  range: number,
  attackSpeed: number
): TowerTypeComponent => ({
  type: 'towerType',
  towerType,
  damage,
  range,
  attackSpeed,
  lastAttack: 0
})