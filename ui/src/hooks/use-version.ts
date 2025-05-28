import { useQuery } from "@tanstack/react-query";
import { getApiUrl } from "@/config";

interface VersionInfo {
  version: string;
  gitCommit: string;
  buildDate: string;
}

export function useVersion() {
  return useQuery<VersionInfo>({
    queryKey: ["version"],
    queryFn: async () => {
      const response = await fetch(getApiUrl("/api/version"));
      if (!response.ok) throw new Error("Failed to fetch version info");
      return response.json();
    },
  });
}
