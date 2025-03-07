package main

import (
    "log"
    "github.com/gin-gonic/gin"
    "srun/internal/core"
)

func main() {
    store, err := core.NewSQLiteStorage("srun.db")
    if err != nil {
        log.Fatal(err)
    }

    pm := &core.ProcessManager{
        jobs:    make(map[string]*core.Job),
        store:   store,
        logChan: make(chan core.LogMessage, 1000),
    }

    r := gin.Default()
    setupRoutes(r, pm)
    
    r.Run(":8080")
}
