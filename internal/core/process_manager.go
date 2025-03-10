package core

import (
	"bufio"
	"container/ring"
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
	"github.com/google/uuid"
	"srun/internal/ansi"
)

type ProcessManager struct {
    Mu       sync.RWMutex
    Jobs     map[string]*Job
    Store    Storage
    LogChan  chan LogMessage
}

func (pm *ProcessManager) StartJob(command string, timeout time.Duration) (*Job, error) {
    // Validate timeout range (5m to 8h)
    if timeout < 5*time.Minute || timeout > 8*time.Hour {
        return nil, fmt.Errorf("timeout must be between 5 minutes and 8 hours")
    }

    // Create job with unique ID
    job := &Job{
        ID:        uuid.New().String(),
        Status:    "running",
        StartedAt: time.Now(),
        LogBuffer: ring.New(1000),
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
        return nil, fmt.Errorf("failed to start command: %w", err)
    }

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

    // Handle command output in goroutines
    go func() {
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            processed := ansi.Process(scanner.Text())
            pm.LogChan <- LogMessage{
                JobID:     job.ID,
                Text:      processed.Plain,
                RawText:   processed.Raw,
                Styles:    processed.Styles,
                Progress:  processed.Progress,
                Time:      time.Now(),
            }
            job.LogBuffer.Value = processed.Raw  // Store raw version with ANSI codes
            job.LogBuffer = job.LogBuffer.Next()
        }
    }()

    go func() {
        scanner := bufio.NewScanner(stderr)
        for scanner.Scan() {
            processed := ansi.Process(scanner.Text())
            pm.LogChan <- LogMessage{
                JobID:     job.ID,
                Text:      processed.Plain,
                RawText:   processed.Raw,
                Styles:    processed.Styles,
                Progress:  processed.Progress,
                Time:      time.Now(),
            }
            job.LogBuffer.Value = processed.Raw  // Store raw version with ANSI codes
            job.LogBuffer = job.LogBuffer.Next()
        }
    }()

    // Monitor command completion
    go func() {
        err := cmd.Wait()
        pm.Mu.Lock()
        if err != nil {
            if ctx.Err() == context.DeadlineExceeded {
                job.Status = "timeout"
            } else {
                job.Status = "error"
            }
        } else {
            job.Status = "completed"
        }
        pm.Mu.Unlock()
        
        // Create new job record with final status
        newJob := &Job{
            ID:        uuid.New().String(),
            Status:    job.Status,
            StartedAt: job.StartedAt,
            LogBuffer: job.LogBuffer,
            Cmd:      job.Cmd,
        }
        _ = pm.Store.CreateJob(newJob)
    }()

    return job, nil
}

func NewProcessManager(store Storage) *ProcessManager {
    return &ProcessManager{
        Jobs:    make(map[string]*Job),
        Store:   store,
        LogChan: make(chan LogMessage, 100),
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

    // Start a new job with the same command and a 1-hour default timeout
    newJob, err := pm.StartJob(originalCmd, 1*time.Hour)
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

    job, exists := pm.Jobs[id]
    if !exists {
        return fmt.Errorf("job not found: %s", id)
    }

    // Stop the job if it's running
    if job.Status == "running" {
        job.Cancel()
    }

    // Remove from memory
    delete(pm.Jobs, id)

    // Remove from storage
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

    // Close log channel
    close(pm.LogChan)
}

type Job struct {
    ID        string
    Cmd       *exec.Cmd
    Cancel    context.CancelFunc
    Status    string // running, stopped, completed
    StartedAt time.Time
    LogBuffer *ring.Ring // 1000 elements
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
}
