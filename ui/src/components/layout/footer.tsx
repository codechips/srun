export function Footer() {
  return (
    <footer className="fixed bottom-0 w-full text-center p-2 bg-background/95 backdrop-blur border-t">
      <div className="container max-w-7xl mx-auto text-sm text-muted-foreground">
        MIT License · Git commit: {import.meta.env.VITE_GIT_COMMIT || 'development'}
      </div>
    </footer>
  )
}
