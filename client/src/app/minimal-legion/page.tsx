'use client';

import { useEffect, useRef } from 'react';
import MinimalLegionGameV2 from '@/game/minimal-legion/presentation/components/MinimalLegionGameV2';

export default function MinimalLegionPage() {
  return (
    <div className="min-h-screen bg-gray-900 flex items-center justify-center">
      <div className="relative">
        <MinimalLegionGameV2 />
      </div>
    </div>
  );
}