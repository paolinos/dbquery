package api

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/infrastructure/database"
)

// getDB opens a database connection for the given db name.
func getDB(dataDir, dbName string) (*database.SQLiteDB, error) {
	dbPath := database.GetDBPath(dataDir, dbName)
	return database.OpenExisting(dbPath)
}

// ListTablesHandler returns all tables in the specified database.
func ListTablesHandler(dataDir string) gin.HandlerFunc {
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
				"error":   "Failed to list tables",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    tables,
			"message": "ok",
		})
	}
}

// GetTableSchemaHandler returns the schema for a specific table.
func GetTableSchemaHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbName := c.Param("db")
		tableName := c.Param("table")

		if dbName == "" || tableName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database and table names are required"})
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

		columns, err := db.GetTableColumns(tableName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get table schema",
				"details": err.Error(),
			})
			return
		}

		// Also get CREATE TABLE statement
		var createSQL sql.NullString
		err = db.GetDB().QueryRow(
			"SELECT sql FROM sqlite_master WHERE type='table' AND name=?",
			tableName,
		).Scan(&createSQL)

		schema := ""
		if err == nil && createSQL.Valid {
			schema = createSQL.String
		}

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"name":    tableName,
				"schema":  schema,
				"columns": columns,
			},
			"message": "ok",
		})
	}
}

// GetTableDataHandler returns paginated data from a table.
func GetTableDataHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		dbName := c.Param("db")
		tableName := c.Param("table")

		if dbName == "" || tableName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Database and table names are required"})
			return
		}

		page := 1
		perPage := 100
		// Parse query params if available
		if p, ok := c.GetQuery("page"); ok {
			fmt.Sscanf(p, "%d", &page)
		}
		if pp, ok := c.GetQuery("per_page"); ok {
			fmt.Sscanf(pp, "%d", &perPage)
		}
		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 1000 {
			perPage = 100
		}

		offset := (page - 1) * perPage
		query := fmt.Sprintf("SELECT * FROM [%s] LIMIT %d OFFSET %d", tableName, perPage, offset)

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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to query table",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data":    result,
			"page":    page,
			"perPage": perPage,
			"message": "ok",
		})
	}
}
