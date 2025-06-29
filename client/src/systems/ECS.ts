// ECS 기본 타입 정의

export type EntityId = string

export interface Component {
  type: string
}

export interface Entity {
  id: EntityId
  components: Map<string, Component>
}

export abstract class System {
  abstract requiredComponents: string[]
  
  abstract update(entities: Entity[], deltaTime: number): void
  
  protected getEntitiesWithComponents(entities: Entity[]): Entity[] {
    return entities.filter(entity => 
      this.requiredComponents.every(componentType => 
        entity.components.has(componentType)
      )
    )
  }
  
  protected getComponent<T extends Component>(entity: Entity, type: string): T | undefined {
    return entity.components.get(type) as T | undefined
  }
}

export class World {
  private entities: Map<EntityId, Entity> = new Map()
  private systems: System[] = []
  
  createEntity(id: EntityId): Entity {
    const entity: Entity = {
      id,
      components: new Map()
    }
    this.entities.set(id, entity)
    return entity
  }
  
  removeEntity(id: EntityId): void {
    this.entities.delete(id)
  }
  
  addComponent<T extends Component>(entityId: EntityId, component: T): void {
    const entity = this.entities.get(entityId)
    if (entity) {
      entity.components.set(component.type, component)
    }
  }
  
  removeComponent(entityId: EntityId, componentType: string): void {
    const entity = this.entities.get(entityId)
    if (entity) {
      entity.components.delete(componentType)
    }
  }
  
  getEntity(id: EntityId): Entity | undefined {
    return this.entities.get(id)
  }
  
  getAllEntities(): Entity[] {
    return Array.from(this.entities.values())
  }
  
  addSystem(system: System): void {
    this.systems.push(system)
  }
  
  update(deltaTime: number): void {
    const entities = this.getAllEntities()
    this.systems.forEach(system => {
      system.update(entities, deltaTime)
    })
  }
}