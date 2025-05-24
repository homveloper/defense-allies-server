package rpchandler

import (
	"context"
	"encoding/json"
	"reflect"
)

// Handler 인터페이스 - 모든 RPC 핸들러가 구현해야 하는 기본 인터페이스
type Handler interface{}

// HandlerFunc JSON-RPC 핸들러 함수 타입
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// returnTypeInfo 메서드의 반환값 정보
type returnTypeInfo struct {
	hasReturn bool // 반환값이 있는지
	hasError  bool // 에러를 반환하는지
	returnIdx int  // 반환값의 인덱스
	errorIdx  int  // 에러의 인덱스
}

// MethodWrapper 메서드 호출을 위한 래퍼
type MethodWrapper struct {
	method      reflect.Method // 리플렉션 메서드
	handler     reflect.Value  // 핸들러 인스턴스
	hasContext  bool           // context.Context 파라미터가 있는지
	paramType   reflect.Type   // 파라미터 타입
	needsParams bool           // 파라미터가 필요한지
	returnInfo  returnTypeInfo // 반환값 정보
	methodPath  string         // 메서드 경로 (예: "game.core.GetState")
}

// Call 메서드를 호출합니다
func (mw *MethodWrapper) Call(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// 파라미터 준비
	args := []reflect.Value{}

	// 첫 번째 인자는 항상 receiver (핸들러 인스턴스)
	args = append(args, mw.handler)

	// context.Context 파라미터 추가
	if mw.hasContext {
		args = append(args, reflect.ValueOf(ctx))
	}

	// 파라미터 추가
	if mw.needsParams {
		if mw.paramType == reflect.TypeOf(json.RawMessage{}) {
			// json.RawMessage 타입인 경우 그대로 전달
			args = append(args, reflect.ValueOf(params))
		} else {
			// 구조체 타입인 경우 언마샬링
			paramValue := reflect.New(mw.paramType).Interface()
			if err := json.Unmarshal(params, paramValue); err != nil {
				return nil, err
			}
			args = append(args, reflect.ValueOf(paramValue).Elem())
		}
	}

	// 메서드 호출
	results := mw.method.Func.Call(args)

	// 반환값 처리
	return mw.processResults(results)
}

// processResults 메서드 호출 결과를 처리합니다
func (mw *MethodWrapper) processResults(results []reflect.Value) (interface{}, error) {
	if len(results) == 0 {
		return nil, nil
	}

	var result interface{}
	var err error

	// 반환값 추출
	if mw.returnInfo.hasReturn {
		resultValue := results[mw.returnInfo.returnIdx]
		if !resultValue.IsNil() {
			result = resultValue.Interface()
		}
	}

	// 에러 추출
	if mw.returnInfo.hasError {
		errorValue := results[mw.returnInfo.errorIdx]
		if !errorValue.IsNil() {
			err = errorValue.Interface().(error)
		}
	}

	return result, err
}

// analyzeMethod 메서드를 분석하여 MethodWrapper를 생성합니다
func analyzeMethod(handler reflect.Value, method reflect.Method, methodPath string) *MethodWrapper {
	methodType := method.Type

	mw := &MethodWrapper{
		method:     method,
		handler:    handler,
		methodPath: methodPath,
	}

	// 파라미터 분석 (첫 번째는 receiver이므로 제외)
	numIn := methodType.NumIn()
	paramIndex := 1 // receiver 다음부터

	// context.Context 확인
	if numIn > paramIndex {
		if methodType.In(paramIndex) == reflect.TypeOf((*context.Context)(nil)).Elem() {
			mw.hasContext = true
			paramIndex++
		}
	}

	// 파라미터 타입 확인
	if numIn > paramIndex {
		mw.needsParams = true
		mw.paramType = methodType.In(paramIndex)
	}

	// 반환값 분석
	numOut := methodType.NumOut()
	mw.returnInfo = analyzeReturnTypes(methodType, numOut)

	return mw
}

// analyzeReturnTypes 반환값 타입을 분석합니다
func analyzeReturnTypes(methodType reflect.Type, numOut int) returnTypeInfo {
	info := returnTypeInfo{}

	errorType := reflect.TypeOf((*error)(nil)).Elem()

	for i := 0; i < numOut; i++ {
		outType := methodType.Out(i)

		if outType.Implements(errorType) {
			info.hasError = true
			info.errorIdx = i
		} else {
			info.hasReturn = true
			info.returnIdx = i
		}
	}

	return info
}

// isPublicMethod 메서드가 public인지 확인합니다
func isPublicMethod(method reflect.Method) bool {
	// 메서드 이름이 대문자로 시작하는지 확인
	return method.Name[0] >= 'A' && method.Name[0] <= 'Z'
}
