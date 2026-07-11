package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/infrastructure/database"
)

// ListDatabasesHandler returns all SQLite databases owned by the authenticated user.
func ListDatabasesHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")

		databases, err := database.ListDatabases(dataDir, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to list databases",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    databases,
			"message": "ok",
		})
	}
}
