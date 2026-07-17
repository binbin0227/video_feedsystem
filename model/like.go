package model

import "time"

type Like struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	VideoID   int64     `gorm:"not null;uniqueIndex:idx_user_video" json:"video_id"` // 联合唯一索引：防重复点赞
	AccountID int64     `gorm:"not null;uniqueIndex:idx_user_video;index" json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}
