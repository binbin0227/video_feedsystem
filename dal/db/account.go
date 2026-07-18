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

// FindAccountByUsername 根据用户名查询账号。
func FindAccountByUsername(ctx context.Context, username string) (*model.Account, error) {
	var account model.Account
	err := DB.WithContext(ctx).Where("username = ?", username).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

// FindAccountByID 根据主键查询账号。
func FindAccountByID(ctx context.Context, accountID int64) (*model.Account, error) {
	var account model.Account
	err := DB.WithContext(ctx).First(&account, accountID).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
