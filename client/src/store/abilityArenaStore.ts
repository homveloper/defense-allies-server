import { create } from 'zustand';

interface Player {
  health: number;
  maxHealth: number;
  mana: number;
  maxMana: number;
  stamina: number;
  maxStamina: number;
  level: number;
  experience: number;
  experienceToNext: number;
  killCount: number;
  abilities: string[];
  currentRandomAbility?: string;
}

interface GameStats {
  waveNumber: number;
  enemiesKilled: number;
  damageDealt: number;
  damageTaken: number;
  abilitiesUsed: number;
  timeAlive: number; // seconds
  score: number;
  powerUpsCollected: number;
}

interface AbilityArenaState {
  // Game State
  isGameStarted: boolean;
  isGameOver: boolean;
  isPaused: boolean;
  isAbilitySelectionOpen: boolean;
  
  // Player State
  player: Player;
  
  // Game Stats
  stats: GameStats;
  
  // Available abilities for selection
  availableAbilities: Array<{
    id: string;
    name: string;
    description: string;
    icon: string;
    rarity: 'common' | 'rare' | 'epic' | 'legendary';
  }>;

  // Actions
  startGame: () => void;
  endGame: () => void;
  resetGame: () => void;
  setPaused: (paused: boolean) => void;
  
  // Player Actions
  updatePlayerHealth: (health: number) => void;
  updatePlayerMana: (mana: number) => void;
  updatePlayerStamina: (stamina: number) => void;
  addExperience: (amount: number) => void;
  levelUp: () => void;
  addAbility: (abilityId: string) => void;
  updateCurrentRandomAbility: (abilityName: string) => void;
  
  // Stats Actions
  incrementWave: () => void;
  incrementKills: (count?: number) => void;
  addDamageDealt: (amount: number) => void;
  addDamageTaken: (amount: number) => void;
  incrementAbilitiesUsed: () => void;
  updateTimeAlive: (seconds: number) => void;
  addScore: (points: number) => void;
  incrementPowerUps: () => void;
  
  // Ability Selection
  openAbilitySelection: (abilities: AbilityArenaState['availableAbilities']) => void;
  closeAbilitySelection: () => void;
  selectAbility: (abilityId: string) => void;
}

const initialPlayer: Player = {
  health: 100,
  maxHealth: 100,
  mana: 50,
  maxMana: 50,
  stamina: 100,
  maxStamina: 100,
  level: 1,
  experience: 0,
  experienceToNext: 100,
  killCount: 0,
  abilities: ['basic_attack'], // Start with basic attack
  currentRandomAbility: 'None'
};

const initialStats: GameStats = {
  waveNumber: 1,
  enemiesKilled: 0,
  damageDealt: 0,
  damageTaken: 0,
  abilitiesUsed: 0,
  timeAlive: 0,
  score: 0,
  powerUpsCollected: 0
};

export const useAbilityArenaStore = create<AbilityArenaState>((set, get) => ({
  // Initial State
  isGameStarted: false,
  isGameOver: false,
  isPaused: false,
  isAbilitySelectionOpen: false,
  
  player: { ...initialPlayer },
  stats: { ...initialStats },
  availableAbilities: [],

  // Game Actions
  startGame: () => set({ 
    isGameStarted: true, 
    isGameOver: false, 
    isPaused: false 
  }),

  endGame: () => set({ 
    isGameOver: true, 
    isGameStarted: false,
    isPaused: false
  }),

  resetGame: () => set({
    isGameStarted: false,
    isGameOver: false,
    isPaused: false,
    isAbilitySelectionOpen: false,
    player: { ...initialPlayer },
    stats: { ...initialStats },
    availableAbilities: []
  }),

  setPaused: (paused: boolean) => set({ isPaused: paused }),

  // Player Actions
  updatePlayerHealth: (health: number) => set((state) => ({
    player: { 
      ...state.player, 
      health: Math.max(0, Math.min(health, state.player.maxHealth))
    }
  })),

  updatePlayerMana: (mana: number) => set((state) => ({
    player: { 
      ...state.player, 
      mana: Math.max(0, Math.min(mana, state.player.maxMana))
    }
  })),

  updatePlayerStamina: (stamina: number) => set((state) => ({
    player: { 
      ...state.player, 
      stamina: Math.max(0, Math.min(stamina, state.player.maxStamina))
    }
  })),

  addExperience: (amount: number) => set((state) => {
    const newExp = state.player.experience + amount;
    let newLevel = state.player.level;
    let expToNext = state.player.experienceToNext;
    let shouldLevelUp = false;

    // Check for level up
    if (newExp >= expToNext) {
      newLevel++;
      shouldLevelUp = true;
      expToNext = newLevel * 100; // Simple formula: level * 100
    }

    const newState = {
      player: {
        ...state.player,
        experience: newExp,
        level: newLevel,
        experienceToNext: expToNext
      }
    };

    // Trigger level up modal if needed
    if (shouldLevelUp) {
      // This will be handled by the game scene
      console.log(`Level up! New level: ${newLevel}`);
    }

    return newState;
  }),

  levelUp: () => set((state) => ({
    player: {
      ...state.player,
      level: state.player.level + 1,
      maxHealth: state.player.maxHealth + 20,
      health: state.player.health + 20,
      maxMana: state.player.maxMana + 10,
      mana: state.player.mana + 10,
      experienceToNext: (state.player.level + 1) * 100
    }
  })),

  addAbility: (abilityId: string) => set((state) => ({
    player: {
      ...state.player,
      abilities: [...state.player.abilities, abilityId]
    }
  })),

  updateCurrentRandomAbility: (abilityName: string) => set((state) => ({
    player: {
      ...state.player,
      currentRandomAbility: abilityName
    }
  })),

  // Stats Actions
  incrementWave: () => set((state) => ({
    stats: { ...state.stats, waveNumber: state.stats.waveNumber + 1 }
  })),

  incrementKills: (count = 1) => set((state) => ({
    stats: { ...state.stats, enemiesKilled: state.stats.enemiesKilled + count },
    player: { ...state.player, killCount: state.player.killCount + count }
  })),

  addDamageDealt: (amount: number) => set((state) => ({
    stats: { ...state.stats, damageDealt: state.stats.damageDealt + amount }
  })),

  addDamageTaken: (amount: number) => set((state) => ({
    stats: { ...state.stats, damageTaken: state.stats.damageTaken + amount }
  })),

  incrementAbilitiesUsed: () => set((state) => ({
    stats: { ...state.stats, abilitiesUsed: state.stats.abilitiesUsed + 1 }
  })),

  updateTimeAlive: (seconds: number) => set((state) => ({
    stats: { ...state.stats, timeAlive: seconds }
  })),

  addScore: (points: number) => set((state) => ({
    stats: { ...state.stats, score: state.stats.score + points }
  })),

  incrementPowerUps: () => set((state) => ({
    stats: { ...state.stats, powerUpsCollected: state.stats.powerUpsCollected + 1 }
  })),

  // Ability Selection Actions
  openAbilitySelection: (abilities) => set({
    isAbilitySelectionOpen: true,
    availableAbilities: abilities,
    isPaused: true
  }),

  closeAbilitySelection: () => set({
    isAbilitySelectionOpen: false,
    availableAbilities: [],
    isPaused: false
  }),

  selectAbility: (abilityId: string) => {
    const state = get();
    
    // Add ability to player
    state.addAbility(abilityId);
    
    // Level up the player
    state.levelUp();
    
    // Close selection modal
    state.closeAbilitySelection();
    
    console.log(`Selected ability: ${abilityId}`);
  }
}));