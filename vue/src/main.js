import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

import SignComponent from './components/SignComponent.vue'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')




createApp(SignComponent).mount('#sign-app')
console.log('ddddd')

// import { pushReceive } from '@/pushReceive/pushReceive.js'
// window.pushReceive = pushReceive