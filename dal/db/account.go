package db

import (
	"context"
	"video_feedsystem/model"
)

// 将新用户信息写入数据库
func CreateAccount(ctx context.Context, account *model.Account) error {
	return DB.WithContext(ctx).Create(account).Error
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

// 根据用户名去数据库查询整条记录
func FindByUsername(ctx context.Context, username string) (*model.Account, error) {
	var account model.Account
	err := DB.WithContext(ctx).Where("username = ?", username).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
