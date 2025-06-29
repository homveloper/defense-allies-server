'use client';

import React, { useRef, useEffect, useState, useCallback } from 'react';
import { Application, Graphics, Container } from 'pixi.js';
import { GameService } from '@/game/minimal-legion/application/services/GameService';

interface PixiGameSceneProps {
  selectedTowerType: string | null;
  gameStateHook: any;
}

export default function PixiGameScene({ selectedTowerType, gameStateHook }: PixiGameSceneProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const appRef = useRef<Application | null>(null);
  const gameServiceRef = useRef<GameService | null>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  // 게임 오브젝트 컨테이너들
  const containerRefs = useRef({
    background: null as Container | null,
    player: null as Graphics | null,
    enemies: new Map<string, Graphics>(),
    allies: new Map<string, Graphics>(),
    projectiles: new Map<string, Graphics>(),
    healthBars: new Map<string, Graphics>(),
  });

  // 스프라이트 풀링 시스템
  const spritePools = useRef({
    enemies: [] as Graphics[],
    allies: [] as Graphics[],
    projectiles: [] as Graphics[],
    healthBars: [] as Graphics[],
    particles: [] as Graphics[],
  });

  // 파티클 시스템
  const particles = useRef<Array<{
    id: string;
    graphic: Graphics;
    x: number;
    y: number;
    vx: number;
    vy: number;
    life: number;
    maxLife: number;
    color: number;
    size: number;
  }>>([]);

  const lastUpdateTime = useRef(0);
  const animationFrameRef = useRef<number>();

  // 스프라이트 풀 헬퍼 함수들
  const getFromPool = (poolName: keyof typeof spritePools.current): Graphics => {
    const pool = spritePools.current[poolName];
    if (pool.length > 0) {
      return pool.pop()!;
    }
    return new Graphics();
  };

  const returnToPool = (sprite: Graphics, poolName: keyof typeof spritePools.current): void => {
    sprite.clear();
    sprite.visible = false;
    if (appRef.current) {
      appRef.current.stage.removeChild(sprite);
    }
    spritePools.current[poolName].push(sprite);
  };

  // 파티클 효과 함수들
  const createParticleEffect = (x: number, y: number, color: number, count: number = 8) => {
    if (!appRef.current) return;

    for (let i = 0; i < count; i++) {
      const angle = (Math.PI * 2 * i) / count;
      const speed = 50 + Math.random() * 100;
      const vx = Math.cos(angle) * speed;
      const vy = Math.sin(angle) * speed;
      
      const graphic = getFromPool('particles');
      graphic.clear();
      graphic.circle(0, 0, 2 + Math.random() * 3);
      graphic.fill(color);
      graphic.visible = true;
      appRef.current.stage.addChild(graphic);

      const particle = {
        id: `particle-${Date.now()}-${i}`,
        graphic,
        x,
        y,
        vx,
        vy,
        life: 0.5 + Math.random() * 0.5, // 0.5~1초
        maxLife: 1,
        color,
        size: 2 + Math.random() * 3
      };

      particles.current.push(particle);
    }
  };

  const updateParticles = (deltaTime: number) => {
    for (let i = particles.current.length - 1; i >= 0; i--) {
      const particle = particles.current[i];
      
      // 위치 업데이트
      particle.x += particle.vx * deltaTime;
      particle.y += particle.vy * deltaTime;
      particle.life -= deltaTime;
      
      // 그래픽 업데이트
      particle.graphic.x = particle.x;
      particle.graphic.y = particle.y;
      particle.graphic.alpha = particle.life / particle.maxLife;

      // 수명이 다한 파티클 제거
      if (particle.life <= 0) {
        returnToPool(particle.graphic, 'particles');
        particles.current.splice(i, 1);
      }
    }
  };

  // PixiJS 앱 초기화
  const initializePixi = useCallback(async () => {
    if (!canvasRef.current || appRef.current) return;

    try {
      // PixiJS Application 생성
      const app = new Application();
      await app.init({
        canvas: canvasRef.current,
        width: 1200,
        height: 800,
        backgroundColor: 0x1a1a1a,
        antialias: true,
        autoDensity: true,
        resolution: window.devicePixelRatio || 1,
      });

      appRef.current = app;

      // 배경 및 컨테이너 설정
      setupContainers(app);

      // 게임 서비스 초기화
      gameServiceRef.current = new GameService();
      gameServiceRef.current.initialize();

      // 입력 처리 설정
      setupInputHandlers();

      // 게임 루프 시작
      startGameLoop();

      setIsInitialized(true);
      console.log('PixiJS Game initialized');

    } catch (error) {
      console.error('Failed to initialize PixiJS:', error);
    }
  }, []);

  const setupContainers = (app: Application) => {
    // 배경 그리드 생성
    const background = new Container();
    const grid = new Graphics();
    
    // 배경색
    grid.rect(0, 0, 1200, 800);
    grid.fill(0x1a1a1a);
    
    // 격자 그리기
    grid.stroke({ width: 1, color: 0x333333, alpha: 0.5 });
    for (let x = 0; x <= 1200; x += 50) {
      grid.moveTo(x, 0);
      grid.lineTo(x, 800);
    }
    for (let y = 0; y <= 800; y += 50) {
      grid.moveTo(0, y);
      grid.lineTo(1200, y);
    }

    background.addChild(grid);
    app.stage.addChild(background);
    
    containerRefs.current.background = background;
  };

  const setupInputHandlers = () => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!gameServiceRef.current) return;

      const direction = { x: 0, y: 0 };
      
      switch (event.code) {
        case 'KeyW':
        case 'ArrowUp':
          direction.y = -1;
          break;
        case 'KeyS':
        case 'ArrowDown':
          direction.y = 1;
          break;
        case 'KeyA':
        case 'ArrowLeft':
          direction.x = -1;
          break;
        case 'KeyD':
        case 'ArrowRight':
          direction.x = 1;
          break;
      }

      if (direction.x !== 0 || direction.y !== 0) {
        gameServiceRef.current.movePlayer(direction);
      }
    };

    const handleKeyUp = (event: KeyboardEvent) => {
      if (!gameServiceRef.current) return;

      // WASD 키가 떼어졌을 때 이동 중지
      if (['KeyW', 'KeyS', 'KeyA', 'KeyD', 'ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight'].includes(event.code)) {
        gameServiceRef.current.movePlayer({ x: 0, y: 0 });
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);

    // 정리 함수에서 이벤트 리스너 제거
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  };

  const startGameLoop = () => {
    lastUpdateTime.current = Date.now();

    const gameLoop = () => {
      if (!gameServiceRef.current || !appRef.current) return;

      const now = Date.now();
      const deltaTime = (now - lastUpdateTime.current) / 1000;
      lastUpdateTime.current = now;

      // 게임 상태 업데이트
      gameServiceRef.current.update(deltaTime);
      
      // 렌더링 업데이트
      updateRendering();

      // 파티클 업데이트
      updateParticles(deltaTime);

      // React 상태 동기화
      syncWithReactState();

      animationFrameRef.current = requestAnimationFrame(gameLoop);
    };

    gameLoop();
  };

  const updateRendering = () => {
    if (!gameServiceRef.current || !appRef.current) return;

    const gameState = gameServiceRef.current.getGameState();

    // 플레이어 렌더링
    updatePlayerRendering(gameState.player);

    // 적들 렌더링
    updateEnemiesRendering(gameState.enemies);

    // 아군들 렌더링
    updateAlliesRendering(gameState.allies);

    // 투사체들 렌더링
    updateProjectilesRendering(gameState.projectiles);

    // 사용하지 않는 스프라이트 정리
    cleanupUnusedSprites(gameState);
  };

  const updatePlayerRendering = (player: any) => {
    if (!appRef.current || !player) return;

    if (!containerRefs.current.player) {
      const playerGraphic = new Graphics();
      playerGraphic.circle(0, 0, 10);
      playerGraphic.fill(0x3B82F6); // 파란색
      appRef.current.stage.addChild(playerGraphic);
      containerRefs.current.player = playerGraphic;
    }

    // 플레이어 위치 업데이트 (화면 중앙 기준)
    const playerGraphic = containerRefs.current.player;
    playerGraphic.x = player.position.x + 600; // 화면 중앙 (1200/2)
    playerGraphic.y = player.position.y + 400; // 화면 중앙 (800/2)
  };

  const updateEnemiesRendering = (enemies: any[]) => {
    if (!appRef.current) return;

    for (const enemy of enemies) {
      let enemyGraphic = containerRefs.current.enemies.get(enemy.id);

      if (!enemyGraphic) {
        enemyGraphic = getFromPool('enemies');
        enemyGraphic.circle(0, 0, enemy.size / 2);
        enemyGraphic.fill(parseInt(enemy.color.replace('#', '0x')));
        enemyGraphic.visible = true;
        appRef.current.stage.addChild(enemyGraphic);
        containerRefs.current.enemies.set(enemy.id, enemyGraphic);
      }

      // 위치 업데이트
      enemyGraphic.x = enemy.position.x + 600;
      enemyGraphic.y = enemy.position.y + 400;

      // 체력바 렌더링
      renderHealthBar(enemy, enemy.position.x + 600, enemy.position.y + 400 - 15);
    }
  };

  const updateAlliesRendering = (allies: any[]) => {
    if (!appRef.current) return;

    for (const ally of allies) {
      let allyGraphic = containerRefs.current.allies.get(ally.id);

      if (!allyGraphic) {
        allyGraphic = getFromPool('allies');
        allyGraphic.circle(0, 0, ally.size / 2);
        allyGraphic.fill(parseInt(ally.color.replace('#', '0x')));
        allyGraphic.visible = true;
        appRef.current.stage.addChild(allyGraphic);
        containerRefs.current.allies.set(ally.id, allyGraphic);
      }

      // 위치 업데이트
      allyGraphic.x = ally.position.x + 600;
      allyGraphic.y = ally.position.y + 400;

      // 체력바 렌더링
      renderHealthBar(ally, ally.position.x + 600, ally.position.y + 400 - 15);
    }
  };

  const updateProjectilesRendering = (projectiles: any[]) => {
    if (!appRef.current) return;

    for (const projectile of projectiles) {
      let projectileGraphic = containerRefs.current.projectiles.get(projectile.id);

      if (!projectileGraphic) {
        projectileGraphic = getFromPool('projectiles');
        projectileGraphic.circle(0, 0, 2);
        projectileGraphic.fill(parseInt(projectile.color.replace('#', '0x')));
        projectileGraphic.visible = true;
        appRef.current.stage.addChild(projectileGraphic);
        containerRefs.current.projectiles.set(projectile.id, projectileGraphic);
      }

      // 위치 업데이트
      projectileGraphic.x = projectile.position.x + 600;
      projectileGraphic.y = projectile.position.y + 400;
    }
  };

  const cleanupUnusedSprites = (gameState: any) => {
    if (!appRef.current) return;

    // 현재 존재하는 엔티티 ID 수집
    const currentEnemyIds = new Set(gameState.enemies.map((e: any) => e.id));
    const currentAllyIds = new Set(gameState.allies.map((a: any) => a.id));
    const currentProjectileIds = new Set(gameState.projectiles.map((p: any) => p.id));

    // 사용하지 않는 적 스프라이트 제거
    for (const [id, graphic] of containerRefs.current.enemies.entries()) {
      if (!currentEnemyIds.has(id)) {
        // 적 사망 파티클 효과
        createParticleEffect(graphic.x, graphic.y, 0xFF4444, 6);
        returnToPool(graphic, 'enemies');
        containerRefs.current.enemies.delete(id);
      }
    }

    // 사용하지 않는 아군 스프라이트 제거
    for (const [id, graphic] of containerRefs.current.allies.entries()) {
      if (!currentAllyIds.has(id)) {
        // 아군 사망 파티클 효과
        createParticleEffect(graphic.x, graphic.y, 0x44FF44, 6);
        returnToPool(graphic, 'allies');
        containerRefs.current.allies.delete(id);
      }
    }

    // 사용하지 않는 투사체 스프라이트 제거
    for (const [id, graphic] of containerRefs.current.projectiles.entries()) {
      if (!currentProjectileIds.has(id)) {
        // 투사체 충돌 파티클 효과 (작은 효과)
        createParticleEffect(graphic.x, graphic.y, 0xFFFF44, 3);
        returnToPool(graphic, 'projectiles');
        containerRefs.current.projectiles.delete(id);
      }
    }

    // 사용하지 않는 체력바 제거
    const allEntityIds = new Set([
      ...currentEnemyIds,
      ...currentAllyIds
    ]);
    
    for (const [id, graphic] of containerRefs.current.healthBars.entries()) {
      if (!allEntityIds.has(id)) {
        returnToPool(graphic, 'healthBars');
        containerRefs.current.healthBars.delete(id);
      }
    }
  };

  const renderHealthBar = (entity: any, x: number, y: number) => {
    if (!appRef.current) return;

    const healthPercent = entity.health.currentValue / entity.health.maximumValue;
    if (healthPercent >= 1) return; // 체력이 가득 찬 경우 렌더링하지 않음

    let healthBarGraphic = containerRefs.current.healthBars.get(entity.id);

    if (!healthBarGraphic) {
      healthBarGraphic = getFromPool('healthBars');
      healthBarGraphic.visible = true;
      appRef.current.stage.addChild(healthBarGraphic);
      containerRefs.current.healthBars.set(entity.id, healthBarGraphic);
    }

    // 체력바 초기화
    healthBarGraphic.clear();

    const barWidth = 24;
    const barHeight = 4;
    const barX = x - barWidth / 2;
    const barY = y;

    // 배경 (빨간색 - 잃은 체력)
    healthBarGraphic.rect(barX, barY, barWidth, barHeight);
    healthBarGraphic.fill(0xFF0000);

    // 현재 체력 (초록색)
    if (healthPercent > 0) {
      healthBarGraphic.rect(barX, barY, barWidth * healthPercent, barHeight);
      healthBarGraphic.fill(0x00FF00);
    }

    // 테두리
    healthBarGraphic.rect(barX, barY, barWidth, barHeight);
    healthBarGraphic.stroke({ width: 1, color: 0x000000, alpha: 0.8 });

    // 위치 업데이트
    healthBarGraphic.x = 0;
    healthBarGraphic.y = 0;
  };

  const syncWithReactState = () => {
    if (!gameServiceRef.current || !gameStateHook.setGameState) return;

    const gameState = gameServiceRef.current.getGameState();
    
    // GameService 상태를 useGameState 형식으로 변환
    const convertedState = {
      wave: gameState.wave,
      health: gameState.player?.health?.currentValue || 100,
      maxHealth: gameState.player?.health?.maximumValue || 100,
      gold: gameState.score, // 점수를 골드로 매핑
      score: gameState.score,
      enemies: gameState.enemies.map((enemy: any) => ({
        id: enemy.id,
        position: [enemy.position.x, enemy.position.y, 0] as [number, number, number],
        health: enemy.health.currentValue,
        maxHealth: enemy.health.maximumValue,
        speed: enemy.speed,
        pathIndex: 0,
        value: 10,
        type: 'basic' as const
      })),
      towers: [], // 타워는 아직 구현되지 않음
      isWaveActive: gameState.enemies.length > 0,
      waveTimer: 0,
      gameStatus: gameState.isGameRunning ? 'playing' as const : 'game-over' as const
    };

    gameStateHook.setGameState(convertedState);
  };

  // 컴포넌트 마운트 시 초기화
  useEffect(() => {
    initializePixi();

    // 정리 함수
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
      
      if (appRef.current) {
        appRef.current.destroy(true, { children: true });
        appRef.current = null;
      }
    };
  }, [initializePixi]);

  return (
    <div className="absolute inset-0">
      <canvas
        ref={canvasRef}
        className="w-full h-full"
        style={{ 
          display: 'block',
          width: '100%',
          height: '100%'
        }}
      />
      
      {/* 초기화 상태 표시 */}
      {!isInitialized && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-900">
          <div className="text-white text-xl">
            Initializing PixiJS Game...
          </div>
        </div>
      )}

      {/* 디버그 정보 (개발 중에만) */}
      {process.env.NODE_ENV === 'development' && isInitialized && (
        <div className="absolute top-4 left-4 text-white text-sm bg-black bg-opacity-50 p-2 rounded">
          <div className="text-blue-400 font-bold">🛡️ LEGION SYSTEM ACTIVE</div>
          <div>Enemies: <span className="text-red-400">{gameServiceRef.current?.getGameState().enemies.length || 0}</span></div>
          <div>Army Size: <span className="text-green-400">{gameServiceRef.current?.getGameState().allies.length || 0}</span></div>
          <div>Projectiles: {gameServiceRef.current?.getGameState().projectiles.length || 0}</div>
          <div>Player Health: {gameServiceRef.current?.getGameState().player?.health?.currentValue || 0}</div>
          <div>Wave: {gameServiceRef.current?.getGameState().wave || 1}</div>
          <div className="text-yellow-300">80% conversion rate - Build your legion!</div>
        </div>
      )}
    </div>
  );
}