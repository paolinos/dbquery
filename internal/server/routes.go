package server

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/dbquery/dbquery/internal/infrastructure/database"
	"github.com/dbquery/dbquery/internal/server/api"
	"github.com/dbquery/dbquery/internal/server/middleware"
)

// SetupRouter creates and configures the Gin router with all routes.
func SetupRouter(dataDir string, frontendPath string, authDB *database.AuthDB) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(middleware.SetupCORS())

	// Public API routes (no auth required)
	publicGroup := r.Group("/api")
	{
		// Health check
		publicGroup.GET("/health", api.HealthHandler())

		// Check if any users exist (for first-time setup)
		publicGroup.GET("/auth/has-users", api.HasUsersHandler(authDB))

		// Register first user (only when no users exist)
		publicGroup.POST("/auth/register", api.RegisterHandler(authDB))

		// Login
		publicGroup.POST("/auth/login", api.LoginHandler(authDB))
	}

	// Protected API routes (auth required)
	protectedGroup := r.Group("/api")
	protectedGroup.Use(middleware.AuthMiddleware(authDB))
	{
		// List all databases (filtered by user)
		protectedGroup.GET("/databases", api.ListDatabasesHandler(dataDir))

		// List tables in a database
		protectedGroup.GET("/databases/:db/tables", api.ListTablesHandler(dataDir))

		// Get table schema
		protectedGroup.GET("/databases/:db/tables/:table/schema", api.GetTableSchemaHandler(dataDir))

		// Get table data (paginated)
		protectedGroup.GET("/databases/:db/tables/:table/data", api.GetTableDataHandler(dataDir))

		// Execute SQL query
		protectedGroup.POST("/databases/:db/query", api.ExecuteQueryHandler(dataDir))

		// Get autocomplete suggestions
		protectedGroup.GET("/databases/:db/autocomplete", api.GetTableAutocompleteHandler(dataDir))

		// Upload Excel file
		protectedGroup.POST("/upload", api.UploadExcelHandler(dataDir))

		// Auth: logout
		protectedGroup.POST("/auth/logout", api.LogoutHandler(authDB))

		// Auth: current user info
		protectedGroup.GET("/auth/me", api.MeHandler(authDB))
	}

	// Serve frontend static files
	serveFrontend(r, frontendPath)

	return r
}

// serveFrontend serves the frontend static files from disk.
// If the directory doesn't exist, it sets up a fallback that returns a helpful message.
func serveFrontend(r *gin.Engine, frontendPath string) {
	absPath, err := filepath.Abs(frontendPath)
	if err != nil {
		absPath = frontendPath
	}

	// Check if frontend directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		// Frontend not built yet — set up a placeholder
		r.NoRoute(func(c *gin.Context) {
			if c.Request.URL.Path == "/" {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusOK, `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>DBQuery</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { margin: 0; padding: 2rem; }
    .container { max-width: 800px; margin: 0 auto; text-align: center; }
    pre { text-align: left; background: var(--card-background-color); padding: 1rem; border-radius: 8px; }
  </style>
</head>
<body>
  <div class="container">
    <h1>🚀 DBQuery</h1>
    <p>Excel to SQLite Manager with SQL Query Interface</p>
    <article>
      <p>The frontend is not built yet. Please build it with:</p>
      <pre><code>cd web && npm install && npm run build</code></pre>
      <p>Then restart the server.</p>
      <p>In the meantime, you can use the API directly:</p>
      <ul style="text-align: left;">
        <li><code>POST /api/auth/login</code> — Login</li>
        <li><code>GET /api/databases</code> — List databases</li>
        <li><code>GET /api/databases/:db/tables</code> — List tables</li>
        <li><code>POST /api/databases/:db/query</code> — Execute SQL</li>
        <li><code>POST /api/upload</code> — Upload Excel file</li>
      </ul>
    </article>
  </div>
</body>
</html>`)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		})
		return
	}

	// Serve static files
	r.Use(func(c *gin.Context) {
		// Skip API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.Next()
			return
		}

		requestPath := c.Request.URL.Path
		if requestPath == "/" {
			requestPath = "/index.html"
		}

		fullPath := filepath.Join(absPath, requestPath)
		if _, err := os.Stat(fullPath); err == nil {
			c.File(fullPath)
			c.Abort()
			return
		}

		// For SPA routing, serve index.html
		c.File(filepath.Join(absPath, "index.html"))
		c.Abort()
	})
}
