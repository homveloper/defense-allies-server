'use client';

import { useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import Phaser from 'phaser';
import { MainScene } from './scenes/MainScene';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';
import { GameHUD } from './GameHUD';

const MinimalLegionGame = () => {
  const router = useRouter();
  const gameRef = useRef<Phaser.Game | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
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
      backgroundColor: '#1a1a2e',
    };

    gameRef.current = new Phaser.Game(config);
    setGame(gameRef.current);

    return () => {
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

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-gray-900">
      <div className="relative">
        <div ref={containerRef} className="game-container" />
        <GameHUD onResumeGame={handleResumeGame} onExitGame={handleExitGame} />
      </div>
    </div>
  );
};

export default MinimalLegionGame;