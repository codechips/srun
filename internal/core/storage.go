package core

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
    db *sql.DB
}

func (s *SQLiteStorage) SaveJob(job *Job) error {
    // TODO: Implement database persistence
    return nil
}

func (s *SQLiteStorage) GetJob(id string) (*Job, error) {
    // TODO: Implement database lookup
    return nil, nil
}

func (s *SQLiteStorage) ListJobs() ([]*Job, error) {
    // TODO: Implement database query
    return nil, nil
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
