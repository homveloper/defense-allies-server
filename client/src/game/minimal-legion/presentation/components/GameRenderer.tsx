'use client';

import { useEffect, useRef } from 'react';
import { Entity } from '../../types/minimalLegion';

interface GameRendererProps {
  player: Entity | null;
  allies: Entity[];
  enemies: Entity[];
  projectiles: Entity[];
  camera: { x: number; y: number };
  onUpdate: (deltaTime: number) => void;
}

export const GameRenderer = ({ 
  player, 
  allies, 
  enemies, 
  projectiles, 
  camera, 
  onUpdate 
}: GameRendererProps) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const animationFrameRef = useRef<number>();
  const lastTimeRef = useRef<number>(0);

  // 렌더링 함수
  const renderEntity = (ctx: CanvasRenderingContext2D, entity: Entity) => {
    const screenX = entity.position.x - camera.x;
    const screenY = entity.position.y - camera.y;
    
    // 화면 밖 컬링
    if (screenX < -50 || screenX > 1250 || screenY < -50 || screenY > 850) {
      return;
    }
    
    ctx.save();
    ctx.fillStyle = entity.color;
    ctx.translate(screenX, screenY);
    
    if (entity.type === 'projectile') {
      // 투사체를 더 크고 명확하게 렌더링
      ctx.beginPath();
      ctx.arc(0, 0, Math.max(entity.size, 8), 0, Math.PI * 2); // 최소 8px 크기
      ctx.fill();
      
      // 투사체 외곽선 추가
      ctx.strokeStyle = '#FFFFFF';
      ctx.lineWidth = 1;
      ctx.stroke();
      
    } else {
      // 캐릭터는 사각형으로 렌더링
      ctx.fillRect(-entity.size / 2, -entity.size / 2, entity.size, entity.size);
      
      // 체력바 (체력이 감소한 경우만)
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

  // 그리드 렌더링
  const renderGrid = (ctx: CanvasRenderingContext2D, canvas: HTMLCanvasElement) => {
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
  };

  // 디버그 정보 렌더링
  const renderDebugInfo = (ctx: CanvasRenderingContext2D) => {
    ctx.fillStyle = 'rgba(0, 0, 0, 0.8)';
    ctx.fillRect(10, 10, 250, 140);
    
    ctx.fillStyle = '#FFFFFF';
    ctx.font = '11px monospace';
    ctx.fillText(`Enemies: ${enemies.length}`, 15, 25);
    ctx.fillText(`Allies: ${allies.length}`, 15, 40);
    ctx.fillText(`Projectiles: ${projectiles.length}`, 15, 55);
    if (player) {
      ctx.fillText(`Player: (${Math.round(player.position.x)}, ${Math.round(player.position.y)})`, 15, 70);
    }
    ctx.fillText(`Camera: (${Math.round(camera.x)}, ${Math.round(camera.y)})`, 15, 85);
    
    // 투사체 상세 정보
    if (projectiles.length > 0) {
      ctx.fillStyle = '#00FF00';
      ctx.fillText(`Projectile Details:`, 15, 105);
      const proj = projectiles[0];
      ctx.fillText(`  Pos: (${Math.round(proj.position.x)}, ${Math.round(proj.position.y)})`, 15, 120);
      ctx.fillText(`  Color: ${proj.color}, Size: ${proj.size}`, 15, 135);
    }
  };

  // 게임 루프
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    
    const gameLoop = (currentTime: number) => {
      const deltaTime = (currentTime - lastTimeRef.current) / 1000;
      lastTimeRef.current = currentTime;
      
      // 게임 업데이트
      onUpdate(deltaTime);
      
      // 화면 클리어
      ctx.fillStyle = '#1F2937';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      
      // 그리드 렌더링
      renderGrid(ctx, canvas);
      
      // 엔티티 렌더링 (순서 중요)
      enemies.forEach(enemy => renderEntity(ctx, enemy));
      allies.forEach(ally => renderEntity(ctx, ally));
      
      projectiles.forEach(projectile => renderEntity(ctx, projectile));
      
      if (player) renderEntity(ctx, player);
      
      // 디버그 정보
      renderDebugInfo(ctx);
      
      animationFrameRef.current = requestAnimationFrame(gameLoop);
    };
    
    animationFrameRef.current = requestAnimationFrame(gameLoop);
    
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
      }
    };
  }, [player, allies, enemies, projectiles, camera, onUpdate]);

  return (
    <canvas
      ref={canvasRef}
      width={1200}
      height={800}
      className="absolute inset-0"
    />
  );
};