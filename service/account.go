package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/pkg/apperr"
	"video_feedsystem/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	maxUsernameLength = 32
	minPasswordLength = 8
	maxPasswordLength = 72
)

// Register 校验注册信息、加密密码并创建账号。
func Register(ctx context.Context, username, password string) error {
	// 1. 校验合法性
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

	// 2. 检查用户名是否已被注册
	exists, err := db.CheckUsernameExist(ctx, username)
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "注册失败，请稍后再试", err)
	}
	if exists {
		return apperr.New(apperr.KindConflict, "用户名已被注册")
	}

	// 3. 使用雪花算法生成账号 ID
	accountID, err := utils.GenerateID()
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "生成账号 ID 失败", err)
	}

	// 4. 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return apperr.Wrap(apperr.KindInternal, "密码加密失败", err)
	}

	// 5. 打包并调用 db.CreateAccount
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

func Login(ctx context.Context, username, password string) (string, error) {
	// 1. 校验合法性
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return "", apperr.New(apperr.KindInvalid, "用户名或密码不能为空")
	}

	// 2. 调用 db.FindByUsername 查询账号
	account, err := db.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperr.New(apperr.KindUnauthorized, "用户名或密码错误")
		}
		return "", apperr.Wrap(apperr.KindInternal, "登录失败，请稍后再试", err)
	}

	// 3. 将加密后的前端密码与数据库密码进行比对
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)); err != nil {
		return "", apperr.New(apperr.KindUnauthorized, "用户名或密码错误")
	}

	// 4. 返回 token
	token, err := utils.GenerateToken(account.ID)
	if err != nil {
		return "", apperr.Wrap(apperr.KindInternal, "生成登录凭证失败，请稍后再试", err)
	}
	return token, nil
}
