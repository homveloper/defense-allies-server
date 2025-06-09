package snapshots

import (
	"time"
)

// EventCountPolicy 이벤트 개수 기반 스냅샷 정책
type EventCountPolicy struct {
	threshold int
}

// NewEventCountPolicy 이벤트 개수 기반 정책 생성
func NewEventCountPolicy(threshold int) *EventCountPolicy {
	if threshold <= 0 {
		threshold = 10 // 기본값
	}
	return &EventCountPolicy{
		threshold: threshold,
	}
}

func (p *EventCountPolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	return eventCount > 0 && eventCount%p.threshold == 0
}

func (p *EventCountPolicy) GetSnapshotInterval() int {
	return p.threshold
}

func (p *EventCountPolicy) GetPolicyName() string {
	return "EventCountPolicy"
}

// TimeBasedPolicy 시간 기반 스냅샷 정책
type TimeBasedPolicy struct {
	interval     time.Duration
	lastSnapshot map[string]time.Time
}

// NewTimeBasedPolicy 시간 기반 정책 생성
func NewTimeBasedPolicy(interval time.Duration) *TimeBasedPolicy {
	if interval <= 0 {
		interval = 1 * time.Hour // 기본값: 1시간
	}
	return &TimeBasedPolicy{
		interval:     interval,
		lastSnapshot: make(map[string]time.Time),
	}
}

func (p *TimeBasedPolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	aggregateID := aggregate.ID()
	lastTime, exists := p.lastSnapshot[aggregateID]

	if !exists {
		p.lastSnapshot[aggregateID] = time.Now()
		return true // 첫 번째 스냅샷
	}

	if time.Since(lastTime) >= p.interval {
		p.lastSnapshot[aggregateID] = time.Now()
		return true
	}

	return false
}

func (p *TimeBasedPolicy) GetSnapshotInterval() int {
	return int(p.interval.Minutes())
}

func (p *TimeBasedPolicy) GetPolicyName() string {
	return "TimeBasedPolicy"
}

// VersionBasedPolicy 버전 기반 스냅샷 정책
type VersionBasedPolicy struct {
	versionInterval int
}

// NewVersionBasedPolicy 버전 기반 정책 생성
func NewVersionBasedPolicy(versionInterval int) *VersionBasedPolicy {
	if versionInterval <= 0 {
		versionInterval = 5 // 기본값: 5버전마다
	}
	return &VersionBasedPolicy{
		versionInterval: versionInterval,
	}
}

func (p *VersionBasedPolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	version := aggregate.Version()
	return version > 0 && version%p.versionInterval == 0
}

func (p *VersionBasedPolicy) GetSnapshotInterval() int {
	return p.versionInterval
}

func (p *VersionBasedPolicy) GetPolicyName() string {
	return "VersionBasedPolicy"
}

// CompositePolicy 복합 정책 (여러 정책 조합)
type CompositePolicy struct {
	policies []SnapshotPolicy
	operator string // "AND" 또는 "OR"
}

// NewCompositePolicy 복합 정책 생성
func NewCompositePolicy(operator string, policies ...SnapshotPolicy) *CompositePolicy {
	if operator != "AND" && operator != "OR" {
		operator = "OR" // 기본값
	}
	return &CompositePolicy{
		policies: policies,
		operator: operator,
	}
}

func (p *CompositePolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	if len(p.policies) == 0 {
		return false
	}

	if p.operator == "AND" {
		// 모든 정책이 true여야 함
		for _, policy := range p.policies {
			if !policy.ShouldCreateSnapshot(aggregate, eventCount) {
				return false
			}
		}
		return true
	} else {
		// 하나라도 true면 됨 (OR)
		for _, policy := range p.policies {
			if policy.ShouldCreateSnapshot(aggregate, eventCount) {
				return true
			}
		}
		return false
	}
}

func (p *CompositePolicy) GetSnapshotInterval() int {
	if len(p.policies) == 0 {
		return 0
	}

	if p.operator == "AND" {
		// 가장 큰 간격 반환
		maxInterval := 0
		for _, policy := range p.policies {
			interval := policy.GetSnapshotInterval()
			if interval > maxInterval {
				maxInterval = interval
			}
		}
		return maxInterval
	} else {
		// 가장 작은 간격 반환 (OR)
		minInterval := int(^uint(0) >> 1) // max int
		for _, policy := range p.policies {
			interval := policy.GetSnapshotInterval()
			if interval < minInterval {
				minInterval = interval
			}
		}
		return minInterval
	}
}

func (p *CompositePolicy) GetPolicyName() string {
	return "CompositePolicy"
}

// CustomPolicy 커스텀 정책 (함수 기반)
type CustomPolicy struct {
	name     string
	interval int
	checkFn  func(aggregate Aggregate, eventCount int) bool
}

// NewCustomPolicy 커스텀 정책 생성
func NewCustomPolicy(name string, interval int, checkFn func(Aggregate, int) bool) *CustomPolicy {
	return &CustomPolicy{
		name:     name,
		interval: interval,
		checkFn:  checkFn,
	}
}

func (p *CustomPolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	if p.checkFn == nil {
		return false
	}
	return p.checkFn(aggregate, eventCount)
}

func (p *CustomPolicy) GetSnapshotInterval() int {
	return p.interval
}

func (p *CustomPolicy) GetPolicyName() string {
	return p.name
}

// AlwaysPolicy 항상 스냅샷 생성 정책 (테스트용)
type AlwaysPolicy struct{}

func NewAlwaysPolicy() *AlwaysPolicy {
	return &AlwaysPolicy{}
}

func (p *AlwaysPolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	return true
}

func (p *AlwaysPolicy) GetSnapshotInterval() int {
	return 1
}

func (p *AlwaysPolicy) GetPolicyName() string {
	return "AlwaysPolicy"
}

// NeverPolicy 절대 스냅샷 생성하지 않는 정책 (테스트용)
type NeverPolicy struct{}

func NewNeverPolicy() *NeverPolicy {
	return &NeverPolicy{}
}

func (p *NeverPolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	return false
}

func (p *NeverPolicy) GetSnapshotInterval() int {
	return 0
}

func (p *NeverPolicy) GetPolicyName() string {
	return "NeverPolicy"
}

// AdaptivePolicy 적응형 정책 (성능 기반)
type AdaptivePolicy struct {
	baseThreshold    int
	performanceData  map[string]*PerformanceMetrics
	adaptationFactor float64
}

type PerformanceMetrics struct {
	AverageRestoreTime time.Duration
	EventCount         int
	LastMeasurement    time.Time
}

// NewAdaptivePolicy 적응형 정책 생성
func NewAdaptivePolicy(baseThreshold int, adaptationFactor float64) *AdaptivePolicy {
	return &AdaptivePolicy{
		baseThreshold:    baseThreshold,
		performanceData:  make(map[string]*PerformanceMetrics),
		adaptationFactor: adaptationFactor,
	}
}

func (p *AdaptivePolicy) ShouldCreateSnapshot(aggregate Aggregate, eventCount int) bool {
	aggregateID := aggregate.ID()

	// 성능 데이터가 없으면 기본 임계값 사용
	metrics, exists := p.performanceData[aggregateID]
	if !exists {
		return eventCount >= p.baseThreshold
	}

	// 복원 시간이 길수록 더 자주 스냅샷 생성
	adaptedThreshold := float64(p.baseThreshold)
	if metrics.AverageRestoreTime > 100*time.Millisecond {
		adaptedThreshold *= p.adaptationFactor
	}

	return eventCount >= int(adaptedThreshold)
}

func (p *AdaptivePolicy) GetSnapshotInterval() int {
	return p.baseThreshold
}

func (p *AdaptivePolicy) GetPolicyName() string {
	return "AdaptivePolicy"
}

// UpdatePerformanceMetrics 성능 메트릭 업데이트
func (p *AdaptivePolicy) UpdatePerformanceMetrics(aggregateID string, restoreTime time.Duration, eventCount int) {
	metrics, exists := p.performanceData[aggregateID]
	if !exists {
		metrics = &PerformanceMetrics{}
		p.performanceData[aggregateID] = metrics
	}

	// 이동 평균 계산
	if metrics.AverageRestoreTime == 0 {
		metrics.AverageRestoreTime = restoreTime
	} else {
		metrics.AverageRestoreTime = (metrics.AverageRestoreTime + restoreTime) / 2
	}

	metrics.EventCount = eventCount
	metrics.LastMeasurement = time.Now()
}
