import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  // root: '.', 
  // base: '/vue/',
  plugins: [
    vue(),
  ],
  build: {
    emptyOutDir: true,
    outDir: '../view',       // <-- output to public/
    assetsDir: '../public/assets',       // <-- output to public/assets
    rollupOptions: {
      input: {
        // main: fileURLToPath(new URL('./index.html', import.meta.url)),
        index: fileURLToPath(new URL('./index.html', import.meta.url)),
        pushSubscription: fileURLToPath(new URL('./pushSubscription.html', import.meta.url)),
        signGoogle: fileURLToPath(new URL('./signGoogle.html', import.meta.url)),
        signTmp: fileURLToPath(new URL('./signTmp.html', import.meta.url)),
      }
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
