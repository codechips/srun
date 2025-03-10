import { JobList } from "@/components/features/jobs/job-list"

export function JobsPage() {
  return (
    <div className="py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Jobs</h1>
        <button className="bg-primary text-primary-foreground hover:bg-primary/90 px-4 py-2 rounded-md">
          New Job
        </button>
      </div>
      <JobList />
    </div>
  )
}
