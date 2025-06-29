import { create } from 'zustand';
import { Entity, Vector2D, GameState, UpgradeOption, Ability } from './types/minimalLegion';
import { EnemySpawnSystem } from './systems/enemySpawnSystem';
import { CombatSystem } from './systems/combatSystem';
import { MovementSystem } from './systems/movementSystem';
import { CollisionSystem } from './systems/collisionSystem';

interface MinimalLegionStore {
  // Game State
  gameState: GameState;
  wave: number;
  score: number;
  playTime: number;
  
  // Player
  player: Entity;
  playerLevel: number;
  experience: number;
  experienceToNextLevel: number;
  
  // Entities
  allies: Entity[];
  enemies: Entity[];
  projectiles: Entity[];
  
  // Upgrades & Abilities
  availableUpgrades: UpgradeOption[];
  activeAbilities: Ability[];
  abilitySlots: (Ability | null)[];
  
  // Game Configuration
  maxAllies: number;
  
  // Camera
  camera: { x: number; y: number };
  
  // Systems
  enemySpawnSystem: EnemySpawnSystem;
  combatSystem: CombatSystem;
  movementSystem: MovementSystem;
  collisionSystem: CollisionSystem;
  
  // Actions
  startGame: () => void;
  pauseGame: () => void;
  resumeGame: () => void;
  gameOver: () => void;
  
  // Player Actions
  movePlayer: (direction: Vector2D) => void;
  gainExperience: (amount: number) => void;
  levelUp: () => void;
  selectUpgrade: (upgrade: UpgradeOption) => void;
  
  // Entity Management
  spawnEnemy: (enemy: Entity) => void;
  convertEnemyToAlly: (enemyId: string) => void;
  removeEntity: (type: 'enemy' | 'ally' | 'projectile', id: string) => void;
  spawnProjectile: (projectile: Entity) => void;
  
  // Ability System
  useAbility: (slotIndex: number) => void;
  addAbility: (ability: Ability) => void;
  
  // Update Loop
  updateGame: (deltaTime: number) => void;
}

const initialPlayer: Entity = {
  id: 'player',
  position: { x: 600, y: 400 },
  velocity: { x: 0, y: 0 },
  health: 100,
  maxHealth: 100,
  damage: 10,
  speed: 5,
  size: 20,
  color: '#3B82F6',
  type: 'player'
};

export const useMinimalLegionStore = create<MinimalLegionStore>((set, get) => ({
  // Initial State
  gameState: 'menu',
  wave: 0,
  score: 0,
  playTime: 0,
  
  player: initialPlayer,
  playerLevel: 1,
  experience: 0,
  experienceToNextLevel: 100,
  
  allies: [],
  enemies: [],
  projectiles: [],
  
  availableUpgrades: [],
  activeAbilities: [],
  abilitySlots: [null, null, null, null],
  
  maxAllies: 5,
  
  // Camera
  camera: { x: 0, y: 0 },
  
  // Initialize systems
  enemySpawnSystem: new EnemySpawnSystem(),
  combatSystem: new CombatSystem(),
  movementSystem: new MovementSystem(),
  collisionSystem: new CollisionSystem(),
  
  // Game Control Actions
  startGame: () => {
    const newEnemySpawnSystem = new EnemySpawnSystem();
    newEnemySpawnSystem.startNewWave();
    
    set({
      gameState: 'playing',
      wave: 1,
      score: 0,
      playTime: 0,
      player: { ...initialPlayer },
      playerLevel: 1,
      experience: 0,
      experienceToNextLevel: 100,
      allies: [],
      enemies: [],
      projectiles: [],
      maxAllies: 5,
      camera: { x: 0, y: 0 },
      enemySpawnSystem: newEnemySpawnSystem,
      combatSystem: new CombatSystem(),
      movementSystem: new MovementSystem(),
      collisionSystem: new CollisionSystem()
    });
    
    console.log('Game started! Wave 1 begins.');
  },
  
  pauseGame: () => set({ gameState: 'paused' }),
  resumeGame: () => set({ gameState: 'playing' }),
  gameOver: () => set({ gameState: 'gameover' }),
  
  // Player Actions
  movePlayer: (direction) => {
    set((state) => ({
      player: {
        ...state.player,
        velocity: direction
      }
    }));
  },
  
  gainExperience: (amount) => {
    set((state) => {
      const newExperience = state.experience + amount;
      if (newExperience >= state.experienceToNextLevel) {
        // 레벨에 따라 경험치 요구량 증가 (1.2배씩)
        const newRequirement = Math.floor(state.experienceToNextLevel * 1.2);
        return {
          experience: newExperience - state.experienceToNextLevel,
          experienceToNextLevel: newRequirement,
          playerLevel: state.playerLevel + 1,
          gameState: 'levelup'
        };
      }
      return { experience: newExperience };
    });
  },
  
  levelUp: () => {
    set((state) => ({
      playerLevel: state.playerLevel + 1,
      gameState: 'levelup'
    }));
  },
  
  selectUpgrade: (upgrade) => {
    // Apply upgrade effects here
    set({ gameState: 'playing' });
  },
  
  // Entity Management
  spawnEnemy: (enemy) => {
    set((state) => ({
      enemies: [...state.enemies, enemy]
    }));
  },
  
  convertEnemyToAlly: (enemyId) => {
    set((state) => {
      const enemy = state.enemies.find(e => e.id === enemyId);
      if (!enemy || state.allies.length >= state.maxAllies) return state;
      
      const ally: Entity = {
        ...enemy,
        color: '#3B82F6',
        type: 'ally',
        health: enemy.maxHealth
      };
      
      return {
        enemies: state.enemies.filter(e => e.id !== enemyId),
        allies: [...state.allies, ally],
        score: state.score + 10
      };
    });
  },
  
  removeEntity: (type, id) => {
    set((state) => {
      switch (type) {
        case 'enemy':
          return { enemies: state.enemies.filter(e => e.id !== id) };
        case 'ally':
          return { allies: state.allies.filter(a => a.id !== id) };
        case 'projectile':
          return { projectiles: state.projectiles.filter(p => p.id !== id) };
        default:
          return state;
      }
    });
  },
  
  spawnProjectile: (projectile) => {
    set((state) => ({
      projectiles: [...state.projectiles, projectile]
    }));
  },
  
  // Ability System
  useAbility: (slotIndex) => {
    const ability = get().abilitySlots[slotIndex];
    if (!ability || ability.currentCooldown > 0) return;
    
    // Execute ability effect
    // This will be implemented based on ability type
  },
  
  addAbility: (ability) => {
    set((state) => {
      const emptySlotIndex = state.abilitySlots.findIndex(slot => slot === null);
      if (emptySlotIndex === -1) return state;
      
      const newSlots = [...state.abilitySlots];
      newSlots[emptySlotIndex] = ability;
      
      return {
        abilitySlots: newSlots,
        activeAbilities: [...state.activeAbilities, ability]
      };
    });
  },
  
  // Update Loop
  updateGame: (deltaTime) => {
    const state = get();
    if (state.gameState !== 'playing') return;
    
    // Update enemy spawn system
    state.enemySpawnSystem.update(deltaTime, state.wave, state.player.position, (enemy) => {
      get().spawnEnemy(enemy);
    });
    
    // Update combat system
    state.combatSystem.update(
      deltaTime, 
      state.player, 
      state.allies, 
      state.enemies, 
      (projectile) => get().spawnProjectile(projectile)
    );
    
    // Update movement system
    const { updatedAllies, updatedEnemies, updatedProjectiles } = state.movementSystem.update(
      deltaTime,
      state.player,
      state.allies,
      state.enemies,
      state.projectiles
    );
    
    // Check collisions
    const collisionResult = state.collisionSystem.checkCollisions(
      state.player,
      updatedAllies,
      updatedEnemies,
      updatedProjectiles
    );
    
    set((currentState) => {
      let newState = {
        ...currentState,
        playTime: currentState.playTime + deltaTime,
        allies: updatedAllies,
        enemies: updatedEnemies,
        projectiles: updatedProjectiles
      };
      
      // Update player position (무한 맵)
      newState.player = {
        ...newState.player,
        position: {
          x: newState.player.position.x + newState.player.velocity.x * newState.player.speed * 60 * deltaTime,
          y: newState.player.position.y + newState.player.velocity.y * newState.player.speed * 60 * deltaTime
        }
      };
      
      // Update camera to follow player
      newState.camera = {
        x: newState.player.position.x - 600, // 화면 중앙에 플레이어
        y: newState.player.position.y - 400
      };
      
      // Apply collision damage to player
      if (collisionResult.playerDamage > 0) {
        newState.player = {
          ...newState.player,
          health: Math.max(0, newState.player.health - collisionResult.playerDamage)
        };
        
        if (newState.player.health <= 0) {
          newState.gameState = 'gameover';
        }
      }
      
      // Apply collision damage to allies
      newState.allies = newState.allies.map(ally => {
        const damage = collisionResult.allyDamage.find(d => d.allyId === ally.id);
        if (damage) {
          return {
            ...ally,
            health: Math.max(0, ally.health - damage.damage)
          };
        }
        return ally;
      }).filter(ally => ally.health > 0);
      
      // Apply projectile hits
      for (const hit of collisionResult.projectileHits) {
        // Remove projectile
        newState.projectiles = newState.projectiles.filter(p => p.id !== hit.projectileId);
        
        // Apply damage to target
        if (hit.targetId === newState.player.id) {
          newState.player = {
            ...newState.player,
            health: Math.max(0, newState.player.health - hit.damage)
          };
          if (newState.player.health <= 0) {
            newState.gameState = 'gameover';
          }
        } else {
          // Check if it's an enemy
          const enemyIndex = newState.enemies.findIndex(e => e.id === hit.targetId);
          if (enemyIndex !== -1) {
            const enemy = newState.enemies[enemyIndex];
            const newHealth = enemy.health - hit.damage;
            
            if (newHealth <= 0) {
              // Enemy defeated - convert to ally or give experience
              if (newState.allies.length < newState.maxAllies) {
                state.convertEnemyToAlly(enemy.id);
              } else {
                // Give experience instead
                state.gainExperience(10);
                newState.enemies = newState.enemies.filter(e => e.id !== enemy.id);
                newState.score = newState.score + 10;
              }
            } else {
              // Update enemy health
              newState.enemies = newState.enemies.map(e => 
                e.id === hit.targetId ? { ...e, health: newHealth } : e
              );
            }
          } else {
            // Check if it's an ally
            newState.allies = newState.allies.map(ally => {
              if (ally.id === hit.targetId) {
                return {
                  ...ally,
                  health: Math.max(0, ally.health - hit.damage)
                };
              }
              return ally;
            }).filter(ally => ally.health > 0);
          }
        }
      }
      
      // Check wave completion
      if (newState.enemies.length === 0 && state.enemySpawnSystem.isWaveComplete()) {
        newState.wave = newState.wave + 1;
        state.enemySpawnSystem.startNewWave();
      }
      
      return newState;
    });
  }
}));