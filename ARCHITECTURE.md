# Script Execution Service Architecture

## Core Components

1. **Web API Layer (Gin Framework)**
   - REST endpoints with JSON responses
   - WebSocket/SSE for real-time logs
   - Rate limiting middleware
   - Request validation middleware

2. **Script Execution Engine**
   ```go
   type ScriptProcess struct {
       ID        uuid.UUID
       Cmd       *exec.Cmd
       Status    string // running, stopped, completed
       StartTime time.Time
       LogBuffer *circular.Buffer // 10k line capacity
       Cancel    context.CancelFunc
   }
   
   type ScriptManager struct {
       Processes sync.Map
       LogDir    string
       DB        *sqlx.DB
   }
   ```

3. **Logging System**
   - Dual storage: in-memory ring buffer + persistent files
   - ANSI escape code preservation
   - Log rotation strategy:
     ```bash
     logs/
       2023-09-01/
         script_abc123.log
         script_def456.log
     ```

4. **Storage Layer (SQLite)**
   ```sql
   CREATE TABLE scripts (
     id TEXT PRIMARY KEY,
     command TEXT,
     status TEXT,
     created_at DATETIME,
     stopped_at DATETIME,
     exit_code INTEGER
   );

   CREATE TABLE log_chunks (
     script_id TEXT,
     timestamp DATETIME,
     chunk TEXT,
     is_error BOOLEAN
   );
   ```

## API Endpoints

```go
// Start new script
POST /api/scripts
Body: { "command": "python data_processor.py", "timeout": 3600 }

// List scripts
GET /api/scripts

// Stream logs
GET /api/scripts/{id}/logs
(WebSocket or SSE)

// Control scripts
POST /api/scripts/{id}/kill
POST /api/scripts/{id}/restart
```

## Concurrency Model

1. Process isolation with exec.CommandContext
2. Goroutine-per-script with cleanup hooks:
   ```go
   func (sm *ScriptManager) Start(command string) error {
       ctx, cancel := context.WithTimeout(context.Background(), 8*time.Hour)
       cmd := exec.CommandContext(ctx, "bash", "-c", command)
       
       // Set process output to teeing:
       // - Memory buffer
       // - Persistent file
       // - Active WebSocket connections
   }
   ```

## WebUI Considerations

1. ANSI rendering using `xterm.js`
2. Auto-refresh status table
3. WebSocket-based log display component
4. Streaming response processing

## Security Measures

1. Command injection protection:
   ```go
   func sanitizeInput(cmd string) bool {
       return !strings.ContainsAny(cmd, "&|;") 
   }
   ```

2. Process resource limits:
   ```go
   cmd.SysProcAttr = &syscall.SysProcAttr{
       Setpgid: true,
       Cpulimit: 0.8, // Max 80% CPU
   }
   ```

## Deployment Strategy

Single binary with embedded assets:
```bash
./script-service \
  -port=8080 \
  -log-dir=/var/log/script-service \
  -db-file=/var/lib/script-service/data.db
```

## First Implementation Steps

1. Scaffold Gin routes
2. Implement ScriptProcess lifecycle management
3. Create log tee implementation (memory + file)
4. Add SQLite persistence layer
