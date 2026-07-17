package apperr

// Kind 表示错误类别，HTTP 层会根据它选择状态码。
type Kind string

const (
	KindInvalid      Kind = "INVALID_ARGUMENT"
	KindUnauthorized Kind = "UNAUTHORIZED"
	KindForbidden    Kind = "FORBIDDEN"
	KindNotFound     Kind = "NOT_FOUND"
	KindConflict     Kind = "CONFLICT"
	KindInternal     Kind = "INTERNAL_ERROR"
)

// AppError 同时保存给用户看的信息和供开发者排查的底层错误。
type AppError struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	return e.Message
}

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
