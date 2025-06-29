import { Entity } from '../types/minimalLegion';

export class ConversionSystem {
  private conversionRate = 0.3; // 30% 기본 변환 확률

  constructor() {}

  /**
   * 죽은 적들을 찾아 반환
   */
  checkForDeadEnemies(enemies: Entity[]): Entity[] {
    return enemies.filter(enemy => enemy.health <= 0);
  }

  /**
   * 적을 아군으로 변환
   */
  convertEnemyToAlly(deadEnemy: Entity, spawnAlly: (ally: Entity) => void): void {
    const ally: Entity = {
      ...deadEnemy,
      id: this.generateAllyId(),
      type: 'ally',
      velocity: { x: 0, y: 0 },
      health: deadEnemy.maxHealth, // 풀 체력으로 부활
      damage: this.calculateAllyDamage(deadEnemy.damage),
      speed: this.getAllySpeed(),
      color: this.getAllyColor()
    };

    spawnAlly(ally);
  }

  private generateAllyId(): string {
    return `ally_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  private calculateAllyDamage(enemyDamage: number): number {
    const ALLY_DAMAGE_RATIO = 0.75;
    return Math.floor(enemyDamage * ALLY_DAMAGE_RATIO);
  }

  private getAllySpeed(): number {
    return 4; // 아군 기본 속도
  }

  private getAllyColor(): string {
    return '#3B82F6'; // 아군 색상
  }

  /**
   * 적이 변환되어야 하는지 확률 체크
   */
  shouldConvertEnemy(deadEnemy: Entity): boolean {
    return Math.random() < this.conversionRate;
  }

  /**
   * 변환 확률 설정
   */
  setConversionRate(rate: number): void {
    this.conversionRate = Math.max(0, Math.min(1, rate)); // 0-1 사이로 제한
  }

  /**
   * 전체 적 변환 프로세스 처리
   */
  processEnemyConversions(
    enemies: Entity[], 
    spawnAlly: (ally: Entity) => void
  ): { remainingEnemies: Entity[]; convertedEnemyIds: string[] } {
    const deadEnemies = this.checkForDeadEnemies(enemies);
    const convertedEnemyIds: string[] = [];

    // 죽은 적들 중 변환될 적들 처리
    deadEnemies.forEach(deadEnemy => {
      if (this.shouldConvertEnemy(deadEnemy)) {
        this.convertEnemyToAlly(deadEnemy, spawnAlly);
        convertedEnemyIds.push(deadEnemy.id);
      }
    });

    // 살아있는 적들만 반환
    const remainingEnemies = enemies.filter(enemy => enemy.health > 0);

    return {
      remainingEnemies,
      convertedEnemyIds
    };
  }
}