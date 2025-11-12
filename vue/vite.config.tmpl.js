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
    // outDir: '../public/vue-tmpl',
    emptyOutDir: false, // ← SPA側出力を消さない
    rollupOptions: {
      input: {
        outsideApp: fileURLToPath(new URL('./src/outsideApp.js', import.meta.url))
      },
      output: {
        entryFileNames: 'assets/outsideApp.js',
        assetFileNames: assetInfo => {
          if (assetInfo.name && assetInfo.name.endsWith('.css')) {
            return 'assets/outsideApp.css' // ✅ CSSファイル名を固定化！
          }
          return 'assets/[name][extname]'
        }
      }
    }
  }
})
