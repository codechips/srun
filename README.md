# srun - modern remote script and command runner

The name *srun* stands for **s**cript **run**ner (but is also an obscene curse word in Russian).

[![Go Version](https://img.shields.io/badge/go-1.21%2B-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Script and command execution system with web interface featuring:

- ✅ Real-time ANSI-compatible terminal output
- 🗄️ SQLite-backed job persistence
- 🔒 Mutex-protected concurrent access
- 🔄 Ring buffer log storage (1000 entries)
- 🌐 WebSocket-based log streaming
- 📦 Single-binary deployment

## Installation

### Binary Installation

```bash
# Download latest release
curl -L https://github.com/codechips/srun/releases/latest/download/srun-linux-amd64 | tar xz
chmod +x srun
./srun
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

## CLI Flags
| Flag          | Default                              | Description                                                                 |
|---------------|--------------------------------------|-----------------------------------------------------------------------------|
| `-port`       | `8000`                               | HTTP server port (also via `SRUN_PORT` environment variable)                |
| `-db`         | Platform-specific config directory*  | SQLite database path (auto-created if missing)                              |

*Default database locations:  
- **Linux**: `$HOME/.config/srun/srun.db`  
- **macOS**: `$HOME/Library/Application Support/srun/srun.db`  
- **Windows**: `%APPDATA%\srun\srun.db`  

## Process Management (from process_manager.go)
- Job lifecycle management with PID tracking
- Automatic process cleanup on termination
- Context-based cancellation for graceful shutdown
- Status transitions: running → {completed, stopped, failed}
- Restart functionality with command preservation

## Development

To develop locally start the UI server:

```bash
cd ui
pnpm install
pnpm dev
```

Then start the Go server:

```bash
go run cmd/srun/main.go
```

In development the calls to the API server are proxied through the Vite dev server. See [vite.config.ts](ui/vite.config.ts).

## Contributing
1. Fork repository
2. Create feature branch (`feat/...` or `fix/...`)
3. Submit PR with

## License
MIT Licensed - See [LICENSE](LICENSE) for details.
