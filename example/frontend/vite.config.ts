import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [svelte()],
  build: {
    manifest: true,
    rollupOptions: {
      input: resolve(__dirname, 'src/main.ts'),
    },
    outDir: resolve(__dirname, '../static'),
    emptyOutDir: true,
  },
  server: {
    cors: true,
  },
})
