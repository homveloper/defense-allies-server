'use client';

import { useAbilityArenaStore } from '@/store/abilityArenaStore';

interface PauseModalProps {
  onResume: () => void;
  onRestart: () => void;
}

export function PauseModal({ onResume, onRestart }: PauseModalProps) {
  const store = useAbilityArenaStore();

  if (!store.isPaused || store.isAbilitySelectionOpen || store.isGameOver) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-sm w-full mx-4 text-center">
        <h2 className="text-xl font-bold mb-6">Game Paused</h2>
        
        <div className="space-y-3">
          <button
            onClick={onResume}
            className="w-full px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded transition-colors"
          >
            Resume Game
          </button>
          <button
            onClick={onRestart}
            className="w-full px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded transition-colors"
          >
            Restart Game
          </button>
          <a
            href="/games"
            className="block w-full px-4 py-2 bg-gray-600 hover:bg-gray-700 text-white rounded transition-colors"
          >
            Back to Games
          </a>
        </div>
      </div>
    </div>
  );
}