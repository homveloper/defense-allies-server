'use client'

import React, { useRef, useEffect, useState } from 'react'
import { useFrame } from '@react-three/fiber'
import { Plane } from '@react-three/drei'
import * as THREE from 'three'

interface HealthBarProps {
  current: number
  max: number
  position: [number, number, number] // 월드 위치
  width?: number
  height?: number
  showBorder?: boolean
  animationSpeed?: number // 보간 속도 (0-1, 기본값: 0.1)
}

export default function HealthBar({ 
  current, 
  max, 
  position, 
  width = 0.6, 
  height = 0.06,
  showBorder = true,
  animationSpeed = 0.15
}: HealthBarProps) {
  const backgroundRef = useRef<THREE.Mesh>(null)
  const fillRef = useRef<THREE.Mesh>(null)
  const borderRef = useRef<THREE.Mesh>(null)
  
  // 현재 표시되는 체력 (보간용)
  const [displayHealth, setDisplayHealth] = useState(current)
  const targetHealth = current / max
  const currentDisplayPercent = displayHealth / max

  // 부드러운 체력 보간 애니메이션
  useFrame((state, delta) => {
    const targetDisplayHealth = current
    
    if (Math.abs(displayHealth - targetDisplayHealth) > 0.1) {
      // 선형 보간으로 부드럽게 변화
      const diff = targetDisplayHealth - displayHealth
      const change = diff * animationSpeed
      
      setDisplayHealth(prev => {
        const newValue = prev + change
        // 목표값에 매우 가까우면 정확히 설정
        if (Math.abs(newValue - targetDisplayHealth) < 0.5) {
          return targetDisplayHealth
        }
        return newValue
      })
    }
  })

  // 체력바 위치와 스케일 업데이트
  useFrame(() => {
    if (fillRef.current) {
      // 체력바 길이 조정
      const healthPercent = Math.max(0, Math.min(1, currentDisplayPercent))
      fillRef.current.scale.x = healthPercent
      
      // 체력바 위치 조정 (왼쪽 정렬)
      fillRef.current.position.x = -(width * (1 - healthPercent)) / 2
    }
  })

  // 체력에 따른 색상 계산
  const getHealthColor = (percent: number): string => {
    if (percent > 0.6) return '#22c55e' // 초록색
    if (percent > 0.3) return '#f59e0b' // 주황색
    return '#ef4444' // 빨간색
  }

  // 체력이 0 이하면 렌더링하지 않음
  if (max <= 0 || current <= 0) {
    return null
  }

  return (
    <group position={position}>
      {/* 체력바 배경 */}
      <Plane
        ref={backgroundRef}
        args={[width, height]}
        rotation={[-Math.PI / 2, 0, 0]}
      >
        <meshBasicMaterial 
          color="#333333" 
          transparent={true}
          opacity={0.8}
        />
      </Plane>
      
      {/* 체력바 채우기 */}
      <Plane
        ref={fillRef}
        position={[0, 0.001, 0]}
        args={[width, height]}
        rotation={[-Math.PI / 2, 0, 0]}
      >
        <meshBasicMaterial 
          color={getHealthColor(currentDisplayPercent)}
          transparent={true}
          opacity={0.9}
        />
      </Plane>
      
      {/* 체력바 테두리 */}
      {showBorder && (
        <mesh
          ref={borderRef}
          position={[0, 0.002, 0]}
          rotation={[-Math.PI / 2, 0, 0]}
        >
          <ringGeometry args={[width * 0.48, width * 0.52, 4]} />
          <meshBasicMaterial 
            color="#ffffff" 
            transparent={true}
            opacity={0.7}
            side={THREE.DoubleSide}
          />
        </mesh>
      )}
      
      {/* 데미지 표시 효과 (체력이 감소 중일 때) */}
      {displayHealth > current && (
        <Plane
          position={[0, 0.003, 0]}
          args={[width, height]}
          rotation={[-Math.PI / 2, 0, 0]}
        >
          <meshBasicMaterial 
            color="#ff0000"
            transparent={true}
            opacity={0.3}
          />
        </Plane>
      )}
      
      {/* 회복 표시 효과 (체력이 증가 중일 때) */}
      {displayHealth < current && (
        <Plane
          position={[0, 0.003, 0]}
          args={[width, height]}
          rotation={[-Math.PI / 2, 0, 0]}
        >
          <meshBasicMaterial 
            color="#00ff00"
            transparent={true}
            opacity={0.3}
          />
        </Plane>
      )}
    </group>
  )
}