package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/core/models"
)

// ExecuteQueryHandler executes a SQL query against the specified database.
func ExecuteQueryHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbName := c.Param("db")
		if dbName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database name is required"})
			return
		}

		var req models.QueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request body",
				"details": "Query is required",
			})
			return
		}

		query := strings.TrimSpace(req.Query)
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Query cannot be empty"})
			return
		}

		// Basic validation: prevent execution of multiple statements (for safety)
		if strings.Count(query, ";") > 1 {
			// Allow trailing semicolons
			trimmed := strings.TrimRight(query, "; \t\n\r")
			if strings.Count(trimmed, ";") > 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Multiple statements are not supported",
					"details": "Please execute one statement at a time",
				})
				return
			}
		}

		// Remove trailing semicolons for consistency
		query = strings.TrimRight(query, "; \t\n\r")

		db, err := getDB(dataDir, dbName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Database not found",
				"details": err.Error(),
			})
			return
		}
		defer db.Close()

		result, err := db.ExecuteQuery(query)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Query execution failed",
				"details": err.Error(),
			})
			return
		}

		// Log query for audit
		fmt.Printf("[QUERY] db=%s query=%q rows=%d\n", dbName, query, len(result.Rows))

		c.JSON(http.StatusOK, gin.H{
			"data":    result,
			"message": "ok",
		})
	}
}

// GetTableAutocompleteHandler returns table and column names for autocomplete.
func GetTableAutocompleteHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbName := c.Param("db")
		if dbName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database name is required"})
			return
		}

		db, err := getDB(dataDir, dbName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Database not found",
				"details": err.Error(),
			})
			return
		}
		defer db.Close()

		tables, err := db.GetTables()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get tables",
				"details": err.Error(),
			})
			return
		}

		// Build autocomplete data
		type TableSuggestion struct {
			Name    string   `json:"name"`
			Columns []string `json:"columns"`
		}
		suggestions := make([]TableSuggestion, len(tables))
		for i, t := range tables {
			cols := make([]string, len(t.Columns))
			for j, col := range t.Columns {
				cols[j] = col.Name
			}
			suggestions[i] = TableSuggestion{
				Name:    t.Name,
				Columns: cols,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    suggestions,
			"message": "ok",
		})
	}
}
