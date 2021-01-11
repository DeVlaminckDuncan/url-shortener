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

	r.POST("/short-urls", func(c *gin.Context) {
		handler.CreateShortURL(c)
	})

	r.PUT("/short-urls/:id", func(c *gin.Context) {
		handler.UpdateShortURL(c)
	})

	r.DELETE("/short-urls/:id", func(c *gin.Context) {
		handler.DeleteShortURL(c)
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

	r.GET("/user/:userID", func(c *gin.Context) {
		handler.GetUser(c)
	})

	r.PUT("/user/:userID", func(c *gin.Context) {
		handler.UpdateUser(c)
	})

	r.DELETE("/user/:userID", func(c *gin.Context) {
		handler.DeleteUser(c)
	})

	r.NoRoute(func(c *gin.Context) {
		shortURL := c.Request.URL.Path[1:]

		if len(shortURL) == 8 {
			handler.RedirectShortURL(c)
		} else {
			handler.NotFound(c)
		}
	})

	store.InitializeStore()

	err := r.Run(":9001")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server:\n%v", err))
	}
}
