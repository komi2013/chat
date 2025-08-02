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
app.mount('#app')

checkFaviconBadge()