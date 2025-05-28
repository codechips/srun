import { useEffect, useRef } from "react";
import { Terminal } from "@xterm/xterm";
import { getWsUrl } from "@/config";
import "@xterm/xterm/css/xterm.css";

interface JobTerminalProps {
  jobId: string;
}

export function JobTerminal({ jobId }: JobTerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const terminal = useRef<Terminal | null>(null);

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

    const wsUrl = getWsUrl(`/api/jobs/${jobId}/logs`);
    const ws = new WebSocket(wsUrl);

    ws.onmessage = (event) => {
      try {
        const { data } = event;
        // Handle carriage returns for progress updates
        if (data.includes("\r") && !data.includes("\n")) {
          terminal.current?.write("\r" + data);
        } else {
          terminal.current?.writeln(data);
        }
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
