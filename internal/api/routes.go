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
package api

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "srun/internal/core"
)

type CreateJobRequest struct {
    Command string        `json:"command" binding:"required"`
    Timeout time.Duration `json:"timeout"` // in seconds
}

func createJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req CreateJobRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Invalid request: " + err.Error(),
            })
            return
        }

        // Set default timeout if not specified
        if req.Timeout == 0 {
            req.Timeout = time.Hour // 1 hour default
        }

        // Validate timeout range (5m to 8h)
        if req.Timeout < 5*time.Minute || req.Timeout > 8*time.Hour {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Timeout must be between 5 minutes and 8 hours",
            })
            return
        }

        // Start the job
        job, err := pm.StartJob(req.Command, req.Timeout)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to start job: " + err.Error(),
            })
            return
        }

        // Return job information
        c.JSON(http.StatusCreated, gin.H{
            "id":         job.ID,
            "command":    req.Command,
            "status":     job.Status,
            "started_at": job.StartedAt,
        })
    }
}
