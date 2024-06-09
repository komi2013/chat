<script setup>
import { RouterLink, RouterView } from 'vue-router';
import { ref, onUpdated } from 'vue';
import { upsertData } from './my/indexDB.js';
import { bookmark } from './pushReceive/bookmark.js';
import { channelJoin } from './pushReceive/channelJoin.js';
import { channelUpsert } from './pushReceive/channelUpsert.js';
import { message } from './pushReceive/message.js';
import { emoji } from './pushReceive/emoji.js';
import { messageEdit } from './pushReceive/messageEdit.js';
import { rookie } from './pushReceive/rookie.js';
import { thread } from './pushReceive/thread.js';
import { threadEdit } from './pushReceive/threadEdit.js';

navigator.serviceWorker.addEventListener('message', async (event) => {
  processNotificationData(event.data.notificationData);
});

function processNotificationData(notificationData) {
  const data = JSON.parse(notificationData);
  switch (data[1]) {
    case 'bookmark':
      bookmark(data);
      break;
    case 'channelJoin':
      channelJoin(data);
      break;
    case 'channel':
      channelUpsert(data);
      break;
    case 'community_join':
      console.log('community_join', data);
      break;
    case 'emoji':
      emoji(data);
      break;
    case 'message':
      message(data);
      break;
    case 'msgEdit':
      messageEdit(data);
      break;
    case 'rookie':
      rookie(data);
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
