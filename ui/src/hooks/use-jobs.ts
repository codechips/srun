import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { getApiUrl } from "@/config";

export interface Job {
  id: string;
  pid: number;
  command: string;
  status: string;
  startedAt: string;
  completedAt?: string;
}

export function useJobs() {
  return useQuery<Job[]>({
    queryKey: ["jobs"],
    queryFn: async () => {
      const response = await fetch(getApiUrl("/api/jobs"));
      if (!response.ok) throw new Error("Failed to fetch jobs");
      return response.json();
    },
    refetchInterval: 5000, // Refresh every 5 seconds
  });
}

export function useCreateJob() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (command: string) => {
      const response = await fetch(getApiUrl("/api/jobs"), {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command }),
      });
      if (!response.ok) throw new Error("Failed to create job");
      return response.json();
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["jobs"] });
    },
  });
}

export function useJobActions() {
  const queryClient = useQueryClient();

  const stopJob = useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(getApiUrl(`/api/jobs/${id}/stop`), { method: "POST" });
      if (!response.ok) throw new Error("Failed to stop job");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["jobs"] });
      toast.success("Job stopped successfully");
    },
    onError: (error) => {
      toast.error(`Failed to stop job: ${error.message}`);
    },
  });

  const restartJob = useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(getApiUrl(`/api/jobs/${id}/restart`), {
        method: "POST",
      });
      if (!response.ok) throw new Error("Failed to restart job");
      return response.json();
    },
    onSuccess: (data, id) => {
      queryClient.setQueryData(["jobs", id], data);
      queryClient.invalidateQueries({ queryKey: ["jobs"] });
      toast.success("Job restarted successfully");
    },
    onError: (error) => {
      toast.error(`Failed to restart job: ${error.message}`);
    },
  });

  const removeJob = useMutation({
    mutationFn: async (id: string) => {
      const response = await fetch(getApiUrl(`/api/jobs/${id}`), { method: "DELETE" });
      if (!response.ok) throw new Error("Failed to remove job");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["jobs"] });
      toast.success("Job removed successfully");
    },
    onError: (error) => {
      toast.error(`Failed to remove job: ${error.message}`);
    },
  });

  return {
    stopJob,
    restartJob,
    removeJob,
  };
}
