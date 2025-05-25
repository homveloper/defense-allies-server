package rpchandler

import (
	"fmt"
)

// Composer 여러 핸들러를 조합하는 빌더
type Composer struct {
	handlers map[string]Handler
}

// NewComposer 새로운 Composer를 생성합니다
func NewComposer() *Composer {
	return &Composer{
		handlers: make(map[string]Handler),
	}
}

// Add 핸들러를 추가합니다 (메서드 체이닝 지원)
func (c *Composer) Add(name string, handler Handler) *Composer {
	if name == "" || handler == nil {
		// 에러가 발생해도 체이닝을 계속할 수 있도록 현재 인스턴스 반환
		return c
	}
	
	c.handlers[name] = handler
	return c
}

// Compose 조합된 핸들러를 생성합니다
func (c *Composer) Compose() Handler {
	if len(c.handlers) == 0 {
		return &CompositeHandler{handlers: make(map[string]Handler)}
	}
	
	// 핸들러 맵 복사 (불변성 보장)
	handlersCopy := make(map[string]Handler)
	for name, handler := range c.handlers {
		handlersCopy[name] = handler
	}
	
	return &CompositeHandler{
		handlers: handlersCopy,
	}
}

// GetHandlerNames 추가된 핸들러 이름들을 반환합니다
func (c *Composer) GetHandlerNames() []string {
	names := make([]string, 0, len(c.handlers))
	for name := range c.handlers {
		names = append(names, name)
	}
	return names
}

// HasHandler 핸들러가 추가되어 있는지 확인합니다
func (c *Composer) HasHandler(name string) bool {
	_, exists := c.handlers[name]
	return exists
}

// GetHandlerCount 추가된 핸들러 수를 반환합니다
func (c *Composer) GetHandlerCount() int {
	return len(c.handlers)
}

// Clear 모든 핸들러를 제거합니다
func (c *Composer) Clear() *Composer {
	c.handlers = make(map[string]Handler)
	return c
}

// CompositeHandler 여러 핸들러를 조합한 핸들러
type CompositeHandler struct {
	handlers map[string]Handler
}

// GetHandlers 조합된 핸들러들을 반환합니다
func (ch *CompositeHandler) GetHandlers() map[string]Handler {
	// 불변성을 위해 복사본 반환
	handlersCopy := make(map[string]Handler)
	for name, handler := range ch.handlers {
		handlersCopy[name] = handler
	}
	return handlersCopy
}

// GetHandlerNames 조합된 핸들러 이름들을 반환합니다
func (ch *CompositeHandler) GetHandlerNames() []string {
	names := make([]string, 0, len(ch.handlers))
	for name := range ch.handlers {
		names = append(names, name)
	}
	return names
}

// HasHandler 특정 핸들러가 포함되어 있는지 확인합니다
func (ch *CompositeHandler) HasHandler(name string) bool {
	_, exists := ch.handlers[name]
	return exists
}

// GetHandlerCount 조합된 핸들러 수를 반환합니다
func (ch *CompositeHandler) GetHandlerCount() int {
	return len(ch.handlers)
}

// GetHandler 특정 이름의 핸들러를 반환합니다
func (ch *CompositeHandler) GetHandler(name string) (Handler, error) {
	handler, exists := ch.handlers[name]
	if !exists {
		return nil, fmt.Errorf("handler not found: %s", name)
	}
	return handler, nil
}

// 편의 함수들

// ComposeHandlers 여러 핸들러를 한 번에 조합합니다
func ComposeHandlers(handlers map[string]Handler) Handler {
	composer := NewComposer()
	for name, handler := range handlers {
		composer.Add(name, handler)
	}
	return composer.Compose()
}

// QuickCompose 빠른 조합을 위한 헬퍼 함수
func QuickCompose(pairs ...interface{}) Handler {
	if len(pairs)%2 != 0 {
		// 잘못된 인자 수인 경우 빈 CompositeHandler 반환
		return &CompositeHandler{handlers: make(map[string]Handler)}
	}
	
	composer := NewComposer()
	for i := 0; i < len(pairs); i += 2 {
		name, ok1 := pairs[i].(string)
		handler, ok2 := pairs[i+1].(Handler)
		
		if ok1 && ok2 {
			composer.Add(name, handler)
		}
	}
	
	return composer.Compose()
}
