package model

import "time"

// Account 表示账号表；用户名由唯一索引保证不能重复，Password 只保存 bcrypt 哈希。
type Account struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Username  string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
