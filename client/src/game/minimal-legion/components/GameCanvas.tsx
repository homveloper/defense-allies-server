'use client';

import { useEffect, useRef } from 'react';
import { useMinimalLegionStore } from '../useMinimalLegionStore';
import { Entity } from '../types/minimalLegion';

export default function GameCanvas() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationFrameRef = useRef<number>();
  const lastTimeRef = useRef<number>(0);
  
  const {
    gameState,
    player,
    allies,
    enemies,
    projectiles,
    camera,
    updateGame,
    movePlayer
  } = useMinimalLegionStore();

  // Handle keyboard input
  useEffect(() => {
    const keys: { [key: string]: boolean } = {};
    
    const handleKeyDown = (e: KeyboardEvent) => {
      const key = e.key.toLowerCase();
      keys[key] = true;
      
      // 즉시 이동 대응
      updateMovement();
    };
    
    const handleKeyUp = (e: KeyboardEvent) => {
      const key = e.key.toLowerCase();
      keys[key] = false;
      
      // 즉시 이동 대응
      updateMovement();
    };
    
    const updateMovement = () => {
      const direction = { x: 0, y: 0 };
      
      if (keys['w']) direction.y = -1;
      if (keys['s']) direction.y = 1;
      if (keys['a']) direction.x = -1;
      if (keys['d']) direction.x = 1;
      
      // Normalize diagonal movement
      if (direction.x !== 0 && direction.y !== 0) {
        direction.x *= 0.707;
        direction.y *= 0.707;
      }
      
      movePlayer(direction);
    };
    
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('keyup', handleKeyUp);
    
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('keyup', handleKeyUp);
    };
  }, [movePlayer]);

  // Render entities with camera offset
  const renderEntity = (ctx: CanvasRenderingContext2D, entity: Entity) => {
    // 카메라 좌표계로 변환
    const screenX = entity.position.x - camera.x;
    const screenY = entity.position.y - camera.y;
    
    // 화면 밖에 있으면 렌더링 스킵
    if (screenX < -50 || screenX > 1250 || screenY < -50 || screenY > 850) {
      return;
    }
    
    ctx.save();
    ctx.fillStyle = entity.color;
    ctx.translate(screenX, screenY);
    
    if (entity.type === 'projectile') {
      ctx.beginPath();
      ctx.arc(0, 0, entity.size / 2, 0, Math.PI * 2);
      ctx.fill();
    } else {
      // Render as square for characters
      ctx.fillRect(-entity.size / 2, -entity.size / 2, entity.size, entity.size);
      
      // Health bar
      if (entity.type !== 'projectile' && entity.health < entity.maxHealth) {
        const barWidth = entity.size;
        const barHeight = 4;
        const healthPercentage = entity.health / entity.maxHealth;
        
        ctx.fillStyle = 'rgba(0, 0, 0, 0.5)';
        ctx.fillRect(-barWidth / 2, -entity.size / 2 - 10, barWidth, barHeight);
        
        ctx.fillStyle = entity.type === 'enemy' ? '#EF4444' : '#10B981';
        ctx.fillRect(-barWidth / 2, -entity.size / 2 - 10, barWidth * healthPercentage, barHeight);
      }
    }
    
    ctx.restore();
  };

  // Game loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    const gameLoop = (currentTime: number) => {
      const deltaTime = (currentTime - lastTimeRef.current) / 1000;
      lastTimeRef.current = currentTime;
      
      if (gameState === 'playing') {
        updateGame(deltaTime);
      }
      
      // Clear canvas
      ctx.fillStyle = '#1F2937';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      
      // Draw grid pattern with camera offset
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)';
      ctx.lineWidth = 1;
      
      const gridSize = 50;
      const startX = Math.floor(camera.x / gridSize) * gridSize;
      const startY = Math.floor(camera.y / gridSize) * gridSize;
      
      for (let x = startX; x < camera.x + canvas.width; x += gridSize) {
        const screenX = x - camera.x;
        if (screenX >= 0 && screenX <= canvas.width) {
          ctx.beginPath();
          ctx.moveTo(screenX, 0);
          ctx.lineTo(screenX, canvas.height);
          ctx.stroke();
        }
      }
      
      for (let y = startY; y < camera.y + canvas.height; y += gridSize) {
        const screenY = y - camera.y;
        if (screenY >= 0 && screenY <= canvas.height) {
          ctx.beginPath();
          ctx.moveTo(0, screenY);
          ctx.lineTo(canvas.width, screenY);
          ctx.stroke();
        }
      }
      
      // Render all entities
      enemies.forEach(enemy => renderEntity(ctx, enemy));
      allies.forEach(ally => renderEntity(ctx, ally));
      projectiles.forEach(projectile => renderEntity(ctx, projectile));
      renderEntity(ctx, player);
      
      animationFrameRef.current = requestAnimationFrame(gameLoop);
    };
    
    animationFrameRef.current = requestAnimationFrame(gameLoop);
    
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [gameState, player, allies, enemies, projectiles, updateGame]);

  return (
    <canvas
      ref={canvasRef}
      width={1200}
      height={800}
      className="absolute inset-0"
    />
  );
}