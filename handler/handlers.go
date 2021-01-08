package handler

import (
	"net/http"

	"github.com/devlaminckduncan/url-shortener/shortener"
	"github.com/devlaminckduncan/url-shortener/store"
	"github.com/gin-gonic/gin"
)

type urlCreationRequest struct {
	Name    string `json:"name" binding:"required"`
	LongURL string `json:"long_url" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

type userSignupRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type userLoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

// TODO: check every request for empty values
// TODO: add codes to JSON data eg "status_code": "URL_CREATED"

func CreateShortURL(c *gin.Context) {
	var creationRequest urlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL := shortener.GenerateShortURL(creationRequest.LongURL, creationRequest.UserID)
	err := store.SaveURL(shortURL, creationRequest.Name, creationRequest.LongURL, creationRequest.UserID)
	if err != "" {
		status := http.StatusBadRequest
		if err == "NON_EXISTING_USER" {
			status = http.StatusUnauthorized
		}

		c.JSON(status, gin.H{
			"status_code": err,
		})
		return
	}

	host := "http://localhost:9001/"
	c.JSON(200, gin.H{
		"message":   "Short URL created successfully",
		"short_url": host + shortURL,
	})
}

func RedirectShortURL(c *gin.Context) {
	shortURL := c.Request.URL.Path[1:]
	longURL := store.GetLongURL(shortURL)

	c.Redirect(301, longURL)
}

func GetUserShortenedURLs(c *gin.Context) {
	userID := c.Request.URL.Path[12:]

	urls := store.GetUserShortenedURLs(userID)

	c.JSON(200, urls)
}

func CreateUser(c *gin.Context) {
	var userData userSignupRequest
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := store.User{
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Username:  userData.Username,
		Email:     userData.Email,
		Password:  userData.Password,
	}

	token := store.SaveUser(user)

	if token != "" {
		c.JSON(200, gin.H{
			"message": "User created successfully",
			"token":   token,
		})
	} else {
		c.JSON(401, gin.H{
			"message": "Something went wrong",
		})
	}
}

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

	token := store.CheckLogin(user)
	if token != "" {
		c.JSON(200, gin.H{
			"message": "User logged in successfully",
			"token":   token,
		})
	} else {
		c.JSON(401, gin.H{
			"message": "Something went wrong",
		})
	}
}

func NotFound(c *gin.Context) {
	c.JSON(404, gin.H{
		"message": "Not found",
	})
}
