package db

import (
	"context"
	"errors"
	"time"

	"video_feedsystem/model"
)

// ErrFollowNotFound 表示取消关注时没有找到对应关系。
var ErrFollowNotFound = errors.New("follow relation not found")

// SocialAccountRow 承接关注关系和账号信息的 JOIN 查询结果。
type SocialAccountRow struct {
	RelationID int64     `gorm:"column:relation_id"`
	AccountID  int64     `gorm:"column:account_id"`
	Username   string    `gorm:"column:username"`
	FollowedAt time.Time `gorm:"column:followed_at"`
}

// CreateFollow 将关注关系写入数据库。
func CreateFollow(ctx context.Context, social *model.Social) error {
	return DB.WithContext(ctx).Create(social).Error
}

// DeleteFollow 删除指定关注者与被关注者之间的关系。
func DeleteFollow(ctx context.Context, followerID, vloggerID int64) error {
	result := DB.WithContext(ctx).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Delete(&model.Social{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFollowNotFound
	}

	return nil
}

// CheckFollowExist 判断指定关注关系是否存在。
func CheckFollowExist(ctx context.Context, followerID, vloggerID int64) (bool, error) {
	var count int64
	err := DB.WithContext(ctx).
		Model(&model.Social{}).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Count(&count).Error

	return count > 0, err
}

// ListFollowingAccounts 按关注关系 ID 倒序查询当前用户关注的账号。
func ListFollowingAccounts(ctx context.Context, followerID, cursor int64, limit int) ([]SocialAccountRow, error) {
	var rows []SocialAccountRow

	query := DB.WithContext(ctx).
		Table("socials AS s").
		Select("s.id AS relation_id, s.vlogger_id AS account_id, a.username, s.created_at AS followed_at").
		Joins("JOIN accounts AS a ON a.id = s.vlogger_id").
		Where("s.follower_id = ?", followerID).
		Order("s.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("s.id < ?", cursor)
	}

	err := query.Scan(&rows).Error
	return rows, err
}

// ListFollowerAccounts 按关注关系 ID 倒序查询当前用户的粉丝账号。
func ListFollowerAccounts(ctx context.Context, vloggerID, cursor int64, limit int) ([]SocialAccountRow, error) {
	var rows []SocialAccountRow

	query := DB.WithContext(ctx).
		Table("socials AS s").
		Select("s.id AS relation_id, s.follower_id AS account_id, a.username, s.created_at AS followed_at").
		Joins("JOIN accounts AS a ON a.id = s.follower_id").
		Where("s.vlogger_id = ?", vloggerID).
		Order("s.id DESC").
		Limit(limit)

	if cursor > 0 {
		query = query.Where("s.id < ?", cursor)
	}

	err := query.Scan(&rows).Error
	return rows, err
}
