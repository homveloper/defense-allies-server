export interface Vector2D {
  x: number;
  y: number;
}

export type GameState = 'menu' | 'playing' | 'paused' | 'levelup' | 'gameover';

export type EntityType = 'player' | 'ally' | 'enemy' | 'projectile';

export interface Entity {
  id: string;
  position: Vector2D;
  velocity: Vector2D;
  health: number;
  maxHealth: number;
  damage: number;
  speed: number;
  size: number;
  color: string;
  type: EntityType;
  target?: string;
  owner?: string;
}

export type UpgradeType = 'player' | 'legion' | 'utility';

export interface UpgradeOption {
  id: string;
  name: string;
  description: string;
  type: UpgradeType;
  effect: UpgradeEffect;
  icon?: string;
  maxLevel?: number;
  currentLevel?: number;
}

export interface UpgradeEffect {
  stat: string;
  value: number;
  isPercentage?: boolean;
}

export type AbilityType = 'instant' | 'duration' | 'permanent' | 'debuff';

export interface Ability {
  id: string;
  name: string;
  description: string;
  type: AbilityType;
  cooldown: number;
  currentCooldown: number;
  duration?: number;
  effect: AbilityEffect;
  icon?: string;
}

export interface AbilityEffect {
  type: string;
  value: number;
  radius?: number;
  target?: 'self' | 'allies' | 'enemies' | 'all';
}

export interface WaveConfig {
  waveNumber: number;
  enemyCount: number;
  enemyTypes: EnemyType[];
  spawnInterval: number;
  isBossWave?: boolean;
}

export interface EnemyType {
  id: string;
  name: string;
  health: number;
  damage: number;
  speed: number;
  size: number;
  color: string;
  experience: number;
  special?: string[];
}