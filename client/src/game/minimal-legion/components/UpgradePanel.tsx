'use client';

import { useState } from 'react';
import { useMinimalLegionStore } from '../useMinimalLegionStore';
import { UpgradeOption } from '../types/minimalLegion';

// Sample upgrades for demonstration
const sampleUpgrades: UpgradeOption[] = [
  {
    id: 'player_speed',
    name: '이동속도 증가',
    description: '플레이어 이동속도 +10%',
    type: 'player',
    effect: { stat: 'speed', value: 0.1, isPercentage: true }
  },
  {
    id: 'player_damage',
    name: '공격력 증가',
    description: '플레이어 공격력 +20%',
    type: 'player',
    effect: { stat: 'damage', value: 0.2, isPercentage: true }
  },
  {
    id: 'max_allies',
    name: '최대 군단 수',
    description: '최대 군단 수 +5',
    type: 'legion',
    effect: { stat: 'maxAllies', value: 5 }
  },
  {
    id: 'ally_health',
    name: '군단 체력',
    description: '모든 군단원 체력 +25%',
    type: 'legion',
    effect: { stat: 'allyHealth', value: 0.25, isPercentage: true }
  },
  {
    id: 'exp_gain',
    name: '경험치 획득량',
    description: '경험치 획득량 +20%',
    type: 'utility',
    effect: { stat: 'experienceGain', value: 0.2, isPercentage: true }
  }
];

export default function UpgradePanel() {
  const { selectUpgrade, playerLevel } = useMinimalLegionStore();
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

  // Get 3 random upgrades
  const availableUpgrades = sampleUpgrades
    .sort(() => Math.random() - 0.5)
    .slice(0, 3);

  const handleSelect = () => {
    if (selectedIndex !== null) {
      selectUpgrade(availableUpgrades[selectedIndex]);
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'player': return 'border-blue-500 bg-blue-500/10';
      case 'legion': return 'border-green-500 bg-green-500/10';
      case 'utility': return 'border-yellow-500 bg-yellow-500/10';
      default: return 'border-gray-500 bg-gray-500/10';
    }
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'player': return '플레이어';
      case 'legion': return '군단';
      case 'utility': return '유틸리티';
      default: return type;
    }
  };

  return (
    <div className="absolute inset-0 bg-black/80 flex items-center justify-center">
      <div className="bg-gray-900 p-8 rounded-lg max-w-4xl">
        <h2 className="text-3xl font-bold text-white text-center mb-2">레벨 업!</h2>
        <p className="text-gray-400 text-center mb-8">현재 레벨: {playerLevel}</p>
        
        <div className="grid grid-cols-3 gap-4 mb-8">
          {availableUpgrades.map((upgrade, index) => (
            <button
              key={upgrade.id}
              onClick={() => setSelectedIndex(index)}
              className={`
                p-6 rounded-lg border-2 transition-all cursor-pointer
                ${selectedIndex === index 
                  ? 'scale-105 shadow-lg ' + getTypeColor(upgrade.type)
                  : 'border-gray-700 bg-gray-800 hover:border-gray-600'
                }
              `}
            >
              <div className={`text-sm font-semibold mb-2 ${
                selectedIndex === index ? 'text-white' : 'text-gray-400'
              }`}>
                {getTypeLabel(upgrade.type)}
              </div>
              <h3 className="text-xl font-bold text-white mb-2">{upgrade.name}</h3>
              <p className="text-gray-300">{upgrade.description}</p>
            </button>
          ))}
        </div>
        
        <div className="text-center">
          <button
            onClick={handleSelect}
            disabled={selectedIndex === null}
            className={`
              px-8 py-3 rounded-lg font-semibold transition-all
              ${selectedIndex !== null
                ? 'bg-blue-600 hover:bg-blue-700 text-white'
                : 'bg-gray-700 text-gray-500 cursor-not-allowed'
              }
            `}
          >
            선택하기
          </button>
        </div>
      </div>
    </div>
  );
}