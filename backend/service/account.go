package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AccountSearchItem 表示用户搜索结果中的业务字段。
type AccountSearchItem struct {
	AccountID         int64
	Username          string
	ReceivedLikeCount int64
	FollowerCount     int64
}

// AccountProfile 表示用户主页需要展示的基础信息和统计数据。
type AccountProfile struct {
	AccountID         int64
	Username          string
	CreatedAt         time.Time
	VideoCount        int64
	ReceivedLikeCount int64
	FollowingCount    int64
	FollowerCount     int64
}

const (
	maxUsernameLength  = 32
	minPasswordLength  = 8
	maxPasswordLength  = 72
	accountSearchLimit = 20
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
	exists, err := db.CheckUsernameExist(ctx, username)
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "注册失败，请稍后再试", err)
	}
	if exists {
		return apperr.New(apperr.KindConflict, "用户名已被注册")
	}
	accountID, err := utils.GenerateID()
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "生成账号 ID 失败", err)
	}
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
func Login(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", apperr.New(apperr.KindInvalid, "用户名或密码不能为空")
	}
	account, err := db.FindAccountByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperr.New(apperr.KindUnauthorized, "用户名或密码错误")
		}
		return "", apperr.Wrap(apperr.KindInternal, "登录失败，请稍后再试", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return "", apperr.New(apperr.KindUnauthorized, "用户名或密码错误")
	}
	token, err := utils.GenerateToken(account.ID)
	if err != nil {
		return "", apperr.Wrap(apperr.KindInternal, "生成登录凭证失败，请稍后再试", err)
	}
	return token, nil
}

// GetAccountProfile 查询指定账号的主页信息。
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

// SearchAccounts 根据用户名关键词搜索账号。
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
