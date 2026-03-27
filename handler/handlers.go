package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/vancenceho/pixiurl/shortener"
	"github.com/vancenceho/pixiurl/store"
	"net/http"

)

// Request model definition
type UrlCreationRequest struct {
	LongUrl string `json:"long_url" binding:"required"`
	UserId string `json:"user_id" binding:"required"`
}

func CreateShortUrl(c *gin.Context) {
	// TODO: implementation to be added here
	var creationRequest UrlCreationRequest

	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortUrl := shortener.GenerateShortLink(creationRequest.LongUrl, creationRequest.UserId)
	store.SaveUrlMapping(shortUrl, creationRequest.LongUrl, creationRequest.UserId)

	host := "http://localhost:9808/api/v1/"

	c.JSON(200, gin.H{
		"message": "short url created successfully",
		"short_url": host + "shorturl/" + shortUrl,
	})
}

func HandleShortUrlRedirect(c *gin.Context) {
	// TODO: Implementation to be added here
}


