// versioning/schema_migration.go - 스키마 마이그레이션
package cqrsx

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// SchemaMigrator는 이벤트 스키마 마이그레이션을 담당합니다
type SchemaMigrator struct {
	collection      *mongo.Collection
	upgrader        *EventUpgrader
	migrationStatus *MigrationStatus
}

// MigrationStatus는 마이그레이션 진행 상황을 추적합니다
type MigrationStatus struct {
	TotalEvents     int64      `json:"totalEvents"`
	ProcessedEvents int64      `json:"processedEvents"`
	UpgradedEvents  int64      `json:"upgradedEvents"`
	ErrorCount      int64      `json:"errorCount"`
	StartTime       time.Time  `json:"startTime"`
	EndTime         *time.Time `json:"endTime,omitempty"`
	Status          string     `json:"status"`
}

// NewSchemaMigrator는 새로운 스키마 마이그레이터를 생성합니다
func NewSchemaMigrator(collection *mongo.Collection, upgrader *EventUpgrader) *SchemaMigrator {
	return &SchemaMigrator{
		collection: collection,
		upgrader:   upgrader,
		migrationStatus: &MigrationStatus{
			Status: "ready",
		},
	}
}

// MigrateAllEvents는 모든 이벤트를 최신 스키마로 마이그레이션합니다
func (sm *SchemaMigrator) MigrateAllEvents(ctx context.Context, dryRun bool) error {
	log.Printf("Starting schema migration (dry run: %v)", dryRun)

	sm.migrationStatus.StartTime = time.Now()
	sm.migrationStatus.Status = "running"

	// 전체 이벤트 수 계산
	totalCount, err := sm.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count documents: %w", err)
	}
	sm.migrationStatus.TotalEvents = totalCount

	// 배치 단위로 처리
	batchSize := int64(1000)
	processedCount := int64(0)

	for processedCount < totalCount {
		// 배치 조회
		pipeline := []bson.M{
			{"$skip": processedCount},
			{"$limit": batchSize},
		}

		cursor, err := sm.collection.Aggregate(ctx, pipeline)
		if err != nil {
			return fmt.Errorf("failed to aggregate documents: %w", err)
		}

		// 배치 처리
		batchUpgraded, batchErrors := sm.processBatch(ctx, cursor, dryRun)
		cursor.Close(ctx)

		// 진행 상황 업데이트
		sm.migrationStatus.ProcessedEvents += batchSize
		sm.migrationStatus.UpgradedEvents += batchUpgraded
		sm.migrationStatus.ErrorCount += batchErrors

		processedCount += batchSize

		// 진행 상황 로그
		progress := float64(processedCount) / float64(totalCount) * 100
		log.Printf("Migration progress: %.2f%% (%d/%d events)", progress, processedCount, totalCount)
	}

	// 마이그레이션 완료
	now := time.Now()
	sm.migrationStatus.EndTime = &now
	sm.migrationStatus.Status = "completed"

	log.Printf("Schema migration completed: %d events processed, %d upgraded, %d errors",
		sm.migrationStatus.ProcessedEvents,
		sm.migrationStatus.UpgradedEvents,
		sm.migrationStatus.ErrorCount)

	return nil
}

func (sm *SchemaMigrator) processBatch(ctx context.Context, cursor *mongo.Cursor, dryRun bool) (int64, int64) {
	var upgraded int64
	var errors int64

	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			errors++
			continue
		}

		// 이벤트 스트림인지 단일 이벤트인지 확인
		if events, ok := doc["events"].(bson.A); ok {
			// 이벤트 스트림 처리
			upgradedEvents, errorEvents := sm.processEventStream(ctx, doc, events, dryRun)
			upgraded += upgradedEvents
			errors += errorEvents
		} else {
			// 단일 이벤트 처리
			if sm.processEvent(ctx, doc, dryRun) {
				upgraded++
			} else {
				errors++
			}
		}
	}

	return upgraded, errors
}

func (sm *SchemaMigrator) processEventStream(ctx context.Context, streamDoc bson.M, events bson.A, dryRun bool) (int64, int64) {
	var upgraded int64
	var errors int64
	var needsUpdate bool

	for i, eventItem := range events {
		eventDoc, ok := eventItem.(bson.M)
		if !ok {
			errors++
			continue
		}

		eventType, ok := eventDoc["eventType"].(string)
		if !ok {
			errors++
			continue
		}

		version, ok := eventDoc["version"].(int32)
		if !ok {
			// 버전이 없는 경우 v1으로 간주
			version = 1
		}

		data, ok := eventDoc["data"].(bson.M)
		if !ok {
			errors++
			continue
		}

		// 이벤트 업그레이드 시도
		upgradedData, newVersion, err := sm.upgrader.UpgradeEvent(eventType, int(version), convertBsonMToMap(data))
		if err != nil {
			errors++
			continue
		}

		// 버전이 변경된 경우
		if newVersion > int(version) {
			eventDoc["data"] = upgradedData
			eventDoc["version"] = newVersion
			events[i] = eventDoc
			needsUpdate = true
			upgraded++
		}
	}

	// 실제 업데이트 수행 (dry run이 아닌 경우)
	if needsUpdate && !dryRun {
		filter := bson.M{"_id": streamDoc["_id"]}
		update := bson.M{"$set": bson.M{"events": events}}

		if _, err := sm.collection.UpdateOne(ctx, filter, update); err != nil {
			log.Printf("Failed to update stream %v: %v", streamDoc["_id"], err)
			errors += upgraded // 전체를 에러로 카운트
			upgraded = 0
		}
	}

	return upgraded, errors
}

func (sm *SchemaMigrator) processEvent(ctx context.Context, eventDoc bson.M, dryRun bool) bool {
	eventType, ok := eventDoc["eventType"].(string)
	if !ok {
		return false
	}

	version, ok := eventDoc["version"].(int32)
	if !ok {
		version = 1
	}

	data, ok := eventDoc["data"].(bson.M)
	if !ok {
		return false
	}

	// 이벤트 업그레이드 시도
	upgradedData, newVersion, err := sm.upgrader.UpgradeEvent(eventType, int(version), convertBsonMToMap(data))
	if err != nil {
		return false
	}

	// 버전이 변경된 경우
	if newVersion > int(version) && !dryRun {
		filter := bson.M{"_id": eventDoc["_id"]}
		update := bson.M{
			"$set": bson.M{
				"data":    upgradedData,
				"version": newVersion,
			},
		}

		if _, err := sm.collection.UpdateOne(ctx, filter, update); err != nil {
			log.Printf("Failed to update event %v: %v", eventDoc["_id"], err)
			return false
		}

		return true
	}

	return false
}

// GetMigrationStatus는 현재 마이그레이션 상태를 반환합니다
func (sm *SchemaMigrator) GetMigrationStatus() *MigrationStatus {
	return sm.migrationStatus
}

// 유틸리티 함수
func convertBsonMToMap(bsonM bson.M) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range bsonM {
		result[k] = v
	}
	return result
}
