package store

import "time"

// ShortenedURL contains a short URL with its associated data
type ShortenedURL struct {
	ID        string    `xorm:"pk not null unique"`
	Name      string    `xorm:"not null"`
	CreatedAt time.Time `xorm:"not null default CURRENT_TIMESTAMP created"`
	ShortURL  string    `xorm:"not null unique"`
	LongURL   string    `xorm:"not null"`
}
