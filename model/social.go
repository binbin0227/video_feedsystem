package model

import "time"

type Social struct {
	ID         int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	FollowerID int64     `gorm:"not null;uniqueIndex:idx_follow" json:"follower_id"` // 粉丝
	VloggerID  int64     `gorm:"not null;uniqueIndex:idx_follow;index" json:"vlogger_id"`  // 被关注的博主
	CreatedAt  time.Time `json:"created_at"`
}
