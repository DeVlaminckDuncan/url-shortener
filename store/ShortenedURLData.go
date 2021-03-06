package store

// ShortenedURLData contains a list of ShortenedURLs and analytics (a list of ShortenedURLVisitsHistory)
type ShortenedURLData struct {
	ShortenedURLObject ShortenedURL `json:"shortenedURL"`
	Analytics          []string     `json:"analytics"`
}
