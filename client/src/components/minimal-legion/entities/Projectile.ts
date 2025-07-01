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

    // Add to scene first
    scene.add.existing(this);
    
    // Then enable physics - wait for next frame to ensure proper initialization
    scene.physics.world.enable(this);
    
    // Ensure body exists before configuring it
    if (this.body) {
      this.body.setCircle(4);
      this.body.setCollideWorldBounds(false);
      console.log('Projectile physics body initialized at:', x, y);
    } else {
      console.error('Failed to create physics body for projectile');
    }
    
    console.log('Projectile created at:', x, y, 'Body exists:', !!this.body);
  }

  fire(targetX: number, targetY: number, damage: number) {
    this.damage = damage;

    if (!this.body) {
      console.error('Projectile body not found when firing!');
      return;
    }

    const angle = Phaser.Math.Angle.Between(this.x, this.y, targetX, targetY);
    const velocityX = Math.cos(angle) * this.speed;
    const velocityY = Math.sin(angle) * this.speed;
    
    // Set velocity and verify it was applied
    this.body.setVelocity(velocityX, velocityY);
    
    // Verify velocity was set
    console.log(`Projectile fired from (${this.x.toFixed(1)}, ${this.y.toFixed(1)}) to (${targetX.toFixed(1)}, ${targetY.toFixed(1)})`);
    console.log(`Calculated velocity: (${velocityX.toFixed(1)}, ${velocityY.toFixed(1)})`);
    console.log(`Actual body velocity: (${this.body.velocity.x.toFixed(1)}, ${this.body.velocity.y.toFixed(1)})`);
    console.log(`Body position: (${this.body.x.toFixed(1)}, ${this.body.y.toFixed(1)})`);

    // Destroy after 3 seconds if it doesn't hit anything
    this.scene.time.delayedCall(3000, () => {
      if (this.active) {
        console.log('Projectile auto-destroyed after timeout');
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