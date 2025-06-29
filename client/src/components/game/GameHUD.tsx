'use client';

import React from 'react';

interface GameHUDProps {
  gameState: any;
  selectedTowerType: string | null;
  onTowerSelect: (type: string | null) => void;
  onStartWave: () => void;
}

export default function GameHUD({ gameState, selectedTowerType, onTowerSelect, onStartWave }: GameHUDProps) {
  return (
    <div className="absolute inset-0 pointer-events-none">
      {/* Top HUD */}
      <div className="absolute top-4 left-4 right-4 flex justify-between items-start pointer-events-auto">
        <div className="bg-black/50 text-white p-4 rounded-lg">
          <h2 className="text-xl font-bold">Tower Defense</h2>
          <p className="text-sm opacity-80">Work in Progress</p>
        </div>
        
        <div className="bg-black/50 text-white p-4 rounded-lg">
          <div className="text-sm opacity-80">Wave</div>
          <div className="text-2xl font-bold">1</div>
        </div>
      </div>

      {/* Bottom HUD */}
      <div className="absolute bottom-4 left-4 right-4 pointer-events-auto">
        <div className="bg-black/50 text-white p-4 rounded-lg flex justify-center gap-4">
          <button 
            onClick={onStartWave}
            className="px-6 py-2 bg-green-600 hover:bg-green-700 rounded-lg font-semibold transition-colors"
          >
            Start Wave
          </button>
        </div>
      </div>
    </div>
  );
}