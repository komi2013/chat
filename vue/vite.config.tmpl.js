// vite.config.tmpl.js
import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: false, // ← SPA側出力を消さない
    rollupOptions: {
      input: {
        outsideApp: fileURLToPath(new URL('./src/outsideApp.js', import.meta.url))
      },
      output: {
        entryFileNames: 'assets/outsideApp.js'
      }
    }
  }
})
