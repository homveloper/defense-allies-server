package rpchandler

import (
	"testing"
)

// TestDebugRegistration 등록 과정을 디버깅하는 테스트
func TestDebugRegistration(t *testing.T) {
	registry := NewRegistry()
	
	t.Logf("Before registration - method count: %d", registry.GetMethodCount())
	
	err := registry.RegisterHandler("game", &GameHandler{})
	if err != nil {
		t.Fatalf("Failed to register GameHandler: %v", err)
	}
	
	t.Logf("After registration - method count: %d", registry.GetMethodCount())
	
	methods := registry.GetMethodNames()
	t.Logf("Registered methods: %v", methods)
	
	contexts := registry.GetAllRPCContexts()
	t.Logf("RPC contexts count: %d", len(contexts))
	
	for name, ctx := range contexts {
		t.Logf("Context: %s -> %+v", name, ctx.GetInfo())
	}
}
