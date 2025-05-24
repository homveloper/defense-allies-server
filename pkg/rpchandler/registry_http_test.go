package rpchandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRegistryHTTPHandler Registry의 HTTP 핸들러 기능 테스트
func TestRegistryHTTPHandler(t *testing.T) {
	// Registry 생성 및 핸들러 등록
	registry := NewRegistry()
	gameHandler := &GameHandler{}
	registry.RegisterHandler("game", gameHandler)
	registry.RegisterHandler("player", &PlayerHandler{})

	// HTTP 서버 생성 (Registry를 직접 핸들러로 사용)
	server := httptest.NewServer(registry)
	defer server.Close()

	t.Run("ValidJSONRPCCall", func(t *testing.T) {
		// 유효한 JSON-RPC 요청
		request := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "game.GetStatus",
			ID:      1,
		}

		jsonData, _ := json.Marshal(request)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		t.Logf("Response: %+v", response)

		if response.JSONRPC != "2.0" {
			t.Error("Expected JSON-RPC version 2.0")
		}

		if response.Error != nil {
			t.Errorf("Expected no error, got: %+v", response.Error)
		}

		if response.Result == nil {
			t.Error("Expected result, got nil")
		}
	})

	t.Run("JSONRPCWithParams", func(t *testing.T) {
		// 파라미터가 있는 JSON-RPC 요청
		params := map[string]string{"game_id": "12345"}
		paramsJSON, _ := json.Marshal(params)

		request := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "game.GetState",
			Params:  json.RawMessage(paramsJSON),
			ID:      2,
		}

		jsonData, _ := json.Marshal(request)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var response JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		t.Logf("Response with params: %+v", response)

		if response.Error != nil {
			t.Errorf("Expected no error, got: %+v", response.Error)
		}
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		// 존재하지 않는 메서드 호출
		request := JSONRPCRequest{
			JSONRPC: "2.0",
			Method:  "nonexistent.Method",
			ID:      3,
		}

		jsonData, _ := json.Marshal(request)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var response JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		t.Logf("Error response: %+v", response)

		if response.Error == nil {
			t.Error("Expected error for nonexistent method")
		}

		if response.Error.Code != -32603 {
			t.Errorf("Expected error code -32603, got %d", response.Error.Code)
		}
	})

	t.Run("InvalidJSONRPCVersion", func(t *testing.T) {
		// 잘못된 JSON-RPC 버전
		request := map[string]interface{}{
			"jsonrpc": "1.0",
			"method":  "game.GetStatus",
			"id":      4,
		}

		jsonData, _ := json.Marshal(request)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		var response JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Error == nil {
			t.Error("Expected error for invalid JSON-RPC version")
		}

		if response.Error.Code != -32600 {
			t.Errorf("Expected error code -32600, got %d", response.Error.Code)
		}
	})

	t.Run("InvalidHTTPMethod", func(t *testing.T) {
		// GET 메서드로 요청 (POST만 허용)
		resp, err := http.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", resp.StatusCode)
		}
	})

	t.Run("CORSOptions", func(t *testing.T) {
		// OPTIONS 요청 (CORS preflight)
		req, _ := http.NewRequest("OPTIONS", server.URL, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", resp.StatusCode)
		}

		// CORS 헤더 확인
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Expected CORS header Access-Control-Allow-Origin: *")
		}
	})
}

// TestRegistryAsHTTPHandler Registry를 HTTP Mux에 등록하는 테스트
func TestRegistryAsHTTPHandler(t *testing.T) {
	// Registry 생성
	registry := NewRegistry()
	registry.RegisterHandler("game", &GameHandler{})

	// HTTP Mux에 Registry 등록
	mux := http.NewServeMux()
	mux.Handle("/rpc", registry)            // 단일 경로
	mux.Handle("/api/v1/jsonrpc", registry) // API 경로

	// 테스트 서버 생성
	server := httptest.NewServer(mux)
	defer server.Close()

	// 두 경로 모두에서 동일하게 작동하는지 테스트
	paths := []string{"/rpc", "/api/v1/jsonrpc"}

	for _, path := range paths {
		t.Run("Path_"+path, func(t *testing.T) {
			request := JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "game.GetStatus",
				ID:      1,
			}

			jsonData, _ := json.Marshal(request)
			resp, err := http.Post(server.URL+path, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", path, resp.StatusCode)
			}

			var response JSONRPCResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response from %s: %v", path, err)
			}

			if response.Error != nil {
				t.Errorf("Expected no error from %s, got: %+v", path, response.Error)
			}

			t.Logf("Response from %s: %+v", path, response)
		})
	}
}
