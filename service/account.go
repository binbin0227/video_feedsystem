package service

import (
	"context"
	"errors"
	"strings"
	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/utils"

	"golang.org/x/crypto/bcrypt"
)

var ErrUsernameTaken = errors.New("用户名已被注册")

// 新用户注册
func Register(ctx context.Context, username, password string) error {
	// 0. 格式化
	username = strings.TrimSpace(username)
	// 1. 查重
	isExist, err := db.CheckUsernameExist(ctx, username)
	if err != nil {
		return err // err: 数据库崩溃
	}
	if isExist {
		return ErrUsernameTaken
	}

	// 2. 分配 ID
	accountID, err := utils.GenerateID()
	if err != nil {
		return errors.New("服务器发号器异常，请稍后再试")
	}

	// 3. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密异常")
	}

	// 4. 将数据打包进结构体
	account := &model.Account{
		ID:       accountID,
		Username: username,
		Password: string(hashedPassword),
	}

	// 5. 存入数据库
	return db.CreateAccount(ctx, account)
}

// 用户登录
func Login(ctx context.Context, username, password string) (string, error) {
	// 1. 查询用户信息
	account, err := db.FindByUsername(ctx, username)
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 2. 校验密码
	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 3. 登录成功则返回 token
	token, err := utils.GenerateToken(account.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
