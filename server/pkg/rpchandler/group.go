package rpchandler

import (
	"fmt"
	"sort"
	"strings"
)

// Group 계층적 API 경로 관리
type Group struct {
	prefix   string
	handlers map[string]Handler
	registry *Registry
}

// RegisterHandler 그룹에 핸들러를 등록합니다
func (g *Group) RegisterHandler(name string, handler Handler, options ...RegisterOption) error {
	if name == "" {
		return fmt.Errorf("handler name cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	// 핸들러 저장
	g.handlers[name] = handler

	// 전체 경로 생성
	fullPath := g.prefix + "." + name

	// 옵션 적용
	opts := newRegisterOptions()
	opts.applyOptions(options)

	// Registry에 메서드 등록
	return g.registry.scanAndRegisterMethods(fullPath, handler, opts)
}

// Group 하위 그룹을 생성합니다
func (g *Group) Group(suffix string) *Group {
	if suffix == "" {
		return g
	}

	fullPrefix := g.prefix + "." + suffix
	return g.registry.Group(fullPrefix)
}

// RegisterComposite CompositeHandler를 등록합니다
func (g *Group) RegisterComposite(composite *CompositeHandler) error {
	if composite == nil {
		return fmt.Errorf("composite handler cannot be nil")
	}

	// CompositeHandler의 모든 핸들러를 등록
	for name, handler := range composite.handlers {
		if err := g.RegisterHandler(name, handler); err != nil {
			return fmt.Errorf("failed to register composite handler %s: %w", name, err)
		}
	}

	return nil
}

// GetMethodNames 이 그룹에 속한 메서드 이름들을 반환합니다
func (g *Group) GetMethodNames() []string {
	var names []string
	prefix := g.prefix + "."

	for methodName := range g.registry.contexts {
		if strings.HasPrefix(methodName, prefix) {
			names = append(names, methodName)
		}
	}

	sort.Strings(names)
	return names
}

// GetHandlerNames 이 그룹에 등록된 핸들러 이름들을 반환합니다
func (g *Group) GetHandlerNames() []string {
	names := make([]string, 0, len(g.handlers))
	for name := range g.handlers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasHandler 핸들러가 등록되어 있는지 확인합니다
func (g *Group) HasHandler(name string) bool {
	_, exists := g.handlers[name]
	return exists
}

// GetPrefix 그룹의 프리픽스를 반환합니다
func (g *Group) GetPrefix() string {
	return g.prefix
}

// GetMethodCount 이 그룹에 속한 메서드 수를 반환합니다
func (g *Group) GetMethodCount() int {
	count := 0
	prefix := g.prefix + "."

	for methodName := range g.registry.contexts {
		if strings.HasPrefix(methodName, prefix) {
			count++
		}
	}

	return count
}

// GetHandlerCount 이 그룹에 등록된 핸들러 수를 반환합니다
func (g *Group) GetHandlerCount() int {
	return len(g.handlers)
}

// Clear 이 그룹의 모든 핸들러를 제거합니다
func (g *Group) Clear() {
	// 등록된 메서드들 제거
	prefix := g.prefix + "."

	for methodName := range g.registry.contexts {
		if strings.HasPrefix(methodName, prefix) {
			delete(g.registry.contexts, methodName)
		}
	}

	// 핸들러 맵 초기화
	g.handlers = make(map[string]Handler)
}
