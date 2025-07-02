export interface WaveDifficulty {
  enemyCount: number;
  enemyHealthMultiplier: number;
  enemyDamageMultiplier: number;
  enemySpeedMultiplier: number;
  spawnRate: number; // 밀리초 단위
  maxConcurrentEnemies: number;
  isBossWave: boolean;
  difficultyType: 'easy' | 'normal' | 'hard' | 'peak' | 'rest';
}

export class DifficultyManager {
  private baseEnemyCount: number = 5;
  private difficultyPattern: string[] = [
    'easy',    // 1
    'easy',    // 2
    'normal',  // 3
    'normal',  // 4
    'hard',    // 5 - 첫 번째 도전
    'rest',    // 6 - 휴식
    'normal',  // 7
    'hard',    // 8
    'hard',    // 9
    'peak',    // 10 - 보스 웨이브
    'rest',    // 11 - 휴식
    'normal',  // 12
  ];

  constructor() {
    console.log('DifficultyManager initialized with mountain-like difficulty curve');
  }

  public getWaveDifficulty(wave: number): WaveDifficulty {
    // 12웨이브 주기로 반복
    const cyclePosition = ((wave - 1) % 12);
    const difficultyType = this.difficultyPattern[cyclePosition] || 'normal';
    const cycle = Math.floor((wave - 1) / 12);
    
    // 기본값 설정
    let config: WaveDifficulty = {
      enemyCount: this.baseEnemyCount,
      enemyHealthMultiplier: 1,
      enemyDamageMultiplier: 1,
      enemySpeedMultiplier: 1,
      spawnRate: 2000,
      maxConcurrentEnemies: 15,
      isBossWave: false,
      difficultyType: difficultyType as 'easy' | 'normal' | 'hard' | 'peak' | 'rest'
    };

    // 난이도 타입별 설정
    switch (difficultyType) {
      case 'easy':
        config = {
          ...config,
          enemyCount: this.baseEnemyCount + cycle * 2,
          enemyHealthMultiplier: 0.8 + cycle * 0.1,
          enemyDamageMultiplier: 0.8 + cycle * 0.1,
          enemySpeedMultiplier: 0.9,
          spawnRate: 2500,
          maxConcurrentEnemies: 10 + cycle * 2
        };
        break;
        
      case 'normal':
        config = {
          ...config,
          enemyCount: this.baseEnemyCount + 3 + cycle * 3,
          enemyHealthMultiplier: 1 + cycle * 0.15,
          enemyDamageMultiplier: 1 + cycle * 0.15,
          enemySpeedMultiplier: 1,
          spawnRate: 2000,
          maxConcurrentEnemies: 15 + cycle * 3
        };
        break;
        
      case 'hard':
        config = {
          ...config,
          enemyCount: this.baseEnemyCount + 5 + cycle * 4,
          enemyHealthMultiplier: 1.2 + cycle * 0.2,
          enemyDamageMultiplier: 1.2 + cycle * 0.2,
          enemySpeedMultiplier: 1.1,
          spawnRate: 1500,
          maxConcurrentEnemies: 20 + cycle * 4
        };
        break;
        
      case 'peak':
        config = {
          ...config,
          enemyCount: this.baseEnemyCount + 8 + cycle * 5,
          enemyHealthMultiplier: 1.5 + cycle * 0.3,
          enemyDamageMultiplier: 1.5 + cycle * 0.3,
          enemySpeedMultiplier: 1.2,
          spawnRate: 1200,
          maxConcurrentEnemies: 25 + cycle * 5,
          isBossWave: true
        };
        break;
        
      case 'rest':
        config = {
          ...config,
          enemyCount: this.baseEnemyCount + cycle,
          enemyHealthMultiplier: 0.7 + cycle * 0.05,
          enemyDamageMultiplier: 0.7 + cycle * 0.05,
          enemySpeedMultiplier: 0.8,
          spawnRate: 3000,
          maxConcurrentEnemies: 8 + cycle
        };
        break;
    }

    // 웨이브별 추가 조정
    if (wave === 5) {
      // 5웨이브 특별 조정 - 너무 어렵지 않게
      config.enemyCount = Math.min(config.enemyCount, 12);
      config.maxConcurrentEnemies = Math.min(config.maxConcurrentEnemies, 15);
    }

    return config;
  }

  public getDifficultyDescription(wave: number): string {
    const difficulty = this.getWaveDifficulty(wave);
    const descriptions: Record<string, string> = {
      'easy': '🌱 쉬운 난이도',
      'normal': '🌿 보통 난이도',
      'hard': '🔥 어려운 난이도',
      'peak': '⚡ 최고 난이도!',
      'rest': '😌 휴식 구간'
    };
    
    return descriptions[difficulty.difficultyType] || '보통 난이도';
  }

  public getUpcomingDifficulty(currentWave: number, ahead: number = 3): string[] {
    const upcoming: string[] = [];
    for (let i = 1; i <= ahead; i++) {
      const wave = currentWave + i;
      upcoming.push(`Wave ${wave}: ${this.getDifficultyDescription(wave)}`);
    }
    return upcoming;
  }

  public shouldShowWarning(wave: number): boolean {
    const difficulty = this.getWaveDifficulty(wave);
    return difficulty.difficultyType === 'peak' || difficulty.difficultyType === 'hard';
  }

  public getWaveReward(wave: number): { experience: number; score: number } {
    const difficulty = this.getWaveDifficulty(wave);
    const baseReward = wave * 50;
    
    const multipliers: Record<string, number> = {
      'easy': 0.8,
      'normal': 1,
      'hard': 1.5,
      'peak': 2,
      'rest': 0.6
    };
    
    const multiplier = multipliers[difficulty.difficultyType] || 1;
    
    return {
      experience: Math.floor(baseReward * multiplier),
      score: Math.floor(baseReward * multiplier)
    };
  }
}