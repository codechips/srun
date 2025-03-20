# srun - Modern Remote Script & Command Runner

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Enterprise-grade command execution system with web interface, featuring:

✅ Real-time ANSI-compatible terminal output  
🗄️ SQLite-backed job persistence  
🔒 Mutex-protected concurrent access  
🔄 Ring buffer log storage (1000 entries)  
🌐 WebSocket-based log streaming  
📦 Single-binary deployment

## Installation

### Binary Installation
```bash
# Download latest release
curl -L https://github.com/codechips/srun/releases/latest/download/srun-linux-amd64 | tar xz
chmod +x srun
sudo mv srun /usr/local/bin/
```

### From Source
```bash
git clone https://github.com/codechips/srun.git
cd srun
./build.sh  # Requires Go 1.21+, Node 18+, pnpm
```

## Usage
```bash
# Start server (default port 8000)
./srun -port=8080

# Environment alternative
SRUN_PORT=8080 ./srun
```

Access the web UI at `http://localhost:8080`

## Process Management (from process_manager.go)
- Job lifecycle management with PID tracking
- Automatic process cleanup on termination
- Context-based cancellation for graceful shutdown
- Status transitions: running → {completed, stopped, failed}
- Restart functionality with command preservation

## API Endpoints
| Endpoint                | Method | Description                          |
|-------------------------|--------|--------------------------------------|
| `/api/version`          | GET    | Get build version info               |
| `/api/jobs`             | POST   | Create new job (JSON: `{"command":...}`) |
| `/api/jobs/:id/logs`    | GET    | Stream logs via WebSocket            |
| `/api/jobs/:id/restart` | POST   | Restart stopped job                  |
| `/api/jobs/:id/stop`    | POST   | Stop running job                     |

## Storage System
- SQLite database (`srun.db`) with:
  ```sql
  CREATE TABLE jobs (
    id TEXT PRIMARY KEY,
    command TEXT,
    pid INTEGER,
    status TEXT,
    created_at DATETIME,
    stopped_at DATETIME
  );
  CREATE TABLE job_logs (
    job_id TEXT REFERENCES jobs(id),
    content TEXT,
    created_at DATETIME
  );
  ```
- Batch log writes (10 messages/chunk)
- ANSI-aware log processing and storage
- Automatic schema migrations

## Development

### Build System
```bash
./build.sh  # Builds:
            # 1. UI with pnpm
            # 2. Embeds assets in Go binary
            # 3. Compiles with git version info
```

### Key Implementation Details
- **Concurrency**: `sync.RWMutex` for job map access
- **Log Handling**:
  - 50ms flush interval for log batches
  - 1000-entry ring buffer for in-memory storage
  - WebSocket message batching (10ms coalescing)
- **Process Supervision**:
  - Automatic failure detection (100ms health check)
  - PID validation with `syscall.Signal(0)`
  - Exit code parsing for failure states

## Contributing
1. Fork repository
2. Create feature branch (`feat/...` or `fix/...`)
3. Maintain:
   - SQL schema compatibility
   - ANSI escape handling
   - Mutex locking conventions
4. Submit PR with:
   - Code changes
   - Migration SQL if needed
   - Updated UI components
   - Concurrency safety analysis

## License
MIT Licensed - See [LICENSE](LICENSE) for details.
