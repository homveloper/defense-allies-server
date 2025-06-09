// migrator.go - 이벤트 저장소 마이그레이션 도구
package cqrsx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventStoreMigrator는 스토어 간 데이터 마이그레이션을 담당합니다
type EventStoreMigrator struct {
	source   EventStore
	target   EventStore
	config   MigrationConfig
	progress *MigrationProgress
	mu       sync.RWMutex
}

// MigrationConfig는 마이그레이션 설정입니다
type MigrationConfig struct {
	BatchSize       int           `json:"batchSize"`       // 배치 크기
	MaxConcurrency  int           `json:"maxConcurrency"`  // 최대 동시 처리
	VerifyData      bool          `json:"verifyData"`      // 데이터 검증 여부
	ContinueOnError bool          `json:"continueOnError"` // 오류 시 계속 진행
	Timeout         time.Duration `json:"timeout"`         // 타임아웃
	DryRun          bool          `json:"dryRun"`          // 시뮬레이션 모드
}

// MigrationProgress는 마이그레이션 진행 상황을 추적합니다
type MigrationProgress struct {
	TotalStreams     int           `json:"totalStreams"`
	ProcessedStreams int           `json:"processedStreams"`
	TotalEvents      int64         `json:"totalEvents"`
	ProcessedEvents  int64         `json:"processedEvents"`
	ErrorCount       int           `json:"errorCount"`
	StartTime        time.Time     `json:"startTime"`
	EstimatedTime    time.Duration `json:"estimatedTime"`
	CurrentStream    string        `json:"currentStream"`
	Status           string        `json:"status"`
}

// MigrationResult는 마이그레이션 결과입니다
type MigrationResult struct {
	Success         bool             `json:"success"`
	ProcessedEvents int64            `json:"processedEvents"`
	ErrorCount      int              `json:"errorCount"`
	Duration        time.Duration    `json:"duration"`
	Errors          []MigrationError `json:"errors"`
}

// MigrationError는 마이그레이션 오류 정보입니다
type MigrationError struct {
	StreamName string    `json:"streamName"`
	EventID    string    `json:"eventId"`
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
}

// NewEventStoreMigrator는 새로운 마이그레이터를 생성합니다
func NewEventStoreMigrator(source, target EventStore) *EventStoreMigrator {
	return &EventStoreMigrator{
		source: source,
		target: target,
		config: DefaultMigrationConfig(),
		progress: &MigrationProgress{
			Status: "initialized",
		},
	}
}

// SetConfig는 마이그레이션 설정을 변경합니다
func (m *EventStoreMigrator) SetConfig(config MigrationConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// GetProgress는 현재 진행 상황을 반환합니다
func (m *EventStoreMigrator) GetProgress() MigrationProgress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.progress
}

// Migrate는 전체 마이그레이션을 실행합니다
func (m *EventStoreMigrator) Migrate(ctx context.Context) error {
	m.updateProgress(func(p *MigrationProgress) {
		p.Status = "starting"
		p.StartTime = time.Now()
	})

	// 마이그레이션할 스트림 목록 수집
	streams, err := m.collectStreams(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect streams: %w", err)
	}

	m.updateProgress(func(p *MigrationProgress) {
		p.TotalStreams = len(streams)
		p.Status = "collecting_events"
	})

	// 총 이벤트 수 계산
	totalEvents, err := m.countTotalEvents(ctx, streams)
	if err != nil {
		return fmt.Errorf("failed to count events: %w", err)
	}

	m.updateProgress(func(p *MigrationProgress) {
		p.TotalEvents = totalEvents
		p.Status = "migrating"
	})

	// 병렬 마이그레이션 실행
	result, err := m.migrateStreams(ctx, streams)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	m.updateProgress(func(p *MigrationProgress) {
		p.Status = "completed"
	})

	if !result.Success {
		return fmt.Errorf("migration completed with %d errors", result.ErrorCount)
	}

	return nil
}

// MigrateStream은 단일 스트림을 마이그레이션합니다
func (m *EventStoreMigrator) MigrateStream(ctx context.Context, aggregateID uuid.UUID) error {
	m.updateProgress(func(p *MigrationProgress) {
		p.CurrentStream = aggregateID.String()
	})

	// 소스에서 이벤트 로드
	events, err := m.source.Load(ctx, aggregateID)
	if err != nil {
		return fmt.Errorf("failed to load events from source: %w", err)
	}

	if len(events) == 0 {
		return nil // 빈 스트림
	}

	// Dry run 모드 확인
	if m.config.DryRun {
		m.updateProgress(func(p *MigrationProgress) {
			p.ProcessedEvents += int64(len(events))
			p.ProcessedStreams++
		})
		return nil
	}

	// 배치 단위로 타겟에 저장
	return m.saveBatches(ctx, events)
}

// Verify는 마이그레이션 결과를 검증합니다
func (m *EventStoreMigrator) Verify(ctx context.Context) error {
	if !m.config.VerifyData {
		return nil
	}

	m.updateProgress(func(p *MigrationProgress) {
		p.Status = "verifying"
	})

	// 스트림별 검증 (샘플링)
	streams, err := m.collectStreams(ctx)
	if err != nil {
		return err
	}

	sampleSize := min(len(streams), 100) // 최대 100개 스트림 검증
	errorCount := 0

	for i := 0; i < sampleSize; i++ {
		if err := m.verifyStream(ctx, streams[i]); err != nil {
			errorCount++
			if errorCount > 10 { // 10개 이상 오류 시 중단
				return fmt.Errorf("verification failed: too many errors")
			}
		}
	}

	return nil
}

// Private methods

func (m *EventStoreMigrator) collectStreams(ctx context.Context) ([]uuid.UUID, error) {
	// QueryableEventStore인 경우 스트림 목록 조회
	if queryable, ok := m.source.(QueryableEventStore); ok {
		events, err := queryable.FindEvents(ctx, EventQuery{Limit: 10000})
		if err != nil {
			return nil, err
		}

		// 고유한 집합체 ID 수집
		streamMap := make(map[uuid.UUID]bool)
		for _, event := range events {
			streamMap[event.string()] = true
		}

		streams := make([]uuid.UUID, 0, len(streamMap))
		for id := range streamMap {
			streams = append(streams, id)
		}

		return streams, nil
	}

	// 일반 EventStore인 경우 다른 방법 필요
	// 실제 구현에서는 MongoDB 컬렉션을 직접 조회하거나
	// 별도의 스트림 레지스트리를 사용
	return nil, fmt.Errorf("source store does not support stream enumeration")
}

func (m *EventStoreMigrator) countTotalEvents(ctx context.Context, streams []uuid.UUID) (int64, error) {
	var total int64

	for _, streamID := range streams {
		events, err := m.source.Load(ctx, streamID)
		if err != nil {
			continue // 오류 무시하고 계속
		}
		total += int64(len(events))
	}

	return total, nil
}

func (m *EventStoreMigrator) migrateStreams(ctx context.Context, streams []uuid.UUID) (*MigrationResult, error) {

	startTime := time.Now()

	result := &MigrationResult{
		Success: true,
	}

	// 워커 풀 생성
	jobs := make(chan uuid.UUID, len(streams))
	results := make(chan error, len(streams))

	// 워커 시작
	for i := 0; i < m.config.MaxConcurrency; i++ {
		go m.migrationWorker(ctx, jobs, results)
	}

	// 작업 전송
	for _, streamID := range streams {
		jobs <- streamID
	}
	close(jobs)

	// 결과 수집
	for i := 0; i < len(streams); i++ {
		err := <-results
		if err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, MigrationError{
				Error:     err.Error(),
				Timestamp: time.Now(),
			})

			if !m.config.ContinueOnError {
				result.Success = false
				break
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

func (m *EventStoreMigrator) migrationWorker(ctx context.Context, jobs <-chan uuid.UUID, results chan<- error) {
	for streamID := range jobs {
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
			err := m.MigrateStream(ctx, streamID)
			results <- err
		}
	}
}

func (m *EventStoreMigrator) saveBatches(ctx context.Context, events []Event) error {
	batchSize := m.config.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	for i := 0; i < len(events); i += batchSize {
		end := min(i+batchSize, len(events))
		batch := events[i:end]

		expectedVersion := 0
		if len(batch) > 0 {
			expectedVersion = batch[0].Version() - 1
		}

		if err := m.target.Save(ctx, batch, expectedVersion); err != nil {
			return fmt.Errorf("failed to save batch: %w", err)
		}

		m.updateProgress(func(p *MigrationProgress) {
			p.ProcessedEvents += int64(len(batch))
		})
	}

	m.updateProgress(func(p *MigrationProgress) {
		p.ProcessedStreams++
	})

	return nil
}

func (m *EventStoreMigrator) verifyStream(ctx context.Context, streamID uuid.UUID) error {
	sourceEvents, err := m.source.Load(ctx, streamID)
	if err != nil {
		return err
	}

	targetEvents, err := m.target.Load(ctx, streamID)
	if err != nil {
		return err
	}

	if len(sourceEvents) != len(targetEvents) {
		return fmt.Errorf("event count mismatch for stream %s: source=%d, target=%d",
			streamID, len(sourceEvents), len(targetEvents))
	}

	// 이벤트 내용 검증 (간단한 비교)
	for i, sourceEvent := range sourceEvents {
		targetEvent := targetEvents[i]
		if sourceEvent.Version() != targetEvent.Version() ||
			sourceEvent.EventType() != targetEvent.EventType() {
			return fmt.Errorf("event mismatch at position %d for stream %s", i, streamID)
		}
	}

	return nil
}

func (m *EventStoreMigrator) updateProgress(updateFunc func(*MigrationProgress)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	updateFunc(m.progress)

	// 예상 완료 시간 계산
	if m.progress.ProcessedEvents > 0 && m.progress.TotalEvents > 0 {
		elapsed := time.Since(m.progress.StartTime)
		remaining := float64(m.progress.TotalEvents-m.progress.ProcessedEvents) /
			float64(m.progress.ProcessedEvents) * float64(elapsed)
		m.progress.EstimatedTime = time.Duration(remaining)
	}
}

// DefaultMigrationConfig는 기본 마이그레이션 설정을 반환합니다
func DefaultMigrationConfig() MigrationConfig {
	return MigrationConfig{
		BatchSize:       100,
		MaxConcurrency:  4,
		VerifyData:      true,
		ContinueOnError: false,
		Timeout:         30 * time.Minute,
		DryRun:          false,
	}
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MigrationProgressMonitor는 마이그레이션 진행 상황을 모니터링합니다
type MigrationProgressMonitor struct {
	migrator *EventStoreMigrator
	interval time.Duration
	stopCh   chan struct{}
}

// NewMigrationProgressMonitor는 새로운 진행 상황 모니터를 생성합니다
func NewMigrationProgressMonitor(migrator *EventStoreMigrator, interval time.Duration) *MigrationProgressMonitor {
	return &MigrationProgressMonitor{
		migrator: migrator,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start는 모니터링을 시작합니다
func (m *MigrationProgressMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			progress := m.migrator.GetProgress()
			m.logProgress(progress)
		}
	}
}

// Stop은 모니터링을 중지합니다
func (m *MigrationProgressMonitor) Stop() {
	close(m.stopCh)
}

func (m *MigrationProgressMonitor) logProgress(progress MigrationProgress) {
	if progress.TotalEvents > 0 {
		percentage := float64(progress.ProcessedEvents) / float64(progress.TotalEvents) * 100
		fmt.Printf("[Migration] Status: %s, Progress: %.1f%% (%d/%d events), ETA: %v\n",
			progress.Status, percentage, progress.ProcessedEvents, progress.TotalEvents, progress.EstimatedTime)
	} else {
		fmt.Printf("[Migration] Status: %s, Processed: %d events\n",
			progress.Status, progress.ProcessedEvents)
	}
}
