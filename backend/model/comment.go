package model

import "time"

// Comment 表示评论表；idx_video_comment 支持按视频和评论 ID 倒序分页。
type Comment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false;index:idx_video_comment,priority:2" json:"id"`
	VideoID   int64     `gorm:"not null;index:idx_video_comment,priority:1" json:"video_id"`
	AccountID int64     `gorm:"not null" json:"account_id"`
	Content   string    `gorm:"type:varchar(500);not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`

	// Video 和 Account 只用于建立外键及按需预加载，不直接序列化到接口响应。
	Video   Video   `gorm:"foreignKey:VideoID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Account Account `gorm:"foreignKey:AccountID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
