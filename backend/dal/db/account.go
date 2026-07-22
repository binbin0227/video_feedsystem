package db

import (
	"context"
	"strings"
	"time"

	"video_feedsystem/model"
)

// AccountProfileRow 承接用户主页聚合查询结果。
type AccountProfileRow struct {
	AccountID         int64     `gorm:"column:account_id"`
	Username          string    `gorm:"column:username"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	VideoCount        int64     `gorm:"column:video_count"`
	ReceivedLikeCount int64     `gorm:"column:received_like_count"`
	FollowingCount    int64     `gorm:"column:following_count"`
	FollowerCount     int64     `gorm:"column:follower_count"`
}

// AccountSearchRow 承接用户搜索列表所需的账号和统计字段。
type AccountSearchRow struct {
	AccountID         int64  `gorm:"column:account_id"`
	Username          string `gorm:"column:username"`
	ReceivedLikeCount int64  `gorm:"column:received_like_count"`
	FollowerCount     int64  `gorm:"column:follower_count"`
}

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

// FindAccountProfile 查询账号基础信息及作品、获赞、关注和粉丝数量。
func FindAccountProfile(ctx context.Context, accountID int64) (*AccountProfileRow, error) {
	var row AccountProfileRow

	err := DB.WithContext(ctx).
		Table("accounts AS a").
		Select(`a.id AS account_id, a.username, a.created_at AS created_at, 
	(SELECT COUNT(*) FROM videos AS v WHERE v.author_id = a.id) AS video_count, 
	COALESCE((SELECT SUM(v.like_count) FROM videos AS v WHERE v.author_id = a.id),0) AS received_like_count, 
	(SELECT COUNT(*) FROM socials AS s WHERE s.follower_id = a.id) AS following_count, 
	(SELECT COUNT(*) FROM socials AS s WHERE s.vlogger_id = a.id) AS follower_count`).
		Where("a.id = ?", accountID).
		Take(&row).Error

	if err != nil {
		return nil, err
	}

	return &row, nil
}

// escapeLikeKeyword 转义 LIKE 模式中的特殊字符，避免用户输入被当成通配符。
func escapeLikeKeyword(keyword string) string {
	replacer := strings.NewReplacer(
		"!", "!!",
		"%", "!%",
		"_", "!_",
	)

	return replacer.Replace(keyword)
}

// SearchAccountsByUsername 按用户名模糊搜索，并优先返回粉丝数较多的账号。
func SearchAccountsByUsername(ctx context.Context, keyword string, limit int) ([]AccountSearchRow, error) {
	var rows []AccountSearchRow

	escapedKeyword := escapeLikeKeyword(keyword)
	pattern := "%" + escapedKeyword + "%"

	err := DB.WithContext(ctx).
		Table("accounts AS a").
		Select(`
			a.id AS account_id,
			a.username,

			COALESCE(
				(SELECT SUM(v.like_count)
				 FROM videos AS v
				 WHERE v.author_id = a.id),
				0
			) AS received_like_count,

			(SELECT COUNT(*)
			 FROM socials AS s
			 WHERE s.vlogger_id = a.id) AS follower_count
		`).
		Where("a.username LIKE ? ESCAPE '!'", pattern).
		Order("follower_count DESC").
		Order("a.id DESC").
		Limit(limit).
		Scan(&rows).Error

	return rows, err
}
