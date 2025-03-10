package core

import "database/sql"

var migrations = []string{
    `CREATE TABLE IF NOT EXISTS jobs (
        id TEXT PRIMARY KEY,
        command TEXT NOT NULL,
        pid INTEGER,
        status TEXT CHECK(status IN ('running', 'stopped', 'completed', 'failed', 'timeout')) NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        stopped_at DATETIME,
        exit_code INTEGER
    )`,
    `CREATE TABLE IF NOT EXISTS job_logs (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        job_id TEXT NOT NULL,
        content TEXT NOT NULL,
        log_level TEXT CHECK(log_level IN ('stdout', 'stderr')) NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
    )`,
}

func migrate(db *sql.DB) error {
    for _, query := range migrations {
        if _, err := db.Exec(query); err != nil {
            return err
        }
    }
    return nil
}
