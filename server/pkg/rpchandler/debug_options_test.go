package rpchandler

import (
	"testing"
)

// TestDebugOptions 옵션 적용 과정을 디버깅하는 테스트
func TestDebugOptions(t *testing.T) {
	registry := NewRegistry()
	gameHandler := &GameHandler{}
	
	// 함수 이름 추출 확인
	getStateName := getFunctionName(gameHandler.GetState)
	processRawDataName := getFunctionName(gameHandler.ProcessRawData)
	
	t.Logf("GetState function name: %s", getStateName)
	t.Logf("ProcessRawData function name: %s", processRawDataName)
	
	// 옵션 생성 및 확인
	opts := newRegisterOptions()
	
	// WithIgnore 옵션 적용
	ignoreOption := WithIgnore(gameHandler.GetState, gameHandler.ProcessRawData)
	ignoreOption(opts)
	
	t.Logf("Ignored methods after WithIgnore: %v", opts.GetIgnoredMethods())
	
	// 각 메서드가 제외되는지 확인
	t.Logf("Should ignore GetState: %v", opts.shouldIgnoreMethod("GetState"))
	t.Logf("Should ignore ProcessRawData: %v", opts.shouldIgnoreMethod("ProcessRawData"))
	t.Logf("Should ignore GetStatus: %v", opts.shouldIgnoreMethod("GetStatus"))
	t.Logf("Should ignore Ping: %v", opts.shouldIgnoreMethod("Ping"))
	
	// 실제 등록 과정 확인
	err := registry.RegisterHandler("debug", gameHandler, WithIgnore(gameHandler.GetState, gameHandler.ProcessRawData))
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}
	
	methods := registry.GetMethodNamesWithPrefix("debug")
	t.Logf("Actually registered methods: %v", methods)
	
	// 각 메서드별 등록 여부 확인
	for _, method := range []string{"debug.GetState", "debug.ProcessRawData", "debug.GetStatus", "debug.Ping"} {
		found := false
		for _, registered := range methods {
			if registered == method {
				found = true
				break
			}
		}
		t.Logf("Method %s registered: %v", method, found)
	}
}
