import { Position } from './MovementHandler'
import { Enemy } from './Enemy'

export type TowerType = 'basic' | 'splash' | 'slow' | 'laser'

export interface TowerConfig {
  type: TowerType
  damage: number
  range: number
  attackSpeed: number // 밀리초
  cost: number
  color: string
  size: { radius: number, height: number }
}

export class Tower {
  public readonly id: string
  public readonly type: TowerType
  public readonly config: TowerConfig
  
  private position: Position
  private lastAttackTime: number
  private level: number
  private targetEnemy: Enemy | null

  constructor(
    id: string,
    config: TowerConfig,
    position: Position
  ) {
    this.id = id
    this.type = config.type
    this.config = config
    this.position = { ...position }
    this.lastAttackTime = 0
    this.level = 1
    this.targetEnemy = null
  }

  // 매 프레임 업데이트
  update(deltaTime: number, enemies: Enemy[]): void {
    // 타겟 선택
    this.selectTarget(enemies)
    
    // 공격 시도
    this.tryAttack()
  }

  // 사거리 내 적 중에서 타겟 선택
  private selectTarget(enemies: Enemy[]): void {
    const aliveEnemies = enemies.filter(enemy => 
      enemy.isAliveStatus() && enemy.isMoving()
    )

    if (aliveEnemies.length === 0) {
      this.targetEnemy = null
      return
    }

    // 현재 타겟이 여전히 유효한지 확인
    if (this.targetEnemy && 
        this.targetEnemy.isAliveStatus() && 
        this.isInRange(this.targetEnemy)) {
      return // 현재 타겟 유지
    }

    // 새 타겟 선택 - 사거리 내에서 가장 진행률이 높은 적
    const enemiesInRange = aliveEnemies.filter(enemy => this.isInRange(enemy))
    
    if (enemiesInRange.length === 0) {
      this.targetEnemy = null
      return
    }

    // 진행률이 가장 높은 적을 타겟으로 선택
    this.targetEnemy = enemiesInRange.reduce((prev, current) => 
      current.getProgress() > prev.getProgress() ? current : prev
    )
  }

  // 적이 사거리 내에 있는지 확인
  private isInRange(enemy: Enemy): boolean {
    const enemyPos = enemy.getPosition()
    const distance = Math.sqrt(
      Math.pow(this.position.x - enemyPos.x, 2) +
      Math.pow(this.position.z - enemyPos.z, 2)
    )
    return distance <= this.config.range
  }

  // 공격 시도
  private tryAttack(): boolean {
    if (!this.targetEnemy || !this.canAttack()) {
      return false
    }

    this.performAttack()
    this.lastAttackTime = Date.now()
    return true
  }

  // 공격 가능한지 확인
  private canAttack(): boolean {
    const now = Date.now()
    return now - this.lastAttackTime >= this.config.attackSpeed
  }

  // 실제 공격 수행
  private performAttack(): void {
    if (!this.targetEnemy) return

    const damage = this.getDamage()
    const isDead = this.targetEnemy.takeDamage(damage)

    // 타입별 특수 효과
    this.applySpecialEffects()

    if (isDead) {
      this.targetEnemy = null
    }
  }

  // 타입별 특수 효과 적용
  private applySpecialEffects(): void {
    switch (this.type) {
      case 'slow':
        // 슬로우 효과 (속도 50% 감소, 2초간)
        if (this.targetEnemy) {
          const originalSpeed = this.targetEnemy.config.speed
          this.targetEnemy.setSpeed(originalSpeed * 0.5)
          setTimeout(() => {
            if (this.targetEnemy?.isAliveStatus()) {
              this.targetEnemy.setSpeed(originalSpeed)
            }
          }, 2000)
        }
        break
      
      case 'splash':
        // 스플래시 데미지 (주변 적들에게 50% 데미지)
        // TODO: 구현 필요
        break
      
      case 'laser':
        // 관통 데미지 (직선상의 모든 적에게 데미지)
        // TODO: 구현 필요
        break
    }
  }

  // 현재 데미지 계산 (레벨 보정 포함)
  private getDamage(): number {
    return this.config.damage * this.level
  }

  // 위치 반환
  getPosition(): Position {
    return { ...this.position }
  }

  // 레벨 반환
  getLevel(): number {
    return this.level
  }

  // 레벨업
  levelUp(): void {
    this.level++
  }

  // 현재 타겟 반환
  getCurrentTarget(): Enemy | null {
    return this.targetEnemy
  }

  // 마지막 공격 시간 반환 (공격 효과 표시용)
  getLastAttackTime(): number {
    return this.lastAttackTime
  }

  // 공격 중인지 확인
  isAttacking(): boolean {
    const now = Date.now()
    return now - this.lastAttackTime < 200 // 0.2초간 공격 효과 표시
  }

  // 사거리 반환
  getRange(): number {
    return this.config.range
  }

  // 업그레이드 비용 계산
  getUpgradeCost(): number {
    return this.config.cost * this.level
  }

  // 판매 가격 계산
  getSellPrice(): number {
    return Math.floor(this.config.cost * this.level * 0.7)
  }

  // 디버그용 정보
  getDebugInfo(): any {
    return {
      id: this.id,
      type: this.type,
      level: this.level,
      position: this.position,
      damage: this.getDamage(),
      range: this.config.range,
      target: this.targetEnemy?.id || 'none',
      lastAttack: Date.now() - this.lastAttackTime
    }
  }
}