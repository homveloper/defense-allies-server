'use client'

import React, { useState } from 'react'
import SettingsModal from './SettingsModal'
import { useTheme } from '@/contexts/ThemeContext'

interface GameHUDProps {
  gameState: any
  selectedTowerType: string | null
  onTowerSelect: (type: string | null) => void
  onStartWave: () => void
}

export default function GameHUD({ gameState, selectedTowerType, onTowerSelect, onStartWave }: GameHUDProps) {
  const [showSettings, setShowSettings] = useState(false)
  const { theme, colors, toggleTheme } = useTheme()

  const towers = [
    { id: 'basic', color: 'bg-blue-600', name: '기사단 요새', cost: 50 },
    { id: 'slow', color: 'bg-amber-500', name: '상인 길드', cost: 75 },
    { id: 'splash', color: 'bg-purple-600', name: '마법사 탑', cost: 100 },
    { id: 'laser', color: 'bg-green-500', name: '대성당', cost: 150 },
    { id: 'wall', color: 'bg-gray-600', name: '성벽 요새', cost: 200 },
    { id: 'support', color: 'bg-yellow-500', name: '왕궁', cost: 300 },
  ]

  return (
    <>
      {/* Top Game Info Bar */}
      <div 
        className="absolute top-11 left-0 right-0 h-[50px] backdrop-blur-sm px-6 py-2 border-b"
        style={{ 
          backgroundColor: colors.bg.overlay,
          borderColor: colors.border.primary 
        }}
      >
        <div className="flex justify-between items-center h-full">
          <div>
            <p className="text-sm font-medium" style={{ color: colors.text.primary }}>
              웨이브 {gameState.wave}/10
            </p>
            <p className="text-xs" style={{ color: colors.text.secondary }}>
              체력: {gameState.health}/{gameState.maxHealth}
            </p>
            <p className="text-xs text-amber-600">적: {gameState.enemies.length}마리</p>
          </div>
          
          <div className="text-center">
            <p className="text-sm font-medium text-amber-600">골드: {gameState.gold}</p>
            <p className="text-xs" style={{ color: colors.text.secondary }}>
              점수: {gameState.score.toLocaleString()}
            </p>
            <p className="text-xs" style={{ color: colors.text.accent }}>
              상태: {gameState.isWaveActive ? '진행중' : '대기'}
            </p>
          </div>
          
          <div className="flex items-center gap-2">
            {/* Theme Toggle Button */}
            <button 
              onClick={toggleTheme}
              className="w-8 h-8 rounded-full flex items-center justify-center transition-colors border"
              style={{ 
                backgroundColor: colors.bg.tertiary,
                borderColor: colors.border.primary 
              }}
              title={theme === 'light' ? '다크모드로 변경' : '라이트모드로 변경'}
            >
              <span className="text-sm" style={{ color: colors.text.primary }}>
                {theme === 'light' ? '🌙' : '☀️'}
              </span>
            </button>
            
            {/* Settings Button */}
            <button 
              onClick={() => setShowSettings(true)}
              className="w-8 h-8 rounded-full flex items-center justify-center transition-colors border"
              style={{ 
                backgroundColor: colors.bg.tertiary,
                borderColor: colors.border.primary 
              }}
            >
              <span className="text-sm" style={{ color: colors.text.primary }}>⚙</span>
            </button>
          </div>
        </div>
      </div>

      {/* Tower Selection Panel - Bottom Right */}
      <div 
        className="absolute bottom-4 right-4 backdrop-blur-sm rounded-lg border p-3"
        style={{ 
          backgroundColor: colors.bg.overlay,
          borderColor: colors.border.primary 
        }}
      >
        <p className="text-xs text-center font-medium mb-2" style={{ color: colors.text.primary }}>
          타워 선택
        </p>
        <div className="grid grid-cols-3 gap-1 mb-2">
          {towers.map((tower) => (
            <button
              key={tower.id}
              onClick={() => onTowerSelect(tower.id)}
              disabled={gameState.gold < tower.cost}
              className={`w-6 h-6 text-xs rounded ${tower.color} ${
                selectedTowerType === tower.id ? 'ring-1' : ''
              } ${
                gameState.gold < tower.cost ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-80 transition-all transform hover:scale-110'
              }`}
              style={{
                borderColor: selectedTowerType === tower.id ? colors.text.accent : 'transparent'
              }}
              title={`${tower.name} (${tower.cost}G)`}
            />
          ))}
        </div>
        {/* Deselect Button */}
        <button
          onClick={() => onTowerSelect(null)}
          className={`w-full h-6 text-xs rounded transition-colors flex items-center justify-center border ${
            selectedTowerType === null ? 'ring-1' : ''
          }`}
          style={{
            backgroundColor: colors.bg.tertiary,
            borderColor: colors.border.primary,
            color: colors.text.primary
          }}
          title="선택 해제"
        >
          <span className="text-sm">👆</span>
        </button>
      </div>

      {/* Selected Tower Info - Bottom Left */}
      {selectedTowerType && (
        <div 
          className="absolute bottom-4 left-4 backdrop-blur-sm rounded-lg border p-4 min-w-[140px]"
          style={{ 
            backgroundColor: colors.bg.overlay,
            borderColor: colors.border.primary 
          }}
        >
          <p className="text-xs font-medium mb-2" style={{ color: colors.text.primary }}>
            선택된 타워
          </p>
          <p className="text-sm font-medium" style={{ color: colors.text.primary }}>
            {towers.find(t => t.id === selectedTowerType)?.name}
          </p>
          <p className="text-xs mt-1 text-amber-600">
            비용: {towers.find(t => t.id === selectedTowerType)?.cost}G
          </p>
          {gameState.gold < (towers.find(t => t.id === selectedTowerType)?.cost || 0) && (
            <p className="text-xs text-red-500 mt-1">골드 부족</p>
          )}
        </div>
      )}

      {/* Wave Control - Bottom Center */}
      {!gameState.isWaveActive && gameState.enemies.length === 0 && (
        <div className="absolute bottom-20 left-1/2 transform -translate-x-1/2">
          <button
            onClick={() => {
              console.log('Wave start button clicked!')
              console.log('onStartWave function:', onStartWave)
              onStartWave()
            }}
            className="bg-green-500 hover:bg-green-600 text-white font-bold py-3 px-6 rounded-full text-base shadow-lg min-h-[48px] min-w-[120px] border border-green-400"
          >
            웨이브 {gameState.wave} 시작
          </button>
        </div>
      )}

      {/* Wave Progress Indicator */}
      {gameState.isWaveActive && (
        <div className="absolute bottom-20 left-1/2 transform -translate-x-1/2">
          <div 
            className="backdrop-blur-sm rounded-lg border px-4 py-2"
            style={{ 
              backgroundColor: colors.bg.overlay,
              borderColor: colors.border.primary 
            }}
          >
            <p className="text-xs text-center" style={{ color: colors.text.primary }}>
              웨이브 {gameState.wave} 진행중
            </p>
            <p className="text-xs text-center" style={{ color: colors.text.secondary }}>
              적: {gameState.enemies.length}마리
            </p>
          </div>
        </div>
      )}

      {/* Settings Modal */}
      {showSettings && (
        <SettingsModal onClose={() => setShowSettings(false)} />
      )}
    </>
  )
}