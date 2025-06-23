package cqrsx_test

import (
	"context"
	"cqrs/cqrsx/v2"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 테스트용 오류 정의
var (
	ErrTemporary = errors.New("temporary error")
	ErrPermanent = errors.New("permanent error")
)

// 도메인 특화 재시도 정책 (예시)
func isRetryable(err error) bool {
	return !errors.Is(err, ErrPermanent)
}

func TestRetryer_Do(t *testing.T) {
	t.Run("성공 케이스", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer()

		// 재시도 가능한 함수 정의
		count := 0
		fn := func() error {
			count++
			if count < 3 {
				return ErrTemporary
			}
			return nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		err := retryer.Do(ctx, fn)

		// 검증
		assert.NoError(t, err, "재시도 후 성공해야 함")
		assert.Equal(t, 3, count, "정확히 3번 시도해야 함")
	})

	t.Run("최대 시도 횟수 초과", func(t *testing.T) {
		// 최대 시도 횟수가 3인 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(3),
			cqrsx.WithConstantBackoff(10*time.Millisecond),
		)

		// 항상 실패하는 함수
		count := 0
		fn := func() error {
			count++
			return ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		err := retryer.Do(ctx, fn)

		// 검증
		require.Error(t, err, "오류가 반환되어야 함")
		assert.Equal(t, ErrTemporary, err, "원본 오류가 반환되어야 함")
		assert.Equal(t, 3, count, "정확히 3번 시도해야 함")
	})

	t.Run("최대 시간 초과", func(t *testing.T) {
		// 최대 시간이 100ms인 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxDuration(100*time.Millisecond),
			cqrsx.WithConstantBackoff(50*time.Millisecond),
		)

		// 항상 실패하는 함수
		count := 0
		fn := func() error {
			count++
			return ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		start := time.Now()
		err := retryer.Do(ctx, fn)
		duration := time.Since(start)

		// 검증
		require.Error(t, err, "오류가 반환되어야 함")
		assert.Equal(t, ErrTemporary, err, "원본 오류가 반환되어야 함")
		assert.GreaterOrEqual(t, duration, 100*time.Millisecond, "최소 100ms 이상 실행되어야 함")
		assert.LessOrEqual(t, count, 3, "최대 3번까지만 시도해야 함")
	})

	t.Run("컨텍스트 취소", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithConstantBackoff(100 * time.Millisecond),
		)

		// 항상 실패하는 함수
		count := 0
		fn := func() error {
			count++
			return ErrTemporary
		}

		// 취소 가능한 컨텍스트 생성
		ctx, cancel := context.WithCancel(context.Background())

		// 첫 번째 시도 후 취소
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		// 함수 실행 및 재시도
		err := retryer.Do(ctx, fn)

		// 검증
		require.Error(t, err, "오류가 반환되어야 함")
		assert.ErrorIs(t, err, context.Canceled, "컨텍스트 취소 오류여야 함")
		assert.LessOrEqual(t, count, 2, "최대 2번까지만 시도해야 함")
	})

	t.Run("재시도 조건에 따른 중단", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(5),
			cqrsx.WithConstantBackoff(10*time.Millisecond),
		)

		// 첫 번째 시도에서 영구적 오류 반환
		count := 0
		fn := func() error {
			count++
			if count == 1 {
				return ErrPermanent // 재시도하지 않아야 하는 오류
			}
			return nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도 (도메인 정책 사용)
		err := retryer.Do(ctx, fn, cqrsx.WithRetryPredicate(isRetryable))

		// 검증
		require.Error(t, err, "오류가 반환되어야 함")
		assert.Equal(t, ErrPermanent, err, "원본 오류가 그대로 반환되어야 함")
		assert.Equal(t, 1, count, "정확히 1번만 시도해야 함")
	})
}

func TestRetryer_DoWithResult(t *testing.T) {
	t.Run("성공 케이스", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer()

		// 재시도 가능한 함수 정의
		count := 0
		fn := func() error {
			count++
			if count < 3 {
				return ErrTemporary
			}
			return nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result := retryer.DoWithResult(ctx, fn)

		// 검증
		assert.True(t, result.Success, "성공 플래그가 true여야 함")
		assert.NoError(t, result.Error, "오류가 없어야 함")
		assert.Equal(t, 3, result.Attempts, "정확히 3번 시도해야 함")
		assert.NotZero(t, result.Duration, "소요 시간이 기록되어야 함")
		assert.Empty(t, result.FailureReason, "실패 이유가 없어야 함")
	})

	t.Run("최대 시도 횟수 초과", func(t *testing.T) {
		// 최대 시도 횟수가 3인 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(3),
			cqrsx.WithConstantBackoff(10*time.Millisecond),
		)

		// 항상 실패하는 함수
		count := 0
		fn := func() error {
			count++
			return ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result := retryer.DoWithResult(ctx, fn)

		// 검증
		assert.False(t, result.Success, "성공 플래그가 false여야 함")
		assert.Equal(t, ErrTemporary, result.Error, "원본 오류가 반환되어야 함")
		assert.Equal(t, 3, result.Attempts, "정확히 3번 시도해야 함")
		assert.NotZero(t, result.Duration, "소요 시간이 기록되어야 함")
		assert.Contains(t, result.FailureReason, "maximum retry attempts", "실패 이유가 최대 시도 횟수 초과여야 함")
	})
}

func TestRetryerAdapter_DoWithValue(t *testing.T) {
	t.Run("성공 케이스", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer()

		// 어댑터 생성
		adapter := cqrsx.NewRetryerAdapter[string](retryer)

		// 재시도 가능한 함수 정의
		count := 0
		fn := func() (string, error) {
			count++
			if count < 3 {
				return "", ErrTemporary
			}
			return "success", nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result, value := adapter.DoWithValue(ctx, fn)

		// 검증
		assert.True(t, result.Success, "성공 플래그가 true여야 함")
		assert.NoError(t, result.Error, "오류가 없어야 함")
		assert.Equal(t, 3, result.Attempts, "정확히 3번 시도해야 함")
		assert.Equal(t, "success", value, "올바른 값이 반환되어야 함")
	})

	t.Run("실패 케이스", func(t *testing.T) {
		// 최대 시도 횟수가 3인 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(3),
			cqrsx.WithConstantBackoff(10*time.Millisecond),
		)

		// 어댑터 생성
		adapter := cqrsx.NewRetryerAdapter[string](retryer)

		// 항상 실패하는 함수
		count := 0
		fn := func() (string, error) {
			count++
			return "", ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result, value := adapter.DoWithValue(ctx, fn)

		// 검증
		assert.False(t, result.Success, "성공 플래그가 false여야 함")
		assert.Equal(t, ErrTemporary, result.Error, "원본 오류가 반환되어야 함")
		assert.Equal(t, 3, result.Attempts, "정확히 3번 시도해야 함")
		assert.Equal(t, "", value, "기본값이 반환되어야 함")
	})

	t.Run("다양한 타입 테스트 - 정수", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer()

		// 정수 타입 어댑터 생성
		adapter := cqrsx.NewRetryerAdapter[int](retryer)

		// 재시도 가능한 함수 정의
		count := 0
		fn := func() (int, error) {
			count++
			if count < 2 {
				return 0, ErrTemporary
			}
			return 42, nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result, value := adapter.DoWithValue(ctx, fn)

		// 검증
		assert.True(t, result.Success, "성공 플래그가 true여야 함")
		assert.Equal(t, 2, result.Attempts, "정확히 2번 시도해야 함")
		assert.Equal(t, 42, value, "올바른 값이 반환되어야 함")
	})

	t.Run("다양한 타입 테스트 - 구조체", func(t *testing.T) {
		// 테스트용 구조체 정의
		type Person struct {
			Name string
			Age  int
		}

		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer()

		// 구조체 타입 어댑터 생성
		adapter := cqrsx.NewRetryerAdapter[Person](retryer)

		// 재시도 가능한 함수 정의
		count := 0
		fn := func() (Person, error) {
			count++
			if count < 2 {
				return Person{}, ErrTemporary
			}
			return Person{Name: "John", Age: 30}, nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result, value := adapter.DoWithValue(ctx, fn)

		// 검증
		assert.True(t, result.Success, "성공 플래그가 true여야 함")
		assert.Equal(t, 2, result.Attempts, "정확히 2번 시도해야 함")
		assert.Equal(t, "John", value.Name, "올바른 이름이 반환되어야 함")
		assert.Equal(t, 30, value.Age, "올바른 나이가 반환되어야 함")
	})
}

func TestRetryOptions(t *testing.T) {
	t.Run("옵션 조합", func(t *testing.T) {
		// 여러 옵션을 조합한 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(5),
			cqrsx.WithMaxDuration(200*time.Millisecond),
			cqrsx.WithBackoff(10*time.Millisecond, 100*time.Millisecond, 2.0),
			cqrsx.WithJitter(0.2),
		)

		// 재시도 가능한 함수 정의
		count := 0
		fn := func() error {
			count++
			if count < 3 {
				return ErrTemporary
			}
			return nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result := retryer.DoWithResult(ctx, fn)

		// 검증
		assert.True(t, result.Success, "성공 플래그가 true여야 함")
		assert.Equal(t, 3, result.Attempts, "정확히 3번 시도해야 함")
	})

	t.Run("실행 시 옵션 오버라이드", func(t *testing.T) {
		// 기본 설정으로 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(10),
			cqrsx.WithConstantBackoff(100*time.Millisecond),
		)

		// 항상 실패하는 함수
		count := 0
		fn := func() error {
			count++
			return ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 실행 시 옵션 오버라이드
		result := retryer.DoWithResult(ctx, fn,
			cqrsx.WithMaxAttempts(3),                       // 10에서 3으로 오버라이드
			cqrsx.WithConstantBackoff(10*time.Millisecond), // 100ms에서 10ms로 오버라이드
		)

		// 검증
		assert.False(t, result.Success, "성공 플래그가 false여야 함")
		assert.Equal(t, 3, result.Attempts, "정확히 3번 시도해야 함 (오버라이드된 값)")
	})
}

func TestRetryPredicate(t *testing.T) {
	t.Run("특정 오류만 재시도", func(t *testing.T) {
		// 기본 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(5),
			cqrsx.WithConstantBackoff(10*time.Millisecond),
			cqrsx.WithRetryPredicate(func(err error) bool {
				// ErrTemporary만 재시도
				return errors.Is(err, ErrTemporary)
			}),
		)

		// 다양한 오류를 반환하는 함수
		count := 0
		fn := func() error {
			count++
			if count == 1 {
				return ErrTemporary // 재시도 가능
			}
			if count == 2 {
				return ErrPermanent // 재시도 불가능
			}
			return nil
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		result := retryer.DoWithResult(ctx, fn)

		// 검증
		assert.False(t, result.Success, "성공 플래그가 false여야 함")
		assert.Equal(t, ErrPermanent, result.Error, "ErrPermanent가 반환되어야 함")
		assert.Equal(t, 2, result.Attempts, "정확히 2번 시도해야 함")
	})
}

func TestBackoffStrategies(t *testing.T) {
	t.Run("지수 백오프", func(t *testing.T) {
		// 지수 백오프 설정으로 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(4),
			cqrsx.WithBackoff(10*time.Millisecond, 1*time.Second, 2.0),
			cqrsx.WithJitter(0), // 지터 없음 (테스트 예측 가능성)
		)

		// 항상 실패하는 함수
		startTimes := make([]time.Time, 0, 4)
		fn := func() error {
			startTimes = append(startTimes, time.Now())
			return ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		retryer.DoWithResult(ctx, fn)

		// 검증 (대략적인 간격 확인)
		require.Equal(t, 4, len(startTimes), "정확히 4번 시도해야 함")

		// 첫 번째 시도는 즉시 실행
		// 두 번째 시도는 약 10ms 후
		// 세 번째 시도는 약 20ms 후 (10ms * 2)
		// 네 번째 시도는 약 40ms 후 (20ms * 2)
		if len(startTimes) >= 3 {
			interval1 := startTimes[1].Sub(startTimes[0])
			interval2 := startTimes[2].Sub(startTimes[1])

			// 두 번째 간격이 첫 번째 간격보다 커야 함 (지수 증가)
			assert.Greater(t, interval2, interval1, "지수 백오프: 두 번째 간격이 첫 번째 간격보다 커야 함")
		}
	})

	t.Run("일정 백오프", func(t *testing.T) {
		// 일정 백오프 설정으로 Retryer 생성
		retryer := cqrsx.NewRetryer(
			cqrsx.WithMaxAttempts(4),
			cqrsx.WithConstantBackoff(20*time.Millisecond),
			cqrsx.WithJitter(0), // 지터 없음 (테스트 예측 가능성)
		)

		// 항상 실패하는 함수
		startTimes := make([]time.Time, 0, 4)
		fn := func() error {
			startTimes = append(startTimes, time.Now())
			return ErrTemporary
		}

		// 컨텍스트 생성
		ctx := context.Background()

		// 함수 실행 및 재시도
		retryer.DoWithResult(ctx, fn)

		// 검증 (대략적인 간격 확인)
		require.Equal(t, 4, len(startTimes), "정확히 4번 시도해야 함")

		if len(startTimes) >= 3 {
			interval1 := startTimes[1].Sub(startTimes[0])
			interval2 := startTimes[2].Sub(startTimes[1])

			// 두 간격이 비슷해야 함 (일정 백오프)
			ratio := interval2.Seconds() / interval1.Seconds()
			assert.InDelta(t, 1.0, ratio, 0.5, "일정 백오프: 간격이 비슷해야 함")
		}
	})
}
