import { JobList } from "@/components/jobs/job-list";
import { CreateJobDialog } from "@/components/jobs/create-job-dialog";

export function JobsPage() {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editCommand, setEditCommand] = useState("");

  return (
    <div className="py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Jobs</h1>
        <CreateJobDialog 
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          initialCommand={editCommand}
        />
      </div>
      <JobList onEditJob={(command) => {
        setEditCommand(command);
        setDialogOpen(true);
      }} />
    </div>
  );
}
