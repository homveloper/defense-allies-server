'use client';

import React, { useRef, useEffect, useState } from 'react';
import dynamic from 'next/dynamic';
import { GameEngineService } from '@/game/minimal-legion/domain/services/GameEngineService';
import { EnemyService } from '@/game/minimal-legion/domain/services/EnemyService';
import { EnemyRepository } from '@/game/minimal-legion/data/repositories/EnemyRepository';
import { GameStateRepository } from '@/game/minimal-legion/data/repositories/GameStateRepository';

interface PhaserGameSceneProps {
  selectedTowerType: string | null;
  gameStateHook: any;
}

export default function PhaserGameScene({ selectedTowerType, gameStateHook }: PhaserGameSceneProps) {
  const gameRef = useRef<Phaser.Game | null>(null);
  const gameSceneRef = useRef<GameScene | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isInitialized, setIsInitialized] = useState(false);

  // 기존 minimal-legion 게임 엔진 인스턴스 생성
  const gameEngineServiceRef = useRef<GameEngineService | null>(null);

  useEffect(() => {
    if (!containerRef.current || gameRef.current) return;

    const initializeGame = async () => {
      try {
        // Phaser를 동적으로 로드
        const Phaser = (await import('phaser')).default;
        
        // GameScene을 동적으로 생성
        class GameScene extends Phaser.Scene {
          private gameEngineService: any;
          private lastUpdateTime: number = 0;
          private entitySprites: Map<string, Phaser.GameObjects.Sprite> = new Map();
          private playerSprite: Phaser.GameObjects.Sprite | null = null;
          private gameStateUpdateCallback?: (gameState: any) => void;

          constructor() {
            super({ key: 'GameScene' });
          }

          init(data: { gameEngineService: any, onGameStateUpdate?: (gameState: any) => void }) {
            this.gameEngineService = data.gameEngineService;
            this.gameStateUpdateCallback = data.onGameStateUpdate;
          }

          preload() {
            this.createColorSprites();
          }

          create() {
            this.createBackground();
            
            if (this.gameEngineService) {
              this.gameEngineService.initialize();
            }

            this.setupInput();

            this.time.addEvent({
              delay: 16,
              callback: this.gameUpdate,
              callbackScope: this,
              loop: true
            });

            console.log('Phaser GameScene created and initialized');
          }

          private createColorSprites() {
            this.add.graphics()
              .fillStyle(0x3B82F6)
              .fillCircle(10, 10, 10)
              .generateTexture('player', 20, 20);

            this.add.graphics()
              .fillStyle(0xDC2626)
              .fillCircle(9, 9, 9)
              .generateTexture('enemy', 18, 18);

            this.add.graphics()
              .fillStyle(0x10B981)
              .fillCircle(8, 8, 8)
              .generateTexture('ally', 16, 16);

            this.add.graphics()
              .fillStyle(0xFBBF24)
              .fillCircle(2, 2, 2)
              .generateTexture('projectile', 4, 4);
          }

          private createBackground() {
            const background = this.add.graphics();
            background.fillStyle(0x1a1a1a);
            background.fillRect(0, 0, this.scale.width, this.scale.height);
            
            background.lineStyle(1, 0x333333, 0.5);
            for (let x = 0; x < this.scale.width; x += 50) {
              background.lineBetween(x, 0, x, this.scale.height);
            }
            for (let y = 0; y < this.scale.height; y += 50) {
              background.lineBetween(0, y, this.scale.width, y);
            }
          }

          private setupInput() {
            this.input.keyboard?.on('keydown', (event: KeyboardEvent) => {
              if (!this.gameEngineService) return;

              const direction = { x: 0, y: 0 };
              
              if (event.code === 'KeyW' || event.code === 'ArrowUp') direction.y = -1;
              if (event.code === 'KeyS' || event.code === 'ArrowDown') direction.y = 1;
              if (event.code === 'KeyA' || event.code === 'ArrowLeft') direction.x = -1;
              if (event.code === 'KeyD' || event.code === 'ArrowRight') direction.x = 1;

              if (direction.x !== 0 || direction.y !== 0) {
                this.gameEngineService.movePlayer(direction);
              }
            });
          }

          private gameUpdate() {
            if (!this.gameEngineService) return;

            const now = Date.now();
            const deltaTime = (now - this.lastUpdateTime) / 1000;
            this.lastUpdateTime = now;

            this.gameEngineService.update(deltaTime);
            
            const gameState = this.gameEngineService.getGameState();
            this.updateSprites(gameState);
            
            if (this.gameStateUpdateCallback) {
              this.gameStateUpdateCallback(gameState);
            }
          }

          private updateSprites(gameState: any) {
            this.clearUnusedSprites(gameState);

            if (gameState.player) {
              this.updatePlayerSprite(gameState.player);
            }

            gameState.enemies.forEach((enemy: any) => {
              this.updateEntitySprite(enemy, 'enemy');
            });

            gameState.allies.forEach((ally: any) => {
              this.updateEntitySprite(ally, 'ally');
            });

            gameState.projectiles.forEach((projectile: any) => {
              this.updateEntitySprite(projectile, 'projectile');
            });
          }

          private updatePlayerSprite(player: any) {
            if (!this.playerSprite) {
              this.playerSprite = this.add.sprite(
                player.position.x + this.scale.width / 2,
                player.position.y + this.scale.height / 2,
                'player'
              );
            }

            this.playerSprite.setPosition(
              player.position.x + this.scale.width / 2,
              player.position.y + this.scale.height / 2
            );
          }

          private updateEntitySprite(entity: any, textureKey: string) {
            let sprite = this.entitySprites.get(entity.id);

            if (!sprite) {
              sprite = this.add.sprite(
                entity.position.x + this.scale.width / 2,
                entity.position.y + this.scale.height / 2,
                textureKey
              );
              this.entitySprites.set(entity.id, sprite);
            }

            sprite.setPosition(
              entity.position.x + this.scale.width / 2,
              entity.position.y + this.scale.height / 2
            );
          }

          private clearUnusedSprites(gameState: any) {
            const currentEntityIds = new Set([
              ...gameState.enemies.map((e: any) => e.id),
              ...gameState.allies.map((a: any) => a.id),
              ...gameState.projectiles.map((p: any) => p.id)
            ]);

            this.entitySprites.forEach((sprite, id) => {
              if (!currentEntityIds.has(id)) {
                sprite.destroy();
                this.entitySprites.delete(id);
              }
            });
          }
        }

        // 기존 게임 엔진 서비스 초기화 (기존 구조 유지)
        const enemyRepository = new EnemyRepository();
        const gameStateRepository = new GameStateRepository();
        const enemyService = new EnemyService(enemyRepository);
        const gameEngineService = new GameEngineService(enemyService, gameStateRepository);
        
        gameEngineServiceRef.current = gameEngineService;

        // Phaser 게임 설정
        const config: Phaser.Types.Core.GameConfig = {
          type: Phaser.AUTO,
          width: 1200,
          height: 800,
          parent: containerRef.current!,
          backgroundColor: '#1a1a1a',
          physics: {
            default: 'arcade',
            arcade: {
              gravity: { x: 0, y: 0 }, // 탑뷰 게임이므로 중력 없음
              debug: false
            }
          },
          scene: GameScene,
          scale: {
            mode: Phaser.Scale.FIT,
            autoCenter: Phaser.Scale.CENTER_BOTH
          }
        };

        // Phaser 게임 인스턴스 생성
        gameRef.current = new Phaser.Game(config);

        // 게임 씬이 생성되면 게임 엔진 전달
        gameRef.current.events.once('ready', () => {
          const scene = gameRef.current?.scene.getScene('GameScene') as any;
          if (scene) {
            gameSceneRef.current = scene;
            
            // 게임 엔진 서비스와 상태 업데이트 콜백 전달
            scene.scene.restart({
              gameEngineService,
              onGameStateUpdate: (gameState: any) => {
                // React 게임 상태 훅과 동기화
                if (gameStateHook.setGameState) {
                  // minimal-legion 게임 상태를 useGameState 형식으로 변환
                  const convertedState = convertGameState(gameState);
                  gameStateHook.setGameState(convertedState);
                }
              }
            });
            
            setIsInitialized(true);
            console.log('Phaser game initialized with minimal-legion game engine');
          }
        });
      } catch (error) {
        console.error('Failed to initialize Phaser game:', error);
      }
    };

    initializeGame();

    return () => {
      if (gameRef.current) {
        gameRef.current.destroy(true);
        gameRef.current = null;
      }
    };
  }, []);

  // 게임 상태 변환 함수 (minimal-legion → useGameState 형식)
  const convertGameState = (minimalLegionState: any) => {
    return {
      wave: minimalLegionState.wave,
      health: minimalLegionState.player?.health || 100,
      maxHealth: minimalLegionState.player?.maxHealth || 100,
      gold: minimalLegionState.score, // 점수를 골드로 매핑
      score: minimalLegionState.score,
      enemies: minimalLegionState.enemies.map((enemy: any) => ({
        id: enemy.id,
        position: [enemy.position.x, enemy.position.y, 0] as [number, number, number],
        health: enemy.health,
        maxHealth: enemy.maxHealth,
        speed: enemy.speed,
        pathIndex: 0,
        value: 10,
        type: 'basic' as const
      })),
      towers: [], // 타워는 아직 구현되지 않음
      isWaveActive: minimalLegionState.enemies.length > 0,
      waveTimer: 0,
      gameStatus: minimalLegionState.player?.health > 0 ? 'playing' as const : 'game-over' as const
    };
  };

  // React에서 게임 엔진에 액세스할 수 있는 메서드들
  useEffect(() => {
    if (isInitialized && gameSceneRef.current) {
      // selectedTowerType 변경 처리 등 추가 로직
      console.log('Selected tower type:', selectedTowerType);
    }
  }, [selectedTowerType, isInitialized]);

  return (
    <div className="absolute inset-0">
      <div 
        ref={containerRef} 
        className="w-full h-full"
        style={{ 
          width: '100%', 
          height: '100%',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center'
        }}
      />
      
      {/* 초기화 상태 표시 */}
      {!isInitialized && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-900">
          <div className="text-white text-xl">
            Initializing Phaser Game Engine...
          </div>
        </div>
      )}
    </div>
  );
}