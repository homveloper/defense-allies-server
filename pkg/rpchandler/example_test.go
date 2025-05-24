package rpchandler

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// 테스트용 핸들러들

// GameHandler 게임 관련 핸들러
type GameHandler struct{}

// GetStatus 게임 상태 조회 (파라미터 없음)
func (g *GameHandler) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":  "running",
		"players": 42,
		"uptime":  3600,
	}
}

// Ping 핑 테스트 (에러만 반환)
func (g *GameHandler) Ping() error {
	return nil
}

// GetStateParams 게임 상태 조회 파라미터
type GetStateParams struct {
	GameID string `json:"game_id"`
}

// GetState 게임 상태 조회 (구조체 파라미터)
func (g *GameHandler) GetState(ctx context.Context, params GetStateParams) (map[string]interface{}, error) {
	if params.GameID == "" {
		return nil, fmt.Errorf("game_id is required")
	}
	
	return map[string]interface{}{
		"game_id": params.GameID,
		"status":  "active",
		"wave":    5,
		"lives":   10,
	}, nil
}

// ProcessRawData Raw JSON 처리
func (g *GameHandler) ProcessRawData(ctx context.Context, params json.RawMessage) error {
	// Raw JSON 데이터 처리
	fmt.Printf("Processing raw data: %s\n", string(params))
	return nil
}

// PlayerHandler 플레이어 관련 핸들러
type PlayerHandler struct{}

// GetProfileParams 프로필 조회 파라미터
type GetProfileParams struct {
	PlayerID string `json:"player_id"`
}

// GetProfile 플레이어 프로필 조회
func (p *PlayerHandler) GetProfile(ctx context.Context, params GetProfileParams) (map[string]interface{}, error) {
	return map[string]interface{}{
		"id":    params.PlayerID,
		"name":  "Player" + params.PlayerID,
		"level": 25,
		"rank":  "Gold",
	}, nil
}

// TestRPCContext RPCContext 기본 테스트
func TestRPCContext(t *testing.T) {
	// Registry 생성
	registry := NewRegistry()
	
	// 핸들러 등록
	err := registry.RegisterHandler("game", &GameHandler{})
	if err != nil {
		t.Fatalf("Failed to register GameHandler: %v", err)
	}
	
	err = registry.RegisterHandler("player", &PlayerHandler{})
	if err != nil {
		t.Fatalf("Failed to register PlayerHandler: %v", err)
	}
	
	// 등록된 메서드 확인
	methods := registry.GetMethodNames()
	expectedMethods := []string{
		"game.GetStatus",
		"game.Ping", 
		"game.GetState",
		"game.ProcessRawData",
		"player.GetProfile",
	}
	
	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(methods))
	}
	
	// RPCContext 확인
	for _, methodName := range expectedMethods {
		ctx, exists := registry.GetRPCContext(methodName)
		if !exists {
			t.Errorf("RPCContext not found for method: %s", methodName)
			continue
		}
		
		if !ctx.IsValid() {
			t.Errorf("Invalid RPCContext for method: %s", methodName)
		}
		
		t.Logf("Method: %s, Signature: %s", methodName, ctx.GetMethodSignature())
	}
}

// TestMethodCalls 메서드 호출 테스트
func TestMethodCalls(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterHandler("game", &GameHandler{})
	registry.RegisterHandler("player", &PlayerHandler{})
	
	ctx := context.Background()
	
	// 파라미터 없는 메서드 테스트
	result, err := registry.CallMethod(ctx, "game.GetStatus", nil)
	if err != nil {
		t.Errorf("Failed to call game.GetStatus: %v", err)
	} else {
		t.Logf("game.GetStatus result: %+v", result)
	}
	
	// 에러만 반환하는 메서드 테스트
	result, err = registry.CallMethod(ctx, "game.Ping", nil)
	if err != nil {
		t.Errorf("Failed to call game.Ping: %v", err)
	} else {
		t.Logf("game.Ping result: %+v", result)
	}
	
	// 구조체 파라미터 메서드 테스트
	params := json.RawMessage(`{"game_id": "12345"}`)
	result, err = registry.CallMethod(ctx, "game.GetState", params)
	if err != nil {
		t.Errorf("Failed to call game.GetState: %v", err)
	} else {
		t.Logf("game.GetState result: %+v", result)
	}
	
	// Raw JSON 파라미터 메서드 테스트
	rawParams := json.RawMessage(`{"action": "test", "data": [1,2,3]}`)
	result, err = registry.CallMethod(ctx, "game.ProcessRawData", rawParams)
	if err != nil {
		t.Errorf("Failed to call game.ProcessRawData: %v", err)
	} else {
		t.Logf("game.ProcessRawData result: %+v", result)
	}
	
	// 플레이어 프로필 조회 테스트
	playerParams := json.RawMessage(`{"player_id": "user123"}`)
	result, err = registry.CallMethod(ctx, "player.GetProfile", playerParams)
	if err != nil {
		t.Errorf("Failed to call player.GetProfile: %v", err)
	} else {
		t.Logf("player.GetProfile result: %+v", result)
	}
}

// TestMethodInfo 메서드 정보 테스트
func TestMethodInfo(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterHandler("game", &GameHandler{})
	
	// 메서드 정보 조회
	info, err := registry.GetMethodInfo("game.GetState")
	if err != nil {
		t.Errorf("Failed to get method info: %v", err)
	} else {
		t.Logf("Method info: %+v", info)
	}
	
	// 모든 메서드 정보 조회
	allInfo := registry.GetAllMethodInfo()
	for methodName, info := range allInfo {
		t.Logf("Method %s info: %+v", methodName, info)
	}
	
	// 메서드 시그니처 조회
	signatures := registry.GetMethodSignatures()
	for methodName, signature := range signatures {
		t.Logf("Method %s signature: %s", methodName, signature)
	}
}
