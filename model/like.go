package model

import "time"

type Like struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false;index:idx_account_like,priority:2" json:"id"`
	VideoID   int64     `gorm:"not null;uniqueIndex:idx_user_video" json:"video_id"`
	AccountID int64     `gorm:"not null;uniqueIndex:idx_user_video;index:idx_account_like,priority:1" json:"account_id"`
	CreatedAt time.Time `json:"created_at"`
}
