package handler

import (
	"net/http"

	"github.com/devlaminckduncan/url-shortener/shortener"
	"github.com/devlaminckduncan/url-shortener/store"
	"github.com/gin-gonic/gin"
)

type urlCreationRequest struct {
	Name    string `json:"name" binding:"required"`
	LongURL string `json:"longURL" binding:"required"`
	UserID  string `json:"userID" binding:"required"`
}

type userLoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

// TODO: check every request for empty values
// TODO: add codes to JSON data eg "statusCode": "URL_CREATED"

// CreateShortURL takes a name, a long URL and a user ID and creates a new ShortenedURL
func CreateShortURL(c *gin.Context) {
	var creationRequest urlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL := shortener.GenerateShortURL(creationRequest.LongURL, creationRequest.UserID)
	statusCode, err := store.SaveURL(shortURL, creationRequest.Name, creationRequest.LongURL, creationRequest.UserID)
	if statusCode != "" || err != nil {
		status := http.StatusBadRequest
		if statusCode == "NON_EXISTING_USER" {
			status = http.StatusUnauthorized
		}

		c.JSON(status, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":  "Short URL created successfully",
		"shortURL": "http://" + c.Request.Host + "/" + shortURL,
	})
}

// RedirectShortURL takes a short URL redirects you to the long URL from the database and creates a new ShortenedURLVisitsHistory
func RedirectShortURL(c *gin.Context) {
	shortURL := c.Request.URL.Path[1:]
	longURL := store.GetLongURL(shortURL)

	c.Redirect(301, longURL)
}

// GetUserShortenedURLs takes a user ID and returns the user's ShortenedURLs
func GetUserShortenedURLs(c *gin.Context) {
	userID := c.Request.URL.Path[12:]

	urls, statusCode, err := store.GetUserShortenedURLs(userID)
	if statusCode != "" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, urls)
}

// CreateUser takes a first name, a last name, a username, an email and a password and creates a new User and returns a new user token
func CreateUser(c *gin.Context) {
	var user store.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, statusCode, err := store.SaveUser(user)
	if statusCode != "" || err != nil {
		c.JSON(401, gin.H{
			"message":    "Something went wrong",
			"statusCode": statusCode,
			"error":      err,
		})
	} else {
		c.JSON(200, gin.H{
			"message": "User created successfully",
			"token":   token,
		})
	}
}

// CheckUserLogin takes a username or email and a password and checks if the user exists and provided a correct password and returns a new user token
func CheckUserLogin(c *gin.Context) {
	var userData userLoginRequest
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := store.User{
		Username: userData.Username,
		Email:    userData.Email,
		Password: userData.Password,
	}

	token, statusCode, err := store.CheckLogin(user)
	if statusCode != "" || err != nil {
		c.JSON(401, gin.H{
			"message":    "Something went wrong",
			"statusCode": statusCode,
			"error":      err,
		})
	} else {
		c.JSON(200, gin.H{
			"message": "User logged in successfully",
			"token":   token,
		})
	}
}

// NotFound returns a 404 with a "Not found" message
func NotFound(c *gin.Context) {
	c.JSON(404, gin.H{
		"message": "Not found",
	})
}
