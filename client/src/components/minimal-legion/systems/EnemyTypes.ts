export interface EnemyTypeConfig {
  id: string;
  name: string;
  description: string;
  // 시각적 특성
  shape: 'circle' | 'square' | 'triangle' | 'pentagon' | 'star' | 'diamond';
  size: number; // 기본 크기 배수
  color: number; // 16진수 색상
  borderColor?: number;
  borderWidth?: number;
  // 능력치 배수
  healthMultiplier: number;
  damageMultiplier: number;
  speedMultiplier: number;
  // 특수 능력
  abilities: EnemyAbility[];
  // 보상
  experienceReward: number;
  scoreReward: number;
}

export interface EnemyAbility {
  type: 'ranged' | 'melee' | 'explosive' | 'heal' | 'shield' | 'speed_burst' | 'summon';
  value?: number;
  cooldown?: number;
}

// 기본 적 타입들
export const ENEMY_TYPES: Record<string, EnemyTypeConfig> = {
  // 기본형
  grunt: {
    id: 'grunt',
    name: '그런트',
    description: '가장 기본적인 적',
    shape: 'circle',
    size: 1,
    color: 0xe74c3c, // 빨간색
    healthMultiplier: 1,
    damageMultiplier: 1,
    speedMultiplier: 1,
    abilities: [{ type: 'melee' }],
    experienceReward: 10,
    scoreReward: 10
  },
  
  // 빠른 타입
  scout: {
    id: 'scout',
    name: '스카웃',
    description: '빠르지만 약한 적',
    shape: 'triangle',
    size: 0.8,
    color: 0xf39c12, // 주황색
    healthMultiplier: 0.6,
    damageMultiplier: 0.8,
    speedMultiplier: 1.8,
    abilities: [{ type: 'melee' }, { type: 'speed_burst', cooldown: 3000 }],
    experienceReward: 15,
    scoreReward: 15
  },
  
  // 탱커형
  tank: {
    id: 'tank',
    name: '탱크',
    description: '느리지만 강력한 적',
    shape: 'square',
    size: 1.5,
    color: 0x7f8c8d, // 회색
    borderColor: 0x34495e,
    borderWidth: 3,
    healthMultiplier: 2.5,
    damageMultiplier: 1.5,
    speedMultiplier: 0.6,
    abilities: [{ type: 'melee' }, { type: 'shield', value: 50 }],
    experienceReward: 30,
    scoreReward: 30
  },
  
  // 원거리형
  archer: {
    id: 'archer',
    name: '아처',
    description: '원거리 공격 전문',
    shape: 'diamond',
    size: 1,
    color: 0x27ae60, // 녹색
    healthMultiplier: 0.8,
    damageMultiplier: 1.2,
    speedMultiplier: 0.9,
    abilities: [{ type: 'ranged', cooldown: 1500 }],
    experienceReward: 20,
    scoreReward: 20
  },
  
  // 폭발형
  bomber: {
    id: 'bomber',
    name: '봄버',
    description: '죽을 때 폭발하는 적',
    shape: 'pentagon',
    size: 1.2,
    color: 0xe67e22, // 진한 주황
    borderColor: 0xd35400,
    borderWidth: 2,
    healthMultiplier: 1,
    damageMultiplier: 0.8,
    speedMultiplier: 1.1,
    abilities: [{ type: 'melee' }, { type: 'explosive', value: 25 }],
    experienceReward: 25,
    scoreReward: 25
  },
  
  // 치유형
  healer: {
    id: 'healer',
    name: '힐러',
    description: '주변 적을 치유하는 적',
    shape: 'star',
    size: 1,
    color: 0x3498db, // 파란색
    borderColor: 0x2980b9,
    borderWidth: 2,
    healthMultiplier: 0.9,
    damageMultiplier: 0.5,
    speedMultiplier: 0.8,
    abilities: [{ type: 'heal', value: 5, cooldown: 2000 }],
    experienceReward: 35,
    scoreReward: 35
  },
  
  // 엘리트형
  elite: {
    id: 'elite',
    name: '엘리트',
    description: '강력한 상급 적',
    shape: 'star',
    size: 1.3,
    color: 0x8e44ad, // 보라색
    borderColor: 0xffd700,
    borderWidth: 3,
    healthMultiplier: 1.8,
    damageMultiplier: 1.5,
    speedMultiplier: 1.2,
    abilities: [{ type: 'melee' }, { type: 'ranged', cooldown: 2000 }],
    experienceReward: 50,
    scoreReward: 50
  },
  
  // 소환형
  summoner: {
    id: 'summoner',
    name: '서머너',
    description: '미니언을 소환하는 적',
    shape: 'pentagon',
    size: 1.1,
    color: 0x16a085, // 청록색
    borderColor: 0x1abc9c,
    borderWidth: 2,
    healthMultiplier: 1.2,
    damageMultiplier: 0.7,
    speedMultiplier: 0.7,
    abilities: [{ type: 'summon', value: 2, cooldown: 5000 }],
    experienceReward: 40,
    scoreReward: 40
  },
  
  // 보스형
  boss: {
    id: 'boss',
    name: '보스',
    description: '웨이브 보스',
    shape: 'star',
    size: 2,
    color: 0xc0392b, // 진한 빨강
    borderColor: 0xffd700,
    borderWidth: 4,
    healthMultiplier: 5,
    damageMultiplier: 2,
    speedMultiplier: 0.8,
    abilities: [
      { type: 'melee' }, 
      { type: 'ranged', cooldown: 1000 },
      { type: 'summon', value: 3, cooldown: 8000 }
    ],
    experienceReward: 100,
    scoreReward: 100
  }
};

// 웨이브별 적 타입 분포
export function getEnemyTypesForWave(wave: number): string[] {
  const types: string[] = ['grunt']; // 항상 기본형 포함
  
  if (wave >= 2) types.push('scout');
  if (wave >= 3) types.push('archer');
  if (wave >= 4) types.push('tank');
  if (wave >= 5) types.push('bomber');
  if (wave >= 6) types.push('healer');
  if (wave >= 8) types.push('elite');
  if (wave >= 10) types.push('summoner');
  
  // 10웨이브마다 보스
  if (wave % 10 === 0) {
    types.push('boss');
  }
  
  return types;
}

// 랜덤하게 적 타입 선택 (웨이브 기반)
export function getRandomEnemyType(wave: number): EnemyTypeConfig {
  const availableTypes = getEnemyTypesForWave(wave);
  const weights: Record<string, number> = {
    grunt: 40,
    scout: 20,
    archer: 15,
    tank: 10,
    bomber: 8,
    healer: 5,
    elite: 5,
    summoner: 3,
    boss: 1
  };
  
  // 가중치 기반 랜덤 선택
  const totalWeight = availableTypes.reduce((sum, type) => sum + (weights[type] || 1), 0);
  let random = Math.random() * totalWeight;
  
  for (const type of availableTypes) {
    random -= weights[type] || 1;
    if (random <= 0) {
      return ENEMY_TYPES[type];
    }
  }
  
  return ENEMY_TYPES.grunt; // 기본값
}

// 특정 난이도에 맞는 적 타입 추천
export function getEnemyTypesForDifficulty(difficultyType: string): string[] {
  switch (difficultyType) {
    case 'easy':
      return ['grunt', 'scout'];
    case 'normal':
      return ['grunt', 'scout', 'archer', 'tank'];
    case 'hard':
      return ['archer', 'tank', 'bomber', 'healer', 'elite'];
    case 'peak':
      return ['elite', 'summoner', 'boss', 'tank', 'bomber'];
    case 'rest':
      return ['grunt', 'scout'];
    default:
      return ['grunt'];
  }
}