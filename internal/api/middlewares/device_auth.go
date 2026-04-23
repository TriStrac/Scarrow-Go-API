package middlewares

import (
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
)

func DeviceAuthMiddleware(deviceRepo repository.DeviceRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.GetHeader("X-Device-ID")
		deviceSecret := c.GetHeader("X-Device-Secret")

		if deviceID == "" || deviceSecret == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-Device-ID and X-Device-Secret headers required"})
			c.Abort()
			return
		}

		device, err := deviceRepo.FindByID(deviceID)
		if err != nil || device == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		if device.Secret != deviceSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
			c.Abort()
			return
		}

		c.Set("deviceID", deviceID)
		c.Next()
	}
}