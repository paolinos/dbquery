package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/infrastructure/database"
)

// AuthMiddleware validates JWT tokens from the Authorization header.
func AuthMiddleware(authDB *database.AuthDB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization format, expected 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate against the database
		valid, userID, err := authDB.IsTokenValid(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to validate token",
			})
			c.Abort()
			return
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Look up the user to get the username
		user, err := authDB.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Set user info in context for downstream handlers
		c.Set("userID", user.ID)
		c.Set("username", user.Username)
		c.Set("token", token)

		c.Next()
	}
}
