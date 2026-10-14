import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// In development the frontend runs on its own port (5173) and this proxy
// forwards /api to the Go backend (8080), avoiding CORS. In production it is
// not needed: the Go binary serves everything from the same origin.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
