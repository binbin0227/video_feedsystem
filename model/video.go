package model

import "time"

type Video struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	AuthorID    int64     `gorm:"index;not null" json:"author_id"`
	Title       string    `gorm:"type:varchar(128);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	PlayURL     string    `gorm:"type:varchar(255);not null" json:"play_url"`
	CoverURL    string    `gorm:"type:varchar(255);not null" json:"cover_url"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"` // 刷视频时需要按时间倒序排,因此加上索引
	UpdatedAt   time.Time `json:"updated_at"`
	LikeCount   int       `gorm:"default:0" json:"like_count"`
	Popularity  int       `gorm:"default:0" json:"popularity"`
}