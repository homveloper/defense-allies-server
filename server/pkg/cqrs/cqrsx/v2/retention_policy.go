// retention_policy.go - 상태 보존 정책
package cqrsx

import (
	"context"
	"sort"
	"time"
)

// KeepLastNPolicy는 최신 N개의 상태만 보존하는 정책입니다
type KeepLastNPolicy struct {
	count int
}

// NewKeepLastNPolicy는 새로운 KeepLastN 정책을 생성합니다
func NewKeepLastNPolicy(count int) *KeepLastNPolicy {
	if count <= 0 {
		count = 1 // 최소 1개는 보존
	}
	return &KeepLastNPolicy{count: count}
}

// ShouldKeep은 상태를 보존할지 결정합니다
func (p *KeepLastNPolicy) ShouldKeep(state *AggregateState) bool {
	// 개별 상태로는 판단할 수 없으므로 항상 true 반환
	// GetCleanupCandidates에서 실제 로직 수행
	return true
}

// GetCleanupCandidates는 정리 대상 상태들을 반환합니다
func (p *KeepLastNPolicy) GetCleanupCandidates(ctx context.Context, states []*AggregateState) []*AggregateState {
	if len(states) <= p.count {
		return nil // 정리할 대상 없음
	}

	// 버전으로 정렬 (최신순)
	sorted := make([]*AggregateState, len(states))
	copy(sorted, states)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version > sorted[j].Version
	})

	// 오래된 상태들을 정리 대상으로 반환
	return sorted[p.count:]
}

// TimeBasedPolicy는 시간 기반 보존 정책입니다
type TimeBasedPolicy struct {
	maxAge time.Duration
}

// NewTimeBasedPolicy는 새로운 시간 기반 정책을 생성합니다
func NewTimeBasedPolicy(maxAge time.Duration) *TimeBasedPolicy {
	return &TimeBasedPolicy{maxAge: maxAge}
}

// ShouldKeep은 상태를 보존할지 결정합니다
func (p *TimeBasedPolicy) ShouldKeep(state *AggregateState) bool {
	return time.Since(state.Timestamp) <= p.maxAge
}

// GetCleanupCandidates는 정리 대상 상태들을 반환합니다
func (p *TimeBasedPolicy) GetCleanupCandidates(ctx context.Context, states []*AggregateState) []*AggregateState {
	var candidates []*AggregateState
	cutoff := time.Now().Add(-p.maxAge)

	for _, state := range states {
		if state.Timestamp.Before(cutoff) {
			candidates = append(candidates, state)
		}
	}

	return candidates
}

// SizeBasedPolicy는 크기 기반 보존 정책입니다
type SizeBasedPolicy struct {
	maxTotalSize int64
}

// NewSizeBasedPolicy는 새로운 크기 기반 정책을 생성합니다
func NewSizeBasedPolicy(maxTotalSize int64) *SizeBasedPolicy {
	return &SizeBasedPolicy{maxTotalSize: maxTotalSize}
}

// ShouldKeep은 상태를 보존할지 결정합니다
func (p *SizeBasedPolicy) ShouldKeep(state *AggregateState) bool {
	// 개별 상태로는 판단할 수 없으므로 항상 true 반환
	return true
}

// GetCleanupCandidates는 정리 대상 상태들을 반환합니다
func (p *SizeBasedPolicy) GetCleanupCandidates(ctx context.Context, states []*AggregateState) []*AggregateState {
	// 버전으로 정렬 (최신순)
	sorted := make([]*AggregateState, len(states))
	copy(sorted, states)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version > sorted[j].Version
	})

	var totalSize int64
	var keepIndex int

	// 최신부터 누적하여 크기 제한까지 보존
	for i, state := range sorted {
		totalSize += state.Size()
		if totalSize > p.maxTotalSize {
			keepIndex = i
			break
		}
		keepIndex = i + 1
	}

	// 크기 제한을 초과하는 오래된 상태들을 정리 대상으로 반환
	if keepIndex < len(sorted) {
		return sorted[keepIndex:]
	}

	return nil
}

// CompositePolicy는 여러 정책을 조합한 정책입니다
type CompositePolicy struct {
	policies []RetentionPolicy
	mode     CompositeMode
}

type CompositeMode int

const (
	CompositeModeAND CompositeMode = iota // 모든 정책을 만족해야 보존
	CompositeModeOR                       // 하나의 정책만 만족하면 보존
)

// NewCompositePolicy는 새로운 조합 정책을 생성합니다
func NewCompositePolicy(mode CompositeMode, policies ...RetentionPolicy) *CompositePolicy {
	return &CompositePolicy{
		policies: policies,
		mode:     mode,
	}
}

// ShouldKeep은 상태를 보존할지 결정합니다
func (p *CompositePolicy) ShouldKeep(state *AggregateState) bool {
	if len(p.policies) == 0 {
		return true
	}

	switch p.mode {
	case CompositeModeAND:
		for _, policy := range p.policies {
			if !policy.ShouldKeep(state) {
				return false
			}
		}
		return true

	case CompositeModeOR:
		for _, policy := range p.policies {
			if policy.ShouldKeep(state) {
				return true
			}
		}
		return false

	default:
		return true
	}
}

// GetCleanupCandidates는 정리 대상 상태들을 반환합니다
func (p *CompositePolicy) GetCleanupCandidates(ctx context.Context, states []*AggregateState) []*AggregateState {
	if len(p.policies) == 0 {
		return nil
	}

	// 모든 정책의 정리 대상을 수집
	candidateMap := make(map[*AggregateState]int)

	for _, policy := range p.policies {
		candidates := policy.GetCleanupCandidates(ctx, states)
		for _, candidate := range candidates {
			candidateMap[candidate]++
		}
	}

	var result []*AggregateState

	switch p.mode {
	case CompositeModeAND:
		// 모든 정책에서 정리 대상으로 선정된 상태들만
		for state, count := range candidateMap {
			if count == len(p.policies) {
				result = append(result, state)
			}
		}

	case CompositeModeOR:
		// 하나 이상의 정책에서 정리 대상으로 선정된 상태들
		for state := range candidateMap {
			result = append(result, state)
		}
	}

	return result
}

// NoRetentionPolicy는 모든 상태를 보존하는 정책입니다 (기본값)
type NoRetentionPolicy struct{}

// NewNoRetentionPolicy는 새로운 무보존 정책을 생성합니다
func NewNoRetentionPolicy() *NoRetentionPolicy {
	return &NoRetentionPolicy{}
}

// ShouldKeep은 모든 상태를 보존합니다
func (p *NoRetentionPolicy) ShouldKeep(state *AggregateState) bool {
	return true
}

// GetCleanupCandidates는 정리 대상이 없음을 반환합니다
func (p *NoRetentionPolicy) GetCleanupCandidates(ctx context.Context, states []*AggregateState) []*AggregateState {
	return nil
}

// 편의 함수들

// KeepLast는 최신 N개만 보존하는 정책을 생성합니다
func KeepLast(count int) RetentionPolicy {
	return NewKeepLastNPolicy(count)
}

// KeepForDuration은 특정 기간 동안만 보존하는 정책을 생성합니다
func KeepForDuration(duration time.Duration) RetentionPolicy {
	return NewTimeBasedPolicy(duration)
}

// KeepWithinSize는 특정 크기 내에서만 보존하는 정책을 생성합니다
func KeepWithinSize(maxSize int64) RetentionPolicy {
	return NewSizeBasedPolicy(maxSize)
}

// CombineWithAND는 AND 조건으로 정책들을 조합합니다
func CombineWithAND(policies ...RetentionPolicy) RetentionPolicy {
	return NewCompositePolicy(CompositeModeAND, policies...)
}

// CombineWithOR는 OR 조건으로 정책들을 조합합니다
func CombineWithOR(policies ...RetentionPolicy) RetentionPolicy {
	return NewCompositePolicy(CompositeModeOR, policies...)
}
