<script setup>
import { ref, computed, onBeforeMount } from 'vue';
import { onBeforeRouteUpdate, useRouter } from 'vue-router';
import DrawerColumn from '../components/DrawerColumn.vue';
import EditBox from '../components/EditBox.vue';
import Messages from '../components/Messages.vue';

import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit, adjustHeight, textareaRefs } from '../my/other.js';
import { removeMark } from '../my/markdown.js';

import { getSubstring, getParam } from '../my/strings.js';

const props = defineProps({
  message_id: '',
  channel_id: '',
  back_id: ''
})

const message = ref('');
const threadHead = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();

const messages = computed(() => {
  return messagesStore.messages;
});


const msg = {
  messageTxt: '',
  messageID: ''
};

const channel = ref('');

async function fetchChannel() {
  console.log('props.channel_id', props.channel_id);
  try {
    const data = await getIDB('channel', props.channel_id);
    channel.value = data;
  } catch (error) {
    channel.value = null;
  }
}

async function fetchThreadHead() {
  try {
    const data = await getIDB('threadHead', props.message_id);
    threadHead.value = data;
    let message = {};
    message.messageID = data.parentID;
    message.parentID = data.parentID;
    message.messageTxt = data.messageTxt;
    message.aliasName = data.aliasName;
    message.aliasImg = data.aliasImg;
    message.createdAt = data.createdAt;
    message.emojis = data.emojis;
    messagesStore.insert(message);
  } catch (error) {
    console.log('no threadHead', error);
    try {
      const message = await getIDB('message', props.message_id);
      threadHead.value = message;
      threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
      threadHead.value.messageTxt = message.messageTxt;
      threadHead.value.backURL = '/channel/' + props.channel_id + '/';
      console.log('threadHead.value', threadHead.value);
      messagesStore.insert(message);
    } catch (error) {
      console.log('no message', error);
      try {
        const message = await getIDB('thread', props.message_id);
        threadHead.value = message;
        threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
        threadHead.value.messageTxt = message.messageTxt;
        threadHead.value.backURL = '/thread/' + props.channel_id + '/' + props.back_id + '/noback/';
        console.log('threadHead.value', threadHead.value.messageID);
        messagesStore.insert(message);
      } catch (error) {
        console.error('Failed to fetch thread:', error);
        threadHead.value = null;
      }
    }
  }
}

const fetchMessageData = () => {
  return new Promise((resolve, reject) => {
    console.log('latee', props.message_id);
    getIDBs('thread', 'parentIDIndex', props.message_id)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(message => {
          console.log('message', message);
          messagesStore.insert(message);
        });
        resolve();
      })
      .catch((error) => {
        reject(error);
      });
  });
};

onBeforeMount(async () => {
  await messagesStore.deleteAll();
  await fetchChannel();
  await fetchThreadHead();
  await fetchMessageData();

  parentID.value = props.message_id;
  console.log('parentID.value', parentID.value);
  const content = document.getElementById('content');
  content.scrollTop = content.scrollHeight;
  window.scrollTo(0, content.scrollHeight);
});


onBeforeRouteUpdate((to, from, next) => {
  if (from.path != to.path) {
    console.log('ページ遷移が検出されました:', from.path, to.path);
    location.href = to.path;
  }
});

</script>

<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <div style="width: 90%;" v-html="threadHead.title"> </div>
  <div style="line-height: 50px;width: 50px;">
    <RouterLink v-if="threadHead.backURL" :to="threadHead.backURL"> ⬅ </RouterLink>
  </div>
</div>

  <template v-if="channel">
    <Messages :channel="channel" :messages="messages" />
  </template>

<div class="editText">
  <template v-if="parentID">
    <EditBox :channel="channel" :message="msg" :parent_id="parentID" />
  </template>
</div>
<br>
</div>
</template>

<style>

#content {
  height: 100%;
  overflow-y: auto;
}

@media screen and (min-width : 701px) { 

}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
  }
  .headTitle div {
    display: table-cell;
  }
}
</style>

