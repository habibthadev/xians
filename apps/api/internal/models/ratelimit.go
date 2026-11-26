package models

import "time"

// RateLimit is a fixed-window counter persisted in Postgres so limits hold
// across serverless instances and cold starts.
type RateLimit struct {
	Key         string    `gorm:"primaryKey;type:text"`
	WindowStart time.Time `gorm:"primaryKey;type:timestamptz"`
	Count       int       `gorm:"not null;default:0"`
}

func (RateLimit) TableName() string {
	return "rate_limits"
}
