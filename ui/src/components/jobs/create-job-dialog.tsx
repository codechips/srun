import { useState, useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
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
  initialCommand?: string;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}

export function CreateJobDialog({ 
  onJobCreated, 
  initialCommand = "", 
  open, 
  onOpenChange 
}: CreateJobDialogProps) {
  const [command, setCommand] = useState(initialCommand);
  const createJob = useCreateJob();
  const queryClient = useQueryClient();

  // Reset command when dialog opens with new initialCommand
  useEffect(() => {
    if (open) {
      setCommand(initialCommand);
    }
  }, [open, initialCommand]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createJob.mutate(command, {
      onSuccess: (data) => {
        toast.success("Job created successfully");
        onOpenChange?.(false);
        onJobCreated?.(data.id);
        setCommand("");
        setTimeout(() => {
          queryClient.invalidateQueries({ queryKey: ['jobs'] });
        }, 1000);
      },
      onError: (error) => {
        toast.error(`Failed to create job: ${error.message}`);
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[850px]">
        <DialogHeader>
          <DialogTitle>{initialCommand.trim() ? 'Edit Job' : 'Create New Job'}</DialogTitle>
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
            {initialCommand ? 'Update & Run' : 'Create Job'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
