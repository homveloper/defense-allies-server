package utils

import (
	"encoding/json"
	"net/http"
)

// APIResponse 표준 API 응답 구조체
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SendJSONResponse JSON 응답을 전송합니다
func SendJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// SendSuccessResponse 성공 응답을 전송합니다
func SendSuccessResponse(w http.ResponseWriter, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	SendJSONResponse(w, http.StatusOK, response)
}

// SendErrorResponse 에러 응답을 전송합니다
func SendErrorResponse(w http.ResponseWriter, statusCode int, errorMsg string) {
	response := APIResponse{
		Success: false,
		Error:   errorMsg,
	}
	SendJSONResponse(w, statusCode, response)
}

// SendCreatedResponse 생성 성공 응답을 전송합니다
func SendCreatedResponse(w http.ResponseWriter, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	SendJSONResponse(w, http.StatusCreated, response)
}
