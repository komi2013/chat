<script setup>
import { ref, computed, onBeforeMount } from 'vue'
import DrawerColumn from '../components/DrawerColumn.vue'
import EditBox from '../components/EditBox.vue'
import Messages from '../components/Messages.vue'

import { useMessagesStore } from '../stores/messages.js';
import { useChannelsStore } from '../stores/channels.js';
import { get_formated_time } from '../my/get_formated_time.js';
import { getIDB, getIDBs, upsertData } from '../my/indexDB.js';
import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '../my/emoji.js';
import { isOtherOpen, otherMessageId, openOther, closeOther, selectOther, activeEdit, adjustHeight, textareaRefs } from '../my/other.js';
import { textToHtml } from '../my/textToHtml.js';

import { getSubstring, removeHtmlTags } from '../my/strings.js';

const props = defineProps({
  message_id: '',
  channel_id: ''
})

const message = ref('');
const threadHead = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();
messagesStore.deleteAll();
const messages = computed(() => {
  console.log('computed', messagesStore.messages);
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
      threadHead.value.title = getSubstring(removeHtmlTags(textToHtml(message.messageTxt)), 0, 8);
      threadHead.value.messageTxt = message.messageTxt;
      console.log('threadHead.value', threadHead.value.messageID);
      messagesStore.insert(message);
    } catch (error) {
      console.error('Failed to fetch message:', error);
      threadHead.value = null;
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
  await fetchChannel();
  await fetchThreadHead();
  await fetchMessageData();

  parentID.value = threadHead.value.parentID;
  console.log('parentID.value', parentID.value);
  const content = document.getElementById('content');
  content.scrollTop = content.scrollHeight;
  window.scrollTo(0, content.scrollHeight);
});

</script>

<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <div v-html="threadHead.title"> </div>
</div>

  <template v-if="channel">
    <Messages :channel="channel" :messages="messages" :parent_id="parentID" />
  </template>

<div class="msgBox">
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
    height: 50px;
    display: table;
  }
}
</style>

