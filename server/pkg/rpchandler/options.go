package rpchandler

import (
	"reflect"
	"runtime"
)

// RegisterOptions 핸들러 등록 옵션
type RegisterOptions struct {
	ignoredMethods map[string]bool // 제외할 메서드 이름들
}

// RegisterOption 등록 옵션 함수 타입
type RegisterOption func(*RegisterOptions)

// newRegisterOptions 기본 등록 옵션을 생성합니다
func newRegisterOptions() *RegisterOptions {
	return &RegisterOptions{
		ignoredMethods: make(map[string]bool),
	}
}

// applyOptions 옵션들을 적용합니다
func (opts *RegisterOptions) applyOptions(options []RegisterOption) {
	for _, option := range options {
		option(opts)
	}
}

// shouldIgnoreMethod 메서드를 제외해야 하는지 확인합니다
func (opts *RegisterOptions) shouldIgnoreMethod(methodName string) bool {
	return opts.ignoredMethods[methodName]
}

// WithIgnoreNames 메서드 이름으로 제외할 함수들을 지정합니다
func WithIgnoreNames(methodNames ...string) RegisterOption {
	return func(opts *RegisterOptions) {
		for _, name := range methodNames {
			opts.ignoredMethods[name] = true
		}
	}
}

// WithIgnore 함수 포인터로 제외할 함수들을 지정합니다
func WithIgnore(funcs ...interface{}) RegisterOption {
	return func(opts *RegisterOptions) {
		for _, fn := range funcs {
			if methodName := getFunctionName(fn); methodName != "" {
				opts.ignoredMethods[methodName] = true
			}
		}
	}
}

// WithIgnoreReflect 리플렉션 정보로 제외할 함수들을 지정합니다
func WithIgnoreReflect(methods ...reflect.Method) RegisterOption {
	return func(opts *RegisterOptions) {
		for _, method := range methods {
			opts.ignoredMethods[method.Name] = true
		}
	}
}

// getFunctionName 함수 포인터에서 메서드 이름을 추출합니다
func getFunctionName(fn interface{}) string {
	if fn == nil {
		return ""
	}

	// 함수 포인터의 타입 확인
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return ""
	}

	// runtime.FuncForPC를 사용해서 함수 이름 추출
	pc := fnValue.Pointer()
	funcInfo := runtime.FuncForPC(pc)
	if funcInfo == nil {
		return ""
	}

	fullName := funcInfo.Name()

	// 메서드 이름만 추출
	// 패턴들:
	// "package.(*Type).Method" -> "Method"
	// "package.Type.Method" -> "Method"
	// "package.function" -> "function"

	// 마지막 점 이후의 문자열 추출
	if lastDot := findLastDot(fullName); lastDot != -1 {
		methodName := fullName[lastDot+1:]

		// 메서드 이름에서 "-fm" 같은 접미사 제거 (Go 컴파일러가 추가하는 경우)
		if dashIndex := findDash(methodName); dashIndex != -1 {
			methodName = methodName[:dashIndex]
		}

		return methodName
	}

	return fullName
}

// findLastDot 마지막 점의 위치를 찾습니다
func findLastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

func findDash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '-' {
			return i
		}
	}
	return -1
}

// 추가 옵션들 (향후 확장 가능)

// WithPrefix 메서드 이름에 프리픽스를 추가합니다
func WithPrefix(prefix string) RegisterOption {
	return func(opts *RegisterOptions) {
		// 향후 구현 예정
	}
}

// WithMethodFilter 커스텀 메서드 필터를 적용합니다
func WithMethodFilter(filter func(methodName string) bool) RegisterOption {
	return func(opts *RegisterOptions) {
		// 향후 구현 예정
	}
}

// 디버깅용 함수들

// GetIgnoredMethods 제외된 메서드 목록을 반환합니다
func (opts *RegisterOptions) GetIgnoredMethods() []string {
	methods := make([]string, 0, len(opts.ignoredMethods))
	for method := range opts.ignoredMethods {
		methods = append(methods, method)
	}
	return methods
}

// HasIgnoredMethods 제외된 메서드가 있는지 확인합니다
func (opts *RegisterOptions) HasIgnoredMethods() bool {
	return len(opts.ignoredMethods) > 0
}
