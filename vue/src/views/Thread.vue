<script setup>
import { ref, computed, onMounted } from 'vue';

import Advertisement from '@/components/Advertisement.vue';
import DrawerThread from '@/components/DrawerThread.vue';
import EditBox from '@/components/EditBox.vue';
import Messages from '@/components/Messages.vue';

import { useMessagesStore } from '@/stores/messages.js';
import { useChannelsStore } from '@/stores/channels.js';

import { isEmojiOpen, selectedMessageId, openEmoji, closeEmoji, selectEmoji, calcEmoji, emojiPath } from '@/my/emoji.js';
import { removeMark } from '@/my/markdown.js';

const props = defineProps({
  channel_id: String,
  parentID: String,
  backID: String,
  messageID: String
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

onMounted(async () => {
  channel.value = await getIDB('channel', props.channel_id)
  aliases.value = await getIDBs('alias', 'channelIDIndex', props.channel_id, 10000);
  groups.value = await getIDBs('group', 'channelIDIndex', props.channel_id, 10000);
  threadHead.value = await getIDB('threadHead', props.parentID);
  await makeThreadHead()
  document.title = threadHead.value.title
  const content = await document.getElementById('content');
  content.scrollTop = await content.scrollHeight;
  await window.scrollTo(0, content.scrollHeight);
  readStatus()
  fetched.value = true
});

const message = ref('');
const parentID = ref('');
const messagesStore = useMessagesStore();
const messages = computed(() => {
  return messagesStore.messages;
});

const msg = {
  messageTxt: '',
  messageID: '',
  parentID: props.parentID
};

async function makeThreadHead() {
  if (threadHead.value) {
    if (threadHead.value.aliasName == channel.value.myname) {
      threadHead.value.edit = true;
    }
    let message = threadHead.value;
    message.messageID = threadHead.value.parentID
    message.createdAt = threadHead.value.updatedAt
    messagesStore.insert(message)
  } else {
    const threadHeadValue = {
      channelID: channel.value.channelID,
      parentID: props.parentID,
      title: '新規スレッド',
      messageTxt: '',
      aliasName: channel.value.myname,
      aliasNames: [channel.value.myname],
      displayStatus: 0,
      newThread: true
    }
    if (props.parentID.includes('@')) {
      const parts = props.parentID.split('@');
      const toWhom = parts[0] === channel.value.myname ? parts[1] : parts[0];
      threadHeadValue.title = getSubstring(toWhom, 0, 12);
      threadHeadValue.messageTxt = toWhom;
      threadHeadValue.aliasNames = [...parts]
      parts.forEach(part => {
        const group = groups.value.find(g => g.groupName === part);
        if (group) {
          threadHeadValue.aliasNames.push(...group.aliasNames);
        }
      });
      threadHeadValue.aliasNames = [...new Set(threadHeadValue.aliasNames)]
    }
    threadHead.value = threadHeadValue;
    if (props.backID) {
      const message = await getIDB('thread', props.parentID);
      threadHead.value = message;
      threadHead.value.parentID = props.parentID;
      threadHead.value.title = getSubstring(removeMark(message.messageTxt), 0, 12);
      threadHead.value.messageTxt = message.messageTxt;
      // threadHead.value.threadType = 0;
      threadHead.value.backID = props.backID
      threadHead.value.newReply = true
      const originalThreadHead = await getIDB('threadHead', props.backID)
      threadHead.value.aliasNames = [
        ...new Set([
          ...originalThreadHead.aliasNames,
          ...threadHeadValue.aliasNames
        ])
      ]
      threadHead.value.adminNames = originalThreadHead.adminNames
      messagesStore.insert(message);
    }
  }
}

function readStatus () {
  if (threadHead.value.displayStatus && threadHead.value.displayStatus == 1 || threadHead.value.displayStatus == 2) {
    threadHead.value.displayStatus = 0;
    updIDBone('threadHead', props.parentID, 'displayStatus', 0);
    revertFaviconBadge()
  }
}

function backTo() {
  const backID = threadHead.value.backID;
  if (backID) {
    // const secondPart = backID.replace(props.channel_id, '');
    location.href = '/thread/' + props.channel_id + '/' + backID + '/';
  } else {
    location.href = '/channel/' + props.channel_id + '/';
  }
}

</script>

<template>
<DrawerThread/>
<div id="content">
  <div v-if="fetched">
    <div v-if="threadHead">
      <div class="headTitle">
        <a :href="'/threadHead/' + threadHead.parentID + '/'">
          {{threadHead.title}}
        </a>
      </div>
      <div class="headIcon">
        <span>
          <a @click="backTo"> ⬅ </a>
        </span>
        <span :class="[{ 'selected': copyable, 'emoji-stamp': copyable }]">
          <a @click="copyable = !copyable" > 📄 </a>
        </span>
      </div>
    </div>
    <Messages 
      :channel="channel"
      :aliases="aliases"
      :groups="groups"
      :messages="messages"
      :threadHead="threadHead"
      :copyable="copyable"
      :messageID="messageID" />

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

@media screen and (min-width : 701px) { 
  .headTitle {
    width: 50%;
    min-height: 40px;
    display: inline-block;
  }
  .headIcon {
    width: 40%;
    text-align: right;
    display: inline-block;
  }

}

@media screen and (max-width : 700px) {
  .headTitle {
    width: 45%;
    min-height: 40px;
    margin-left: 50px;
    display: inline-block;
  }
  .headIcon {
    width: 35%;
    text-align: right;
    display: inline-block;
  }

}
</style>

