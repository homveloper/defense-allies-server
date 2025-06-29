'use client'

import React, { useState, useRef, useEffect } from 'react'
import { Canvas } from '@react-three/fiber'
import { OrthographicCamera } from '@react-three/drei'
import { useTheme } from '@/contexts/ThemeContext'
import ThreeTower from './ThreeTower'
import SmoothThreeEnemy from './SmoothThreeEnemy'
import GameGrid from './GameGrid'

interface GameSceneProps {
  selectedTowerType: string | null
  gameStateHook: any
}

// 20x20 그리드 설정
const GRID_SIZE = 20
const CELL_SIZE = 1

// 적 이동 경로 (단순한 S자 경로)
const ENEMY_PATH = [
  { x: 0, y: 10 },   // 시작점 (왼쪽 중앙)
  { x: 5, y: 10 },
  { x: 5, y: 5 },
  { x: 15, y: 5 },
  { x: 15, y: 15 },
  { x: 19, y: 15 }   // 끝점 (오른쪽)
]

// 경로를 Set으로 변환 (빠른 검색용)
const PATH_CELLS = new Set<string>()
for (let i = 0; i < ENEMY_PATH.length - 1; i++) {
  const start = ENEMY_PATH[i]
  const end = ENEMY_PATH[i + 1]
  
  if (start.x === end.x) {
    // 수직 라인
    const minY = Math.min(start.y, end.y)
    const maxY = Math.max(start.y, end.y)
    for (let y = minY; y <= maxY; y++) {
      PATH_CELLS.add(`${start.x}-${y}`)
    }
  } else {
    // 수평 라인
    const minX = Math.min(start.x, end.x)
    const maxX = Math.max(start.x, end.x)
    for (let x = minX; x <= maxX; x++) {
      PATH_CELLS.add(`${x}-${start.y}`)
    }
  }
}

export default function ThreeGameScene({ selectedTowerType, gameStateHook }: GameSceneProps) {
  const [hoveredCell, setHoveredCell] = useState<{ x: number, y: number } | null>(null)
  const { colors } = useTheme()
  const { gameState, placeTower, damageEnemy } = gameStateHook

  // 적 이동 애니메이션 - 단일 루프로 통합
  useEffect(() => {
    let animationId: number
    let lastTime = 0
    
    const animate = (currentTime: number) => {
      const deltaTime = currentTime - lastTime
      lastTime = currentTime
      
      // 고정 타임스텝 (16.67ms = 60fps)
      const fixedDelta = 16.67
      
      gameState.enemies.forEach((enemy: any) => {
        if (enemy.pathIndex < ENEMY_PATH.length - 1) {
          const currentPoint = ENEMY_PATH[enemy.pathIndex]
          const nextPoint = ENEMY_PATH[enemy.pathIndex + 1]
          
          const dx = nextPoint.x - currentPoint.x
          const dy = nextPoint.y - currentPoint.y
          const distance = Math.sqrt(dx * dx + dy * dy)
          
          if (distance > 0) {
            // 고정된 속도 (프레임 레이트 독립적)
            const speed = enemy.speed * (fixedDelta / 1000) // 초당 속도
            const moveX = (dx / distance) * speed
            const moveY = (dy / distance) * speed
            
            // 다음 포인트에 도달했는지 확인
            const currentX = enemy.position[0]
            const currentY = enemy.position[2]
            const distanceToNext = Math.sqrt(
              Math.pow(nextPoint.x - currentX, 2) + 
              Math.pow(nextPoint.y - currentY, 2)
            )
            
            if (distanceToNext < 0.02) {
              // 다음 웨이포인트로 이동
              enemy.pathIndex++
              if (enemy.pathIndex >= ENEMY_PATH.length - 1) {
                // 경로 끝에 도달
                enemy.position[0] = nextPoint.x
                enemy.position[2] = nextPoint.y
              }
            } else {
              // 계속 이동 (선형 이동)
              enemy.position[0] += moveX
              enemy.position[2] += moveY
            }
          }
        }
      })
      
      animationId = requestAnimationFrame(animate)
    }
    
    animationId = requestAnimationFrame(animate)
    
    // 정리 함수로 이전 루프 중단
    return () => {
      if (animationId) {
        cancelAnimationFrame(animationId)
      }
    }
  }, []) // 의존성 제거 - 한 번만 실행

  // 타워 공격 시스템
  useEffect(() => {
    if (gameState.towers.length === 0 || gameState.enemies.length === 0) return

    const attackInterval = setInterval(() => {
      gameState.towers.forEach((tower: any) => {
        const now = Date.now()
        if (now - tower.lastAttack < tower.attackSpeed) return

        // 사거리 내 적 찾기
        const enemiesInRange = gameState.enemies.filter((enemy: any) => {
          const distance = Math.sqrt(
            Math.pow(tower.position[0] - enemy.position[0], 2) + 
            Math.pow(tower.position[2] - enemy.position[2], 2)
          )
          return distance <= tower.range && enemy.health > 0
        })

        if (enemiesInRange.length > 0) {
          // 가장 가까운 적 공격
          const target = enemiesInRange[0]
          tower.lastAttack = now
          
          // 공격 연출 (0.5초 뒤 데미지 적용)
          setTimeout(() => {
            if (damageEnemy && target.health > 0) {
              damageEnemy(target.id, tower.damage)
            }
          }, 500)
        }
      })
    }, 100)

    return () => clearInterval(attackInterval)
  }, [gameState.towers, gameState.enemies, damageEnemy])

  // 셀 클릭 처리
  const handleCellClick = (x: number, z: number) => {
    if (!selectedTowerType) return
    
    console.log('Cell clicked:', { x, z })
    
    const cellKey = `${x}-${z}`
    const hasTower = gameState.towers.some((tower: any) =>
      Math.floor(tower.position[0]) === x &&
      Math.floor(tower.position[2]) === z
    )

    console.log('Cell check:', { cellKey, isPath: PATH_CELLS.has(cellKey), hasTower })

    // 경로나 기존 타워가 있는 곳에는 배치 불가
    if (!PATH_CELLS.has(cellKey) && !hasTower) {
      const worldPos: [number, number, number] = [x, 0, z]
      console.log('Placing tower at:', worldPos)
      placeTower(worldPos, selectedTowerType as any)
    }
  }

  // 셀 호버 처리
  const handleCellHover = (x: number, z: number) => {
    setHoveredCell({ x, y: z })
  }

  // 셀 호버 떠남 처리
  const handleCellLeave = () => {
    setHoveredCell(null)
  }

  return (
    <div className="w-full h-full" style={{ backgroundColor: colors.bg.secondary }}>
      <Canvas
        style={{ width: '100%', height: '100%' }}
      >
        {/* 2D 탑뷰를 위한 직교 카메라 */}
        <OrthographicCamera
          makeDefault
          position={[10, 20, 10]}
          rotation={[-Math.PI / 2, 0, 0]}
          zoom={30}
          near={0.1}
          far={1000}
        />

        {/* 조명 */}
        <ambientLight intensity={0.8} />
        <directionalLight position={[0, 10, 0]} intensity={0.5} />

        {/* 그리드와 경로 */}
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

        {/* 타워들 */}
        {gameState.towers.map((tower: any) => (
          <ThreeTower
            key={tower.id}
            tower={tower}
            colors={colors}
            isHovered={
              hoveredCell &&
              Math.floor(tower.position[0]) === hoveredCell.x &&
              Math.floor(tower.position[2]) === hoveredCell.y
            }
          />
        ))}

        {/* 적들 */}
        {gameState.enemies.map((enemy: any) => (
          <SmoothThreeEnemy
            key={enemy.id}
            enemy={enemy}
          />
        ))}
      </Canvas>
    </div>
  )
}