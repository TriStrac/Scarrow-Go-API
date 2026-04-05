package middlewares

import (
	"fmt"

	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

// ActivityLogMiddleware logs user actions based on mutating HTTP methods
func ActivityLogMiddleware(activityLogService service.ActivityLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Proceed with the request
		c.Next()

		// Only log for successful mutating operations (POST, PATCH, DELETE, PUT)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			method := c.Request.Method
			if method == "POST" || method == "PATCH" || method == "DELETE" || method == "PUT" {
				userID, exists := c.Get("userID")
				if !exists {
					// Some operations like login/register might not have userID in context yet
					// but we might want to log them too if we can extract ID later or if it's public.
					// For now, let's just log if userID exists.
					return
				}

				path := c.FullPath()
				action := fmt.Sprintf("%s %s", method, path)
				module := extractModule(path)

				// Log in background or directly
				// s.LogActivity(userID.(string), action, module)
				_ = activityLogService.LogActivity(userID.(string), action, module)
			}
		}
	}
}

func extractModule(path string) string {
	// Simple module extraction based on path (e.g., /api/users/foo -> users)
	// You can make this more robust as needed.
	if len(path) < 5 {
		return "unknown"
	}
	// skip "/api/"
	parts := fmt.Sprintf("%v", path)
	// Just a simple placeholder for now
	return parts
}
