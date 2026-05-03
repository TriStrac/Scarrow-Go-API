package middlewares

import (
	"os"

	"github.com/gin-gonic/gin"
)

// DevBypassMiddleware allows bypass with a special header for development/testing
// DO NOT use in production!
func DevBypassMiddleware() gin.HandlerFunc {
	bypassSecret := os.Getenv("DEV_BYPASS_SECRET")
	if bypassSecret == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		bypassHeader := c.GetHeader("X-Dev-Bypass")
		if bypassHeader == bypassSecret {
			c.Set("userID", "dev-bypass-user")
			c.Next()
			return
		}
		c.Next()
	}
}