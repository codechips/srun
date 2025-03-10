package api

import (
	"fmt"
	"io"
	"net/http"
	"srun/internal/core"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateJobRequest struct {
	Command string `json:"command" binding:"required"`
}

func SetupRoutes(r *gin.Engine, pm *core.ProcessManager) {
	// Job management endpoints
	r.POST("/api/jobs", createJobHandler(pm))
	r.GET("/api/jobs", listJobsHandler(pm))
	r.GET("/api/jobs/:id", getJobHandler(pm))
	r.DELETE("/api/jobs/:id", stopJobHandler(pm))
	r.POST("/api/jobs/:id/restart", restartJobHandler(pm))

	// Log streaming endpoint
	r.GET("/api/jobs/:id/logs", streamLogsHandler(pm))
}

func listJobsHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobs, err := pm.ListJobs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to list jobs: " + err.Error(),
			})
			return
		}

		// Convert jobs to response format
		var response []gin.H
		for _, job := range jobs {
			resp := gin.H{
				"id":        job.ID,
				"command":   job.Command,
				"status":    job.Status,
				"pid":       job.PID,
				"startedAt": job.StartedAt.Format(time.RFC3339),
			}
			// Only include completedAt if it's not zero time
			if !job.CompletedAt.IsZero() {
				resp["completedAt"] = job.CompletedAt.Format(time.RFC3339)
			}
			response = append(response, resp)
		}

		c.JSON(http.StatusOK, response)
	}
}

func getJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		job, err := pm.GetJob(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get job: " + err.Error(),
			})
			return
		}
		if job == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job not found",
			})
			return
		}

		resp := gin.H{
			"id":        job.ID,
			"command":   job.Command,
			"status":    job.Status,
			"pid":       job.PID,
			"startedAt": job.StartedAt.Format(time.RFC3339),
		}
		// Only include completedAt if it's not zero time
		if !job.CompletedAt.IsZero() {
			resp["completedAt"] = job.CompletedAt.Format(time.RFC3339)
		}
		c.JSON(http.StatusOK, resp)
	}
}

func stopJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := pm.StopJob(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to stop job: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Job stopped successfully",
		})
	}
}

func restartJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		job, err := pm.RestartJob(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to restart job: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          job.ID,
			"command":     job.Command,
			"status":      job.Status,
			"pid":         job.PID,
			"startedAt":   job.StartedAt.Format(time.RFC3339),
			"completedAt": job.CompletedAt.Format(time.RFC3339),
		})
	}
}

func streamLogsHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Check if job exists
		job, err := pm.GetJob(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get job: " + err.Error(),
			})
			return
		}
		if job == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Job not found",
			})
			return
		}

		// Get historical logs
		logs, err := pm.Store.GetJobLogs(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get logs: " + err.Error(),
			})
			return
		}

		// Set headers for plain text streaming
		c.Header("Content-Type", "text/plain")
		c.Header("X-Content-Type-Options", "nosniff")

		// Send historical logs first
		for _, log := range logs {
			fmt.Fprintln(c.Writer, log.RawText)
		}
		c.Writer.Flush()

		// If job is still running, stream new logs
		if job.Status == "running" {
			// Create channel for this client
			clientChan := make(chan core.LogMessage, 10)

			// Subscribe to log channel
			go func() {
				for msg := range pm.LogChan {
					if msg.JobID == id {
						clientChan <- msg
					}
				}
			}()

			// Stream logs until connection closes
			c.Stream(func(w io.Writer) bool {
				select {
				case msg := <-clientChan:
					fmt.Fprintln(w, msg.RawText)
					return true
				case <-c.Done():
					close(clientChan)
					return false
				}
			})
		}
	}
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

		// Start the job without timeout
		job, err := pm.StartJob(req.Command)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to start job: " + err.Error(),
			})
			return
		}

		// Return job information
		resp := gin.H{
			"id":        job.ID,
			"command":   req.Command,
			"status":    job.Status,
			"pid":       job.PID,
			"startedAt": job.StartedAt.Format(time.RFC3339),
		}
		// Only include completedAt if it's not zero time
		if !job.CompletedAt.IsZero() {
			resp["completedAt"] = job.CompletedAt.Format(time.RFC3339)
		}
		c.JSON(http.StatusCreated, resp)
	}
}
