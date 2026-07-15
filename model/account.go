package model

import "time"

type Account struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"` 
	Username  string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}