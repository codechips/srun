import path from "path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { execSync } from 'child_process';

// Get git commit hash
const getGitCommitHash = () => {
  try {
    return execSync('git rev-parse --short HEAD').toString().trim();
  } catch {
    return 'unknown';
  }
};

export default defineConfig({
  plugins: [react(), tailwindcss()],
  base: '/', // Ensure assets are loaded from root path
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8000",
        changeOrigin: true,
        ws: true, // Enable WebSocket proxy
      },
    },
  },
  define: {
    'import.meta.env.VITE_GIT_COMMIT': JSON.stringify(getGitCommitHash()),
  },
});
