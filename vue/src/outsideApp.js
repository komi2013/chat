import './assets/main.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'

import Advertisement from '@/components/Advertisement.vue'
import Drawer from '@/components/Drawer.vue'
import { pushReceive } from '@/pushReceive/pushReceive.js'

// 🔥 Pinia を1つ作る
const pinia = createPinia()

// Drawer
const drawerApp = createApp(Drawer)
drawerApp.use(pinia)
drawerApp.mount('#drawer_column')

// Advertisement（同じ Pinia を使う）
createApp(Advertisement).use(pinia).mount('#ad-right1')
createApp(Advertisement).use(pinia).mount('#ad-right2')
createApp(Advertisement).use(pinia).mount('#ad-right3')

// 🔥 Vue外から使う入口
import { setActivePinia } from 'pinia'
window.pushReceive = (data, fromPush = false) => {
  setActivePinia(pinia)
  return pushReceive(data, fromPush)
}
