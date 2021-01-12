package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/devlaminckduncan/url-shortener/shortener"
	"github.com/devlaminckduncan/url-shortener/store"
	"github.com/dgrijalva/jwt-go"
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

type tokenHeader struct {
	Authorization string `header:"Authorization" binding:"required"`
}

func parseTokenWithClaims(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &store.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET_JWT_KEY")), nil
	})

	return token, err
}

func checkSecurityToken(c *gin.Context) (bool, string, string, error) {
	var tokenHeaderData tokenHeader
	if err := c.ShouldBindHeader(&tokenHeaderData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false, "", "ERROR_BINDING_HEADER", err
	}

	tokenString := strings.Replace(tokenHeaderData.Authorization, "Bearer ", "", 1)
	var newTokenString string

	token, err := parseTokenWithClaims(tokenString)
	if err != nil && strings.Contains(err.Error(), "token is expired by") {
		claims, _ := token.Claims.(*store.JWTClaims)
		user, statusCode, err := store.GetUser(claims.Username)
		if statusCode != "OK" || err != nil {
			return false, "", statusCode, err
		}

		statusCode, err = store.DeleteSecurityToken(tokenString)
		if statusCode != "OK" || err != nil {
			return false, "", statusCode, err
		}

		newTokenString, statusCode, err = store.GenerateSecurityToken(user)
		if statusCode != "OK" || err != nil {
			return false, "", statusCode, err
		}

		token, err = parseTokenWithClaims(newTokenString)
		if err != nil {
			return false, "", "ERROR_GENERATING_TOKEN", err
		}
	}

	if claims, ok := token.Claims.(*store.JWTClaims); !ok || !token.Valid {
		return false, "", "INVALID_TOKEN", err
	} else if userExists, statusCode, err := store.CheckUserExists(claims.Username); err != nil || !userExists {
		return false, "", statusCode, err
	}

	if newTokenString != "" {
		return true, newTokenString, "OK", nil
	}

	return true, "", "OK", nil
}

// RedirectShortURL takes a short URL redirects you to the long URL from the database and creates a new ShortenedURLVisitsHistory
func RedirectShortURL(c *gin.Context) {
	shortURL := c.Request.URL.Path[1:]
	longURL := store.GetLongURL(shortURL)

	c.Redirect(301, longURL)
}

// UpdateShortURL takes a name and a long URL and updates the ShortenedURL in the database
func UpdateShortURL(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {
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
		statusCode, err = store.UpdateShortenedURL(shortenedURL)
		if statusCode != "OK" || err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"statusCode": statusCode,
			})
			return
		}

		c.JSON(200, gin.H{
			"message":    "Short URL updated successfully",
			"statusCode": statusCode,
			"newToken":   newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
}

// DeleteShortURL deletes the ShortenedURL in the database
func DeleteShortURL(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {

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
			"newToken":   newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
}

// CreateShortURL takes a name, a long URL and a user ID and creates a new ShortenedURL
func CreateShortURL(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {

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
			"newToken":   newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
}

// GetUserShortenedURLs takes a user ID and returns the user's ShortenedURLs
func GetUserShortenedURLs(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {

		userID := c.Param("userID")

		urls, statusCode, err := store.GetUserShortenedURLs(userID)
		if statusCode != "OK" || err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"statusCode": statusCode,
				"error":      err,
			})
			return
		}

		c.JSON(200, gin.H{
			"urls":     urls,
			"newToken": newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
}

// GetUser returns user information by user ID
func GetUser(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {

		userID := c.Param("userID")

		user, statusCode, err := store.GetUser(userID)
		if statusCode != "OK" || err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"statusCode": statusCode,
				"error":      err,
			})
			return
		}

		c.JSON(200, gin.H{
			"user":     user,
			"newToken": newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
}

// UpdateUser takes a first name, a last name, a username, an email and a password and updates the User in the database
func UpdateUser(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {

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
			"newToken":   newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
}

// DeleteUser deletes the User in the database
func DeleteUser(c *gin.Context) {
	ok, newToken, statusCode, err := checkSecurityToken(c)
	if ok {

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
			"newToken":   newToken,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"statusCode": statusCode,
			"err":        err,
		})
	}
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
