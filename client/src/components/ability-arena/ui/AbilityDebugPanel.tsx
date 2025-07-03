'use client';

import { useState, useEffect } from 'react';
import { useAbilityArenaStore } from '@/store/abilityArenaStore';

interface AbilityDebugInfo {
  // Current Random Ability
  currentAbility: {
    name: string;
    id: string;
    description: string;
    cooldown: number;
    manaCost: number;
    tags: string[];
  } | null;
  
  // Player Ability System State
  attributes: Record<string, {
    baseValue: number;
    currentValue: number;
    maxValue?: number;
    modifiers: any[];
  }>;
  
  // Active Effects
  activeEffects: Array<{
    id: string;
    name: string;
    duration: number;
    remaining: number;
    stacks: number;
    type: string;
  }>;
  
  // Cooldowns
  cooldowns: Record<string, {
    remaining: number;
    total: number;
    percentage: number;
  }>;
  
  // Granted Abilities
  grantedAbilities: Array<{
    id: string;
    name: string;
    description: string;
  }>;
  
  // Tags
  activeTags: string[];
}

export function AbilityDebugPanel() {
  const store = useAbilityArenaStore();
  const [debugInfo, setDebugInfo] = useState<AbilityDebugInfo | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);
  const [selectedTab, setSelectedTab] = useState<'ability' | 'attributes' | 'effects' | 'cooldowns' | 'abilities' | 'tags'>('ability');

  // Get debug info from the game scene
  useEffect(() => {
    const interval = setInterval(() => {
      if (typeof window !== 'undefined' && (window as any).currentArenaPlayer) {
        const player = (window as any).currentArenaPlayer;
        const abilitySystem = player.abilitySystem;
        
        if (abilitySystem) {
          const debugData = abilitySystem.getDebugInfo();
          const currentRandomAbility = player.currentRandomAbility;
          
          const info: AbilityDebugInfo = {
            currentAbility: currentRandomAbility ? {
              name: currentRandomAbility.name,
              id: currentRandomAbility.id,
              description: currentRandomAbility.description,
              cooldown: currentRandomAbility.cooldown,
              manaCost: currentRandomAbility.manaCost || 0,
              tags: currentRandomAbility.tags || []
            } : null,
            attributes: debugData.attributes || {},
            activeEffects: debugData.activeEffects || [],
            cooldowns: debugData.cooldowns || {},
            grantedAbilities: debugData.abilities || [],
            activeTags: debugData.tags || []
          };
          
          setDebugInfo(info);
        }
      }
    }, 500); // Update every 500ms

    return () => clearInterval(interval);
  }, []);

  if (!store.isGameStarted || store.isGameOver) {
    return null;
  }

  return (
    <div className="fixed left-4 top-1/2 transform -translate-y-1/2 z-50">
      {/* Toggle Button */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="bg-gray-800 hover:bg-gray-700 text-white px-3 py-2 rounded-r-lg shadow-lg transition-colors"
        style={{ writingMode: 'vertical-rl', textOrientation: 'mixed' }}
      >
        {isExpanded ? '◀' : '▶'} Debug Panel
      </button>

      {/* Debug Panel */}
      {isExpanded && (
        <div className="bg-gray-900/95 text-white rounded-lg shadow-xl p-4 ml-1 w-80 max-h-96 overflow-hidden flex flex-col">
          {/* Header */}
          <div className="flex justify-between items-center mb-3">
            <h3 className="text-lg font-bold text-green-400">Ability Debug</h3>
            <button
              onClick={() => setIsExpanded(false)}
              className="text-gray-400 hover:text-white text-xl"
            >
              ×
            </button>
          </div>

          {/* Tab Navigation */}
          <div className="flex flex-wrap gap-1 mb-3 text-xs">
            {[
              { key: 'ability', label: 'Current' },
              { key: 'attributes', label: 'Attrs' },
              { key: 'effects', label: 'Effects' },
              { key: 'cooldowns', label: 'Cooldowns' },
              { key: 'abilities', label: 'Abilities' },
              { key: 'tags', label: 'Tags' }
            ].map(tab => (
              <button
                key={tab.key}
                onClick={() => setSelectedTab(tab.key as any)}
                className={`px-2 py-1 rounded ${
                  selectedTab === tab.key 
                    ? 'bg-blue-600 text-white' 
                    : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto text-xs">
            {selectedTab === 'ability' && (
              <div className="space-y-2">
                <h4 className="text-yellow-400 font-semibold">Current Random Ability</h4>
                {debugInfo?.currentAbility ? (
                  <div className="bg-gray-800 p-2 rounded">
                    <div><strong>Name:</strong> {debugInfo.currentAbility.name}</div>
                    <div><strong>ID:</strong> {debugInfo.currentAbility.id}</div>
                    <div><strong>Description:</strong> {debugInfo.currentAbility.description}</div>
                    <div><strong>Cooldown:</strong> {debugInfo.currentAbility.cooldown}ms</div>
                    <div><strong>Mana Cost:</strong> {debugInfo.currentAbility.manaCost}</div>
                    <div><strong>Tags:</strong> {debugInfo.currentAbility.tags.join(', ')}</div>
                  </div>
                ) : (
                  <div className="text-gray-400">No current ability</div>
                )}
              </div>
            )}

            {selectedTab === 'attributes' && (
              <div className="space-y-2">
                <h4 className="text-yellow-400 font-semibold">Player Attributes</h4>
                {debugInfo?.attributes && Object.entries(debugInfo.attributes).map(([name, attr]) => (
                  <div key={name} className="bg-gray-800 p-2 rounded">
                    <div className="font-semibold text-blue-300">{name}</div>
                    <div>Base: {attr.baseValue}</div>
                    <div>Current: {attr.currentValue}</div>
                    {attr.maxValue && <div>Max: {attr.maxValue}</div>}
                    {attr.modifiers.length > 0 && (
                      <div>Modifiers: {attr.modifiers.length}</div>
                    )}
                  </div>
                ))}
              </div>
            )}

            {selectedTab === 'effects' && (
              <div className="space-y-2">
                <h4 className="text-yellow-400 font-semibold">Active Effects</h4>
                {debugInfo?.activeEffects.length ? (
                  debugInfo.activeEffects.map((effect, index) => (
                    <div key={index} className="bg-gray-800 p-2 rounded">
                      <div className="font-semibold text-purple-300">{effect.name}</div>
                      <div>ID: {effect.id}</div>
                      <div>Duration: {effect.duration}ms</div>
                      <div>Remaining: {effect.remaining}ms</div>
                      <div>Stacks: {effect.stacks}</div>
                      <div>Type: {effect.type}</div>
                    </div>
                  ))
                ) : (
                  <div className="text-gray-400">No active effects</div>
                )}
              </div>
            )}

            {selectedTab === 'cooldowns' && (
              <div className="space-y-2">
                <h4 className="text-yellow-400 font-semibold">Ability Cooldowns</h4>
                {debugInfo?.cooldowns && Object.entries(debugInfo.cooldowns).map(([abilityId, cooldown]) => (
                  <div key={abilityId} className="bg-gray-800 p-2 rounded">
                    <div className="font-semibold text-orange-300">{abilityId}</div>
                    <div>Remaining: {Math.ceil(cooldown.remaining / 1000)}s</div>
                    <div>Total: {Math.ceil(cooldown.total / 1000)}s</div>
                    <div className="w-full bg-gray-600 rounded-full h-2 mt-1">
                      <div 
                        className="bg-orange-500 h-2 rounded-full transition-all duration-100" 
                        style={{ width: `${Math.max(0, 100 - cooldown.percentage)}%` }}
                      ></div>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {selectedTab === 'abilities' && (
              <div className="space-y-2">
                <h4 className="text-yellow-400 font-semibold">Granted Abilities</h4>
                {debugInfo?.grantedAbilities.length ? (
                  debugInfo.grantedAbilities.map((ability, index) => (
                    <div key={index} className="bg-gray-800 p-2 rounded">
                      <div className="font-semibold text-green-300">{ability.name}</div>
                      <div>ID: {ability.id}</div>
                      <div>Description: {ability.description}</div>
                    </div>
                  ))
                ) : (
                  <div className="text-gray-400">No granted abilities</div>
                )}
              </div>
            )}

            {selectedTab === 'tags' && (
              <div className="space-y-2">
                <h4 className="text-yellow-400 font-semibold">Active Tags</h4>
                {debugInfo?.activeTags.length ? (
                  <div className="flex flex-wrap gap-1">
                    {debugInfo.activeTags.map((tag, index) => (
                      <span 
                        key={index} 
                        className="bg-blue-600 px-2 py-1 rounded text-xs"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                ) : (
                  <div className="text-gray-400">No active tags</div>
                )}
              </div>
            )}
          </div>

          {/* Footer */}
          <div className="mt-2 pt-2 border-t border-gray-700 text-xs text-gray-400">
            Updates every 500ms
          </div>
        </div>
      )}
    </div>
  );
}