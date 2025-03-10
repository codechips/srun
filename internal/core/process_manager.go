package core

import (
	"bufio"
	"container/ring"
	"context"
	"fmt"
	"os/exec"
	"srun/internal/ansi"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type ProcessManager struct {
    Mu        sync.RWMutex
    Jobs      map[string]*Job
    Store     Storage
    LogChan   chan LogMessage
    logBuffer []LogMessage
    logMu     sync.Mutex
}

func (pm *ProcessManager) StartJob(command string) (*Job, error) {
    // Create job with unique ID
    job := &Job{
        ID:        uuid.New().String(),
        Command:   command,
        Status:    "running",
        StartedAt: time.Now(),
        LogBuffer: ring.New(1000),
    }

    // Create context without timeout - just use background context
    ctx, cancel := context.WithCancel(context.Background())
    job.Cancel = cancel

    // Prepare command
    cmd := exec.CommandContext(ctx, "sh", "-c", command)
    job.Cmd = cmd

    // Set up pipes for stdout and stderr
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
    }

    // Start the command
    if err := cmd.Start(); err != nil {
        // Update job status to failed since command couldn't even start
        job.Status = "failed"
        job.CompletedAt = time.Now()
        
        // Store the failed job
        pm.Mu.Lock()
        pm.Jobs[job.ID] = job
        pm.Mu.Unlock()
        
        if err := pm.Store.CreateJob(job); err != nil {
            return nil, fmt.Errorf("failed to create failed job: %w", err)
        }
        
        return job, nil
    }

    // Set the PID
    job.PID = cmd.Process.Pid

    // Store job in manager
    pm.Mu.Lock()
    pm.Jobs[job.ID] = job
    pm.Mu.Unlock()

    // Create job in storage
    if err := pm.Store.CreateJob(job); err != nil {
        // Cleanup if storage fails
        job.Cancel()
        delete(pm.Jobs, job.ID)
        return nil, fmt.Errorf("failed to create job: %w", err)
    }

    // Start a goroutine to check for immediate failure
    go func() {
        // Give the process a small window to fail
        time.Sleep(100 * time.Millisecond)
        
        // Check if process has already exited
        if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
            // Process has already exited, but we need to wait for the exit code
            // which will be handled in the completion monitoring goroutine
            // Don't mark as failed here
            return
        }
    }()

    // Handle command output in goroutines
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            processed := ansi.Process(scanner.Text())
            msg := LogMessage{
                JobID:     job.ID,
                Text:      processed.Plain,
                RawText:   processed.Raw,
                Styles:    processed.Styles,
                Progress:  processed.Progress,
                Time:      time.Now(),
            }

            // Store in ring buffer
            job.LogBuffer.Value = processed.Raw
            job.LogBuffer = job.LogBuffer.Next()

            // Add to batch buffer
            pm.logMu.Lock()
            pm.logBuffer = append(pm.logBuffer, msg)
            pm.logMu.Unlock()

            // Debug log
            fmt.Printf("Received log for job %s: %s\n", job.ID, processed.Raw)

            // Send to real-time channel
            pm.LogChan <- msg
        }
    }()

    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() {
            processed := ansi.Process(scanner.Text())
            msg := LogMessage{
                JobID:     job.ID,
                Text:      processed.Plain,
                RawText:   processed.Raw,
                Styles:    processed.Styles,
                Progress:  processed.Progress,
                Time:      time.Now(),
            }

            // Store in ring buffer
            job.LogBuffer.Value = processed.Raw
            job.LogBuffer = job.LogBuffer.Next()

            // Add to batch buffer
            pm.logMu.Lock()
            pm.logBuffer = append(pm.logBuffer, msg)
            pm.logMu.Unlock()

            // Send to real-time channel
            pm.LogChan <- msg
        }
    }()

    // Monitor command completion
    go func() {
        err := cmd.Wait()
        pm.Mu.Lock()
        job.CompletedAt = time.Now()
        if err != nil {
            if ctx.Err() == context.DeadlineExceeded {
                job.Status = "timeout"
            } else if ctx.Err() == context.Canceled {
                // Job was intentionally stopped, keep the "stopped" status
                if job.Status != "stopped" {
                    job.Status = "failed"
                }
            } else {
                // Check if the error contains an exit code
                if exitErr, ok := err.(*exec.ExitError); ok {
                    if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
                        fmt.Printf("Command failed with exit code: %d\n", status.ExitStatus())
                        job.Status = "failed"
                    }
                } else {
                    job.Status = "failed"
                }
            }
        } else {
            job.Status = "completed"
        }
        pm.Mu.Unlock()

        // Flush any remaining logs before updating status
        pm.flushLogs()

        // Update existing job record with final status
        if err := pm.Store.UpdateJobStatus(job.ID, job.Status); err != nil {
            fmt.Printf("Failed to update job status: %v\n", err)
        }
    }()

    return job, nil
}

func NewProcessManager(store Storage) *ProcessManager {
    pm := &ProcessManager{
        Jobs:      make(map[string]*Job),
        Store:     store,
        LogChan:   make(chan LogMessage, 100),
        logBuffer: make([]LogMessage, 0, 1000),
    }
    pm.startLogWriter()
    return pm
}

func (pm *ProcessManager) startLogWriter() {
    ticker := time.NewTicker(5 * time.Second)
    go func() {
        for range ticker.C {
            pm.flushLogs()
        }
    }()
}

func (pm *ProcessManager) flushLogs() {
    pm.logMu.Lock()
    if len(pm.logBuffer) == 0 {
        pm.logMu.Unlock()
        return
    }

    // Copy buffer and clear it
    logsToWrite := make([]LogMessage, len(pm.logBuffer))
    copy(logsToWrite, pm.logBuffer)
    pm.logBuffer = pm.logBuffer[:0]
    pm.logMu.Unlock()

    // Write to storage
    if err := pm.Store.BatchWriteLogs(logsToWrite); err != nil {
        fmt.Printf("Error writing logs: %v\n", err)
        // Could add retry logic here
    }
}

func (pm *ProcessManager) GetJob(id string) (*Job, error) {
    pm.Mu.RLock()
    defer pm.Mu.RUnlock()

    job, exists := pm.Jobs[id]
    if !exists {
        // Try loading from storage
        return pm.Store.GetJob(id)
    }
    return job, nil
}

func (pm *ProcessManager) ListJobs() ([]*Job, error) {
    pm.Mu.RLock()
    defer pm.Mu.RUnlock()

    return pm.Store.ListJobs()
}

func (pm *ProcessManager) StopJob(id string) error {
    pm.Mu.Lock()
    defer pm.Mu.Unlock()

    job, exists := pm.Jobs[id]
    if !exists {
        return fmt.Errorf("job not found: %s", id)
    }

    if job.Status != "running" {
        return fmt.Errorf("job is not running: %s", id)
    }

    // Cancel the context and wait for process to finish
    job.Cancel()
    job.Status = "stopped"

    // Create new job record with stopped status
    newJob := &Job{
        ID:        uuid.New().String(),
        Command:   job.Command,
        Status:    job.Status,
        StartedAt: job.StartedAt,
        LogBuffer: job.LogBuffer,
        Cmd:      job.Cmd,
    }
    if err := pm.Store.CreateJob(newJob); err != nil {
        return fmt.Errorf("failed to create job status: %w", err)
    }

    return nil
}

func (pm *ProcessManager) RestartJob(id string) (*Job, error) {
    pm.Mu.Lock()
    oldJob, exists := pm.Jobs[id]
    pm.Mu.Unlock()

    if !exists {
        return nil, fmt.Errorf("job not found: %s", id)
    }

    // Stop the old job if it's still running
    if oldJob.Status == "running" {
        if err := pm.StopJob(id); err != nil {
            return nil, fmt.Errorf("failed to stop old job: %w", err)
        }
    }

    // Get the original command from the old job's Cmd
    originalCmd := oldJob.Cmd.Args[2] // Skip "sh" and "-c"

    // Start a new job with the same command
    newJob, err := pm.StartJob(originalCmd)
    if err != nil {
        return nil, fmt.Errorf("failed to restart job: %w", err)
    }

    return newJob, nil
}

func (pm *ProcessManager) GetJobLogs(id string) []string {
    pm.Mu.RLock()
    defer pm.Mu.RUnlock()

    job, exists := pm.Jobs[id]
    if !exists {
        return nil
    }

    // Convert ring buffer to slice
    var logs []string
    if job.LogBuffer != nil {
        job.LogBuffer.Do(func(v interface{}) {
            if v != nil {
                logs = append(logs, v.(string))
            }
        })
    }
    return logs
}

func (pm *ProcessManager) RemoveJob(id string) error {
    pm.Mu.Lock()
    defer pm.Mu.Unlock()

    // Try to get the job from memory first
    job, exists := pm.Jobs[id]
    if exists && job.Status == "running" {
        // If job is running, cancel it
        job.Cancel()
        // Wait a moment for the job to clean up
        time.Sleep(100 * time.Millisecond)
    }

    // Remove from memory if it exists
    delete(pm.Jobs, id)

    // Remove from storage (this will cascade delete logs due to foreign key)
    if err := pm.Store.RemoveJob(id); err != nil {
        return fmt.Errorf("failed to remove job from storage: %w", err)
    }

    return nil
}

func (pm *ProcessManager) Cleanup() {
    pm.Mu.Lock()
    defer pm.Mu.Unlock()

    // Stop all running jobs
    for _, job := range pm.Jobs {
        if job.Status == "running" {
            job.Cancel()
        }
    }

    // Flush remaining logs
    pm.flushLogs()

    // Close log channel
    close(pm.LogChan)
}

type Job struct {
    ID          string
    Cmd         *exec.Cmd
    Command     string             // Store command string directly
    PID         int               // Process ID
    Cancel      context.CancelFunc
    Status      string // running, stopped, completed
    StartedAt   time.Time
    CompletedAt time.Time         // When the job finished (success or failure)
    LogBuffer   *ring.Ring // 1000 elements
}

type LogMessage struct {
    JobID     string
    Text      string                // Plain text without ANSI codes
    RawText   string                // Original text with ANSI codes
    Styles    map[int][]string      // Style information
    Progress  *ansi.ProgressInfo    // Progress information if detected
    Time      time.Time
}

type Storage interface {
    CreateJob(job *Job) error
    GetJob(id string) (*Job, error)
    ListJobs() ([]*Job, error)
    RemoveJob(id string) error
    BatchWriteLogs(logs []LogMessage) error
    GetJobLogs(id string) ([]LogMessage, error)
    UpdateJobStatus(id string, status string) error
}
