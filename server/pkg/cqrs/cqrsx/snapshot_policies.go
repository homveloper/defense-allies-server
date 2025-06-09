package cqrsx

import (
	"time"

	"cqrs"
)

// SnapshotPolicy defines when snapshots should be created
type SnapshotPolicy interface {
	// ShouldCreateSnapshot determines if a snapshot should be created
	ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool

	// GetSnapshotInterval returns the snapshot creation interval
	GetSnapshotInterval() int

	// GetPolicyName returns the policy name
	GetPolicyName() string
}

// EventCountPolicy creates snapshots based on event count
type EventCountPolicy struct {
	threshold int
}

// NewEventCountPolicy creates an event count based policy
func NewEventCountPolicy(threshold int) *EventCountPolicy {
	if threshold <= 0 {
		threshold = 10 // default value
	}
	return &EventCountPolicy{
		threshold: threshold,
	}
}

func (p *EventCountPolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	return eventCount > 0 && eventCount%p.threshold == 0
}

func (p *EventCountPolicy) GetSnapshotInterval() int {
	return p.threshold
}

func (p *EventCountPolicy) GetPolicyName() string {
	return "EventCountPolicy"
}

// TimeBasedPolicy creates snapshots based on time intervals
type TimeBasedPolicy struct {
	interval     time.Duration
	lastSnapshot map[string]time.Time
}

// NewTimeBasedPolicy creates a time based policy
func NewTimeBasedPolicy(interval time.Duration) *TimeBasedPolicy {
	if interval <= 0 {
		interval = 1 * time.Hour // default: 1 hour
	}
	return &TimeBasedPolicy{
		interval:     interval,
		lastSnapshot: make(map[string]time.Time),
	}
}

func (p *TimeBasedPolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	aggregateID := aggregate.ID()
	lastTime, exists := p.lastSnapshot[aggregateID]

	if !exists {
		p.lastSnapshot[aggregateID] = time.Now()
		return true // first snapshot
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

// VersionBasedPolicy creates snapshots based on version intervals
type VersionBasedPolicy struct {
	versionInterval int
}

// NewVersionBasedPolicy creates a version based policy
func NewVersionBasedPolicy(versionInterval int) *VersionBasedPolicy {
	if versionInterval <= 0 {
		versionInterval = 5 // default: every 5 versions
	}
	return &VersionBasedPolicy{
		versionInterval: versionInterval,
	}
}

func (p *VersionBasedPolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	version := aggregate.Version()
	return version > 0 && version%p.versionInterval == 0
}

func (p *VersionBasedPolicy) GetSnapshotInterval() int {
	return p.versionInterval
}

func (p *VersionBasedPolicy) GetPolicyName() string {
	return "VersionBasedPolicy"
}

// CompositePolicy combines multiple policies
type CompositePolicy struct {
	policies []SnapshotPolicy
	operator string // "AND" or "OR"
}

// NewCompositePolicy creates a composite policy
func NewCompositePolicy(operator string, policies ...SnapshotPolicy) *CompositePolicy {
	if operator != "AND" && operator != "OR" {
		operator = "OR" // default
	}
	return &CompositePolicy{
		policies: policies,
		operator: operator,
	}
}

func (p *CompositePolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	if len(p.policies) == 0 {
		return false
	}

	if p.operator == "AND" {
		// All policies must return true
		for _, policy := range p.policies {
			if !policy.ShouldCreateSnapshot(aggregate, eventCount) {
				return false
			}
		}
		return true
	} else {
		// Any policy returning true is sufficient (OR)
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
		// Return the largest interval
		maxInterval := 0
		for _, policy := range p.policies {
			interval := policy.GetSnapshotInterval()
			if interval > maxInterval {
				maxInterval = interval
			}
		}
		return maxInterval
	} else {
		// Return the smallest interval (OR)
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

// CustomPolicy allows function-based policies
type CustomPolicy struct {
	name     string
	interval int
	checkFn  func(aggregate cqrs.AggregateRoot, eventCount int) bool
}

// NewCustomPolicy creates a custom policy
func NewCustomPolicy(name string, interval int, checkFn func(cqrs.AggregateRoot, int) bool) *CustomPolicy {
	return &CustomPolicy{
		name:     name,
		interval: interval,
		checkFn:  checkFn,
	}
}

func (p *CustomPolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
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

// AlwaysPolicy always creates snapshots (for testing)
type AlwaysPolicy struct{}

func NewAlwaysPolicy() *AlwaysPolicy {
	return &AlwaysPolicy{}
}

func (p *AlwaysPolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	return true
}

func (p *AlwaysPolicy) GetSnapshotInterval() int {
	return 1
}

func (p *AlwaysPolicy) GetPolicyName() string {
	return "AlwaysPolicy"
}

// NeverPolicy never creates snapshots (for testing)
type NeverPolicy struct{}

func NewNeverPolicy() *NeverPolicy {
	return &NeverPolicy{}
}

func (p *NeverPolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	return false
}

func (p *NeverPolicy) GetSnapshotInterval() int {
	return 0
}

func (p *NeverPolicy) GetPolicyName() string {
	return "NeverPolicy"
}

// AdaptivePolicy adapts based on performance metrics
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

// NewAdaptivePolicy creates an adaptive policy
func NewAdaptivePolicy(baseThreshold int, adaptationFactor float64) *AdaptivePolicy {
	return &AdaptivePolicy{
		baseThreshold:    baseThreshold,
		performanceData:  make(map[string]*PerformanceMetrics),
		adaptationFactor: adaptationFactor,
	}
}

func (p *AdaptivePolicy) ShouldCreateSnapshot(aggregate cqrs.AggregateRoot, eventCount int) bool {
	aggregateID := aggregate.ID()

	// Use base threshold if no performance data
	metrics, exists := p.performanceData[aggregateID]
	if !exists {
		return eventCount >= p.baseThreshold
	}

	// Create snapshots more frequently if restore time is long
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

// UpdatePerformanceMetrics updates performance metrics
func (p *AdaptivePolicy) UpdatePerformanceMetrics(aggregateID string, restoreTime time.Duration, eventCount int) {
	metrics, exists := p.performanceData[aggregateID]
	if !exists {
		metrics = &PerformanceMetrics{}
		p.performanceData[aggregateID] = metrics
	}

	// Calculate moving average
	if metrics.AverageRestoreTime == 0 {
		metrics.AverageRestoreTime = restoreTime
	} else {
		metrics.AverageRestoreTime = (metrics.AverageRestoreTime + restoreTime) / 2
	}

	metrics.EventCount = eventCount
	metrics.LastMeasurement = time.Now()
}
