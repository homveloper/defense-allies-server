'use client';

import { useAbilityArenaStore } from '@/store/abilityArenaStore';

interface GameOverModalProps {
  onRestart: () => void;
}

export function GameOverModal({ onRestart }: GameOverModalProps) {
  const store = useAbilityArenaStore();

  if (!store.isGameOver) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4 text-center">
        <h2 className="text-2xl font-bold text-red-600 mb-4">Game Over</h2>
        
        <div className="space-y-2 mb-6">
          <div className="text-lg">Final Score: <span className="font-bold">{store.stats.score}</span></div>
          <div>Wave Reached: {store.stats.waveNumber}</div>
          <div>Enemies Killed: {store.stats.enemiesKilled}</div>
          <div>Time Survived: {Math.floor(store.stats.timeAlive / 60)}:{(store.stats.timeAlive % 60).toString().padStart(2, '0')}</div>
          <div>Damage Dealt: {store.stats.damageDealt}</div>
          <div>Abilities Used: {store.stats.abilitiesUsed}</div>
        </div>

        <div className="flex gap-3">
          <button
            onClick={onRestart}
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded transition-colors"
          >
            Play Again
          </button>
          <a
            href="/games"
            className="flex-1 px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors text-center"
          >
            Back to Games
          </a>
        </div>
      </div>
    </div>
  );
}