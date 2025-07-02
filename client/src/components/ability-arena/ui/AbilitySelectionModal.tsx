'use client';

import { useAbilityArenaStore } from '@/store/abilityArenaStore';

export function AbilitySelectionModal() {
  const store = useAbilityArenaStore();

  if (!store.isAbilitySelectionOpen) {
    return null;
  }

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
        <h2 className="text-xl font-bold text-center mb-4">Choose an Ability</h2>
        
        <div className="space-y-3">
          {store.availableAbilities.map((ability) => (
            <button
              key={ability.id}
              onClick={() => store.selectAbility(ability.id)}
              className="w-full p-4 text-left border rounded-lg hover:bg-gray-50 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="text-2xl">{ability.icon}</div>
                <div>
                  <h3 className="font-semibold">{ability.name}</h3>
                  <p className="text-sm text-gray-600">{ability.description}</p>
                </div>
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}