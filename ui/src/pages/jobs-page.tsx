import { JobList } from "@/components/features/jobs/job-list"
import { CreateJobDialog } from "@/components/features/jobs/create-job-dialog"

export function JobsPage() {
  return (
    <div className="py-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Jobs</h1>
        <CreateJobDialog />
      </div>
      <JobList />
    </div>
  )
}
