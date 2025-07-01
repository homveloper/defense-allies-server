'use client';

import dynamic from 'next/dynamic';
import { Suspense } from 'react';

const MinimalLegionGame = dynamic(
  () => import('@/components/minimal-legion/MinimalLegionGame'),
  {
    ssr: false,
    loading: () => (
      <div className="w-full h-screen flex items-center justify-center bg-gray-900">
        <div className="text-white text-xl">Loading game...</div>
      </div>
    ),
  }
);

export default function MinimalLegionPage() {
  return (
    <div className="w-full h-screen bg-gray-900">
      <Suspense fallback={<div>Loading...</div>}>
        <MinimalLegionGame />
      </Suspense>
    </div>
  );
}