import { useState } from "react";
import React from "react";
import { useJobs, useJobActions } from "@/hooks/use-jobs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { MoreVertical, Play, Square, Trash } from "lucide-react";
import { Button } from "@/components/ui/button";
import { JobTerminal } from "./job-terminal";

function LoadingTable() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">ID</TableHead>
          <TableHead className="w-[80px]">PID</TableHead>
          <TableHead className="w-[100px]">Status</TableHead>
          <TableHead>Command</TableHead>
          <TableHead className="w-[180px]">Started</TableHead>
          <TableHead className="w-[180px]">Completed</TableHead>
          <TableHead className="w-[100px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {[...Array(5)].map((_, i) => (
          <TableRow key={i}>
            <TableCell>
              <Skeleton className="h-4 w-20" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-10" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-6 w-16" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-full" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-4 w-32" />
            </TableCell>
            <TableCell>
              <Skeleton className="h-8 w-8 rounded-full ml-auto" />
            </TableCell>
          </TableRow>
            {expandedJobId === job.id && (
              <TableRow>
                <TableCell colSpan={7} className="p-0">
                  <div className="p-4">
                    <JobTerminal jobId={job.id} />
                  </div>
                </TableCell>
              </TableRow>
            )}
          </React.Fragment>
        ))}
      </TableBody>
    </Table>
  );
}

export function JobList() {
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null);
  const { data: jobs, isLoading } = useJobs();
  const { stopJob, restartJob, removeJob } = useJobActions();

  const handleRowClick = (jobId: string) => {
    setExpandedJobId(expandedJobId === jobId ? null : jobId);
  };

  const handleStop = (id: string) => stopJob(id);
  const handleRestart = (id: string) => restartJob(id);
  const handleRemove = (id: string) => removeJob(id);

  if (isLoading) {
    return <LoadingTable />;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-[100px]">ID</TableHead>
          <TableHead className="w-[80px]">PID</TableHead>
          <TableHead className="w-[100px]">Status</TableHead>
          <TableHead>Command</TableHead>
          <TableHead className="w-[180px]">Started</TableHead>
          <TableHead className="w-[180px]">Completed</TableHead>
          <TableHead className="w-[100px] text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {jobs?.map((job) => (
          <React.Fragment key={job.id}>
            <TableRow 
              key={job.id} 
              className="cursor-pointer hover:bg-muted/50"
              onClick={(e) => {
                // Prevent row click when clicking dropdown menu
                if ((e.target as HTMLElement).closest('.dropdown-trigger')) {
                  return;
                }
                handleRowClick(job.id);
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
                  <Button variant="ghost" className="h-8 w-8 p-0 dropdown-trigger">
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {job.status === "running" ? (
                    <DropdownMenuItem onClick={() => handleStop(job.id)}>
                      <Square className="mr-2 h-4 w-4" />
                      <span>Stop</span>
                    </DropdownMenuItem>
                  ) : (
                    <DropdownMenuItem onClick={() => handleRestart(job.id)}>
                      <Play className="mr-2 h-4 w-4" />
                      <span>Restart</span>
                    </DropdownMenuItem>
                  )}
                  <DropdownMenuItem
                    onClick={() => handleRemove(job.id)}
                    className="text-red-600"
                  >
                    <Trash className="mr-2 h-4 w-4" />
                    <span>Remove</span>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
