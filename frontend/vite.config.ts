import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [svelte()],
  build: {
    // Emit manifest.json so the Go asset resolver can find hashed filenames
    manifest: true,
    rollupOptions: {
      input: resolve(__dirname, 'src/main.ts'),
    },
    // Output directly into the Go embed directory
    outDir: resolve(__dirname, '../static'),
    emptyOutDir: true,
  },
  server: {
    // Allow the Go server to load scripts cross-origin in dev mode
    cors: true,
  },
})
