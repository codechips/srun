import type { ReactNode } from "react";
import { Footer } from "./footer";

export function RootLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <div className="flex flex-col">
        <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="container max-w-7xl mx-auto flex h-14 items-center">
            <div className="mr-4 flex">
              <a className="mr-6 flex items-center space-x-2" href="/">
                <span className="app-title text-2xl text-gray-700">srun</span>
              </a>
            </div>
          </div>
        </header>
        <main className="flex-1 pb-16">
          <div className="container max-w-7xl mx-auto">{children}</div>
        </main>
        <Footer />
      </div>
    </div>
  );
}
