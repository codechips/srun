package core

import (
    "bufio"
    "context"
    "fmt"
    "os/exec"
    "sync"
    "time"
    "container/ring"
)

type ProcessManager struct {
    Mu       sync.RWMutex
    Jobs     map[string]*Job
    Store    Storage
    LogChan  chan LogMessage
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
    
    // Update storage
    if err := pm.Store.SaveJob(job); err != nil {
        return fmt.Errorf("failed to save job status: %w", err)
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
    JobID string
    Text  string
    Time  time.Time
}

type Storage interface {
    SaveJob(job *Job) error
    GetJob(id string) (*Job, error)
    ListJobs() ([]*Job, error)
}
