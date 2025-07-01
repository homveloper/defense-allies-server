import Phaser from 'phaser';

export class Projectile extends Phaser.GameObjects.Graphics {
  public body!: Phaser.Physics.Arcade.Body;
  public damage: number = 0;
  private speed: number = 600;

  constructor(scene: Phaser.Scene, x: number, y: number) {
    super(scene, { x, y });

    // Draw projectile (small yellow circle)
    this.fillStyle(0xf1c40f);
    this.fillCircle(0, 0, 4);

    // Physics
    scene.physics.world.enable(this);
    this.body.setCircle(4);

    scene.add.existing(this);
  }

  fire(targetX: number, targetY: number, damage: number) {
    this.damage = damage;

    const angle = Phaser.Math.Angle.Between(this.x, this.y, targetX, targetY);
    this.body.setVelocity(
      Math.cos(angle) * this.speed,
      Math.sin(angle) * this.speed
    );

    // Destroy after 2 seconds if it doesn't hit anything
    this.scene.time.delayedCall(2000, () => {
      if (this.active) {
        this.destroy();
      }
    });
  }

  update() {
    // Check if projectile is out of bounds
    if (
      this.x < -50 ||
      this.x > 1250 ||
      this.y < -50 ||
      this.y > 850
    ) {
      this.destroy();
    }
  }
}