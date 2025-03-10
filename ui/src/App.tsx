import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { Toaster } from "sonner"
import { RootLayout } from "./components/layout/root-layout"
import { JobsPage } from "./pages/jobs-page"

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RootLayout>
        <JobsPage />
      </RootLayout>
      <Toaster />
    </QueryClientProvider>
  )
}

export default App
