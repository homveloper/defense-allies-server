import Phaser from 'phaser';
import { MainScene } from '../scenes/MainScene';
import { EnemyTypeConfig, EnemyAbility } from '../systems/EnemyTypes';

export class Enemy extends Phaser.GameObjects.Container {
  public body!: Phaser.Physics.Arcade.Body;
  public health: number;
  private maxHealth: number;
  private sprite: Phaser.GameObjects.Graphics;
  private healthBar: Phaser.GameObjects.Graphics;
  private moveSpeed: number;
  private target: Phaser.GameObjects.GameObject | null = null;
  private attackPower: number;
  private attackRange: number = 100;
  private meleeRange: number = 25;
  private lastFireTime: number = 0;
  private lastMeleeTime: number = 0;
  private fireRate: number = 1500;
  private meleeRate: number = 1200;
  private enemyType: EnemyTypeConfig;
  private abilities: EnemyAbility[];
  private abilityTimers: Map<string, number> = new Map();
  private shield: number = 0;

  constructor(
    scene: Phaser.Scene, 
    x: number, 
    y: number, 
    _wave: number = 1,
    healthMultiplier: number = 1,
    damageMultiplier: number = 1,
    speedMultiplier: number = 1,
    enemyType?: EnemyTypeConfig
  ) {
    super(scene, x, y);

    // 적 타입 설정 (없으면 기본 grunt)
    this.enemyType = enemyType || {
      id: 'grunt',
      name: '그런트',
      description: '기본 적',
      shape: 'circle',
      size: 1,
      color: 0xe74c3c,
      healthMultiplier: 1,
      damageMultiplier: 1,
      speedMultiplier: 1,
      abilities: [{ type: 'melee' }],
      experienceReward: 10,
      scoreReward: 10
    };
    
    this.abilities = [...this.enemyType.abilities];
    
    // Base enemy stats
    const baseHealth = 20;
    const baseAttack = 5;
    const baseSpeed = 3;
    
    // Apply type multipliers and difficulty multipliers
    this.maxHealth = Math.floor(baseHealth * this.enemyType.healthMultiplier * healthMultiplier);
    this.health = this.maxHealth;
    this.moveSpeed = baseSpeed * this.enemyType.speedMultiplier * speedMultiplier;
    this.attackPower = Math.floor(baseAttack * this.enemyType.damageMultiplier * damageMultiplier);
    
    // Shield 능력 체크
    const shieldAbility = this.abilities.find(a => a.type === 'shield');
    if (shieldAbility) {
      this.shield = shieldAbility.value || 0;
    }
    
    console.log(`${this.enemyType.name} spawned: HP=${this.maxHealth}, ATK=${this.attackPower}, SPD=${this.moveSpeed.toFixed(1)}`);

    // Create enemy sprite based on type
    this.sprite = scene.add.graphics();
    this.createSprite();
    this.add(this.sprite);

    // Create health bar
    this.healthBar = scene.add.graphics();
    this.add(this.healthBar);
    this.updateHealthBar();

    // Physics
    scene.physics.world.enable(this);
    this.body.setCollideWorldBounds(true);
    const bodySize = 15 * this.enemyType.size;
    this.body.setCircle(bodySize);

    scene.add.existing(this);
  }
  
  private createSprite() {
    const size = 15 * this.enemyType.size;
    
    // 테두리가 있으면 먼저 그리기
    if (this.enemyType.borderColor && this.enemyType.borderWidth) {
      this.sprite.lineStyle(this.enemyType.borderWidth, this.enemyType.borderColor);
    }
    
    // 메인 색상 설정
    this.sprite.fillStyle(this.enemyType.color);
    
    // 모양별로 그리기
    switch (this.enemyType.shape) {
      case 'circle':
        this.sprite.fillCircle(0, 0, size);
        if (this.enemyType.borderColor) {
          this.sprite.strokeCircle(0, 0, size);
        }
        break;
        
      case 'square':
        this.sprite.fillRect(-size, -size, size * 2, size * 2);
        if (this.enemyType.borderColor) {
          this.sprite.strokeRect(-size, -size, size * 2, size * 2);
        }
        break;
        
      case 'triangle':
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(-size, size);
        this.sprite.lineTo(size, size);
        this.sprite.closePath();
        this.sprite.fill();
        if (this.enemyType.borderColor) {
          this.sprite.strokePath();
        }
        break;
        
      case 'diamond':
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(size, 0);
        this.sprite.lineTo(0, size);
        this.sprite.lineTo(-size, 0);
        this.sprite.closePath();
        this.sprite.fill();
        if (this.enemyType.borderColor) {
          this.sprite.strokePath();
        }
        break;
        
      case 'pentagon':
        const pentagonAngle = (Math.PI * 2) / 5;
        this.sprite.beginPath();
        for (let i = 0; i < 5; i++) {
          const x = Math.cos(pentagonAngle * i - Math.PI / 2) * size;
          const y = Math.sin(pentagonAngle * i - Math.PI / 2) * size;
          if (i === 0) {
            this.sprite.moveTo(x, y);
          } else {
            this.sprite.lineTo(x, y);
          }
        }
        this.sprite.closePath();
        this.sprite.fill();
        if (this.enemyType.borderColor) {
          this.sprite.strokePath();
        }
        break;
        
      case 'star':
        const starAngle = (Math.PI * 2) / 10;
        this.sprite.beginPath();
        for (let i = 0; i < 10; i++) {
          const radius = i % 2 === 0 ? size : size * 0.5;
          const x = Math.cos(starAngle * i - Math.PI / 2) * radius;
          const y = Math.sin(starAngle * i - Math.PI / 2) * radius;
          if (i === 0) {
            this.sprite.moveTo(x, y);
          } else {
            this.sprite.lineTo(x, y);
          }
        }
        this.sprite.closePath();
        this.sprite.fill();
        if (this.enemyType.borderColor) {
          this.sprite.strokePath();
        }
        break;
    }
    
    // Shield 표시
    if (this.shield > 0) {
      this.sprite.lineStyle(2, 0x3498db, 0.5);
      this.sprite.strokeCircle(0, 0, size + 5);
    }
  }
  
  private createFlashSprite() {
    const size = 15 * this.enemyType.size;
    
    // 흰색으로 그리기 (플래시 효과)
    this.sprite.fillStyle(0xffffff);
    
    switch (this.enemyType.shape) {
      case 'circle':
        this.sprite.fillCircle(0, 0, size);
        break;
      case 'square':
        this.sprite.fillRect(-size, -size, size * 2, size * 2);
        break;
      case 'triangle':
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(-size, size);
        this.sprite.lineTo(size, size);
        this.sprite.closePath();
        this.sprite.fill();
        break;
      case 'diamond':
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(size, 0);
        this.sprite.lineTo(0, size);
        this.sprite.lineTo(-size, 0);
        this.sprite.closePath();
        this.sprite.fill();
        break;
      case 'pentagon':
        const pentagonAngle = (Math.PI * 2) / 5;
        this.sprite.beginPath();
        for (let i = 0; i < 5; i++) {
          const x = Math.cos(pentagonAngle * i - Math.PI / 2) * size;
          const y = Math.sin(pentagonAngle * i - Math.PI / 2) * size;
          if (i === 0) {
            this.sprite.moveTo(x, y);
          } else {
            this.sprite.lineTo(x, y);
          }
        }
        this.sprite.closePath();
        this.sprite.fill();
        break;
      case 'star':
        const starAngle = (Math.PI * 2) / 10;
        this.sprite.beginPath();
        for (let i = 0; i < 10; i++) {
          const radius = i % 2 === 0 ? size : size * 0.5;
          const x = Math.cos(starAngle * i - Math.PI / 2) * radius;
          const y = Math.sin(starAngle * i - Math.PI / 2) * radius;
          if (i === 0) {
            this.sprite.moveTo(x, y);
          } else {
            this.sprite.lineTo(x, y);
          }
        }
        this.sprite.closePath();
        this.sprite.fill();
        break;
    }
  }
  
  private createSpeedBurstSprite() {
    const size = 15 * this.enemyType.size;
    
    // 노란색으로 그리기 (스피드 버스트 효과)
    this.sprite.fillStyle(0xffff00);
    
    switch (this.enemyType.shape) {
      case 'circle':
        this.sprite.fillCircle(0, 0, size);
        break;
      case 'square':
        this.sprite.fillRect(-size, -size, size * 2, size * 2);
        break;
      case 'triangle':
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(-size, size);
        this.sprite.lineTo(size, size);
        this.sprite.closePath();
        this.sprite.fill();
        break;
      case 'diamond':
        this.sprite.beginPath();
        this.sprite.moveTo(0, -size);
        this.sprite.lineTo(size, 0);
        this.sprite.lineTo(0, size);
        this.sprite.lineTo(-size, 0);
        this.sprite.closePath();
        this.sprite.fill();
        break;
      case 'pentagon':
        const pentagonAngle = (Math.PI * 2) / 5;
        this.sprite.beginPath();
        for (let i = 0; i < 5; i++) {
          const x = Math.cos(pentagonAngle * i - Math.PI / 2) * size;
          const y = Math.sin(pentagonAngle * i - Math.PI / 2) * size;
          if (i === 0) {
            this.sprite.moveTo(x, y);
          } else {
            this.sprite.lineTo(x, y);
          }
        }
        this.sprite.closePath();
        this.sprite.fill();
        break;
      case 'star':
        const starAngle = (Math.PI * 2) / 10;
        this.sprite.beginPath();
        for (let i = 0; i < 10; i++) {
          const radius = i % 2 === 0 ? size : size * 0.5;
          const x = Math.cos(starAngle * i - Math.PI / 2) * radius;
          const y = Math.sin(starAngle * i - Math.PI / 2) * radius;
          if (i === 0) {
            this.sprite.moveTo(x, y);
          } else {
            this.sprite.lineTo(x, y);
          }
        }
        this.sprite.closePath();
        this.sprite.fill();
        break;
    }
  }

  setTarget(target: Phaser.GameObjects.GameObject) {
    this.target = target;
  }

  takeDamage(amount: number) {
    let actualDamage = amount;
    
    // Shield 처리
    if (this.shield > 0) {
      const shieldDamage = Math.min(this.shield, amount);
      this.shield -= shieldDamage;
      actualDamage -= shieldDamage;
      console.log(`Shield absorbed ${shieldDamage} damage. Remaining shield: ${this.shield}`);
    }
    
    this.health -= actualDamage;
    this.updateHealthBar();

    // Flash effect - 흰색으로 잠시 변경
    this.sprite.clear();
    this.createFlashSprite();

    this.scene.time.delayedCall(100, () => {
      if (this.sprite && this.active) {
        this.sprite.clear();
        this.createSprite();
      }
    });
  }

  update() {
    if (!this.target || !this.target.active) return;

    const distance = Phaser.Math.Distance.Between(
      this.x,
      this.y,
      this.target.body!.position.x,
      this.target.body!.position.y
    );

    const now = this.scene.time.now;

    // Always move towards target unless very close for melee
    if (distance > this.meleeRange) {
      const angle = Phaser.Math.Angle.Between(
        this.x,
        this.y,
        this.target.body!.position.x,
        this.target.body!.position.y
      );

      this.body.setVelocity(
        Math.cos(angle) * this.moveSpeed * 60,
        Math.sin(angle) * this.moveSpeed * 60
      );
    } else {
      // Very close - slow down but keep moving slightly
      const angle = Phaser.Math.Angle.Between(
        this.x,
        this.y,
        this.target.body!.position.x,
        this.target.body!.position.y
      );

      this.body.setVelocity(
        Math.cos(angle) * this.moveSpeed * 20,
        Math.sin(angle) * this.moveSpeed * 20
      );
    }

    // 능력 기반 공격
    this.executeAbilities(distance, now);
    
    // 기본 공격
    if (distance <= this.meleeRange && now - this.lastMeleeTime > this.meleeRate) {
      this.meleeAttack();
      this.lastMeleeTime = now;
    }
  }

  private meleeAttack() {
    if (!this.target || !this.target.active || !this.attackPower) return;

    const mainScene = this.scene as MainScene;
    const damage = this.attackPower || 5; // Default damage if undefined
    
    console.log(`Enemy dealing ${damage} melee damage`);
    mainScene.dealMeleeDamage(this.target, damage);

    // Visual feedback for melee attack
    if (this.sprite && this.active) {
      this.sprite.clear();
      this.sprite.fillStyle(0xffffff);
      this.sprite.fillCircle(0, 0, 18);

      this.scene.time.delayedCall(100, () => {
        if (this.sprite && this.active) {
          this.sprite.clear();
          this.sprite.fillStyle(0xe74c3c);
          this.sprite.fillCircle(0, 0, 15);
        }
      });
    }
  }

  private executeAbilities(distance: number, now: number) {
    this.abilities.forEach(ability => {
      const cooldown = ability.cooldown || 1000;
      const lastUsed = this.abilityTimers.get(ability.type) || 0;
      
      if (now - lastUsed < cooldown) return;
      
      switch (ability.type) {
        case 'ranged':
          if (distance <= this.attackRange && distance > this.meleeRange) {
            this.fireProjectile();
            this.abilityTimers.set(ability.type, now);
          }
          break;
          
        case 'heal':
          this.healNearbyEnemies(ability.value || 5);
          this.abilityTimers.set(ability.type, now);
          break;
          
        case 'speed_burst':
          this.speedBurst();
          this.abilityTimers.set(ability.type, now);
          break;
          
        case 'explosive':
          // 폭발은 죽을 때 실행됨
          break;
          
        case 'summon':
          this.summonMinions(ability.value || 1);
          this.abilityTimers.set(ability.type, now);
          break;
      }
    });
  }
  
  private fireProjectile() {
    if (!this.target || !this.target.active) return;

    const mainScene = this.scene as MainScene;
    mainScene.fireProjectile(
      this.x,
      this.y,
      this.target.body!.position.x,
      this.target.body!.position.y,
      this.attackPower,
      true // isEnemy
    );
  }
  
  private healNearbyEnemies(healAmount: number) {
    const mainScene = this.scene as MainScene;
    const enemies = (mainScene as unknown as { enemies: { children: { entries: Enemy[] } } }).enemies.children.entries;
    
    enemies.forEach(enemy => {
      if (enemy === this || !enemy.active) return;
      
      const distance = Phaser.Math.Distance.Between(this.x, this.y, enemy.x, enemy.y);
      if (distance <= 100) { // 힐 범위
        enemy.heal(healAmount);
      }
    });
    
    // 치유 효과 표시
    this.showHealEffect();
  }
  
  private heal(amount: number) {
    this.health = Math.min(this.maxHealth, this.health + amount);
    this.updateHealthBar();
    
    // 힐 텍스트 표시
    const healText = this.scene.add.text(this.x, this.y - 30, `+${amount}`, {
      font: '12px Arial',
      color: '#00ff00'
    });
    healText.setOrigin(0.5);
    
    this.scene.tweens.add({
      targets: healText,
      y: healText.y - 20,
      alpha: 0,
      duration: 1000,
      onComplete: () => healText.destroy()
    });
  }
  
  private showHealEffect() {
    const healCircle = this.scene.add.graphics();
    healCircle.lineStyle(3, 0x00ff00, 0.8);
    healCircle.strokeCircle(this.x, this.y, 100);
    
    this.scene.tweens.add({
      targets: healCircle,
      alpha: 0,
      scale: 1.5,
      duration: 500,
      onComplete: () => healCircle.destroy()
    });
  }
  
  private speedBurst() {
    const originalSpeed = this.moveSpeed;
    this.moveSpeed *= 2;
    
    // 스피드 버스트 효과 표시 - 노란색으로 변경
    this.sprite.clear();
    this.createSpeedBurstSprite();
    
    this.scene.time.delayedCall(1000, () => {
      this.moveSpeed = originalSpeed;
      if (this.sprite && this.active) {
        this.sprite.clear();
        this.createSprite();
      }
    });
  }
  
  private summonMinions(count: number) {
    const mainScene = this.scene as MainScene;
    
    for (let i = 0; i < count; i++) {
      const angle = (Math.PI * 2 / count) * i;
      const distance = 50;
      const x = this.x + Math.cos(angle) * distance;
      const y = this.y + Math.sin(angle) * distance;
      
      // 작은 미니언 생성
      const minion = new Enemy(
        this.scene,
        x,
        y,
        1,
        0.5, // 체력 50%
        0.7, // 공격력 70%
        1.2, // 속도 120%
        {
          id: 'minion',
          name: '미니언',
          description: '소환된 미니언',
          shape: 'circle',
          size: 0.6,
          color: this.enemyType.color,
          healthMultiplier: 0.5,
          damageMultiplier: 0.7,
          speedMultiplier: 1.2,
          abilities: [{ type: 'melee' }],
          experienceReward: 5,
          scoreReward: 5
        }
      );
      
      (mainScene as unknown as { enemies: { add: (enemy: Enemy) => void } }).enemies.add(minion);
      if (this.target) {
        minion.setTarget(this.target);
      }
    }
  }
  
  public explode() {
    const explosiveAbility = this.abilities.find(a => a.type === 'explosive');
    if (!explosiveAbility) return;
    
    const explosionDamage = explosiveAbility.value || 25;
    const explosionRadius = 80;
    
    // 폭발 효과 표시
    const explosion = this.scene.add.graphics();
    explosion.fillStyle(0xff4500, 0.8);
    explosion.fillCircle(this.x, this.y, explosionRadius);
    
    this.scene.tweens.add({
      targets: explosion,
      scale: { from: 0.1, to: 1 },
      alpha: { from: 1, to: 0 },
      duration: 300,
      onComplete: () => explosion.destroy()
    });
    
    // 범위 내 플레이어와 아군에게 데미지
    const mainScene = this.scene as MainScene;
    const player = (mainScene as unknown as { player: { active: boolean; x: number; y: number; takeDamage: (amount: number) => void } }).player;
    const allies = (mainScene as unknown as { allies: { children: { entries: { active: boolean; x: number; y: number; takeDamage: (amount: number) => void }[] } } }).allies.children.entries;
    
    // 플레이어 데미지 체크
    if (player && player.active) {
      const playerDistance = Phaser.Math.Distance.Between(this.x, this.y, player.x, player.y);
      if (playerDistance <= explosionRadius) {
        player.takeDamage(explosionDamage);
      }
    }
    
    // 아군 데미지 체크
    allies.forEach((ally) => {
      if (!ally.active) return;
      const allyDistance = Phaser.Math.Distance.Between(this.x, this.y, ally.x, ally.y);
      if (allyDistance <= explosionRadius) {
        ally.takeDamage(explosionDamage);
      }
    });
  }
  
  public getRewards(): { experience: number; score: number } {
    return {
      experience: this.enemyType.experienceReward,
      score: this.enemyType.scoreReward
    };
  }

  private updateHealthBar() {
    this.healthBar.clear();

    // Background
    this.healthBar.fillStyle(0x000000);
    this.healthBar.fillRect(-20, -25, 40, 4);

    // Health
    const healthPercent = this.health / this.maxHealth;
    this.healthBar.fillStyle(0xff0000);
    this.healthBar.fillRect(-20, -25, 40 * healthPercent, 4);
  }
}