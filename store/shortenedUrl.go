package store

import (
	"time"

	"github.com/google/uuid"
)

type ShortenedUrl struct {
	Id        uuid.UUID
	CreatedAt time.Time `xorm:"created"`
	ShortUrl  string
	LongUrl   string
	Visits    int
}
