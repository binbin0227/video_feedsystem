package model

import "time"

type Comment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false;index:idx_video_comment,priority:2" json:"id"`
	VideoID   int64     `gorm:"not null;index:idx_video_comment,priority:1" json:"video_id"`
	AccountID int64     `gorm:"not null" json:"account_id"`
	Content   string    `gorm:"type:varchar(500);not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`

	Video   Video   `gorm:"foreignKey:VideoID;references:ID;constraint:OnDelete:CASCADE;" json:"-"`
	Account Account `gorm:"foreignKey:AccountID;references:ID;" json:"-"`
}
