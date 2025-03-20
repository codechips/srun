import { useQuery } from "@tanstack/react-query";

interface VersionInfo {
  version: string;
  gitCommit: string;
  buildDate: string;
}

export function useVersion() {
  return useQuery<VersionInfo>({
    queryKey: ["version"],
    queryFn: async () => {
      const response = await fetch("/api/version");
      if (!response.ok) throw new Error("Failed to fetch version info");
      return response.json();
    },
  });
}
