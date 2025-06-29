import { System, Entity } from './ECS'
import { HealthComponent } from './components'

export class HealthSystem extends System {
  requiredComponents = ['health']
  
  private damageQueue: Array<{ entityId: string, damage: number }> = []
  
  update(entities: Entity[], deltaTime: number): void {
    // 대기 중인 데미지 처리
    this.processDamageQueue(entities)
  }
  
  dealDamage(entityId: string, damage: number): void {
    this.damageQueue.push({ entityId, damage })
  }
  
  private processDamageQueue(entities: Entity[]): void {
    while (this.damageQueue.length > 0) {
      const { entityId, damage } = this.damageQueue.shift()!
      
      const entity = entities.find(e => e.id === entityId)
      if (!entity) continue
      
      const health = this.getComponent<HealthComponent>(entity, 'health')
      if (!health) continue
      
      health.current = Math.max(0, health.current - damage)
      
      // 체력이 0이 되면 엔티티에 'dead' 마크 추가 (다른 시스템에서 처리)
      if (health.current <= 0) {
        entity.components.set('dead', { type: 'dead' })
      }
    }
  }
  
  heal(entityId: string, amount: number, entities: Entity[]): void {
    const entity = entities.find(e => e.id === entityId)
    if (!entity) return
    
    const health = this.getComponent<HealthComponent>(entity, 'health')
    if (!health) return
    
    health.current = Math.min(health.max, health.current + amount)
  }
}