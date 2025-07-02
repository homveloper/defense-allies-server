'use client';

import { useRouter } from 'next/navigation';
import { Card, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';

export default function GamesPage() {
  const router = useRouter();

  const games = [
    {
      id: 'minimal-legion',
      title: '미니멀 군단',
      description: '혼자서 시작해 거대한 군단을 만들어가는 액션 로그라이크 게임',
      status: 'available',
      difficulty: '쉬움',
      players: '1인',
      image: '/images/minimal-legion-thumb.png',
      route: '/minimal-legion',
    },
    {
      id: 'ability-arena',
      title: '🏟️ Ability Arena',
      description: 'GAS 어빌리티 시스템을 테스트할 수 있는 아레나 배틀 게임',
      status: 'available',
      difficulty: '보통',
      players: '1인',
      image: '/images/ability-arena-thumb.png',
      route: '/ability-arena',
      tag: 'NEW'
    },
    {
      id: 'tower-defense',
      title: '타워 디펜스',
      description: '전략적인 타워 배치로 적의 침입을 막는 디펜스 게임',
      status: 'coming-soon',
      difficulty: '보통',
      players: '1-4인',
      image: '/images/tower-defense-thumb.png',
      route: '/tower-defense',
    },
    {
      id: 'space-survival',
      title: '우주 생존',
      description: '우주 공간에서 살아남기 위한 서바이벌 게임',
      status: 'coming-soon',
      difficulty: '어려움',
      players: '1-2인',
      image: '/images/space-survival-thumb.png',
      route: '/space-survival',
    },
  ];

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <header className="px-6 py-4 border-b border-gray-800">
        <h1 className="text-2xl font-semibold">게임 선택</h1>
        <p className="text-sm text-gray-400 mt-1">플레이할 게임을 선택하세요</p>
      </header>

      <main className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {games.map((game) => (
            <Card
              key={game.id}
              variant={game.status === 'available' ? 'interactive' : 'default'}
              className={`bg-gray-800 border-gray-700 ${
                game.status !== 'available' ? 'opacity-60' : ''
              }`}
            >
              <CardHeader className="space-y-4">
                <div className="aspect-video bg-gray-700 rounded-lg flex items-center justify-center">
                  <span className="text-gray-500 text-sm">게임 이미지</span>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-start justify-between">
                    <CardTitle className="text-white">{game.title}</CardTitle>
                    <div className="flex gap-1">
                      {game.tag === 'NEW' && (
                        <Badge variant="success">NEW</Badge>
                      )}
                      {game.status === 'coming-soon' && (
                        <Badge variant="waiting">준비중</Badge>
                      )}
                    </div>
                  </div>
                  
                  <CardDescription className="text-gray-400">
                    {game.description}
                  </CardDescription>
                  
                  <div className="flex gap-2 text-xs">
                    <Badge variant="level">{game.difficulty}</Badge>
                    <Badge variant="rank">{game.players}</Badge>
                  </div>
                </div>

                <Button
                  fullWidth
                  disabled={game.status !== 'available'}
                  onClick={() => {
                    if (game.status === 'available') {
                      router.push(game.route);
                    }
                  }}
                >
                  {game.status === 'available' ? '게임 시작' : '준비중'}
                </Button>
              </CardHeader>
            </Card>
          ))}
        </div>

        <div className="mt-8 text-center">
          <Button
            variant="ghost"
            onClick={() => router.back()}
            className="text-gray-400 hover:text-white"
          >
            ← 돌아가기
          </Button>
        </div>
      </main>
    </div>
  );
}