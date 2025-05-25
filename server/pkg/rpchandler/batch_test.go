package rpchandler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestBatchProcessing JSON-RPC 배치 처리 테스트
func TestBatchProcessing(t *testing.T) {
	// Registry 생성 및 핸들러 등록
	registry := NewRegistry()
	registry.RegisterHandler("game", &GameHandler{})
	registry.RegisterHandler("player", &PlayerHandler{})

	// HTTP 서버 생성
	server := httptest.NewServer(registry)
	defer server.Close()

	t.Run("SingleRequest", func(t *testing.T) {
		// 단일 요청 테스트
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

		var response JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		t.Logf("Single request response: %+v", response)

		if response.JSONRPC != "2.0" {
			t.Error("Expected JSON-RPC version 2.0")
		}

		if response.ID != 1 {
			t.Errorf("Expected ID 1, got %v", response.ID)
		}
	})

	t.Run("BatchRequest", func(t *testing.T) {
		// 배치 요청 테스트
		batchRequest := []JSONRPCRequest{
			{
				JSONRPC: "2.0",
				Method:  "game.GetStatus",
				ID:      1,
			},
			{
				JSONRPC: "2.0",
				Method:  "game.Ping",
				ID:      2,
			},
			{
				JSONRPC: "2.0",
				Method:  "game.GetState",
				Params:  json.RawMessage(`{"game_id": "12345"}`),
				ID:      3,
			},
		}

		jsonData, _ := json.Marshal(batchRequest)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make batch request: %v", err)
		}
		defer resp.Body.Close()

		var responses []JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		t.Logf("Batch request responses: %+v", responses)

		// 응답 개수 확인
		if len(responses) != len(batchRequest) {
			t.Errorf("Expected %d responses, got %d", len(batchRequest), len(responses))
		}

		// 순서 확인
		for i, response := range responses {
			expectedID := batchRequest[i].ID
			if response.ID != expectedID {
				t.Errorf("Response %d: expected ID %v, got %v", i, expectedID, response.ID)
			}

			if response.JSONRPC != "2.0" {
				t.Errorf("Response %d: expected JSON-RPC version 2.0", i)
			}
		}
	})

	t.Run("BatchRequestWithErrors", func(t *testing.T) {
		// 에러가 포함된 배치 요청 테스트
		batchRequest := []JSONRPCRequest{
			{
				JSONRPC: "2.0",
				Method:  "game.GetStatus",
				ID:      1,
			},
			{
				JSONRPC: "2.0",
				Method:  "nonexistent.Method",
				ID:      2,
			},
			{
				JSONRPC: "1.0", // 잘못된 버전
				Method:  "game.Ping",
				ID:      3,
			},
		}

		jsonData, _ := json.Marshal(batchRequest)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make batch request: %v", err)
		}
		defer resp.Body.Close()

		var responses []JSONRPCResponse
		if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
			t.Fatalf("Failed to decode batch response: %v", err)
		}

		t.Logf("Batch request with errors responses: %+v", responses)

		// 첫 번째 응답은 성공
		if responses[0].Error != nil {
			t.Errorf("Expected first response to succeed, got error: %+v", responses[0].Error)
		}

		// 두 번째 응답은 메서드 없음 에러
		if responses[1].Error == nil {
			t.Error("Expected second response to have error for nonexistent method")
		}

		// 세 번째 응답은 버전 에러
		if responses[2].Error == nil {
			t.Error("Expected third response to have error for invalid version")
		} else if responses[2].Error.Code != -32600 {
			t.Errorf("Expected error code -32600, got %d", responses[2].Error.Code)
		}
	})

	t.Run("EmptyBatchRequest", func(t *testing.T) {
		// 빈 배치 요청 테스트
		emptyBatch := []JSONRPCRequest{}

		jsonData, _ := json.Marshal(emptyBatch)
		resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make empty batch request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty batch, got %d", resp.StatusCode)
		}
	})
}

// OrderTestHandler 순서 테스트용 핸들러
type OrderTestHandler struct {
	callOrder []int
}

// Method1 첫 번째 메서드
func (h *OrderTestHandler) Method1() int {
	h.callOrder = append(h.callOrder, 1)
	time.Sleep(10 * time.Millisecond) // 약간의 지연
	return 1
}

// Method2 두 번째 메서드
func (h *OrderTestHandler) Method2() int {
	h.callOrder = append(h.callOrder, 2)
	time.Sleep(5 * time.Millisecond) // 더 짧은 지연
	return 2
}

// Method3 세 번째 메서드
func (h *OrderTestHandler) Method3() int {
	h.callOrder = append(h.callOrder, 3)
	return 3
}

// TestBatchOrderGuarantee 배치 처리 순서 보장 테스트
func TestBatchOrderGuarantee(t *testing.T) {
	// 순서를 확인할 수 있는 특별한 핸들러
	handler := &OrderTestHandler{
		callOrder: make([]int, 0),
	}

	// Registry에 등록
	registry := NewRegistry()
	registry.RegisterHandler("order", handler)

	server := httptest.NewServer(registry)
	defer server.Close()

	// 배치 요청 (순서: 1, 2, 3)
	batchRequest := []JSONRPCRequest{
		{JSONRPC: "2.0", Method: "order.Method1", ID: 1},
		{JSONRPC: "2.0", Method: "order.Method2", ID: 2},
		{JSONRPC: "2.0", Method: "order.Method3", ID: 3},
	}

	jsonData, _ := json.Marshal(batchRequest)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to make batch request: %v", err)
	}
	defer resp.Body.Close()

	var responses []JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		t.Fatalf("Failed to decode batch response: %v", err)
	}

	// 호출 순서 확인
	expectedOrder := []int{1, 2, 3}
	if len(handler.callOrder) != len(expectedOrder) {
		t.Errorf("Expected %d calls, got %d", len(expectedOrder), len(handler.callOrder))
	}

	for i, expected := range expectedOrder {
		if i >= len(handler.callOrder) || handler.callOrder[i] != expected {
			t.Errorf("Expected call order %v, got %v", expectedOrder, handler.callOrder)
			break
		}
	}

	// 응답 순서 확인
	for i, response := range responses {
		expectedID := i + 1
		if response.ID != expectedID {
			t.Errorf("Response %d: expected ID %d, got %v", i, expectedID, response.ID)
		}
	}

	t.Logf("Call order: %v", handler.callOrder)
	t.Logf("Responses: %+v", responses)
}
