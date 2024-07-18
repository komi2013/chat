<script setup>
import { RouterLink, RouterView } from 'vue-router';
import { ref, onUpdated } from 'vue';
import { upsertData } from './my/indexDB.js';
import { bookmark } from './pushReceive/bookmark.js';
import { channelJoin } from './pushReceive/channelJoin.js';
import { channelAdd } from './pushReceive/channelAdd.js';
import { channelEdit } from './pushReceive/channelEdit.js';
import { emoji } from './pushReceive/emoji.js';
import { groupAliasEdit } from './pushReceive/groupAliasEdit.js';
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
    case 'channelAdd':
      channelAdd(data);
      break;
    case 'channelEdit':
      channelEdit(data);
      break;
    case 'emoji':
      emoji(data);
      break;
    case 'groupAliasEdit':
      groupAliasEdit(data);
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
