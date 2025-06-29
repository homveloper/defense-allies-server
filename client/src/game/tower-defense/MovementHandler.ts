// 이동 처리를 담당하는 클래스
export interface Position {
  x: number
  y: number
  z: number
}

export interface Waypoint {
  x: number
  y: number
}

export class MovementHandler {
  private position: Position
  private speed: number
  private path: Waypoint[]
  private currentPathIndex: number
  private isMoving: boolean

  constructor(
    initialPosition: Position,
    speed: number,
    path: Waypoint[]
  ) {
    this.position = { ...initialPosition }
    this.speed = speed
    this.path = [...path]
    this.currentPathIndex = 0
    this.isMoving = true
  }

  // 위치 업데이트 (프레임 레이트 독립적)
  update(deltaTime: number): boolean {
    if (!this.isMoving || this.currentPathIndex >= this.path.length - 1) {
      return false // 이동 완료
    }

    const targetWaypoint = this.path[this.currentPathIndex + 1]
    if (!targetWaypoint) {
      this.isMoving = false
      return false
    }

    // 현재 위치에서 목표 웨이포인트까지의 벡터
    const dx = targetWaypoint.x - this.position.x
    const dz = targetWaypoint.y - this.position.z
    const distance = Math.sqrt(dx * dx + dz * dz)

    // 목표에 도달했는지 확인
    if (distance < 0.02) {
      this.currentPathIndex++
      this.position.x = targetWaypoint.x
      this.position.z = targetWaypoint.y
      return true // 계속 이동 중
    }

    // 정규화된 방향 벡터
    const directionX = dx / distance
    const directionZ = dz / distance

    // 이동 거리 계산 (deltaTime은 밀리초)
    const moveDistance = this.speed * (deltaTime / 1000)

    // 새 위치 계산
    this.position.x += directionX * moveDistance
    this.position.z += directionZ * moveDistance

    return true // 계속 이동 중
  }

  // 현재 위치 반환
  getPosition(): Position {
    return { ...this.position }
  }

  // 이동 중인지 확인
  isCurrentlyMoving(): boolean {
    return this.isMoving && this.currentPathIndex < this.path.length - 1
  }

  // 경로 완주했는지 확인
  hasReachedEnd(): boolean {
    return this.currentPathIndex >= this.path.length - 1
  }

  // 진행률 반환 (0-1)
  getProgress(): number {
    if (this.path.length <= 1) return 1
    return Math.min(this.currentPathIndex / (this.path.length - 1), 1)
  }

  // 속도 변경
  setSpeed(newSpeed: number): void {
    this.speed = newSpeed
  }

  // 강제로 특정 위치로 이동
  setPosition(newPosition: Position): void {
    this.position = { ...newPosition }
  }

  // 이동 중단
  stop(): void {
    this.isMoving = false
  }

  // 이동 재개
  resume(): void {
    this.isMoving = true
  }
}