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

	// Create a filesystem handler for the embedded files
	staticFS := http.FS(static.StaticFiles)
	// fileServer := http.FileServer(staticFS)

	// Serve static files
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", staticFS)
	})

	// r.GET("/assets/*filepath", func(c *gin.Context) {
	// 	c.Request.URL.Path = c.Param("filepath")
	// 	fileServer.ServeHTTP(c.Writer, c.Request)
	// })

	r.Run(":" + port)
}
