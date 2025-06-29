import Matter from 'matter-js';
import { GameEntity } from '../../domain/entities/GameEntity';
import { Position } from '../../domain/value-objects/Position';

export interface PhysicsBody {
  id: string;
  body: Matter.Body;
  entity: GameEntity;
  type: 'player' | 'enemy' | 'ally' | 'projectile';
}

export class PhysicsService {
  private engine: Matter.Engine;
  private world: Matter.World;
  private bodies: Map<string, PhysicsBody> = new Map();
  
  // 물리 세계 설정
  private readonly worldBounds = {
    width: 1200,
    height: 800
  };

  constructor() {
    // Matter.js 엔진 초기화
    this.engine = Matter.Engine.create();
    this.world = this.engine.world;
    
    // 중력 비활성화 (탑다운 게임)
    this.engine.world.gravity.y = 0;
    this.engine.world.gravity.x = 0;
    
    this.setupWorldBounds();
    
    console.log('PhysicsService initialized with Matter.js');
  }

  // 월드 경계 설정
  private setupWorldBounds(): void {
    const thickness = 50;
    const { width, height } = this.worldBounds;
    
    // 경계벽 생성 (보이지 않는 벽)
    const walls = [
      // 상단
      Matter.Bodies.rectangle(0, -thickness/2, width, thickness, { isStatic: true }),
      // 하단  
      Matter.Bodies.rectangle(0, height + thickness/2, width, thickness, { isStatic: true }),
      // 좌측
      Matter.Bodies.rectangle(-thickness/2, height/2, thickness, height, { isStatic: true }),
      // 우측
      Matter.Bodies.rectangle(width + thickness/2, height/2, thickness, height, { isStatic: true })
    ];
    
    Matter.World.add(this.world, walls);
  }

  // 엔티티에 물리 바디 추가
  addEntity(entity: GameEntity, type: 'player' | 'enemy' | 'ally' | 'projectile'): void {
    if (this.bodies.has(entity.id)) {
      console.log(`Entity ${entity.id} already has physics body`);
      return; // 이미 존재함
    }

    let body: Matter.Body;
    const options = this.getBodyOptions(type);
    
    // 엔티티 타입에 따른 바디 생성
    if (type === 'projectile') {
      // 투사체는 작은 원형
      body = Matter.Bodies.circle(
        entity.position.x + 600, // 화면 중앙 기준 좌표 변환
        entity.position.y + 400,
        entity.size / 2,
        options
      );
    } else {
      // 다른 엔티티들은 원형
      body = Matter.Bodies.circle(
        entity.position.x + 600,
        entity.position.y + 400,
        entity.size / 2,
        options
      );
    }

    // 바디에 커스텀 데이터 추가
    body.label = `${type}_${entity.id}`;
    
    const physicsBody: PhysicsBody = {
      id: entity.id,
      body,
      entity,
      type
    };

    this.bodies.set(entity.id, physicsBody);
    Matter.World.add(this.world, body);
    
    console.log(`Added ${type} physics body for entity ${entity.id}. Total bodies: ${this.bodies.size}`);
  }

  // 타입별 바디 옵션 설정
  private getBodyOptions(type: 'player' | 'enemy' | 'ally' | 'projectile'): Matter.IBodyDefinition {
    const baseOptions: Matter.IBodyDefinition = {
      friction: 0.1,
      frictionAir: 0.01,
      restitution: 0.1, // 탄성
    };

    switch (type) {
      case 'player':
        return {
          ...baseOptions,
          density: 0.001,
          frictionAir: 0.05, // 플레이어는 좀 더 빠르게 멈춤
        };
      
      case 'enemy':
        return {
          ...baseOptions,
          density: 0.0008,
          frictionAir: 0.02,
        };
      
      case 'ally':
        return {
          ...baseOptions,
          density: 0.0008,
          frictionAir: 0.02,
        };
      
      case 'projectile':
        return {
          ...baseOptions,
          density: 0.0001,
          frictionAir: 0.001, // 투사체는 공기저항 최소
          isSensor: true, // 충돌 감지만 하고 물리적 반응 없음
        };
      
      default:
        return baseOptions;
    }
  }

  // 엔티티 제거
  removeEntity(entityId: string): void {
    const physicsBody = this.bodies.get(entityId);
    if (physicsBody) {
      Matter.World.remove(this.world, physicsBody.body);
      this.bodies.delete(entityId);
    }
  }

  // 엔티티 위치 업데이트 (게임 로직 → 물리)
  updateEntityPosition(entityId: string, position: Position): void {
    const physicsBody = this.bodies.get(entityId);
    if (physicsBody) {
      const worldX = position.x + 600;
      const worldY = position.y + 400;
      Matter.Body.setPosition(physicsBody.body, { x: worldX, y: worldY });
    }
  }

  // 엔티티 속도 설정
  setEntityVelocity(entityId: string, velocity: { x: number; y: number }): void {
    const physicsBody = this.bodies.get(entityId);
    if (physicsBody) {
      Matter.Body.setVelocity(physicsBody.body, velocity);
    }
  }

  // 물리 시뮬레이션 업데이트
  update(deltaTime: number): void {
    // Matter.js 엔진 업데이트 (60FPS 기준)
    Matter.Engine.update(this.engine, deltaTime * 1000);
    
    // 물리 바디 위치를 게임 엔티티에 동기화
    this.syncPhysicsToEntities();
  }

  // 물리 바디 위치를 게임 엔티티에 동기화
  private syncPhysicsToEntities(): void {
    for (const [entityId, physicsBody] of this.bodies) {
      const { body, entity, type } = physicsBody;
      
      // 투사체는 물리 엔진이 완전히 제어
      if (type === 'projectile') {
        const gameX = body.position.x - 600;
        const gameY = body.position.y - 400;
        entity.moveTo(new Position(gameX, gameY));
      }
      // 다른 엔티티들은 부분적으로 물리 영향 받음
      else {
        // 충돌로 인한 미세한 위치 조정만 적용
        const gameX = body.position.x - 600;
        const gameY = body.position.y - 400;
        
        // 현재 위치와 물리 위치의 차이가 클 경우에만 조정
        const currentPos = entity.position;
        const distance = Math.sqrt(
          Math.pow(gameX - currentPos.x, 2) + Math.pow(gameY - currentPos.y, 2)
        );
        
        if (distance > 5) { // 5픽셀 이상 차이날 때만 조정
          const adjustedX = currentPos.x + (gameX - currentPos.x) * 0.3; // 30%만 적용
          const adjustedY = currentPos.y + (gameY - currentPos.y) * 0.3;
          entity.moveTo(new Position(adjustedX, adjustedY));
        }
      }
    }
  }

  // 특정 엔티티 주변의 다른 엔티티들 찾기
  getNearbyEntities(entityId: string, radius: number, targetTypes?: string[]): GameEntity[] {
    const physicsBody = this.bodies.get(entityId);
    if (!physicsBody) return [];

    const centerPosition = physicsBody.body.position;
    const nearbyEntities: GameEntity[] = [];

    for (const [otherId, otherBody] of this.bodies) {
      if (otherId === entityId) continue;
      
      // 타입 필터링
      if (targetTypes && !targetTypes.includes(otherBody.type)) continue;

      const distance = Math.sqrt(
        Math.pow(otherBody.body.position.x - centerPosition.x, 2) +
        Math.pow(otherBody.body.position.y - centerPosition.y, 2)
      );

      if (distance <= radius) {
        nearbyEntities.push(otherBody.entity);
      }
    }

    return nearbyEntities;
  }

  // 충돌 이벤트 리스너 설정
  onCollisionStart(callback: (pairs: Matter.IPair[]) => void): void {
    Matter.Events.on(this.engine, 'collisionStart', (event) => {
      callback(event.pairs);
    });
  }

  onCollisionEnd(callback: (pairs: Matter.IPair[]) => void): void {
    Matter.Events.on(this.engine, 'collisionEnd', (event) => {
      callback(event.pairs);
    });
  }

  // 디버그용: 모든 물리 바디 정보 가져오기
  getDebugInfo(): { totalBodies: number; entitiesByType: Record<string, number> } {
    const entitiesByType: Record<string, number> = {};
    
    for (const physicsBody of this.bodies.values()) {
      entitiesByType[physicsBody.type] = (entitiesByType[physicsBody.type] || 0) + 1;
    }

    return {
      totalBodies: this.bodies.size,
      entitiesByType
    };
  }

  // 정리
  dispose(): void {
    Matter.World.clear(this.world, false);
    Matter.Engine.clear(this.engine);
    this.bodies.clear();
  }
}