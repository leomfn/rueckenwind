import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 8081,
    proxy: {
      '/data': 'http://localhost:8080',
      '/static': 'http://localhost:8080',
    },
  },
})
