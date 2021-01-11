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

type urlUpdateRequest struct {
	Name    string `json:"name"`
	LongURL string `json:"longURL"`
}

type userLoginRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password" binding:"required"`
}

type userUpdateRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// TODO: check every request for empty values
// TODO: add codes to JSON data eg "statusCode": "URL_CREATED"

// RedirectShortURL takes a short URL redirects you to the long URL from the database and creates a new ShortenedURLVisitsHistory
func RedirectShortURL(c *gin.Context) {
	shortURL := c.Request.URL.Path[1:]
	longURL := store.GetLongURL(shortURL)

	c.Redirect(301, longURL)
}

// UpdateShortURL takes a name and a long URL and updates the ShortenedURL in the database
func UpdateShortURL(c *gin.Context) {
	var urlData urlUpdateRequest
	if err := c.ShouldBindJSON(&urlData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var shortenedURL = store.ShortenedURL{
		ID:      c.Param("id"),
		Name:    urlData.Name,
		LongURL: urlData.LongURL,
	}
	statusCode, err := store.UpdateShortenedURL(shortenedURL)
	if statusCode != "OK" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "Short URL updated successfully",
		"statusCode": statusCode,
	})
}

// DeleteShortURL deletes the ShortenedURL in the database
func DeleteShortURL(c *gin.Context) {
	id := c.Param("id")

	statusCode, err := store.DeleteShortenedURL(id)
	if statusCode != "OK" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "Short URL deleted successfully",
		"statusCode": statusCode,
	})
}

// CreateShortURL takes a name, a long URL and a user ID and creates a new ShortenedURL
func CreateShortURL(c *gin.Context) {
	var creationRequest urlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortURL := shortener.GenerateShortURL(creationRequest.LongURL, creationRequest.UserID)
	statusCode, err := store.SaveURL(shortURL, creationRequest.Name, creationRequest.LongURL, creationRequest.UserID)
	if statusCode != "OK" || err != nil {
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
		"message":    "Short URL created successfully",
		"statusCode": statusCode,
		"shortURL":   "http://" + c.Request.Host + "/" + shortURL,
	})
}

// GetUserShortenedURLs takes a user ID and returns the user's ShortenedURLs
func GetUserShortenedURLs(c *gin.Context) {
	userID := c.Param("userID")

	urls, statusCode, err := store.GetUserShortenedURLs(userID)
	if statusCode != "OK" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, urls)
}

// GetUser returns user information by user ID
func GetUser(c *gin.Context) {
	userID := c.Param("userID")

	user, statusCode, err := store.GetUser(userID)
	if statusCode != "OK" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, user)
}

// UpdateUser takes a first name, a last name, a username, an email and a password and updates the User in the database
func UpdateUser(c *gin.Context) {
	userID := c.Param("userID")

	var userData userUpdateRequest
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user = store.User{
		ID:        userID,
		FirstName: userData.FirstName,
		LastName:  userData.LastName,
		Username:  userData.Username,
		Email:     userData.Email,
		Password:  userData.Password,
	}

	statusCode, err := store.UpdateUser(user)
	if statusCode != "OK" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "User updated successfully",
		"statusCode": statusCode,
	})
}

// DeleteUser deletes the User in the database
func DeleteUser(c *gin.Context) {
	userID := c.Param("userID")

	statusCode, err := store.DeleteUser(userID)
	if statusCode != "OK" || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "User deleted successfully",
		"statusCode": statusCode,
	})
}

// CreateUser takes a first name, a last name, a username, an email and a password and creates a new User and returns a new user token
func CreateUser(c *gin.Context) {
	var user store.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, statusCode, err := store.SaveUser(user)
	if statusCode != "OK" || err != nil {
		c.JSON(401, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "User created successfully",
		"statusCode": statusCode,
		"token":      token,
	})
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
	if statusCode != "OK" || err != nil {
		c.JSON(401, gin.H{
			"statusCode": statusCode,
			"error":      err,
		})
		return
	}

	c.JSON(200, gin.H{
		"message":    "User logged in successfully",
		"statusCode": statusCode,
		"token":      token,
	})
}

// NotFound returns a 404 with a "Not found" message
func NotFound(c *gin.Context) {
	c.JSON(404, gin.H{
		"message": "Not found",
	})
}
