// 적 색상
function getEnemyColor(type: string): string {
  switch (type) {
    case 'fast': return '#10b981'  // 빠른 적 - 녹색
    case 'tank': return '#ef4444'  // 탱크 적 - 빨간색
    default: return '#f59e0b'      // 기본 적 - 주황색
  }
}

// 적 크기
function getEnemySize(type: string): number {
  switch (type) {
    case 'fast': return 6   // 빠른 적 - 작음
    case 'tank': return 12  // 탱크 적 - 큼
    default: return 8       // 기본 적 - 중간
  }
}

export interface EnemyData {
  id: string
  position: [number, number, number]
  health: number
  maxHealth: number
  speed: number
  pathIndex: number
  value: number
  type: 'basic' | 'fast' | 'tank'
}

export function renderEnemy(
  enemy: EnemyData, 
  ctx: CanvasRenderingContext2D, 
  cellSize: number
) {
  const x = (enemy.position[0] + 10) * cellSize
  const y = (enemy.position[2] + 10) * cellSize
  const size = getEnemySize(enemy.type)
  
  // 적 몸체
  ctx.fillStyle = getEnemyColor(enemy.type)
  ctx.beginPath()
  ctx.arc(x, y, size, 0, Math.PI * 2)
  ctx.fill()
  
  // 적 테두리
  ctx.strokeStyle = '#000000'
  ctx.lineWidth = 1
  ctx.stroke()
  
  // 타입별 특별한 표시
  if (enemy.type === 'tank') {
    // 탱크는 내부에 작은 원 추가
    ctx.fillStyle = '#ffffff'
    ctx.beginPath()
    ctx.arc(x, y, size - 4, 0, Math.PI * 2)
    ctx.fill()
    
    ctx.fillStyle = getEnemyColor(enemy.type)
    ctx.beginPath()
    ctx.arc(x, y, size - 6, 0, Math.PI * 2)
    ctx.fill()
  } else if (enemy.type === 'fast') {
    // 빠른 적은 번개 모양 표시
    ctx.strokeStyle = '#ffffff'
    ctx.lineWidth = 2
    ctx.beginPath()
    ctx.moveTo(x - 2, y - 3)
    ctx.lineTo(x + 1, y)
    ctx.lineTo(x - 1, y)
    ctx.lineTo(x + 2, y + 3)
    ctx.stroke()
  }
  
  // 체력바 (항상 수평)
  const healthPercent = enemy.health / enemy.maxHealth
  const barWidth = Math.max(16, size * 2.5) // 적 크기에 비례한 체력바
  const barHeight = 3
  const barY = y - size - 8 // 적 위쪽에 위치
  
  // 체력바 배경
  ctx.fillStyle = '#333333'
  ctx.fillRect(x - barWidth / 2, barY, barWidth, barHeight)
  
  // 체력바 채우기
  if (healthPercent > 0.6) {
    ctx.fillStyle = '#22c55e' // 초록색
  } else if (healthPercent > 0.3) {
    ctx.fillStyle = '#f59e0b' // 주황색
  } else {
    ctx.fillStyle = '#ef4444' // 빨간색
  }
  ctx.fillRect(x - barWidth / 2, barY, barWidth * healthPercent, barHeight)
  
  // 체력바 테두리
  ctx.strokeStyle = '#ffffff'
  ctx.lineWidth = 1
  ctx.strokeRect(x - barWidth / 2, barY, barWidth, barHeight)
  
  // 데미지 받은 효과 (체력이 감소했을 때)
  if (healthPercent < 1) {
    const damageAlpha = (1 - healthPercent) * 0.3
    ctx.fillStyle = `rgba(255, 0, 0, ${damageAlpha})`
    ctx.beginPath()
    ctx.arc(x, y, size + 2, 0, Math.PI * 2)
    ctx.fill()
  }
}