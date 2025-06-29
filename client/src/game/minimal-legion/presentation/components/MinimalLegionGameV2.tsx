'use client';

import { useEffect, useRef } from 'react';
import { useGameEngine } from '../hooks/useGameEngine';
import { GameRenderer } from './GameRenderer';

export default function MinimalLegionGameV2() {
  const {
    gameState,
    startGame,
    updateGame,
    movePlayer,
    pauseGame,
    resumeGame,
    getDebugInfo
  } = useGameEngine();
  
  // 키 상태 및 이동 인터벌 참조
  const keysRef = useRef<{ [key: string]: boolean }>({});
  const movementIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // 키보드 입력 처리
  useEffect(() => {
    
    const handleKeyDown = (e: KeyboardEvent) => {
      const key = e.key.toLowerCase();
      
      // 키 반복 방지
      if (e.repeat) return;
      
      keysRef.current[key] = true;
      
      // ESC로 일시정지 (즉시 처리)
      if (key === 'escape') {
        if (gameState.isPlaying) {
          pauseGame();
        } else {
          resumeGame();
        }
        return;
      }
      
      // 이동 키가 처음 눌렸을 때 즉시 이동 시작
      if (['w', 'a', 's', 'd'].includes(key)) {
        updateMovement();
        
        // 연속 이동을 위한 인터벌 시작
        if (!movementIntervalRef.current) {
          movementIntervalRef.current = setInterval(updateMovement, 16); // 60fps
        }
      }
    };
    
    const handleKeyUp = (e: KeyboardEvent) => {
      const key = e.key.toLowerCase();
      keysRef.current[key] = false;
      
      // 모든 이동 키가 해제되었는지 확인
      const movementKeys = ['w', 'a', 's', 'd'];
      const anyMovementKeyPressed = movementKeys.some(k => keysRef.current[k]);
      
      if (!anyMovementKeyPressed) {
        // 모든 이동 키가 해제되면 정지
        movePlayer({ x: 0, y: 0 });
        
        // 인터벌 정리
        if (movementIntervalRef.current) {
          clearInterval(movementIntervalRef.current);
          movementIntervalRef.current = null;
        }
      } else {
        // 일부 키가 여전히 눌려있으면 이동 업데이트
        updateMovement();
      }
    };
    
    const updateMovement = () => {
      const direction = { x: 0, y: 0 };
      
      if (keysRef.current['w']) direction.y = -1;
      if (keysRef.current['s']) direction.y = 1;
      if (keysRef.current['a']) direction.x = -1;
      if (keysRef.current['d']) direction.x = 1;
      
      // 대각선 이동 정규화
      if (direction.x !== 0 && direction.y !== 0) {
        direction.x *= 0.707;
        direction.y *= 0.707;
      }
      
      movePlayer(direction);
    };
    
    // 포커스 잃을 때 모든 키 해제
    const handleBlur = () => {
      keysRef.current = {};
      movePlayer({ x: 0, y: 0 });
      if (movementIntervalRef.current) {
        clearInterval(movementIntervalRef.current);
        movementIntervalRef.current = null;
      }
    };
    
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    window.addEventListener('blur', handleBlur);
    
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
      window.removeEventListener('blur', handleBlur);
      
      if (movementIntervalRef.current) {
        clearInterval(movementIntervalRef.current);
      }
    };
  }, [gameState.isPlaying, movePlayer, pauseGame, resumeGame]);

  // 디버그 정보 로깅
  useEffect(() => {
    if (gameState.isPlaying) {
      const debugInfo = getDebugInfo();
      if (debugInfo) {
        console.log('Debug Info:', debugInfo);
      }
    }
  }, [gameState.enemies.length, getDebugInfo, gameState.isPlaying]);

  return (
    <div className="relative w-[1200px] h-[800px] bg-gray-800 rounded-lg overflow-hidden">
      {/* 게임 렌더러 */}
      <GameRenderer
        player={gameState.player}
        allies={gameState.allies}
        enemies={gameState.enemies}
        projectiles={gameState.projectiles}
        camera={gameState.camera}
        onUpdate={updateGame}
      />
      
      {/* HUD */}
      <div className="absolute top-4 left-4 right-4 flex justify-between items-start pointer-events-none">
        {/* 좌측: 플레이어 정보 */}
        <div className="bg-black/50 p-4 rounded-lg text-white">
          <div className="text-sm opacity-80">Wave {gameState.wave}</div>
          <div className="text-sm opacity-80">Score: {gameState.score}</div>
          {gameState.player && (
            <div className="text-xs opacity-60">
              HP: {gameState.player.health}/{gameState.player.maxHealth}
            </div>
          )}
        </div>

        {/* 우측: 게임 정보 */}
        <div className="bg-black/50 p-4 rounded-lg text-white text-right">
          <div className="text-sm opacity-80">Enemies: {gameState.enemies.length}</div>
          <div className="text-sm opacity-80">Allies: {gameState.allies.length}</div>
          <div className="text-sm opacity-80">Projectiles: {gameState.projectiles.length}</div>
        </div>
      </div>

      {/* 메뉴 화면 */}
      {!gameState.isPlaying && !gameState.player && (
        <div className="absolute inset-0 bg-black/80 flex items-center justify-center">
          <div className="text-center text-white">
            <h1 className="text-6xl font-bold mb-8">미니멀 군단 V2</h1>
            <p className="text-xl mb-4">3-Tier Architecture</p>
            <p className="text-sm mb-8 opacity-80">렌더링과 상태가 분리된 새로운 아키텍처</p>
            <button
              onClick={startGame}
              className="px-8 py-4 bg-blue-600 hover:bg-blue-700 rounded-lg text-xl font-semibold transition-colors pointer-events-auto"
            >
              게임 시작
            </button>
          </div>
        </div>
      )}

      {/* 일시정지 화면 */}
      {!gameState.isPlaying && gameState.player && (
        <div className="absolute inset-0 bg-black/50 flex items-center justify-center">
          <div className="bg-gray-900 p-8 rounded-lg text-white text-center">
            <h2 className="text-3xl font-bold mb-4">일시 정지</h2>
            <p className="text-sm opacity-80 mb-4">ESC로 게임을 재개하세요</p>
            <button
              onClick={resumeGame}
              className="px-6 py-3 bg-green-600 hover:bg-green-700 rounded-lg font-semibold transition-colors pointer-events-auto"
            >
              계속하기
            </button>
          </div>
        </div>
      )}

      {/* 컨트롤 가이드 */}
      <div className="absolute bottom-4 left-4 bg-black/50 p-3 rounded-lg text-white text-xs opacity-80">
        <div>🎮 WASD: 이동 (멀티 터치 지원)</div>
        <div>⏸️ ESC: 일시정지</div>
        <div className="text-green-400 text-[10px] mt-1">대각선 이동 가능</div>
      </div>
    </div>
  );
}