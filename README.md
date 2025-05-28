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

## Reverse Proxy Configuration

`srun` can be deployed behind a reverse proxy and served under a subpath (e.g., `https://yourdomain.com/srun/`). The application dynamically adapts its base path based on a header provided by the reverse proxy.

### How it Works

1.  **Backend Adaptation**: The Go backend inspects the `X-Forwarded-Prefix` HTTP header (or a similar header configured in your proxy) on incoming requests. This header should contain the base path under which `srun` is being served (e.g., `/srun`).
2.  **Frontend Injection**: The backend injects this base path into the `index.html` served to the client. Specifically, it sets a `<base href="...">` tag and a JavaScript global `window.APP_BASE_PATH`.
3.  **Frontend Usage**: The frontend UI (built with Vite using relative asset paths `./`) uses this injected base path to correctly construct URLs for API calls, WebSocket connections, and static assets.

This allows `srun` to operate correctly without needing a compile-time or startup-time base path configuration.

### Example: Caddy

Here's an example Caddyfile configuration to serve `srun` on `localhost:8000` under the `/srun/` subpath:

```caddyfile
yourdomain.com {
    # Route requests for /srun/ to the srun application
    route /srun/* {
        # Strip the /srun prefix before forwarding to the backend
        uri strip_prefix /srun

        # Set the X-Forwarded-Prefix header so srun knows its public base path
        # This is crucial for the backend to generate correct links and for the UI to work.
        header_up X-Forwarded-Prefix /srun

        # Proxy requests to your srun application.
        # Caddy v2 automatically handles WebSocket proxying for paths like /api/jobs/:id/logs.
        reverse_proxy localhost:8000 # Assuming srun is running on port 8000
    }

    # Other configurations for your domain...
    # For example, to serve a static site at the root:
    # root * /srv/my-static-site
    # file_server
}
```

**Key points for Caddy (or other proxies):**

*   **`uri strip_prefix`**: Ensures that `srun` receives requests at its root (e.g., `/api/jobs` instead of `/srun/api/jobs`).
*   **`header_up X-Forwarded-Prefix`**: Informs `srun` of its public-facing base path. If your proxy uses a different header to convey this information (e.g., `X-Base-Path`), you'll need to adjust the Go backend code in `cmd/srun/main.go` to read that specific header.
*   **WebSocket Proxying**: Caddy's `reverse_proxy` handles WebSocket upgrades automatically, which is necessary for real-time log streaming.

With this setup, you can access `srun` at `http://yourdomain.com/srun/`.

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
