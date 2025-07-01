import Phaser from 'phaser';
import { Player } from './Player';

export class RotatingOrb extends Phaser.GameObjects.Graphics {
  public body!: Phaser.Physics.Arcade.Body;
  private player: Player;
  private rotationAngle: number = 0;
  private radius: number = 60;
  private rotationSpeed: number = 3; // degrees per frame
  private damage: number = 15;
  private lastDamageTime: number = 0;
  private damageCooldown: number = 500; // 0.5초 쿨다운

  constructor(scene: Phaser.Scene, player: Player, startAngle: number = 0) {
    super(scene, { x: player.x, y: player.y });

    this.player = player;
    this.rotationAngle = startAngle;

    // Create orb visual (glowing blue circle)
    this.fillStyle(0x00ffff, 0.8);
    this.fillCircle(0, 0, 8);
    
    // Add glow effect
    this.lineStyle(2, 0x00ffff, 0.6);
    this.strokeCircle(0, 0, 12);

    // Physics
    scene.physics.world.enable(this);
    this.body.setCircle(8);
    this.body.setCollideWorldBounds(false);

    scene.add.existing(this);
  }

  update() {
    if (!this.player || !this.player.active) {
      this.destroy();
      return;
    }

    // 플레이어 주변을 회전
    this.rotationAngle += this.rotationSpeed;
    if (this.rotationAngle >= 360) {
      this.rotationAngle -= 360;
    }

    // 각도를 라디안으로 변환하고 위치 계산
    const radians = Phaser.Math.DegToRad(this.rotationAngle);
    const newX = this.player.x + Math.cos(radians) * this.radius;
    const newY = this.player.y + Math.sin(radians) * this.radius;

    this.setPosition(newX, newY);

    // 시각적 효과를 위한 약간의 펄스 효과
    const pulse = Math.sin(this.rotationAngle * 0.1) * 0.1 + 1;
    this.setScale(pulse);
  }

  // 적과 충돌했을 때 호출
  hitEnemy(enemy: { takeDamage?: (amount: number) => void }) {
    const now = this.scene.time.now;
    if (now - this.lastDamageTime > this.damageCooldown) {
      // 데미지 주기
      if (enemy && typeof enemy.takeDamage === 'function') {
        enemy.takeDamage(this.damage);
        this.lastDamageTime = now;

        // 타격 효과
        this.createHitEffect();
        
        console.log(`Rotating orb hit enemy for ${this.damage} damage`);
      }
    }
  }

  private createHitEffect() {
    // 타격 시 반짝이는 효과
    this.clear();
    this.fillStyle(0xffffff, 1);
    this.fillCircle(0, 0, 12);
    
    this.scene.time.delayedCall(100, () => {
      if (this.active) {
        this.clear();
        this.fillStyle(0x00ffff, 0.8);
        this.fillCircle(0, 0, 8);
        
        this.lineStyle(2, 0x00ffff, 0.6);
        this.strokeCircle(0, 0, 12);
      }
    });
  }

  getDamage(): number {
    return this.damage;
  }
}