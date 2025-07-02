'use client'

import React, { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/Button'
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card'
import { Input, SearchInput } from '@/components/ui/Input'
import { Toggle } from '@/components/ui/Toggle'
import { Checkbox } from '@/components/ui/Checkbox'
import { Badge } from '@/components/ui/Badge'
import { Progress } from '@/components/ui/Progress'
import { Toast } from '@/components/ui/Toast'

export default function ShowcasePage() {
  const router = useRouter()
  const [toggleStates, setToggleStates] = useState({
    sound: true,
    vibration: false,
  })
  
  const [checkboxStates, setCheckboxStates] = useState({
    terms: true,
    marketing: false,
  })

  const [searchValue, setSearchValue] = useState('')

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Header */}
      <header className="bg-white px-6 py-4 border-b border-slate-200">
        <h1 className="text-xl font-semibold text-slate-900">UI Component Showcase</h1>
        <p className="text-sm text-slate-500">Defense Allies 디자인 시스템</p>
      </header>

      {/* Content */}
      <main className="px-6 py-6 space-y-8">
        {/* Game Start Section */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Game Selection</h2>
          <div className="space-y-3">
            <Card variant="interactive" onClick={() => router.push('/minimal-legion')}>
              <CardHeader>
                <CardTitle>미니멀 군단 (Minimal Legion)</CardTitle>
                <CardDescription>
                  혼자서 시작해 거대한 군단을 만들어가는 액션 로그라이크 게임
                </CardDescription>
              </CardHeader>
            </Card>
            
            <Card variant="interactive" onClick={() => router.push('/ability-arena')}>
              <CardHeader>
                <div className="flex items-center gap-2">
                  <CardTitle>🏟️ Ability Arena</CardTitle>
                  <Badge variant="success">NEW</Badge>
                </div>
                <CardDescription>
                  GAS 어빌리티 시스템을 테스트할 수 있는 아레나 배틀 게임
                </CardDescription>
              </CardHeader>
            </Card>
          </div>
        </section>

        {/* Color Palette */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Color Palette</h2>
          <div className="space-y-3">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Primary Colors</p>
              <div className="flex gap-4">
                <div className="text-center">
                  <div className="w-10 h-10 bg-blue-600 rounded-lg mb-1"></div>
                  <p className="text-[10px] text-slate-500">Main</p>
                  <p className="text-[10px] text-slate-500">#2563EB</p>
                </div>
                <div className="text-center">
                  <div className="w-10 h-10 bg-green-500 rounded-lg mb-1"></div>
                  <p className="text-[10px] text-slate-500">Secondary</p>
                  <p className="text-[10px] text-slate-500">#10B981</p>
                </div>
                <div className="text-center">
                  <div className="w-10 h-10 bg-amber-500 rounded-lg mb-1"></div>
                  <p className="text-[10px] text-slate-500">Accent</p>
                  <p className="text-[10px] text-slate-500">#F59E0B</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Typography */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Typography</h2>
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold text-slate-900">Heading 1 - 24px/32px</h1>
            <h2 className="text-xl font-semibold text-slate-900">Heading 2 - 20px/28px</h2>
            <h3 className="text-lg font-semibold text-slate-900">Heading 3 - 18px/24px</h3>
            <p className="text-base text-slate-900">Body Large - 16px/24px</p>
            <p className="text-sm text-slate-900">Body Regular - 14px/20px</p>
            <p className="text-xs text-slate-500">Body Small - 12px/16px</p>
          </div>
        </section>

        {/* Button Components */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Button Components</h2>
          
          <div className="space-y-4">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Primary Buttons</p>
              <div className="space-y-3">
                <Button size="large" fullWidth>Large Primary Button</Button>
                <div className="flex gap-3">
                  <Button size="medium">Medium Primary</Button>
                  <Button size="small">Small Primary</Button>
                </div>
              </div>
            </div>

            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Button Variants</p>
              <div className="space-y-3">
                <Button variant="secondary" fullWidth>Secondary Button</Button>
                <div className="flex gap-3">
                  <Button variant="danger">Danger Button</Button>
                  <Button variant="ghost">Ghost Button</Button>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Card Components */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Card Components</h2>
          
          <div className="space-y-3">
            <Card>
              <CardHeader>
                <CardTitle>Card Title</CardTitle>
                <CardDescription>
                  Card description with some additional information that spans multiple lines.
                </CardDescription>
              </CardHeader>
            </Card>

            <div className="grid grid-cols-2 gap-3">
              <Card variant="interactive">
                <div className="flex items-start gap-3">
                  <div className="w-6 h-6 bg-blue-600 rounded-full"></div>
                  <div>
                    <CardTitle>Interactive</CardTitle>
                    <CardDescription>Clickable card with hover effects</CardDescription>
                  </div>
                </div>
              </Card>

              <Card variant="status" statusColor="green">
                <CardTitle>Success Status</CardTitle>
                <CardDescription>Operation completed successfully</CardDescription>
              </Card>
            </div>
          </div>
        </section>

        {/* Input Components */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Input Components</h2>
          
          <div className="space-y-3">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Text Input</p>
              <Input placeholder="플레이어 닉네임 입력" />
            </div>

            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Search Input</p>
              <SearchInput 
                placeholder="친구 검색..." 
                value={searchValue}
                onSearch={setSearchValue}
              />
            </div>
          </div>
        </section>

        {/* Toggle & Checkbox */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Toggle & Checkbox</h2>
          
          <div className="space-y-3">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Toggle Switch</p>
              <div className="space-y-2">
                <Toggle 
                  label="사운드 효과"
                  checked={toggleStates.sound}
                  onChange={(checked) => setToggleStates({...toggleStates, sound: checked})}
                />
                <Toggle 
                  label="진동 피드백"
                  checked={toggleStates.vibration}
                  onChange={(checked) => setToggleStates({...toggleStates, vibration: checked})}
                />
              </div>
            </div>

            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Checkbox</p>
              <div className="space-y-2">
                <Checkbox 
                  label="약관 동의"
                  checked={checkboxStates.terms}
                  onChange={(checked) => setCheckboxStates({...checkboxStates, terms: checked})}
                />
                <Checkbox 
                  label="마케팅 수신 동의"
                  checked={checkboxStates.marketing}
                  onChange={(checked) => setCheckboxStates({...checkboxStates, marketing: checked})}
                />
              </div>
            </div>
          </div>
        </section>

        {/* Status Indicators */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Status Indicators</h2>
          
          <div className="space-y-3">
            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Progress Bar</p>
              <Progress 
                value={1250} 
                max={2000} 
                variant="experience" 
                label="경험치"
                showValue
              />
            </div>

            <div>
              <p className="text-xs font-medium text-slate-500 uppercase mb-2">Health Bar</p>
              <Progress 
                value={75} 
                max={100} 
                variant="health" 
                label="체력"
                showValue
              />
            </div>
          </div>
        </section>

        {/* Badges & Labels */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Badges & Labels</h2>
          
          <div className="flex flex-wrap gap-2">
            <Badge variant="online">온라인</Badge>
            <Badge variant="waiting">대기중</Badge>
            <Badge variant="offline">오프라인</Badge>
            <Badge variant="level">Lv.15</Badge>
            <Badge variant="rank">골드</Badge>
          </div>
        </section>

        {/* Notifications */}
        <section>
          <h2 className="text-base font-semibold text-slate-900 mb-3">Notifications</h2>
          
          <div className="space-y-3">
            <Toast
              variant="success"
              title="게임 승리!"
              description="경험치 +150, 골드 +300을 획득했습니다."
            />
            
            <Toast
              variant="error"
              title="연결 오류"
              description="서버와의 연결이 끊어졌습니다. 다시 시도해주세요."
            />
          </div>
        </section>
      </main>
    </div>
  )
}