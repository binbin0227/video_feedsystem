package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"video_feedsystem/dal/db"
	"video_feedsystem/dal/redis"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AccountSearchItem struct {
	AccountID         int64
	Username          string
	ReceivedLikeCount int64
	FollowerCount     int64
}

type AccountProfile struct {
	AccountID         int64
	Username          string
	CreatedAt         time.Time
	VideoCount        int64
	ReceivedLikeCount int64
	FollowingCount    int64
	FollowerCount     int64
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

const (
	maxUsernameLength  = 32
	minPasswordLength  = 8
	maxPasswordLength  = 72
	accountSearchLimit = 20
	refreshTokenTTL    = 7 * 24 * time.Hour
)

// Register 校验注册信息、加密密码并创建账号。
func Register(ctx context.Context, username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return apperr.New(apperr.KindInvalid, "用户名或密码不能为空")
	}
	if utf8.RuneCountInString(username) > maxUsernameLength {
		return apperr.New(apperr.KindInvalid, "用户名不能超过 32 个字符")
	}
	passwordLength := len([]byte(password))
	if passwordLength < minPasswordLength {
		return apperr.New(apperr.KindInvalid, "密码不能少于 8 个字节")
	}
	if passwordLength > maxPasswordLength {
		return apperr.New(apperr.KindInvalid, "密码不能超过 72 个字节")
	}

	// 检查用户名是否已被注册
	exist, err := db.CheckUsernameExist(ctx, username)
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "注册失败，请稍后再试", err)
	}
	if exist {
		return apperr.New(apperr.KindConflict, "用户名已被注册")
	}

	// 生成账号 ID
	accountID, err := utils.GenerateID()
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "生成账号 ID 失败", err)
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "密码加密失败", err)
	}

	account := &model.Account{
		ID:       accountID,
		Username: username,
		Password: string(hashedPassword),
	}
	if err := db.CreateAccount(ctx, account); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return apperr.New(apperr.KindConflict, "用户名已被注册")
		}
		return apperr.Wrap(apperr.KindInternal, "注册失败，请稍后再试", err)
	}

	return nil
}

// Login 校验用户名和密码，成功后签发 JWT。
func Login(ctx context.Context, username, password string) (*AuthTokens, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, apperr.New(apperr.KindInvalid, "用户名或密码不能为空")
	}

	// 检查用户是否存在
	account, err := db.FindAccountByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.KindUnauthorized, "用户名或密码错误")
		}
		return nil, apperr.Wrap(apperr.KindInternal, "登录失败，请稍后再试", err)
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return nil, apperr.New(apperr.KindUnauthorized, "用户名或密码错误")
	}

	accessToken, err := utils.GenerateAccessToken(account.ID)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成登录凭证失败，请稍后再试", err)
	}

	refreshToken, refreshTokenHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成刷新凭证失败，请稍后再试", err)
	}

	if err := redis.SaveRefreshToken(ctx, refreshTokenHash, account.ID, refreshTokenTTL); err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "保存登录状态失败，请稍后再试", err)
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func RefreshAuthTokens(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, apperr.New(apperr.KindInvalid, "刷新凭证不能为空")
	}

	refreshTokenHash := utils.HashRefreshToken(refreshToken)
	accountID, found, err := redis.FindRefreshTokenAccountID(ctx, refreshTokenHash)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "刷新登录状态失败，请稍后再试", err)
	}
	if !found {
		return nil, apperr.New(apperr.KindUnauthorized, "刷新凭证无效或已过期")
	}

	accessToken, err := utils.GenerateAccessToken(accountID)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成登录凭证失败，请稍后再试", err)
	}

	newRefreshToken, newRefreshTokenHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "生成刷新凭证失败，请稍后再试", err)
	}

	rotated, err := redis.RotateRefreshToken(ctx, refreshTokenHash, newRefreshTokenHash, accountID, refreshTokenTTL)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "刷新登录状态失败，请稍后再试", err)
	}
	if !rotated {
		return nil, apperr.New(apperr.KindUnauthorized, "刷新凭证无效或已过期")
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func Logout(ctx context.Context, accountID int64, accessTokenID string, accessTokenExpiresAt time.Time, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return apperr.New(apperr.KindInvalid, "刷新凭证不能为空")
	}
	if accountID <= 0 || accessTokenID == "" {
		return apperr.New(apperr.KindUnauthorized, "登录状态无效")
	}

	accessTokenBlacklistTTL := time.Until(accessTokenExpiresAt)
	if accessTokenBlacklistTTL <= 0 {
		return apperr.New(apperr.KindUnauthorized, "Token 已过期或无效")
	}

	refreshTokenHash := utils.HashRefreshToken(refreshToken)
	if err := redis.RevokeLoginSession(ctx, refreshTokenHash, accountID, accessTokenID, accessTokenBlacklistTTL); err != nil {
		return apperr.Wrap(apperr.KindInternal, "退出登录失败，请稍后再试", err)
	}

	return nil
}

// 查询指定账号的主页信息。
func GetAccountProfile(ctx context.Context, accountID int64) (*AccountProfile, error) {
	if accountID <= 0 {
		return nil, apperr.New(apperr.KindInvalid, "用户ID不合法")
	}

	row, err := db.FindAccountProfile(ctx, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.New(apperr.KindNotFound, "用户不存在")
		}
		return nil, apperr.Wrap(apperr.KindInternal, "查询用户主页失败，请稍后再试", err)
	}

	return &AccountProfile{
		AccountID:         row.AccountID,
		Username:          row.Username,
		CreatedAt:         row.CreatedAt,
		VideoCount:        row.VideoCount,
		ReceivedLikeCount: row.ReceivedLikeCount,
		FollowingCount:    row.FollowingCount,
		FollowerCount:     row.FollowerCount,
	}, nil
}

// 根据用户名关键词搜索账号
func SearchAccounts(ctx context.Context, keyword string) ([]AccountSearchItem, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, apperr.New(apperr.KindInvalid, "搜索关键词不能为空")
	}

	rows, err := db.SearchAccountsByUsername(ctx, keyword, accountSearchLimit)
	if err != nil {
		return nil, apperr.Wrap(apperr.KindInternal, "搜索用户失败，请稍后再试", err)
	}

	accounts := make([]AccountSearchItem, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, AccountSearchItem{
			AccountID:         row.AccountID,
			Username:          row.Username,
			ReceivedLikeCount: row.ReceivedLikeCount,
			FollowerCount:     row.FollowerCount,
		})
	}

	return accounts, nil
}
