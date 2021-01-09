package store

import "time"

// ShortenedURL contains a short URL with its associated data
type ShortenedURL struct {
	ID        string    `json:"id" xorm:"pk not null unique"`
	Name      string    `json:"name" xorm:"not null"`
	CreatedAt time.Time `json:"createdAt" xorm:"not null default CURRENT_TIMESTAMP created"`
	ShortURL  string    `json:"shortURL" xorm:"not null unique"`
	LongURL   string    `json:"longURL" xorm:"not null"`
}
