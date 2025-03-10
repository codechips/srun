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

func (s *SQLiteStorage) SaveJob(job *Job) error {
    // First try to update existing job
    result, err := s.db.Exec(
        `UPDATE jobs 
         SET status = ?, 
             command = ?,
             pid = ?
         WHERE id = ?`,
        job.Status,
        job.Cmd.Args[2], // Skip "sh" "-c" to get actual command
        job.Cmd.Process.Pid,
        job.ID,
    )
    if err != nil {
        return fmt.Errorf("failed to update job: %w", err)
    }

    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }

    // If no rows were updated, insert new job
    if rows == 0 {
        _, err = s.db.Exec(
            `INSERT INTO jobs (id, command, pid, status, created_at) 
             VALUES (?, ?, ?, ?, ?)`,
            job.ID,
            job.Cmd.Args[2],
            job.Cmd.Process.Pid,
            job.Status,
            job.StartedAt,
        )
        if err != nil {
            return fmt.Errorf("failed to insert job: %w", err)
        }
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
