'use client'

import React, { useState, useRef, useEffect, useCallback } from 'react'
import { Canvas } from '@react-three/fiber'
import { OrthographicCamera } from '@react-three/drei'
import { useTheme } from '@/contexts/ThemeContext'
import { World } from '@/systems/ECS'
import { MovementSystem } from '@/systems/MovementSystem'
import { RenderSystem, RenderData } from '@/systems/RenderSystem'
import { HealthSystem } from '@/systems/HealthSystem'
import {
  createPositionComponent,
  createMovementComponent,
  createRenderComponent,
  createHealthComponent,
  createEnemyTypeComponent,
  createTowerTypeComponent
} from '@/systems/components'
import GameGrid from './GameGrid'
import ECSEntity from './ECSEntity'

interface ECSGameSceneProps {
  selectedTowerType: string | null
  gameStateHook: any
}

// 20x20 그리드 설정
const GRID_SIZE = 20
const CELL_SIZE = 1

// 적 이동 경로
const ENEMY_PATH = [
  { x: 0, y: 10 },
  { x: 5, y: 10 },
  { x: 5, y: 5 },
  { x: 15, y: 5 },
  { x: 15, y: 15 },
  { x: 19, y: 15 }
]

// 경로 셀들
const PATH_CELLS = new Set<string>()
for (let i = 0; i < ENEMY_PATH.length - 1; i++) {
  const start = ENEMY_PATH[i]
  const end = ENEMY_PATH[i + 1]
  
  if (start.x === end.x) {
    const minY = Math.min(start.y, end.y)
    const maxY = Math.max(start.y, end.y)
    for (let y = minY; y <= maxY; y++) {
      PATH_CELLS.add(`${start.x}-${y}`)
    }
  } else {
    const minX = Math.min(start.x, end.x)
    const maxX = Math.max(start.x, end.x)
    for (let x = minX; x <= maxX; x++) {
      PATH_CELLS.add(`${x}-${start.y}`)
    }
  }
}

export default function ECSGameScene({ selectedTowerType, gameStateHook }: ECSGameSceneProps) {
  const [hoveredCell, setHoveredCell] = useState<{ x: number, y: number } | null>(null)
  const [renderData, setRenderData] = useState<RenderData[]>([])
  const { colors } = useTheme()
  const { gameState, placeTower } = gameStateHook
  
  // ECS World 초기화
  const worldRef = useRef<World>(new World())
  const movementSystemRef = useRef<MovementSystem>(new MovementSystem())
  const renderSystemRef = useRef<RenderSystem>(new RenderSystem())
  const healthSystemRef = useRef<HealthSystem>(new HealthSystem())
  
  // 시스템 초기화
  useEffect(() => {
    const world = worldRef.current
    world.addSystem(movementSystemRef.current)
    world.addSystem(renderSystemRef.current)
    world.addSystem(healthSystemRef.current)
    
    // 렌더 시스템 콜백 설정
    renderSystemRef.current.setRenderCallback(setRenderData)
  }, [])
  
  // 게임 상태에서 ECS로 엔티티 동기화
  useEffect(() => {
    const world = worldRef.current
    
    // 기존 엔티티 제거
    const existingEntities = world.getAllEntities()
    existingEntities.forEach(entity => {
      world.removeEntity(entity.id)
    })
    
    // 적 엔티티 생성
    gameState.enemies.forEach((enemy: any) => {
      const entity = world.createEntity(enemy.id)
      
      world.addComponent(entity.id, createPositionComponent(
        enemy.position[0], 0, enemy.position[2]
      ))
      
      world.addComponent(entity.id, createMovementComponent(
        enemy.speed, ENEMY_PATH
      ))
      
      const enemyColors = {
        fast: '#10b981',
        tank: '#ef4444',
        basic: '#f59e0b'
      }
      
      const enemySizes = {
        fast: { radius: 0.15, height: 0.2 },
        tank: { radius: 0.3, height: 0.4 },
        basic: { radius: 0.2, height: 0.3 }
      }
      
      world.addComponent(entity.id, createRenderComponent(
        'enemy',
        enemyColors[enemy.type] || enemyColors.basic,
        enemySizes[enemy.type] || enemySizes.basic
      ))
      
      world.addComponent(entity.id, createHealthComponent(enemy.maxHealth))
      
      world.addComponent(entity.id, createEnemyTypeComponent(
        enemy.type, enemy.value
      ))
      
      // 현재 체력 설정
      const healthComponent = world.getEntity(entity.id)?.components.get('health')
      if (healthComponent) {
        (healthComponent as any).current = enemy.health
      }
    })
    
    // 타워 엔티티 생성
    gameState.towers.forEach((tower: any) => {
      const entity = world.createEntity(tower.id)
      
      world.addComponent(entity.id, createPositionComponent(
        tower.position[0], 0, tower.position[2]
      ))
      
      const towerColors = {
        basic: '#2563eb',
        slow: '#f59e0b',
        splash: '#8b5cf6',
        laser: '#22c55e'
      }
      
      world.addComponent(entity.id, createRenderComponent(
        'tower',
        towerColors[tower.type] || towerColors.basic,
        { radius: 0.3, height: 0.6 }
      ))
      
      world.addComponent(entity.id, createTowerTypeComponent(
        tower.type, tower.damage, tower.range, tower.attackSpeed
      ))
    })
  }, [gameState.enemies, gameState.towers])
  
  // ECS 업데이트 루프
  useEffect(() => {
    let lastTime = 0
    let animationId: number
    
    const update = (currentTime: number) => {
      const deltaTime = currentTime - lastTime || 16.67
      lastTime = currentTime
      
      worldRef.current.update(deltaTime)
      
      animationId = requestAnimationFrame(update)
    }
    
    animationId = requestAnimationFrame(update)
    
    return () => {
      if (animationId) {
        cancelAnimationFrame(animationId)
      }
    }
  }, [])
  
  // 셀 클릭 처리
  const handleCellClick = useCallback((x: number, z: number) => {
    if (!selectedTowerType) return
    
    const cellKey = `${x}-${z}`
    const hasTower = gameState.towers.some((tower: any) =>
      Math.floor(tower.position[0]) === x &&
      Math.floor(tower.position[2]) === z
    )
    
    if (!PATH_CELLS.has(cellKey) && !hasTower) {
      const worldPos: [number, number, number] = [x, 0, z]
      placeTower(worldPos, selectedTowerType as any)
    }
  }, [selectedTowerType, gameState.towers, placeTower])
  
  const handleCellHover = useCallback((x: number, z: number) => {
    setHoveredCell({ x, y: z })
  }, [])
  
  const handleCellLeave = useCallback(() => {
    setHoveredCell(null)
  }, [])
  
  return (
    <div className="w-full h-full" style={{ backgroundColor: colors.bg.secondary }}>
      <Canvas style={{ width: '100%', height: '100%' }}>
        <OrthographicCamera
          makeDefault
          position={[10, 20, 10]}
          rotation={[-Math.PI / 2, 0, 0]}
          zoom={30}
          near={0.1}
          far={1000}
        />
        
        <ambientLight intensity={0.8} />
        <directionalLight position={[0, 10, 0]} intensity={0.5} />
        
        <GameGrid 
          gridSize={GRID_SIZE}
          cellSize={CELL_SIZE}
          pathCells={PATH_CELLS}
          colors={colors}
          hoveredCell={hoveredCell}
          selectedTowerType={selectedTowerType}
          gameState={gameState}
          onCellClick={handleCellClick}
          onCellHover={handleCellHover}
          onCellLeave={handleCellLeave}
        />
        
        {/* ECS 엔티티들 렌더링 */}
        {renderData.map(data => (
          <ECSEntity
            key={data.entityId}
            data={data}
            colors={colors}
            isHovered={
              hoveredCell &&
              Math.floor(data.position.x) === hoveredCell.x &&
              Math.floor(data.position.z) === hoveredCell.y
            }
          />
        ))}
      </Canvas>
    </div>
  )
}