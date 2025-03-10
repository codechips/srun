package core

import (
    "context"
    "container/ring"
    "database/sql"
    "fmt"
    "os/exec"
    "time"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
    db *sql.DB
}

func (s *SQLiteStorage) CreateJob(job *Job) error {
    _, err := s.db.Exec(
        `INSERT INTO jobs (id, command, pid, status, created_at) 
         VALUES (?, ?, ?, ?, ?)`,
        job.ID,
        job.Cmd.Args[2], // Skip "sh" "-c" to get actual command
        job.Cmd.Process.Pid,
        job.Status,
        job.StartedAt,
    )
    if err != nil {
        return fmt.Errorf("failed to create job: %w", err)
    }
    return nil
}

func (s *SQLiteStorage) GetJob(id string) (*Job, error) {
    row := s.db.QueryRow(
        `SELECT id, command, status, created_at 
         FROM jobs 
         WHERE id = ?`,
        id,
    )

    var (
        jobID      string
        command    string
        status     string
        createdAt  time.Time
    )

    if err := row.Scan(&jobID, &command, &status, &createdAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to scan job: %w", err)
    }

    // Create a new Job struct with the retrieved data
    job := &Job{
        ID:        jobID,
        Status:    status,
        StartedAt: createdAt,
        LogBuffer: ring.New(1000),
    }

    // Only create Cmd if job is not completed/stopped
    if status == "running" {
        ctx, cancel := context.WithCancel(context.Background())
        job.Cancel = cancel
        job.Cmd = exec.CommandContext(ctx, "sh", "-c", command)
    }

    return job, nil
}

func (s *SQLiteStorage) RemoveJob(id string) error {
    _, err := s.db.Exec("DELETE FROM jobs WHERE id = ?", id)
    if err != nil {
        return fmt.Errorf("failed to remove job: %w", err)
    }
    return nil
}

func (s *SQLiteStorage) ListJobs() ([]*Job, error) {
    rows, err := s.db.Query(
        `SELECT id, command, status, created_at 
         FROM jobs 
         ORDER BY created_at DESC`,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query jobs: %w", err)
    }
    defer rows.Close()

    var jobs []*Job
    for rows.Next() {
        var (
            jobID      string
            command    string
            status     string
            createdAt  time.Time
        )

        if err := rows.Scan(&jobID, &command, &status, &createdAt); err != nil {
            return nil, fmt.Errorf("failed to scan job row: %w", err)
        }

        job := &Job{
            ID:        jobID,
            Status:    status,
            StartedAt: createdAt,
            LogBuffer: ring.New(1000),
        }

        // Only create Cmd if job is not completed/stopped
        if status == "running" {
            ctx, cancel := context.WithCancel(context.Background())
            job.Cancel = cancel
            job.Cmd = exec.CommandContext(ctx, "sh", "-c", command)
        }

        jobs = append(jobs, job)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating job rows: %w", err)
    }

    return jobs, nil
}

func (s *SQLiteStorage) BatchWriteLogs(logs []LogMessage) error {
    tx, err := s.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT INTO job_logs (job_id, content, log_level, created_at)
        VALUES (?, ?, ?, ?)
    `)
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()

    for _, log := range logs {
        _, err = stmt.Exec(
            log.JobID,
            log.RawText,
            "stdout", // TODO: Add proper log level detection
            log.Time,
        )
        if err != nil {
            return fmt.Errorf("failed to insert log: %w", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    return nil
}

func (s *SQLiteStorage) GetJobLogs(jobID string) ([]LogMessage, error) {
    rows, err := s.db.Query(`
        SELECT content, created_at 
        FROM job_logs 
        WHERE job_id = ? 
        ORDER BY created_at ASC`,
        jobID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to query logs: %w", err)
    }
    defer rows.Close()

    var logs []LogMessage
    for rows.Next() {
        var content string
        var createdAt time.Time
        
        if err := rows.Scan(&content, &createdAt); err != nil {
            return nil, fmt.Errorf("failed to scan log row: %w", err)
        }

        processed := ansi.Process(content)
        logs = append(logs, LogMessage{
            JobID:    jobID,
            Text:     processed.Plain,
            RawText:  processed.Raw,
            Styles:   processed.Styles,
            Progress: processed.Progress,
            Time:     createdAt,
        })
    }

    return logs, nil
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    if err := migrate(db); err != nil {
        return nil, fmt.Errorf("migration failed: %w", err)
    }

    return &SQLiteStorage{db: db}, nil
}
