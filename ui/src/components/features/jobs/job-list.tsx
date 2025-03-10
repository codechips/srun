import { useQuery } from "@tanstack/react-query"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { MoreVertical, Play, Square, Trash } from "lucide-react"
import { Button } from "@/components/ui/button"

interface Job {
  id: string
  pid: number
  command: string
  status: string
  startedAt: string
  completedAt?: string
}

export function JobList() {
  const { data: jobs, isLoading } = useQuery<Job[]>({
    queryKey: ['jobs'],
    queryFn: async () => {
      const response = await fetch('/api/jobs')
      return response.json()
    }
  })

  const handleStop = async (id: string) => {
    await fetch(`/api/jobs/${id}`, { method: 'DELETE' })
  }

  const handleRestart = async (id: string) => {
    await fetch(`/api/jobs/${id}/restart`, { method: 'POST' })
  }

  const handleRemove = async (id: string) => {
    // TODO: Add confirmation dialog
    await fetch(`/api/jobs/${id}`, { method: 'DELETE' })
  }

  if (isLoading) {
    return <div>Loading...</div>
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
          <TableRow key={job.id}>
            <TableCell className="font-mono">{job.id.slice(0, 8)}</TableCell>
            <TableCell>{job.pid}</TableCell>
            <TableCell>
              <Badge variant={job.status === 'running' ? 'default' : 'secondary'}>
                {job.status}
              </Badge>
            </TableCell>
            <TableCell className="font-mono">{job.command}</TableCell>
            <TableCell>{new Date(job.startedAt).toLocaleString()}</TableCell>
            <TableCell>
              {job.completedAt ? new Date(job.completedAt).toLocaleString() : '-'}
            </TableCell>
            <TableCell className="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" className="h-8 w-8 p-0">
                    <MoreVertical className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  {job.status === 'running' ? (
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
  )
}
