package store

// UserShortenedURL is a link between ShortenedURL and User
type UserShortenedURL struct {
	UserID         string `xorm:"not null"`
	ShortenedURLID string `xorm:"not null"`
}