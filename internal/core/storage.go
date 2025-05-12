package core

import (
	"container/ring"
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"srun/internal/ansi"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db *sql.DB
}

func (s *SQLiteStorage) CreateJob(job *Job) error {
	fmt.Printf("Creating job in database: %s with status %s\n", job.ID, job.Status)
	var stoppedAt interface{}
	if !job.CompletedAt.IsZero() {
		stoppedAt = job.CompletedAt
	}

	_, err := s.db.Exec(
		`INSERT INTO jobs (id, command, pid, status, created_at, stopped_at) 
         VALUES (?, ?, ?, ?, ?, ?)`,
		job.ID,
		job.Command,
		job.PID,
		job.Status,
		job.StartedAt,
		stoppedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	fmt.Printf("Successfully created job in database: %s\n", job.ID)
	return nil
}

func (s *SQLiteStorage) GetJob(id string) (*Job, error) {
	row := s.db.QueryRow(
		`SELECT id, command, pid, status, created_at, stopped_at 
         FROM jobs 
         WHERE id = ?`,
		id,
	)

	var (
		jobID     string
		command   string
		pid       int
		status    string
		createdAt time.Time
		stoppedAt sql.NullTime
	)

	if err := row.Scan(&jobID, &command, &pid, &status, &createdAt, &stoppedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan job: %w", err)
	}

	// Create a new Job struct with the retrieved data
	job := &Job{
		ID:          jobID,
		Command:     command,
		PID:         pid,
		Status:      status,
		StartedAt:   createdAt,
		CompletedAt: stoppedAt.Time,
		LogBuffer:   ring.New(1000),
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
	result, err := s.db.Exec("DELETE FROM jobs WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to remove job: %w", err)
	}

	// Check if any row was actually deleted
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("job not found: %s", id)
	}

	return nil
}

func (s *SQLiteStorage) ListJobs() ([]*Job, error) {
	rows, err := s.db.Query(
		`SELECT id, command, pid, status, created_at, stopped_at 
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
			jobID     string
			command   string
			pid       int
			status    string
			createdAt time.Time
			stoppedAt sql.NullTime
		)

		if err := rows.Scan(&jobID, &command, &pid, &status, &createdAt, &stoppedAt); err != nil {
			return nil, fmt.Errorf("failed to scan job row: %w", err)
		}

		job := &Job{
			ID:          jobID,
			Command:     command,
			PID:         pid,
			Status:      status,
			StartedAt:   createdAt,
			CompletedAt: stoppedAt.Time,
			LogBuffer:   ring.New(1000),
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
	fmt.Printf("Starting batch write of %d logs\n", len(logs))

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

	for i, log := range logs {
		fmt.Printf("Writing log %d/%d for job %s\n", i+1, len(logs), log.JobID)
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
	fmt.Printf("Successfully wrote %d logs to database\n", len(logs))
	return nil
}

func (s *SQLiteStorage) UpdateJobStatus(id string, status string) error {
	_, err := s.db.Exec(
		`UPDATE jobs 
         SET status = ?, 
             stopped_at = CURRENT_TIMESTAMP
         WHERE id = ?`,
		status,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
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
			JobID:   jobID,
			Text:    processed.Plain,
			RawText: processed.Raw,
			Time:    createdAt,
		})
	}

	return logs, nil
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	fmt.Printf("Opening SQLite database at: %s\n", dbPath)

	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Set proper permissions
	if err := os.Chmod(dbPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set database permissions: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}
