import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

import { pushReceive } from '@/pushReceive/pushReceive.js'

navigator.serviceWorker.addEventListener('message', async (event) => {
  pushReceive(event.data.notificationData, true)
})

const app = createApp(App)
app.use(createPinia())
app.use(router)

// Vue コンポーネント内のエラーをキャッチ
app.config.errorHandler = (err, instance, info) => {
  console.error('Vue error:', err, info)

  sendErrorLog({
    message: err.message,
    stack: err.stack,
    info,
    component: instance?.$options?.name || '(anonymous)'
  })
}

app.mount('#app')

checkFaviconBadge()