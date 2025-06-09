// hybrid_store.go - 하이브리드 방식 구현
package cqrsx

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// HybridEventStore는 Hot/Cold 데이터 분리 방식 저장소입니다
type HybridEventStore struct {
	hotStore       EventStore          // 최근 데이터용 (StreamEventStore)
	coldStore      QueryableEventStore // 과거 데이터용 (DocumentEventStore)
	archiveManager *ArchiveManager
	config         HybridConfig
	metrics        *hybridMetrics
}

// HybridConfig는 하이브리드 저장소 설정입니다
type HybridConfig struct {
	HotDataThreshold  time.Duration `json:"hotDataThreshold"`  // 예: 30일
	ArchiveInterval   time.Duration `json:"archiveInterval"`   // 예: 6시간
	MaxHotEvents      int           `json:"maxHotEvents"`      // Hot 스토어 최대 이벤트 수
	EnableAutoArchive bool          `json:"enableAutoArchive"` // 자동 아카이빙 활성화
}

type hybridMetrics struct {
	mu                sync.RWMutex
	hotOperations     int64
	coldOperations    int64
	archiveOperations int64
	totalSaveTime     time.Duration
	totalLoadTime     time.Duration
	errors            int64
	lastOperation     time.Time
	lastArchive       time.Time
}

// ArchiveManager는 Hot/Cold 데이터 이동을 관리합니다
type ArchiveManager struct {
	hotStore  EventStore
	coldStore EventStore
	config    HybridConfig
	running   bool
	stopCh    chan struct{}
	mu        sync.RWMutex
}

// NewHybridEventStore는 새로운 하이브리드 이벤트 저장소를 생성합니다
func NewHybridEventStore(hotStore EventStore, coldStore QueryableEventStore, config HybridConfig) *HybridEventStore {
	archiveManager := &ArchiveManager{
		hotStore:  hotStore,
		coldStore: coldStore,
		config:    config,
		stopCh:    make(chan struct{}),
	}

	hybrid := &HybridEventStore{
		hotStore:       hotStore,
		coldStore:      coldStore,
		archiveManager: archiveManager,
		config:         config,
		metrics:        &hybridMetrics{},
	}

	// 자동 아카이빙 활성화
	if config.EnableAutoArchive {
		go hybrid.startAutoArchiving()
	}

	return hybrid
}

// Save는 이벤트들을 저장합니다 (항상 Hot 스토어에 저장)
func (h *HybridEventStore) Save(ctx context.Context, events []Event, expectedVersion int) error {
	if len(events) == 0 {
		return nil
	}

	start := time.Now()
	defer func() {
		h.updateMetrics("save", time.Since(start), "hot", nil)
	}()

	// 모든 새 이벤트는 Hot 스토어에 저장
	err := h.hotStore.Save(ctx, events, expectedVersion)
	if err != nil {
		h.updateMetrics("save", time.Since(start), "hot", err)
		return fmt.Errorf("failed to save to hot store: %w", err)
	}

	// Hot 스토어 크기 확인 및 아카이빙 트리거
	if h.config.EnableAutoArchive {
		go h.checkAndTriggerArchive()
	}

	return nil
}

// Load는 집합체의 모든 이벤트를 로드합니다 (Hot + Cold 통합)
func (h *HybridEventStore) Load(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
	start := time.Now()
	defer func() {
		h.updateMetrics("load", time.Since(start), "hybrid", nil)
	}()

	// Hot 스토어에서 최근 이벤트 로드
	hotEvents, err := h.hotStore.Load(ctx, aggregateID)
	if err != nil {
		h.updateMetrics("load", time.Since(start), "hot", err)
		return nil, fmt.Errorf("failed to load from hot store: %w", err)
	}

	// Cold 스토어에서 과거 이벤트 로드
	coldEvents, err := h.loadFromColdStore(ctx, aggregateID)
	if err != nil {
		h.updateMetrics("load", time.Since(start), "cold", err)
		return nil, fmt.Errorf("failed to load from cold store: %w", err)
	}

	// 이벤트 통합 및 정렬
	allEvents := h.mergeEvents(coldEvents, hotEvents)

	return allEvents, nil
}

// LoadFrom은 특정 버전부터 이벤트를 로드합니다
func (h *HybridEventStore) LoadFrom(ctx context.Context, aggregateID uuid.UUID, fromVersion int) ([]Event, error) {
	allEvents, err := h.Load(ctx, aggregateID)
	if err != nil {
		return nil, err
	}

	// 버전 필터링
	var filteredEvents []Event
	for _, event := range allEvents {
		if event.Version() >= fromVersion {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// FindEvents는 복잡한 쿼리로 이벤트를 찾습니다 (주로 Cold 스토어 활용)
func (h *HybridEventStore) FindEvents(ctx context.Context, query EventQuery) ([]Event, error) {
	start := time.Now()
	defer func() {
		h.updateMetrics("query", time.Since(start), "cold", nil)
	}()

	// Cold 스토어에서 검색 (복잡한 쿼리 지원)
	return h.coldStore.FindEvents(ctx, query)
}

// CountEvents는 쿼리 조건에 맞는 이벤트 수를 반환합니다
func (h *HybridEventStore) CountEvents(ctx context.Context, query EventQuery) (int64, error) {
	return h.coldStore.CountEvents(ctx, query)
}

// GetMetrics는 성능 메트릭을 반환합니다
func (h *HybridEventStore) GetMetrics() StoreMetrics {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()

	hotMetrics := h.hotStore.GetMetrics()
	coldMetrics := h.coldStore.GetMetrics()

	return StoreMetrics{
		SaveOperations:  hotMetrics.SaveOperations + coldMetrics.SaveOperations,
		LoadOperations:  hotMetrics.LoadOperations + coldMetrics.LoadOperations,
		AverageSaveTime: (hotMetrics.AverageSaveTime + coldMetrics.AverageSaveTime) / 2,
		AverageLoadTime: (hotMetrics.AverageLoadTime + coldMetrics.AverageLoadTime) / 2,
		ErrorCount:      h.metrics.errors,
		LastOperation:   h.metrics.lastOperation,
		StorageStrategy: StrategyHybrid,
	}
}

// Close는 연결을 정리합니다
func (h *HybridEventStore) Close() error {
	// 자동 아카이빙 중지
	h.archiveManager.Stop()

	// 각 스토어 정리
	if err := h.hotStore.Close(); err != nil {
		return err
	}
	if err := h.coldStore.Close(); err != nil {
		return err
	}

	return nil
}

// Archive Operations

// StartAutoArchiving은 자동 아카이빙을 시작합니다
func (h *HybridEventStore) startAutoArchiving() {
	h.archiveManager.Start(context.Background())
}

// ArchiveOldEvents는 오래된 이벤트를 Cold 스토어로 이동합니다
func (h *HybridEventStore) ArchiveOldEvents(ctx context.Context) error {
	return h.archiveManager.ArchiveOldEvents(ctx, time.Now().Add(-h.config.HotDataThreshold))
}

// Private helper methods

func (h *HybridEventStore) loadFromColdStore(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
	// Cold 스토어에서 집합체 이벤트 검색
	query := EventQuery{
		AggregateIDs: []uuid.UUID{aggregateID},
	}

	return h.coldStore.FindEvents(ctx, query)
}

func (h *HybridEventStore) mergeEvents(coldEvents, hotEvents []Event) []Event {
	// 두 이벤트 슬라이스를 버전 순으로 병합
	allEvents := append(coldEvents, hotEvents...)

	// 버전으로 정렬
	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Version() < allEvents[j].Version()
	})

	return allEvents
}

func (h *HybridEventStore) checkAndTriggerArchive() {
	h.metrics.mu.RLock()
	lastArchive := h.metrics.lastArchive
	h.metrics.mu.RUnlock()

	// 아카이빙 간격 확인
	if time.Since(lastArchive) < h.config.ArchiveInterval {
		return
	}

	// 비동기 아카이빙 실행
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
		defer cancel()

		if err := h.ArchiveOldEvents(ctx); err != nil {
			fmt.Printf("Archive operation failed: %v\n", err)
		}
	}()
}

func (h *HybridEventStore) updateMetrics(operation string, duration time.Duration, store string, err error) {
	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	switch store {
	case "hot":
		h.metrics.hotOperations++
	case "cold":
		h.metrics.coldOperations++
	case "hybrid":
		// 통합 연산
	}

	if operation == "save" {
		h.metrics.totalSaveTime += duration
	} else {
		h.metrics.totalLoadTime += duration
	}

	if err != nil {
		h.metrics.errors++
	}

	h.metrics.lastOperation = time.Now()
}

// ArchiveManager 구현

// Start는 아카이브 매니저를 시작합니다
func (am *ArchiveManager) Start(ctx context.Context) {
	am.mu.Lock()
	if am.running {
		am.mu.Unlock()
		return
	}
	am.running = true
	am.mu.Unlock()

	ticker := time.NewTicker(am.config.ArchiveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-am.stopCh:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-am.config.HotDataThreshold)
			if err := am.ArchiveOldEvents(ctx, cutoff); err != nil {
				fmt.Printf("Archive operation failed: %v\n", err)
			}
		}
	}
}

// Stop은 아카이브 매니저를 중지합니다
func (am *ArchiveManager) Stop() {
	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.running {
		return
	}

	am.running = false
	close(am.stopCh)
}

// ArchiveOldEvents는 지정된 시간보다 오래된 이벤트를 아카이빙합니다
func (am *ArchiveManager) ArchiveOldEvents(ctx context.Context, cutoff time.Time) error {
	// Hot 스토어에서 QueryableEventStore 인터페이스가 필요한 경우
	// 여기서는 단순화된 버전으로 구현

	fmt.Printf("Archiving events older than %s\n", cutoff.Format(time.RFC3339))

	// 실제 구현에서는:
	// 1. Hot 스토어에서 cutoff 이전 이벤트 찾기
	// 2. Cold 스토어로 이동
	// 3. Hot 스토어에서 제거
	// 4. 트랜잭션으로 일관성 보장

	return nil
}
