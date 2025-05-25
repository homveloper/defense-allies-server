package rpchandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHTTPHandler HTTP 핸들러 테스트
func TestHTTPHandler(t *testing.T) {
	// Registry 및 핸들러 설정
	registry := NewRegistry()
	registry.RegisterHandler("game", &GameHandler{})
	registry.RegisterHandler("player", &PlayerHandler{})
	
	// HTTP 핸들러 생성
	httpHandler := NewHTTPHandler(registry, "/api/v1/rpc")
	
	// HTTP Mux 설정
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	
	// 테스트 서버 생성
	server := httptest.NewServer(mux)
	defer server.Close()
	
	t.Run("ListMethods", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/rpc/methods")
		if err != nil {
			t.Fatalf("Failed to get methods: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		t.Logf("Methods response: %+v", result)
		
		if !result["success"].(bool) {
			t.Error("Expected success to be true")
		}
	})
	
	t.Run("CallMethodDirect", func(t *testing.T) {
		// 파라미터 없는 메서드 호출
		reqBody := RPCRequest{
			Method: "game.GetStatus",
		}
		
		jsonData, _ := json.Marshal(reqBody)
		resp, err := http.Post(server.URL+"/api/v1/rpc/call", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to call method: %v", err)
		}
		defer resp.Body.Close()
		
		var result RPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		t.Logf("RPC call response: %+v", result)
		
		if !result.Success {
			t.Errorf("Expected success, got error: %s", result.Error)
		}
	})
	
	t.Run("CallMethodWithParams", func(t *testing.T) {
		// 파라미터가 있는 메서드 호출
		params := map[string]string{"game_id": "12345"}
		paramsJSON, _ := json.Marshal(params)
		
		reqBody := RPCRequest{
			Method: "game.GetState",
			Params: json.RawMessage(paramsJSON),
		}
		
		jsonData, _ := json.Marshal(reqBody)
		resp, err := http.Post(server.URL+"/api/v1/rpc/call", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to call method: %v", err)
		}
		defer resp.Body.Close()
		
		var result RPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		t.Logf("RPC call with params response: %+v", result)
		
		if !result.Success {
			t.Errorf("Expected success, got error: %s", result.Error)
		}
	})
	
	t.Run("CallIndividualMethod", func(t *testing.T) {
		// 개별 메서드 엔드포인트 호출
		params := map[string]string{"player_id": "user123"}
		paramsJSON, _ := json.Marshal(params)
		
		resp, err := http.Post(server.URL+"/api/v1/rpc/method/player.GetProfile", "application/json", bytes.NewBuffer(paramsJSON))
		if err != nil {
			t.Fatalf("Failed to call individual method: %v", err)
		}
		defer resp.Body.Close()
		
		var result RPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		t.Logf("Individual method call response: %+v", result)
		
		if !result.Success {
			t.Errorf("Expected success, got error: %s", result.Error)
		}
	})
	
	t.Run("GetMethodInfo", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/rpc/info/game.GetState")
		if err != nil {
			t.Fatalf("Failed to get method info: %v", err)
		}
		defer resp.Body.Close()
		
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		t.Logf("Method info response: %+v", result)
		
		if !result["success"].(bool) {
			t.Error("Expected success to be true")
		}
	})
	
	t.Run("ErrorHandling", func(t *testing.T) {
		// 존재하지 않는 메서드 호출
		reqBody := RPCRequest{
			Method: "nonexistent.Method",
		}
		
		jsonData, _ := json.Marshal(reqBody)
		resp, err := http.Post(server.URL+"/api/v1/rpc/call", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to call method: %v", err)
		}
		defer resp.Body.Close()
		
		var result RPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		t.Logf("Error response: %+v", result)
		
		if result.Success {
			t.Error("Expected failure for nonexistent method")
		}
		
		if !strings.Contains(result.Error, "method not found") {
			t.Errorf("Expected 'method not found' error, got: %s", result.Error)
		}
	})
}

// TestHTTPHandlerCURL CURL 명령어 예제를 위한 테스트
func TestHTTPHandlerCURL(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterHandler("game", &GameHandler{})
	
	httpHandler := NewHTTPHandler(registry, "/api/v1/rpc")
	mux := http.NewServeMux()
	httpHandler.RegisterRoutes(mux)
	
	server := httptest.NewServer(mux)
	defer server.Close()
	
	t.Logf("Test server running at: %s", server.URL)
	t.Logf("CURL examples:")
	t.Logf("1. List methods:")
	t.Logf("   curl %s/api/v1/rpc/methods", server.URL)
	t.Logf("")
	t.Logf("2. Call method via /call endpoint:")
	t.Logf("   curl -X POST %s/api/v1/rpc/call \\", server.URL)
	t.Logf("     -H 'Content-Type: application/json' \\")
	t.Logf("     -d '{\"method\": \"game.GetStatus\"}'")
	t.Logf("")
	t.Logf("3. Call method with parameters:")
	t.Logf("   curl -X POST %s/api/v1/rpc/call \\", server.URL)
	t.Logf("     -H 'Content-Type: application/json' \\")
	t.Logf("     -d '{\"method\": \"game.GetState\", \"params\": {\"game_id\": \"12345\"}}'")
	t.Logf("")
	t.Logf("4. Call individual method endpoint:")
	t.Logf("   curl -X POST %s/api/v1/rpc/method/game.GetStatus", server.URL)
	t.Logf("")
	t.Logf("5. Get method info:")
	t.Logf("   curl %s/api/v1/rpc/info/game.GetState", server.URL)
}
