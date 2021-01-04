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

	ginEngine := gin.Default()

	ginEngine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the URL Shortener API",
		})
	})

	ginEngine.POST("/create-short-url", func(c *gin.Context) {
		// TODO: check if the user is logged in
		// can be checked here or in handlers.go idk
		// or in a function that checks it or something

		handler.CreateShortUrl(c)
	})

	ginEngine.GET("/:shortUrl", func(c *gin.Context) {
		handler.RedirectShortUrl(c)
	})

	store.InitializeStore()

	err := ginEngine.Run(":9001")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server:\n%v", err))
	}

}
