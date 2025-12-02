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
    emptyOutDir: false,
    // outDir: '../public',       // <-- output to public/
    // assetsDir: 'assets',       // <-- output to public/assets
    rollupOptions: {
      input: {
        // main: fileURLToPath(new URL('./index.html', import.meta.url)),
        index: fileURLToPath(new URL('./view/index.html', import.meta.url)),
        pushSubscription: fileURLToPath(new URL('./view/pushSubscription.html', import.meta.url)),
        signGoogle: fileURLToPath(new URL('./view/signGoogle.html', import.meta.url)),
        signTmp: fileURLToPath(new URL('./view/signTmp.html', import.meta.url)),
      }
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
