'use client'

import React, { useState } from 'react'
import GameHUD from '@/components/game/GameHUD'
import { useGameState } from '@/hooks/useGameState'
import { ThemeProvider, useTheme } from '@/contexts/ThemeContext'

// Temporary placeholder component
function GamePlaceholder() {
  const { colors } = useTheme()
  
  return (
    <div 
      className="w-full h-full flex items-center justify-center"
      style={{ backgroundColor: colors.bg.secondary }}
    >
      <div 
        className="text-center p-8 rounded-lg border"
        style={{ 
          backgroundColor: colors.bg.primary,
          borderColor: colors.border.primary,
          color: colors.text.primary 
        }}
      >
        <h2 className="text-2xl font-bold mb-4">새로운 게임 씬 준비중</h2>
        <p style={{ color: colors.text.secondary }}>
          깔끔한 새 게임 시스템을 구현하고 있습니다.
        </p>
      </div>
    </div>
  )
}

export default function GamePage() {
  const [selectedTowerType, setSelectedTowerType] = useState<string | null>(null)
  const gameStateHook = useGameState()
  const { gameState, startWave } = gameStateHook

  return (
    <ThemeProvider>
      <div className="relative w-full h-screen overflow-hidden">
        {/* Placeholder Game Scene */}
        <GamePlaceholder />
        
        {/* HUD Overlay - 유지됨 */}
        <GameHUD 
          gameState={gameState}
          selectedTowerType={selectedTowerType}
          onTowerSelect={setSelectedTowerType}
          onStartWave={startWave}
        />
      </div>
    </ThemeProvider>
  )
}