import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

type JobStatus = "completed" | "running" | "failed" | "stopped";

interface JobStatusBadgeProps {
  status: JobStatus;
}

const statusStyles = {
  completed: "bg-green-500/15 text-green-700 hover:bg-green-500/25",
  running: "bg-yellow-500/15 text-yellow-700 hover:bg-yellow-500/25",
  failed: "bg-red-500/15 text-red-700 hover:bg-red-500/25",
  stopped: "bg-muted text-muted-foreground hover:bg-muted/80"
} as const;

export function JobStatusBadge({ status }: JobStatusBadgeProps) {
  return (
    <Badge 
      variant="secondary"
      className={cn(
        "font-medium",
        statusStyles[status]
      )}
    >
      {status}
    </Badge>
  );
}
