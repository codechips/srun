import { useState } from "react";
import { useJobs, useJobActions } from "@/hooks/use-jobs";
import { JobRow } from "./job-row";
import { CreateJobDialog } from "./create-job-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";

function LoadingTable() {
  return (
    <>
      <CreateJobDialog />
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
          ))}
        </TableBody>
      </Table>
    </>
  );
}

interface JobListProps {
  onEditJob: (command: string) => void;
}

export function JobList({ onEditJob }: JobListProps) {
  const [expandedJobId, setExpandedJobId] = useState<string | null>(null);
  const { data: jobs, isLoading } = useJobs();
  const { stopJob, restartJob, removeJob } = useJobActions();

  const handleRestart = (id: string) => {
    restartJob.mutate(id, {
      onSuccess: (newJob) => {
        setExpandedJobId(newJob.id);
      }
    });
  };

  const handleJobCreated = (id: string) => {
    setExpandedJobId(id);
  };

  if (isLoading) {
    return <LoadingTable />;
  }

  return (
    <>
      <CreateJobDialog onJobCreated={handleJobCreated} />
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
              expanded={expandedJobId === job.id}
              onExpand={setExpandedJobId}
              onStop={stopJob.mutate}
              onRestart={handleRestart}
              onRemove={removeJob.mutate}
              onEdit={onEditJob}
            />
          ))}
        </TableBody>
      </Table>
    </>
  );
}
