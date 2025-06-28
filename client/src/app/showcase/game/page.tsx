'use client'

import React from 'react'
import { Card, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Progress } from '@/components/ui/Progress'

// Game-specific components
function TowerCard({ name, cost, level = 1 }: { name: string; cost: number; level?: number }) {
  return (
    <Card variant="interactive" className="w-[100px]">
      <CardContent className="p-3 text-center">
        <div className="w-[60px] h-10 bg-blue-600 rounded mx-auto mb-2 relative">
          <div className="absolute inset-x-2 inset-y-2 bg-white rounded-sm"></div>
          {level > 1 && (
            <div className="absolute -top-2 -right-2 w-4 h-4 bg-amber-500 rounded-full flex items-center justify-center">
              <span className="text-[8px] font-bold text-white">{level}</span>
            </div>
          )}
        </div>
        <p className="text-xs font-medium text-slate-900">{name}</p>
        <p className="text-[10px] text-slate-500">{cost} 골드</p>
      </CardContent>
    </Card>
  )
}

function ResourceDisplay() {
  return (
    <Card className="w-full">
      <CardContent className="p-3 flex items-center justify-around">
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-amber-500 rounded-full flex items-center justify-center">
            <span className="text-[10px] font-bold text-white">$</span>
          </div>
          <span className="text-sm font-medium text-slate-900">1,250</span>
        </div>
        
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-red-500 rounded-full flex items-center justify-center">
            <span className="text-[10px] font-bold text-white">♥</span>
          </div>
          <span className="text-sm font-medium text-slate-900">85/100</span>
        </div>
        
        <div className="flex items-center gap-2">
          <div className="w-6 h-6 bg-green-500 rounded-full flex items-center justify-center">
            <span className="text-[10px] font-bold text-white">~</span>
          </div>
          <span className="text-sm font-medium text-slate-900">7</span>
        </div>
      </CardContent>
    </Card>
  )
}

function WaveProgress() {
  return (
    <Card className="w-full">
      <CardContent className="p-4">
        <div className="flex justify-between items-center mb-2">
          <h3 className="text-xs font-medium text-slate-900">웨이브 7/15</h3>
          <Badge variant="waiting" className="text-[10px]">진행중</Badge>
        </div>
        <Progress value={7} max={15} className="mb-3" />
        <div className="space-y-1">
          <p className="text-[10px] text-slate-500">적 15/30 처치</p>
          <p className="text-[10px] text-slate-500">다음 웨이브까지 15초</p>
        </div>
      </CardContent>
    </Card>
  )
}

function FriendListItem({ name, level, isOnline }: { name: string; level: number; isOnline: boolean }) {
  return (
    <Card className="w-full">
      <CardContent className="p-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="relative">
            <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center">
              <span className="text-xs font-bold text-white">{name[0]}</span>
            </div>
            {isOnline && (
              <div className="absolute -bottom-0.5 -right-0.5 w-2 h-2 bg-green-500 rounded-full border border-white"></div>
            )}
          </div>
          <div>
            <p className="text-sm font-medium text-slate-900">{name}</p>
            <p className="text-xs text-slate-500">{isOnline ? '온라인' : '오프라인'} • 레벨 {level}</p>
          </div>
        </div>
        <Button size="small" className="text-[10px] py-1 px-3 h-6">초대</Button>
      </CardContent>
    </Card>
  )
}

function GameHistoryItem() {
  return (
    <Card className="w-full">
      <CardContent className="p-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-6 h-6 bg-green-500 rounded flex items-center justify-center">
            <span className="text-xs font-bold text-white">W</span>
          </div>
          <div>
            <p className="text-sm font-medium text-slate-900">승리 - 웨이브 15</p>
            <p className="text-xs text-slate-500">2시간 전 • +150 XP</p>
          </div>
        </div>
        <div className="text-right">
          <p className="text-xs font-medium text-green-500">+300</p>
          <p className="text-[10px] text-slate-500">골드</p>
        </div>
      </CardContent>
    </Card>
  )
}

export default function GameShowcasePage() {
  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white px-6 py-4 border-b border-slate-200">
        <h1 className="text-xl font-semibold text-slate-900">Game Components</h1>
        <p className="text-sm text-slate-500">Defense Allies 게임 전용 컴포넌트</p>
      </header>

      {/* Content */}
      <main className="px-6 py-6 space-y-8">
        {/* Tower Selection */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Tower Selection</h2>
          <div className="flex gap-3 overflow-x-auto pb-2">
            <TowerCard name="기본 타워" cost={100} />
            <TowerCard name="스플래시 타워" cost={200} level={2} />
            <TowerCard name="슬로우 타워" cost={150} />
            <TowerCard name="레이저 타워" cost={300} level={3} />
          </div>
        </section>

        {/* Resource Display */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Resource Display</h2>
          <ResourceDisplay />
        </section>

        {/* Wave Progress */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Wave Progress</h2>
          <WaveProgress />
        </section>

        {/* Friend List */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Friend List</h2>
          <div className="space-y-2">
            <FriendListItem name="Alice_Player" level={23} isOnline={true} />
            <FriendListItem name="Bob_Gamer" level={19} isOnline={false} />
            <FriendListItem name="Charlie_Pro" level={31} isOnline={true} />
          </div>
        </section>

        {/* Game History */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Game History</h2>
          <div className="space-y-2">
            <GameHistoryItem />
            <Card className="w-full">
              <CardContent className="p-3 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 bg-red-500 rounded flex items-center justify-center">
                    <span className="text-xs font-bold text-white">L</span>
                  </div>
                  <div>
                    <p className="text-sm font-medium text-slate-900">패배 - 웨이브 8</p>
                    <p className="text-xs text-slate-500">5시간 전 • +50 XP</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-xs font-medium text-red-500">+100</p>
                  <p className="text-[10px] text-slate-500">골드</p>
                </div>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* Spacing Guide */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Spacing Guide</h2>
          <p className="text-xs font-medium text-slate-500 uppercase mb-3">Spacing System (4px base unit)</p>
          <div className="flex items-end gap-4">
            <div>
              <div className="w-1 h-5 bg-blue-600 mb-1"></div>
              <p className="text-[10px] text-slate-500">xs: 4px</p>
            </div>
            <div>
              <div className="w-2 h-5 bg-blue-600 mb-1"></div>
              <p className="text-[10px] text-slate-500">sm: 8px</p>
            </div>
            <div>
              <div className="w-4 h-5 bg-blue-600 mb-1"></div>
              <p className="text-[10px] text-slate-500">md: 16px</p>
            </div>
            <div>
              <div className="w-6 h-5 bg-blue-600 mb-1"></div>
              <p className="text-[10px] text-slate-500">lg: 24px</p>
            </div>
            <div>
              <div className="w-8 h-5 bg-blue-600 mb-1"></div>
              <p className="text-[10px] text-slate-500">xl: 32px</p>
            </div>
          </div>
        </section>
      </main>

      {/* Navigation */}
      <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-slate-200 px-6 py-3">
        <div className="flex justify-around">
          <Button variant="ghost" size="small" onClick={() => window.location.href = '/showcase'}>
            기본 컴포넌트
          </Button>
          <Button variant="ghost" size="small" disabled>
            게임 컴포넌트
          </Button>
          <Button variant="ghost" size="small" onClick={() => window.location.href = '/home'}>
            홈으로
          </Button>
        </div>
      </nav>
    </div>
  )
}