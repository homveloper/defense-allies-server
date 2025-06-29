'use client';

import { useMinimalLegionStore } from '../useMinimalLegionStore';

export default function GameHUD() {
  const {
    wave,
    score,
    playTime,
    player,
    playerLevel,
    experience,
    experienceToNextLevel,
    allies,
    maxAllies,
    enemies
  } = useMinimalLegionStore();

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const experiencePercentage = (experience / experienceToNextLevel) * 100;

  return (
    <>
      {/* Top Bar */}
      <div className="absolute top-0 left-0 right-0 p-4 flex justify-between items-start pointer-events-none">
        {/* Left: Player Info */}
        <div className="bg-black/50 p-4 rounded-lg text-white">
          <div className="text-sm opacity-80">레벨 {playerLevel}</div>
          <div className="mb-2">
            <div className="flex items-center gap-2">
              <div className="text-red-500">❤️</div>
              <div className="w-32 h-4 bg-gray-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-red-500 transition-all duration-300"
                  style={{ width: `${(player.health / player.maxHealth) * 100}%` }}
                />
              </div>
              <div className="text-xs">{player.health}/{player.maxHealth}</div>
            </div>
          </div>
          <div>
            <div className="flex items-center gap-2">
              <div className="text-yellow-500">⭐</div>
              <div className="w-32 h-4 bg-gray-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-yellow-500 transition-all duration-300"
                  style={{ width: `${experiencePercentage}%` }}
                />
              </div>
              <div className="text-xs">{experience}/{experienceToNextLevel}</div>
            </div>
          </div>
        </div>

        {/* Center: Wave Info */}
        <div className="bg-black/50 p-4 rounded-lg text-white text-center">
          <div className="text-2xl font-bold">웨이브 {wave}</div>
          <div className="text-sm opacity-80">남은 적: {enemies.length}</div>
        </div>

        {/* Right: Game Info */}
        <div className="bg-black/50 p-4 rounded-lg text-white text-right">
          <div className="text-sm opacity-80">군단</div>
          <div className="text-xl font-bold">{allies.length}/{maxAllies}</div>
          <div className="text-sm opacity-80 mt-2">점수</div>
          <div className="text-xl font-bold">{score.toLocaleString()}</div>
          <div className="text-sm opacity-80 mt-2">{formatTime(playTime)}</div>
        </div>
      </div>

      {/* Bottom: Abilities */}
      <div className="absolute bottom-4 left-1/2 transform -translate-x-1/2 flex gap-2">
        {[0, 1, 2, 3].map((index) => (
          <div
            key={index}
            className="w-16 h-16 bg-black/50 border-2 border-gray-600 rounded-lg flex items-center justify-center text-white"
          >
            <span className="text-xs opacity-50">{index + 1}</span>
          </div>
        ))}
      </div>
    </>
  );
}