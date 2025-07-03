'use client';

import React from 'react';
import { TacticalArenaGame } from '../../components/tactical-arena/TacticalArenaGame';

export default function TacticalArenaPage() {
  return (
    <div className="min-h-screen bg-gray-900">
      <div className="container mx-auto py-4">
        <header className="text-center mb-6">
          <h1 className="text-4xl font-bold text-white mb-2">
            Tactical Arena Demo
          </h1>
          <p className="text-gray-300 text-lg">
            GAS v2 Turn-Based Combat System Showcase
          </p>
          <p className="text-gray-400 text-sm mt-2">
            XCOM/Fire Emblem style tactical combat with N:M unit support
          </p>
        </header>

        <div className="mb-4">
          <div className="bg-gray-800 rounded-lg p-4 text-white">
            <h2 className="text-xl font-bold mb-3">How to Play</h2>
            <div className="grid md:grid-cols-2 gap-4 text-sm">
              <div>
                <h3 className="font-semibold text-blue-400 mb-2">Basic Controls</h3>
                <ul className="space-y-1">
                  <li>• Click units to select them</li>
                  <li>• Click tiles to move selected unit</li>
                  <li>• Use action buttons in the side panel</li>
                  <li>• End phases and turns with control buttons</li>
                </ul>
              </div>
              <div>
                <h3 className="font-semibold text-green-400 mb-2">Combat Mechanics</h3>
                <ul className="space-y-1">
                  <li>• Phase-based turns (Movement → Action → Bonus)</li>
                  <li>• Resource management (Action/Movement points)</li>
                  <li>• Cover system for tactical positioning</li>
                  <li>• Initiative-based turn order</li>
                </ul>
              </div>
            </div>
          </div>
        </div>

        <TacticalArenaGame 
          width={800}
          height={600}
          playerUnits={2}
          enemyUnits={2}
          mapSize={{ width: 8, height: 6 }}
        />

        <div className="mt-6 bg-gray-800 rounded-lg p-4 text-white">
          <h2 className="text-xl font-bold mb-3">Features Demonstrated</h2>
          <div className="grid md:grid-cols-3 gap-4 text-sm">
            <div>
              <h3 className="font-semibold text-purple-400 mb-2">GAS v2 Integration</h3>
              <ul className="space-y-1">
                <li>• Turn-Based Resource System</li>
                <li>• Phase-Based Execution</li>
                <li>• Initiative & Turn Order</li>
                <li>• Enhanced Ability Context</li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-yellow-400 mb-2">Tactical Features</h3>
              <ul className="space-y-1">
                <li>• Grid-based movement</li>
                <li>• Line-of-sight system</li>
                <li>• Cover mechanics</li>
                <li>• Pathfinding & validation</li>
              </ul>
            </div>
            <div>
              <h3 className="font-semibold text-red-400 mb-2">Combat Abilities</h3>
              <ul className="space-y-1">
                <li>• Move (Movement phase)</li>
                <li>• Attack (Main action)</li>
                <li>• Aimed Shot (High accuracy)</li>
                <li>• Overwatch (Reaction system)</li>
              </ul>
            </div>
          </div>
        </div>

        <div className="mt-4 text-center">
          <p className="text-gray-400 text-sm">
            This demo showcases the complete integration of GAS v2 turn-based systems with tactical combat mechanics.
            The system supports scalable N:M unit combat with full resource management and phase-based execution.
          </p>
        </div>
      </div>
    </div>
  );
}