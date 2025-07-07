<script setup>
import { ref, computed, onBeforeMount, onMounted, nextTick } from 'vue';
import { onBeforeRouteUpdate, useRouter } from 'vue-router';

import Advertisement from '@/components/Advertisement.vue';
import DrawerThread from '@/components/DrawerThread.vue';
import EditBox from '@/components/EditBox.vue';
import Messages from '@/components/Messages.vue';

import { useMessagesStore } from '@/stores/messages.js';
import { useChannelsStore } from '@/stores/channels.js';

import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '@/my/emoji.js';
import { removeMark } from '@/my/markdown.js';

const props = defineProps({
  channel_id: '',
  message_id: '',
  backID: ''
})
localStorage.setItem('channelID', props.channel_id);
const channel = ref(null);
const aliases = ref(null);
const groups = ref(null);
const threadHead = ref({
  parentID: '',
  title: '',
});
const fetched = ref(false);
const copyable = ref(false);
// const copyableValue = computed(() => copyable.value);
onBeforeMount(async () => {
  channel.value = await getIDB('channel', props.channel_id);
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.channel_id, 10000);
  groups.value = await getIDBs('group', 'channelIDIndex', props.channel_id, 10000);
  threadHead.value = await getIDB('threadHead', props.channel_id + props.message_id);
  await makeThreadHead();
  fetched.value = await true;
  const content = await document.getElementById('content');
  content.scrollTop = await content.scrollHeight;
  await window.scrollTo(0, content.scrollHeight);
  readStatus();
});

const message = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
});

// /thread/I0JH/I0JH1tamySv/?backID=I0JHvaj
// /thread/I0JH/gdK/

const msg = {
  messageTxt: '',
  messageID: '',
  parentID: props.channel_id + props.message_id
};

async function makeThreadHead() {
  if (threadHead.value) {
    if (threadHead.value.aliasName == channel.value.myname) {
      threadHead.value.edit = true;
    }
    let message = threadHead.value;
    message.messageID = threadHead.value.parentID;
    messagesStore.insert(message);
  } else {
    const threadHeadValue = {
      channelID: channel.value.channelID,
      parentID: props.message_id,
      title: '新規スレッド',
      messageTxt: '',
      aliasName: channel.value.myname,
      aliasImg: channel.value.myimg,
      aliasNames: [channel.value.myname], 
      displayStatus: 0
    }
    if (props.message_id.includes('@')) {
      const parts = props.message_id.split('@');
      const toWhom = parts[0] === channel.value.myname ? parts[1] : parts[0];
      threadHeadValue.title = getSubstring(toWhom, 0, 12);
      threadHeadValue.messageTxt = toWhom;
      threadHeadValue.aliasNames = [...parts];
      parts.forEach(part => {
        const group = groups.value.find(g => g.groupName === part);
        if (group) {
          threadHeadValue.aliasNames.push(...group.aliasNames);
        }
      });
      threadHeadValue.aliasNames = [...new Set(threadHeadValue.aliasNames)];
    }
    threadHead.value = threadHeadValue;
    if (props.backID) {
      const message = await getIDB('thread', props.channel_id + props.message_id);
      threadHead.value = message;
      threadHead.value.parentID = props.message_id;
      threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
      threadHead.value.messageTxt = message.messageTxt;
      threadHead.value.threadType = 0;
      threadHead.value.backID = props.backID;
      messagesStore.insert(message);
    }
  }
}

function readStatus () {
  if (threadHead.value.displayStatus && threadHead.value.displayStatus == 1 || threadHead.value.displayStatus == 2) {
    threadHead.value.displayStatus = 0;
    updIDBone('threadHead', props.channel_id + props.message_id, 'displayStatus', 0)
      .catch((error) => {
        console.error(error);
      });
    const favicon = document.querySelector('link[rel="icon"]');
    favicon.href = '/favicon.ico';
  }
}

function backTo() {
  const backID = threadHead.value.backID;
  if (backID) {
    const secondPart = backID.replace(props.channel_id, '');
    location.href = '/thread/' + props.channel_id + '/' + secondPart + '/';
  } else {
    location.href = '/channel/' + props.channel_id + '/';
  }
}

</script>

<template>
<DrawerThread />
<div id="content">
  <div v-if="fetched">
    <div v-if="threadHead" class="headTitle">
      <div>
        <a :href="'/threadHead/' + threadHead.parentID + '/'">
          {{threadHead.title}}
        </a>
      </div>
      <span>
        <a @click="backTo"> ⬅ </a>
      </span>
      <span :class="{ 'selected': copyable }">
        <a @click="copyable = !copyable" > 📄 </a>
      </span>
    </div>
    <Messages 
      :channel="channel"
      :aliases="aliases"
      :groups="groups"
      :messages="messages"
      :threadHead="threadHead"
      :copyable="copyable" />

    <div class="editText">
      <EditBox 
        :channel="channel"
        :aliases="aliases"
        :groups="groups"
        :message="msg"
        :threadHead="threadHead" />
    </div>
  <br>
  </div>
  <div v-if="!fetched"><br><br> Loading... or Something Went </div>
</div>
<div id="ad_right"> <Advertisement /> <Advertisement /> <Advertisement /> </div>

</template>

<style>

#content {
  height: 100%;
  overflow-y: auto;
}

.headTitle div {
  width: 90%;
}

.headTitle span {
  line-height: 50px;
  width: 50px;
  text-align: center;
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

