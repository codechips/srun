package api

import (
	"net/http"
	"srun/internal/core"
	"srun/internal/version"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type CreateJobRequest struct {
	Command string `json:"command" binding:"required"`
}

func SetupRoutes(r *gin.Engine, pm *core.ProcessManager) {
	// Version endpoint
	r.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version.GetInfo())
	})

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

func sendBatch(ws *websocket.Conn, batch []core.LogMessage) {
	var combined strings.Builder
	for _, msg := range batch {
		combined.WriteString(msg.RawText)
	}
	ws.WriteMessage(websocket.TextMessage, []byte(combined.String()))
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

		// Initialize empty response slice
		response := make([]gin.H, 0)

		// Only process jobs if not nil
		if jobs != nil {
			// Convert jobs to response format
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

		// Check if job exists first
		job, err := pm.GetJob(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job: " + err.Error()})
			return
		}
		if job == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Job not found"})
			return
		}

		// Upgrade to WebSocket connection
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
			EnableCompression: true,
		}

		// Log headers for debugging
		for k, v := range c.Request.Header {
			gin.DefaultWriter.Write([]byte("Header: " + k + ": " + v[0] + "\n"))
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			gin.DefaultErrorWriter.Write([]byte("WebSocket upgrade failed: " + err.Error() + "\n"))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection: " + err.Error()})
			return
		}
		defer ws.Close()

		// Set WebSocket read deadline to prevent hanging connections
		ws.SetReadDeadline(time.Now().Add(24 * time.Hour))

		// Get historical logs first
		logs, err := pm.Store.GetJobLogs(id)
		if err != nil {
			ws.WriteJSON(gin.H{"error": "Failed to get logs: " + err.Error()})
			return
		}

		// Send historical logs
		for _, log := range logs {
			if err := ws.WriteMessage(websocket.TextMessage, []byte(log.RawText)); err != nil {
				return
			}
		}

		// If job is running, subscribe to real-time logs
		if job.Status == "running" {
			// Create buffered channel for this client
			clientChan := make(chan core.LogMessage, 1000)
			defer close(clientChan)

			// Start goroutine to forward messages from LogChan to client
			go func() {
				batch := make([]core.LogMessage, 0, 10)
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case msg := <-pm.LogChan:
						if msg.JobID == id {
							batch = append(batch, msg)
							if len(batch) >= 10 {
								sendBatch(ws, batch)
								batch = batch[:0]
							}
						}
					case <-ticker.C:
						if len(batch) > 0 {
							sendBatch(ws, batch)
							batch = batch[:0]
						}
					case <-c.Done():
						return
					}
				}
			}()

			// Wait for context done
			<-c.Done()
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
