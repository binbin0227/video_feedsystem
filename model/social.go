package model

import "time"

type Social struct {
	ID         int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	FollowerID int64     `gorm:"uniqueIndex:idx_follow" json:"follower_id"` // 粉丝
	VloggerID  int64     `gorm:"uniqueIndex:idx_follow" json:"vlogger_id"`  // 被关注的博主
	CreatedAt  time.Time `json:"created_at"`
}