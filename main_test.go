package main

import (
	"defense-allies-server/internal/serverapp/health"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHomeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(homeHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"message": "Welcome to Defense Allies Server", "status": "running"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHealthHandler(t *testing.T) {
	// HealthApp을 사용한 테스트로 변경
	healthApp := health.NewHealthApp(nil) // Redis 없이 테스트

	mux := http.NewServeMux()
	healthApp.RegisterRoutes(mux)

	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// JSON 응답 확인
	if !strings.Contains(rr.Body.String(), "status") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}
