package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"srun/internal/api"
	"path/filepath"
	"srun/internal/core"
	"srun/internal/static"

	"path"
	"fmt"
	"io/fs"
	"strings"

	"github.com/gin-gonic/gin"
)

func defaultDBPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Printf("Warning: Couldn't get user config directory: %v", err)
		return "srun.db"
	}
	
	appDir := filepath.Join(configDir, "srun")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		log.Printf("Warning: Couldn't create application directory: %v", err)
		return "srun.db"
	}
	
	return filepath.Join(appDir, "srun.db")
}

var dbPath string
var port string
var basePathFlag string
var trustedProxiesFlag string

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
	// Configure flags
	flag.StringVar(&port, "port", "8000", "Port to listen on")
	flag.StringVar(&dbPath, "db", defaultDBPath(), "SQLite database path")
	flag.StringVar(&basePathFlag, "base-path", "/", "Base path for the application (e.g., / or /srun)")
	flag.StringVar(&trustedProxiesFlag, "trusted-proxies", "", "Comma-separated list of trusted proxy IPs (e.g., '127.0.0.1,192.168.1.100')")
	flag.Parse()

	store, err := core.NewSQLiteStorage(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	pm := &core.ProcessManager{
		Jobs:    make(map[string]*core.Job),
		Store:   store,
		LogChan: make(chan core.LogMessage, 1000),
	}

	if envPort := os.Getenv("SRUN_PORT"); port == "" && envPort != "" {
		port = envPort
	}

	r := gin.Default()

	// Set trusted proxies if the flag is provided
	if trustedProxiesFlag != "" {
		proxies := strings.Split(trustedProxiesFlag, ",")
		cleanedProxies := make([]string, 0, len(proxies))
		for _, p := range proxies {
			trimmedP := strings.TrimSpace(p)
			if trimmedP != "" {
				cleanedProxies = append(cleanedProxies, trimmedP)
			}
		}
		if len(cleanedProxies) > 0 {
			r.SetTrustedProxies(cleanedProxies)
		}
	}

	// Normalize base path
	cleanBasePath := path.Clean("/" + basePathFlag)
	if cleanBasePath != "/" && strings.HasSuffix(cleanBasePath, "/") {
		cleanBasePath = cleanBasePath[:len(cleanBasePath)-1]
	}
	if cleanBasePath == "" {
		cleanBasePath = "/"
	}

	// Determine the router group
	var routerGroup gin.IRouter
	if cleanBasePath == "/" {
		routerGroup = r
	} else {
		routerGroup = r.Group(cleanBasePath)
	}

	// API routes first
	api.SetupRoutes(routerGroup, pm)

	// Create a filesystem handler for the embedded files
	distFS, err := fs.Sub(static.StaticFiles, "dist")
	if err != nil {
		panic(err)
	}

	serveIndexHTML := func(c *gin.Context) {
		indexHTMLBytes, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Could not load index.html: "+err.Error())
			return
		}

		// Inject base tag and JavaScript variable
		htmlContent := string(indexHTMLBytes)
		baseHref := cleanBasePath
		if !strings.HasSuffix(baseHref, "/") {
			baseHref += "/"
		}

		// Inject <base href="...">
		if cleanBasePath != "/" {
			htmlContent = strings.Replace(htmlContent, "<head>", fmt.Sprintf("<head>\n    <base href=\"%s\">", baseHref), 1)
		}

		// Inject JavaScript global for base path
		scriptToInject := fmt.Sprintf("<script>window.APP_BASE_PATH = \"%s\";</script>", cleanBasePath)
		htmlContent = strings.Replace(htmlContent, "<head>", "<head>\n    "+scriptToInject, 1)

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlContent))
	}

	// Static file handlers in specific order
	// 1. Root path
	routerGroup.GET("/", serveIndexHTML)

	// 2. Assets directory
	routerGroup.GET("/assets/*filepath", func(c *gin.Context) {
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
	routerGroup.GET("/nazar.svg", func(c *gin.Context) {
		data, err := fs.ReadFile(distFS, "nazar.svg")
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", data)
	})

	// 4. Finally, catch-all for SPA routes
	if _, ok := routerGroup.(*gin.RouterGroup); ok {
		// If it's a group, define a catch-all for SPA routes within the group.
		// This must be after specific file/asset routes for that group.
		routerGroup.GET("/*any", serveIndexHTML)
	} else {
		// If routerGroup is the main engine (r), use NoRoute for SPA catch-all.
		r.NoRoute(serveIndexHTML)
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Using database at: %s", dbPath)
	log.Printf("Application base path: %s", cleanBasePath)
	r.Run(":" + port)
}
