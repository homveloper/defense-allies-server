'use client'

import React, { useState, useRef, useEffect, useCallback } from 'react'
import { Canvas } from '@react-three/fiber'
import { OrthographicCamera } from '@react-three/drei'
import { useTheme } from '@/contexts/ThemeContext'
import { GameWorld, GameWorldConfig } from '@/game/tower-defense/GameWorld'
import { Enemy } from '@/game/tower-defense/Enemy'
import { Tower } from '@/game/tower-defense/Tower'
import { MapGenerator, MapConfig } from '@/game/tower-defense/MapConfig'
import GameGrid from './GameGrid'
import ClassBasedEnemy from './ClassBasedEnemy'
import ClassBasedTower from './ClassBasedTower'

interface ClassBasedGameSceneProps {
  selectedTowerType: string | null
  gameStateHook: any
}

// 게임 설정
const CELL_SIZE = 1

// 현재 사용할 맵 (나중에 설정으로 변경 가능)
// 다양한 맵을 테스트해보려면 아래 중 하나를 선택:
// const CURRENT_MAP_CONFIG = MapGenerator.createSquareMap(10)
// const CURRENT_MAP_CONFIG = MapGenerator.createRectangleMap(12, 8)
// const CURRENT_MAP_CONFIG = MapGenerator.createLShapeMap(10)
// const CURRENT_MAP_CONFIG = MapGenerator.createUShapeMap(12, 8)
// const CURRENT_MAP_CONFIG = MapGenerator.createCrossMap(11)

const CURRENT_MAP_CONFIG = MapGenerator.createSquareMap(10)

export default function ClassBasedGameScene({ selectedTowerType, gameStateHook }: ClassBasedGameSceneProps) {
  const [hoveredCell, setHoveredCell] = useState<{ x: number, y: number } | null>(null)
  const [currentMapConfig] = useState<MapConfig>(CURRENT_MAP_CONFIG)
  const { colors } = useTheme()
  const { gameState, placeTower } = gameStateHook
  
  // 게임 월드 초기화
  const gameWorldRef = useRef<GameWorld>()
  const [enemies, setEnemies] = useState<Enemy[]>([])
  const [towers, setTowers] = useState<Tower[]>([])
  
  // 경로 셀들 계산
  const pathCells = MapGenerator.getPathCells(currentMapConfig.enemyPath)
  
  // 게임 월드 초기화
  useEffect(() => {
    const config: GameWorldConfig = {
      mapConfig: currentMapConfig,
      enemySpawnPoint: { 
        x: currentMapConfig.spawnPoint.x, 
        y: 0, 
        z: currentMapConfig.spawnPoint.y 
      }
    }
    
    gameWorldRef.current = new GameWorld(config)
  }, [currentMapConfig])
  
  // 기존 게임 상태를 게임 월드와 동기화
  useEffect(() => {
    if (!gameWorldRef.current) return
    
    const world = gameWorldRef.current
    
    // 기존 엔티티들 제거
    const currentEnemies = world.getEnemies()
    const currentTowers = world.getTowers()
    
    currentEnemies.forEach(enemy => world.removeEnemy(enemy.id))
    currentTowers.forEach(tower => world.removeTower(tower.id))
    
    // 새 적들 추가
    gameState.enemies.forEach((enemy: any) => {
      const newEnemy = world.createEnemy(enemy.id, enemy.type)
      if (newEnemy) {
        // 기존 상태 복사
        if (enemy.health !== newEnemy.getHealth()) {
          const damage = newEnemy.getHealth() - enemy.health
          newEnemy.takeDamage(damage)
        }
        
        // 위치 설정 (private 접근 문제로 다른 방법 사용)
        // newEnemy['movementHandler'].setPosition({
        //   x: enemy.position[0],
        //   y: 0,
        //   z: enemy.position[2]
        // })
      }
    })
    
    // 새 타워들 추가
    gameState.towers.forEach((tower: any) => {
      const newTower = world.createTower(
        tower.id,
        tower.type,
        { x: tower.position[0], y: 0, z: tower.position[2] }
      )
      // 골드 복구 (createTower에서 차감했으므로)
      if (newTower) {
        world.addGold(newTower.config.cost)
      }
    })
    
  }, [gameState.enemies, gameState.towers])
  
  // 게임 루프
  useEffect(() => {
    if (!gameWorldRef.current) return
    
    let animationId: number
    
    const gameLoop = () => {
      if (gameWorldRef.current) {
        gameWorldRef.current.update()
        
        // 상태 업데이트
        setEnemies([...gameWorldRef.current.getEnemies()])
        setTowers([...gameWorldRef.current.getTowers()])
      }
      
      animationId = requestAnimationFrame(gameLoop)
    }
    
    animationId = requestAnimationFrame(gameLoop)
    
    return () => {
      if (animationId) {
        cancelAnimationFrame(animationId)
      }
    }
  }, [])
  
  // 셀 클릭 처리
  const handleCellClick = useCallback((x: number, z: number) => {
    if (!selectedTowerType || !gameWorldRef.current) return
    
    const cellKey = `${x}-${z}`
    
    // 경로, 블록된 셀이 아닌 경우에만 타워 배치
    if (!pathCells.has(cellKey) && !currentMapConfig.blockedCells?.has(cellKey)) {
      const tower = gameWorldRef.current.createTower(
        `tower-${Date.now()}-${Math.random()}`,
        selectedTowerType as any,
        { x, y: 0, z }
      )
      
      if (tower) {
        // 기존 게임 상태에도 반영
        const worldPos: [number, number, number] = [x, 0, z]
        placeTower(worldPos, selectedTowerType as any)
      }
    }
  }, [selectedTowerType, placeTower, pathCells, currentMapConfig.blockedCells])
  
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
          position={[
            currentMapConfig.size.width / 2, 
            20, 
            currentMapConfig.size.height / 2
          ]}
          rotation={[-Math.PI / 2, 0, 0]}
          zoom={40}
          near={0.1}
          far={1000}
        />
        
        <ambientLight intensity={0.8} />
        <directionalLight position={[0, 10, 0]} intensity={0.5} />
        
        <GameGrid 
          gridSize={currentMapConfig.size.width}
          cellSize={CELL_SIZE}
          pathCells={pathCells}
          colors={colors}
          hoveredCell={hoveredCell}
          selectedTowerType={selectedTowerType}
          gameState={gameState}
          onCellClick={handleCellClick}
          onCellHover={handleCellHover}
          onCellLeave={handleCellLeave}
          mapConfig={currentMapConfig}
        />
        
        {/* 적들 렌더링 */}
        {enemies.map(enemy => (
          <ClassBasedEnemy
            key={enemy.id}
            enemy={enemy}
          />
        ))}
        
        {/* 타워들 렌더링 */}
        {towers.map(tower => (
          <ClassBasedTower
            key={tower.id}
            tower={tower}
            colors={colors}
            isHovered={
              hoveredCell &&
              Math.floor(tower.getPosition().x) === hoveredCell.x &&
              Math.floor(tower.getPosition().z) === hoveredCell.y
            }
          />
        ))}
      </Canvas>
    </div>
  )
}