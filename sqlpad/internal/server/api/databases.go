package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sqlpad/sqlpad/internal/infrastructure/database"
)

// ListDatabasesHandler returns all SQLite databases in the data directory.
func ListDatabasesHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		databases, err := database.ListDatabases(dataDir)
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
