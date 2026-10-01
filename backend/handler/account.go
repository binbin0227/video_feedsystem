package handler

import (
	"net/http"

	"video_feedsystem/pkg/apperr"
	"video_feedsystem/pkg/httpx"
	"video_feedsystem/service"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthTokensResponse struct {
	AccessToken                 string `json:"access_token"`
	RefreshToken                string `json:"refresh_token"`
	AccessTokenExpiresInSeconds int    `json:"expires_in"`
}

// 处理用户注册请求
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	if err := service.Register(c.Request.Context(), req.Username, req.Password); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "账号注册成功！"})
}

// 处理用户登录请求并返回 JWT
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	result, err := service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, AuthTokensResponse{
		AccessToken:                 result.AccessToken,
		RefreshToken:                result.RefreshToken,
		AccessTokenExpiresInSeconds: 1800,
	})
}

func RefreshAuthTokens(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	result, err := service.RefreshAuthTokens(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, AuthTokensResponse{
		AccessToken:                 result.AccessToken,
		RefreshToken:                result.RefreshToken,
		AccessTokenExpiresInSeconds: 1800,
	})
}

func Logout(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteError(c, apperr.New(apperr.KindInvalid, "JSON 解析失败"))
		return
	}

	accountID, err := getAccountID(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	accessTokenID, accessTokenExpiresAt, err := getAccessTokenMetadata(c)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	if err := service.Logout(c.Request.Context(), accountID, accessTokenID, accessTokenExpiresAt, req.RefreshToken); err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "退出登录成功"})
}

// 返回指定账号的主页信息
func GetAccountProfile(c *gin.Context) {
	accountID, err := parsePositiveInt64Query(c, "account_id")
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	profile, err := service.GetAccountProfile(c.Request.Context(), accountID)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"profile": newAccountProfileResponse(profile),
	})
}

// 根据用户名关键词搜索用户
func SearchAccounts(c *gin.Context) {
	keyword := c.Query("keyword")
	accounts, err := service.SearchAccounts(c.Request.Context(), keyword)
	if err != nil {
		httpx.WriteError(c, err)
		return
	}

	c.JSON(http.StatusOK, AccountSearchResponse{
		Accounts: newAccountSearchListResponse(accounts),
	})
}
