package cqrsx

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

var (
	ErrMaxAttemptsReached = errors.New("maximum retry attempts reached")
	ErrMaxDurationReached = errors.New("maximum retry duration reached")
)

// RetryResult는 재시도 작업의 결과를 나타냅니다.
type RetryResult struct {
	Success       bool
	Error         error
	Attempts      int
	Duration      time.Duration
	FailureReason string
}

// RetryableFunc는 재시도 가능한 함수 타입입니다.
type RetryableFunc func() error

// RetryPredicate는 오류가 재시도 가능한지 판단하는 함수 타입입니다.
type RetryPredicate func(error) bool

// Always는 항상 재시도하는 조건입니다.
func Always(err error) bool {
	return true
}

// RetryOption은 재시도 옵션을 설정하는 함수 타입입니다.
type RetryOption func(*RetryConfig)

// RetryConfig는 재시도 설정을 정의합니다.
type RetryConfig struct {
	MaxAttempts     int            // 최대 시도 횟수 (0은 무제한)
	MaxDuration     time.Duration  // 최대 재시도 시간 (0은 무제한)
	InitialInterval time.Duration  // 초기 대기 시간
	MaxInterval     time.Duration  // 최대 대기 시간
	Multiplier      float64        // 대기 시간 증가 배수 (지수 백오프)
	JitterFactor    float64        // 지터 계수 (0.0 ~ 1.0)
	ShouldRetry     RetryPredicate // 재시도 조건
}

// Retryer는 재시도 로직을 제공하는 인터페이스입니다.
type Retryer interface {
	// Do는 지정된 함수를 재시도 정책에 따라 실행합니다.
	Do(ctx context.Context, fn RetryableFunc, opts ...RetryOption) error

	// DoWithResult는 지정된 함수를 재시도 정책에 따라 실행하고 결과를 반환합니다.
	DoWithResult(ctx context.Context, fn RetryableFunc, opts ...RetryOption) *RetryResult
}

// defaultRetryer는 Retryer 인터페이스의 기본 구현입니다.
type defaultRetryer struct {
	defaultConfig RetryConfig
}

// NewRetryer는 기본 설정으로 새 Retryer를 생성합니다.
func NewRetryer(opts ...RetryOption) Retryer {
	config := RetryConfig{
		MaxAttempts:     3, // 기본 3회 시도
		MaxDuration:     0, // 무제한
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,    // 지수 백오프
		JitterFactor:    0.1,    // 10% 지터
		ShouldRetry:     Always, // 기본적으로 항상 재시도
	}

	for _, opt := range opts {
		opt(&config)
	}

	return &defaultRetryer{
		defaultConfig: config,
	}
}

// 제한 시간 동안 시도 횟수를 제한하지 않습니다.
func WithUnlimitedAttempts(duration time.Duration) RetryOption {
	return func(config *RetryConfig) {
		config.MaxAttempts = 0
		config.MaxDuration = duration
	}
}

// WithMaxAttempts는 최대 시도 횟수를 설정합니다.
func WithMaxAttempts(attempts int) RetryOption {
	return func(config *RetryConfig) {
		config.MaxAttempts = attempts
	}
}

// WithMaxDuration은 최대 재시도 시간을 설정합니다.
func WithMaxDuration(duration time.Duration) RetryOption {
	return func(config *RetryConfig) {
		config.MaxDuration = duration
	}
}

// WithConstantBackoff는 일정한 대기 시간을 설정합니다.
func WithConstantBackoff(interval time.Duration) RetryOption {
	return func(config *RetryConfig) {
		config.InitialInterval = interval
		config.MaxInterval = interval
		config.Multiplier = 1.0
	}
}

// WithBackoff는 지수 백오프 설정을 구성합니다.
func WithBackoff(initialInterval, maxInterval time.Duration, multiplier float64) RetryOption {
	return func(config *RetryConfig) {
		config.InitialInterval = initialInterval
		config.MaxInterval = maxInterval
		config.Multiplier = multiplier
	}
}

// WithJitter는 지터 계수를 설정합니다.
func WithJitter(factor float64) RetryOption {
	return func(config *RetryConfig) {
		if factor < 0.0 {
			factor = 0.0
		}
		if factor > 1.0 {
			factor = 1.0
		}
		config.JitterFactor = factor
	}
}

// WithRetryPredicate는 재시도 조건을 설정합니다.
func WithRetryPredicate(predicate RetryPredicate) RetryOption {
	return func(config *RetryConfig) {
		config.ShouldRetry = predicate
	}
}

// Do는 지정된 함수를 재시도 정책에 따라 실행합니다.
func (r *defaultRetryer) Do(ctx context.Context, fn RetryableFunc, opts ...RetryOption) error {
	result := r.DoWithResult(ctx, fn, opts...)
	return result.Error
}

// DoWithResult는 지정된 함수를 재시도 정책에 따라 실행하고 결과를 반환합니다.
func (r *defaultRetryer) DoWithResult(ctx context.Context, fn RetryableFunc, opts ...RetryOption) *RetryResult {
	// 기본 설정 복사
	config := r.defaultConfig

	// 옵션 적용
	for _, opt := range opts {
		opt(&config)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	startTime := time.Now()
	interval := config.InitialInterval

	// 최대 시도 횟수 설정 (0은 무제한)
	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = int(^uint(0) >> 1) // 최대 int 값
	}

	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 함수 실행
		err := fn()
		if err == nil {
			// 성공
			return &RetryResult{
				Success:  true,
				Error:    nil,
				Attempts: attempt + 1,
				Duration: time.Since(startTime),
			}
		}

		lastErr = err

		// 재시도 여부 확인
		if config.ShouldRetry != nil && !config.ShouldRetry(err) {
			// 재시도 불가능한 오류
			return &RetryResult{
				Success:       false,
				Error:         err,
				Attempts:      attempt + 1,
				Duration:      time.Since(startTime),
				FailureReason: "non-retryable error",
			}
		}

		// 마지막 시도였으면 실패 반환
		if attempt == maxAttempts-1 {
			return &RetryResult{
				Success:       false,
				Error:         lastErr,
				Attempts:      attempt + 1,
				Duration:      time.Since(startTime),
				FailureReason: ErrMaxAttemptsReached.Error(),
			}
		}

		// 최대 시간 확인
		if config.MaxDuration > 0 && time.Since(startTime) >= config.MaxDuration {
			return &RetryResult{
				Success:       false,
				Error:         lastErr,
				Attempts:      attempt + 1,
				Duration:      time.Since(startTime),
				FailureReason: ErrMaxDurationReached.Error(),
			}
		}

		// 지터 계산
		jitterDelta := float64(interval) * config.JitterFactor * (rng.Float64()*2 - 1)
		sleepTime := time.Duration(float64(interval) + jitterDelta)

		// 남은 최대 시간 계산 (MaxDuration이 설정된 경우)
		if config.MaxDuration > 0 {
			remainingTime := config.MaxDuration - time.Since(startTime)
			if remainingTime < sleepTime {
				sleepTime = remainingTime
			}

			// 남은 시간이 없으면 종료
			if remainingTime <= 0 {
				return &RetryResult{
					Success:       false,
					Error:         lastErr,
					Attempts:      attempt + 1,
					Duration:      time.Since(startTime),
					FailureReason: ErrMaxDurationReached.Error(),
				}
			}
		}

		// 대기 후 재시도
		select {
		case <-ctx.Done():
			return &RetryResult{
				Success:       false,
				Error:         ctx.Err(),
				Attempts:      attempt + 1,
				Duration:      time.Since(startTime),
				FailureReason: "context cancelled",
			}
		case <-time.After(sleepTime):
			// 다음 시도 준비
		}

		// 다음 간격 계산 (지수 백오프)
		interval = time.Duration(float64(interval) * config.Multiplier)
		if interval > config.MaxInterval {
			interval = config.MaxInterval
		}
	}

	// 이 코드는 실행되지 않아야 함 (위에서 항상 반환됨)
	return &RetryResult{
		Success:       false,
		Error:         lastErr,
		Attempts:      maxAttempts,
		Duration:      time.Since(startTime),
		FailureReason: "unknown",
	}
}

// RetryerAdapter는 특정 타입에 대한 재시도 로직을 제공하는 어댑터입니다.
type RetryerAdapter[T any] struct {
	Retryer Retryer
}

// NewRetryerAdapter는 새 RetryerAdapter를 생성합니다.
func NewRetryerAdapter[T any](retryer Retryer) *RetryerAdapter[T] {
	return &RetryerAdapter[T]{Retryer: retryer}
}

// DoWithValue는 값을 반환하는 함수를 재시도 정책에 따라 실행합니다.
func (r *RetryerAdapter[T]) DoWithValue(
	ctx context.Context,
	fn func() (T, error),
	opts ...RetryOption,
) (*RetryResult, T) {
	var zeroValue T
	var lastValue T
	var success bool

	result := r.Retryer.DoWithResult(ctx, func() error {
		value, err := fn()
		if err != nil {
			return err
		}
		lastValue = value
		success = true
		return nil
	}, opts...)

	if success {
		return result, lastValue
	}

	return result, zeroValue
}
