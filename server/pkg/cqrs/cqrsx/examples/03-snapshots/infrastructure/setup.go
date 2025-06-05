package infrastructure

import (
	"time"

	"defense-allies-server/pkg/cqrs/cqrsx"
	"defense-allies-server/pkg/cqrs/cqrsx/examples/03-snapshots/snapshots"
)

// Infrastructure 인프라 구성 요소들
type Infrastructure struct {
	MongoClient     *cqrsx.MongoClientManager
	EventStore      cqrsx.EventStore
	SnapshotStore   snapshots.SnapshotStore
	SnapshotManager snapshots.SnapshotManager
	OrderRepo       *OrderRepository
}

// SetupInfrastructure 인프라 설정
func SetupInfrastructure(config *InfraConfig) (*Infrastructure, error) {
	// MongoDB 클라이언트 설정
	mongoClient, err := cqrsx.NewMongoClientManager(&cqrsx.MongoConfig{
		URI:                    config.MongoURI,
		Database:               config.Database,
		ConnectTimeout:         config.ConnectTimeout,
		ServerSelectionTimeout: config.ServerSelectionTimeout,
		MaxPoolSize:            int(config.MaxPoolSize),
	})
	if err != nil {
		return nil, err
	}

	// 이벤트 스토어 설정
	eventStore := cqrsx.NewMongoEventStore(mongoClient, config.EventsCollection)

	// 스냅샷 스토어 설정
	snapshotStore := snapshots.NewMongoSnapshotStore(mongoClient, config.SnapshotsCollection)

	// 스냅샷 직렬화기 설정
	serializer, err := createSerializer(config.SnapshotConfig)
	if err != nil {
		return nil, err
	}

	// 스냅샷 정책 설정
	policy := createPolicy(config.SnapshotConfig)

	// 스냅샷 매니저 설정
	snapshotManager := snapshots.NewDefaultSnapshotManager(
		snapshotStore,
		serializer,
		policy,
		config.SnapshotConfig,
	)

	// 주문 리포지토리 설정
	orderRepo := NewOrderRepository(eventStore, snapshotManager)

	return &Infrastructure{
		MongoClient:     mongoClient,
		EventStore:      eventStore,
		SnapshotStore:   snapshotStore,
		SnapshotManager: snapshotManager,
		OrderRepo:       orderRepo,
	}, nil
}

// InfraConfig 인프라 설정
type InfraConfig struct {
	// MongoDB 설정
	MongoURI               string
	Database               string
	EventsCollection       string
	SnapshotsCollection    string
	ConnectTimeout         time.Duration
	ServerSelectionTimeout time.Duration
	MaxPoolSize            uint64

	// 스냅샷 설정
	SnapshotConfig *snapshots.SnapshotConfiguration
}

// DefaultInfraConfig 기본 인프라 설정
func DefaultInfraConfig() *InfraConfig {
	return &InfraConfig{
		MongoURI:               "mongodb://localhost:27017",
		Database:               "cqrs_snapshots_demo",
		EventsCollection:       "events",
		SnapshotsCollection:    "snapshots",
		ConnectTimeout:         10 * time.Second,
		ServerSelectionTimeout: 5 * time.Second,
		MaxPoolSize:            10,
		SnapshotConfig: &snapshots.SnapshotConfiguration{
			Enabled:                  true,
			DefaultPolicy:            "event_count",
			DefaultSerializer:        "json",
			DefaultCompression:       "gzip",
			EventCountThreshold:      5,
			TimeIntervalMinutes:      60,
			SizeThresholdBytes:       1024 * 1024, // 1MB
			MaxSnapshotsPerAggregate: 3,
			CleanupIntervalHours:     24,
			RetentionDays:            30,
			AsyncCreation:            false,
			BatchSize:                100,
			CompressionLevel:         6,
			EnableMetrics:            true,
			MetricsInterval:          300, // 5분
			AlertThresholds: map[string]float64{
				"max_restore_time_ms": 1000,
				"max_snapshot_size":   10 * 1024 * 1024, // 10MB
			},
		},
	}
}

// TestInfraConfig 테스트용 인프라 설정
func TestInfraConfig() *InfraConfig {
	config := DefaultInfraConfig()
	config.Database = "cqrs_snapshots_test"
	config.SnapshotConfig.EventCountThreshold = 3 // 더 자주 스냅샷 생성
	config.SnapshotConfig.MaxSnapshotsPerAggregate = 2
	return config
}

// PerformanceTestInfraConfig 성능 테스트용 인프라 설정
func PerformanceTestInfraConfig() *InfraConfig {
	config := DefaultInfraConfig()
	config.Database = "cqrs_snapshots_perf"
	config.SnapshotConfig.EventCountThreshold = 10
	config.SnapshotConfig.MaxSnapshotsPerAggregate = 5
	config.SnapshotConfig.AsyncCreation = true
	return config
}

// createSerializer 직렬화기 생성
func createSerializer(config *snapshots.SnapshotConfiguration) (snapshots.SnapshotSerializer, error) {
	factory := snapshots.NewSerializerFactory()

	options := map[string]interface{}{
		"pretty_print": false,
	}

	return factory.CreateSerializer(
		config.DefaultSerializer,
		config.DefaultCompression,
		options,
	)
}

// createPolicy 정책 생성
func createPolicy(config *snapshots.SnapshotConfiguration) snapshots.SnapshotPolicy {
	switch config.DefaultPolicy {
	case "event_count":
		return snapshots.NewEventCountPolicy(config.EventCountThreshold)
	case "time_based":
		interval := time.Duration(config.TimeIntervalMinutes) * time.Minute
		return snapshots.NewTimeBasedPolicy(interval)
	case "version_based":
		return snapshots.NewVersionBasedPolicy(config.EventCountThreshold)
	case "always":
		return snapshots.NewAlwaysPolicy()
	case "never":
		return snapshots.NewNeverPolicy()
	case "adaptive":
		return snapshots.NewAdaptivePolicy(config.EventCountThreshold, 0.8)
	default:
		// 기본값: 이벤트 개수 기반
		return snapshots.NewEventCountPolicy(config.EventCountThreshold)
	}
}

// CreateCompositePolicy 복합 정책 생성 예제
func CreateCompositePolicy(config *snapshots.SnapshotConfiguration) snapshots.SnapshotPolicy {
	// 이벤트 개수 정책과 시간 기반 정책을 OR로 결합
	eventCountPolicy := snapshots.NewEventCountPolicy(config.EventCountThreshold)
	timeBasedPolicy := snapshots.NewTimeBasedPolicy(time.Duration(config.TimeIntervalMinutes) * time.Minute)

	return snapshots.NewCompositePolicy("OR", eventCountPolicy, timeBasedPolicy)
}

// CreateCustomPolicy 커스텀 정책 생성 예제
func CreateCustomPolicy() snapshots.SnapshotPolicy {
	// 비즈니스 로직 기반 커스텀 정책
	return snapshots.NewCustomPolicy(
		"BusinessLogicPolicy",
		5,
		func(aggregate snapshots.Aggregate, eventCount int) bool {
			// 예: 주문이 확정되었을 때만 스냅샷 생성
			// 실제로는 aggregate를 Order로 캐스팅하여 상태 확인
			return eventCount >= 5 && eventCount%5 == 0
		},
	)
}

// Cleanup 인프라 정리
func (infra *Infrastructure) Cleanup() error {
	if infra.MongoClient != nil {
		return infra.MongoClient.Close(nil)
	}
	return nil
}

// GetStats 인프라 통계 조회
func (infra *Infrastructure) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 스냅샷 통계
	if infra.SnapshotStore != nil {
		snapshotStats, err := infra.SnapshotStore.GetSnapshotStats(nil)
		if err == nil {
			stats["snapshots"] = snapshotStats
		}
	}

	// 추가 통계들...
	stats["timestamp"] = time.Now()

	return stats, nil
}

// ConfigPresets 설정 프리셋들
var ConfigPresets = map[string]*snapshots.SnapshotConfiguration{
	"development": {
		Enabled:                  true,
		DefaultPolicy:            "event_count",
		DefaultSerializer:        "json",
		DefaultCompression:       "none",
		EventCountThreshold:      3,
		MaxSnapshotsPerAggregate: 2,
		AsyncCreation:            false,
		EnableMetrics:            true,
	},
	"production": {
		Enabled:                  true,
		DefaultPolicy:            "adaptive",
		DefaultSerializer:        "bson",
		DefaultCompression:       "gzip",
		EventCountThreshold:      10,
		MaxSnapshotsPerAggregate: 5,
		AsyncCreation:            true,
		EnableMetrics:            true,
		RetentionDays:            90,
	},
	"high_performance": {
		Enabled:                  true,
		DefaultPolicy:            "event_count",
		DefaultSerializer:        "bson",
		DefaultCompression:       "gzip",
		EventCountThreshold:      5,
		MaxSnapshotsPerAggregate: 10,
		AsyncCreation:            true,
		BatchSize:                1000,
		EnableMetrics:            false, // 성능 우선
	},
	"testing": {
		Enabled:                  true,
		DefaultPolicy:            "always",
		DefaultSerializer:        "json",
		DefaultCompression:       "none",
		EventCountThreshold:      1,
		MaxSnapshotsPerAggregate: 1,
		AsyncCreation:            false,
		EnableMetrics:            false,
	},
}
