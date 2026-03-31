package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vancenceho/pixiurl/shortener"
	"github.com/vancenceho/pixiurl/store"
)

// Request model definition
type UrlCreationRequest struct {
	LongUrl string `json:"long_url" binding:"required"`
	UserId string `json:"user_id" binding:"required"`
}

func CreateShortUrl(c *gin.Context) {
	// Get the request body and map it to the UrlCreationRequest struct
	var creationRequest UrlCreationRequest

	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate a short URL
	shortUrl := shortener.GenerateShortLink(creationRequest.LongUrl, creationRequest.UserId)

	// Save the URL mapping to the database
	store.SaveUrlMapping(shortUrl, creationRequest.LongUrl, creationRequest.UserId)

	// Return the short URL
	host := "http://pixi.url"

	c.JSON(200, gin.H{
		"message": "short url created successfully",
		"short_url": host + "/" + shortUrl,
	})
}

func HandleShortUrlRedirect(c *gin.Context) {
	// Get the short URL from the request parameters
	shortUrl := c.Params.ByName("shortUrl")

	initialUrl, err := store.RetrieveInitialUrl(shortUrl)
	if err != nil {
		// true system error (redis/postgres error)
		c.JSON(http.StatusInternalServerError, gin.H{"error | internal server error: ": err.Error()})

		return
	}

	if initialUrl == "" {
		// not found in redis or postgres
		c.JSON(http.StatusNotFound, gin.H{"error | short url not found: ": shortUrl})

		return
	}

	// Redirect to the initial URL
	c.Redirect(302, initialUrl)

}
