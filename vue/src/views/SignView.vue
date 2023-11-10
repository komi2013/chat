<script setup>
import { ref } from 'vue'
import {subscribe} from '../my/subscribe.js'
import {subscription_post} from '../my/subscription_post.js'
import { useChannelsStore } from '../stores/channels.js'
import { useMessagesStore } from '../stores/messages.js'

const count = ref(0)
const userID = ref('')

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

function setCookie() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('service-worker.js');
    navigator.serviceWorker.ready
      .then(function(registration) {
        return registration.pushManager.getSubscription();
      })
      .then(function(subscription) {
        if (!subscription) {
          subscribe()
        } else {
          console.log(JSON.stringify(subscription));
          subscription_post(JSON.stringify(subscription),userID.value)
        }
      });
  }
}

</script>

<template>

  <!-- <DrawerColumn :channels="channels" /> -->
  <input v-model="userID" placeholder="seijiro" />
  <button @click="setCookie">click</button>
  <div>{{count}}</div>
</template>