package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/infrastructure/database"
	"github.com/dbquery/dbquery/internal/infrastructure/excel"
)

const maxUploadSize = 50 << 20 // 50MB

// UploadExcelHandler handles Excel file upload and imports data into SQLite.
// The resulting database file is prefixed with the user's ID for isolation.
func UploadExcelHandler(dataDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")

		// Limit upload size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "No file provided",
				"details": "Please upload an Excel file with the field name 'file'",
			})
			return
		}
		defer file.Close()

		// Validate file extension
		if !excel.IsExcelFile(header.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid file format",
				"details": "Please upload an .xlsx file",
			})
			return
		}

		// Read file data
		fileData, err := io.ReadAll(file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to read file",
				"details": err.Error(),
			})
			return
		}

		// Parse Excel file
		parser := excel.NewParser()
		sheets, err := parser.ParseReader(fileData, header.Filename)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Failed to parse Excel file",
				"details": err.Error(),
			})
			return
		}

		// Determine database name from the filename (without extension)
		dbName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
		// Sanitize db name
		dbName = sanitizeDBName(dbName)
		if dbName == "" {
			dbName = "imported"
		}

		// Get user-scoped database path
		dbPath := database.GetUserDBPath(dataDir, userID, dbName)

		// Ensure data directory exists
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create data directory",
				"details": err.Error(),
			})
			return
		}

		// Open or create the database
		db, err := database.NewSQLiteDB(dbPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to open database",
				"details": err.Error(),
			})
			return
		}
		defer db.Close()

		// Process each sheet
		var createdTables []struct {
			Name    string `json:"name"`
			Rows    int    `json:"rows"`
			Columns int    `json:"columns"`
		}

		for _, sheet := range sheets {
			tableName := excel.NormalizeSheetName(sheet.Name)
			if tableName == "" {
				continue
			}

			// Create table if not exists, add missing columns
			if err := db.CreateTable(tableName, sheet.Headers, sheet.Rows); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   fmt.Sprintf("Failed to create table for sheet '%s'", sheet.Name),
					"details": err.Error(),
				})
				return
			}

			// Insert data
			if err := db.InsertRows(tableName, sheet.Headers, sheet.Rows); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   fmt.Sprintf("Failed to insert data for sheet '%s'", sheet.Name),
					"details": err.Error(),
				})
				return
			}

			createdTables = append(createdTables, struct {
				Name    string `json:"name"`
				Rows    int    `json:"rows"`
				Columns int    `json:"columns"`
			}{
				Name:    tableName,
				Rows:    len(sheet.Rows),
				Columns: len(sheet.Headers),
			})
		}

		// Get full table list for response
		tables, _ := db.GetTables()

		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"database": dbName,
				"tables":   tables,
				"imported": createdTables,
			},
			"message": fmt.Sprintf("Successfully imported %d sheet(s) into database '%s'", len(sheets), dbName),
		})
	}
}

// sanitizeDBName removes unsafe characters from database names.
func sanitizeDBName(name string) string {
	name = strings.TrimSpace(name)
	// Remove path separators and special chars
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, ":", "")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
