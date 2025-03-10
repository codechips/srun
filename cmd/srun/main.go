package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"srun/internal/api"
	"srun/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
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
	var port int
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.Parse()
	
	// Check environment variable if flag not set
	if port == 8080 {
		if envPort := os.Getenv("SRUN_PORT"); envPort != "" {
			port = int(envPort)
		}
	}

	r := gin.Default()
	api.SetupRoutes(r, pm)
	r.Run(":" + strconv.Itoa(port))
}
