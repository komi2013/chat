<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import { useMessagesStore } from './stores/messages.js'
import { upsertData } from './my/indexDB.js'

setInterval(() => {console.log(navigator.onLine)}, 1000)

const messagesStore = useMessagesStore()
// if ('serviceWorker' in navigator) {
//   navigator.serviceWorker.addEventListener('message', event => {
//     console.log('before_message_parse');
//     console.log(event);
//     const data = JSON.parse(event.data.notificationData);
//     // const data = event.data.notificationData;
//     switch (data[0]) {
//       case 'message':
//         const messageData = data.slice(1)
//         console.log('messageData', messageData)
//         upsertData(JSON.stringify(messageData), 'message', 'messageID', messageData[1])
//           .then((message) => {
//             console.log(message);  // 成功時のメッセージをログに表示
//           })
//           .catch((error) => {
//             console.error(error);  // エラー時のメッセージをログに表示
//           });
//         messagesStore.update(messageData)
//         break
//       case 'community_join':
//   // arr = append(arr, channelID)
//   // arr = append(arr, aliasName)
//   // arr = append(arr, aliasImg)
//         console.log('community_join', data)
//         break
//       case 'orange':
//         console.log('This is an orange.')
//         break
//       default:
//         console.log('Unknown fruit.')
//     }

//   })
// }

// service-worker.js

navigator.serviceWorker.addEventListener('message', async (event) => {
  processNotificationData(event.data.notificationData);
});

function processNotificationData(notificationData) {
  const data = JSON.parse(notificationData);
  switch (data[0]) {
    case 'message':
      const messageData = data.slice(1);
      console.log('messageData', messageData[1]);
      const obj = {
        messageID: data[1],
        channelID: data[2],
        messageTxt: data[3],
        messageType: data[4],
        aliasName: data[5],
        aliasImg: data[6],
        editFlg: data[7],
        parentID: data[8],
        emojis: data[9],
        createdAt: data[10]
      };
      upsertData(obj, 'message', 'messageID', data[1])
        .then((message) => {
          console.log(message);  // 成功時のメッセージをログに表示
        })
        .catch((error) => {
          console.error(error);  // エラー時のメッセージをログに表示
        });
      messagesStore.update(messageData);
      break;
    case 'community_join':
      // arr = append(arr, channelID)
      // arr = append(arr, aliasName)
      // arr = append(arr, aliasImg)
      console.log('community_join', data);
      break;
    case 'orange':
      console.log('This is an orange.');
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
