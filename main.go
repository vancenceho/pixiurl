package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	handlers "github.com/vancenceho/pixiurl/handler"
	"github.com/vancenceho/pixiurl/store"
)
func main() {
	r := gin.Default()
	r.GET("/", func(c*gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello PixiURL, just like TinyURL; it is a URL shortener! 🚀 ",
		})
	})

	r.POST("api/v1/create-short-url", func(c *gin.Context) {
		handlers.CreateShortUrl(c)
	})

	r.GET("/:shortUrl", func(c *gin.Context) {
		handlers.HandleShortUrlRedirect(c)
	})

	// NOTE: store initialization happens here
	store.InitializeStore()
	
	err := r.Run(":9808")

	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error: %v", err))
	}
}