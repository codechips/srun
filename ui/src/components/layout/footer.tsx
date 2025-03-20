import { useVersion } from "@/hooks/use-version";

export function Footer() {
  const { data: version } = useVersion();

  return (
    <footer className="bg-background border-t py-3">
      <div className="container max-w-7xl mx-auto text-sm text-muted-foreground">
        MIT License · v{version?.version || "dev"} ({version?.gitCommit?.slice(0, 7) || "unknown"}) · Built at {version?.buildDate || ""}
      </div>
    </footer>
  );
}
