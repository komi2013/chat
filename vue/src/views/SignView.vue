<script setup>
import { ref } from 'vue'
import {subscribe} from '../my/subscribe.js'
import {subscription_post} from '../my/subscription_post.js'
import { useChannelsStore } from '../stores/channels.js'
import { useMessagesStore } from '../stores/messages.js'

const count = ref(0)
const userID = ref('')
const alias = ref('')

const props = defineProps({
  channels: '',
})


const channelsStore = useChannelsStore()
const messagesStore = useMessagesStore()

const request = new Request('/Init/', {
  method: 'POST',
});
fetch(request)
.then((response)=>{
  if(!response.ok){
    throw new Error();
  }
  return response.json()
})
.then((json)=>{
  // channels.value = json[1]
  channelsStore.insert(json[1])
  messagesStore.insert(json[2])
  // msgs.value = json[2]
})
.catch((reason)=>{
  console.log(reason)
});

// function setCookie(userID, alias) {
//   if ('serviceWorker' in navigator) {
//     navigator.serviceWorker.register('service-worker.js');
//     navigator.serviceWorker.ready
//       .then(function(registration) {
//         return registration.pushManager.getSubscription();
//       })
//       .then(function(subscription) {
//         console.log('userID, alias value', userID.value, alias.value)
//         console.log('userID, alias', userID, alias)
//         if (!subscription) {
//           subscribe('seijiro', 'sei1')
//         } else {
//           console.log(JSON.stringify(subscription))
          
//           subscription_post(JSON.stringify(subscription),
//             'seijiro', 'sei1')
//         }
//       });
//   }
// }

</script>

<template>
  <div id="content">
  <!-- <DrawerColumn :channels="channels" /> -->
  <input v-model="userID" placeholder="seijiro" />
  <input v-model="alias" placeholder="sei1" />
  <button @click="setCookie(userID, alias)">click</button>
  <div>{{count}}</div>
</div>
</template>