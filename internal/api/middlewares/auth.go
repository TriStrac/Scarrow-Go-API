package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware validates the JWT token in the Authorization header and checks user verification
func AuthMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check dev bypass first
		bypassSecret := os.Getenv("DEV_BYPASS_SECRET")
		if bypassSecret != "" && c.GetHeader("X-Dev-Bypass") == bypassSecret {
			c.Set("userID", "dev-bypass-user")
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Expecting "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		secret := os.Getenv("JWT_SECRET")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		// Production-level Security Check: Ensure user exists and is verified
		user, err := userRepo.FindByID(userID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User no longer exists"})
			c.Abort()
			return
		}

		if !user.IsVerified {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account not verified. Please verify your phone number."})
			c.Abort()
			return
		}

		// Set the user ID in the context so subsequent handlers can access it
		c.Set("userID", userID)
		c.Next()
	}
}
