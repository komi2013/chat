<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import Form from './components/Form.vue'
import DrawerColumn from './components/DrawerColumn.vue'
import { useMessagesStore } from './stores/messages.js'
// import * as my from './my'
// import {subscriptionRegister} from './my/subscribe'

let closedTime = null
const msg5 = ref('')
let initData = ref(null)

const props = defineProps({
  conn: {
    type: Object,
    required: true
  },
  message: ref('')
})

setInterval(() => {console.log(navigator.onLine)}, 1000)
const msgs = ref('')
const channels = ref([[1,"channel face 1","description 1"],[1,"channel init 1","description 1"]])


const messagesStore = useMessagesStore()
// my.subscriptionRegister()
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.addEventListener('message', event => {
    const notificationData = event.data.notificationData;
    let data = ["","6545c74e71aee3b6bb941191","micro plastic ","1","sei1","/me.jpg","1",0,null,"2023-11-10T14:30:23.352583421Z"]
    console.log(data)
    console.log(notificationData)
    console.log(JSON.parse(notificationData))
    messagesStore.update(JSON.parse(notificationData))
    console.log(`Received data from Service Worker: "${notificationData}"`);
  });
}

</script>

<template>

  <DrawerColumn />
  
<!--   <div id="content">
    <RouterView @custom-event="handleCustomEvent" @custom-event2="handleCustomEvent" />
  </div> -->
<br><br>
<RouterView />
<br><br><br><br><br>
      <RouterLink to="/">Home</RouterLink><br>
      <RouterLink to="/about">About</RouterLink><br>
      <RouterLink to="/sign" >Sign</RouterLink><br>
      <RouterLink to="/channel/1/" >channel q</RouterLink><br>
      <button @click="setPush">setPush</button>
</template>

<style scoped>

</style>
