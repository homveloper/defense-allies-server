'use client';

import { useState, useEffect } from 'react';
import { useAbilityArenaStore } from '@/store/abilityArenaStore';

interface DevSettings {
  // Player Settings
  playerInvincible: boolean;
  playerInfiniteHealth: boolean;
  playerInfiniteMana: boolean;
  playerSpeedMultiplier: number;
  
  // Enemy Settings
  enemyInvincible: boolean;
  enemySpawnEnabled: boolean;
  enemySpawnRate: number; // enemies per second
  enemySpeedMultiplier: number;
  enemyHealthMultiplier: number;
  enemyDamageMultiplier: number;
  maxEnemyCount: number;
  
  // Wave Settings
  waveProgressionEnabled: boolean;
  currentWaveOverride: number | null;
  
  // Gameplay Settings
  timeScale: number;
  debugMode: boolean;
}

export function DevControlPanel() {
  const store = useAbilityArenaStore();
  const [isExpanded, setIsExpanded] = useState(false);
  const [selectedTab, setSelectedTab] = useState<'player' | 'enemy' | 'spawn' | 'wave' | 'debug'>('player');
  const [settings, setSettings] = useState<DevSettings>({
    playerInvincible: false,
    playerInfiniteHealth: false,
    playerInfiniteMana: false,
    playerSpeedMultiplier: 1.0,
    enemyInvincible: false,
    enemySpawnEnabled: true,
    enemySpawnRate: 1.0,
    enemySpeedMultiplier: 1.0,
    enemyHealthMultiplier: 1.0,
    enemyDamageMultiplier: 1.0,
    maxEnemyCount: 50,
    waveProgressionEnabled: true,
    currentWaveOverride: null,
    timeScale: 1.0,
    debugMode: false
  });

  // Apply settings to game
  useEffect(() => {
    if (typeof window !== 'undefined') {
      (window as any).devSettings = settings;
      
      // Apply time scale to Phaser game
      const game = (window as any).phaserGame;
      if (game) {
        game.loop.timeScale = settings.timeScale;
      }
      
      // Apply to current player if exists
      const player = (window as any).currentArenaPlayer;
      if (player) {
        // Player invincibility
        if (player.abilitySystem) {
          if (settings.playerInvincible) {
            player.abilitySystem.addTag('invincible');
          } else {
            player.abilitySystem.removeTag('invincible');
          }
        }
        
        // Speed multiplier
        if (player.body) {
          (player as any).speedMultiplier = settings.playerSpeedMultiplier;
        }
      }
      
      // Apply to current scene if exists
      const scene = (window as any).currentArenaScene;
      if (scene) {
        // Update enemy spawn settings
        if (scene.enemySpawner) {
          scene.enemySpawner.setSpawnEnabled(settings.enemySpawnEnabled);
          scene.enemySpawner.setSpawnRate(settings.enemySpawnRate);
          scene.enemySpawner.setMaxEnemyCount(settings.maxEnemyCount);
        }
        
        // Apply enemy settings to existing enemies
        if (scene.enemies) {
          scene.enemies.children.entries.forEach((enemy: any) => {
            if (enemy.abilitySystem) {
              if (settings.enemyInvincible) {
                enemy.abilitySystem.addTag('invincible');
              } else {
                enemy.abilitySystem.removeTag('invincible');
              }
            }
            
            // Apply multipliers
            (enemy as any).speedMultiplier = settings.enemySpeedMultiplier;
            (enemy as any).healthMultiplier = settings.enemyHealthMultiplier;
            (enemy as any).damageMultiplier = settings.enemyDamageMultiplier;
          });
        }
      }
    }
  }, [settings]);

  const updateSetting = <K extends keyof DevSettings>(key: K, value: DevSettings[K]) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  };

  const resetSettings = () => {
    setSettings({
      playerInvincible: false,
      playerInfiniteHealth: false,
      playerInfiniteMana: false,
      playerSpeedMultiplier: 1.0,
      enemyInvincible: false,
      enemySpawnEnabled: true,
      enemySpawnRate: 1.0,
      enemySpeedMultiplier: 1.0,
      enemyHealthMultiplier: 1.0,
      enemyDamageMultiplier: 1.0,
      maxEnemyCount: 50,
      waveProgressionEnabled: true,
      currentWaveOverride: null,
      timeScale: 1.0,
      debugMode: false
    });
  };

  const killAllEnemies = () => {
    if (typeof window !== 'undefined') {
      const scene = (window as any).currentArenaScene;
      if (scene && scene.enemies) {
        scene.enemies.children.entries.forEach((enemy: any) => {
          if (enemy.active && enemy.takeDamage) {
            enemy.takeDamage(9999);
          }
        });
      }
    }
  };

  const spawnWave = () => {
    if (typeof window !== 'undefined') {
      const scene = (window as any).currentArenaScene;
      if (scene && scene.enemySpawner) {
        scene.enemySpawner.spawnWave();
      }
    }
  };

  if (!store.isGameStarted || store.isGameOver) {
    return null;
  }

  return (
    <div className="fixed right-4 top-1/2 transform -translate-y-1/2 z-50">
      {/* Toggle Button */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="bg-red-800 hover:bg-red-700 text-white px-3 py-2 rounded-l-lg shadow-lg transition-colors"
        style={{ writingMode: 'vertical-rl', textOrientation: 'mixed' }}
      >
        {isExpanded ? '▶' : '◀'} Dev Panel
      </button>

      {/* Control Panel */}
      {isExpanded && (
        <div className="bg-red-900/95 text-white rounded-lg shadow-xl p-4 mr-1 w-80 max-h-96 overflow-hidden flex flex-col">
          {/* Header */}
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-lg font-bold text-red-400">🛠️ Dev Controls</h3>
            <div className="flex gap-2">
              <button
                onClick={resetSettings}
                className="text-xs bg-gray-700 hover:bg-gray-600 px-2 py-1 rounded"
                title="Reset all settings"
              >
                Reset
              </button>
              <button
                onClick={() => setIsExpanded(false)}
                className="text-gray-400 hover:text-white text-xl"
              >
                ×
              </button>
            </div>
          </div>

          {/* Tab Navigation */}
          <div className="flex flex-wrap gap-1 mb-3 text-xs">
            {[
              { key: 'player', label: 'Player' },
              { key: 'enemy', label: 'Enemy' },
              { key: 'spawn', label: 'Spawn' },
              { key: 'wave', label: 'Wave' },
              { key: 'debug', label: 'Debug' }
            ].map(tab => (
              <button
                key={tab.key}
                onClick={() => setSelectedTab(tab.key as any)}
                className={`px-2 py-1 rounded ${
                  selectedTab === tab.key 
                    ? 'bg-red-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto text-xs space-y-3">
            {selectedTab === 'player' && (
              <div>
                <h4 className="text-yellow-400 font-semibold mb-2">👤 Player Settings</h4>
                
                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.playerInvincible}
                    onChange={(e) => updateSetting('playerInvincible', e.target.checked)}
                    className="rounded"
                  />
                  <span>Invincible</span>
                </label>

                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.playerInfiniteHealth}
                    onChange={(e) => updateSetting('playerInfiniteHealth', e.target.checked)}
                    className="rounded"
                  />
                  <span>Infinite Health</span>
                </label>

                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.playerInfiniteMana}
                    onChange={(e) => updateSetting('playerInfiniteMana', e.target.checked)}
                    className="rounded"
                  />
                  <span>Infinite Mana</span>
                </label>

                <div className="mb-2">
                  <label className="block mb-1">Speed Multiplier: {settings.playerSpeedMultiplier}x</label>
                  <input
                    type="range"
                    min="0.1"
                    max="5.0"
                    step="0.1"
                    value={settings.playerSpeedMultiplier}
                    onChange={(e) => updateSetting('playerSpeedMultiplier', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>
              </div>
            )}

            {selectedTab === 'enemy' && (
              <div>
                <h4 className="text-yellow-400 font-semibold mb-2">👹 Enemy Settings</h4>
                
                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.enemyInvincible}
                    onChange={(e) => updateSetting('enemyInvincible', e.target.checked)}
                    className="rounded"
                  />
                  <span>Enemy Invincible</span>
                </label>

                <div className="mb-2">
                  <label className="block mb-1">Speed: {settings.enemySpeedMultiplier}x</label>
                  <input
                    type="range"
                    min="0.1"
                    max="3.0"
                    step="0.1"
                    value={settings.enemySpeedMultiplier}
                    onChange={(e) => updateSetting('enemySpeedMultiplier', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div className="mb-2">
                  <label className="block mb-1">Health: {settings.enemyHealthMultiplier}x</label>
                  <input
                    type="range"
                    min="0.1"
                    max="5.0"
                    step="0.1"
                    value={settings.enemyHealthMultiplier}
                    onChange={(e) => updateSetting('enemyHealthMultiplier', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div className="mb-2">
                  <label className="block mb-1">Damage: {settings.enemyDamageMultiplier}x</label>
                  <input
                    type="range"
                    min="0.1"
                    max="5.0"
                    step="0.1"
                    value={settings.enemyDamageMultiplier}
                    onChange={(e) => updateSetting('enemyDamageMultiplier', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>

                <button
                  onClick={killAllEnemies}
                  className="w-full bg-red-600 hover:bg-red-700 py-2 rounded mt-2"
                >
                  💀 Kill All Enemies
                </button>
              </div>
            )}

            {selectedTab === 'spawn' && (
              <div>
                <h4 className="text-yellow-400 font-semibold mb-2">🎯 Spawn Settings</h4>
                
                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.enemySpawnEnabled}
                    onChange={(e) => updateSetting('enemySpawnEnabled', e.target.checked)}
                    className="rounded"
                  />
                  <span>Enemy Spawning</span>
                </label>

                <div className="mb-2">
                  <label className="block mb-1">Spawn Rate: {settings.enemySpawnRate}x</label>
                  <input
                    type="range"
                    min="0.1"
                    max="5.0"
                    step="0.1"
                    value={settings.enemySpawnRate}
                    onChange={(e) => updateSetting('enemySpawnRate', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div className="mb-2">
                  <label className="block mb-1">Max Enemies: {settings.maxEnemyCount}</label>
                  <input
                    type="range"
                    min="1"
                    max="200"
                    step="1"
                    value={settings.maxEnemyCount}
                    onChange={(e) => updateSetting('maxEnemyCount', parseInt(e.target.value))}
                    className="w-full"
                  />
                </div>

                <button
                  onClick={spawnWave}
                  className="w-full bg-blue-600 hover:bg-blue-700 py-2 rounded mt-2"
                >
                  🌊 Spawn Wave
                </button>
              </div>
            )}

            {selectedTab === 'wave' && (
              <div>
                <h4 className="text-yellow-400 font-semibold mb-2">🌊 Wave Settings</h4>
                
                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.waveProgressionEnabled}
                    onChange={(e) => updateSetting('waveProgressionEnabled', e.target.checked)}
                    className="rounded"
                  />
                  <span>Wave Progression</span>
                </label>

                <div className="mb-2">
                  <label className="block mb-1">Current Wave: {store.stats.waveNumber}</label>
                  <div className="flex gap-2">
                    <button
                      onClick={() => store.incrementWave()}
                      className="flex-1 bg-green-600 hover:bg-green-700 py-1 rounded text-xs"
                    >
                      Next Wave
                    </button>
                    <button
                      onClick={() => {
                        if (store.stats.waveNumber > 1) {
                          // No direct decrement method, so we'd need to add one
                          console.log('Previous wave functionality would need store update');
                        }
                      }}
                      className="flex-1 bg-orange-600 hover:bg-orange-700 py-1 rounded text-xs"
                      disabled={store.stats.waveNumber <= 1}
                    >
                      Prev Wave
                    </button>
                  </div>
                </div>

                <div className="mb-2">
                  <label className="block mb-1">Override Wave (0 = off)</label>
                  <input
                    type="number"
                    min="0"
                    max="999"
                    value={settings.currentWaveOverride || 0}
                    onChange={(e) => {
                      const val = parseInt(e.target.value);
                      updateSetting('currentWaveOverride', val === 0 ? null : val);
                    }}
                    className="w-full bg-gray-700 text-white p-1 rounded"
                  />
                </div>
              </div>
            )}

            {selectedTab === 'debug' && (
              <div>
                <h4 className="text-yellow-400 font-semibold mb-2">🐛 Debug Settings</h4>
                
                <label className="flex items-center gap-2 mb-2">
                  <input
                    type="checkbox"
                    checked={settings.debugMode}
                    onChange={(e) => updateSetting('debugMode', e.target.checked)}
                    className="rounded"
                  />
                  <span>Debug Mode</span>
                </label>

                <div className="mb-2">
                  <label className="block mb-1">Time Scale: {settings.timeScale}x</label>
                  <input
                    type="range"
                    min="0.1"
                    max="3.0"
                    step="0.1"
                    value={settings.timeScale}
                    onChange={(e) => updateSetting('timeScale', parseFloat(e.target.value))}
                    className="w-full"
                  />
                </div>

                <div className="space-y-1 text-xs">
                  <div>Current FPS: {typeof window !== 'undefined' ? '60' : 'N/A'}</div>
                  <div>Active Enemies: {typeof window !== 'undefined' && (window as any).currentArenaScene?.enemies?.children?.entries?.length || 0}</div>
                  <div>Player Health: {store.player.health}/{store.player.maxHealth}</div>
                  <div>Player Mana: {store.player.mana}/{store.player.maxMana}</div>
                </div>
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="mt-2 pt-2 border-t border-red-700 text-xs text-gray-400">
            🚨 Development Tools
          </div>
        </div>
      )}
    </div>
  );
}