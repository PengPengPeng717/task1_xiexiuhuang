package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func APIKey() gin.HandlerFunc {
	key := os.Getenv("API_KEY")
	if key == "" {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		if c.GetHeader("X-API-Key") == key || c.Query("api_key") == key {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
