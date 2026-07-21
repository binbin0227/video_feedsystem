package model

import "time"

type Social struct {
	ID int64 `gorm:"primaryKey;autoIncrement:false;index:idx_follower_relation,priority:2;index:idx_vlogger_relation,priority:2" json:"id"`
	FollowerID int64 `gorm:"not null;uniqueIndex:idx_follow;index:idx_follower_relation,priority:1" json:"follower_id"`
	VloggerID int64 `gorm:"not null;uniqueIndex:idx_follow;index:idx_vlogger_relation,priority:1" json:"vlogger_id"`
	CreatedAt time.Time `json:"created_at"`
}
