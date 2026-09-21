package model

import "time"

type OutboxEvent struct {
	ID          int64      `gorm:"primaryKey;autoIncrement:false"`
	EventType   string     `gorm:"type:varchar(64);not null"`
	Payload     string     `gorm:"type:json;not null"`
	RetryCount  int        `gorm:"not null;default:0"`
	NextRetryAt time.Time  `gorm:"not null;index:idx_outbox_pending,priority:2"`
	PublishedAt *time.Time `gorm:"index:idx_outbox_pending,priority:1"`
	CreatedAt   time.Time
}
