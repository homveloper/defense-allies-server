package rpchandler

import (
	"testing"
)

// TestOptionsPattern 옵션 패턴 테스트
func TestOptionsPattern(t *testing.T) {
	registry := NewRegistry()

	t.Run("WithIgnoreNames", func(t *testing.T) {
		// 메서드 이름으로 제외
		err := registry.RegisterHandler("game1", &GameHandler{}, WithIgnoreNames("GetState", "Ping"))
		if err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}

		methods := registry.GetMethodNames()
		t.Logf("Registered methods (ignore names): %v", methods)

		// GetState와 Ping이 제외되었는지 확인
		for _, method := range methods {
			if method == "game1.GetState" || method == "game1.Ping" {
				t.Errorf("Method %s should be ignored", method)
			}
		}

		// GetStatus는 등록되어야 함
		found := false
		for _, method := range methods {
			if method == "game1.GetStatus" {
				found = true
				break
			}
		}
		if !found {
			t.Error("GetStatus method should be registered")
		}
	})

	t.Run("WithIgnore", func(t *testing.T) {
		// 함수 포인터로 제외
		gameHandler := &GameHandler{}

		err := registry.RegisterHandler("game2", gameHandler, WithIgnore(gameHandler.GetState, gameHandler.ProcessRawData))
		if err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}

		methods := registry.GetMethodNamesWithPrefix("game2")
		t.Logf("Registered methods (ignore funcs): %v", methods)

		// GetState와 ProcessRawData가 제외되었는지 확인
		for _, method := range methods {
			if method == "game2.GetState" || method == "game2.ProcessRawData" {
				t.Errorf("Method %s should be ignored", method)
			}
		}

		// GetStatus와 Ping은 등록되어야 함
		expectedMethods := []string{"game2.GetStatus", "game2.Ping"}
		for _, expected := range expectedMethods {
			found := false
			for _, method := range methods {
				if method == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Method %s should be registered", expected)
			}
		}
	})

	t.Run("NoOptions", func(t *testing.T) {
		// 옵션 없이 등록 (모든 메서드 등록)
		err := registry.RegisterHandler("game3", &GameHandler{})
		if err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}

		methods := registry.GetMethodNamesWithPrefix("game3")
		t.Logf("Registered methods (no options): %v", methods)

		// 모든 public 메서드가 등록되어야 함
		expectedMethods := []string{"game3.GetStatus", "game3.Ping", "game3.GetState", "game3.ProcessRawData"}
		for _, expected := range expectedMethods {
			found := false
			for _, method := range methods {
				if method == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Method %s should be registered", expected)
			}
		}
	})

	t.Run("CombinedOptions", func(t *testing.T) {
		// 여러 옵션 조합
		gameHandler := &GameHandler{}

		err := registry.RegisterHandler("game4", gameHandler,
			WithIgnoreNames("Ping"),
			WithIgnore(gameHandler.ProcessRawData))
		if err != nil {
			t.Fatalf("Failed to register handler: %v", err)
		}

		methods := registry.GetMethodNamesWithPrefix("game4")
		t.Logf("Registered methods (combined options): %v", methods)

		// Ping과 ProcessRawData가 제외되었는지 확인
		for _, method := range methods {
			if method == "game4.Ping" || method == "game4.ProcessRawData" {
				t.Errorf("Method %s should be ignored", method)
			}
		}

		// GetStatus와 GetState는 등록되어야 함
		expectedMethods := []string{"game4.GetStatus", "game4.GetState"}
		for _, expected := range expectedMethods {
			found := false
			for _, method := range methods {
				if method == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Method %s should be registered", expected)
			}
		}
	})
}

// TestFunctionNameExtraction 함수 이름 추출 테스트
func TestFunctionNameExtraction(t *testing.T) {
	gameHandler := &GameHandler{}

	testCases := []struct {
		name     string
		fn       interface{}
		expected string
	}{
		{"GetStatus", gameHandler.GetStatus, "GetStatus"},
		{"GetState", gameHandler.GetState, "GetState"},
		{"Ping", gameHandler.Ping, "Ping"},
		{"ProcessRawData", gameHandler.ProcessRawData, "ProcessRawData"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getFunctionName(tc.fn)
			t.Logf("Function name for %s: %s", tc.name, result)

			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

// TestGroupWithOptions 그룹에서 옵션 패턴 테스트
func TestGroupWithOptions(t *testing.T) {
	registry := NewRegistry()
	group := registry.Group("api")

	gameHandler := &GameHandler{}

	err := group.RegisterHandler("game", gameHandler, WithIgnore(gameHandler.GetState))
	if err != nil {
		t.Fatalf("Failed to register handler to group: %v", err)
	}

	methods := registry.GetMethodNamesWithPrefix("api.game")
	t.Logf("Group methods with options: %v", methods)

	// GetState가 제외되었는지 확인
	for _, method := range methods {
		if method == "api.game.GetState" {
			t.Error("GetState should be ignored in group")
		}
	}

	// 다른 메서드들은 등록되어야 함
	expectedMethods := []string{"api.game.GetStatus", "api.game.Ping", "api.game.ProcessRawData"}
	for _, expected := range expectedMethods {
		found := false
		for _, method := range methods {
			if method == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Method %s should be registered in group", expected)
		}
	}
}
