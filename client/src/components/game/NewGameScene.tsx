'use client'

import React, { useState, useRef, useEffect } from 'react'
import { useTheme } from '@/contexts/ThemeContext'
import { renderTower } from './Tower'
import { renderEnemy } from './Enemy'

interface GameSceneProps {
  selectedTowerType: string | null
  gameStateHook: any
}

// 20x20 그리드 설정
const GRID_SIZE = 20
const CELL_SIZE = 30
const CANVAS_SIZE = GRID_SIZE * CELL_SIZE

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

export default function NewGameScene({ selectedTowerType, gameStateHook }: GameSceneProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [hoveredCell, setHoveredCell] = useState<{ x: number, y: number } | null>(null)
  const { colors } = useTheme()
  const { gameState, placeTower, damageEnemy } = gameStateHook

  // 적 이동 애니메이션 - 60fps로 부드럽게
  useEffect(() => {
    if (gameState.enemies.length === 0) return

    let animationId: number
    
    const animate = () => {
      gameState.enemies.forEach((enemy: any) => {
        if (enemy.pathIndex < ENEMY_PATH.length - 1) {
          const currentPoint = ENEMY_PATH[enemy.pathIndex]
          const nextPoint = ENEMY_PATH[enemy.pathIndex + 1]
          
          const dx = nextPoint.x - currentPoint.x
          const dy = nextPoint.y - currentPoint.y
          const distance = Math.sqrt(dx * dx + dy * dy)
          
          if (distance > 0) {
            const speed = enemy.speed * 0.016 // 프레임당 속도 (60fps 기준)
            const moveX = (dx / distance) * speed
            const moveY = (dy / distance) * speed
            
            // 다음 포인트에 도달했는지 확인
            const currentX = enemy.position[0] + 10
            const currentY = enemy.position[2] + 10
            const distanceToNext = Math.sqrt(
              Math.pow(nextPoint.x - currentX, 2) + 
              Math.pow(nextPoint.y - currentY, 2)
            )
            
            if (distanceToNext < 0.05) {
              // 다음 웨이포인트로 이동
              enemy.pathIndex++
              if (enemy.pathIndex >= ENEMY_PATH.length - 1) {
                // 경로 끝에 도달
                enemy.position[0] = nextPoint.x - 10
                enemy.position[2] = nextPoint.y - 10
              }
            } else {
              // 계속 이동
              enemy.position[0] += moveX
              enemy.position[2] += moveY
            }
          }
        }
      })
      
      animationId = requestAnimationFrame(animate)
    }
    
    animate()

    return () => cancelAnimationFrame(animationId)
  }, [gameState.enemies])

  // 타워 공격 시스템
  useEffect(() => {
    if (gameState.towers.length === 0 || gameState.enemies.length === 0) return

    const attackInterval = setInterval(() => {
      gameState.towers.forEach((tower: any) => {
        const now = Date.now()
        if (now - tower.lastAttack < tower.attackSpeed) return

        // 사거리 내 적 찾기
        const towerGridX = Math.floor(tower.position[0] + 10)
        const towerGridY = Math.floor(tower.position[2] + 10)
        
        const enemiesInRange = gameState.enemies.filter((enemy: any) => {
          const enemyGridX = Math.floor(enemy.position[0] + 10)
          const enemyGridY = Math.floor(enemy.position[2] + 10)
          const distance = Math.sqrt(
            Math.pow(towerGridX - enemyGridX, 2) + 
            Math.pow(towerGridY - enemyGridY, 2)
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


  // 실시간 캔버스 업데이트를 위한 별도 루프
  useEffect(() => {
    let renderAnimationId: number
    
    const renderLoop = () => {
      const canvas = canvasRef.current
      if (!canvas) {
        renderAnimationId = requestAnimationFrame(renderLoop)
        return
      }

      const ctx = canvas.getContext('2d')
      if (!ctx) {
        renderAnimationId = requestAnimationFrame(renderLoop)
        return
      }

      // 배경 지우기
      ctx.fillStyle = colors.game.background
      ctx.fillRect(0, 0, CANVAS_SIZE, CANVAS_SIZE)

      // 그리드 그리기
      ctx.strokeStyle = colors.border.secondary
      ctx.lineWidth = 1
      for (let i = 0; i <= GRID_SIZE; i++) {
        // 세로선
        ctx.beginPath()
        ctx.moveTo(i * CELL_SIZE, 0)
        ctx.lineTo(i * CELL_SIZE, CANVAS_SIZE)
        ctx.stroke()
        
        // 가로선
        ctx.beginPath()
        ctx.moveTo(0, i * CELL_SIZE)
        ctx.lineTo(CANVAS_SIZE, i * CELL_SIZE)
        ctx.stroke()
      }

      // 경로 하이라이트
      ctx.fillStyle = '#ef444450' // 반투명 빨간색
      PATH_CELLS.forEach(cellKey => {
        const [x, y] = cellKey.split('-').map(Number)
        ctx.fillRect(x * CELL_SIZE + 1, y * CELL_SIZE + 1, CELL_SIZE - 2, CELL_SIZE - 2)
      })

      // 경로 테두리
      ctx.strokeStyle = '#ef4444'
      ctx.lineWidth = 2
      PATH_CELLS.forEach(cellKey => {
        const [x, y] = cellKey.split('-').map(Number)
        ctx.strokeRect(x * CELL_SIZE + 1, y * CELL_SIZE + 1, CELL_SIZE - 2, CELL_SIZE - 2)
      })

      // 호버 효과
      if (hoveredCell && selectedTowerType) {
        const cellKey = `${hoveredCell.x}-${hoveredCell.y}`
        const hasTower = gameState.towers.some((tower: any) =>
          Math.floor(tower.position[0] + 10) === hoveredCell.x &&
          Math.floor(tower.position[2] + 10) === hoveredCell.y
        )
        
        if (!PATH_CELLS.has(cellKey) && !hasTower) {
          ctx.fillStyle = colors.text.accent + '40' // 반투명 파란색
          ctx.fillRect(
            hoveredCell.x * CELL_SIZE + 1,
            hoveredCell.y * CELL_SIZE + 1,
            CELL_SIZE - 2,
            CELL_SIZE - 2
          )
        }
      }

      // 타워 그리기
      gameState.towers.forEach((tower: any) => {
        const gridX = Math.floor(tower.position[0] + 10)
        const gridY = Math.floor(tower.position[2] + 10)
        const isHovered = hoveredCell && 
                         gridX === hoveredCell.x && 
                         gridY === hoveredCell.y
        
        renderTower(tower, ctx, CELL_SIZE, colors, isHovered)
      })

      // 적 그리기
      gameState.enemies.forEach((enemy: any) => {
        renderEnemy(enemy, ctx, CELL_SIZE)
      })
      
      renderAnimationId = requestAnimationFrame(renderLoop)
    }
    
    renderAnimationId = requestAnimationFrame(renderLoop)
    
    return () => cancelAnimationFrame(renderAnimationId)
  }, [colors, hoveredCell, selectedTowerType])

  // 마우스 이벤트
  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const x = Math.floor((e.clientX - rect.left) / CELL_SIZE)
    const y = Math.floor((e.clientY - rect.top) / CELL_SIZE)
    
    if (x >= 0 && x < GRID_SIZE && y >= 0 && y < GRID_SIZE) {
      setHoveredCell({ x, y })
    } else {
      setHoveredCell(null)
    }
  }

  const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!hoveredCell || !selectedTowerType) return

    const cellKey = `${hoveredCell.x}-${hoveredCell.y}`
    const hasTower = gameState.towers.some((tower: any) =>
      Math.floor(tower.position[0] + 10) === hoveredCell.x &&
      Math.floor(tower.position[2] + 10) === hoveredCell.y
    )

    // 경로나 기존 타워가 있는 곳에는 배치 불가
    if (!PATH_CELLS.has(cellKey) && !hasTower) {
      const worldPos: [number, number, number] = [
        hoveredCell.x - 10, // 그리드 좌표를 월드 좌표로 변환
        0,
        hoveredCell.y - 10
      ]
      placeTower(worldPos, selectedTowerType as any)
    }
  }

  return (
    <div 
      className="w-full h-full flex items-center justify-center"
      style={{ backgroundColor: colors.bg.secondary }}
    >
      <canvas
        ref={canvasRef}
        width={CANVAS_SIZE}
        height={CANVAS_SIZE}
        className="border-2 cursor-pointer"
        style={{ borderColor: colors.border.primary }}
        onMouseMove={handleMouseMove}
        onMouseLeave={() => setHoveredCell(null)}
        onClick={handleClick}
      />
    </div>
  )
}

