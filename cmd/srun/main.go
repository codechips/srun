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
	// Handle vite.svg favicon
	r.GET("/vite.svg", func(c *gin.Context) {
		c.FileFromFS("vite.svg", http.FS(static.StaticFiles))
	})
	// Handle all assets
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets/"+c.Param("filepath"), http.FS(static.StaticFiles))
	})
	// All other routes fall back to index.html for client-side routing
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(static.StaticFiles))
	})

	r.Run(":" + port)
}
