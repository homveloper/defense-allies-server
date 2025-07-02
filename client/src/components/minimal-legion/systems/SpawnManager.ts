import { 
  SpawnPattern, 
  SpawnPosition,
  CircularPattern,
  WavePattern,
  RandomPattern,
  SpiralPattern,
  CrossPattern,
  CornerPattern
} from './SpawnPattern';

export class SpawnManager {
  private patterns: SpawnPattern[] = [];
  private currentPatternIndex: number = 0;
  private patternChangeInterval: number = 15000; // 15초마다 패턴 변경
  private lastPatternChangeTime: number = 0;
  private waveBasedPatterns: Map<number, SpawnPattern[]> = new Map();

  constructor() {
    this.initializePatterns();
    this.initializeWavePatterns();
  }

  private initializePatterns() {
    // 모든 사용 가능한 패턴 초기화
    this.patterns = [
      new RandomPattern(),
      new CircularPattern(),
      new WavePattern('left'),
      new WavePattern('right'),
      new WavePattern('top'),
      new WavePattern('bottom'),
      new SpiralPattern(),
      new CrossPattern(),
      new CornerPattern()
    ];
  }

  private initializeWavePatterns() {
    // 웨이브별 추천 패턴 설정
    this.waveBasedPatterns.set(1, [new RandomPattern(), new WavePattern('left')]);
    this.waveBasedPatterns.set(2, [new CircularPattern(), new WavePattern('right')]);
    this.waveBasedPatterns.set(3, [new SpiralPattern(), new CrossPattern()]);
    this.waveBasedPatterns.set(4, [new CornerPattern(), new CircularPattern()]);
    // 5웨이브 이상은 모든 패턴 사용
  }

  public getCurrentPattern(): SpawnPattern {
    return this.patterns[this.currentPatternIndex];
  }

  public getPatternForWave(wave: number): SpawnPattern {
    const wavePatterns = this.waveBasedPatterns.get(wave);
    if (wavePatterns && wavePatterns.length > 0) {
      return wavePatterns[Math.floor(Math.random() * wavePatterns.length)];
    }
    // 특정 웨이브 패턴이 없으면 랜덤 선택
    return this.patterns[Math.floor(Math.random() * this.patterns.length)];
  }

  public shouldChangePattern(currentTime: number): boolean {
    if (currentTime - this.lastPatternChangeTime > this.patternChangeInterval) {
      this.lastPatternChangeTime = currentTime;
      return true;
    }
    return false;
  }

  public changeToNextPattern(): void {
    this.currentPatternIndex = (this.currentPatternIndex + 1) % this.patterns.length;
    console.log(`Pattern changed to: ${this.getCurrentPattern().name}`);
  }

  public changeToRandomPattern(): void {
    const newIndex = Math.floor(Math.random() * this.patterns.length);
    this.currentPatternIndex = newIndex;
    console.log(`Pattern changed to: ${this.getCurrentPattern().name}`);
  }

  public getSpawnPositions(
    count: number, 
    centerX: number, 
    centerY: number, 
    width: number, 
    height: number,
    pattern?: SpawnPattern
  ): SpawnPosition[] {
    const activePattern = pattern || this.getCurrentPattern();
    return activePattern.getPositions(count, centerX, centerY, width, height);
  }

  public spawnEnemiesWithPattern<T extends Phaser.GameObjects.GameObject>(
    scene: Phaser.Scene,
    enemyClass: new (scene: Phaser.Scene, x: number, y: number, wave: number, healthMult?: number, damageMult?: number, speedMult?: number, enemyType?: unknown) => T,
    count: number,
    wave: number,
    pattern?: SpawnPattern,
    onSpawn?: (enemy: T, position: SpawnPosition) => void,
    healthMultiplier: number = 1,
    damageMultiplier: number = 1,
    speedMultiplier: number = 1
  ): T[] {
    const enemies: T[] = [];
    const activePattern = pattern || this.getPatternForWave(wave);
    
    const positions = this.getSpawnPositions(
      count,
      scene.cameras.main.centerX,
      scene.cameras.main.centerY,
      scene.cameras.main.width,
      scene.cameras.main.height,
      activePattern
    );

    console.log(`Spawning ${count} enemies with ${activePattern.name}`);

    positions.forEach((pos, index) => {
      // 약간의 딜레이를 두고 생성하여 시각적 효과 증대
      scene.time.delayedCall(index * 100, () => {
        const enemy = new enemyClass(scene, pos.x, pos.y, wave, healthMultiplier, damageMultiplier, speedMultiplier);
        enemies.push(enemy);
        
        if (onSpawn) {
          onSpawn(enemy, pos);
        }
      });
    });

    return enemies;
  }

  public getPatternInfo(): { current: string; available: string[] } {
    return {
      current: this.getCurrentPattern().name,
      available: this.patterns.map(p => p.name)
    };
  }
}