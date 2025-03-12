package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"srun/internal/api"
	"srun/internal/core"
	"srun/internal/static"

	"github.com/gin-gonic/gin"
)

func main() {
	store, err := core.NewSQLiteStorage("srun.db")
	if err != nil {
		log.Fatal(err)
	}

	pm := &core.ProcessManager{
		Jobs:    make(map[string]*core.Job),
		Store:   store,
		LogChan: make(chan core.LogMessage, 1000),
	}

	// Port configuration with flag and env var
	var port string
	flag.StringVar(&port, "port", "8000", "Port to listen on")
	flag.Parse()

	// Check environment variable if flag not set
	if port == "" {
		if envPort := os.Getenv("SRUN_PORT"); envPort != "" {
			port = envPort
		}
	}

	r := gin.Default()
	
	// API routes
	api.SetupRoutes(r, pm)

	// Serve static files
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(static.StaticFiles))
	})
	r.NoRoute(func(c *gin.Context) {
		// Try to serve static file
		if _, err := static.StaticFiles.Open(c.Request.URL.Path[1:]); err == nil {
			c.FileFromFS(c.Request.URL.Path[1:], http.FS(static.StaticFiles))
			return
		}
		// Fall back to index.html for client-side routing
		c.FileFromFS("index.html", http.FS(static.StaticFiles))
	})

	r.Run(":" + port)
}
