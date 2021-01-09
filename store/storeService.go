package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/devlaminckduncan/url-shortener/shortener"

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

	var engine *xorm.Engine
	engine, err := xorm.NewEngine(databaseDriver, connectionString)
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
		fmt.Println("Creating new database " + databaseName)
		_, err = generalEngine.Exec("CREATE DATABASE IF NOT EXISTS " + databaseName)
		if err != nil {
			panic(fmt.Sprintf("Failed to create database:\n%d", err))
		}
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
		fmt.Println("Failed to generate password hash:\n", err)
		return err
	}
	user.Password = hash

	_, err = storeService.URLShortenerDB.Insert(&user)
	if err != nil {
		fmt.Println("Failed to insert data into table User:\n", err)
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
		fmt.Println("Failed to insert data into table ShortenedURL:\n", err)
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
		fmt.Println("Failed to insert url data into table UserShortenedURL:\n", err)
		return err
	}

	fmt.Println("Finished seeding database!")
	return nil
}

func checkUserExists(id string) (bool, error) {
	var user User
	userExists, err := storeService.URLShortenerDB.Table(&user).Where("ID = ?", id).Exist()
	if err != nil {
		fmt.Println("Failed to fetch User data:\n", err)
		return false, err
	}
	return userExists, nil
}

// GetLongURL returns the long URL based on the short URL
func GetLongURL(shortURL string) string {
	var shortenedURL ShortenedURL
	_, err := storeService.URLShortenerDB.Table(&shortenedURL).Select("ID, LongURL").Where("ShortURL = ?", shortURL).Get(&shortenedURL)
	if err != nil {
		fmt.Println("Failed to fetch ShortenedURL data:\n", err)
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
func SaveURL(shortURL string, name string, longURL string, userID string) (string, error) {
	userExists, err := checkUserExists(userID)
	if err != nil {
		return "ERROR_CHECKING_USER", err
	} else if !userExists {
		return "NON_EXISTING_USER", err
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
			return "DUPLICATE_URL", err
		}

		fmt.Println("Failed to insert data into table ShortenedURL:\n", err)
		return "ERROR_INSERTING_SHORTENEDURL", err
	}

	var userShortenedURL = UserShortenedURL{
		UserID:         userID,
		ShortenedURLID: id,
	}

	_, err = storeService.URLShortenerDB.Insert(&userShortenedURL)
	if err != nil {
		fmt.Println("Failed to insert url data into table UserShortenedURL:\n", err)
		return "ERROR_INSERTING_USERSHORTENEDURL", err
	}

	return "OK", nil
}

// UpdateShortenedURL updates the given shortenedURL object in the database
func UpdateShortenedURL(shortenedURL ShortenedURL) (string, error) {
	// TODO: what if the shortenedURL doesn't exist in the database?

	_, err := storeService.URLShortenerDB.ID(shortenedURL.ID).Update(&shortenedURL)
	if err != nil {
		fmt.Println("Failed to update data in table ShortenedURL:\n", err)
		return "ERROR_UPDATING_SHORTENEDURL", err
	}

	return "OK", nil
}

// DeleteShortenedURL deletes the given shortenedURL object in the database
func DeleteShortenedURL(id string) (string, error) {
	var shortenedURL ShortenedURL
	_, err := storeService.URLShortenerDB.Table(&shortenedURL).Where("ID = ?", id).Get(&shortenedURL)
	if err != nil {
		fmt.Println("Failed to fetch ShortenedURL data:\n", err)
		return "ERROR_FETCHING_SHORTENEDURL", err
	}

	_, err = storeService.URLShortenerDB.Delete(&shortenedURL)
	if err != nil {
		fmt.Println("Failed to delete data from table ShortenedURL:\n", err)
		return "ERROR_DELETING_SHORTENEDURL", err
	}

	var usershortenedURL UserShortenedURL
	_, err = storeService.URLShortenerDB.Table(&usershortenedURL).Where("ShortenedURLID = ?", id).Get(&usershortenedURL)
	if err != nil {
		fmt.Println("Failed to fetch UserShortenedURL data:\n", err)
		return "ERROR_FETCHING_USERSHORTENEDURL", err
	}

	_, err = storeService.URLShortenerDB.Delete(&usershortenedURL)
	if err != nil {
		fmt.Println("Failed to delete data from table UserShortenedURL:\n", err)
		return "ERROR_DELETING_USERSHORTENEDURL", err
	}

	var analytics []ShortenedURLVisitsHistory
	err = storeService.URLShortenerDB.Table(&ShortenedURLVisitsHistory{}).Find(&analytics, &ShortenedURLVisitsHistory{ShortenedURLID: id})
	if err != nil {
		fmt.Println("Failed to fetch ShortenedURLVisitsHistory data:\n", err)
		return "ERROR_FETCHING_SHORTENEDURLVISITSHISTORY", err
	}

	for _, item := range analytics {
		_, err = storeService.URLShortenerDB.Delete(&item)
		if err != nil {
			fmt.Println("Failed to delete data from table ShortenedURLVisitsHistory:\n", err)
			return "ERROR_DELETING_SHORTENEDURLVISITSHISTORY", err
		}
	}

	return "OK", nil
}

// GetUserShortenedURLs returns all ShortenedURL objects with analytics that a user created.
func GetUserShortenedURLs(userID string) ([]ShortenedURLData, string, error) {
	// TODO: add JSON names in the models ShortenedURLData, ShortenedURL, ShortenedURLVisitsHistory
	var shortenedURLData []ShortenedURLData

	userExists, err := checkUserExists(userID)
	if err != nil {
		return shortenedURLData, "ERROR_CHECKING_USER", err
	} else if !userExists {
		return shortenedURLData, "NON_EXISTING_USER", err
	}

	// Get the ShortenedURLIDs by userID from table UserShortenedURL
	var userShortenedURLs []UserShortenedURL
	err = storeService.URLShortenerDB.Table(&UserShortenedURL{}).Select("ShortenedURLID").Find(&userShortenedURLs, &UserShortenedURL{UserID: userID})
	if err != nil {
		fmt.Println("Failed to fetch UserShortenedURL data:\n", err)
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
			fmt.Println("Failed to fetch ShortenedURL data:\n", err)
			return []ShortenedURLData{}, "ERROR_FETCHING_SHORTENEDURL", err
		}

		var urlAnalytics []string
		err := storeService.URLShortenerDB.Table(&ShortenedURLVisitsHistory{}).Select("VisitedAt").Find(&urlAnalytics, &ShortenedURLVisitsHistory{ShortenedURLID: data.ShortenedURLObject.ID})
		if err == nil {
			data.Analytics = urlAnalytics
		} else if !strings.Contains(err.Error(), "Error 1054") {
			fmt.Println("Failed to fetch ShortenedURLVisitsHistory data:\n", err)
			return []ShortenedURLData{}, "ERROR_FETCHING_ANALYTICS", err
		}

		shortenedURLData = append(shortenedURLData, data)
	}

	return shortenedURLData, "OK", nil
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

func generatePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// SaveUser inserts a User object into the database
func SaveUser(user User) (string, string, error) {
	user.ID = uuid.NewV4().String()

	hash, err := generatePasswordHash(user.Password)
	if err != nil {
		fmt.Println("Failed to generate password hash:\n", err)
		return "", "ERROR_GENERATING_HASH", err
	}
	user.Password = hash

	_, err = storeService.URLShortenerDB.Insert(&user)
	if err != nil {
		fmt.Println("Failed to insert data into table User:\n", err)
		return "", "ERROR_INSERTING_USER", err
	}

	token := generateSecurityToken(user)

	return token, "", nil
}

// GetUser returns a User object by ID
func GetUser(userID string) (User, string, error) {
	var user User
	_, err := storeService.URLShortenerDB.Table(&user).Where("ID = ?", user.ID).Get(&user)
	if err != nil {
		fmt.Println("Failed to fetch User data:\n", err)
		return user, "ERROR_FETCHING_USER", err
	}

	return user, "OK", nil
}

// UpdateUser returns a User object by ID
func UpdateUser(user User) (string, error) {
	_, err := storeService.URLShortenerDB.Update(&user)
	if err != nil {
		fmt.Println("Failed to update data in table User:\n", err)
		return "ERROR_UPDATING_USER", err
	}

	return "OK", nil
}

// DeleteUser returns a User object by ID
func DeleteUser(user User) (string, error) {
	_, err := storeService.URLShortenerDB.Delete(&user)
	if err != nil {
		fmt.Println("Failed to delete data from table User:\n", err)
		return "ERROR_DELETING_USER", err
	}

	// TODO: delete the user's ShortenedURLs, ShortenedURLVisitsHistory and UserShortenedURLs

	return "OK", nil
}

// CheckLogin compares the given password with the password hash from the database and returns a new token if they match
func CheckLogin(user User) (string, string, error) {
	var userFromDatabase User
	_, err := storeService.URLShortenerDB.Table(&user).Select("ID, Password").Where("Username = ? OR Email = ?", user.Username, user.Email).Get(&userFromDatabase)

	err = bcrypt.CompareHashAndPassword([]byte(userFromDatabase.Password), []byte(user.Password))
	if err != nil {
		fmt.Println(err)
		return "", "ERROR_FETCHING_USER", err
	}

	user.ID = userFromDatabase.ID
	token := generateSecurityToken(user)

	return token, "OK", nil
}
