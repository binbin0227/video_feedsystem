package model

import "time"

type Social struct {
	ID         int64     `gorm:"primaryKey;autoIncrement:false;index:idx_follower_relation,priority:2;index:idx_vlogger_relation,priority:2" json:"id"`
	FollowerID int64     `gorm:"not null;uniqueIndex:idx_follow;index:idx_follower_relation,priority:1" json:"follower_id"`
	VloggerID  int64     `gorm:"not null;uniqueIndex:idx_follow;index:idx_vlogger_relation,priority:1" json:"vlogger_id"`
	CreatedAt  time.Time `json:"created_at"`

	// Follower 和 Vlogger 分别建立指向 accounts.id 的外键，账号删除后同步清理关注关系。
	Follower Account `gorm:"foreignKey:FollowerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Vlogger  Account `gorm:"foreignKey:VloggerID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
