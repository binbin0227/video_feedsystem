package model

import "time"

type Video struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false;index:idx_author_video,priority:2" json:"id"`
	AuthorID    int64     `gorm:"not null;index:idx_author_video,priority:1;uniqueIndex:idx_author_media,priority:1" json:"author_id"`
	Title       string    `gorm:"type:varchar(128);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	PlayURL     string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_author_media,priority:2" json:"play_url"`
	CoverURL    string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_author_media,priority:3" json:"cover_url"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LikeCount   int       `gorm:"default:0" json:"like_count"`
	Popularity  int       `gorm:"default:0" json:"popularity"`

	// Author 建立 videos.author_id 到 accounts.id 的真实外键关联。
	Author Account `gorm:"foreignKey:AuthorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
}
