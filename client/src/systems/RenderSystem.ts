import { System, Entity } from './ECS'
import { PositionComponent, RenderComponent, HealthComponent, EnemyTypeComponent, TowerTypeComponent } from './components'

// 렌더링 시스템은 React Three Fiber와 통합하기 위해 콜백 기반으로 동작
export type RenderCallback = (renderData: RenderData[]) => void

export interface RenderData {
  entityId: string
  position: { x: number, y: number, z: number }
  render: RenderComponent
  health?: HealthComponent
  enemyType?: EnemyTypeComponent
  towerType?: TowerTypeComponent
}

export class RenderSystem extends System {
  requiredComponents = ['position', 'render']
  private renderCallback: RenderCallback | null = null
  
  setRenderCallback(callback: RenderCallback): void {
    this.renderCallback = callback
  }
  
  update(entities: Entity[], deltaTime: number): void {
    if (!this.renderCallback) return
    
    const renderableEntities = this.getEntitiesWithComponents(entities)
    
    const renderData: RenderData[] = renderableEntities.map(entity => {
      const position = this.getComponent<PositionComponent>(entity, 'position')!
      const render = this.getComponent<RenderComponent>(entity, 'render')!
      const health = this.getComponent<HealthComponent>(entity, 'health')
      const enemyType = this.getComponent<EnemyTypeComponent>(entity, 'enemyType')
      const towerType = this.getComponent<TowerTypeComponent>(entity, 'towerType')
      
      // 회전 업데이트 (적만)
      if (enemyType) {
        const rotationSpeed = enemyType.enemyType === 'fast' ? 0.1 : 
                             enemyType.enemyType === 'tank' ? 0.02 : 0.05
        render.rotation += rotationSpeed
      }
      
      return {
        entityId: entity.id,
        position: { x: position.x, y: position.y, z: position.z },
        render,
        health,
        enemyType,
        towerType
      }
    })
    
    this.renderCallback(renderData)
  }
}