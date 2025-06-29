'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import GameButton from '@/components/ui/GameButton/GameButton'

interface GameCard {
  id: string
  title: string
  description: string
  path: string
  icon: string
  color: string
  available: boolean
}

const games: GameCard[] = [
  {
    id: 'minimal-legion',
    title: '미니멀 군단',
    description: '혼자서 시작해 거대한 군단을 만들어보세요!',
    path: '/minimal-legion',
    icon: '⚔️',
    color: '#3b82f6',
    available: true
  },
  {
    id: 'tower-defense',
    title: '타워 디펜스',
    description: '적의 침입을 막아라!',
    path: '/games/tower-defense',
    icon: '🏰',
    color: '#10b981',
    available: true
  },
  {
    id: 'puzzle',
    title: '퍼즐 게임',
    description: '머리를 써서 문제를 해결하세요',
    path: '/games/puzzle',
    icon: '🧩',
    color: '#8b5cf6',
    available: false
  },
  {
    id: 'arcade',
    title: '아케이드',
    description: '빠른 반응속도가 필요해요',
    path: '/games/arcade',
    icon: '🎮',
    color: '#ec4899',
    available: false
  }
]

export default function GamesPage() {
  const router = useRouter()

  const handleGameSelect = (game: GameCard) => {
    if (game.available) {
      router.push(game.path)
    }
  }

  return (
    <div className="min-h-screen flex flex-col bg-gray-900">
      {/* 헤더 */}
      <header className="p-6 border-b border-gray-700 bg-gray-800">
        <h1 className="text-3xl font-bold text-center text-white">
          미니게임 모음
        </h1>
      </header>

      {/* 게임 목록 */}
      <main className="flex-1 p-6">
        <div className="max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {games.map((game) => (
            <div
              key={game.id}
              className={`
                relative rounded-2xl p-6 transition-all duration-300 bg-gray-800
                ${game.available 
                  ? 'cursor-pointer hover:scale-105 hover:shadow-xl' 
                  : 'cursor-not-allowed opacity-50'
                }
              `}
              style={{
                border: `2px solid ${game.available ? game.color : '#374151'}`,
                boxShadow: game.available 
                  ? `0 0 20px ${game.color}33` 
                  : 'none'
              }}
              onClick={() => handleGameSelect(game)}
            >
              {/* 게임 아이콘 */}
              <div 
                className="text-6xl mb-4 text-center"
                style={{ 
                  filter: game.available ? 'none' : 'grayscale(100%)'
                }}
              >
                {game.icon}
              </div>

              {/* 게임 정보 */}
              <h3 
                className="text-xl font-bold mb-2"
                style={{ color: game.available ? game.color : '#9CA3AF' }}
              >
                {game.title}
              </h3>
              <p className="text-sm mb-4 text-gray-400">
                {game.description}
              </p>

              {/* 플레이 버튼 */}
              {game.available ? (
                <GameButton
                  onClick={() => handleGameSelect(game)}
                  variant="primary"
                  className="w-full"
                  style={{ backgroundColor: game.color }}
                >
                  플레이
                </GameButton>
              ) : (
                <div className="text-center py-2 px-4 rounded-lg bg-gray-700 text-gray-400">
                  준비중
                </div>
              )}

              {/* 뱃지 */}
              {game.available && (
                <div 
                  className="absolute top-2 right-2 px-2 py-1 rounded text-xs font-bold"
                  style={{ 
                    backgroundColor: game.id === 'minimal-legion' ? '#ef4444' : '#fbbf24',
                    color: game.id === 'minimal-legion' ? '#ffffff' : '#78350f'
                  }}
                >
                  {game.id === 'minimal-legion' ? 'NEW' : 'BETA'}
                </div>
              )}
            </div>
          ))}
        </div>

        {/* 설명 */}
        <div className="mt-12 max-w-2xl mx-auto text-center p-6 rounded-lg bg-gray-800 border border-gray-700">
          <p className="text-gray-400">
            다양한 미니게임을 즐겨보세요! 더 많은 게임이 곧 추가될 예정입니다.
          </p>
        </div>
      </main>

      {/* 하단 네비게이션 */}
      <footer className="p-4 border-t border-gray-700 bg-gray-800">
        <nav className="flex justify-center gap-6">
          <button
            onClick={() => router.push('/')}
            className="px-4 py-2 rounded-lg transition-colors bg-gray-700 text-gray-300 hover:bg-gray-600"
          >
            홈으로
          </button>
        </nav>
      </footer>
    </div>
  )
}