'use client'

import React, { useRef, useMemo } from 'react'
import { useFrame } from '@react-three/fiber'
import { Cylinder, Ring } from '@react-three/drei'
import * as THREE from 'three'

interface ThreeTowerProps {
  tower: {
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
  colors: any
  isHovered?: boolean
}

// 타워 색상
function getTowerColor(type: string): string {
  switch (type) {
    case 'basic': return '#2563eb'    // 기사단 요새 - 파란색
    case 'slow': return '#f59e0b'     // 상인 길드 - 주황색  
    case 'splash': return '#8b5cf6'   // 마법사 탑 - 보라색
    case 'laser': return '#22c55e'    // 대성당 - 초록색
    default: return '#6b7280'
  }
}

export default function ThreeTower({ tower, colors, isHovered = false }: ThreeTowerProps) {
  const meshRef = useRef<THREE.Mesh>(null)
  const attackEffectRef = useRef<THREE.Mesh>(null)
  
  const towerColor = getTowerColor(tower.type)
  
  // 공격 애니메이션 효과
  useFrame(() => {
    const timeSinceAttack = Date.now() - tower.lastAttack
    if (attackEffectRef.current) {
      if (timeSinceAttack < 200) {
        const alpha = 1 - (timeSinceAttack / 200)
        attackEffectRef.current.visible = true
        ;(attackEffectRef.current.material as THREE.MeshBasicMaterial).opacity = alpha * 0.5
      } else {
        attackEffectRef.current.visible = false
      }
    }
  })
  
  // 타워 높이 (타입별 다름)
  const towerHeight = useMemo(() => {
    switch (tower.type) {
      case 'basic': return 0.6
      case 'splash': return 0.8
      case 'laser': return 1.0
      case 'slow': return 0.5
      default: return 0.6
    }
  }, [tower.type])
  
  return (
    <group position={[tower.position[0], 0, tower.position[2]]}>
      {/* 타워 베이스 */}
      <Cylinder
        ref={meshRef}
        position={[0, towerHeight / 2, 0]}
        args={[0.3, 0.3, towerHeight, 8]}
      >
        <meshLambertMaterial color={towerColor} />
      </Cylinder>
      
      {/* 타워 내부 코어 */}
      <Cylinder
        position={[0, towerHeight / 2, 0]}
        args={[0.15, 0.15, towerHeight * 0.8, 8]}
      >
        <meshLambertMaterial color={colors.bg.primary} />
      </Cylinder>
      
      {/* 타입별 특별한 장식 */}
      {tower.type === 'basic' && (
        // 기사단 - 방패
        <mesh position={[0, towerHeight * 0.7, 0.25]} rotation={[0, 0, 0]}>
          <cylinderGeometry args={[0.1, 0.1, 0.05, 6]} />
          <meshLambertMaterial color="#f59e0b" />
        </mesh>
      )}
      
      {tower.type === 'splash' && (
        // 마법사 - 크리스탈
        <mesh position={[0, towerHeight + 0.1, 0]} rotation={[0, Math.PI / 4, 0]}>
          <boxGeometry args={[0.1, 0.2, 0.1]} />
          <meshLambertMaterial color="#c084fc" emissive="#4c1d95" emissiveIntensity={0.2} />
        </mesh>
      )}
      
      {tower.type === 'laser' && (
        // 대성당 - 십자가
        <group position={[0, towerHeight + 0.1, 0]}>
          <mesh>
            <boxGeometry args={[0.03, 0.15, 0.03]} />
            <meshLambertMaterial color="#f59e0b" />
          </mesh>
          <mesh>
            <boxGeometry args={[0.1, 0.03, 0.03]} />
            <meshLambertMaterial color="#f59e0b" />
          </mesh>
        </group>
      )}
      
      {tower.type === 'slow' && (
        // 상인 길드 - 동전
        <mesh position={[0, towerHeight + 0.05, 0]} rotation={[Math.PI / 2, 0, 0]}>
          <cylinderGeometry args={[0.08, 0.08, 0.02, 8]} />
          <meshLambertMaterial color="#f59e0b" />
        </mesh>
      )}
      
      {/* 레벨 표시 구체 */}
      {tower.level > 1 && (
        <mesh position={[0.2, towerHeight * 0.8, 0.2]}>
          <sphereGeometry args={[0.05, 8, 8]} />
          <meshLambertMaterial color="#ffffff" />
        </mesh>
      )}
      
      {/* 공격 범위 표시 (호버 시) */}
      {isHovered && (
        <Ring
          position={[0, 0.01, 0]}
          rotation={[-Math.PI / 2, 0, 0]}
          args={[tower.range - 0.1, tower.range, 32]}
        >
          <meshBasicMaterial 
            color={towerColor} 
            transparent={true} 
            opacity={0.3}
            side={THREE.DoubleSide}
          />
        </Ring>
      )}
      
      {/* 공격 효과 */}
      <mesh
        ref={attackEffectRef}
        position={[0, towerHeight / 2, 0]}
        visible={false}
      >
        <sphereGeometry args={[0.4, 16, 16]} />
        <meshBasicMaterial 
          color="#ffff00" 
          transparent={true}
          opacity={0.5}
        />
      </mesh>
    </group>
  )
}