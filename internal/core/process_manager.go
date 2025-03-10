package core

import (
    "bufio"
    "context"
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
