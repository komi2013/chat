import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  base: '/vue/',
  plugins: [
    vue(),
  ],
  build: {
    outDir: '../public/vue',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        // JS の出力ファイル名を固定
        entryFileNames: 'assets/index.js',
        // 追加で分割されるチャンクの命名
        chunkFileNames: 'assets/[name].[hash].js',
        // CSS の出力も固定
        // assetFileNames: 'assets/[name].[hash].[ext]',
        assetFileNames: (assetInfo) => {
          if (assetInfo.name && assetInfo.name.endsWith('.css')) {
            return 'assets/index.css';
          }
          return 'assets/[name].[hash].[ext]';
        }
      }
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
