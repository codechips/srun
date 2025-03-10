import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useCreateJob } from "@/hooks/use-jobs";
import { toast } from "sonner";

interface CreateJobDialogProps {
  onJobCreated?: (jobId: string) => void;
}

export function CreateJobDialog({ onJobCreated }: CreateJobDialogProps) {
  const [command, setCommand] = useState("");
  const [open, setOpen] = useState(false);
  const createJob = useCreateJob();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createJob.mutate(command, {
      onSuccess: (data) => {
        toast.success("Job created successfully");
        onJobCreated?.(data.id);
        setOpen(false);
        setCommand("");
      },
      onError: (error) => {
        toast.error(`Failed to create job: ${error.message}`);
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>New Job</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[850px]">
        <DialogHeader>
          <DialogTitle>Create New Job</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Textarea
              id="command"
              placeholder="Enter command or script..."
              value={command}
              onChange={(e) => setCommand(e.target.value)}
              className="font-mono"
              rows={10}
            />
          </div>
          <Button type="submit" disabled={!command.trim()}>
            Create Job
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
