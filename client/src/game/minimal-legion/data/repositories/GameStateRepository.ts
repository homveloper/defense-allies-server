import { Entity } from '../../types/minimalLegion';

export interface IGameStateRepository {
  getPlayer(): Entity | null;
  setPlayer(player: Entity): void;
  getAllies(): Entity[];
  setAllies(allies: Entity[]): void;
  getProjectiles(): Entity[];
  setProjectiles(projectiles: Entity[]): void;
  getCamera(): { x: number; y: number };
  setCamera(camera: { x: number; y: number }): void;
  getWave(): number;
  setWave(wave: number): void;
  getScore(): number;
  setScore(score: number): void;
}

export class GameStateRepository implements IGameStateRepository {
  private player: Entity | null = null;
  private allies: Entity[] = [];
  private projectiles: Entity[] = [];
  private camera = { x: 0, y: 0 };
  private wave = 1;
  private score = 0;

  getPlayer(): Entity | null {
    return this.player;
  }

  setPlayer(player: Entity): void {
    this.player = { ...player };
  }

  getAllies(): Entity[] {
    return [...this.allies];
  }

  setAllies(allies: Entity[]): void {
    this.allies = allies.map(ally => ({ ...ally }));
  }

  getProjectiles(): Entity[] {
    return [...this.projectiles];
  }

  setProjectiles(projectiles: Entity[]): void {
    this.projectiles = projectiles.map(proj => ({ ...proj }));
  }

  getCamera(): { x: number; y: number } {
    return { ...this.camera };
  }

  setCamera(camera: { x: number; y: number }): void {
    this.camera = { ...camera };
  }

  getWave(): number {
    return this.wave;
  }

  setWave(wave: number): void {
    this.wave = wave;
  }

  getScore(): number {
    return this.score;
  }

  setScore(score: number): void {
    this.score = score;
  }
}