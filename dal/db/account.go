package db

import (
	"context"

	"video_feedsystem/model"
)

// CreateAccount 将新账号写入数据库。
func CreateAccount(ctx context.Context, account *model.Account) error {
	return DB.WithContext(ctx).Create(account).Error
}

// CheckUsernameExist 检查用户名是否已经存在。
func CheckUsernameExist(ctx context.Context, username string) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).Model(&model.Account{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// FindByUsername 根据用户名查询账号。
func FindByUsername(ctx context.Context, username string) (*model.Account, error) {
	var account model.Account
	err := DB.WithContext(ctx).Where("username = ?", username).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
