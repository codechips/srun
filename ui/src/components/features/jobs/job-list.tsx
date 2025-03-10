import { useQuery } from "@tanstack/react-query"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

interface Job {
  id: string
  command: string
  status: string
  startedAt: string
}

export function JobList() {
  const { data: jobs, isLoading } = useQuery<Job[]>({
    queryKey: ['jobs'],
    queryFn: async () => {
      const response = await fetch('/api/jobs')
      return response.json()
    }
  })

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div className="grid gap-4">
      {jobs?.map((job) => (
        <Card key={job.id}>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span className="font-mono text-sm">{job.command}</span>
              <Badge variant={job.status === 'running' ? 'default' : 'secondary'}>
                {job.status}
              </Badge>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              Started: {new Date(job.startedAt).toLocaleString()}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
