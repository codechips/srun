import { useJobs, useJobActions } from "@/hooks/use-jobs";
import { JobRow } from "./job-row";
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
                <TableCell colSpan={7} className="p-0 border-0">
                  <div className="p-4 bg-muted/50 rounded-lg m-2">
                    <JobTerminal jobId={job.id} />
                  </div>
                </TableCell>
              </TableRow>
            )}
          </>
        ))}
      </TableBody>
    </Table>
  );
}

export function JobList() {
  const { data: jobs, isLoading } = useJobs();
  const { stopJob, restartJob, removeJob } = useJobActions();

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
          <JobRow
            key={job.id}
            job={job}
            onStop={stopJob}
            onRestart={restartJob}
            onRemove={removeJob}
          />
        ))}
      </TableBody>
    </Table>
  );
}
