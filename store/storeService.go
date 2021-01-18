package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/devlaminckduncan/url-shortener/shortener"

	"golang.org/x/crypto/bcrypt"
	"xorm.io/xorm/log"
	"xorm.io/xorm/names"

	// _ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql" // The XORM engine uses this package.
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql" // The XORM engine uses this package.

	// _ "github.com/denisenkom/go-mssqldb" // The XORM engine uses this package.
	uuid "github.com/satori/go.uuid"
	"xorm.io/xorm"
)

type storageService struct {
	URLShortenerDB *xorm.Engine
}

var storeService = &storageService{}
var logger *os.File

// InitializeStore creates the database if it doesn't exist and it synchronizes the tables.
func InitializeStore() {
	logWriter, err := os.Create("logs/sql.log")
	if err != nil {
		panic(fmt.Sprintf("Failed to create log writer:\n%d", err))
	}

	databaseDriver := "mysql"
	// databaseDriver := "sqlserver"
	// connectionString := os.Getenv(strings.ToUpper(databaseDriver) + "_CONNECTION_STRING")
	connectionString := os.Getenv("HEROKU_MYSQL_CONNECTION_STRING")

	var engine *xorm.Engine
	engine, err = xorm.NewEngine(databaseDriver, connectionString)
	// engine, err := xorm.NewEngine("mssql", connectionString)
	if err != nil {
		panic(fmt.Sprintf("Failed to open database:\n%d", err))
	}

	_, err = engine.DBMetas()
	databaseExists := true
	if err != nil {
		if strings.Contains(err.Error(), "Unknown database") {
			databaseExists = false
		} else {
			panic(fmt.Sprintf("Failed to get database metas:\n%d", err))
		}
	}
	if !databaseExists {
		generalEngine, err := xorm.NewEngine(databaseDriver, strings.Split(connectionString, "/")[0]+"/")
		if err != nil {
			panic(fmt.Sprintf("Failed to open database:\n%d", err))
		}

		databaseName := strings.Split(connectionString, "/")[1]
		fmt.Println("Creating new database " + databaseName + "...")
		_, err = generalEngine.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName)
		if err != nil {
			panic(fmt.Sprintf("Failed to create database:\n%d", err))
		}
	}

	engine.SetLogger(log.NewSimpleLogger(logWriter))

	storeService.URLShortenerDB = engine

	// Make the database names the same as the model names
	engine.SetMapper(names.SameMapper{})

	fmt.Println("Syncing tables...")
	engine.Sync2(new(ShortenedURL))
	engine.Sync2(new(ShortenedURLVisitsHistory))
	engine.Sync2(new(UserShortenedURL))
	engine.Sync2(new(User))
	engine.Sync2(new(UserToken))
	fmt.Println("Finished syncing tables!")

	if !databaseExists {
		err = seedDatabase()
		if err != nil {
			panic(fmt.Sprintf("Failed to seed database:\n%d", err))
		}
	}
}

func seedDatabase() error {
	fmt.Println("Seeding database...")

	user := User{
		ID:        "e0dba740-fc4b-4977-872c-d360239e6b1b",
		FirstName: "Duncan",
		LastName:  "De Vlaminck",
		Username:  "DuncanDV",
		Email:     "duncan.de.vlaminck@student.howest.be",
		Password:  "Astrongpassword.1",
	}

	hash, err := generatePasswordHash(user.Password)
	if err != nil {
		logError("Failed to generate password hash:\n" + err.Error())
		return err
	}
	user.Password = hash

	_, err = storeService.URLShortenerDB.Insert(&user)
	if err != nil {
		logError("Failed to insert data into table User:\n" + err.Error())
		return err
	}

	userTokens := []UserToken{
		{
			UserID: user.ID,
			Token:  []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6IkR1bmNhbkRWIiwiZXhwIjoxNzAwMDAwMDAwfQ.X4Ju07IIAx0wij-iUGgMZn8XSHTT3u5RtGYL7eSmEb4"),
		},
		{
			UserID: user.ID,
			Token:  []byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VybmFtZSI6IkR1bmNhbkRWIiwiZXhwIjoxNjAwMDAwMDAwfQ.vMZd0ser22aw92SWR-hAf5fmnGUFMF8isD8HY3eEfLM"),
		},
	}

	_, err = storeService.URLShortenerDB.Insert(&userTokens)
	if err != nil {
		logError("Failed to insert data into table UserToken:\n" + err.Error())
		return err
	}

	shortenedURLs := []ShortenedURL{
		{
			ID:      "ca48ac65-cb99-45eb-9fcb-ba69d1edb631",
			Name:    "XORM manual",
			LongURL: "https://gobook.io/read/gitea.com/xorm/manual-en-US/",
		},
		{
			ID:      "161b2f3e-5295-11eb-ae93-0242ac130002",
			Name:    "Go tutorial for beginners",
			LongURL: "https://app.pluralsight.com/library/courses/getting-started-with-go/table-of-contents",
		},
	}

	for index := range shortenedURLs {
		shortURL := shortener.GenerateShortURL(shortenedURLs[index].LongURL, user.ID)
		shortenedURLs[index].ShortURL = shortURL
	}

	_, err = storeService.URLShortenerDB.Insert(&shortenedURLs)
	if err != nil {
		logError("Failed to insert data into table ShortenedURL:\n" + err.Error())
		return err
	}

	var userShortenedURLs []UserShortenedURL

	for _, shortenedURL := range shortenedURLs {
		userShortenedURLs = append(userShortenedURLs, UserShortenedURL{
			UserID:         user.ID,
			ShortenedURLID: shortenedURL.ID,
		})
	}

	_, err = storeService.URLShortenerDB.Insert(&userShortenedURLs)
	if err != nil {
		logError("Failed to insert url data into table UserShortenedURL:\n" + err.Error())
		return err
	}

	fmt.Println("Finished seeding database!")
	return nil
}

func logError(errStr string) {
	fmt.Println(errStr)
	storeService.URLShortenerDB.Logger().Errorf(errStr)
}

func checkShortenedURLExists(id string) (bool, error) {
	var shortenedURL ShortenedURL
	shortenedURLExists, err := storeService.URLShortenerDB.Table(&shortenedURL).Where("ID = ?", id).Exist()
	if err != nil {
		logError("Failed to fetch ShortenedURL data:\n" + err.Error())
		return false, err
	}

	return shortenedURLExists, nil
}

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// GenerateSecurityToken creates a new security token using a username and ID and saves it in the database
func GenerateSecurityToken(user User) (string, string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &JWTClaims{
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(os.Getenv("SECRET_JWT_KEY"))
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		logError("Failed to create token string:\n" + err.Error())
		return "", "ERROR_CREATING_TOKEN", err
	}

	userToken := UserToken{
		UserID: user.ID,
		Token:  []byte(tokenString),
	}
	_, err = storeService.URLShortenerDB.Insert(&userToken)
	if err != nil {
		logError("Failed to insert data into table UserToken:\n" + err.Error())
		return "", "ERROR_INSERTING_USERTOKEN", err
	}

	return tokenString, "OK", nil
}

// CheckSecurityTokenExists checks whether the given security token exists in the database and if it expired
func CheckSecurityTokenExists(tokenString string) (bool, string, error) {
	tokenExists, err := storeService.URLShortenerDB.Table(&UserToken{}).Where("Token = ?", tokenString).Exist()
	if err != nil {
		logError("Failed to fetch UserToken data:\n" + err.Error())
		return false, "ERROR_FETCHING_USERTOKEN", err
	}
	if !tokenExists {
		return false, "NON_EXISTING_USERTOKEN", nil
	}

	return true, "OK", nil
}

// DeleteSecurityToken deletes the given security token from the database
func DeleteSecurityToken(token string) (string, error) {
	tokenExists, statusCode, err := CheckSecurityTokenExists(token)
	if statusCode != "OK" || err != nil || !tokenExists {
		return statusCode, err
	}

	var userToken UserToken
	_, err = storeService.URLShortenerDB.Table(&userToken).Where("Token = ?", token).Get(&userToken)
	if err != nil {
		logError("Failed to fetch UserToken data:\n" + err.Error())
		return "ERROR_FETCHING_USERTOKEN", err
	}

	_, err = storeService.URLShortenerDB.Delete(&userToken)
	if err != nil {
		logError("Failed to delete data from table UserToken:\n" + err.Error())
		return "ERROR_DELETING_USERTOKEN", err
	}

	return "OK", nil
}

// CheckUserExists checks if the given user ID, username or email exists in the database
func CheckUserExists(uniqueValue string) (bool, string, error) {
	var user User
	userExists, err := storeService.URLShortenerDB.Table(&user).Where("ID = ? OR Username = ? OR Email = ?", uniqueValue, uniqueValue, uniqueValue).Exist()
	if err != nil {
		logError("Failed to fetch User data:\n" + err.Error())
		return false, "ERROR_FETCHING_USER", err
	}

	return userExists, "OK", nil
}

// GetLongURL returns the long URL based on the short URL
func GetLongURL(shortURL string) string {
	var shortenedURL ShortenedURL
	_, err := storeService.URLShortenerDB.Table(&shortenedURL).Select("ID, LongURL").Where("ShortURL = ?", shortURL).Get(&shortenedURL)
	if err != nil {
		logError("Failed to fetch ShortenedURL data:\n" + err.Error())
	}

	var visitsHistory = ShortenedURLVisitsHistory{
		ShortenedURLID: shortenedURL.ID,
	}

	_, err = storeService.URLShortenerDB.Insert(&visitsHistory)
	if err != nil {
		logError("Failed to insert data into table ShortenedURLVisitsHistory:\n" + err.Error())
	}

	return shortenedURL.LongURL
}

// SaveURL inserts a ShortenedURL object and a UserShortenedURL object into the database
func SaveURL(shortURL string, name string, longURL string, userID string) (ShortenedURL, string, error) {
	userExists, statusCode, err := CheckUserExists(userID)
	if statusCode != "OK" || err != nil {
		return ShortenedURL{}, statusCode, err
	} else if !userExists {
		return ShortenedURL{}, "NON_EXISTING_USER", nil
	}

	id := uuid.NewV4().String()

	var shortenedURL = ShortenedURL{
		ID:       id,
		Name:     name,
		ShortURL: shortURL,
		LongURL:  longURL,
	}

	_, err = storeService.URLShortenerDB.Insert(&shortenedURL)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062") {
			return ShortenedURL{}, "DUPLICATE_URL", err
		}

		logError("Failed to insert data into table ShortenedURL:\n" + err.Error())
		return ShortenedURL{}, "ERROR_INSERTING_SHORTENEDURL", err
	}

	var userShortenedURL = UserShortenedURL{
		UserID:         userID,
		ShortenedURLID: id,
	}

	_, err = storeService.URLShortenerDB.Insert(&userShortenedURL)
	if err != nil {
		logError("Failed to insert url data into table UserShortenedURL:\n" + err.Error())
		return ShortenedURL{}, "ERROR_INSERTING_USERSHORTENEDURL", err
	}

	now := time.Now()
	shortenedURL.CreatedAt = now

	return shortenedURL, "OK", nil
}

// UpdateShortenedURL updates the given shortenedURL object in the database
func UpdateShortenedURL(shortenedURL ShortenedURL) (string, error) {
	shortenedURLExists, err := checkShortenedURLExists(shortenedURL.ID)
	if err != nil {
		return "ERROR_FETCHING_SHORTENEDURL", err
	}
	if !shortenedURLExists {
		return "NON_EXISTING_SHORTENEDURL", nil
	}

	_, err = storeService.URLShortenerDB.ID(shortenedURL.ID).Update(&shortenedURL)
	if err != nil {
		logError("Failed to update data in table ShortenedURL:\n" + err.Error())
		return "ERROR_UPDATING_SHORTENEDURL", err
	}

	return "OK", nil
}

// DeleteShortenedURL deletes the given shortenedURL object in the database
func DeleteShortenedURL(id string) (string, error) {
	var shortenedURL ShortenedURL
	_, err := storeService.URLShortenerDB.Table(&shortenedURL).Where("ID = ?", id).Get(&shortenedURL)
	if err != nil {
		logError("Failed to fetch ShortenedURL data:\n" + err.Error())
		return "ERROR_FETCHING_SHORTENEDURL", err
	}

	shortenedURLExists, err := checkShortenedURLExists(id)
	if err != nil {
		return "ERROR_FETCHING_SHORTENEDURL", err
	}
	if !shortenedURLExists {
		return "NON_EXISTING_SHORTENEDURL", nil
	}

	_, err = storeService.URLShortenerDB.Delete(&shortenedURL)
	if err != nil {
		logError("Failed to delete data from table ShortenedURL:\n" + err.Error())
		return "ERROR_DELETING_SHORTENEDURL", err
	}

	var usershortenedURL UserShortenedURL
	_, err = storeService.URLShortenerDB.Table(&usershortenedURL).Where("ShortenedURLID = ?", id).Get(&usershortenedURL)
	if err != nil {
		logError("Failed to fetch UserShortenedURL data:\n" + err.Error())
		return "ERROR_FETCHING_USERSHORTENEDURL", err
	}

	_, err = storeService.URLShortenerDB.Delete(&usershortenedURL)
	if err != nil {
		logError("Failed to delete data from table UserShortenedURL:\n" + err.Error())
		return "ERROR_DELETING_USERSHORTENEDURL", err
	}

	var analytics []ShortenedURLVisitsHistory
	err = storeService.URLShortenerDB.Table(&ShortenedURLVisitsHistory{}).Find(&analytics, &ShortenedURLVisitsHistory{ShortenedURLID: id})
	if err != nil {
		logError("Failed to fetch ShortenedURLVisitsHistory data:\n" + err.Error())
		return "ERROR_FETCHING_SHORTENEDURLVISITSHISTORY", err
	}

	for _, item := range analytics {
		_, err = storeService.URLShortenerDB.Delete(&item)
		if err != nil {
			logError("Failed to delete data from table ShortenedURLVisitsHistory:\n" + err.Error())
			return "ERROR_DELETING_SHORTENEDURLVISITSHISTORY", err
		}
	}

	return "OK", nil
}

// GetUserShortenedURLs returns all ShortenedURL objects with analytics that a user created.
func GetUserShortenedURLs(userID string) ([]ShortenedURLData, string, error) {
	var shortenedURLData []ShortenedURLData

	userExists, statusCode, err := CheckUserExists(userID)
	if statusCode != "OK" || err != nil {
		return shortenedURLData, statusCode, err
	} else if !userExists {
		return shortenedURLData, "NON_EXISTING_USER", nil
	}

	// Get the ShortenedURLIDs by userID from table UserShortenedURL
	var userShortenedURLs []UserShortenedURL
	err = storeService.URLShortenerDB.Table(&UserShortenedURL{}).Select("ShortenedURLID").Find(&userShortenedURLs, &UserShortenedURL{UserID: userID})
	if err != nil {
		logError("Failed to fetch UserShortenedURL data:\n" + err.Error())
		return shortenedURLData, "ERROR_FETCHING_USERSHORTENEDURL", err
	}

	// Get the ShortenedURLs using the ShortenedURLIDs from userShortenedURLs
	// Get the analytics using the ShortenedURLIDs from shortenedURLData.ShortenedURLs
	for _, userShortenedURL := range userShortenedURLs {
		var (
			data         ShortenedURLData
			shortenedURL ShortenedURL
		)
		_, err = storeService.URLShortenerDB.Table(&shortenedURL).Where("ID = ?", userShortenedURL.ShortenedURLID).Get(&shortenedURL)
		if err == nil {
			data.ShortenedURLObject = shortenedURL
		} else {
			logError("Failed to fetch ShortenedURL data:\n" + err.Error())
			return []ShortenedURLData{}, "ERROR_FETCHING_SHORTENEDURL", err
		}

		var urlAnalytics []string
		err := storeService.URLShortenerDB.Table(&ShortenedURLVisitsHistory{}).Select("VisitedAt").Find(&urlAnalytics, &ShortenedURLVisitsHistory{ShortenedURLID: data.ShortenedURLObject.ID})
		if err == nil {
			data.Analytics = urlAnalytics
		} else if !strings.Contains(err.Error(), "Error 1054") {
			logError("Failed to fetch ShortenedURLVisitsHistory data:\n" + err.Error())
			return []ShortenedURLData{}, "ERROR_FETCHING_ANALYTICS", err
		}

		shortenedURLData = append(shortenedURLData, data)
	}

	return shortenedURLData, "OK", nil
}

// SaveUser inserts a User object into the database
func SaveUser(user User) (string, string, string, error) {
	user.ID = uuid.NewV4().String()

	hash, err := generatePasswordHash(user.Password)
	if err != nil {
		logError("Failed to generate password hash:\n" + err.Error())
		return "", "", "ERROR_GENERATING_HASH", err
	}
	user.Password = hash

	_, err = storeService.URLShortenerDB.Insert(&user)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062") {
			return "", "", "DUPLICATE_USER", err
		}

		logError("Failed to insert data into table User:\n" + err.Error())
		return "", "", "ERROR_INSERTING_USER", err
	}

	token, statusCode, err := GenerateSecurityToken(user)
	if statusCode != "OK" || err != nil {
		return "", "", statusCode, err
	}

	return token, user.ID, "OK", nil
}

// GetUser returns a User object by ID, username or email
func GetUser(uniqueValue string) (User, string, error) {
	userExists, statusCode, err := CheckUserExists(uniqueValue)
	if statusCode != "OK" || err != nil {
		return User{}, statusCode, err
	} else if !userExists {
		return User{}, "NON_EXISTING_USER", nil
	}

	var user User
	_, err = storeService.URLShortenerDB.Table(&user).Select("ID, FirstName, LastName, Username, Email").Where("ID = ? OR Username = ? OR Email = ?", uniqueValue, uniqueValue, uniqueValue).Get(&user)
	if err != nil {
		logError("Failed to fetch User data:\n" + err.Error())
		return User{}, "ERROR_FETCHING_USER", err
	}

	return user, "OK", nil
}

// UpdateUser updates the given user object in the database
func UpdateUser(user User, token string) (string, string, error) {
	userExists, statusCode, err := CheckUserExists(user.ID)
	if statusCode != "OK" || err != nil {
		return "", statusCode, err
	}
	if !userExists {
		return "", "NON_EXISTING_USER", nil
	}

	if user.Password != "" {
		hash, err := generatePasswordHash(user.Password)
		if err != nil {
			logError("Failed to generate password hash:\n" + err.Error())
			return "", "ERROR_GENERATING_HASH", err
		}

		user.Password = hash
	}

	var oldUser User
	_, err = storeService.URLShortenerDB.Table(&oldUser).Select("ID, FirstName, LastName, Username, Email").Where("ID = ?", user.ID).Get(&oldUser)
	if err != nil {
		logError("Failed to fetch User data:\n" + err.Error())
		return "", "ERROR_FETCHING_USER", err
	}

	_, err = storeService.URLShortenerDB.ID(user.ID).Update(&user)
	if err != nil {
		if strings.Contains(err.Error(), "Error 1062") {
			return "", "DUPLICATE_USER", err
		}

		logError("Failed to update data in table User:\n" + err.Error())
		return "", "ERROR_UPDATING_USER", err
	}

	// Create a new token if the username was changed because the payload of the token contains the username
	if user.Username != oldUser.Username {
		newToken, statusCode, err := GenerateSecurityToken(user)
		if statusCode != "OK" || err != nil {
			return "", statusCode, err
		}

		statusCode, err = DeleteSecurityToken(token)
		if statusCode != "OK" || err != nil {
			return "", statusCode, err
		}

		return newToken, "OK", nil
	}

	return "", "OK", nil
}

// DeleteUser returns a User object by ID
func DeleteUser(id string) (string, error) {
	userExists, statusCode, err := CheckUserExists(id)
	if statusCode != "OK" || err != nil {
		return statusCode, err
	} else if !userExists {
		return "NON_EXISTING_USER", nil
	}

	// Delete the User
	_, err = storeService.URLShortenerDB.Delete(&User{ID: id})
	if err != nil {
		logError("Failed to delete data from table User:\n" + err.Error())
		return "ERROR_DELETING_USER", err
	}

	// Delete the UserTokens
	_, err = storeService.URLShortenerDB.Delete(&UserToken{UserID: id})
	if err != nil {
		logError("Failed to delete data from table User:\n" + err.Error())
		return "ERROR_DELETING_USER", err
	}

	// Get the ShortenedURLIDs by userID from table UserShortenedURL
	var userShortenedURLs []UserShortenedURL
	err = storeService.URLShortenerDB.Table(&UserShortenedURL{}).Select("ShortenedURLID").Find(&userShortenedURLs, &UserShortenedURL{UserID: id})
	if err != nil {
		logError("Failed to fetch UserShortenedURL data:\n" + err.Error())
		return "ERROR_FETCHING_USERSHORTENEDURL", err
	}

	// Delete the UserShortenedURLs, ShortenedURLs and analytics (ShortenedURLVisitsHistory)
	for _, userShortenedURL := range userShortenedURLs {
		_, err = storeService.URLShortenerDB.Delete(&userShortenedURL)
		if err != nil {
			logError("Failed to delete data from table UserShortenedURL:\n" + err.Error())
			return "ERROR_DELETING_USERSHORTENEDURL", err
		}

		_, err = storeService.URLShortenerDB.Delete(&ShortenedURL{ID: userShortenedURL.ShortenedURLID})
		if err != nil {
			logError("Failed to delete data from table UserShortenedURL:\n" + err.Error())
			return "ERROR_DELETING_SHORTENEDURL", err
		}

		_, err = storeService.URLShortenerDB.Delete(&ShortenedURLVisitsHistory{ShortenedURLID: userShortenedURL.ShortenedURLID})
		if err != nil {
			logError("Failed to delete data from table ShortenedURLVisitsHistory:\n" + err.Error())
			return "ERROR_DELETING_SHORTENEDURLVISITSHISTORY", err
		}
	}

	return "OK", nil
}

// CheckLogin compares the given password with the password hash from the database and returns a new token if they match
func CheckLogin(user User) (string, string, string, error) {
	var uniqueValue string
	if user.Username != "" {
		uniqueValue = user.Username
	} else {
		uniqueValue = user.Email
	}

	userExists, statusCode, err := CheckUserExists(uniqueValue)
	if statusCode != "OK" || err != nil {
		return "", "", statusCode, err
	} else if !userExists {
		return "", "", "NON_EXISTING_USER", nil
	}

	var userFromDatabase User
	_, err = storeService.URLShortenerDB.Table(&user).Select("ID, Password").Where("Username = ? OR Email = ?", user.Username, user.Email).Get(&userFromDatabase)
	if err != nil {
		return "", "", "ERROR_FETCHING_USER", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userFromDatabase.Password), []byte(user.Password))
	if err != nil {
		return "", "", "WRONG_PASSWORD", err
	}

	user.ID = userFromDatabase.ID
	token, statusCode, err := GenerateSecurityToken(user)
	if statusCode != "OK" || err != nil {
		return "", "", statusCode, err
	}

	return token, user.ID, "OK", nil
}
