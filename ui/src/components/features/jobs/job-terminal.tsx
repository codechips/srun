import { useEffect, useRef } from 'react'
import { Terminal } from '@xterm/xterm'
import '@xterm/xterm/css/xterm.css'

interface JobTerminalProps {
  jobId: string
}

export function JobTerminal({ jobId }: JobTerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const terminal = useRef<Terminal>()

  useEffect(() => {
    if (!terminalRef.current) return

    // Initialize terminal
    terminal.current = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'monospace',
      theme: {
        background: '#1a1b1e'
      }
    })
    terminal.current.open(terminalRef.current)

    // Connect to WebSocket for logs
    const ws = new WebSocket(`ws://${window.location.host}/api/jobs/${jobId}/logs`)
    
    ws.onopen = () => {
      terminal.current?.writeln('Connected to log stream...\r\n')
    }

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      if (data.error) {
        terminal.current?.writeln(`\r\nError: ${data.error}`)
        return
      }
      terminal.current?.writeln(data.text)
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      terminal.current?.writeln('\r\nError: WebSocket connection failed')
    }

    ws.onclose = () => {
      terminal.current?.writeln('\r\nLog stream disconnected')
    }

    return () => {
      ws.close()
      terminal.current?.dispose()
    }
  }, [jobId])

  return <div ref={terminalRef} className="h-[500px] bg-[#1a1b1e] rounded-md" />
}
