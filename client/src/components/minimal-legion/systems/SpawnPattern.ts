export interface SpawnPosition {
  x: number;
  y: number;
}

export interface SpawnPattern {
  name: string;
  description: string;
  getPositions: (count: number, centerX: number, centerY: number, width: number, height: number) => SpawnPosition[];
}

export class CircularPattern implements SpawnPattern {
  name = '원형 패턴';
  description = '플레이어 주변에 원형으로 생성';

  getPositions(count: number, centerX: number, centerY: number, width: number, height: number): SpawnPosition[] {
    const positions: SpawnPosition[] = [];
    const radius = 300; // 원의 반지름
    
    for (let i = 0; i < count; i++) {
      const angle = (i / count) * Math.PI * 2;
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      
      // 화면 밖으로 나가지 않도록 제한
      positions.push({
        x: Math.max(50, Math.min(width - 50, x)),
        y: Math.max(50, Math.min(height - 50, y))
      });
    }
    
    return positions;
  }
}

export class WavePattern implements SpawnPattern {
  name = '파도 패턴';
  description = '한쪽에서 다른쪽으로 이동하며 생성';
  
  constructor(private direction: 'left' | 'right' | 'top' | 'bottom' = 'left') {}

  getPositions(count: number, _centerX: number, _centerY: number, width: number, height: number): SpawnPosition[] {
    const positions: SpawnPosition[] = [];
    
    switch (this.direction) {
      case 'left':
        for (let i = 0; i < count; i++) {
          positions.push({
            x: -50,
            y: (height / (count + 1)) * (i + 1)
          });
        }
        break;
      case 'right':
        for (let i = 0; i < count; i++) {
          positions.push({
            x: width + 50,
            y: (height / (count + 1)) * (i + 1)
          });
        }
        break;
      case 'top':
        for (let i = 0; i < count; i++) {
          positions.push({
            x: (width / (count + 1)) * (i + 1),
            y: -50
          });
        }
        break;
      case 'bottom':
        for (let i = 0; i < count; i++) {
          positions.push({
            x: (width / (count + 1)) * (i + 1),
            y: height + 50
          });
        }
        break;
    }
    
    return positions;
  }
}

export class RandomPattern implements SpawnPattern {
  name = '무작위 패턴';
  description = '화면 가장자리에 무작위로 생성';

  getPositions(count: number, _centerX: number, _centerY: number, width: number, height: number): SpawnPosition[] {
    const positions: SpawnPosition[] = [];
    
    for (let i = 0; i < count; i++) {
      const side = Math.floor(Math.random() * 4);
      let x, y;
      
      switch (side) {
        case 0: // Top
          x = Math.random() * width;
          y = -50;
          break;
        case 1: // Right
          x = width + 50;
          y = Math.random() * height;
          break;
        case 2: // Bottom
          x = Math.random() * width;
          y = height + 50;
          break;
        default: // Left
          x = -50;
          y = Math.random() * height;
      }
      
      positions.push({ x, y });
    }
    
    return positions;
  }
}

export class SpiralPattern implements SpawnPattern {
  name = '나선형 패턴';
  description = '중심에서 나선형으로 퍼지며 생성';

  getPositions(count: number, centerX: number, centerY: number, width: number, height: number): SpawnPosition[] {
    const positions: SpawnPosition[] = [];
    const maxRadius = Math.min(width, height) / 3;
    
    for (let i = 0; i < count; i++) {
      const angle = (i / 3) * Math.PI * 2;
      const radius = (i / count) * maxRadius + 100;
      const x = centerX + Math.cos(angle) * radius;
      const y = centerY + Math.sin(angle) * radius;
      
      positions.push({
        x: Math.max(50, Math.min(width - 50, x)),
        y: Math.max(50, Math.min(height - 50, y))
      });
    }
    
    return positions;
  }
}

export class CrossPattern implements SpawnPattern {
  name = '십자 패턴';
  description = '십자 모양으로 생성';

  getPositions(count: number, centerX: number, centerY: number, width: number, height: number): SpawnPosition[] {
    const positions: SpawnPosition[] = [];
    const spacing = 100;
    const perArm = Math.floor(count / 4);
    const remainder = count % 4;
    
    // 상
    for (let i = 0; i < perArm; i++) {
      positions.push({
        x: centerX,
        y: centerY - (i + 1) * spacing
      });
    }
    
    // 우
    for (let i = 0; i < perArm; i++) {
      positions.push({
        x: centerX + (i + 1) * spacing,
        y: centerY
      });
    }
    
    // 하
    for (let i = 0; i < perArm; i++) {
      positions.push({
        x: centerX,
        y: centerY + (i + 1) * spacing
      });
    }
    
    // 좌
    for (let i = 0; i < perArm + remainder; i++) {
      positions.push({
        x: centerX - (i + 1) * spacing,
        y: centerY
      });
    }
    
    // 화면 범위 제한
    return positions.map(pos => ({
      x: Math.max(50, Math.min(width - 50, pos.x)),
      y: Math.max(50, Math.min(height - 50, pos.y))
    }));
  }
}

export class CornerPattern implements SpawnPattern {
  name = '모서리 패턴';
  description = '화면 모서리에서 생성';

  getPositions(count: number, _centerX: number, _centerY: number, width: number, height: number): SpawnPosition[] {
    const positions: SpawnPosition[] = [];
    const corners = [
      { x: 50, y: 50 },           // 좌상
      { x: width - 50, y: 50 },   // 우상
      { x: width - 50, y: height - 50 }, // 우하
      { x: 50, y: height - 50 }   // 좌하
    ];
    
    for (let i = 0; i < count; i++) {
      const corner = corners[i % 4];
      const offset = Math.floor(i / 4) * 50;
      
      positions.push({
        x: corner.x + (i % 2 === 0 ? offset : -offset),
        y: corner.y + (i % 2 === 1 ? offset : -offset)
      });
    }
    
    return positions;
  }
}