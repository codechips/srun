package main

import (
	"log"
	"srun/internal/core"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine, pm *core.ProcessManager) {
	// Example route setup
	r.POST("/jobs", func(c *gin.Context) {
		// Implementation here
	})
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

	r := gin.Default()
	setupRoutes(r, pm)
	r.Run(":8080")
}
