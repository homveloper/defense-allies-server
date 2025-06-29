'use client'

import React from 'react'
import { Plane } from '@react-three/drei'
import { MapConfig } from '@/game/tower-defense/MapConfig'

interface GameGridProps {
  gridSize: number
  cellSize: number
  pathCells: Set<string>
  colors: any
  hoveredCell: { x: number, y: number } | null
  selectedTowerType: string | null
  gameState: any
  onCellClick: (x: number, z: number) => void
  onCellHover: (x: number, z: number) => void
  onCellLeave: () => void
  mapConfig: MapConfig
}

export default function GameGrid({ 
  gridSize, 
  cellSize, 
  pathCells, 
  colors, 
  hoveredCell, 
  selectedTowerType, 
  gameState,
  onCellClick,
  onCellHover,
  onCellLeave,
  mapConfig
}: GameGridProps) {
  
  // 그리드 셀들 생성
  const gridCells = []
  
  for (let x = 0; x < mapConfig.size.width; x++) {
    for (let z = 0; z < mapConfig.size.height; z++) {
      const cellKey = `${x}-${z}`
      const isPath = pathCells.has(cellKey)
      const isBlocked = mapConfig.blockedCells?.has(cellKey) || false
      const isHovered = hoveredCell && hoveredCell.x === x && hoveredCell.y === z
      const hasTower = gameState.towers.some((tower: any) =>
        Math.floor(tower.position[0]) === x &&
        Math.floor(tower.position[2]) === z
      )
      
      let cellColor = colors.game.background
      let opacity = 0.1
      
      if (isBlocked) {
        cellColor = '#64748b' // 블록된 영역 - 회색
        opacity = 0.8
      } else if (isPath) {
        cellColor = '#ef4444' // 경로 - 빨간색
        opacity = 0.7
      } else if (isHovered && selectedTowerType && !hasTower && !isBlocked) {
        cellColor = colors.text.accent || '#3b82f6' // 호버 - 파란색
        opacity = 0.5
      }
      
      gridCells.push(
        <Plane
          key={`${x}-${z}`}
          position={[x, 0, z]}
          rotation={[-Math.PI / 2, 0, 0]}
          args={[cellSize * 0.9, cellSize * 0.9]}
          onClick={(e) => {
            e.stopPropagation()
            onCellClick(x, z)
          }}
          onPointerEnter={(e) => {
            e.stopPropagation()
            onCellHover(x, z)
          }}
          onPointerLeave={(e) => {
            e.stopPropagation()
            onCellLeave()
          }}
        >
          <meshBasicMaterial 
            color={cellColor}
            transparent={true}
            opacity={opacity}
          />
        </Plane>
      )
    }
  }
  
  // 그리드 라인들
  const gridLines = []
  
  // 세로 라인들
  for (let i = 0; i <= mapConfig.size.width; i++) {
    gridLines.push(
      <mesh key={`vertical-${i}`} position={[i - 0.5, 0.001, mapConfig.size.height / 2 - 0.5]}>
        <boxGeometry args={[0.02, 0.01, mapConfig.size.height]} />
        <meshBasicMaterial color={colors.border.secondary} />
      </mesh>
    )
  }
  
  // 가로 라인들
  for (let i = 0; i <= mapConfig.size.height; i++) {
    gridLines.push(
      <mesh key={`horizontal-${i}`} position={[mapConfig.size.width / 2 - 0.5, 0.001, i - 0.5]}>
        <boxGeometry args={[mapConfig.size.width, 0.01, 0.02]} />
        <meshBasicMaterial color={colors.border.secondary} />
      </mesh>
    )
  }
  
  return (
    <group>
      {/* 바닥 */}
      <Plane
        position={[mapConfig.size.width / 2 - 0.5, -0.01, mapConfig.size.height / 2 - 0.5]}
        rotation={[-Math.PI / 2, 0, 0]}
        args={[mapConfig.size.width, mapConfig.size.height]}
      >
        <meshBasicMaterial color={colors.game.background} />
      </Plane>
      
      {/* 그리드 셀들 */}
      {gridCells}
      
      {/* 그리드 라인들 */}
      {gridLines}
    </group>
  )
}