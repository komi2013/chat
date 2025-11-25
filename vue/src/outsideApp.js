import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

// import App from './App.vue'
// import router from './router'

import Advertisement from '@/components/Advertisement.vue';
import Drawer from '@/components/Drawer.vue'


// import SignComponent from './components/SignComponent.vue'

import { pushReceive } from '@/pushReceive/pushReceive.js'

// const app = createApp(App)

// app.use(createPinia())
// app.use(router)

// app.mount('#app')

window.pushReceive = pushReceive

createApp(Drawer).mount('#drawer_column')

createApp(Advertisement).mount('#ad-right1')
createApp(Advertisement).mount('#ad-right2')
createApp(Advertisement).mount('#ad-right3')

// createApp(SignComponent).mount('#sign-app')
// console.log('SignComponent_Test')