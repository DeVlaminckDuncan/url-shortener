package store

import "github.com/google/uuid"

type UserShortenedUrl struct {
	UserId         uuid.UUID
	ShortenedUrlId uuid.UUID
}
