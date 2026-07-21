package model

import "time"

type Video struct {
	ID          int64     `gorm:"primaryKey;autoIncrement:false;index:idx_author_video,priority:2" json:"id"`
	AuthorID    int64     `gorm:"not null;index:idx_author_video,priority:1" json:"author_id"`
	Title       string    `gorm:"type:varchar(128);not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	PlayURL     string    `gorm:"type:varchar(255);not null" json:"play_url"`
	CoverURL    string    `gorm:"type:varchar(255);not null" json:"cover_url"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LikeCount   int       `gorm:"default:0" json:"like_count"`
	Popularity  int       `gorm:"default:0" json:"popularity"`
}
