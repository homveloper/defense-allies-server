'use client';

import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import Phaser from 'phaser';
import { MainScene } from './scenes/MainScene';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';
import { GameHUD } from './GameHUD';
import { DebugPanel } from './DebugPanel';
import { GameOverModal } from './GameOverModal';
import { LevelUpModal } from './LevelUpModal';

const MinimalLegionGame = () => {
  const router = useRouter();
  const gameRef = useRef<Phaser.Game | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const mainSceneRef = useRef<MainScene | null>(null);
  const [debugData, setDebugData] = useState<Record<string, unknown>>({});
  const { setGame, togglePause, resetGame } = useMinimalLegionStore();

  useEffect(() => {
    if (!containerRef.current || gameRef.current) return;

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.AUTO,
      width: 1200,
      height: 800,
      parent: containerRef.current,
      physics: {
        default: 'arcade',
        arcade: {
          gravity: { x: 0, y: 0 },
          debug: false,
        },
      },
      scene: [MainScene],
      backgroundColor: '#ffffff',
    };

    gameRef.current = new Phaser.Game(config);
    setGame(gameRef.current);
    
    // Get reference to main scene for debugging
    gameRef.current.events.on('ready', () => {
      mainSceneRef.current = gameRef.current?.scene.getScene('MainScene') as MainScene;
    });

    // Debug data update interval
    const debugInterval = setInterval(() => {
      if (mainSceneRef.current) {
        const debugInfo = mainSceneRef.current.getDebugInfo();
        setDebugData(debugInfo);
      }
    }, 100);

    return () => {
      clearInterval(debugInterval);
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
        setGame(null);
      }
    };
  }, [setGame]);

  const handleResumeGame = () => {
    togglePause();
  };

  const handleExitGame = () => {
    resetGame();
    router.push('/games');
  };

  const handleRestartGame = () => {
    console.log('Restarting game...');
    resetGame();
    
    // Restart the current scene
    if (mainSceneRef.current) {
      mainSceneRef.current.scene.restart();
    }
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50">
      <div className="relative border-2 border-gray-300 rounded-lg shadow-lg">
        <div ref={containerRef} className="game-container rounded-lg overflow-hidden" />
        <GameHUD onResumeGame={handleResumeGame} onExitGame={handleExitGame} />
        <DebugPanel 
          player={debugData.player}
          enemies={(debugData.enemies as unknown[]) || []}
          allies={(debugData.allies as unknown[]) || []}
          rotatingOrbs={(debugData.rotatingOrbs as unknown[]) || []}
        />
        <GameOverModal 
          onRestart={handleRestartGame}
          onExit={handleExitGame}
        />
        <LevelUpModal />
      </div>
    </div>
  );
};

export default MinimalLegionGame;