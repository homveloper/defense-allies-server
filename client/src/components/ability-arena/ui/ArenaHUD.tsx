'use client';

import { useAbilityArenaStore } from '@/store/abilityArenaStore';

export function ArenaHUD() {
  const store = useAbilityArenaStore();

  if (!store.isGameStarted || store.isGameOver) {
    return null;
  }

  return (
    <div className="absolute inset-0 pointer-events-none">
      {/* Top HUD */}
      <div className="absolute top-4 left-4 right-4 flex justify-between items-start pointer-events-auto">
        {/* Player Stats */}
        <div className="bg-black/70 rounded-lg p-3 space-y-2">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-red-500 rounded"></div>
            <span className="text-white text-sm">
              {store.player.health}/{store.player.maxHealth}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-blue-500 rounded"></div>
            <span className="text-white text-sm">
              {store.player.mana}/{store.player.maxMana}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 bg-yellow-500 rounded"></div>
            <span className="text-white text-sm">Level {store.player.level}</span>
          </div>
        </div>

        {/* Game Stats */}
        <div className="bg-black/70 rounded-lg p-3 space-y-1">
          <div className="text-white text-sm">Wave: {store.stats.waveNumber}</div>
          <div className="text-white text-sm">Kills: {store.stats.enemiesKilled}</div>
          <div className="text-white text-sm">Score: {store.stats.score}</div>
          <div className="text-white text-sm">
            Time: {Math.floor(store.stats.timeAlive / 60)}:{(store.stats.timeAlive % 60).toString().padStart(2, '0')}
          </div>
        </div>
      </div>

      {/* Ability Bar */}
      <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 pointer-events-auto">
        <div className="bg-black/70 rounded-lg p-2 flex gap-2">
          <div className="w-12 h-12 bg-gray-600 rounded border border-gray-400 flex items-center justify-center">
            <span className="text-white text-xs">Q</span>
          </div>
          <div className="w-12 h-12 bg-gray-600 rounded border border-gray-400 flex items-center justify-center">
            <span className="text-white text-xs">E</span>
          </div>
          <div className="w-12 h-12 bg-gray-600 rounded border border-gray-400 flex items-center justify-center">
            <span className="text-white text-xs">R</span>
          </div>
        </div>
      </div>

      {/* Controls Help */}
      <div className="absolute bottom-4 right-4 bg-black/70 rounded-lg p-2 text-white text-xs">
        <div>WASD: Move</div>
        <div>Space: Dash</div>
        <div>Left Click: Attack</div>
        <div>Right Click: Fireball</div>
      </div>
    </div>
  );
}