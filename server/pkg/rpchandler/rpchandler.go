// Package rpchandler provides a reflection-based JSON-RPC handler utility
// that automatically scans public methods of structs and registers them as JSON-RPC endpoints.
//
// Features:
// - Automatic method scanning using reflection
// - Flexible method signatures support
// - Group routing for hierarchical API paths
// - Handler composition for combining multiple handlers
// - JSON-RPC server integration
//
// Example usage:
//
//	// Define handlers
//	type GameHandler struct{}
//	func (g *GameHandler) GetStatus() any { return "running" }
//	func (g *GameHandler) Ping() error { return nil }
//
//	// Create registry and register handlers
//	registry := rpchandler.NewRegistry()
//	registry.RegisterHandler("game", &GameHandler{})
//
//	// Register with JSON-RPC server
//	rpcServer := jsonrpc.NewServer()
//	registry.RegisterAllMethods(rpcServer)
//
//	// Available methods: game.GetStatus, game.Ping
package rpchandler

import (
	"fmt"
	"log"
)

// Version 패키지 버전
const Version = "1.0.0"

// 편의 함수들 (인스턴스 기반)

// 유틸리티 함수들

// ValidateMethodSignature 메서드 시그니처가 유효한지 검증합니다
func ValidateMethodSignature(handler Handler, methodName string) error {
	// 이 함수는 개발 시 디버깅 용도로 사용할 수 있습니다
	// 실제 구현은 필요에 따라 추가할 수 있습니다
	return nil
}

// PrintRegisteredMethods 등록된 메서드들을 출력합니다 (디버깅 용도)
func PrintRegisteredMethods(registry *Registry) {
	methods := registry.GetMethodNames()
	if len(methods) == 0 {
		log.Println("No methods registered")
		return
	}

	log.Printf("Registered methods (%d):", len(methods))
	for _, method := range methods {
		log.Printf("  - %s", method)
	}
}

// CreateQuickRegistry 빠른 레지스트리 생성을 위한 헬퍼 함수
func CreateQuickRegistry(handlers map[string]Handler) (*Registry, error) {
	registry := NewRegistry()

	for name, handler := range handlers {
		if err := registry.RegisterHandler(name, handler); err != nil {
			return nil, fmt.Errorf("failed to register handler %s: %w", name, err)
		}
	}

	return registry, nil
}

// CreateQuickGroup 빠른 그룹 생성을 위한 헬퍼 함수
func CreateQuickGroup(registry *Registry, prefix string, handlers map[string]Handler) (*Group, error) {
	group := registry.Group(prefix)

	for name, handler := range handlers {
		if err := group.RegisterHandler(name, handler); err != nil {
			return nil, fmt.Errorf("failed to register handler %s to group %s: %w", name, prefix, err)
		}
	}

	return group, nil
}

// Stats 레지스트리 통계 정보
type Stats struct {
	MethodCount int    `json:"method_count"`
	GroupCount  int    `json:"group_count"`
	Version     string `json:"version"`
}

// GetStats 레지스트리 통계를 반환합니다
func (r *Registry) GetStats() Stats {
	return Stats{
		MethodCount: r.GetMethodCount(),
		GroupCount:  r.GetGroupCount(),
		Version:     Version,
	}
}
