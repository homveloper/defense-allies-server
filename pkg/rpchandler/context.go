package rpchandler

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// InputArgInfo 입력 인자 정보
type InputArgInfo struct {
	Index     int          `json:"index"`       // 인자 순서
	Name      string       `json:"name"`        // 인자 이름 (구조체 필드명 등)
	Type      reflect.Type `json:"-"`           // 인자 타입 (JSON 직렬화 제외)
	TypeName  string       `json:"type_name"`   // 타입 이름 (문자열)
	IsContext bool         `json:"is_context"`  // context.Context 타입인지
	IsPointer bool         `json:"is_pointer"`  // 포인터 타입인지
	IsStruct  bool         `json:"is_struct"`   // 구조체 타입인지
	IsRawJSON bool         `json:"is_raw_json"` // json.RawMessage 타입인지
	Required  bool         `json:"required"`    // 필수 인자인지
}

// OutputArgInfo 출력 인자 정보
type OutputArgInfo struct {
	Index     int          `json:"index"`      // 반환값 순서
	Type      reflect.Type `json:"-"`          // 반환값 타입 (JSON 직렬화 제외)
	TypeName  string       `json:"type_name"`  // 타입 이름 (문자열)
	IsError   bool         `json:"is_error"`   // error 타입인지
	IsPointer bool         `json:"is_pointer"` // 포인터 타입인지
	IsNil     bool         `json:"is_nil"`     // nil 가능한지
}

// RPCContext RPC 함수 호출 컨텍스트
type RPCContext struct {
	FunctionName string          `json:"function_name"` // 함수 이름
	HandlerName  string          `json:"handler_name"`  // 핸들러 이름
	FullPath     string          `json:"full_path"`     // 전체 경로 (handler.function)
	Method       reflect.Method  `json:"-"`             // 리플렉션 메서드 (JSON 직렬화 제외)
	Handler      reflect.Value   `json:"-"`             // 핸들러 인스턴스 (JSON 직렬화 제외)
	InputArgs    []InputArgInfo  `json:"input_args"`    // 입력 인자 정보들
	OutputArgs   []OutputArgInfo `json:"output_args"`   // 출력 인자 정보들

	// 캐시된 정보
	hasContext   bool // context.Context 파라미터가 있는지
	contextIndex int  // context.Context의 인덱스
	hasParams    bool // 파라미터가 있는지
	paramsIndex  int  // 파라미터의 인덱스
	hasReturn    bool // 반환값이 있는지
	returnIndex  int  // 반환값의 인덱스
	hasError     bool // 에러 반환이 있는지
	errorIndex   int  // 에러의 인덱스
}

// NewRPCContext 새로운 RPCContext를 생성합니다
func NewRPCContext(handlerName string, handler reflect.Value, method reflect.Method) *RPCContext {
	ctx := &RPCContext{
		FunctionName: method.Name,
		HandlerName:  handlerName,
		FullPath:     handlerName + "." + method.Name,
		Method:       method,
		Handler:      handler,
		InputArgs:    []InputArgInfo{},
		OutputArgs:   []OutputArgInfo{},
	}

	ctx.analyzeInputArgs()
	ctx.analyzeOutputArgs()
	ctx.buildCache()

	return ctx
}

// analyzeInputArgs 입력 인자들을 분석합니다
func (ctx *RPCContext) analyzeInputArgs() {
	methodType := ctx.Method.Type
	numIn := methodType.NumIn()

	// 첫 번째 인자는 receiver이므로 제외하고 시작
	for i := 1; i < numIn; i++ {
		argType := methodType.In(i)
		argInfo := InputArgInfo{
			Index:     i - 1, // receiver 제외한 실제 인덱스
			Type:      argType,
			TypeName:  argType.String(),
			IsPointer: argType.Kind() == reflect.Ptr,
			Required:  true, // 기본적으로 필수
		}

		// 타입별 분석
		switch {
		case argType == reflect.TypeOf((*context.Context)(nil)).Elem():
			argInfo.IsContext = true
			argInfo.Name = "context"
			argInfo.Required = false // context는 자동 주입되므로 필수 아님

		case argType == reflect.TypeOf(json.RawMessage{}):
			argInfo.IsRawJSON = true
			argInfo.Name = "params"

		case argType.Kind() == reflect.Struct || (argType.Kind() == reflect.Ptr && argType.Elem().Kind() == reflect.Struct):
			argInfo.IsStruct = true
			if argType.Kind() == reflect.Ptr {
				argInfo.Name = argType.Elem().Name()
			} else {
				argInfo.Name = argType.Name()
			}

		default:
			argInfo.Name = fmt.Sprintf("arg%d", i-1)
		}

		ctx.InputArgs = append(ctx.InputArgs, argInfo)
	}
}

// analyzeOutputArgs 출력 인자들을 분석합니다
func (ctx *RPCContext) analyzeOutputArgs() {
	methodType := ctx.Method.Type
	numOut := methodType.NumOut()

	errorType := reflect.TypeOf((*error)(nil)).Elem()

	for i := 0; i < numOut; i++ {
		outType := methodType.Out(i)
		outInfo := OutputArgInfo{
			Index:     i,
			Type:      outType,
			TypeName:  outType.String(),
			IsPointer: outType.Kind() == reflect.Ptr,
			IsNil:     outType.Kind() == reflect.Ptr || outType.Kind() == reflect.Interface,
		}

		// error 타입 확인
		if outType.Implements(errorType) {
			outInfo.IsError = true
		}

		ctx.OutputArgs = append(ctx.OutputArgs, outInfo)
	}
}

// buildCache 캐시 정보를 구축합니다
func (ctx *RPCContext) buildCache() {
	// 입력 인자 캐시
	for i, arg := range ctx.InputArgs {
		if arg.IsContext {
			ctx.hasContext = true
			ctx.contextIndex = i
		} else {
			ctx.hasParams = true
			ctx.paramsIndex = i
		}
	}

	// 출력 인자 캐시
	for i, arg := range ctx.OutputArgs {
		if arg.IsError {
			ctx.hasError = true
			ctx.errorIndex = i
		} else {
			ctx.hasReturn = true
			ctx.returnIndex = i
		}
	}
}

// DecodeParams JSON-RPC 파라미터를 디코딩합니다
func (ctx *RPCContext) DecodeParams(params json.RawMessage) ([]reflect.Value, error) {
	args := []reflect.Value{ctx.Handler} // receiver 추가

	for _, argInfo := range ctx.InputArgs {
		if argInfo.IsContext {
			// context는 나중에 호출 시점에 추가
			continue
		}

		if argInfo.IsRawJSON {
			// json.RawMessage는 그대로 전달
			args = append(args, reflect.ValueOf(params))
		} else if argInfo.IsStruct {
			// 구조체는 언마샬링
			var paramValue reflect.Value
			if argInfo.IsPointer {
				paramValue = reflect.New(argInfo.Type.Elem())
			} else {
				paramValue = reflect.New(argInfo.Type)
			}

			if err := json.Unmarshal(params, paramValue.Interface()); err != nil {
				return nil, fmt.Errorf("failed to unmarshal params for %s: %w", argInfo.Name, err)
			}

			if !argInfo.IsPointer {
				paramValue = paramValue.Elem()
			}

			args = append(args, paramValue)
		} else {
			// 기타 타입들은 기본값으로 처리
			args = append(args, reflect.Zero(argInfo.Type))
		}
	}

	return args, nil
}

// Call 함수를 호출합니다
func (ctx *RPCContext) Call(callCtx context.Context, params json.RawMessage) (interface{}, error) {
	// 파라미터 디코딩
	args, err := ctx.DecodeParams(params)
	if err != nil {
		return nil, err
	}

	// context.Context 추가 (필요한 경우)
	if ctx.hasContext {
		// context를 적절한 위치에 삽입
		finalArgs := make([]reflect.Value, 0, len(args)+1)
		finalArgs = append(finalArgs, args[0]) // receiver

		contextInserted := false
		for i, argInfo := range ctx.InputArgs {
			if argInfo.IsContext {
				finalArgs = append(finalArgs, reflect.ValueOf(callCtx))
				contextInserted = true
			} else {
				// receiver 다음부터의 실제 인자들
				if len(args) > i {
					finalArgs = append(finalArgs, args[i])
				}
			}
		}

		if !contextInserted {
			// context가 첫 번째 파라미터인 경우
			finalArgs = append([]reflect.Value{args[0], reflect.ValueOf(callCtx)}, args[1:]...)
		}

		args = finalArgs
	}

	// 메서드 호출
	results := ctx.Method.Func.Call(args)

	// 결과 변환
	return ctx.ConvertOutput(results)
}

// ConvertOutput 출력 결과를 JSON-RPC 응답으로 변환합니다
func (ctx *RPCContext) ConvertOutput(results []reflect.Value) (interface{}, error) {
	if len(results) == 0 {
		return nil, nil
	}

	var result interface{}
	var err error

	// 반환값 처리
	if ctx.hasReturn {
		returnValue := results[ctx.returnIndex]
		if !returnValue.IsNil() {
			result = returnValue.Interface()
		}
	}

	// 에러 처리
	if ctx.hasError {
		errorValue := results[ctx.errorIndex]
		if !errorValue.IsNil() {
			err = errorValue.Interface().(error)
		}
	}

	return result, err
}

// GetMethodSignature 메서드 시그니처를 문자열로 반환합니다
func (ctx *RPCContext) GetMethodSignature() string {
	signature := fmt.Sprintf("%s(", ctx.FunctionName)

	// 입력 파라미터
	for i, arg := range ctx.InputArgs {
		if i > 0 {
			signature += ", "
		}
		signature += fmt.Sprintf("%s %s", arg.Name, arg.TypeName)
	}

	signature += ")"

	// 출력 파라미터
	if len(ctx.OutputArgs) > 0 {
		signature += " ("
		for i, arg := range ctx.OutputArgs {
			if i > 0 {
				signature += ", "
			}
			signature += arg.TypeName
		}
		signature += ")"
	}

	return signature
}

// IsValid RPCContext가 유효한지 확인합니다
func (ctx *RPCContext) IsValid() bool {
	return ctx.FunctionName != "" && ctx.HandlerName != "" && ctx.Method.Name != ""
}

// GetInfo RPCContext 정보를 맵으로 반환합니다 (디버깅 용도)
func (ctx *RPCContext) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"function_name": ctx.FunctionName,
		"handler_name":  ctx.HandlerName,
		"full_path":     ctx.FullPath,
		"signature":     ctx.GetMethodSignature(),
		"input_count":   len(ctx.InputArgs),
		"output_count":  len(ctx.OutputArgs),
		"has_context":   ctx.hasContext,
		"has_params":    ctx.hasParams,
		"has_return":    ctx.hasReturn,
		"has_error":     ctx.hasError,
	}
}
