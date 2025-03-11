import { useEffect, useRef } from "react";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";

interface JobTerminalProps {
  jobId: string;
}

export function JobTerminal({ jobId }: JobTerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminal = useRef<Terminal>();

  useEffect(() => {
    if (!terminalRef.current) return;

    // Initialize terminal
    terminal.current = new Terminal({
      cursorBlink: false,
      fontSize: 14,
      fontFamily: "monospace",
      convertEol: true,
      theme: {
        background: "#1a1b1e",
        foreground: "#e4e4e7",
        cursor: "#a1a1aa",
      },
    });
    terminal.current.open(terminalRef.current);

    // In development, connect through Vite's proxy
    const wsUrl = `ws://${window.location.host}/api/jobs/${jobId}/logs`;
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.error) {
          terminal.current?.writeln(`\r\nError: ${data.error}`);
          return;
        }

        terminal.current?.writeln(data.text);

        // // Handle carriage returns for progress updates
        // if (data.text.includes("\r") && !data.text.includes("\n")) {
        //   terminal.current?.write("\r" + data.text);
        // } else {
        //   terminal.current?.writeln(data.text);
        // }
      } catch (error) {
        console.error("Failed to parse message:", error, event.data);
        terminal.current?.writeln(
          `\r\nError: Failed to parse message: ${event.data}`,
        );
      }
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
    };

    ws.onclose = () => {
      console.log("WebSocket closed");
    };

    return () => {
      ws.close();
      terminal.current?.dispose();
    };
  }, [jobId]);

  return (
    <div ref={terminalRef} className="h-[500px] bg-[#1a1b1e] rounded-md p-4" />
  );
}
