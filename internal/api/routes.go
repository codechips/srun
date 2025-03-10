package api

import (
	"net/http"
	"srun/internal/core"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type CreateJobRequest struct {
	Command string `json:"command" binding:"required"`
}

func SetupRoutes(r *gin.Engine, pm *core.ProcessManager) {
	// Job management endpoints
	r.POST("/api/jobs", createJobHandler(pm))
	r.GET("/api/jobs", listJobsHandler(pm))
	r.GET("/api/jobs/:id", getJobHandler(pm))
	r.DELETE("/api/jobs/:id", removeJobHandler(pm))
	r.POST("/api/jobs/:id/stop", stopJobHandler(pm))
	r.POST("/api/jobs/:id/restart", restartJobHandler(pm))

	// Log streaming endpoint
	r.GET("/api/jobs/:id/logs", streamLogsHandler(pm))
}

func removeJobHandler(pm *core.ProcessManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := pm.RemoveJob(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to remove job: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Job removed successfully",
		})
	}
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
		// Upgrade to WebSocket connection
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		}
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
			return
		}
		defer ws.Close()

		id := c.Param("id")

		// Check if job exists
		job, err := pm.GetJob(id)
		if err != nil {
			ws.WriteJSON(gin.H{"error": "Failed to get job: " + err.Error()})
			return
		}
		if job == nil {
			ws.WriteJSON(gin.H{"error": "Job not found"})
			return
		}

		// Get historical logs
		logs, err := pm.Store.GetJobLogs(id)
		if err != nil {
			ws.WriteJSON(gin.H{"error": "Failed to get logs: " + err.Error()})
			return
		}

		// Send historical logs first
		for _, log := range logs {
			if err := ws.WriteJSON(gin.H{
				"text": log.RawText,
				"time": log.Time.Format(time.RFC3339),
			}); err != nil {
				return
			}
		}

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
			for {
				select {
				case msg := <-clientChan:
					if err := ws.WriteJSON(gin.H{
						"text": msg.RawText,
						"time": msg.Time.Format(time.RFC3339),
					}); err != nil {
						return
					}
				case <-c.Done():
					close(clientChan)
					return
				}
			}
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
