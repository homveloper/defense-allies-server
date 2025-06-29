import { MovementHandler, Position, Waypoint } from './MovementHandler'

export type EnemyType = 'basic' | 'fast' | 'tank'

export interface EnemyConfig {
  type: EnemyType
  health: number
  speed: number
  value: number
  color: string
  size: { radius: number, height: number }
}

export class Enemy {
  public readonly id: string
  public readonly type: EnemyType
  public readonly maxHealth: number
  public readonly value: number
  public readonly config: EnemyConfig

  private health: number
  private movementHandler: MovementHandler
  private isAlive: boolean
  private rotation: number

  constructor(
    id: string,
    config: EnemyConfig,
    startPosition: Position,
    path: Waypoint[]
  ) {
    this.id = id
    this.type = config.type
    this.maxHealth = config.health
    this.health = config.health
    this.value = config.value
    this.config = config
    this.isAlive = true
    this.rotation = 0

    // 이동 핸들러 초기화
    this.movementHandler = new MovementHandler(startPosition, config.speed, path)
  }

  // 매 프레임 업데이트
  update(deltaTime: number): void {
    if (!this.isAlive) return

    // 이동 처리
    this.movementHandler.update(deltaTime)

    // 회전 애니메이션
    const rotationSpeed = this.getRotationSpeed()
    this.rotation += rotationSpeed * (deltaTime / 1000)
  }

  // 데미지 받기
  takeDamage(damage: number): boolean {
    if (!this.isAlive) return false

    this.health = Math.max(0, this.health - damage)
    
    if (this.health <= 0) {
      this.isAlive = false
      return true // 죽음
    }
    
    return false // 생존
  }

  // 치유
  heal(amount: number): void {
    if (!this.isAlive) return
    this.health = Math.min(this.maxHealth, this.health + amount)
  }

  // 현재 위치 반환
  getPosition(): Position {
    return this.movementHandler.getPosition()
  }

  // 현재 체력 반환
  getHealth(): number {
    return this.health
  }

  // 체력 비율 반환 (0-1)
  getHealthPercent(): number {
    return this.health / this.maxHealth
  }

  // 살아있는지 확인
  isAliveStatus(): boolean {
    return this.isAlive
  }

  // 이동 중인지 확인
  isMoving(): boolean {
    return this.movementHandler.isCurrentlyMoving()
  }

  // 경로 완주했는지 확인
  hasReachedEnd(): boolean {
    return this.movementHandler.hasReachedEnd()
  }

  // 진행률 반환
  getProgress(): number {
    return this.movementHandler.getProgress()
  }

  // 현재 회전값 반환
  getRotation(): number {
    return this.rotation
  }

  // 타입별 회전 속도
  private getRotationSpeed(): number {
    switch (this.type) {
      case 'fast': return Math.PI * 2 // 빠름
      case 'tank': return Math.PI * 0.5 // 느림
      default: return Math.PI // 보통
    }
  }

  // 속도 변경 (버프/디버프용)
  setSpeed(newSpeed: number): void {
    this.movementHandler.setSpeed(newSpeed)
  }

  // 이동 중단/재개
  pauseMovement(): void {
    this.movementHandler.stop()
  }

  resumeMovement(): void {
    this.movementHandler.resume()
  }

  // 강제 사망
  kill(): void {
    this.isAlive = false
    this.health = 0
  }

  // 디버그용 정보
  getDebugInfo(): any {
    return {
      id: this.id,
      type: this.type,
      health: `${this.health}/${this.maxHealth}`,
      position: this.getPosition(),
      isAlive: this.isAlive,
      isMoving: this.isMoving(),
      progress: this.getProgress()
    }
  }
}