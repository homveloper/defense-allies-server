'use client';

import { useMinimalLegionStore } from '../useMinimalLegionStore';

export default function GameOverPanel() {
  const { 
    score, 
    wave, 
    playTime, 
    playerLevel,
    startGame 
  } = useMinimalLegionStore();

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}분 ${secs}초`;
  };

  return (
    <div className="absolute inset-0 bg-black/80 flex items-center justify-center">
      <div className="bg-gray-900 p-8 rounded-lg text-white text-center max-w-md">
        <h2 className="text-4xl font-bold mb-8">게임 오버</h2>
        
        <div className="space-y-4 mb-8">
          <div className="bg-gray-800 p-4 rounded-lg">
            <div className="text-gray-400 text-sm">최종 점수</div>
            <div className="text-3xl font-bold text-yellow-500">{score.toLocaleString()}</div>
          </div>
          
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-gray-800 p-3 rounded-lg">
              <div className="text-gray-400 text-xs">도달 웨이브</div>
              <div className="text-xl font-bold">{wave}</div>
            </div>
            <div className="bg-gray-800 p-3 rounded-lg">
              <div className="text-gray-400 text-xs">플레이 시간</div>
              <div className="text-xl font-bold">{formatTime(playTime)}</div>
            </div>
            <div className="bg-gray-800 p-3 rounded-lg">
              <div className="text-gray-400 text-xs">최종 레벨</div>
              <div className="text-xl font-bold">{playerLevel}</div>
            </div>
          </div>
        </div>
        
        <button
          onClick={startGame}
          className="px-8 py-3 bg-blue-600 hover:bg-blue-700 rounded-lg font-semibold transition-colors"
        >
          다시 시작
        </button>
      </div>
    </div>
  );
}