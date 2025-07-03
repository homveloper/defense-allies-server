'use client';

import React, { useState, useEffect, useRef } from 'react';
import { TacticalArenaEngine } from './engine/TacticalArenaEngine';
import { TacticalUnit } from './entities/TacticalUnit';
import { TurnPhase } from '../../../packages/gas/v2/turn-based/TurnBasedContext';

interface TacticalArenaGameProps {
  width?: number;
  height?: number;
  playerUnits?: number;
  enemyUnits?: number;
  mapSize?: { width: number; height: number };
}

export const TacticalArenaGame: React.FC<TacticalArenaGameProps> = ({
  width = 1000,
  height = 700,
  playerUnits = 2,
  enemyUnits = 2,
  mapSize = { width: 8, height: 6 }
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const engineRef = useRef<TacticalArenaEngine | null>(null);
  
  // Game State
  const [isInitialized, setIsInitialized] = useState(false);
  const [gameState, setGameState] = useState({
    currentTurn: 1,
    currentRound: 1,
    activeUnit: null as TacticalUnit | null,
    currentPhase: TurnPhase.START,
    turnOrder: [] as TacticalUnit[],
    selectedUnit: null as TacticalUnit | null,
    gameResult: null as 'player_victory' | 'enemy_victory' | 'draw' | null
  });

  // UI State
  const [uiState, setUiState] = useState({
    showGridNumbers: true,
    showMovementRange: true,
    showAttackRange: true,
    showTurnOrder: true,
    selectedAction: null as string | null,
    hoveredTile: null as { x: number; y: number } | null
  });

  // Initialize game engine
  useEffect(() => {
    if (!canvasRef.current || isInitialized) return;

    const engine = new TacticalArenaEngine(canvasRef.current, {
      width,
      height,
      mapSize,
      playerUnits,
      enemyUnits
    });

    // Set up event listeners
    engine.on('game-state-changed', (newState) => {
      setGameState(prev => ({ ...prev, ...newState }));
    });

    engine.on('ui-state-changed', (newState) => {
      setUiState(prev => ({ ...prev, ...newState }));
    });

    engine.on('unit-selected', (unit: TacticalUnit) => {
      setGameState(prev => ({ ...prev, selectedUnit: unit }));
    });

    engine.on('action-completed', ({ unitId, actionId, success }) => {
      console.log(`Unit ${unitId} ${success ? 'completed' : 'failed'} action: ${actionId}`);
    });

    engine.on('game-ended', (result) => {
      setGameState(prev => ({ ...prev, gameResult: result }));
    });

    engineRef.current = engine;
    engine.initialize();
    setIsInitialized(true);

    return () => {
      engine.destroy();
    };
  }, [width, height, mapSize, playerUnits, enemyUnits, isInitialized]);

  // Action handlers
  const handleAction = (actionId: string) => {
    if (engineRef.current && gameState.selectedUnit) {
      engineRef.current.executeAction(gameState.selectedUnit.id, actionId);
    }
  };

  const handleEndTurn = () => {
    if (engineRef.current) {
      engineRef.current.endCurrentTurn();
    }
  };

  const handleEndPhase = () => {
    if (engineRef.current) {
      engineRef.current.advancePhase();
    }
  };

  const handleRestart = () => {
    if (engineRef.current) {
      engineRef.current.restart();
      setGameState(prev => ({ ...prev, gameResult: null }));
    }
  };

  // Render available actions for selected unit
  const renderUnitActions = () => {
    if (!gameState.selectedUnit || !gameState.activeUnit) return null;
    
    const isActiveUnit = gameState.selectedUnit.id === gameState.activeUnit.id;
    if (!isActiveUnit) return null;

    const unit = gameState.selectedUnit;
    const availableActions = unit.getAvailableActions(gameState.currentPhase);

    return (
      <div className="bg-gray-800 p-4 rounded-lg">
        <h3 className="text-white text-lg font-bold mb-3">
          {unit.name} - Actions
        </h3>
        <div className="grid grid-cols-2 gap-2">
          {availableActions.map(action => (
            <button
              key={action.id}
              onClick={() => handleAction(action.id)}
              disabled={!action.canUse}
              className={`p-2 rounded text-sm font-medium transition-colors ${
                action.canUse 
                  ? 'bg-blue-600 hover:bg-blue-700 text-white' 
                  : 'bg-gray-600 text-gray-400 cursor-not-allowed'
              }`}
              title={action.description}
            >
              {action.name}
              {action.cost && (
                <div className="text-xs opacity-75">
                  Cost: {Object.entries(action.cost).map(([res, amount]) => 
                    `${res}: ${amount}`
                  ).join(', ')}
                </div>
              )}
            </button>
          ))}
        </div>
      </div>
    );
  };

  // Render turn order
  const renderTurnOrder = () => {
    if (!uiState.showTurnOrder || gameState.turnOrder.length === 0) return null;

    return (
      <div className="bg-gray-800 p-4 rounded-lg">
        <h3 className="text-white text-lg font-bold mb-3">Turn Order</h3>
        <div className="space-y-2">
          {gameState.turnOrder.map((unit, index) => (
            <div
              key={unit.id}
              className={`p-2 rounded flex items-center justify-between ${
                unit.id === gameState.activeUnit?.id
                  ? 'bg-yellow-600 text-black'
                  : unit.faction === 'player'
                  ? 'bg-blue-600 text-white'
                  : 'bg-red-600 text-white'
              }`}
            >
              <span className="font-medium">
                {index + 1}. {unit.name}
              </span>
              <div className="text-sm">
                HP: {unit.stats.health}/{unit.stats.maxHealth}
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // Render game stats
  const renderGameStats = () => {
    return (
      <div className="bg-gray-800 p-4 rounded-lg">
        <h3 className="text-white text-lg font-bold mb-3">Game Status</h3>
        <div className="space-y-2 text-white">
          <div>Round: {gameState.currentRound}</div>
          <div>Turn: {gameState.currentTurn}</div>
          <div>Phase: {gameState.currentPhase}</div>
          {gameState.activeUnit && (
            <div>Active: {gameState.activeUnit.name}</div>
          )}
        </div>
        
        <div className="mt-4 space-y-2">
          <button
            onClick={handleEndPhase}
            className="w-full bg-orange-600 hover:bg-orange-700 text-white py-2 px-4 rounded transition-colors"
          >
            End Phase
          </button>
          <button
            onClick={handleEndTurn}
            className="w-full bg-green-600 hover:bg-green-700 text-white py-2 px-4 rounded transition-colors"
          >
            End Turn
          </button>
        </div>
      </div>
    );
  };

  // Render unit details
  const renderUnitDetails = () => {
    const unit = gameState.selectedUnit;
    if (!unit) return null;

    const resources = unit.getResourceSummary();

    return (
      <div className="bg-gray-800 p-4 rounded-lg">
        <h3 className="text-white text-lg font-bold mb-3">
          {unit.name}
          <span className={`ml-2 px-2 py-1 rounded text-xs ${
            unit.faction === 'player' ? 'bg-blue-600' : 'bg-red-600'
          }`}>
            {unit.faction.toUpperCase()}
          </span>
        </h3>
        
        <div className="space-y-2 text-white text-sm">
          <div>Position: ({unit.position.x}, {unit.position.y})</div>
          <div>Health: {unit.stats.health}/{unit.stats.maxHealth}</div>
          <div>Armor: {unit.stats.armor}</div>
          <div>Accuracy: {unit.stats.accuracy}%</div>
        </div>

        <div className="mt-3">
          <h4 className="text-white font-medium mb-2">Resources</h4>
          <div className="space-y-1 text-sm">
            {Object.entries(resources).map(([resourceId, info]) => (
              <div key={resourceId} className="flex justify-between text-white">
                <span className="capitalize">{resourceId.replace('_', ' ')}:</span>
                <span>{info.current}/{info.max}</span>
              </div>
            ))}
          </div>
        </div>

        {unit.statusEffects.length > 0 && (
          <div className="mt-3">
            <h4 className="text-white font-medium mb-2">Status Effects</h4>
            <div className="space-y-1">
              {unit.statusEffects.map((effect, index) => (
                <div key={index} className="text-xs bg-purple-600 text-white px-2 py-1 rounded">
                  {effect.name} ({effect.duration} turns)
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

  // Render settings panel
  const renderSettings = () => {
    return (
      <div className="bg-gray-800 p-4 rounded-lg">
        <h3 className="text-white text-lg font-bold mb-3">Display Settings</h3>
        <div className="space-y-2">
          <label className="flex items-center text-white">
            <input
              type="checkbox"
              checked={uiState.showGridNumbers}
              onChange={(e) => setUiState(prev => ({ ...prev, showGridNumbers: e.target.checked }))}
              className="mr-2"
            />
            Show Grid Numbers
          </label>
          <label className="flex items-center text-white">
            <input
              type="checkbox"
              checked={uiState.showMovementRange}
              onChange={(e) => setUiState(prev => ({ ...prev, showMovementRange: e.target.checked }))}
              className="mr-2"
            />
            Show Movement Range
          </label>
          <label className="flex items-center text-white">
            <input
              type="checkbox"
              checked={uiState.showAttackRange}
              onChange={(e) => setUiState(prev => ({ ...prev, showAttackRange: e.target.checked }))}
              className="mr-2"
            />
            Show Attack Range
          </label>
          <label className="flex items-center text-white">
            <input
              type="checkbox"
              checked={uiState.showTurnOrder}
              onChange={(e) => setUiState(prev => ({ ...prev, showTurnOrder: e.target.checked }))}
              className="mr-2"
            />
            Show Turn Order
          </label>
        </div>
      </div>
    );
  };

  // Render game over screen
  const renderGameOver = () => {
    if (!gameState.gameResult) return null;

    const resultText = {
      player_victory: '🎉 Player Victory!',
      enemy_victory: '💀 Enemy Victory!',
      draw: '🤝 Draw!'
    }[gameState.gameResult];

    const resultColor = {
      player_victory: 'text-green-400',
      enemy_victory: 'text-red-400',
      draw: 'text-yellow-400'
    }[gameState.gameResult];

    return (
      <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50">
        <div className="bg-gray-800 p-8 rounded-lg text-center">
          <h2 className={`text-3xl font-bold mb-4 ${resultColor}`}>
            {resultText}
          </h2>
          <p className="text-white mb-6">
            Battle lasted {gameState.currentRound} rounds, {gameState.currentTurn} turns
          </p>
          <button
            onClick={handleRestart}
            className="bg-blue-600 hover:bg-blue-700 text-white py-2 px-6 rounded transition-colors"
          >
            Play Again
          </button>
        </div>
      </div>
    );
  };

  return (
    <div className="flex h-screen bg-gray-900">
      {/* Game Canvas */}
      <div className="flex-1 flex items-center justify-center p-4">
        <div className="relative">
          <canvas
            ref={canvasRef}
            width={width}
            height={height}
            className="border border-gray-600 rounded-lg bg-gray-700"
            style={{ maxWidth: '100%', maxHeight: '100%' }}
          />
          {!isInitialized && (
            <div className="absolute inset-0 flex items-center justify-center bg-gray-900 bg-opacity-75 rounded-lg">
              <div className="text-white text-xl">Loading Tactical Arena...</div>
            </div>
          )}
        </div>
      </div>

      {/* UI Panels */}
      <div className="w-80 p-4 space-y-4 overflow-y-auto">
        {renderGameStats()}
        {renderTurnOrder()}
        {renderUnitDetails()}
        {renderUnitActions()}
        {renderSettings()}
      </div>

      {/* Game Over Overlay */}
      {renderGameOver()}
    </div>
  );
};