package main

import (
	"log"
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

	r := gin.Default()
	api.SetupRoutes(r, pm)
	r.Run(":8080")
}
