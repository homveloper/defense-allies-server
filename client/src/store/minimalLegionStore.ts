import { create } from 'zustand';
import Phaser from 'phaser';

interface Player {
  health: number;
  maxHealth: number;
  attackPower: number;
  moveSpeed: number;
  attackSpeed: number;
  range: number;
  level: number;
  experience: number;
  experienceToNext: number;
}

interface GameState {
  game: Phaser.Game | null;
  player: Player;
  allies: number;
  maxAllies: number;
  wave: number;
  score: number;
  enemiesRemaining: number;
  isPaused: boolean;
  isGameOver: boolean;
}

interface MinimalLegionStore extends GameState {
  setGame: (game: Phaser.Game | null) => void;
  updatePlayer: (updates: Partial<Player>) => void;
  addExperience: (amount: number) => void;
  levelUp: () => void;
  addAlly: () => void;
  removeAlly: () => void;
  nextWave: () => void;
  updateScore: (points: number) => void;
  setEnemiesRemaining: (count: number) => void;
  togglePause: () => void;
  gameOver: () => void;
  resetGame: () => void;
}

const initialPlayer: Player = {
  health: 100,
  maxHealth: 100,
  attackPower: 10,
  moveSpeed: 5,
  attackSpeed: 2,
  range: 150,
  level: 1,
  experience: 0,
  experienceToNext: 100,
};

const initialState: GameState = {
  game: null,
  player: initialPlayer,
  allies: 0,
  maxAllies: 5,
  wave: 1,
  score: 0,
  enemiesRemaining: 0,
  isPaused: false,
  isGameOver: false,
};

export const useMinimalLegionStore = create<MinimalLegionStore>((set) => ({
  ...initialState,

  setGame: (game) => set({ game }),

  updatePlayer: (updates) =>
    set((state) => ({
      player: { ...state.player, ...updates },
    })),

  addExperience: (amount) =>
    set((state) => {
      const newExperience = state.player.experience + amount;
      const shouldLevelUp = newExperience >= state.player.experienceToNext;

      return {
        player: {
          ...state.player,
          experience: shouldLevelUp
            ? newExperience - state.player.experienceToNext
            : newExperience,
        },
      };
    }),

  levelUp: () =>
    set((state) => ({
      player: {
        ...state.player,
        level: state.player.level + 1,
        experienceToNext: 100 * (state.player.level + 1),
      },
    })),

  addAlly: () =>
    set((state) => ({
      allies: Math.min(state.allies + 1, state.maxAllies),
    })),

  removeAlly: () =>
    set((state) => ({
      allies: Math.max(state.allies - 1, 0),
    })),

  nextWave: () =>
    set((state) => ({
      wave: state.wave + 1,
    })),

  updateScore: (points) =>
    set((state) => ({
      score: state.score + points,
    })),

  setEnemiesRemaining: (count) => set({ enemiesRemaining: count }),

  togglePause: () =>
    set((state) => ({
      isPaused: !state.isPaused,
    })),

  gameOver: () => set({ isGameOver: true }),

  resetGame: () =>
    set((state) => ({
      ...initialState,
      game: state.game,
    })),
}));