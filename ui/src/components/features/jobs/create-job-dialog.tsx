import { useState, useEffect } from "react"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { JobTerminal } from "./job-terminal"
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"

export function CreateJobDialog() {
  const [command, setCommand] = useState("")
  const [showLogs, setShowLogs] = useState(false)
  const [jobId, setJobId] = useState<string>("")
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: async (command: string) => {
      const response = await fetch('/api/jobs', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command }),
      })
      if (!response.ok) throw new Error('Failed to create job')
      return response.json()
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['jobs'] })
      toast.success('Job created successfully')
      setJobId(data.id)
      setShowLogs(true)
    },
    onError: (error) => {
      toast.error(`Failed to create job: ${error.message}`)
    }
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    createMutation.mutate(command)
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button>New Job</Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[850px]">
        <DialogHeader>
          <DialogTitle>Create New Job</DialogTitle>
        </DialogHeader>
        {showLogs ? (
          <div className="py-4">
            <JobTerminal jobId={jobId} />
          </div>
        ) : (
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
      )}
      </DialogContent>
    </Dialog>
  )
}
