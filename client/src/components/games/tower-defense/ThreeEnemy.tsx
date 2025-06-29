'use client'

import React, { useRef, useMemo, useEffect } from 'react'
import { useFrame } from '@react-three/fiber'
import { Cylinder, Plane } from '@react-three/drei'
import * as THREE from 'three'

interface ThreeEnemyProps {
  enemy: {
    id: string
    position: [number, number, number]
    health: number
    maxHealth: number
    speed: number
    pathIndex: number
    value: number
    type: 'basic' | 'fast' | 'tank'
  }
}

// 적 색상
function getEnemyColor(type: string): string {
  switch (type) {
    case 'fast': return '#10b981'  // 빠른 적 - 녹색
    case 'tank': return '#ef4444'  // 탱크 적 - 빨간색
    default: return '#f59e0b'      // 기본 적 - 주황색
  }
}

// 적 크기
function getEnemySize(type: string): { radius: number, height: number } {
  switch (type) {
    case 'fast': return { radius: 0.15, height: 0.2 }   // 빠른 적 - 작음
    case 'tank': return { radius: 0.3, height: 0.4 }    // 탱크 적 - 큼
    default: return { radius: 0.2, height: 0.3 }        // 기본 적 - 중간
  }
}

export default function ThreeEnemy({ enemy }: ThreeEnemyProps) {
  const groupRef = useRef<THREE.Group>(null)
  const meshRef = useRef<THREE.Mesh>(null)
  const healthBarRef = useRef<THREE.Mesh>(null)
  
  const enemyColor = getEnemyColor(enemy.type)
  const { radius, height } = getEnemySize(enemy.type)
  const healthPercent = enemy.health / enemy.maxHealth
  
  // 적이 피해를 받았을 때 효과
  const damageEffect = useMemo(() => {
    return healthPercent < 1 ? (1 - healthPercent) * 0.5 : 0
  }, [healthPercent])
  
  // Three.js 좌표를 실시간으로 업데이트
  useEffect(() => {
    if (groupRef.current) {
      groupRef.current.position.set(
        enemy.position[0], 
        0, 
        enemy.position[2]
      )
    }
  }, [enemy.position[0], enemy.position[2]])
  
  // 회전 애니메이션 (타입별 다른 속도)
  useFrame((state) => {
    if (meshRef.current) {
      const rotationSpeed = enemy.type === 'fast' ? 0.1 : enemy.type === 'tank' ? 0.02 : 0.05
      meshRef.current.rotation.y += rotationSpeed
    }
  })
  
  // 체력바 색상
  const healthBarColor = useMemo(() => {
    if (healthPercent > 0.6) return '#22c55e' // 초록색
    if (healthPercent > 0.3) return '#f59e0b' // 주황색
    return '#ef4444' // 빨간색
  }, [healthPercent])
  
  return (
    <group ref={groupRef}>
      {/* 적 몸체 */}
      <Cylinder
        ref={meshRef}
        position={[0, height / 2, 0]}
        args={[radius, radius, height, 8]}
      >
        <meshLambertMaterial 
          color={enemyColor}
          emissive={damageEffect > 0 ? "#ff0000" : "#000000"}
          emissiveIntensity={damageEffect}
        />
      </Cylinder>
      
      {/* 타입별 특별한 장식 */}
      {enemy.type === 'tank' && (
        <>
          {/* 탱크 - 두꺼운 장갑 */}
          <Cylinder
            position={[0, height / 2, 0]}
            args={[radius * 0.8, radius * 0.8, height * 0.9, 8]}
          >
            <meshLambertMaterial color="#ffffff" />
          </Cylinder>
          <Cylinder
            position={[0, height / 2, 0]}
            args={[radius * 0.6, radius * 0.6, height * 0.8, 8]}
          >
            <meshLambertMaterial color={enemyColor} />
          </Cylinder>
        </>
      )}
      
      {enemy.type === 'fast' && (
        // 빠른 적 - 번개 효과
        <group position={[0, height * 0.7, 0]}>
          <mesh rotation={[0, 0, Math.PI / 6]}>
            <boxGeometry args={[0.02, 0.1, 0.02]} />
            <meshLambertMaterial color="#ffffff" emissive="#ffffff" emissiveIntensity={0.3} />
          </mesh>
          <mesh rotation={[0, 0, -Math.PI / 6]} position={[0.02, -0.02, 0]}>
            <boxGeometry args={[0.02, 0.06, 0.02]} />
            <meshLambertMaterial color="#ffffff" emissive="#ffffff" emissiveIntensity={0.3} />
          </mesh>
        </group>
      )}
      
      {enemy.type === 'basic' && (
        // 기본 적 - 간단한 상단 장식
        <mesh position={[0, height * 0.9, 0]}>
          <boxGeometry args={[0.05, 0.05, 0.05]} />
          <meshLambertMaterial color="#ffffff" />
        </mesh>
      )}
      
      {/* 체력바 배경 */}
      <Plane
        position={[0, height + 0.2, 0]}
        rotation={[-Math.PI / 2, 0, 0]}
        args={[radius * 3, 0.06]}
      >
        <meshBasicMaterial color="#333333" />
      </Plane>
      
      {/* 체력바 */}
      <Plane
        ref={healthBarRef}
        position={[-radius * 1.5 * (1 - healthPercent), height + 0.21, 0]}
        rotation={[-Math.PI / 2, 0, 0]}
        args={[radius * 3 * healthPercent, 0.04]}
      >
        <meshBasicMaterial color={healthBarColor} />
      </Plane>
      
      {/* 체력바 테두리 */}
      <mesh position={[0, height + 0.215, 0]} rotation={[-Math.PI / 2, 0, 0]}>
        <ringGeometry args={[radius * 1.48, radius * 1.52, 4]} />
        <meshBasicMaterial color="#ffffff" side={THREE.DoubleSide} />
      </mesh>
      
      {/* 그림자 */}
      <Plane
        position={[0, 0.001, 0]}
        rotation={[-Math.PI / 2, 0, 0]}
        args={[radius * 2, radius * 2]}
      >
        <meshBasicMaterial 
          color="#000000" 
          transparent={true}
          opacity={0.2}
        />
      </Plane>
    </group>
  )
}