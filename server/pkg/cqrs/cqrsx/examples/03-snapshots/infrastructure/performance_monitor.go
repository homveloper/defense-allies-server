package infrastructure

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PerformanceMetrics 성능 측정 지표
type PerformanceMetrics struct {
	AggregateID       string        `json:"aggregate_id"`
	OperationType     string        `json:"operation_type"`
	StartTime         time.Time     `json:"start_time"`
	EndTime           time.Time     `json:"end_time"`
	Duration          time.Duration `json:"duration"`
	EventsProcessed   int           `json:"events_processed"`
	SnapshotUsed      bool          `json:"snapshot_used"`
	SnapshotVersion   int           `json:"snapshot_version,omitempty"`
	MemoryUsageBefore int64         `json:"memory_usage_before"`
	MemoryUsageAfter  int64         `json:"memory_usage_after"`
	MemoryDifference  int64         `json:"memory_difference"`
	Success           bool          `json:"success"`
	ErrorMessage      string        `json:"error_message,omitempty"`
}

// PerformanceReport 성능 보고서
type PerformanceReport struct {
	TotalOperations      int                  `json:"total_operations"`
	SuccessfulOps        int                  `json:"successful_operations"`
	FailedOps            int                  `json:"failed_operations"`
	AverageDuration      time.Duration        `json:"average_duration"`
	MinDuration          time.Duration        `json:"min_duration"`
	MaxDuration          time.Duration        `json:"max_duration"`
	TotalEventsProcessed int                  `json:"total_events_processed"`
	SnapshotUsageRate    float64              `json:"snapshot_usage_rate"`
	MemoryEfficiency     float64              `json:"memory_efficiency"`
	OperationsByType     map[string]int       `json:"operations_by_type"`
	Metrics              []PerformanceMetrics `json:"metrics"`
	GeneratedAt          time.Time            `json:"generated_at"`
}

// PerformanceMonitor 성능 모니터링 인터페이스
type PerformanceMonitor interface {
	// StartOperation 작업 시작 추적
	StartOperation(ctx context.Context, aggregateID, operationType string) string

	// EndOperation 작업 종료 추적
	EndOperation(ctx context.Context, operationID string, eventsProcessed int, snapshotUsed bool, snapshotVersion int, err error)

	// RecordMemoryUsage 메모리 사용량 기록
	RecordMemoryUsage(ctx context.Context, operationID string, before, after int64)

	// GetMetrics 성능 지표 조회
	GetMetrics(ctx context.Context, aggregateID string) ([]PerformanceMetrics, error)

	// GenerateReport 성능 보고서 생성
	GenerateReport(ctx context.Context, since time.Time) (*PerformanceReport, error)

	// ClearMetrics 지표 초기화
	ClearMetrics(ctx context.Context) error

	// GetRealTimeStats 실시간 통계 조회
	GetRealTimeStats(ctx context.Context) (map[string]interface{}, error)
}

// InMemoryPerformanceMonitor 메모리 기반 성능 모니터
type InMemoryPerformanceMonitor struct {
	metrics          map[string]PerformanceMetrics
	activeOperations map[string]*PerformanceMetrics
	mutex            sync.RWMutex
}

// NewInMemoryPerformanceMonitor 새로운 메모리 기반 성능 모니터 생성
func NewInMemoryPerformanceMonitor() *InMemoryPerformanceMonitor {
	return &InMemoryPerformanceMonitor{
		metrics:          make(map[string]PerformanceMetrics),
		activeOperations: make(map[string]*PerformanceMetrics),
	}
}

// StartOperation 작업 시작 추적
func (m *InMemoryPerformanceMonitor) StartOperation(ctx context.Context, aggregateID, operationType string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	operationID := fmt.Sprintf("%s_%s_%d", aggregateID, operationType, time.Now().UnixNano())

	metric := &PerformanceMetrics{
		AggregateID:   aggregateID,
		OperationType: operationType,
		StartTime:     time.Now(),
	}

	m.activeOperations[operationID] = metric
	return operationID
}

// EndOperation 작업 종료 추적
func (m *InMemoryPerformanceMonitor) EndOperation(ctx context.Context, operationID string, eventsProcessed int, snapshotUsed bool, snapshotVersion int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metric, exists := m.activeOperations[operationID]
	if !exists {
		return
	}

	metric.EndTime = time.Now()
	metric.Duration = metric.EndTime.Sub(metric.StartTime)
	metric.EventsProcessed = eventsProcessed
	metric.SnapshotUsed = snapshotUsed
	metric.SnapshotVersion = snapshotVersion
	metric.Success = err == nil
	if err != nil {
		metric.ErrorMessage = err.Error()
	}

	// 완료된 메트릭을 저장
	m.metrics[operationID] = *metric

	// 활성 작업에서 제거
	delete(m.activeOperations, operationID)
}

// RecordMemoryUsage 메모리 사용량 기록
func (m *InMemoryPerformanceMonitor) RecordMemoryUsage(ctx context.Context, operationID string, before, after int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if metric, exists := m.activeOperations[operationID]; exists {
		metric.MemoryUsageBefore = before
		metric.MemoryUsageAfter = after
		metric.MemoryDifference = after - before
	}
}

// GetMetrics 성능 지표 조회
func (m *InMemoryPerformanceMonitor) GetMetrics(ctx context.Context, aggregateID string) ([]PerformanceMetrics, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []PerformanceMetrics
	for _, metric := range m.metrics {
		if aggregateID == "" || metric.AggregateID == aggregateID {
			result = append(result, metric)
		}
	}

	return result, nil
}

// GenerateReport 성능 보고서 생성
func (m *InMemoryPerformanceMonitor) GenerateReport(ctx context.Context, since time.Time) (*PerformanceReport, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var relevantMetrics []PerformanceMetrics
	operationsByType := make(map[string]int)
	totalDuration := time.Duration(0)
	minDuration := time.Duration(0)
	maxDuration := time.Duration(0)
	successfulOps := 0
	failedOps := 0
	totalEventsProcessed := 0
	snapshotUsedCount := 0
	totalMemoryDiff := int64(0)

	for _, metric := range m.metrics {
		if metric.StartTime.After(since) {
			relevantMetrics = append(relevantMetrics, metric)
			operationsByType[metric.OperationType]++
			totalDuration += metric.Duration
			totalEventsProcessed += metric.EventsProcessed
			totalMemoryDiff += metric.MemoryDifference

			if metric.Success {
				successfulOps++
			} else {
				failedOps++
			}

			if metric.SnapshotUsed {
				snapshotUsedCount++
			}

			if minDuration == 0 || metric.Duration < minDuration {
				minDuration = metric.Duration
			}
			if metric.Duration > maxDuration {
				maxDuration = metric.Duration
			}
		}
	}

	totalOps := len(relevantMetrics)
	var avgDuration time.Duration
	var snapshotUsageRate float64
	var memoryEfficiency float64

	if totalOps > 0 {
		avgDuration = totalDuration / time.Duration(totalOps)
		snapshotUsageRate = float64(snapshotUsedCount) / float64(totalOps) * 100
		if totalMemoryDiff != 0 {
			memoryEfficiency = float64(totalEventsProcessed) / float64(totalMemoryDiff) * 1000 // events per KB
		}
	}

	return &PerformanceReport{
		TotalOperations:      totalOps,
		SuccessfulOps:        successfulOps,
		FailedOps:            failedOps,
		AverageDuration:      avgDuration,
		MinDuration:          minDuration,
		MaxDuration:          maxDuration,
		TotalEventsProcessed: totalEventsProcessed,
		SnapshotUsageRate:    snapshotUsageRate,
		MemoryEfficiency:     memoryEfficiency,
		OperationsByType:     operationsByType,
		Metrics:              relevantMetrics,
		GeneratedAt:          time.Now(),
	}, nil
}

// ClearMetrics 지표 초기화
func (m *InMemoryPerformanceMonitor) ClearMetrics(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.metrics = make(map[string]PerformanceMetrics)
	m.activeOperations = make(map[string]*PerformanceMetrics)
	return nil
}

// GetRealTimeStats 실시간 통계 조회
func (m *InMemoryPerformanceMonitor) GetRealTimeStats(ctx context.Context) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_metrics":     len(m.metrics),
		"active_operations": len(m.activeOperations),
		"last_updated":      time.Now(),
	}

	// 최근 1분간의 통계
	oneMinuteAgo := time.Now().Add(-time.Minute)
	recentOps := 0
	recentSuccessful := 0
	recentFailed := 0

	for _, metric := range m.metrics {
		if metric.StartTime.After(oneMinuteAgo) {
			recentOps++
			if metric.Success {
				recentSuccessful++
			} else {
				recentFailed++
			}
		}
	}

	stats["recent_operations"] = map[string]interface{}{
		"total":      recentOps,
		"successful": recentSuccessful,
		"failed":     recentFailed,
		"success_rate": func() float64 {
			if recentOps > 0 {
				return float64(recentSuccessful) / float64(recentOps) * 100
			}
			return 0
		}(),
	}

	return stats, nil
}

// BenchmarkResult 벤치마크 결과
type BenchmarkResult struct {
	TestName               string        `json:"test_name"`
	WithSnapshot           bool          `json:"with_snapshot"`
	EventCount             int           `json:"event_count"`
	RestorationTime        time.Duration `json:"restoration_time"`
	MemoryUsage            int64         `json:"memory_usage"`
	PerformanceImprovement float64       `json:"performance_improvement,omitempty"`
}

// PerformanceBenchmark 성능 벤치마크 실행
func (m *InMemoryPerformanceMonitor) PerformanceBenchmark(ctx context.Context, aggregateID string, eventCounts []int) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	for _, eventCount := range eventCounts {
		// 스냅샷 없이 테스트
		withoutSnapshot := m.benchmarkRestore(ctx, aggregateID, eventCount, false)
		withoutSnapshot.TestName = fmt.Sprintf("Restore_%d_events_without_snapshot", eventCount)
		results = append(results, withoutSnapshot)

		// 스냅샷과 함께 테스트
		withSnapshot := m.benchmarkRestore(ctx, aggregateID, eventCount, true)
		withSnapshot.TestName = fmt.Sprintf("Restore_%d_events_with_snapshot", eventCount)

		// 성능 개선 계산
		if withoutSnapshot.RestorationTime > 0 {
			improvement := float64(withoutSnapshot.RestorationTime-withSnapshot.RestorationTime) / float64(withoutSnapshot.RestorationTime) * 100
			withSnapshot.PerformanceImprovement = improvement
		}

		results = append(results, withSnapshot)
	}

	return results, nil
}

// benchmarkRestore 복원 벤치마크 실행
func (m *InMemoryPerformanceMonitor) benchmarkRestore(ctx context.Context, aggregateID string, eventCount int, useSnapshot bool) BenchmarkResult {
	// 실제 복원 로직은 여기서 구현
	// 이는 예시이므로 시뮬레이션된 시간을 사용
	var restorationTime time.Duration
	var memoryUsage int64

	if useSnapshot {
		// 스냅샷 사용 시 더 빠른 복원
		restorationTime = time.Duration(eventCount/10) * time.Millisecond
		memoryUsage = int64(eventCount * 100) // 더 적은 메모리 사용
	} else {
		// 스냅샷 없이 모든 이벤트 재생
		restorationTime = time.Duration(eventCount) * time.Millisecond
		memoryUsage = int64(eventCount * 200) // 더 많은 메모리 사용
	}

	return BenchmarkResult{
		WithSnapshot:    useSnapshot,
		EventCount:      eventCount,
		RestorationTime: restorationTime,
		MemoryUsage:     memoryUsage,
	}
}
