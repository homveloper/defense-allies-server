'use client'

import React, { useRef, useEffect, useState } from 'react'
import { useFrame } from '@react-three/fiber'
import { RoundedBox } from '@react-three/drei'
import * as THREE from 'three'

interface ProgressHealthBarProps {
  current: number
  max: number
  position: [number, number, number]
  width?: number
  height?: number
  depth?: number
  animationSpeed?: number
  showDamageDelay?: boolean // 데미지 딜레이 표시 (빨간 부분)
  cornerRadius?: number
}

export default function ProgressHealthBar({ 
  current, 
  max, 
  position, 
  width = 0.6, 
  height = 0.08,
  depth = 0.02,
  animationSpeed = 0.15,
  showDamageDelay = true,
  cornerRadius = 0.01
}: ProgressHealthBarProps) {
  const groupRef = useRef<THREE.Group>(null)
  const fillRef = useRef<THREE.Mesh>(null)
  const damageDelayRef = useRef<THREE.Mesh>(null)
  
  // 애니메이션 상태
  const [displayHealth, setDisplayHealth] = useState(current)
  const [damageDelayHealth, setDamageDelayHealth] = useState(current)
  
  const targetHealthPercent = Math.max(0, Math.min(1, current / max))
  const displayHealthPercent = Math.max(0, Math.min(1, displayHealth / max))
  const damageDelayPercent = Math.max(0, Math.min(1, damageDelayHealth / max))

  // 체력 감소 애니메이션
  useFrame((state, delta) => {
    const targetHealth = current
    
    // 즉시 체력이 증가한 경우
    if (targetHealth > displayHealth) {
      setDisplayHealth(targetHealth)
      setDamageDelayHealth(targetHealth)
      return
    }
    
    // 체력 감소 시 부드러운 애니메이션
    if (displayHealth > targetHealth) {
      const diff = displayHealth - targetHealth
      const change = Math.max(diff * animationSpeed, 0.5) // 최소 변화량 보장
      
      setDisplayHealth(prev => {
        const newValue = prev - change
        if (newValue <= targetHealth) {
          return targetHealth
        }
        return newValue
      })
    }
    
    // 데미지 딜레이 (빨간 부분) 애니메이션 - 더 느리게
    if (showDamageDelay && damageDelayHealth > displayHealth) {
      const delayDiff = damageDelayHealth - displayHealth
      const delayChange = Math.max(delayDiff * animationSpeed * 0.3, 0.2)
      
      setDamageDelayHealth(prev => {
        const newValue = prev - delayChange
        if (newValue <= displayHealth) {
          return displayHealth
        }
        return newValue
      })
    }
  })

  // 체력에 따른 색상 계산
  const getHealthColor = (percent: number): THREE.Color => {
    if (percent > 0.6) return new THREE.Color('#22c55e') // 초록색
    if (percent > 0.3) return new THREE.Color('#f59e0b') // 주황색
    return new THREE.Color('#ef4444') // 빨간색
  }

  // 체력이 0 이하면 렌더링하지 않음
  if (max <= 0) {
    return null
  }

  return (
    <group ref={groupRef} position={position}>
      {/* 배경 (둥근 모서리) */}
      <RoundedBox
        args={[width, height, depth]}
        radius={cornerRadius}
        smoothness={4}
      >
        <meshBasicMaterial 
          color="#1a1a1a" 
          transparent={true}
          opacity={0.8}
        />
      </RoundedBox>
      
      {/* 데미지 딜레이 표시 (빨간 부분) */}
      {showDamageDelay && damageDelayPercent > displayHealthPercent && (
        <RoundedBox
          ref={damageDelayRef}
          position={[-(width * (1 - damageDelayPercent)) / 2, 0.001, 0]}
          args={[width * damageDelayPercent, height * 0.8, depth * 0.8]}
          radius={cornerRadius * 0.8}
          smoothness={4}
        >
          <meshBasicMaterial 
            color="#dc2626"
            transparent={true}
            opacity={0.6}
          />
        </RoundedBox>
      )}
      
      {/* 체력바 채우기 (메인) */}
      <RoundedBox
        ref={fillRef}
        position={[-(width * (1 - displayHealthPercent)) / 2, 0.002, 0]}
        args={[width * displayHealthPercent, height * 0.9, depth * 0.9]}
        radius={cornerRadius * 0.9}
        smoothness={4}
      >
        <meshBasicMaterial 
          color={getHealthColor(displayHealthPercent)}
          transparent={true}
          opacity={0.9}
        />
      </RoundedBox>
      
      {/* 광택 효과 */}
      <RoundedBox
        position={[0, height * 0.15, depth * 0.1]}
        args={[width * 0.9, height * 0.3, depth * 0.5]}
        radius={cornerRadius * 0.5}
        smoothness={4}
      >
        <meshBasicMaterial 
          color="#ffffff"
          transparent={true}
          opacity={0.2}
        />
      </RoundedBox>
      
      {/* 테두리 */}
      <mesh>
        <ringGeometry args={[width * 0.48, width * 0.52, 8]} />
        <meshBasicMaterial 
          color="#ffffff" 
          transparent={true}
          opacity={0.3}
          side={THREE.DoubleSide}
        />
      </mesh>
      
      {/* 크리티컬 체력 경고 효과 */}
      {displayHealthPercent < 0.2 && displayHealthPercent > 0 && (
        <RoundedBox
          position={[0, 0.003, 0]}
          args={[width, height, depth]}
          radius={cornerRadius}
          smoothness={4}
        >
          <meshBasicMaterial 
            color="#ff0000"
            transparent={true}
            opacity={0.3 * Math.sin(Date.now() * 0.01)} // 깜빡임 효과
          />
        </RoundedBox>
      )}
      
      {/* 체력 회복 효과 */}
      {displayHealth < current && (
        <RoundedBox
          position={[0, 0.004, 0]}
          args={[width, height, depth]}
          radius={cornerRadius}
          smoothness={4}
        >
          <meshBasicMaterial 
            color="#00ff88"
            transparent={true}
            opacity={0.4}
          />
        </RoundedBox>
      )}
    </group>
  )
}