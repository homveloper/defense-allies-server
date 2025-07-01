'use client';

import { useMinimalLegionStore } from '@/store/minimalLegionStore';

export const LevelUpModal = () => {
  const { isLevelUpModalOpen, availableUpgrades, selectUpgrade, player } = useMinimalLegionStore();

  if (!isLevelUpModalOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white border-2 border-blue-400 rounded-lg p-6 max-w-md w-full mx-4 shadow-xl">
        <div className="text-center mb-6">
          <h2 className="text-2xl font-bold text-blue-600 mb-2">레벨 업!</h2>
          <p className="text-gray-800">레벨 {player.level}에 도달했습니다!</p>
          <p className="text-gray-600 text-sm mt-2">업그레이드를 선택하세요:</p>
        </div>
        
        <div className="space-y-3">
          {availableUpgrades.map((upgrade) => (
            <button
              key={upgrade.id}
              onClick={() => selectUpgrade(upgrade)}
              className="w-full bg-gray-50 hover:bg-blue-50 border border-gray-300 hover:border-blue-400 rounded-lg p-4 text-left transition-all duration-200 group shadow-sm hover:shadow-md"
            >
              <div className="text-blue-600 font-semibold group-hover:text-blue-700">
                {upgrade.name}
              </div>
              <div className="text-gray-600 text-sm mt-1 group-hover:text-gray-800">
                {upgrade.description}
              </div>
            </button>
          ))}
        </div>
        
        <div className="text-center mt-4 text-gray-500 text-xs">
          클릭하여 업그레이드를 선택하세요
        </div>
      </div>
    </div>
  );
};