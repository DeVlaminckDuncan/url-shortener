package main

import (
	"fmt"

	"github.com/devlaminckduncan/url-shortener/handler"
	"github.com/devlaminckduncan/url-shortener/store"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // load environment variables

	r := gin.Default()

	// r.GET("/", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "Welcome to the URL Shortener API",
	// 	})
	// })

	r.POST("/create-short-url", func(c *gin.Context) {
		// TODO: check if the user is logged in
		// can be checked here or in handlers.go idk
		// or in a function that checks it or something

		handler.CreateShortURL(c)
	})

	r.GET("/short-urls/:userID", func(c *gin.Context) {
		// TODO: security token
		handler.GetUserShortenedURLs(c)
	})

	r.POST("/signup", func(c *gin.Context) {
		handler.CreateUser(c)
	})

	r.POST("/login", func(c *gin.Context) {
		handler.CheckUserLogin(c)
	})

	r.NoRoute(func(c *gin.Context) {
		// TODO: check if the path is an 8 character long base58 string, return 404 if not -> handler.NotFound()

		handler.RedirectShortURL(c)
	})

	store.InitializeStore()

	err := r.Run(":9001")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server:\n%v", err))
	}

}
