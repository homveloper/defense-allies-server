'use client'

import React, { useRef, useEffect } from 'react'
import { useFrame } from '@react-three/fiber'
import { Cylinder, Plane } from '@react-three/drei'
import * as THREE from 'three'
import { Enemy } from '@/game/tower-defense/Enemy'
import ProgressHealthBar from '@/components/games/shared/ProgressHealthBar'

interface ClassBasedEnemyProps {
  enemy: Enemy
}

export default function ClassBasedEnemy({ enemy }: ClassBasedEnemyProps) {
  const groupRef = useRef<THREE.Group>(null)
  const meshRef = useRef<THREE.Mesh>(null)
  
  const config = enemy.config
  const healthPercent = enemy.getHealthPercent()
  
  // 위치 및 회전 업데이트 (매 프레임)
  useFrame(() => {
    if (groupRef.current) {
      const pos = enemy.getPosition()
      groupRef.current.position.set(pos.x, pos.y, pos.z)
    }
    
    if (meshRef.current) {
      meshRef.current.rotation.y = enemy.getRotation()
    }
  })
  
  // 피해 효과
  const damageEffect = healthPercent < 1 ? (1 - healthPercent) * 0.5 : 0
  
  if (!enemy.isAliveStatus()) {
    return null
  }
  
  return (
    <group ref={groupRef}>
      {/* 적 몸체 */}
      <Cylinder
        ref={meshRef}
        position={[0, config.size.height / 2, 0]}
        args={[config.size.radius, config.size.radius, config.size.height, 8]}
      >
        <meshLambertMaterial 
          color={config.color}
          emissive={damageEffect > 0 ? "#ff0000" : "#000000"}
          emissiveIntensity={damageEffect}
        />
      </Cylinder>
      
      {/* 타입별 장식 */}
      {enemy.type === 'tank' && (
        <>
          <Cylinder
            position={[0, config.size.height / 2, 0]}
            args={[config.size.radius * 0.8, config.size.radius * 0.8, config.size.height * 0.9, 8]}
          >
            <meshLambertMaterial color="#ffffff" />
          </Cylinder>
          <Cylinder
            position={[0, config.size.height / 2, 0]}
            args={[config.size.radius * 0.6, config.size.radius * 0.6, config.size.height * 0.8, 8]}
          >
            <meshLambertMaterial color={config.color} />
          </Cylinder>
        </>
      )}
      
      {enemy.type === 'fast' && (
        <group position={[0, config.size.height * 0.7, 0]}>
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
        <mesh position={[0, config.size.height * 0.9, 0]}>
          <boxGeometry args={[0.05, 0.05, 0.05]} />
          <meshLambertMaterial color="#ffffff" />
        </mesh>
      )}
      
      {/* 체력바 */}
      <ProgressHealthBar
        current={enemy.getHealth()}
        max={enemy.maxHealth}
        position={[0, config.size.height + 0.3, 0]}
        width={Math.max(1.0, config.size.radius * 4.0)}
        height={0.15}
        depth={0.04}
        animationSpeed={0.25}
        showDamageDelay={true}
        cornerRadius={0.02}
      />
      
      {/* 그림자 */}
      <Plane
        position={[0, 0.001, 0]}
        rotation={[-Math.PI / 2, 0, 0]}
        args={[config.size.radius * 2, config.size.radius * 2]}
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