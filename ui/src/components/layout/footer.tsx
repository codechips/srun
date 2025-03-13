export function Footer() {
  return (
    <footer className="bg-background border-t py-3">
      <div className="container max-w-7xl mx-auto text-sm text-muted-foreground">
        MIT License · commit: {import.meta.env.VITE_GIT_COMMIT || "development"}
      </div>
    </footer>
  );
}
