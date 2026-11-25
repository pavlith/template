import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import packageJson from './package.json'

export default defineConfig(async () => ({
  define: {
    __APP_VERSION__: JSON.stringify(packageJson.version),
  },
  plugins: [svelte(), tailwindcss()],
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, './src/lib'),
    },
  },
  build: {
    rollupOptions: {
      // The Quick Pane is a second window with its own webview, so it needs
      // its own HTML entry. In dev, Vite serves /quick-pane.html from the
      // project root already — this is only needed for the production build.
      input: {
        main: path.resolve(__dirname, 'index.html'),
        'quick-pane': path.resolve(__dirname, 'quick-pane.html'),
      },
    },
  },
  clearScreen: false,
  server: {
    port: Number(process.env.WAILS_VITE_PORT ?? 1420),
    strictPort: true,
    host: '127.0.0.1',
  },
}))
