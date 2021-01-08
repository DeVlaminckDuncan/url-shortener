package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"xorm.io/xorm/names"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql" // The XORM engine uses this package.
	uuid "github.com/satori/go.uuid"
	"xorm.io/xorm"
)

type storageService struct {
	URLShortenerDB *xorm.Engine
}

type jwtClaims struct {
	Username string
	jwt.StandardClaims
}

var storeService = &storageService{}

// InitializeStore creates the database if it doesn't exist and it synchronizes the tables.
func InitializeStore() {
	// TODO: logger
	// logWriter, err := os.Create("sql.log")
	// if err != nil {
	// 	panic(fmt.Sprintf("Failed to create log writer:\n%d", err))
	// }

	databaseDriver := "mysql"
	connectionString := os.Getenv(strings.ToUpper(databaseDriver) + "_CONNECTION_STRING")

	generalEngine, err := xorm.NewEngine(databaseDriver, strings.Split(connectionString, "/")[0]+"/")
	if err != nil {
		panic(fmt.Sprintf("Failed to open database:\n%d", err))
	}

	_, err = generalEngine.Exec("CREATE DATABASE IF NOT EXISTS " + strings.Split(connectionString, "/")[1])
	if err != nil {
		panic(fmt.Sprintf("Failed to create database:\n%d", err))
	}

	var engine *xorm.Engine
	engine, err = xorm.NewEngine(databaseDriver, connectionString)
	if err != nil {
		panic(fmt.Sprintf("Failed to open database:\n%d", err))
	}

	// TODO: logger
	// engine.SetLogger(log.NewSimpleLogger(logWriter))

	storeService.URLShortenerDB = engine

	// Make the database names the same as the model names
	engine.SetMapper(names.SameMapper{})

	engine.Sync2(new(ShortenedURL))
	engine.Sync2(new(ShortenedURLVisitsHistory))
	engine.Sync2(new(UserShortenedURL))
	engine.Sync2(new(User))
	engine.Sync2(new(UserToken))
}

func checkUserExists(id string) bool {
	var user User
	userExists, err := storeService.URLShortenerDB.Table(&user).Where("ID = ?", id).Exist()
	if err != nil {
		fmt.Println("Failed to fetch User rows:\n", err)
	}
	return userExists
}

// GetLongURL returns the long URL based on the short URL
func GetLongURL(shortURL string) string {
	var shortenedURL ShortenedURL
	_, err := storeService.URLShortenerDB.Table(&shortenedURL).Select("ID, LongURL").Where("ShortURL = ?", shortURL).Get(&shortenedURL)
	if err != nil {
		fmt.Println("Failed to fetch ShortenedURL rows:\n", err)
		// TODO: log error messages
		// storeService.URLShortenerDB.Logger().Errorf()
	}

	var visitsHistory = ShortenedURLVisitsHistory{
		ShortenedURLID: shortenedURL.ID,
	}

	_, err = storeService.URLShortenerDB.Insert(&visitsHistory)
	if err != nil {
		fmt.Println("Failed to insert data into table ShortenedURLVisitsHistory:\n", err)
	}

	return shortenedURL.LongURL
}

// SaveURL inserts a ShortenedURL object and a UserShortenedURL object into the database
func SaveURL(shortURL string, name string, longURL string, userID string) string {
	if !checkUserExists(userID) {
		return "NON_EXISTING_USER"
	}

	id := uuid.NewV4().String()

	var shortenedURL = ShortenedURL{
		ID:       id,
		Name:     name,
		ShortURL: shortURL,
		LongURL:  longURL,
	}

	_, err := storeService.URLShortenerDB.Insert(&shortenedURL)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062") {
			return "DUPLICATE_URL"
		}

		fmt.Println("Failed to insert data into table ShortenedURL:\n", err)
		return err.Error()
	}

	var userShortenedURL = UserShortenedURL{
		UserID:         userID,
		ShortenedURLID: id,
	}

	_, err = storeService.URLShortenerDB.Insert(&userShortenedURL)
	if err != nil {
		fmt.Println("Failed to insert url data into table UserShortenedURL:\n", err)
		return err.Error()
	}

	return ""
}

// GetUserShortenedURLs returns all ShortenedURL objects that a user created.
func GetUserShortenedURLs(userID string) []ShortenedURL {
	var shortenedURLs []ShortenedURL

	if !checkUserExists(userID) {
		return shortenedURLs
	}

	// Get the ShortenedURLIDs by userID from table UserShortenedURL
	var userShortenedURLs []UserShortenedURL
	err := storeService.URLShortenerDB.Table("UserShortenedURL").Select("ShortenedURLID").Find(&userShortenedURLs, &UserShortenedURL{UserID: userID})
	if err != nil {
		fmt.Println("Failed to fetch UserShortenedURL rows:\n", err)
		return shortenedURLs
	}

	// Get the ShortenedURLs using the ShortenedURLIDs from userShortenedURLs
	for _, userShortenedURL := range userShortenedURLs {
		var shortenedURL ShortenedURL
		_, err = storeService.URLShortenerDB.Table(&shortenedURL).Where("ID = ?", userShortenedURL.ShortenedURLID).Get(&shortenedURL)
		if err == nil {
			shortenedURLs = append(shortenedURLs, shortenedURL)
		} else {
			fmt.Println("Failed to fetch ShortenedURL row:\n", err)
			return []ShortenedURL{}
		}
	}

	return shortenedURLs
}

func generateSecurityToken(user User) string {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &jwtClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(user.Password)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Failed to create token string:\n", err)
		return ""
	}

	var userToken = UserToken{
		UserID: user.ID,
		Token:  []byte(tokenString),
	}
	_, err = storeService.URLShortenerDB.Insert(userToken)
	if err != nil {
		fmt.Println("Failed to insert data into table UserToken:\n", err)
		return ""
	}

	return tokenString
}

// TODO: checkSecurityToken() + delete the token if it expired and create a new one

// SaveUser inserts a User object into the database
func SaveUser(user User) string {
	user.ID = uuid.NewV4().String()

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
	if err != nil {
		fmt.Println("Failed to generate password hash:\n", err)
		return ""
	}
	user.Password = string(hash)

	_, err = storeService.URLShortenerDB.Insert(&user)
	if err != nil {
		fmt.Println("Failed to insert data into table User:\n", err)
	}

	token := generateSecurityToken(user)

	return token
}

// CheckLogin compares the given password with the password hash from the database and returns a new token if they match
func CheckLogin(user User) string {
	var userFromDatabase User
	_, err := storeService.URLShortenerDB.Table(&user).Select("ID, Password").Where("Username = ? OR Email = ?", user.Username, user.Email).Get(&userFromDatabase)

	err = bcrypt.CompareHashAndPassword([]byte(userFromDatabase.Password), []byte(user.Password))
	if err != nil {
		fmt.Println(err)
		return ""
	}

	user.ID = userFromDatabase.ID
	token := generateSecurityToken(user)

	return token
}
