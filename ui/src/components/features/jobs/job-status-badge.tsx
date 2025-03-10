import { Badge } from "@/components/ui/badge";

type JobStatus = "completed" | "running" | "failed" | "stopped";

interface JobStatusBadgeProps {
  status: JobStatus;
}

export function JobStatusBadge({ status }: JobStatusBadgeProps) {
  return (
    <Badge
      className={
        status === "completed"
          ? "bg-green-500 hover:bg-green-600"
          : status === "running"
            ? "bg-yellow-500 hover:bg-yellow-600"
            : status === "failed"
              ? "bg-red-500 hover:bg-red-600"
              : "bg-secondary hover:bg-secondary/80"
      }
    >
      {status}
    </Badge>
  );
}
