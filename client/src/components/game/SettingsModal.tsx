'use client'

import React, { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Toggle } from '@/components/ui/Toggle'

interface SettingsModalProps {
  onClose: () => void
}

export default function SettingsModal({ onClose }: SettingsModalProps) {
  const router = useRouter()
  const [soundEffects, setSoundEffects] = useState(true)
  const [backgroundMusic, setBackgroundMusic] = useState(false)
  const [screenVibration, setScreenVibration] = useState(true)

  const handleExitGame = () => {
    router.push('/home')
  }

  return (
    <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-6">
      {/* Modal */}
      <div className="bg-white rounded-2xl w-full max-w-sm overflow-hidden">
        {/* Header */}
        <div className="bg-slate-50 px-6 py-4 flex items-center justify-between border-b border-slate-200">
          <h2 className="text-lg font-semibold text-slate-900">게임 설정</h2>
          <button
            onClick={onClose}
            className="w-6 h-6 rounded-full bg-slate-100 flex items-center justify-center hover:bg-slate-200 transition-colors"
          >
            <span className="text-slate-500 text-lg leading-none">×</span>
          </button>
        </div>

        {/* Content */}
        <div className="px-6 py-4">
          {/* Sound Settings */}
          <div className="mb-6">
            <h3 className="text-base font-medium text-slate-900 mb-4">사운드</h3>
            
            {/* Sound Effects */}
            <div className="flex items-center justify-between mb-3">
              <span className="text-sm text-slate-500">효과음</span>
              <Toggle
                checked={soundEffects}
                onChange={setSoundEffects}
              />
            </div>

            {/* Background Music */}
            <div className="flex items-center justify-between">
              <span className="text-sm text-slate-500">배경음악</span>
              <Toggle
                checked={backgroundMusic}
                onChange={setBackgroundMusic}
              />
            </div>
          </div>

          <div className="h-px bg-slate-200 mb-6" />

          {/* Game Settings */}
          <div className="mb-6">
            <h3 className="text-base font-medium text-slate-900 mb-4">게임</h3>
            
            {/* Screen Vibration */}
            <div className="flex items-center justify-between mb-4">
              <span className="text-sm text-slate-500">화면 진동</span>
              <Toggle
                checked={screenVibration}
                onChange={setScreenVibration}
              />
            </div>

            {/* Game Speed Info */}
            <div>
              <p className="text-sm text-slate-400">게임 속도: 1x</p>
              <p className="text-xs text-slate-400">매치 시작 시 결정됨</p>
            </div>
          </div>

          <div className="h-px bg-slate-200 mb-4" />

          {/* Action Buttons */}
          <div className="space-y-3">
            {/* Exit Game */}
            <button
              onClick={handleExitGame}
              className="w-full py-2.5 rounded-lg border border-red-200 bg-red-50 text-red-600 font-medium text-sm hover:bg-red-100 transition-colors"
            >
              게임 나가기
            </button>

            {/* Continue Game */}
            <button
              onClick={onClose}
              className="w-full py-2.5 rounded-lg bg-blue-600 text-white font-medium text-sm hover:bg-blue-700 transition-colors"
            >
              게임 계속하기
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}