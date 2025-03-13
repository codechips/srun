package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"srun/internal/api"
	"srun/internal/core"
	"srun/internal/static"

	"fmt"
	"io/fs"
	"strings"

	"github.com/gin-gonic/gin"
)

func ListFilesHandler(c *gin.Context) {
	// Create a string builder to collect the file listing
	var sb strings.Builder

	// Function to recursively list files in a directory
	var listFiles func(string, int) error

	listFiles = func(dir string, level int) error {
		entries, err := fs.ReadDir(static.StaticFiles, dir)
		if err != nil {
			return err
		}

		indent := strings.Repeat("  ", level)
		for _, entry := range entries {
			// Add the file/directory name to the output
			sb.WriteString(fmt.Sprintf("%s- %s\n", indent, entry.Name()))

			// If it's a directory, recursively list its contents
			if entry.IsDir() {
				subDir := dir
				if dir != "" {
					subDir += "/"
				}
				subDir += entry.Name()

				if err := listFiles(subDir, level+1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Start listing from the root of the embedded filesystem
	err := listFiles("", 0)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error listing files: "+err.Error())
		return
	}

	// Return the file listing as plain text
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, sb.String())
}

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

	// API routes first
	api.SetupRoutes(r, pm)

	// Create a filesystem handler for the embedded files
	distFS, err := fs.Sub(static.StaticFiles, "dist")
	if err != nil {
		panic(err)
	}

	// Static file handlers in specific order
	// 1. Root path
	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not load index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 2. Assets directory
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		filepath = strings.TrimPrefix(filepath, "/")
		path := "assets/" + filepath

		data, err := fs.ReadFile(distFS, path)
		if err != nil {
			c.String(http.StatusNotFound, "Asset not found: "+path)
			return
		}

		contentType := "application/octet-stream"
		if strings.HasSuffix(path, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(path, ".css") {
			contentType = "text/css"
		} else if strings.HasSuffix(path, ".svg") {
			contentType = "image/svg+xml"
		}

		c.Data(http.StatusOK, contentType, data)
	})

	// 3. Specific static files (like nazar.svg)
	r.GET("/nazar.svg", func(c *gin.Context) {
		data, err := fs.ReadFile(distFS, "nazar.svg")
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	// 4. Finally, catch-all for SPA routes
	r.NoRoute(func(c *gin.Context) {
		data, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not load index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	r.Run(":" + port)
}
