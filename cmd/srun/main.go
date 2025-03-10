package main

import (
	"flag"
	"log"
	"os"
	"srun/internal/api"
	"srun/internal/core"

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
	flag.StringVar(&port, "port", "", "Port to listen on")
	flag.Parse()

	// Check environment variable if flag not set
	if port == "" {
		if envPort := os.Getenv("SRUN_PORT"); envPort != "" {
			port = envPort
		}
	}

	r := gin.Default()
	api.SetupRoutes(r, pm)
	r.Run(":" + port)
}
