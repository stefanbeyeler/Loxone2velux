import { defineConfig } from 'vite'
import { readFileSync } from 'fs'
import { resolve } from 'path'
import react from '@vitejs/plugin-react'

function getVersion(): string {
  if (process.env.VITE_APP_VERSION) return process.env.VITE_APP_VERSION
  try {
    return readFileSync(resolve(__dirname, '../VERSION'), 'utf-8').trim()
  } catch {
    return 'dev'
  }
}

export default defineConfig({
  plugins: [react()],
  base: './',
  define: {
    __APP_VERSION__: JSON.stringify(getVersion()),
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
