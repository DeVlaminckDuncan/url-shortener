package main

import (
	"fmt"

	"github.com/devlaminckduncan/url-shortener/handler"
	"github.com/devlaminckduncan/url-shortener/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load() // load environment variables

	r := gin.Default()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:8080", "https://go-url-shortener.netlify.app"}
	corsConfig.AddAllowHeaders("Origin", "Authorization")
	corsConfig.AddAllowHeaders("OPTIONS", "GET", "POST", "PUT", "DELETE")

	// Must use CORS before defining any routes or they won't be able to use CORS!
	r.Use(cors.New(corsConfig))

	r.POST("/api/short-urls", func(c *gin.Context) {
		handler.CreateShortURL(c)
	})

	r.PUT("/api/short-urls/:id", func(c *gin.Context) {
		handler.UpdateShortURL(c)
	})

	r.DELETE("/api/short-urls/:id", func(c *gin.Context) {
		handler.DeleteShortURL(c)
	})

	r.GET("/api/short-urls/:userID", func(c *gin.Context) {
		handler.GetUserShortenedURLs(c)
	})

	r.POST("/api/signup", func(c *gin.Context) {
		handler.CreateUser(c)
	})

	r.POST("/api/login", func(c *gin.Context) {
		handler.CheckUserLogin(c)
	})

	r.GET("/api/user/:userID", func(c *gin.Context) {
		handler.GetUser(c)
	})

	r.PUT("/api/user/:userID", func(c *gin.Context) {
		handler.UpdateUser(c)
	})

	r.DELETE("/api/user/:userID", func(c *gin.Context) {
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
