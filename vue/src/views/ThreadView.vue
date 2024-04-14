<script setup>
import { ref, computed, onBeforeMount } from 'vue';
import { onBeforeRouteUpdate, useRouter } from 'vue-router';
import DrawerColumn from '../components/DrawerColumn.vue';
import EditBox from '../components/EditBox.vue';
import Messages from '../components/Messages.vue';

import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';

import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData, updOne } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { removeMark } from '../my/markdown.js';

import { getSubstring, getParam } from '../my/strings.js';

const props = defineProps({
  message_id: '',
  channel_id: ''
})

const message = ref('');
let threadHead = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();

const messages = computed(() => {
  return messagesStore.messages;
});

const msg = {
  messageTxt: '',
  messageID: '',
  parentID: props.message_id
};

const channel = ref('');

async function fetchChannel() {
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
    if (data.aliasName == channel.value.aliasName) {
      data.edit = true;
    }
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
    if (getParam('backID')) {
      const message = await getIDB('thread', props.message_id);
      threadHead.value = message;
      threadHead.value.parentID = props.message_id;
      threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
      threadHead.value.messageTxt = message.messageTxt;
      threadHead.value.backID = getParam('backID');
      messagesStore.insert(message);
    } else {
      const message = await getIDB('message', props.message_id);
      threadHead.value = message;
      threadHead.value.parentID = props.message_id;
      threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
      threadHead.value.messageTxt = message.messageTxt;
      messagesStore.insert(message);
    }
  }
}

function readStatus () {
  if (threadHead.value.displayStatus && threadHead.value.displayStatus == 1 || threadHead.value.displayStatus == 2) {
    threadHead.value.displayStatus = 0;
    updOne('threadHead', props.message_id, 'displayStatus', 0)
      .catch((error) => {
        console.error(error);
      });
    const favicon = document.querySelector('link[rel="icon"]');
    favicon.href = '/favicon.ico';
  }
}

onBeforeMount(async () => {
  await fetchChannel();
  await fetchThreadHead();
  readStatus();
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

function backTo(backID) {
  if (backID) {
    return '/thread/' + props.channel_id + '/' + backID + '/';
  } else {
    return '/channel/' + props.channel_id + '/';
  }
}

</script>

<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <div style="width: 90%;" >
    <a :href="'/threadHead/' + threadHead.parentID + '/'">
      {{threadHead.title}}
    </a>
  </div>
  <div style="line-height: 50px;width: 50px;">
    <a :href="backTo(threadHead.backID)"> ⬅ </a>
   <!--  <span v-if="threadHead.edit" class="emoji" > 🖋 </span> -->
  </div>
</div>

  <template v-if="channel && messages && threadHead">
    <Messages :channel="channel" :messages="messages" :threadHead="threadHead" />
  </template>

<div class="editText" v-if="channel && threadHead">
  <EditBox :channel="channel" :message="msg" :threadHead="threadHead" />
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
  .headTitle {
    margin-left: 50px;
    display: flex;
  }
}

@media screen and (max-width : 700px) {
  .headTitle {
    margin-left: 50px;
    display: flex;
  }
  .headTitle div {
    display: table-cell;
  }
}
</style>

