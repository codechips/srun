import { useState } from "react";
import { TableCell, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { MoreVertical, Play, Square, Trash } from "lucide-react";
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
  onStop: (id: string) => void;
  onRestart: (id: string) => void;
  onRemove: (id: string) => void;
}

export function JobRow({ job, onStop, onRestart, onRemove }: JobRowProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <>
      <TableRow
        className="cursor-pointer hover:bg-muted/50"
        onClick={(e) => {
          if ((e.target as HTMLElement).closest('[role="menuitem"]')) {
            return;
          }
          setIsExpanded(!isExpanded);
        }}
      >
        <TableCell className="font-mono">{job.id.slice(0, 8)}</TableCell>
        <TableCell className="font-mono">{job.pid}</TableCell>
        <TableCell>
          <Badge
            className={
              job.status === "completed"
                ? "bg-green-500 hover:bg-green-600"
                : job.status === "running"
                  ? "bg-yellow-500 hover:bg-yellow-600"
                  : job.status === "failed"
                    ? "bg-red-500 hover:bg-red-600"
                    : "bg-secondary hover:bg-secondary/80"
            }
          >
            {job.status}
          </Badge>
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
                <DropdownMenuItem onClick={() => onRestart(job.id)}>
                  <Play className="mr-2 h-4 w-4" />
                  <span>Restart</span>
                </DropdownMenuItem>
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
      {isExpanded && (
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
