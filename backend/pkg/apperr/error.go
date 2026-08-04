package apperr

// Kind 表示错误类别，HTTP 层会根据它选择状态码。
type Kind string

const (
	KindInvalid         Kind = "INVALID_ARGUMENT"
	KindUnauthorized    Kind = "UNAUTHORIZED"
	KindForbidden       Kind = "FORBIDDEN"
	KindNotFound        Kind = "NOT_FOUND"
	KindConflict        Kind = "CONFLICT"
	KindInternal        Kind = "INTERNAL_ERROR"
	KindTooManyRequests Kind = "TOO_MANY_REQUESTS"
)

// AppError 同时保存给用户看的信息和供开发者排查的底层错误。
type AppError struct {
	Kind    Kind
	Message string
	Cause   error
}

// Error 返回可以展示给调用方的业务错误信息。
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap 返回底层错误，供 errors.Is 和 errors.As 继续判断。
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New 创建普通业务错误，例如参数错误、资源不存在。
func New(kind Kind, message string) *AppError {
	return &AppError{Kind: kind, Message: message}
}

// Wrap 包装数据库等底层错误，Cause 会被记录到日志。
func Wrap(kind Kind, message string, cause error) *AppError {
	return &AppError{Kind: kind, Message: message, Cause: cause}
}
