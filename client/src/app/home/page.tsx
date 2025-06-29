'use client'

import React from 'react'
import { useRouter } from 'next/navigation'

export default function HomePage() {
  const router = useRouter()
  return (
    <div className="min-h-screen bg-slate-50 flex flex-col">
      {/* Header */}
      <header className="bg-white px-6 pt-14 pb-6 border-b border-slate-200">
        <h1 className="text-2xl font-semibold text-slate-900">Defense Allies</h1>
        <p className="text-sm text-slate-500">함께 막아요</p>
      </header>

      {/* Main Content */}
      <main className="flex-1 px-6 py-4">
        {/* Profile Card */}
        <div className="bg-white rounded-xl border border-slate-200 p-4 flex items-center gap-4 mb-5">
          <div className="w-12 h-12 bg-blue-600 rounded-full flex items-center justify-center">
            <span className="text-white font-semibold">P</span>
          </div>
          <div>
            <h2 className="font-medium text-slate-900">플레이어명</h2>
            <p className="text-xs text-slate-500">레벨 15 • 승률 73%</p>
          </div>
        </div>

        {/* Main Action Button */}
        <button 
          onClick={() => router.push('/games')}
          className="w-full bg-blue-600 text-white font-medium py-4 rounded-full mb-6 hover:bg-blue-700 transition-colors"
        >
          게임 시작
        </button>

        {/* Sub Menu Grid */}
        <div className="grid grid-cols-2 gap-3">
          <button 
            onClick={() => router.push('/showcase')}
            className="bg-white rounded-xl border border-slate-200 p-6 text-left hover:border-slate-300 transition-colors"
          >
            <h3 className="font-medium text-slate-900 mb-1">컴포넌트</h3>
            <p className="text-xs text-slate-500">UI 컴포넌트 확인</p>
          </button>
          
          <button className="bg-white rounded-xl border border-slate-200 p-6 text-left hover:border-slate-300 transition-colors">
            <h3 className="font-medium text-slate-900 mb-1">설정</h3>
            <p className="text-xs text-slate-500">게임 환경 설정</p>
          </button>
        </div>
      </main>

      {/* Bottom Navigation */}
      <nav className="bg-white border-t border-slate-200 px-6 py-3">
        <div className="flex justify-around items-center">
          <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
          <div className="w-2 h-2 bg-slate-200 rounded-full"></div>
          <div className="w-2 h-2 bg-slate-200 rounded-full"></div>
        </div>
      </nav>
    </div>
  )
}