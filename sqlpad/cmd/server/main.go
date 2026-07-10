package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	server "github.com/sqlpad/sqlpad/internal/server"
)

func main() {
	// Configuration flags
	port := flag.Int("port", 8080, "HTTP server port")
	dataDir := flag.String("data", "./data", "Directory for SQLite database files")
	frontendPath := flag.String("frontend", "./web/dist", "Path to frontend build directory")
	flag.Parse()

	// Ensure data directory exists
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory %s: %v", *dataDir, err)
	}

	// Setup router
	router := server.SetupRouter(*dataDir, *frontendPath, false)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("SQLPad server starting on http://localhost%s", addr)
	log.Printf("Data directory: %s", *dataDir)
	log.Printf("Frontend path: %s", *frontendPath)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
