<script setup>
import { ref } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import {subscription_post} from '../my/subscription_post.js'
import { useChannelsStore } from '../stores/channels.js'
import { useMessagesStore } from '../stores/messages.js'
import {subscriptionRegister} from '../my/subscribe.js'
import { upsertData } from '../my/indexDB.js'

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
  for (let i = 0; i < json[2].length; i++) {
    const d = json[2][i];
    const alias = {
      aliasName: d[0],
      aliasImg: d[1],
      groupFlg: d[2],
    };
    upsertData(alias, 'alias', 'aliasName', alias.aliasName)
      .then((message) => {
        console.log(message);
      })
      .catch((error) => {
        console.error(error);
      });
  }
  for (const d of json[1]) {
    const channel = {
      channelID: d[0],
      channelName: d[1],
      channelDescription: d[2],
      updatedAt: d[3],
      aliasArray: d[4],
      aliasName: d[5],
      displayStatus: 1
    };
    upsertData(channel, 'channel', 'channelID', channel.channelID)
      .then((message) => {
        console.log(message);
      })
      .catch((error) => {
        console.error(error);
      });
  }
  channelsStore.insert(json[1]);
  // messagesStore.insert(json[2]);
  // msgs.value = json[2]
})
.catch((reason)=>{
  console.log(reason)
});

function setCookie() {
  if ('serviceWorker' in navigator) {
    subscriptionRegister()
    console.log(userID.value, alias.value)
    subscription_post(userID.value, alias.value)
  }
}

function unregister() {
  if ('serviceWorker' in navigator) {
    navigator.serviceWorker.getRegistrations().then(registrations => {
      for (const registration of registrations) {
        registration.unregister();
      }
    });
  }
}

</script>

<template>
  <DrawerColumn />
  <div id="content">
    <br><br>
  <!-- <DrawerColumn :channels="channels" /> -->
  <input v-model="userID" placeholder="seijiro" />
  <input v-model="alias" placeholder="sei1" />
  <br>
  <button @click="setCookie()">setCookie</button>
  <button @click="unregister()">unregister</button>
  <div>{{count}}</div>
</div>
</template>