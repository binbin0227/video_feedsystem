package model

import "time"

type Comment struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	VideoID   int64     `gorm:"index;not null" json:"video_id"` 
	AuthorID  int64     `gorm:"not null" json:"author_id"`
	Content   string    `gorm:"type:varchar(500);not null" json:"content"`
	CreatedAt time.Time `gorm:"index" json:"created_at"` 
}