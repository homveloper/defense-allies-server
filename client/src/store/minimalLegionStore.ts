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

interface Upgrade {
  id: string;
  name: string;
  description: string;
  apply: (player: Player) => Partial<Player>;
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
  isLevelUpModalOpen: boolean;
  availableUpgrades: Upgrade[];
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
  showLevelUpModal: (upgrades: Upgrade[]) => void;
  hideLevelUpModal: () => void;
  selectUpgrade: (upgrade: Upgrade) => void;
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

// 업그레이드 옵션들
const UPGRADES: Upgrade[] = [
  {
    id: 'health_boost',
    name: '생명력 증가',
    description: '최대 체력이 20 증가하고 체력이 모두 회복됩니다',
    apply: (player) => ({ 
      maxHealth: player.maxHealth + 20, 
      health: player.maxHealth + 20 
    })
  },
  {
    id: 'attack_power',
    name: '공격력 강화',
    description: '공격력이 5 증가합니다',
    apply: (player) => ({ attackPower: player.attackPower + 5 })
  },
  {
    id: 'attack_speed',
    name: '공격 속도 증가',
    description: '공격 속도가 20% 증가합니다',
    apply: (player) => ({ attackSpeed: player.attackSpeed * 1.2 })
  },
  {
    id: 'move_speed',
    name: '이동 속도 증가',
    description: '이동 속도가 15% 증가합니다',
    apply: (player) => ({ moveSpeed: player.moveSpeed * 1.15 })
  },
  {
    id: 'range_boost',
    name: '사거리 확장',
    description: '공격 사거리가 30 증가합니다',
    apply: (player) => ({ range: player.range + 30 })
  },
  {
    id: 'max_allies',
    name: '군단 확장',
    description: '최대 동료 수가 2 증가합니다',
    apply: () => ({}) // 특별 처리 필요
  }
];

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
  isLevelUpModalOpen: false,
  availableUpgrades: [],
};

export const useMinimalLegionStore = create<MinimalLegionStore>((set) => ({
  ...initialState,

  setGame: (game) => set({ game }),

  updatePlayer: (updates) =>
    set((state) => {
      const updatedPlayer = { ...state.player, ...updates };
      
      // Validate health values
      if (updates.health !== undefined) {
        updatedPlayer.health = Math.max(0, Math.min(updatedPlayer.maxHealth, updates.health));
        if (isNaN(updatedPlayer.health)) {
          console.error('Invalid health value:', updates.health);
          updatedPlayer.health = state.player.health; // Keep previous value
        }
      }
      
      return { player: updatedPlayer };
    }),

  addExperience: (amount) =>
    set((state) => {
      if (!amount || isNaN(amount) || amount < 0) {
        console.warn('Invalid experience amount:', amount);
        return state;
      }
      
      const newExperience = state.player.experience + amount;
      const shouldLevelUp = newExperience >= state.player.experienceToNext;

      if (shouldLevelUp) {
        console.log('Level up! Showing upgrade selection');
        
        // 랜덤한 3개 업그레이드 선택
        const shuffled = [...UPGRADES].sort(() => Math.random() - 0.5);
        const randomUpgrades = shuffled.slice(0, 3);
        
        return {
          player: {
            ...state.player,
            level: state.player.level + 1,
            experience: newExperience - state.player.experienceToNext,
            experienceToNext: 100 * (state.player.level + 1),
          },
          isLevelUpModalOpen: true,
          availableUpgrades: randomUpgrades,
          isPaused: true, // 게임 일시정지
        };
      }

      return {
        player: {
          ...state.player,
          experience: newExperience,
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

  gameOver: () => {
    console.log('Game Over called');
    set({ isGameOver: true });
  },

  resetGame: () =>
    set((state) => {
      console.log('Store resetGame() called');
      return {
        ...initialState,
        game: state.game,
      };
    }),

  showLevelUpModal: (upgrades) =>
    set({
      isLevelUpModalOpen: true,
      availableUpgrades: upgrades,
      isPaused: true,
    }),

  hideLevelUpModal: () =>
    set({
      isLevelUpModalOpen: false,
      availableUpgrades: [],
      isPaused: false,
    }),

  selectUpgrade: (upgrade) =>
    set((state) => {
      const playerUpdates = upgrade.apply(state.player);
      const newState = {
        player: { ...state.player, ...playerUpdates },
        isLevelUpModalOpen: false,
        availableUpgrades: [],
        isPaused: false,
      };

      console.log(`Upgrade selected: ${upgrade.name}`, playerUpdates);
      
      // 군단 확장 업그레이드 특별 처리
      if (upgrade.id === 'max_allies') {
        return {
          ...newState,
          maxAllies: state.maxAllies + 2,
        };
      }

      return newState;
    }),
}));