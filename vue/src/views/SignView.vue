<script setup>
import { ref } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import {subscription_post} from '../my/subscription_post.js'
import { useChannelsStore } from '../stores/channels.js'
import { useMessagesStore } from '../stores/messages.js'
import {subscriptionRegister} from '../my/subscribe.js'

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
  // console.log(json[1]);

  for (const d of json[1]) {
    const channel = {
      channelID: d[0],
      channelName: d[1],
      channelDescription: d[2],
      updatedAt: d[3],
    };
    upsertData(channel, 'channel', 'channelID', channel.channelID)
      .then((message) => {
        console.log(message);
      })
      .catch((error) => {
        console.error(error);
      });
  }
  for (const d of json[2]) {
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

function openDatabase(table, key) {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('chat', 12);
    request.onerror = (event) => {
      reject(`Error opening database: ${event.target.error}`);
    };
    request.onsuccess = (event) => {
      const db = event.target.result;
      if (!db.objectStoreNames.contains(table)) {
        db.createObjectStore(table, { keyPath: key });
      }
      resolve(db);
    };
    request.onupgradeneeded = (event) => {

    };
  });
}

async function upsertData(data, table, key, objKey) {
  const db = await openDatabase(table, key);
  return new Promise(async (resolve, reject) => {
    const transaction = db.transaction([table], 'readwrite');
    const objectStore = transaction.objectStore(table);
    const existingDataRequest = objectStore.get(objKey);
    existingDataRequest.onsuccess = async () => {
      const existingData = existingDataRequest.result;
      if (existingData) {
        const putRequest = objectStore.put(data);
        putRequest.onsuccess = () => {
          resolve('Data updated successfully');
        };
        putRequest.onerror = (event) => {
          reject(`Error updating data: ${event.target.error}`);
        };
      } else {
        const addRequest = objectStore.add(data);
        addRequest.onsuccess = () => {
          resolve('Data inserted successfully');
        };
        addRequest.onerror = (event) => {
          reject(`Error inserting data: ${event.target.error}`);
        };
      }
    };
    existingDataRequest.onerror = (event) => {
      reject(`Error checking existing data: ${event.target.error}`);
    };
  });
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