'use client';

import { useEffect, useRef, useState } from 'react';
import { useMinimalLegionStore } from '../useMinimalLegionStore';
import GameCanvas from './GameCanvas';
import GameHUD from './GameHUD';
import UpgradePanel from './UpgradePanel';
import GameOverPanel from './GameOverPanel';

export default function MinimalLegionGame() {
  const { 
    gameState, 
    startGame, 
    pauseGame,
    resumeGame 
  } = useMinimalLegionStore();

  useEffect(() => {
    const handleKeyPress = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (gameState === 'playing') {
          pauseGame();
        } else if (gameState === 'paused') {
          resumeGame();
        }
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [gameState, pauseGame, resumeGame]);

  return (
    <div className="relative w-[1200px] h-[800px] bg-gray-800 rounded-lg overflow-hidden">
      <GameCanvas />
      <GameHUD />
      
      {gameState === 'menu' && (
        <div className="absolute inset-0 bg-black/80 flex items-center justify-center">
          <div className="text-center text-white">
            <h1 className="text-6xl font-bold mb-8">미니멀 군단</h1>
            <p className="text-xl mb-8">혼자서 시작해 거대한 군단을 만들어보세요!</p>
            <button
              onClick={startGame}
              className="px-8 py-4 bg-blue-600 hover:bg-blue-700 rounded-lg text-xl font-semibold transition-colors"
            >
              게임 시작
            </button>
          </div>
        </div>
      )}

      {gameState === 'paused' && (
        <div className="absolute inset-0 bg-black/50 flex items-center justify-center">
          <div className="bg-gray-900 p-8 rounded-lg text-white text-center">
            <h2 className="text-3xl font-bold mb-4">일시 정지</h2>
            <button
              onClick={resumeGame}
              className="px-6 py-3 bg-green-600 hover:bg-green-700 rounded-lg font-semibold transition-colors"
            >
              계속하기
            </button>
          </div>
        </div>
      )}

      {gameState === 'levelup' && <UpgradePanel />}
      {gameState === 'gameover' && <GameOverPanel />}
    </div>
  );
}