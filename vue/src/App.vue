<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import { useMessagesStore } from './stores/messages.js'
import { upsertData } from './my/indexDB.js'
import { messageEdit } from './pushReceive/messageEdit.js';

setInterval(() => {console.log(navigator.onLine)}, 1000)

const messagesStore = useMessagesStore()

navigator.serviceWorker.addEventListener('message', async (event) => {
  processNotificationData(event.data.notificationData);
});

function processNotificationData(notificationData) {
  const data = JSON.parse(notificationData);
  switch (data[0]) {
    case 'message':
      const messageData = data.slice(1);
      console.log('messageData', messageData);
      const obj = {
        messageID: data[1],
        channelID: data[2],
        messageTxt: data[3],
        messageType: data[4],
        aliasName: data[5],
        aliasImg: data[6],
        createdAt: data[7]
      };
      upsertData(obj, 'message', 'messageID', data[1])
        .then((message) => {
          console.log(message);  // 成功時のメッセージをログに表示
        })
        .catch((error) => {
          console.error(error);  // エラー時のメッセージをログに表示
        });
      messagesStore.update(obj);
      break;
    case 'community_join':
      // arr = append(arr, channelID)
      // arr = append(arr, aliasName)
      // arr = append(arr, aliasImg)
      console.log('community_join', data);
      break;
    case 'msgEdit':
      console.log(data);
      messageEdit(data);
      break;
    default:
      console.log('Unknown fruit.');
  }
}



</script>

<template>
  <RouterView />
</template>

<style scoped>

</style>
