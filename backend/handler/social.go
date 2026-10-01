package handler

import (
	"net/http"
	"strconv"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/gin-gonic/gin"
)

type FollowRequest struct {
	VloggerID string `json:"vlogger_id"`
}

// 让当前登录用户关注目标账号
func FollowUser(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}
	vloggerID, err := parsePositiveInt64String(req.VloggerID, "vlogger_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	if err := service.FollowUser(c.Request.Context(), followerID, vloggerID); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "关注成功",
	})
}

// 取消当前登录用户对目标账号的关注
func UnfollowUser(c *gin.Context) {
	var req FollowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}
	vloggerID, err := parsePositiveInt64String(req.VloggerID, "vlogger_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	if err := service.UnfollowUser(c.Request.Context(), followerID, vloggerID); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "取消关注成功",
	})
}

// 返回当前登录用户对目标账号的关注状态
func GetFollowStatus(c *gin.Context) {
	vloggerID, err := parsePositiveInt64Query(c, "vlogger_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	following, err := service.CheckFollowStatus(c.Request.Context(), followerID, vloggerID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]bool{
		"is_following": following,
	})
}

// 分页查询当前用户关注的账号
func GetFollowingList(c *gin.Context) {
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	followerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	result, err := service.GetFollowingList(c.Request.Context(), followerID, cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	c.JSON(http.StatusOK, FollowingOrFollowerListResponse{
		Accounts:   newFollowingOrFollowerAccountListResponse(result.Accounts),
		NextCursor: nextCursor,
		HasMore:    result.HasMore,
	})
}

// 分页查询当前用户粉丝的账号
func GetFollowerList(c *gin.Context) {
	cursor, err := parseOptionalCursor(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	limit, err := parseOptionalLimit(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	vloggerID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	result, err := service.GetFollowerList(c.Request.Context(), vloggerID, cursor, limit)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}
	nextCursor := ""
	if result.NextCursor > 0 {
		nextCursor = strconv.FormatInt(result.NextCursor, 10)
	}

	c.JSON(http.StatusOK, FollowingOrFollowerListResponse{
		Accounts:   newFollowingOrFollowerAccountListResponse(result.Accounts),
		HasMore:    result.HasMore,
		NextCursor: nextCursor,
	})
}
