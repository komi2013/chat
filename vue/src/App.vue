<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import HelloWorld from './components/HelloWorld.vue'
import Form from './components/Form.vue'
import DrawerColumn from './components/DrawerColumn.vue'
import {get_formated_time} from './my/get_formated_time.js'
import {init} from './my/init.js'
import {subscribe} from './my/subscribe.js'
import {subscription_post} from './my/subscription_post.js'

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


// function setPush() {
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.register('service-worker.js');
  navigator.serviceWorker.ready
    .then(function(registration) {
      return registration.pushManager.getSubscription();
    })
    .then(function(subscription) {
      // console.log('userID, alias value', userID.value, alias.value)
      // console.log('userID, alias', userID, alias)
      if (!subscription) {
        subscribe('seijiro', 'sei1')
      } else {
        console.log(JSON.stringify(subscription))
        
        subscription_post(JSON.stringify(subscription),
          'seijiro', 'sei1')
      }
    });

  navigator.serviceWorker.addEventListener('message', event => {
    const notificationData = event.data.notificationData;
    console.log(`Received data from Service Worker: "${notificationData}"`);
  });
}


// // Main Page (index.html or any other page)
// if ('serviceWorker' in navigator) {
//   navigator.serviceWorker.register('service-worker.js').then(registration => {
//     // Communicate with the service worker using MessageChannel
//     const { port1, port2 } = new MessageChannel();
//     port1.onmessage = ev => {
//       console.log('[Main Page] Received data from Service Worker:', ev.data);
//     };

//     // // Send the port to the service worker
//     // navigator.serviceWorker.controller.postMessage({ type: 'init', port: port2 });

//     // // Example: Sending a message to the service worker
//     // port2.postMessage({ type: 'hello' });
//   });
// }





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
