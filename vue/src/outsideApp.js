import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

import SignComponent from './components/SignComponent.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')

window.pushReceive = pushReceive

createApp(SignComponent).mount('#sign-app')
console.log('SignComponent_Test')