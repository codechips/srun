export function Footer() {
  return (
    <footer className="bg-background border-t py-6">
      <div className="container max-w-7xl mx-auto text-sm text-muted-foreground">
        MIT License · Git commit: {import.meta.env.VITE_GIT_COMMIT || 'development'}
      </div>
    </footer>
  )
}
