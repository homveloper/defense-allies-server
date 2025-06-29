'use client'

import React from 'react'
import { RenderData } from '@/systems/RenderSystem'

interface ECSEntityProps {
  data: RenderData
  colors: any
  isHovered?: boolean
}

export default function ECSEntity({ data, colors, isHovered = false }: ECSEntityProps) {
  const { position, type, color, size } = data
  
  if (type === 'enemy') {
    return (
      <group position={[position.x, position.y, position.z]}>
        <mesh castShadow receiveShadow>
          <cylinderGeometry args={[size.radius, size.radius, size.height, 8]} />
          <meshStandardMaterial 
            color={isHovered ? colors.primary : color} 
            emissive={isHovered ? color : undefined}
            emissiveIntensity={isHovered ? 0.3 : 0}
          />
        </mesh>
      </group>
    )
  }
  
  if (type === 'tower') {
    return (
      <group position={[position.x, position.y + size.height / 2, position.z]}>
        <mesh castShadow receiveShadow>
          <boxGeometry args={[size.radius * 2, size.height, size.radius * 2]} />
          <meshStandardMaterial 
            color={color}
            emissive={isHovered ? color : undefined}
            emissiveIntensity={isHovered ? 0.3 : 0}
          />
        </mesh>
      </group>
    )
  }
  
  return null
}