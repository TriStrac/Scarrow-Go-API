package middlewares

import (
	"fmt"

	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

var routeActions = map[string]string{
	"POST /api/users/":                    "Registered new account",
	"POST /api/users/verify-registration": "Verified account",
	"POST /api/users/login":               "Logged in",
	"POST /api/users/changePassword":      "Changed password",
	"PATCH /api/users/:userId":            "Updated profile",
	"POST /api/groups/":                   "Created a group",
	"POST /api/groups/join":               "Joined a group via code",
	"POST /api/groups/member":             "Added a member to group",
	"DELETE /api/groups/member":           "Removed a member from group",
	"POST /api/groups/leave":              "Left the group",
	"POST /api/device/":                   "Added a new device",
	"POST /api/messages/":                 "Sent a message",
}

// ActivityLogMiddleware logs user actions with human-readable descriptions
func ActivityLogMiddleware(activityLogService service.ActivityLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			method := c.Request.Method
			path := c.FullPath()

			if method == "GET" {
				return
			}

			userID, exists := c.Get("userID")
			if !exists {
				return
			}

			actionKey := fmt.Sprintf("%s %s", method, path)
			actionDesc, ok := routeActions[actionKey]
			if !ok {
				actionDesc = actionKey // Fallback to raw path if not mapped
			}

			module := extractModule(path)
			_ = activityLogService.LogActivity(userID.(string), actionDesc, module)
		}
	}
}

func extractModule(path string) string {
	if len(path) < 10 {
		return "general"
	}
	// /api/users/... -> users
	// /api/groups/... -> groups
	parts := path[5:]
	for i, char := range parts {
		if char == '/' {
			return parts[:i]
		}
	}
	return parts
}
