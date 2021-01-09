package store

// ShortenedURLData contains a list of ShortenedURLs and analytics (a list of ShortenedURLVisitsHistory)
type ShortenedURLData struct {
	ShortenedURLObject ShortenedURL
	Analytics          []string
}
