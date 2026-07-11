package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const appVersion = "v0.0.1"

// HealthHandler returns server health status, version, and timestamp.
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version":   appVersion,
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}
