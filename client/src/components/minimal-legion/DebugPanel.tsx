'use client';

import { useState, useEffect } from 'react';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';

interface DebugPanelProps {
  player: unknown;
  enemies: unknown[];
  allies: unknown[];
  rotatingOrbs?: unknown[];
}

export const DebugPanel = ({ player, enemies, allies, rotatingOrbs = [] }: DebugPanelProps) => {
  const [isVisible, setIsVisible] = useState(false);
  const [playerData, setPlayerData] = useState<Record<string, unknown>>({});
  const { game } = useMinimalLegionStore();

  useEffect(() => {
    const interval = setInterval(() => {
      if (player && typeof player === 'object') {
        const p = player as Record<string, unknown>;
        setPlayerData({
          x: (p.x as number)?.toFixed?.(2) || 'undefined',
          y: (p.y as number)?.toFixed?.(2) || 'undefined',
          active: p.active as boolean,
          visible: p.visible as boolean,
          bodyExists: !!(p.body as object),
          bodyX: (p.body as { x?: number })?.x?.toFixed?.(2) || 'no body',
          bodyY: (p.body as { y?: number })?.y?.toFixed?.(2) || 'no body',
          velocityX: (p.body as { velocity?: { x?: number } })?.velocity?.x?.toFixed?.(2) || 'no velocity',
          velocityY: (p.body as { velocity?: { y?: number } })?.velocity?.y?.toFixed?.(2) || 'no velocity',
          sceneExists: !!(p.scene as object),
          inDisplayList: (p.scene as { children?: { exists?: (obj: unknown) => boolean } })?.children?.exists?.(player) || false,
          destroyed: p.scene ? !(p.scene as { children: { exists: (obj: unknown) => boolean } }).children.exists(player) : 'no scene',
        });
      }
    }, 100);

    return () => clearInterval(interval);
  }, [player]);

  // Toggle with 'P' key
  useEffect(() => {
    const handleKeyPress = (event: KeyboardEvent) => {
      if (event.key === 'P' || event.key === 'p') {
        event.preventDefault();
        setIsVisible(!isVisible);
      }
    };

    window.addEventListener('keydown', handleKeyPress);
    return () => window.removeEventListener('keydown', handleKeyPress);
  }, [isVisible]);

  if (!isVisible) {
    return (
      <div className="absolute top-4 left-1/2 transform -translate-x-1/2 bg-black/60 text-white px-3 py-1 rounded text-xs">
        Press &apos;P&apos; for Debug Panel
      </div>
    );
  }

  return (
    <div className="absolute top-4 left-4 bg-black/80 backdrop-blur-sm text-white p-4 rounded-lg text-xs font-mono max-w-sm pointer-events-auto">
      <div className="flex justify-between items-center mb-2">
        <h3 className="text-sm font-bold text-yellow-400">Debug Panel</h3>
        <button 
          onClick={() => setIsVisible(false)}
          className="text-red-400 hover:text-red-300"
        >
          ✕
        </button>
      </div>

      <div className="space-y-2">
        {/* Player Debug Info */}
        <div className="border border-blue-400 p-2 rounded">
          <h4 className="text-blue-400 font-semibold mb-1">Player</h4>
          <div className="grid grid-cols-2 gap-1 text-xs">
            <span>Position:</span>
            <span className="text-green-300">({String(playerData.x || 'N/A')}, {String(playerData.y || 'N/A')})</span>
            
            <span>Body Pos:</span>
            <span className="text-green-300">({String(playerData.bodyX || 'N/A')}, {String(playerData.bodyY || 'N/A')})</span>
            
            <span>Velocity:</span>
            <span className="text-blue-300">({String(playerData.velocityX || 'N/A')}, {String(playerData.velocityY || 'N/A')})</span>
            
            <span>Active:</span>
            <span className={playerData.active ? 'text-green-300' : 'text-red-300'}>
              {playerData.active ? 'YES' : 'NO'}
            </span>
            
            <span>Visible:</span>
            <span className={playerData.visible ? 'text-green-300' : 'text-red-300'}>
              {playerData.visible ? 'YES' : 'NO'}
            </span>
            
            <span>Body:</span>
            <span className={playerData.bodyExists ? 'text-green-300' : 'text-red-300'}>
              {playerData.bodyExists ? 'EXISTS' : 'MISSING'}
            </span>
            
            <span>In Scene:</span>
            <span className={playerData.inDisplayList ? 'text-green-300' : 'text-red-300'}>
              {playerData.inDisplayList ? 'YES' : 'NO'}
            </span>
            
            <span>Scene:</span>
            <span className={playerData.sceneExists ? 'text-green-300' : 'text-red-300'}>
              {playerData.sceneExists ? 'EXISTS' : 'MISSING'}
            </span>
            
            <span>Health:</span>
            <span className="text-yellow-300">
              {useMinimalLegionStore.getState().player.health || 'N/A'}
            </span>
          </div>
        </div>

        {/* Game Stats */}
        <div className="border border-yellow-400 p-2 rounded">
          <h4 className="text-yellow-400 font-semibold mb-1">Game Stats</h4>
          <div className="grid grid-cols-2 gap-1 text-xs">
            <span>Enemies:</span>
            <span className="text-red-300">{enemies?.length || 0}</span>
            
            <span>Allies:</span>
            <span className="text-blue-300">{allies?.length || 0}</span>
            
            <span>Orbs:</span>
            <span className="text-cyan-300">{rotatingOrbs?.length || 0}</span>
            
            <span>Game:</span>
            <span className={game ? 'text-green-300' : 'text-red-300'}>
              {game ? 'RUNNING' : 'STOPPED'}
            </span>
          </div>
        </div>

        {/* Performance Info */}
        <div className="border border-purple-400 p-2 rounded">
          <h4 className="text-purple-400 font-semibold mb-1">Performance</h4>
          <div className="grid grid-cols-2 gap-1 text-xs">
            <span>FPS:</span>
            <span className="text-green-300">{game?.loop?.actualFps?.toFixed(1) || 'N/A'}</span>
            
            <span>Objects:</span>
            <span className="text-blue-300">
              {(enemies?.length || 0) + (allies?.length || 0) + 1}
            </span>
          </div>
        </div>

        {/* Controls */}
        <div className="border border-gray-400 p-2 rounded">
          <h4 className="text-gray-400 font-semibold mb-1">Controls</h4>
          <div className="text-xs space-y-1">
            <div>P - Toggle Debug Panel</div>
            <div>WASD - Move Player</div>
          </div>
        </div>

        {/* Warning if player is missing */}
        {(!playerData.active || !playerData.visible || !playerData.bodyExists) && (
          <div className="border border-red-500 p-2 rounded bg-red-900/30">
            <h4 className="text-red-400 font-semibold mb-1">⚠️ WARNING</h4>
            <div className="text-xs text-red-300 space-y-1">
              <div>Player object has issues!</div>
              {!playerData.active && <div>• Player not active</div>}
              {!playerData.visible && <div>• Player not visible</div>}
              {!playerData.bodyExists && <div>• Physics body missing</div>}
              {!playerData.inDisplayList && <div>• Not in display list</div>}
              {!playerData.sceneExists && <div>• Scene missing</div>}
            </div>
          </div>
        )}
        
        {/* Additional debug info */}
        <div className="border border-orange-400 p-2 rounded">
          <h4 className="text-orange-400 font-semibold mb-1">Debug Log</h4>
          <div className="text-xs text-gray-300">
            Check browser console (F12) for detailed logs
          </div>
        </div>
      </div>
    </div>
  );
};