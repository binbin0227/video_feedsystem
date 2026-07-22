package model

import "time"


type Like struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false;index:idx_account_like,priority:2" json:"id"`
	VideoID   int64     `gorm:"not null;uniqueIndex:idx_user_video" json:"video_id"`
	AccountID int64     `gorm:"not null;uniqueIndex:idx_user_video;index:idx_account_like,priority:1" json:"account_id"`
	CreatedAt time.Time `json:"created_at"`

	// Video 和 Account 用于建立外键；视频或账号删除后，其点赞记录随之删除。
	Video   Video   `gorm:"foreignKey:VideoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
