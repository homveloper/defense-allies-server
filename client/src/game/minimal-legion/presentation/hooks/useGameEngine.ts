import { useRef, useEffect, useState } from 'react';
import { EnemyRepository } from '../../data/repositories/EnemyRepository';
import { GameStateRepository } from '../../data/repositories/GameStateRepository';
import { EnemyService } from '../../domain/services/EnemyService';
import { GameEngineService } from '../../domain/services/GameEngineService';
import { Entity } from '../../types/minimalLegion';

export interface GameEngineState {
  player: Entity | null;
  allies: Entity[];
  enemies: Entity[];
  projectiles: Entity[];
  camera: { x: number; y: number };
  wave: number;
  score: number;
  isPlaying: boolean;
}

export const useGameEngine = () => {
  // 의존성 주입 컨테이너
  const engineRef = useRef<{
    enemyRepository: EnemyRepository;
    gameStateRepository: GameStateRepository;
    enemyService: EnemyService;
    gameEngineService: GameEngineService;
  } | null>(null);

  const [gameState, setGameState] = useState<GameEngineState>({
    player: null,
    allies: [],
    enemies: [],
    projectiles: [],
    camera: { x: 0, y: 0 },
    wave: 1,
    score: 0,
    isPlaying: false
  });

  // 의존성 초기화
  useEffect(() => {
    if (!engineRef.current) {
      const enemyRepository = new EnemyRepository();
      const gameStateRepository = new GameStateRepository();
      const enemyService = new EnemyService(enemyRepository);
      const gameEngineService = new GameEngineService(enemyService, gameStateRepository);

      engineRef.current = {
        enemyRepository,
        gameStateRepository,
        enemyService,
        gameEngineService
      };

    }
  }, []);

  const startGame = () => {
    if (engineRef.current) {
      engineRef.current.gameEngineService.initialize();
      setGameState(prev => ({ 
        ...prev, 
        isPlaying: true,
        ...engineRef.current!.gameEngineService.getGameState()
      }));
    }
  };

  const updateGame = (deltaTime: number) => {
    if (engineRef.current && gameState.isPlaying) {
      engineRef.current.gameEngineService.update(deltaTime);
      
      // 게임 상태 동기화
      const newGameState = engineRef.current.gameEngineService.getGameState();
      
      
      setGameState(prev => ({
        ...prev,
        ...newGameState
      }));
    }
  };

  const movePlayer = (direction: { x: number; y: number }) => {
    if (engineRef.current) {
      engineRef.current.gameEngineService.movePlayer(direction);
    }
  };

  const pauseGame = () => {
    setGameState(prev => ({ ...prev, isPlaying: false }));
  };

  const resumeGame = () => {
    setGameState(prev => ({ ...prev, isPlaying: true }));
  };

  // 디버깅용 함수들
  const getDebugInfo = () => {
    if (!engineRef.current) return null;
    
    return {
      enemyCount: engineRef.current.enemyRepository.count(),
      visibleEnemies: gameState.enemies.length,
      totalEnemies: engineRef.current.enemyService.getEnemyCount(),
      playerPosition: gameState.player?.position,
      cameraPosition: gameState.camera
    };
  };

  return {
    gameState,
    startGame,
    updateGame,
    movePlayer,
    pauseGame,
    resumeGame,
    getDebugInfo
  };
};