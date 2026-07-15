package db

import (
	"context"
	"video_feedsystem/model"
)

// 将新用户信息写入数据库
func CreateAccount(ctx context.Context, account *model.Account) error {
	err := DB.WithContext(ctx).Create(account).Error
	return err
}

// 检查用户名是否已经被注册
func CheckUsernameExist(ctx context.Context, username string) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).Model(&model.Account{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
