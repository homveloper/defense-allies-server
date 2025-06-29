// 타워 색상
function getTowerColor(type: string): string {
  switch (type) {
    case 'basic': return '#2563eb'    // 기사단 요새 - 파란색
    case 'slow': return '#f59e0b'     // 상인 길드 - 주황색  
    case 'splash': return '#8b5cf6'   // 마법사 탑 - 보라색
    case 'laser': return '#22c55e'    // 대성당 - 초록색
    default: return '#6b7280'
  }
}

export interface TowerData {
  id: string
  position: [number, number, number]
  type: 'basic' | 'splash' | 'slow' | 'laser'
  level: number
  damage: number
  range: number
  attackSpeed: number
  lastAttack: number
  cost: number
}

export function renderTower(
  tower: TowerData, 
  ctx: CanvasRenderingContext2D, 
  cellSize: number, 
  colors: any,
  isHovered: boolean = false
) {
  const gridX = Math.floor(tower.position[0] + 10)
  const gridY = Math.floor(tower.position[2] + 10)
  
  // 타워 베이스
  ctx.fillStyle = getTowerColor(tower.type)
  ctx.fillRect(
    gridX * cellSize + 3,
    gridY * cellSize + 3,
    cellSize - 6,
    cellSize - 6
  )
  
  // 타워 내부
  ctx.fillStyle = colors.bg.primary
  ctx.fillRect(
    gridX * cellSize + 6,
    gridY * cellSize + 6,
    cellSize - 12,
    cellSize - 12
  )
  
  // 타워 레벨 표시
  if (tower.level > 1) {
    ctx.fillStyle = '#ffffff'
    ctx.font = '8px Arial'
    ctx.textAlign = 'center'
    ctx.fillText(
      tower.level.toString(),
      gridX * cellSize + cellSize / 2,
      gridY * cellSize + cellSize / 2 + 2
    )
  }
  
  // 타워 공격 범위 표시 (호버 시)
  if (isHovered) {
    ctx.strokeStyle = getTowerColor(tower.type) + '40'
    ctx.lineWidth = 2
    const rangeRadius = tower.range * cellSize
    ctx.beginPath()
    ctx.arc(
      (gridX + 0.5) * cellSize,
      (gridY + 0.5) * cellSize,
      rangeRadius,
      0,
      Math.PI * 2
    )
    ctx.stroke()
  }
  
  // 공격 애니메이션 효과
  const timeSinceAttack = Date.now() - tower.lastAttack
  if (timeSinceAttack < 200) { // 0.2초 동안 공격 효과 표시
    const alpha = 1 - (timeSinceAttack / 200)
    ctx.fillStyle = `rgba(255, 255, 0, ${alpha * 0.5})`
    ctx.beginPath()
    ctx.arc(
      (gridX + 0.5) * cellSize,
      (gridY + 0.5) * cellSize,
      cellSize / 3,
      0,
      Math.PI * 2
    )
    ctx.fill()
  }
}