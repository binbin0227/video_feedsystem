package httpx

import (
	"errors"
	"log"
	"net/http"

	"video_feedsystem/pkg/apperr"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 是所有失败请求共用的 JSON 响应结构。
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError 统一记录错误日志并返回 JSON。
func WriteError(c *gin.Context, err error) {
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		appErr = apperr.Wrap(apperr.KindInternal, "服务器内部错误，请稍后再试", err)
	}

	if appErr.Kind == apperr.KindInternal {
		loggedError := appErr.Cause
		if loggedError == nil {
			loggedError = err
		}
		log.Printf("path=%s, error=%v", c.Request.URL.Path, loggedError)
	}

	c.JSON(statusFromKind(appErr.Kind), ErrorResponse{
		Code:    string(appErr.Kind),
		Message: appErr.Message,
	})
}

// statusFromKind 将业务错误类别映射为 HTTP 状态码。
func statusFromKind(kind apperr.Kind) int {
	switch kind {
	case apperr.KindInvalid:
		return http.StatusBadRequest
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
