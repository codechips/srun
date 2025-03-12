import { TableCell, TableRow } from "@/components/ui/table";
import { JobStatusBadge } from "./job-status-badge";
import { Button } from "@/components/ui/button";
import { MoreVertical, Play, Square, Trash, Pencil } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { JobTerminal } from "./job-terminal";
import { Job } from "@/hooks/use-jobs";

interface JobRowProps {
  job: Job;
  expanded: boolean;
  onExpand: (id: string | null) => void;
  onStop: (id: string) => void;
  onRestart: (id: string) => void;
  onRemove: (id: string) => void;
  onEdit: (command: string) => void;
}

export function JobRow({
  job,
  expanded,
  onExpand,
  onStop,
  onRestart,
  onRemove,
}: JobRowProps) {
  return (
    <>
      <TableRow
        className="cursor-pointer hover:bg-muted/50"
        onClick={(e) => {
          if ((e.target as HTMLElement).closest('[role="menuitem"]')) {
            return;
          }
          onExpand(expanded ? null : job.id);
        }}
      >
        <TableCell className="font-mono">{job.id.slice(0, 8)}</TableCell>
        <TableCell className="font-mono">{job.pid}</TableCell>
        <TableCell>
          <JobStatusBadge
            status={
              job.status as "completed" | "running" | "failed" | "stopped"
            }
          />
        </TableCell>
        <TableCell className="font-mono">{job.command}</TableCell>
        <TableCell>{new Date(job.startedAt).toISOString()}</TableCell>
        <TableCell>
          {job.status === "running"
            ? ""
            : job.completedAt
              ? new Date(job.completedAt).toISOString()
              : "-"}
        </TableCell>
        <TableCell className="text-right">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="h-8 w-8 p-0">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {job.status === "running" ? (
                <DropdownMenuItem onClick={() => onStop(job.id)}>
                  <Square className="mr-2 h-4 w-4" />
                  <span>Stop</span>
                </DropdownMenuItem>
              ) : (
                <>
                  <DropdownMenuItem onClick={() => onEdit(job.command)}>
                    <Pencil className="mr-2 h-4 w-4" />
                    <span>Edit & Run</span>
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => onRestart(job.id)}>
                    <Play className="mr-2 h-4 w-4" />
                    <span>
                      {job.status === "failed" ? "Try Again" : "Restart"}
                    </span>
                  </DropdownMenuItem>
                </>
              )}
              <DropdownMenuItem
                onClick={() => onRemove(job.id)}
                className="text-red-600"
              >
                <Trash className="mr-2 h-4 w-4" />
                <span>Remove</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
      {expanded && (
        <TableRow>
          <TableCell colSpan={7} className="p-0 border-0">
            <div className="bg-muted/50">
              <JobTerminal jobId={job.id} />
            </div>
          </TableCell>
        </TableRow>
      )}
    </>
  );
}
