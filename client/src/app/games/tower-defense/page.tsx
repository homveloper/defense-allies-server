'use client'

import React, { useState } from 'react'
import GameHUD from '@/components/games/tower-defense/GameHUD'
import ClassBasedGameScene from '@/components/games/tower-defense/ClassBasedGameScene'
import { useGameState } from '@/hooks/useGameState'
import { ThemeProvider } from '@/contexts/ThemeContext'

export default function TowerDefensePage() {
  const [selectedTowerType, setSelectedTowerType] = useState<string | null>(null)
  const gameStateHook = useGameState()
  const { gameState, startWave } = gameStateHook

  return (
    <ThemeProvider>
      <div className="relative w-full h-screen overflow-hidden">
        {/* Class-based Game Scene */}
        <ClassBasedGameScene 
          selectedTowerType={selectedTowerType}
          gameStateHook={gameStateHook}
        />
        
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