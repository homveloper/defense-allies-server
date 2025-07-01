import Phaser from 'phaser';
import { Player } from '../entities/Player';
import { Enemy } from '../entities/Enemy';
import { Projectile } from '../entities/Projectile';
import { Ally } from '../entities/Ally';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';

export class MainScene extends Phaser.Scene {
  private player!: Player;
  private enemies!: Phaser.Physics.Arcade.Group;
  private allies!: Phaser.Physics.Arcade.Group;
  private projectiles!: Phaser.Physics.Arcade.Group;
  private enemyProjectiles!: Phaser.Physics.Arcade.Group;
  private cursors!: Phaser.Types.Input.Keyboard.CursorKeys;
  private wasd!: {
    up: Phaser.Input.Keyboard.Key;
    down: Phaser.Input.Keyboard.Key;
    left: Phaser.Input.Keyboard.Key;
    right: Phaser.Input.Keyboard.Key;
  };
  private waveStartTime: number = 0;
  private enemySpawnTimer: number = 0;
  private hudText!: Phaser.GameObjects.Text;

  constructor() {
    super({ key: 'MainScene' });
  }

  create() {
    // Input setup
    this.cursors = this.input.keyboard!.createCursorKeys();
    this.wasd = {
      up: this.input.keyboard!.addKey('W'),
      down: this.input.keyboard!.addKey('S'),
      left: this.input.keyboard!.addKey('A'),
      right: this.input.keyboard!.addKey('D'),
    };

    // Groups
    this.enemies = this.physics.add.group({
      classType: Enemy,
      runChildUpdate: true,
    });

    this.allies = this.physics.add.group({
      classType: Ally,
      runChildUpdate: true,
    });

    this.projectiles = this.physics.add.group({
      classType: Projectile,
      runChildUpdate: true,
    });

    this.enemyProjectiles = this.physics.add.group({
      classType: Projectile,
      runChildUpdate: true,
    });

    // Player
    this.player = new Player(this, 600, 400);
    this.add.existing(this.player);

    // Collisions
    this.physics.add.overlap(
      this.projectiles,
      this.enemies,
      this.handleProjectileEnemyCollision,
      undefined,
      this
    );

    this.physics.add.overlap(
      this.enemyProjectiles,
      this.player,
      this.handleEnemyProjectilePlayerCollision,
      undefined,
      this
    );

    this.physics.add.overlap(
      this.enemyProjectiles,
      this.allies,
      this.handleEnemyProjectileAllyCollision,
      undefined,
      this
    );

    // HUD
    this.createHUD();

    // Start first wave
    this.startWave();
  }

  update(_time: number, delta: number) {
    // Player movement
    const moveX =
      (this.cursors.right.isDown || this.wasd.right.isDown ? 1 : 0) -
      (this.cursors.left.isDown || this.wasd.left.isDown ? 1 : 0);
    const moveY =
      (this.cursors.down.isDown || this.wasd.down.isDown ? 1 : 0) -
      (this.cursors.up.isDown || this.wasd.up.isDown ? 1 : 0);

    this.player.move(moveX, moveY);

    // Find nearest enemy for player
    const nearestEnemy = this.findNearestEnemy(this.player.x, this.player.y);
    if (nearestEnemy) {
      this.player.setTarget(nearestEnemy);
    }

    // Enemy spawning
    this.enemySpawnTimer += delta;
    if (this.enemySpawnTimer > 2000 && this.enemies.countActive() < 20) {
      this.spawnEnemy();
      this.enemySpawnTimer = 0;
    }

    // Update HUD
    this.updateHUD();

    // Check wave completion
    const store = useMinimalLegionStore.getState();
    if (store.enemiesRemaining === 0 && this.enemies.countActive() === 0) {
      this.nextWave();
    }
  }

  private createHUD() {
    const style = {
      font: '16px Arial',
      fill: '#ffffff',
    };

    this.hudText = this.add.text(10, 10, '', style);
    this.hudText.setScrollFactor(0);
  }

  private updateHUD() {
    const store = useMinimalLegionStore.getState();
    const hudInfo = [
      `Wave: ${store.wave}`,
      `Health: ${store.player.health}/${store.player.maxHealth}`,
      `Level: ${store.player.level}`,
      `EXP: ${store.player.experience}/${store.player.experienceToNext}`,
      `Allies: ${store.allies}/${store.maxAllies}`,
      `Score: ${store.score}`,
      `Enemies: ${this.enemies.countActive()}`,
    ];

    this.hudText.setText(hudInfo.join('\n'));
  }

  private startWave() {
    const store = useMinimalLegionStore.getState();
    const enemyCount = 5 + store.wave * 2;
    store.setEnemiesRemaining(enemyCount);
    this.waveStartTime = this.time.now;
  }

  private nextWave() {
    const store = useMinimalLegionStore.getState();
    store.nextWave();
    this.time.delayedCall(3000, () => this.startWave());
  }

  private spawnEnemy() {
    const store = useMinimalLegionStore.getState();
    if (store.enemiesRemaining <= 0) return;

    const side = Phaser.Math.Between(0, 3);
    let x, y;

    switch (side) {
      case 0: // Top
        x = Phaser.Math.Between(0, 1200);
        y = -50;
        break;
      case 1: // Right
        x = 1250;
        y = Phaser.Math.Between(0, 800);
        break;
      case 2: // Bottom
        x = Phaser.Math.Between(0, 1200);
        y = 850;
        break;
      default: // Left
        x = -50;
        y = Phaser.Math.Between(0, 800);
    }

    const enemy = new Enemy(this, x, y);
    this.enemies.add(enemy);
    enemy.setTarget(this.player);

    store.setEnemiesRemaining(store.enemiesRemaining - 1);
  }

  private findNearestEnemy(x: number, y: number): Enemy | null {
    let nearest: Enemy | null = null;
    let nearestDistance = Infinity;

    this.enemies.children.entries.forEach((enemy) => {
      if (enemy.active) {
        const distance = Phaser.Math.Distance.Between(
          x,
          y,
          enemy.body!.position.x,
          enemy.body!.position.y
        );
        if (distance < nearestDistance) {
          nearestDistance = distance;
          nearest = enemy as Enemy;
        }
      }
    });

    return nearest;
  }

  private handleProjectileEnemyCollision(
    projectile: Phaser.Types.Physics.Arcade.GameObjectWithBody,
    enemy: Phaser.Types.Physics.Arcade.GameObjectWithBody
  ) {
    const proj = projectile as Projectile;
    const en = enemy as Enemy;

    en.takeDamage(proj.damage);
    proj.destroy();

    if (en.health <= 0) {
      this.convertEnemyToAlly(en);
    }
  }

  private handleEnemyProjectilePlayerCollision(
    projectile: Phaser.Types.Physics.Arcade.GameObjectWithBody,
    _player: Phaser.Types.Physics.Arcade.GameObjectWithBody
  ) {
    const proj = projectile as Projectile;
    this.player.takeDamage(proj.damage);
    proj.destroy();
  }

  private handleEnemyProjectileAllyCollision(
    projectile: Phaser.Types.Physics.Arcade.GameObjectWithBody,
    ally: Phaser.Types.Physics.Arcade.GameObjectWithBody
  ) {
    const proj = projectile as Projectile;
    const al = ally as Ally;

    al.takeDamage(proj.damage);
    proj.destroy();

    if (al.health <= 0) {
      const store = useMinimalLegionStore.getState();
      store.removeAlly();
    }
  }

  private convertEnemyToAlly(enemy: Enemy) {
    const store = useMinimalLegionStore.getState();
    
    // Add experience
    store.addExperience(10);
    store.updateScore(10);

    // Check if we can add more allies
    if (store.allies < store.maxAllies) {
      const ally = new Ally(this, enemy.x, enemy.y);
      ally.setPlayer(this.player);
      this.allies.add(ally);
      store.addAlly();
    }

    enemy.destroy();
  }

  fireProjectile(x: number, y: number, targetX: number, targetY: number, damage: number, isEnemy: boolean = false) {
    const projectile = new Projectile(this, x, y);
    projectile.fire(targetX, targetY, damage);
    
    if (isEnemy) {
      this.enemyProjectiles.add(projectile);
    } else {
      this.projectiles.add(projectile);
    }
  }

  dealMeleeDamage(target: Phaser.GameObjects.GameObject, damage: number) {
    if (target === this.player) {
      this.player.takeDamage(damage);
    } else if (this.allies.children.entries.includes(target)) {
      const ally = target as Ally;
      ally.takeDamage(damage);
      if (ally.health <= 0) {
        const store = useMinimalLegionStore.getState();
        store.removeAlly();
      }
    }
  }
}