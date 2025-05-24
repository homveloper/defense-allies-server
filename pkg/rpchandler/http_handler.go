package rpchandler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPHandler HTTP REST API를 통해 RPC를 호출하는 핸들러
type HTTPHandler struct {
	registry *Registry
	prefix   string // API 경로 프리픽스 (예: "/api/v1/rpc")
}

// NewHTTPHandler 새로운 HTTP 핸들러를 생성합니다
func NewHTTPHandler(registry *Registry, prefix string) *HTTPHandler {
	if prefix == "" {
		prefix = "/api/v1/rpc"
	}
	
	return &HTTPHandler{
		registry: registry,
		prefix:   prefix,
	}
}

// RegisterRoutes HTTP Mux에 라우트를 등록합니다
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	// 전체 RPC 호출 엔드포인트
	mux.HandleFunc(h.prefix+"/call", h.handleRPCCall)
	
	// 메서드별 개별 엔드포인트
	mux.HandleFunc(h.prefix+"/method/", h.handleMethodCall)
	
	// 메타데이터 엔드포인트
	mux.HandleFunc(h.prefix+"/methods", h.handleListMethods)
	mux.HandleFunc(h.prefix+"/info/", h.handleMethodInfo)
}

// RPCRequest RPC 호출 요청 구조체
type RPCRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// RPCResponse RPC 호출 응답 구조체
type RPCResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
	Method  string      `json:"method"`
}

// handleRPCCall 전체 RPC 호출을 처리합니다
// POST /api/v1/rpc/call
func (h *HTTPHandler) handleRPCCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Only POST method allowed")
		return
	}
	
	// 요청 본문 읽기
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()
	
	// JSON 파싱
	var req RPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	
	// 메서드 호출
	result, err := h.registry.CallMethod(r.Context(), req.Method, req.Params)
	
	// 응답 생성
	response := RPCResponse{
		Method: req.Method,
	}
	
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.Result = result
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

// handleMethodCall 개별 메서드 호출을 처리합니다
// POST /api/v1/rpc/method/{methodName}
func (h *HTTPHandler) handleMethodCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, http.StatusMethodNotAllowed, "Only POST method allowed")
		return
	}
	
	// URL에서 메서드 이름 추출
	path := strings.TrimPrefix(r.URL.Path, h.prefix+"/method/")
	methodName := strings.TrimSuffix(path, "/")
	
	if methodName == "" {
		h.sendError(w, http.StatusBadRequest, "Method name is required")
		return
	}
	
	// 요청 본문 읽기 (파라미터)
	var params json.RawMessage
	if r.ContentLength > 0 {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "Failed to read request body")
			return
		}
		defer r.Body.Close()
		params = json.RawMessage(body)
	}
	
	// 메서드 호출
	result, err := h.registry.CallMethod(r.Context(), methodName, params)
	
	// 응답 생성
	response := RPCResponse{
		Method: methodName,
	}
	
	if err != nil {
		response.Success = false
		response.Error = err.Error()
	} else {
		response.Success = true
		response.Result = result
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

// handleListMethods 등록된 메서드 목록을 반환합니다
// GET /api/v1/rpc/methods
func (h *HTTPHandler) handleListMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Only GET method allowed")
		return
	}
	
	methods := h.registry.GetMethodNames()
	signatures := h.registry.GetMethodSignatures()
	
	response := map[string]interface{}{
		"success": true,
		"methods": methods,
		"signatures": signatures,
		"count": len(methods),
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

// handleMethodInfo 특정 메서드 정보를 반환합니다
// GET /api/v1/rpc/info/{methodName}
func (h *HTTPHandler) handleMethodInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "Only GET method allowed")
		return
	}
	
	// URL에서 메서드 이름 추출
	path := strings.TrimPrefix(r.URL.Path, h.prefix+"/info/")
	methodName := strings.TrimSuffix(path, "/")
	
	if methodName == "" {
		h.sendError(w, http.StatusBadRequest, "Method name is required")
		return
	}
	
	// 메서드 정보 조회
	info, err := h.registry.GetMethodInfo(methodName)
	if err != nil {
		h.sendError(w, http.StatusNotFound, fmt.Sprintf("Method not found: %s", methodName))
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"method": methodName,
		"info": info,
	}
	
	h.sendJSON(w, http.StatusOK, response)
}

// sendJSON JSON 응답을 전송합니다
func (h *HTTPHandler) sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// sendError 에러 응답을 전송합니다
func (h *HTTPHandler) sendError(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]interface{}{
		"success": false,
		"error": message,
	}
	h.sendJSON(w, statusCode, response)
}

// GetPrefix API 경로 프리픽스를 반환합니다
func (h *HTTPHandler) GetPrefix() string {
	return h.prefix
}

// SetPrefix API 경로 프리픽스를 설정합니다
func (h *HTTPHandler) SetPrefix(prefix string) {
	h.prefix = prefix
}

// GetRegistry 등록된 Registry를 반환합니다
func (h *HTTPHandler) GetRegistry() *Registry {
	return h.registry
}

// HandleRPCCallDirect 직접 RPC 호출을 처리합니다 (미들웨어 등에서 사용)
func (h *HTTPHandler) HandleRPCCallDirect(ctx context.Context, methodName string, params json.RawMessage) (interface{}, error) {
	return h.registry.CallMethod(ctx, methodName, params)
}
