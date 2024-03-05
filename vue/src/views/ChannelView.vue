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

const props = defineProps({
  id: '',
})

const channel = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
})

const message = {
  messageTxt: '',
  messageID: ''
};

async function fetchData() {
  try {
    const data = await getIDB('channel', props.id);
    channel.value = data;
  } catch (error) {
    channel.value = null;
  }
}

const fetchMessageData = () => {
  return new Promise((resolve, reject) => {
    getIDBs('message', 'channelIDIndex', props.id)
      .then((data) => {
        const latest = data.reverse();
        latest.forEach(message => {
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
  await fetchData();
  await fetchMessageData();
  const content = document.getElementById('content');
  content.scrollTop = content.scrollHeight;
  window.scrollTo(0,content.scrollHeight);
});


const clickEmoji = (messageId, emoji) => {
  const fd = new FormData();
  fd.append('channelID', props.id);
  fd.append('messageID', messageId);
  fd.append('emoji', emoji[0]);
  fd.append('clicked', emoji[2] ? 1 : 0);
  fd.append('aliasName', channel.value.aliasName);
  const request = new Request('/MessageEdit/', {
    method: 'POST',
    body: fd,
  });
  fetch(request)
    .catch((reason)=>{
      alert(reason)
    })
};


</script>

<template>
<DrawerColumn />
<div id="content">
<div class="headTitle">
  <div>
    <RouterLink :to="'/channelInfo/' + channel.channelID"> {{ channel.channelName }} </RouterLink>
  </div>
  <div style="line-height: 50px;">
    &nbsp;
  </div>
</div>

  <template v-if="channel">
    <Messages :channel="channel" :messages="messages" />
  </template>

<div class="msgBox">
  <EditBox :channel="channel" :message="message" />
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

