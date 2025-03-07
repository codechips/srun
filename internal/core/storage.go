package core

import (
    "database/sql"
    "fmt"
    
    _ "github.com/mattn/go-sqlite3"
)

type SRunStorage struct {
    db *sql.DB
}

func NewSRunStorage(dbPath string) (*SRunStorage, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    if err := migrate(db); err != nil {
        return nil, fmt.Errorf("migration failed: %w", err)
    }

    return &SRunStorage{db: db}, nil
}
