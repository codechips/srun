package api

import (
	"github.com/gin-gonic/gin"
	"srun/internal/core"
)

func SetupRoutes(r *gin.Engine, pm *core.ProcessManager) {
	// Job management endpoints
	r.POST("/jobs", createJobHandler(pm))
	r.GET("/jobs", listJobsHandler(pm))
	r.GET("/jobs/:id", getJobHandler(pm))
	r.DELETE("/jobs/:id", stopJobHandler(pm))
	r.POST("/jobs/:id/restart", restartJobHandler(pm))
	
	// Log streaming endpoint
	r.GET("/jobs/:id/logs", streamLogsHandler(pm))
}

func createJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented"})
	}
}

func listJobsHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented"})
	}
}

func getJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented"})
	}
}

func stopJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented"})
	}
}

func restartJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented"})
	}
}

func streamLogsHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(501, gin.H{"error": "not implemented"})
	}
}
