package core

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
    db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    
    return &SQLiteStorage{db: db}, nil
}
