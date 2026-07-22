package httpx

import (
	"context"
	"errors"

	"video_feedsystem/pkg/apperr"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ErrorResponse 是所有失败请求共用的 JSON 响应结构。
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError 统一记录错误日志并返回 JSON。
func WriteError(ctx context.Context, c *app.RequestContext, err error) {
	var appErr *apperr.AppError
	if !errors.As(err, &appErr) {
		appErr = apperr.Wrap(apperr.KindInternal, "服务器内部错误，请稍后再试", err)
	}

	if appErr.Kind == apperr.KindInternal {
		loggedError := appErr.Cause
		if loggedError == nil {
			loggedError = err
		}
		hlog.CtxErrorf(ctx, "path=%s, error=%v", c.Path(), loggedError)
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
		return consts.StatusBadRequest
	case apperr.KindUnauthorized:
		return consts.StatusUnauthorized
	case apperr.KindForbidden:
		return consts.StatusForbidden
	case apperr.KindNotFound:
		return consts.StatusNotFound
	case apperr.KindConflict:
		return consts.StatusConflict
	default:
		return consts.StatusInternalServerError
	}
}
