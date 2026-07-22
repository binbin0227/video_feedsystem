package handler

import (
	"errors"
	"strconv"
	"strings"

	"video_feedsystem/pkg/apperr"

	"github.com/cloudwego/hertz/pkg/app"
)

// getAccountID 从 JWT 中间件写入的上下文中取出用户 ID。
func getAccountID(c *app.RequestContext) (int64, error) {
	value, exists := c.Get("accountID")
	if !exists {
		return 0, apperr.New(apperr.KindUnauthorized, "用户未登录")
	}

	accountID, ok := value.(int64)
	if !ok {
		return 0, apperr.Wrap(apperr.KindInternal, "服务器内部错误，请稍后再试", errors.New("accountID 类型错误"))
	}
	return accountID, nil
}

// parsePositiveInt64Query 读取并校验必须大于 0 的 int64 查询参数。
func parsePositiveInt64Query(c *app.RequestContext, name string) (int64, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" {
		return 0, apperr.New(apperr.KindInvalid, "缺少 "+name+" 参数")
	}

	return parsePositiveInt64String(value, name)
}

// parsePositiveInt64String 将字符串 ID 转成大于 0 的 int64。
func parsePositiveInt64String(value, name string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, apperr.New(apperr.KindInvalid, name+" 不能为空")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperr.New(apperr.KindInvalid, name+" 不合法")
	}

	return id, nil
}

// parseOptionalCursor 解析可选游标，没传时返回 0。
func parseOptionalCursor(c *app.RequestContext) (int64, error) {
	value := strings.TrimSpace(c.Query("cursor"))
	if value == "" {
		return 0, nil
	}

	cursor, err := strconv.ParseInt(value, 10, 64)
	if err != nil || cursor < 0 {
		return 0, apperr.New(apperr.KindInvalid, "cursor 格式错误")
	}

	return cursor, nil
}

// parseOptionalLimit 解析可选数量，没传时返回 0。
func parseOptionalLimit(c *app.RequestContext) (int, error) {
	value := strings.TrimSpace(c.Query("limit"))
	if value == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return 0, apperr.New(apperr.KindInvalid, "limit 格式错误")
	}

	return limit, nil
}
