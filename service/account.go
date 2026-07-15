package service

import (
	"context"
	"errors"
	"video_feedsystem/dal/db"
	"video_feedsystem/model"
	"video_feedsystem/utils"

	"golang.org/x/crypto/bcrypt"
)

// 用户注册 业务逻辑
func RegisterUser(ctx context.Context, username, password string) error {
	// 1. 查重
	isExist, err := db.CheckUsernameExist(ctx, username)
	if err != nil {
		return err // err: 数据库崩溃
	}
	if isExist {
		return errors.New("用户名已被注册，请换一个")
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
