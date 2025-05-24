package rpchandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
)

// Registry RPC 메서드 등록 및 관리
type Registry struct {
	contexts map[string]*RPCContext // RPCContext만 사용
	groups   map[string]*Group
}

// NewRegistry 새로운 Registry를 생성합니다
func NewRegistry() *Registry {
	return &Registry{
		contexts: make(map[string]*RPCContext),
		groups:   make(map[string]*Group),
	}
}

// RegisterHandler 핸들러를 등록합니다 (옵션 패턴 지원)
func (r *Registry) RegisterHandler(name string, handler Handler, options ...RegisterOption) error {
	if name == "" {
		return fmt.Errorf("handler name cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// 옵션 적용
	opts := newRegisterOptions()
	opts.applyOptions(options)

	return r.scanAndRegisterMethods(name, handler, opts)
}

// scanAndRegisterMethods 핸들러의 메서드를 스캔하여 등록합니다
func (r *Registry) scanAndRegisterMethods(prefix string, handler Handler, opts *RegisterOptions) error {
	handlerValue := reflect.ValueOf(handler)
	handlerType := reflect.TypeOf(handler)

	// 메서드는 원래 타입(포인터 포함)에서 스캔해야 함
	// 모든 메서드 스캔
	for i := 0; i < handlerType.NumMethod(); i++ {
		method := handlerType.Method(i)

		// 제외할 메서드인지 확인
		if opts.shouldIgnoreMethod(method.Name) {
			continue
		}

		methodPath := prefix + "." + method.Name

		// RPCContext 생성
		rpcContext := NewRPCContext(prefix, handlerValue, method)
		r.contexts[methodPath] = rpcContext
	}

	return nil
}

// CallMethod 메서드를 호출합니다
func (r *Registry) CallMethod(ctx context.Context, methodName string, params json.RawMessage) (interface{}, error) {
	rpcContext, exists := r.contexts[methodName]
	if !exists {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	return rpcContext.Call(ctx, params)
}

// GetHandlerFunc 메서드에 대한 HandlerFunc를 반환합니다
func (r *Registry) GetHandlerFunc(methodName string) HandlerFunc {
	return func(ctx context.Context, params json.RawMessage) (interface{}, error) {
		return r.CallMethod(ctx, methodName, params)
	}
}

// Group 그룹을 생성하거나 반환합니다
func (r *Registry) Group(prefix string) *Group {
	if group, exists := r.groups[prefix]; exists {
		return group
	}

	group := &Group{
		prefix:   prefix,
		handlers: make(map[string]Handler),
		registry: r,
	}

	r.groups[prefix] = group
	return group
}

// GetMethodNames 등록된 모든 메서드 이름을 반환합니다
func (r *Registry) GetMethodNames() []string {
	names := make([]string, 0, len(r.contexts))
	for name := range r.contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetMethodNamesWithPrefix 특정 프리픽스로 시작하는 메서드 이름을 반환합니다
func (r *Registry) GetMethodNamesWithPrefix(prefix string) []string {
	var names []string
	for name := range r.contexts {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// HasMethod 메서드가 등록되어 있는지 확인합니다
func (r *Registry) HasMethod(methodName string) bool {
	_, exists := r.contexts[methodName]
	return exists
}

// GetMethodCount 등록된 메서드 수를 반환합니다
func (r *Registry) GetMethodCount() int {
	return len(r.contexts)
}

// GetGroupCount 등록된 그룹 수를 반환합니다
func (r *Registry) GetGroupCount() int {
	return len(r.groups)
}

// Clear 모든 등록된 메서드와 그룹을 제거합니다
func (r *Registry) Clear() {
	r.contexts = make(map[string]*RPCContext)
	r.groups = make(map[string]*Group)
}

// GetRPCContext RPCContext를 반환합니다
func (r *Registry) GetRPCContext(methodName string) (*RPCContext, bool) {
	ctx, exists := r.contexts[methodName]
	return ctx, exists
}

// GetAllRPCContexts 모든 RPCContext를 반환합니다
func (r *Registry) GetAllRPCContexts() map[string]*RPCContext {
	contexts := make(map[string]*RPCContext)
	for name, ctx := range r.contexts {
		contexts[name] = ctx
	}
	return contexts
}

// GetMethodSignatures 모든 메서드의 시그니처를 반환합니다
func (r *Registry) GetMethodSignatures() map[string]string {
	signatures := make(map[string]string)
	for name, ctx := range r.contexts {
		signatures[name] = ctx.GetMethodSignature()
	}
	return signatures
}

// GetMethodInfo 메서드 정보를 반환합니다
func (r *Registry) GetMethodInfo(methodName string) (map[string]interface{}, error) {
	ctx, exists := r.contexts[methodName]
	if !exists {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}
	return ctx.GetInfo(), nil
}

// GetAllMethodInfo 모든 메서드 정보를 반환합니다
func (r *Registry) GetAllMethodInfo() map[string]map[string]interface{} {
	info := make(map[string]map[string]interface{})
	for name, ctx := range r.contexts {
		info[name] = ctx.GetInfo()
	}
	return info
}

// ServeHTTP http.Handler 인터페이스 구현 - Registry를 HTTP 핸들러로 사용
func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// JSON-RPC over HTTP 처리
	r.handleJSONRPC(w, req)
}

// handleJSONRPC JSON-RPC 요청을 처리합니다 (단일 요청 및 배치 요청 지원)
func (r *Registry) handleJSONRPC(w http.ResponseWriter, req *http.Request) {
	// CORS 헤더 설정
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// OPTIONS 요청 처리 (CORS preflight)
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// POST 메서드만 허용
	if req.Method != http.MethodPost {
		r.sendErrorResponse(w, http.StatusMethodNotAllowed, "Only POST method allowed")
		return
	}

	// Content-Type 확인
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		r.sendErrorResponse(w, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	// 요청 본문 읽기
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.sendErrorResponse(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer req.Body.Close()

	// 배치 요청인지 단일 요청인지 확인
	if r.isBatchRequest(body) {
		r.handleBatchRequest(w, req, body)
	} else {
		r.handleSingleRequest(w, req, body)
	}
}

// isBatchRequest 배치 요청인지 확인합니다
func (r *Registry) isBatchRequest(body []byte) bool {
	// JSON이 배열로 시작하는지 확인
	trimmed := bytes.TrimSpace(body)
	return len(trimmed) > 0 && trimmed[0] == '['
}

// handleSingleRequest 단일 JSON-RPC 요청을 처리합니다
func (r *Registry) handleSingleRequest(w http.ResponseWriter, req *http.Request, body []byte) {
	// JSON-RPC 요청 파싱
	var rpcReq JSONRPCRequest
	if err := json.Unmarshal(body, &rpcReq); err != nil {
		r.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON-RPC format")
		return
	}

	// 요청 처리
	response := r.processSingleRequest(req.Context(), rpcReq)

	// 응답 전송
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// handleBatchRequest 배치 JSON-RPC 요청을 처리합니다
func (r *Registry) handleBatchRequest(w http.ResponseWriter, req *http.Request, body []byte) {
	// 배치 요청 파싱
	var batchReq []JSONRPCRequest
	if err := json.Unmarshal(body, &batchReq); err != nil {
		r.sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON-RPC batch format")
		return
	}

	// 빈 배치 요청 확인
	if len(batchReq) == 0 {
		r.sendErrorResponse(w, http.StatusBadRequest, "Empty batch request")
		return
	}

	// 배치 요청 처리 (순서 보장, 동기식 실행)
	responses := r.processBatchRequest(req.Context(), batchReq)

	// 응답 전송
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responses)
}

// processSingleRequest 단일 요청을 처리합니다
func (r *Registry) processSingleRequest(ctx context.Context, rpcReq JSONRPCRequest) JSONRPCResponse {
	// JSON-RPC 버전 확인
	if rpcReq.JSONRPC != "2.0" {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32600,
				Message: "Invalid Request",
				Data:    "JSON-RPC version must be 2.0",
			},
			ID: rpcReq.ID,
		}
	}

	// 메서드 호출
	result, err := r.CallMethod(ctx, rpcReq.Method, rpcReq.Params)

	// 응답 생성
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    err.Error(),
			},
			ID: rpcReq.ID,
		}
	}

	return JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      rpcReq.ID,
	}
}

// processBatchRequest 배치 요청을 순서대로 동기식으로 처리합니다
func (r *Registry) processBatchRequest(ctx context.Context, batchReq []JSONRPCRequest) []JSONRPCResponse {
	responses := make([]JSONRPCResponse, len(batchReq))

	// 순서를 보장하기 위해 순차적으로 처리 (동기식)
	for i, req := range batchReq {
		responses[i] = r.processSingleRequest(ctx, req)
	}

	return responses
}

// JSONRPCRequest JSON-RPC 요청 구조체
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id"`
}

// JSONRPCResponse JSON-RPC 응답 구조체
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{}   `json:"id"`
}

// JSONRPCError JSON-RPC 에러 구조체
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// sendJSONRPCResult 성공 응답을 전송합니다
func (r *Registry) sendJSONRPCResult(w http.ResponseWriter, id interface{}, result interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Result:  result,
		ID:      id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// sendJSONRPCError 에러 응답을 전송합니다
func (r *Registry) sendJSONRPCError(w http.ResponseWriter, id interface{}, code int, message string, data interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
		ID: id,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // JSON-RPC는 HTTP 200으로 에러도 전송
	json.NewEncoder(w).Encode(response)
}

// sendErrorResponse 일반 HTTP 에러 응답을 전송합니다
func (r *Registry) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error": message,
	}
	json.NewEncoder(w).Encode(response)
}
