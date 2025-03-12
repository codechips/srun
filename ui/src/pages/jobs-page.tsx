import { useState } from "react";
import { JobList } from "@/components/jobs/job-list";
import { CreateJobDialog } from "@/components/jobs/create-job-dialog";
import { Button } from "@/components/ui/button";

export function JobsPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editCommand, setEditCommand] = useState("");

  const handleNewJob = () => {
    setEditCommand(""); // Reset the command when opening new job dialog
    setDialogOpen(true);
  };

  return (
    <div className="py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Jobs</h1>
        <div className="flex gap-2">
          <Button onClick={handleNewJob}>New Job</Button>
          <CreateJobDialog 
            open={dialogOpen}
            onOpenChange={setDialogOpen}
            initialCommand={editCommand}
          />
        </div>
      </div>
      <JobList onEditJob={(command) => {
        setEditCommand(command);
        setDialogOpen(true);
      }} />
    </div>
  );
}
