package store

import "time"

// ShortenedURLVisitsHistory contains the datetime of every time someone visits a short URL
type ShortenedURLVisitsHistory struct {
	ShortenedURLID string    `xorm:"not null"`
	VisitedAt      time.Time `xorm:"not null default CURRENT_TIMESTAMP created"`
}
