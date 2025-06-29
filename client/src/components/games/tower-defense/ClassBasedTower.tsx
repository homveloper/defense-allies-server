'use client'

import React, { useRef } from 'react'
import { useFrame } from '@react-three/fiber'
import { Cylinder, Ring } from '@react-three/drei'
import * as THREE from 'three'
import { Tower } from '@/game/tower-defense/Tower'

interface ClassBasedTowerProps {
  tower: Tower
  colors: any
  isHovered?: boolean
}

export default function ClassBasedTower({ tower, colors, isHovered = false }: ClassBasedTowerProps) {
  const meshRef = useRef<THREE.Mesh>(null)
  const attackEffectRef = useRef<THREE.Mesh>(null)
  
  const config = tower.config
  const position = tower.getPosition()
  
  // 공격 애니메이션 효과
  useFrame(() => {
    if (attackEffectRef.current) {
      if (tower.isAttacking()) {
        const timeSinceAttack = Date.now() - tower.getLastAttackTime()
        const alpha = 1 - (timeSinceAttack / 200)
        attackEffectRef.current.visible = true
        ;(attackEffectRef.current.material as THREE.MeshBasicMaterial).opacity = alpha * 0.5
      } else {
        attackEffectRef.current.visible = false
      }
    }
  })
  
  return (
    <group position={[position.x, position.y, position.z]}>
      {/* 타워 베이스 */}
      <Cylinder
        ref={meshRef}
        position={[0, config.size.height / 2, 0]}
        args={[config.size.radius, config.size.radius, config.size.height, 8]}
      >
        <meshLambertMaterial color={config.color} />
      </Cylinder>
      
      {/* 타워 내부 코어 */}
      <Cylinder
        position={[0, config.size.height / 2, 0]}
        args={[config.size.radius * 0.5, config.size.radius * 0.5, config.size.height * 0.8, 8]}
      >
        <meshLambertMaterial color={colors.bg.primary} />
      </Cylinder>
      
      {/* 타입별 특별한 장식 */}
      {tower.type === 'basic' && (
        <mesh position={[0, config.size.height * 0.7, config.size.radius * 0.8]} rotation={[0, 0, 0]}>
          <cylinderGeometry args={[0.1, 0.1, 0.05, 6]} />
          <meshLambertMaterial color="#f59e0b" />
        </mesh>
      )}
      
      {tower.type === 'splash' && (
        <mesh position={[0, config.size.height + 0.1, 0]} rotation={[0, Math.PI / 4, 0]}>
          <boxGeometry args={[0.1, 0.2, 0.1]} />
          <meshLambertMaterial color="#c084fc" emissive="#4c1d95" emissiveIntensity={0.2} />
        </mesh>
      )}
      
      {tower.type === 'laser' && (
        <group position={[0, config.size.height + 0.1, 0]}>
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
        <mesh position={[0, config.size.height + 0.05, 0]} rotation={[Math.PI / 2, 0, 0]}>
          <cylinderGeometry args={[0.08, 0.08, 0.02, 8]} />
          <meshLambertMaterial color="#f59e0b" />
        </mesh>
      )}
      
      {/* 레벨 표시 구체 */}
      {tower.getLevel() > 1 && (
        <mesh position={[config.size.radius * 0.7, config.size.height * 0.8, config.size.radius * 0.7]}>
          <sphereGeometry args={[0.05, 8, 8]} />
          <meshLambertMaterial color="#ffffff" />
        </mesh>
      )}
      
      {/* 공격 범위 표시 (호버 시) */}
      {isHovered && (
        <Ring
          position={[0, 0.01, 0]}
          rotation={[-Math.PI / 2, 0, 0]}
          args={[tower.getRange() - 0.1, tower.getRange(), 32]}
        >
          <meshBasicMaterial 
            color={config.color} 
            transparent={true} 
            opacity={0.3}
            side={THREE.DoubleSide}
          />
        </Ring>
      )}
      
      {/* 공격 효과 */}
      <mesh
        ref={attackEffectRef}
        position={[0, config.size.height / 2, 0]}
        visible={false}
      >
        <sphereGeometry args={[config.size.radius + 0.1, 16, 16]} />
        <meshBasicMaterial 
          color="#ffff00" 
          transparent={true}
          opacity={0.5}
        />
      </mesh>
      
      {/* 타겟 라인 표시 (디버그용) */}
      {tower.getCurrentTarget() && (
        <group>
          {/* 간단한 라인 구현 */}
        </group>
      )}
    </group>
  )
}