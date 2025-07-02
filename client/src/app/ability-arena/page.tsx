'use client';

import { useEffect, useRef, useState } from 'react';
import { ArenaHUD } from '@/components/ability-arena/ui/ArenaHUD';
import { AbilitySelectionModal } from '@/components/ability-arena/ui/AbilitySelectionModal';
import { GameOverModal } from '@/components/ability-arena/ui/GameOverModal';
import { PauseModal } from '@/components/ability-arena/ui/PauseModal';
import { useAbilityArenaStore } from '@/store/abilityArenaStore';

export default function AbilityArenaPage() {
  const gameRef = useRef<any>(null);
  const [isGameReady, setIsGameReady] = useState(false);
  const phaserRef = useRef<any>(null);
  const sceneRef = useRef<any>(null);
  const store = useAbilityArenaStore();

  useEffect(() => {
    // Dynamic imports to avoid SSR issues
    const loadPhaser = async () => {
      const [PhaserModule, SceneModule] = await Promise.all([
        import('phaser'),
        import('@/components/ability-arena/scenes/ArenaMainScene')
      ]);
      
      phaserRef.current = PhaserModule;
      sceneRef.current = SceneModule.ArenaMainScene;
      
      // Initialize game after both modules are loaded
      initializeGame();
    };

    loadPhaser();

    // Cleanup function
    return () => {
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
        setIsGameReady(false);
      }
    };
  }, []);

  const initializeGame = () => {
    if (!phaserRef.current || !sceneRef.current || gameRef.current) return;

    const config = {
      type: phaserRef.current.AUTO,
      width: 1200,
      height: 800,
      parent: 'ability-arena-game',
      backgroundColor: '#2c3e50',
      physics: {
        default: 'arcade',
        arcade: {
          gravity: { x: 0, y: 0 },
          debug: false,
        },
      },
      scene: [sceneRef.current],
      scale: {
        mode: phaserRef.current.Scale.FIT,
        autoCenter: phaserRef.current.Scale.CENTER_BOTH,
      },
    };

    gameRef.current = new phaserRef.current.Game(config);
    setIsGameReady(true);
  };

  const handleRestart = () => {
    store.resetGame();
    if (gameRef.current) {
      const scene = gameRef.current.scene.getScene('ArenaMainScene') as any;
      if (scene) {
        scene.scene.restart();
      }
    }
  };

  const handlePause = () => {
    store.setPaused(true);
    if (gameRef.current) {
      const scene = gameRef.current.scene.getScene('ArenaMainScene') as any;
      if (scene) {
        scene.scene.pause();
      }
    }
  };

  const handleResume = () => {
    store.setPaused(false);
    if (gameRef.current) {
      const scene = gameRef.current.scene.getScene('ArenaMainScene') as any;
      if (scene) {
        scene.scene.resume();
      }
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      {/* Header */}
      <div className="bg-gray-800 p-4 flex justify-between items-center">
        <h1 className="text-2xl font-bold">🏟️ Ability Arena</h1>
        <div className="flex gap-4">
          <button
            onClick={handlePause}
            className="px-4 py-2 bg-yellow-600 hover:bg-yellow-700 rounded transition-colors"
            disabled={store.isPaused || store.isGameOver}
          >
            ⏸️ Pause
          </button>
          <button
            onClick={handleRestart}
            className="px-4 py-2 bg-red-600 hover:bg-red-700 rounded transition-colors"
          >
            🔄 Restart
          </button>
          <a
            href="/games"
            className="px-4 py-2 bg-gray-600 hover:bg-gray-700 rounded transition-colors"
          >
            🏠 Back to Games
          </a>
        </div>
      </div>

      {/* Game Container */}
      <div className="relative">
        {/* Phaser Game */}
        <div 
          id="ability-arena-game" 
          className="mx-auto"
          style={{ maxWidth: '1200px', maxHeight: '800px' }}
        />

        {/* Game HUD Overlay */}
        {isGameReady && <ArenaHUD />}

        {/* Game Loading */}
        {!isGameReady && (
          <div className="absolute inset-0 bg-gray-900 flex items-center justify-center min-h-[800px]">
            <div className="text-center">
              <div className="animate-spin w-16 h-16 border-4 border-blue-500 border-t-transparent rounded-full mx-auto mb-4"></div>
              <p className="text-xl">Loading Ability Arena...</p>
              <p className="text-sm text-gray-400 mt-2">
                {!phaserRef.current ? 'Loading Phaser engine...' : 'Initializing game world...'}
              </p>
            </div>
          </div>
        )}

        {/* Modals */}
        <AbilitySelectionModal />
        <GameOverModal onRestart={handleRestart} />
        <PauseModal onResume={handleResume} onRestart={handleRestart} />
      </div>

      {/* Game Info */}
      <div className="max-w-6xl mx-auto p-6">
        <div className="bg-gray-800 rounded-lg p-6">
          <h2 className="text-xl font-bold mb-4">🎮 How to Play</h2>
          <div className="grid md:grid-cols-2 gap-6">
            <div>
              <h3 className="font-semibold mb-2">🎯 Controls</h3>
              <ul className="space-y-1 text-sm text-gray-300">
                <li><kbd className="bg-gray-700 px-1 rounded">WASD</kbd> - Move</li>
                <li><kbd className="bg-gray-700 px-1 rounded">Mouse</kbd> - Aim</li>
                <li><kbd className="bg-gray-700 px-1 rounded">Left Click</kbd> - Basic Attack</li>
                <li><kbd className="bg-gray-700 px-1 rounded">Q</kbd> - Ability 1</li>
                <li><kbd className="bg-gray-700 px-1 rounded">E</kbd> - Ability 2</li>
                <li><kbd className="bg-gray-700 px-1 rounded">R</kbd> - Ultimate</li>
                <li><kbd className="bg-gray-700 px-1 rounded">Space</kbd> - Dash</li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold mb-2">🏆 Objectives</h3>
              <ul className="space-y-1 text-sm text-gray-300">
                <li>• Survive as long as possible</li>
                <li>• Defeat enemies to gain experience</li>
                <li>• Level up to unlock new abilities</li>
                <li>• Collect power-ups for temporary boosts</li>
                <li>• Face increasingly challenging waves</li>
                <li>• Test different ability combinations</li>
              </ul>
            </div>
          </div>
          
          <div className="mt-6 p-4 bg-blue-900/50 rounded-lg">
            <h3 className="font-semibold mb-2">💡 Testing Focus</h3>
            <p className="text-sm text-blue-200">
              This arena is designed to test our Gameplay Ability System (GAS). 
              Try different ability combinations, observe visual effects, test performance with many active effects, 
              and help us identify bugs or balance issues!
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}