'use client';

import { Button } from '@/components/ui/Button';
import { Card, CardHeader, CardTitle } from '@/components/ui/Card';
import { useMinimalLegionStore } from '@/store/minimalLegionStore';

interface GameOverModalProps {
  onRestart: () => void;
  onExit: () => void;
}

export const GameOverModal = ({ onRestart, onExit }: GameOverModalProps) => {
  const { isGameOver, score, wave, allies } = useMinimalLegionStore();

  if (!isGameOver) return null;

  return (
    <div className="absolute inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center pointer-events-auto z-50">
      <Card className="w-96 bg-white border-red-400 shadow-xl">
        <CardHeader className="text-center">
          <div className="mb-4">
            <div className="text-6xl mb-2">💀</div>
            <CardTitle className="text-red-600 text-2xl mb-2">게임 오버</CardTitle>
            <p className="text-gray-600 text-sm">플레이어가 쓰러졌습니다</p>
          </div>

          {/* 게임 결과 */}
          <div className="bg-gray-50 rounded-lg p-4 mb-6 border border-gray-200">
            <h3 className="text-gray-800 font-semibold mb-3">게임 결과</h3>
            <div className="grid grid-cols-2 gap-3 text-sm">
              <div className="bg-white p-2 rounded border border-gray-200">
                <div className="text-gray-500">최종 점수</div>
                <div className="text-yellow-600 font-bold text-lg">{score.toLocaleString()}</div>
              </div>
              <div className="bg-white p-2 rounded border border-gray-200">
                <div className="text-gray-500">도달 웨이브</div>
                <div className="text-blue-600 font-bold text-lg">{wave}</div>
              </div>
              <div className="bg-white p-2 rounded border border-gray-200">
                <div className="text-gray-500">최대 군단</div>
                <div className="text-green-600 font-bold text-lg">{allies}</div>
              </div>
              <div className="bg-white p-2 rounded border border-gray-200">
                <div className="text-gray-500">생존 시간</div>
                <div className="text-purple-600 font-bold text-lg">
                  {Math.floor(wave * 30 / 60)}:{String(wave * 30 % 60).padStart(2, '0')}
                </div>
              </div>
            </div>
          </div>

          {/* 버튼들 */}
          <div className="space-y-3">
            <Button
              variant="primary"
              fullWidth
              size="large"
              onClick={onRestart}
              className="bg-blue-600 hover:bg-blue-700 text-white"
            >
              🔄 다시 도전
            </Button>
            
            <Button
              variant="secondary"
              fullWidth
              onClick={onExit}
              className="bg-gray-500 hover:bg-gray-600 text-white"
            >
              🏠 게임 목록으로
            </Button>
          </div>

          {/* 팁 */}
          <div className="mt-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
            <h4 className="text-blue-600 font-semibold text-sm mb-1">💡 팁</h4>
            <p className="text-gray-700 text-xs">
              적들을 처치해서 아군으로 만들고, 레벨업으로 더 강해지세요!
            </p>
          </div>
        </CardHeader>
      </Card>
    </div>
  );
};