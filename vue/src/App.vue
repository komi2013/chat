<script setup>
import { RouterLink, RouterView } from 'vue-router'
import { ref, onUpdated } from 'vue'
import { useMessagesStore } from './stores/messages.js'
import { upsertData } from './my/indexDB.js'
import { messageEdit } from './pushReceive/messageEdit.js';
import { thread } from './pushReceive/thread.js';
import { threadEdit } from './pushReceive/threadEdit.js';
import { message } from './pushReceive/message.js';

const messagesStore = useMessagesStore()

navigator.serviceWorker.addEventListener('message', async (event) => {
  processNotificationData(event.data.notificationData);
});

function processNotificationData(notificationData) {
  const data = JSON.parse(notificationData);
  switch (data[0]) {
    case 'message':
      message(data);
      break;
    case 'community_join':
      console.log('community_join', data);
      break;
    case 'msgEdit':
      messageEdit(data);
      break;
    case 'thread':
      thread(data);
      break;
    case 'threadEdit':
      threadEdit(data);
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
