package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode 错误代码
type ErrorCode int

const (
	// ErrUnknown 未知错误
	ErrUnknown ErrorCode = iota
	// ErrNotFound 未找到
	ErrNotFound
	// ErrInvalidArg 无效参数
	ErrInvalidArg
	// ErrPermission 权限错误
	ErrPermission
	// ErrTimeout 超时
	ErrTimeout
	// ErrConnection 连接错误
	ErrConnection
	// ErrAuth 认证错误
	ErrAuth
	// ErrRateLimit 限流
	ErrRateLimit
	// ErrInternal 内部错误
	ErrInternal
)

// String 返回错误代码字符串
func (c ErrorCode) String() string {
	switch c {
	case ErrUnknown:
		return "UNKNOWN"
	case ErrNotFound:
		return "NOT_FOUND"
	case ErrInvalidArg:
		return "INVALID_ARG"
	case ErrPermission:
		return "PERMISSION"
	case ErrTimeout:
		return "TIMEOUT"
	case ErrConnection:
		return "CONNECTION"
	case ErrAuth:
		return "AUTH"
	case ErrRateLimit:
		return "RATE_LIMIT"
	case ErrInternal:
		return "INTERNAL"
	default:
		return "UNKNOWN"
	}
}

// HelixError Helix 错误
type HelixError struct {
	Code    ErrorCode
	Message string
	Err     error
	Stack   string
}

// Error 实现 error 接口
func (e *HelixError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *HelixError) Unwrap() error {
	return e.Err
}

// New 创建新错误
func New(code ErrorCode, message string) *HelixError {
	return &HelixError{
		Code:    code,
		Message: message,
		Stack:   getStack(),
	}
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message string) *HelixError {
	return &HelixError{
		Code:    code,
		Message: message,
		Err:     err,
		Stack:   getStack(),
	}
}

// Wrapf 包装错误（格式化）
func Wrapf(err error, code ErrorCode, format string, args ...any) *HelixError {
	return &HelixError{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
		Stack:   getStack(),
	}
}

// Is 检查错误代码
func Is(err error, code ErrorCode) bool {
	var helixErr *HelixError
	if As(err, &helixErr) {
		return helixErr.Code == code
	}
	return false
}

// As 转换为 HelixError
func As(err error, target **HelixError) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(*HelixError); ok {
		*target = e
		return true
	}

	// 检查 Unwrap
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return As(u.Unwrap(), target)
	}

	return false
}

// getStack 获取调用栈
func getStack() string {
	var pcs [10]uintptr
	n := runtime.Callers(3, pcs[:])

	var sb strings.Builder
	for i := 0; i < n; i++ {
		fn := runtime.FuncForPC(pcs[i])
		if fn != nil {
			file, line := fn.FileLine(pcs[i])
			sb.WriteString(fmt.Sprintf("%s:%d\n", file, line))
		}
	}
	return sb.String()
}

// 常用错误构造函数

func NotFound(message string) *HelixError {
	return New(ErrNotFound, message)
}

func NotFoundf(format string, args ...any) *HelixError {
	return New(ErrNotFound, fmt.Sprintf(format, args...))
}

func InvalidArg(message string) *HelixError {
	return New(ErrInvalidArg, message)
}

func InvalidArgf(format string, args ...any) *HelixError {
	return New(ErrInvalidArg, fmt.Sprintf(format, args...))
}

func Permission(message string) *HelixError {
	return New(ErrPermission, message)
}

func Timeout(message string) *HelixError {
	return New(ErrTimeout, message)
}

func Connection(message string) *HelixError {
	return New(ErrConnection, message)
}

func Auth(message string) *HelixError {
	return New(ErrAuth, message)
}

func RateLimit(message string) *HelixError {
	return New(ErrRateLimit, message)
}

func Internal(message string) *HelixError {
	return New(ErrInternal, message)
}

func Internalf(format string, args ...any) *HelixError {
	return New(ErrInternal, fmt.Sprintf(format, args...))
}
