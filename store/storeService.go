package store

import (
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"xorm.io/xorm"
)

type StorageService struct {
	UrlShortenerDb *xorm.Engine
}

var (
	storeService = &StorageService{}
)

func InitializeStore() *StorageService {
	// logWriter, logErr := os.Create("sql.log")
	// if logErr != nil {
	// 	panic(fmt.Sprintf("Failed to create log writer:\n", logErr))
	// }

	connectionString := os.Getenv("MYSQL_CONNECTION_STRING")

	generalEngine, generalEngineErr := xorm.NewEngine("mysql", strings.Split(connectionString, "/")[0]+"/")
	if generalEngineErr != nil {
		panic(fmt.Sprintf("Failed to open database:\n", generalEngineErr))
	}

	_, createDbErr := generalEngine.Exec("CREATE DATABASE IF NOT EXISTS " + strings.Split(connectionString, "/")[1])
	if createDbErr != nil {
		panic(fmt.Sprintf("Failed to create database:\n", createDbErr))
	}

	engine, engineErr := xorm.NewEngine("mysql", connectionString)
	if engineErr != nil {
		panic(fmt.Sprintf("Failed to open database:\n", engineErr))
	}

	// engine.SetLogger(log.NewSimpleLogger(logWriter))

	storeService.UrlShortenerDb = engine

	engine.Sync2(new(ShortenedUrl))
	engine.Sync2(new(User))
	engine.Sync2(new(UserShortenedUrl))

	return storeService
}

func GetLongUrl(shortUrl string) string {

	var shortenedUrl ShortenedUrl
	_, err := storeService.UrlShortenerDb.Table(&shortenedUrl).Where("short_url = ?", shortUrl).Get(&shortenedUrl)

	if err != nil {
		fmt.Println("Failed to fetch rows (storeService.GetLongUrl()):\n", err)
		// TODO: log error messages
		// storeService.UrlShortenerDb.Logger().Errorf()
	}

	fmt.Println(shortenedUrl)

	return shortenedUrl.LongUrl
}

func SaveUrl(shortUrl string, longUrl string, userId string) {
	// TODO: check if the shortUrl already exists WHERE userId
	// TODO: save userId and shortenedUrlId in userShortenedUrl (_, err -> affected, err -> affected.id)

	id := uuid.New()
	fmt.Println(id) // id (blob) -> id (string) hmm
	createdAt := time.Now()

	var shortenedUrl = ShortenedUrl{
		Id:        id,
		CreatedAt: createdAt,
		ShortUrl:  shortUrl,
		LongUrl:   longUrl,
	}

	_, err := storeService.UrlShortenerDb.Insert(&shortenedUrl)

	if err != nil {
		fmt.Println("Failed to insert url data:\n", err)
	}
}
